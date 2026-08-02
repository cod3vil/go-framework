// Package jwtx 封装 JWT 双令牌（Access + Refresh）的签发与校验。
package jwtx

import (
	"errors"
	"time"

	"github.com/cod3vil/go-framework/pkg/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// 令牌类型。
const (
	TypeAccess  = "access"
	TypeRefresh = "refresh"
)

// 校验失败错误。
var (
	ErrInvalidToken = errors.New("jwtx: invalid token")
	ErrWrongType    = errors.New("jwtx: wrong token type")
)

// Claims 业务载荷：用户身份、部门与角色，随 Access Token 下发。
type Claims struct {
	UserID    uint     `json:"uid"`
	Username  string   `json:"uname"`
	DeptID    uint     `json:"dept"`
	RoleKeys  []string `json:"roles"`
	TokenType string   `json:"typ"`
	jwt.RegisteredClaims
}

// Manager 令牌管理器。
type Manager struct {
	secret     []byte
	issuer     string
	accessTTL  time.Duration
	refreshTTL time.Duration
}

// NewManager 根据配置创建管理器。
func NewManager(cfg config.JWTConfig) *Manager {
	return &Manager{
		secret:     []byte(cfg.Secret),
		issuer:     cfg.Issuer,
		accessTTL:  cfg.AccessTTL(),
		refreshTTL: cfg.RefreshTTL(),
	}
}

// Pair 一对令牌。ExpiresIn 为 Access Token 有效秒数，供客户端提前刷新。
type Pair struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int64  `json:"expiresIn"`
}

// GeneratePair 为用户签发 Access + Refresh 令牌对。
func (m *Manager) GeneratePair(userID uint, username string, deptID uint, roleKeys []string) (*Pair, error) {
	access, err := m.generate(TypeAccess, m.accessTTL, userID, username, deptID, roleKeys)
	if err != nil {
		return nil, err
	}
	// Refresh Token 不携带角色，刷新时以数据库最新角色为准。
	refresh, err := m.generate(TypeRefresh, m.refreshTTL, userID, username, deptID, nil)
	if err != nil {
		return nil, err
	}
	return &Pair{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    int64(m.accessTTL.Seconds()),
	}, nil
}

func (m *Manager) generate(typ string, ttl time.Duration, userID uint, username string, deptID uint, roleKeys []string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:    userID,
		Username:  username,
		DeptID:    deptID,
		RoleKeys:  roleKeys,
		TokenType: typ,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			Issuer:    m.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
}

// Parse 解析并校验令牌；wantType 非空时同时校验令牌类型。
func (m *Manager) Parse(tokenStr, wantType string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return m.secret, nil
	})
	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, ErrInvalidToken
	}
	if wantType != "" && claims.TokenType != wantType {
		return nil, ErrWrongType
	}
	return claims, nil
}
