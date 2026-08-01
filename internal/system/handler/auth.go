package handler

import (
	"github.com/cod3vil/go-framework/internal/middleware"
	"github.com/cod3vil/go-framework/internal/system/service"
	"github.com/cod3vil/go-framework/pkg/jwtx"
	"github.com/cod3vil/go-framework/pkg/response"
	"github.com/gin-gonic/gin"
)

// Captcha 获取图形验证码。
// GET /api/v1/auth/captcha
func (h *Handler) Captcha(c *gin.Context) {
	id, img, err := h.svc.Captcha.Generate()
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, gin.H{
		"captchaId":  id,
		"captchaImg": img,
		"enabled":    h.svc.Config.App.CaptchaEnabled,
	})
}

type loginReq struct {
	Username    string `json:"username" binding:"required"`
	Password    string `json:"password" binding:"required"`
	CaptchaID   string `json:"captchaId"`
	CaptchaCode string `json:"captchaCode"`
}

// Login 账号密码登录。
// POST /api/v1/auth/login
func (h *Handler) Login(c *gin.Context) {
	var req loginReq
	if !bindJSON(c, &req) {
		return
	}
	pair, err := h.svc.Login(c.Request.Context(), service.LoginInput{
		Username:    req.Username,
		Password:    req.Password,
		CaptchaID:   req.CaptchaID,
		CaptchaCode: req.CaptchaCode,
		IP:          c.ClientIP(),
		UserAgent:   c.Request.UserAgent(),
	})
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, pair)
}

type refreshReq struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

// Refresh 刷新令牌（旧 Refresh Token 轮换失效）。
// POST /api/v1/auth/refresh
func (h *Handler) Refresh(c *gin.Context) {
	var req refreshReq
	if !bindJSON(c, &req) {
		return
	}
	pair, err := h.svc.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, pair)
}

type logoutReq struct {
	RefreshToken string `json:"refreshToken"`
}

// Logout 登出，当前令牌加入黑名单。
// POST /api/v1/auth/logout
func (h *Handler) Logout(c *gin.Context) {
	var req logoutReq
	_ = c.ShouldBindJSON(&req) // 请求体可选
	if v, ok := c.Get(middleware.CtxClaims); ok {
		if claims, ok := v.(*jwtx.Claims); ok {
			h.svc.Logout(c.Request.Context(), claims, req.RefreshToken)
		}
	}
	response.OK(c, nil)
}

// UserInfo 当前用户信息（资料、角色、权限标识、菜单树）。
// GET /api/v1/auth/userinfo
func (h *Handler) UserInfo(c *gin.Context) {
	info, err := h.svc.UserInfo(c.Request.Context(), middleware.UserID(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, info)
}
