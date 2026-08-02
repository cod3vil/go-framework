package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"github.com/cod3vil/go-framework/internal/system/model"
	"github.com/cod3vil/go-framework/pkg/errs"
	"github.com/cod3vil/go-framework/pkg/tenancy"
	"gorm.io/gorm"
)

// tenantCodeRe 租户编码规则：小写字母开头，字母数字下划线，2-30 位。
var tenantCodeRe = regexp.MustCompile(`^[a-z][a-z0-9_]{1,29}$`)

// 租户相关错误码（system 模块段 2xxx）。
const codeTenantNotFound = 2001

// TenantEnabled 是否启用多租户。
func (s *Service) TenantEnabled() bool {
	return s.Config.Tenant.Enabled && s.Tenancy != nil && s.Tenancy.Supported()
}

// ResolveTenant 按租户编码解析出租户信息（校验存在且启用）。空编码视为主租户。
// 租户注册表始终存于基础库（public）。
func (s *Service) ResolveTenant(ctx context.Context, code string) (*tenancy.Tenant, error) {
	if code == "" || code == tenancy.PrimaryCode {
		return &tenancy.Tenant{Code: tenancy.PrimaryCode, Schema: tenancy.PrimarySchema}, nil
	}
	var t model.SysTenant
	err := s.DB.WithContext(ctx).Where("code = ?", code).First(&t).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errs.New(codeTenantNotFound, "租户不存在")
	}
	if err != nil {
		return nil, errs.ErrInternal.WithCause(err)
	}
	if t.Status != model.StatusEnabled {
		return nil, errs.New(errs.CodeForbidden, "租户已停用")
	}
	return &tenancy.Tenant{Code: t.Code, Schema: t.Schema}, nil
}

// ResolveTenantByCode 实现 middleware.TenantResolver：按编码解析租户及其库句柄。
func (s *Service) ResolveTenantByCode(code string) (*tenancy.Tenant, *gorm.DB, error) {
	ctx := context.Background()
	t, err := s.ResolveTenant(ctx, code)
	if err != nil {
		return nil, nil, err
	}
	db, err := s.TenantDB(t)
	if err != nil {
		return nil, nil, err
	}
	return t, db, nil
}

// TenantDB 返回指定租户的数据库句柄。
func (s *Service) TenantDB(t *tenancy.Tenant) (*gorm.DB, error) {
	if t.Schema == tenancy.PrimarySchema {
		return s.DB, nil
	}
	db, err := s.Tenancy.DB(t.Schema)
	if err != nil {
		return nil, errs.ErrInternal.WithCause(err)
	}
	return db, nil
}

// ListTenants 分页查询租户。
func (s *Service) ListTenants(ctx context.Context, page, pageSize int, name string) ([]model.SysTenant, int64, error) {
	db := s.DB.WithContext(ctx).Model(&model.SysTenant{})
	if name != "" {
		db = db.Where("name LIKE ? OR code LIKE ?", "%"+name+"%", "%"+name+"%")
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, errs.ErrInternal.WithCause(err)
	}
	var list []model.SysTenant
	err := db.Order("id").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	if err != nil {
		return nil, 0, errs.ErrInternal.WithCause(err)
	}
	return list, total, nil
}

// TenantInput 创建租户参数。
type TenantInput struct {
	Code    string
	Name    string
	Contact string
	Remark  string
}

// CreateTenant 开通租户：登记注册表 → 创建 schema → 在其中建表并写入种子数据。
// 整个过程要么全部成功，要么回滚注册表记录并尝试清理 schema。
func (s *Service) CreateTenant(ctx context.Context, in TenantInput, operator uint) (*model.SysTenant, error) {
	if !s.TenantEnabled() {
		return nil, errs.New(errs.CodeForbidden, "未启用多租户（需 PostgreSQL 且 tenant.enabled=true）")
	}
	if !tenantCodeRe.MatchString(in.Code) {
		return nil, errs.New(errs.CodeBadRequest, "租户编码需小写字母开头，仅含字母数字下划线，2-30 位")
	}
	if in.Code == tenancy.PrimaryCode {
		return nil, errs.New(errs.CodeBadRequest, "该编码为保留字")
	}
	var count int64
	s.DB.WithContext(ctx).Model(&model.SysTenant{}).Where("code = ?", in.Code).Count(&count)
	if count > 0 {
		return nil, errs.New(errs.CodeConflict, "租户编码已存在")
	}

	schema := tenancy.SchemaOf(in.Code)
	tenant := model.SysTenant{
		Code: in.Code, Name: in.Name, Schema: schema, Status: model.StatusEnabled,
		Contact: in.Contact, Remark: in.Remark,
	}
	tenant.CreatedBy = operator
	if err := s.DB.WithContext(ctx).Create(&tenant).Error; err != nil {
		return nil, errs.ErrInternal.WithCause(err)
	}

	if err := s.provision(ctx, schema); err != nil {
		// 供给失败：回滚注册记录并清理 schema，避免残留半开通状态。
		s.DB.WithContext(ctx).Unscoped().Delete(&tenant)
		s.DB.WithContext(ctx).Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %q CASCADE", schema))
		return nil, errs.ErrInternal.WithCause(err)
	}
	return &tenant, nil
}

// provision 创建 schema 并在其中建表 + 写入种子数据。
func (s *Service) provision(ctx context.Context, schema string) error {
	if err := s.DB.WithContext(ctx).Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %q", schema)).Error; err != nil {
		return fmt.Errorf("创建 schema 失败: %w", err)
	}
	tdb, err := s.Tenancy.DB(schema)
	if err != nil {
		return err
	}
	if s.MigrateTenant == nil {
		return fmt.Errorf("未配置租户迁移器")
	}
	return s.MigrateTenant(tdb)
}

// getTenant 从基础库查询租户记录。
func (s *Service) getTenant(ctx context.Context, id uint) (*model.SysTenant, error) {
	var t model.SysTenant
	err := s.DB.WithContext(ctx).First(&t, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errs.New(codeTenantNotFound, "租户不存在")
	}
	if err != nil {
		return nil, errs.ErrInternal.WithCause(err)
	}
	return &t, nil
}

// SetTenantStatus 启用/停用租户。
func (s *Service) SetTenantStatus(ctx context.Context, id uint, status int8, operator uint) error {
	t, err := s.getTenant(ctx, id)
	if err != nil {
		return err
	}
	if t.Primary {
		return errs.New(errs.CodeForbidden, "主租户不允许停用")
	}
	if err := s.DB.WithContext(ctx).Model(t).Updates(map[string]any{"status": status, "updated_by": operator}).Error; err != nil {
		return errs.ErrInternal.WithCause(err)
	}
	return nil
}

// DeleteTenant 删除租户并连同其 schema 一并销毁（危险操作）。
func (s *Service) DeleteTenant(ctx context.Context, id uint) error {
	t, err := s.getTenant(ctx, id)
	if err != nil {
		return err
	}
	if t.Primary {
		return errs.New(errs.CodeForbidden, "主租户不允许删除")
	}
	if err := s.DB.WithContext(ctx).Unscoped().Delete(t).Error; err != nil {
		return errs.ErrInternal.WithCause(err)
	}
	if err := s.DB.WithContext(ctx).Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %q CASCADE", t.Schema)).Error; err != nil {
		s.Logger.Warn("删除租户 schema 失败: " + err.Error())
	}
	return nil
}
