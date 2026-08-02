package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/cod3vil/go-framework/internal/middleware"
	"github.com/cod3vil/go-framework/internal/system/model"
	"github.com/cod3vil/go-framework/pkg/errs"
	"github.com/cod3vil/go-framework/pkg/jwtx"
	"github.com/cod3vil/go-framework/pkg/utils"
	"go.uber.org/zap"
)

const (
	loginFailKeyPrefix = "auth:login:fail:"
	maxLoginFails      = 5
	lockDuration       = 10 * time.Minute
)

// LoginInput 登录参数。
type LoginInput struct {
	Username    string
	Password    string
	CaptchaID   string
	CaptchaCode string
	IP          string
	UserAgent   string
}

// Login 账号密码登录：验证码 → 失败锁定 → 密码校验 → 签发令牌 + 登录日志。
func (s *Service) Login(ctx context.Context, in LoginInput) (*jwtx.Pair, error) {
	if s.Config.App.CaptchaEnabled && !s.Captcha.Verify(in.CaptchaID, in.CaptchaCode) {
		return nil, errs.New(errs.CodeBadRequest, "验证码错误或已过期")
	}

	failKey := loginFailKeyPrefix + in.Username
	if fails := s.getFailCount(ctx, failKey); fails >= maxLoginFails {
		return nil, errs.Newf(errs.CodeTooManyReq, "登录失败次数过多，请 %d 分钟后重试", int(lockDuration.Minutes()))
	}

	var user model.SysUser
	err := s.DB.WithContext(ctx).Preload("Roles").Where("username = ?", in.Username).First(&user).Error
	if err != nil || !utils.CheckPassword(user.Password, in.Password) {
		s.recordLoginFail(ctx, in, failKey)
		return nil, errs.New(errs.CodeBadRequest, "用户名或密码错误")
	}
	if user.Status != model.StatusEnabled {
		s.writeLoginLog(ctx, in, model.LoginFailed, "账号已停用")
		return nil, errs.New(errs.CodeForbidden, "账号已停用，请联系管理员")
	}

	pair, err := s.JWT.GeneratePair(user.ID, user.Username, user.DeptID, user.RoleKeys())
	if err != nil {
		return nil, errs.ErrInternal.WithCause(err)
	}

	_ = s.Cache.Del(ctx, failKey)
	now := time.Now()
	s.DB.WithContext(ctx).Model(&user).Updates(map[string]any{
		"last_login_at": now, "last_login_ip": in.IP,
	})
	s.writeLoginLog(ctx, in, model.LoginSuccess, "登录成功")
	return pair, nil
}

// Refresh 用 Refresh Token 换发新令牌对；旧 Refresh Token 立即失效（轮换）。
func (s *Service) Refresh(ctx context.Context, refreshToken string) (*jwtx.Pair, error) {
	claims, err := s.JWT.Parse(refreshToken, jwtx.TypeRefresh)
	if err != nil {
		return nil, errs.ErrUnauthorized
	}
	if banned, _ := s.Cache.Exists(ctx, middleware.BlacklistKeyPrefix+claims.ID); banned {
		return nil, errs.ErrUnauthorized
	}

	var user model.SysUser
	if err := s.DB.WithContext(ctx).Preload("Roles").First(&user, claims.UserID).Error; err != nil {
		return nil, errs.ErrUnauthorized
	}
	if user.Status != model.StatusEnabled {
		return nil, errs.New(errs.CodeForbidden, "账号已停用")
	}

	s.blacklistClaims(ctx, claims)
	pair, err := s.JWT.GeneratePair(user.ID, user.Username, user.DeptID, user.RoleKeys())
	if err != nil {
		return nil, errs.ErrInternal.WithCause(err)
	}
	return pair, nil
}

// Logout 登出：将当前 Access Token（及可选的 Refresh Token）加入黑名单。
func (s *Service) Logout(ctx context.Context, accessClaims *jwtx.Claims, refreshToken string) {
	s.blacklistClaims(ctx, accessClaims)
	if refreshToken != "" {
		if claims, err := s.JWT.Parse(refreshToken, jwtx.TypeRefresh); err == nil {
			s.blacklistClaims(ctx, claims)
		}
	}
}

// blacklistClaims 按剩余有效期将令牌 jti 写入黑名单。
func (s *Service) blacklistClaims(ctx context.Context, claims *jwtx.Claims) {
	ttl := time.Until(claims.ExpiresAt.Time)
	if ttl <= 0 {
		return
	}
	if err := s.Cache.Set(ctx, middleware.BlacklistKeyPrefix+claims.ID, "1", ttl); err != nil {
		s.Logger.Warn("写入令牌黑名单失败", zap.Error(err))
	}
}

// UserInfoOutput 当前登录用户信息，含前端动态路由所需的菜单树与权限标识。
type UserInfoOutput struct {
	User  *model.SysUser   `json:"user"`
	Roles []string         `json:"roles"`
	Perms []string         `json:"perms"`
	Menus []*model.SysMenu `json:"menus"`
}

// UserInfo 返回当前用户的资料、角色、权限标识与可见菜单树。
func (s *Service) UserInfo(ctx context.Context, userID uint) (*UserInfoOutput, error) {
	var user model.SysUser
	if err := s.DB.WithContext(ctx).Preload("Roles").Preload("Dept").First(&user, userID).Error; err != nil {
		return nil, errs.ErrNotFound.WithCause(err)
	}

	menus, err := s.menusForUser(ctx, &user)
	if err != nil {
		return nil, errs.ErrInternal.WithCause(err)
	}

	perms := make([]string, 0)
	if user.IsAdmin() {
		perms = append(perms, "*:*:*")
	}
	visible := make([]*model.SysMenu, 0)
	for _, m := range menus {
		if m.Perm != "" && !user.IsAdmin() {
			perms = append(perms, m.Perm)
		}
		if m.Type != model.MenuTypeButton && m.Visible == 1 {
			visible = append(visible, m)
		}
	}

	return &UserInfoOutput{
		User:  &user,
		Roles: user.RoleKeys(),
		Perms: perms,
		Menus: buildMenuTree(visible, 0),
	}, nil
}

// menusForUser 返回用户可用的全部菜单（admin 为全量，其余按角色并集）。
func (s *Service) menusForUser(ctx context.Context, user *model.SysUser) ([]*model.SysMenu, error) {
	db := s.DB.WithContext(ctx)
	var menus []*model.SysMenu
	if user.IsAdmin() {
		err := db.Where("status = ?", model.StatusEnabled).Order("sort, id").Find(&menus).Error
		return menus, err
	}
	roleIDs := make([]uint, 0, len(user.Roles))
	for _, r := range user.Roles {
		if r.Status == model.StatusEnabled {
			roleIDs = append(roleIDs, r.ID)
		}
	}
	if len(roleIDs) == 0 {
		return menus, nil
	}
	err := db.Distinct("sys_menu.*").
		Joins("JOIN sys_role_menu rm ON rm.sys_menu_id = sys_menu.id AND rm.sys_role_id IN ?", roleIDs).
		Where("sys_menu.status = ?", model.StatusEnabled).
		Order("sys_menu.sort, sys_menu.id").
		Find(&menus).Error
	return menus, err
}

func (s *Service) getFailCount(ctx context.Context, key string) int {
	val, err := s.Cache.Get(ctx, key)
	if err != nil {
		return 0
	}
	n, _ := strconv.Atoi(val)
	return n
}

func (s *Service) recordLoginFail(ctx context.Context, in LoginInput, failKey string) {
	fails := s.getFailCount(ctx, failKey) + 1
	_ = s.Cache.Set(ctx, failKey, strconv.Itoa(fails), lockDuration)
	s.writeLoginLog(ctx, in, model.LoginFailed,
		fmt.Sprintf("用户名或密码错误(第%d次)", fails))
}

func (s *Service) writeLoginLog(ctx context.Context, in LoginInput, status int8, msg string) {
	log := model.SysLoginLog{
		Username:  in.Username,
		IP:        in.IP,
		UserAgent: in.UserAgent,
		Status:    status,
		Msg:       msg,
	}
	if err := s.DB.WithContext(ctx).Create(&log).Error; err != nil {
		s.Logger.Warn("写登录日志失败", zap.Error(err))
	}
}
