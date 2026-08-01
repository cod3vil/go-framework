// Package captcha 提供图形验证码，答案存储在统一 Cache 中，
// 天然支持多实例部署（配合 Redis）。
package captcha

import (
	"context"
	"time"

	"github.com/cod3vil/go-framework/pkg/cache"
	"github.com/mojocn/base64Captcha"
)

const (
	keyPrefix = "captcha:"
	ttl       = 5 * time.Minute
)

// Captcha 图形验证码生成与校验器。
type Captcha struct {
	driver base64Captcha.Driver
	store  base64Captcha.Store
}

// New 创建 4 位数字验证码生成器。
func New(c cache.Cache) *Captcha {
	return &Captcha{
		driver: base64Captcha.NewDriverDigit(80, 240, 4, 0.7, 80),
		store:  &cacheStore{cache: c},
	}
}

// Generate 生成验证码，返回验证码 ID 与 base64 图片。
func (cp *Captcha) Generate() (id, b64 string, err error) {
	c := base64Captcha.NewCaptcha(cp.driver, cp.store)
	id, b64, _, err = c.Generate()
	return id, b64, err
}

// Verify 校验并销毁验证码（一次性）。
func (cp *Captcha) Verify(id, answer string) bool {
	if id == "" || answer == "" {
		return false
	}
	return cp.store.Verify(id, answer, true)
}

// cacheStore 适配 base64Captcha.Store 到框架 Cache 接口。
type cacheStore struct {
	cache cache.Cache
}

func (s *cacheStore) Set(id, value string) error {
	return s.cache.Set(context.Background(), keyPrefix+id, value, ttl)
}

func (s *cacheStore) Get(id string, clear bool) string {
	ctx := context.Background()
	val, err := s.cache.Get(ctx, keyPrefix+id)
	if err != nil {
		return ""
	}
	if clear {
		_ = s.cache.Del(ctx, keyPrefix+id)
	}
	return val
}

func (s *cacheStore) Verify(id, answer string, clear bool) bool {
	return s.Get(id, clear) == answer
}
