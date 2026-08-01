package service

import (
	"context"
	"errors"

	"github.com/cod3vil/go-framework/internal/system/model"
	"github.com/cod3vil/go-framework/pkg/database"
	"github.com/cod3vil/go-framework/pkg/errs"
	"github.com/cod3vil/go-framework/pkg/utils"
	"gorm.io/gorm"
)

// UserQuery 用户列表查询条件。
type UserQuery struct {
	Page     int
	PageSize int
	Username string
	Nickname string
	Status   int8
	DeptID   uint
}

// ListUsers 分页查询用户。
func (s *Service) ListUsers(ctx context.Context, q UserQuery) ([]model.SysUser, int64, error) {
	db := s.DB.WithContext(ctx).Model(&model.SysUser{})
	if q.Username != "" {
		db = db.Where("username LIKE ?", "%"+q.Username+"%")
	}
	if q.Nickname != "" {
		db = db.Where("nickname LIKE ?", "%"+q.Nickname+"%")
	}
	if q.Status != 0 {
		db = db.Where("status = ?", q.Status)
	}
	if q.DeptID != 0 {
		db = db.Where("dept_id = ?", q.DeptID)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, errs.ErrInternal.WithCause(err)
	}
	var users []model.SysUser
	err := db.Preload("Roles").Preload("Dept").
		Order("id").
		Offset((q.Page - 1) * q.PageSize).Limit(q.PageSize).
		Find(&users).Error
	if err != nil {
		return nil, 0, errs.ErrInternal.WithCause(err)
	}
	return users, total, nil
}

// UserInput 创建/更新用户参数。
type UserInput struct {
	Username string
	Password string
	Nickname string
	Email    string
	Phone    string
	Status   int8
	DeptID   uint
	Remark   string
	RoleIDs  []uint
}

// CreateUser 创建用户并分配角色。
func (s *Service) CreateUser(ctx context.Context, in UserInput, operator uint) (*model.SysUser, error) {
	var count int64
	s.DB.WithContext(ctx).Model(&model.SysUser{}).Where("username = ?", in.Username).Count(&count)
	if count > 0 {
		return nil, errs.New(errs.CodeConflict, "用户名已存在")
	}
	hash, err := utils.HashPassword(in.Password)
	if err != nil {
		return nil, errs.ErrInternal.WithCause(err)
	}
	user := model.SysUser{
		Username: in.Username,
		Password: hash,
		Nickname: in.Nickname,
		Email:    in.Email,
		Phone:    in.Phone,
		Status:   in.Status,
		DeptID:   in.DeptID,
		Remark:   in.Remark,
	}
	user.CreatedBy = operator
	err = database.Tx(ctx, s.DB, func(tx *gorm.DB) error {
		if err := tx.Create(&user).Error; err != nil {
			return err
		}
		return s.replaceUserRoles(tx, &user, in.RoleIDs)
	})
	if err != nil {
		return nil, errs.ErrInternal.WithCause(err)
	}
	return &user, nil
}

// GetUser 查询用户详情。
func (s *Service) GetUser(ctx context.Context, id uint) (*model.SysUser, error) {
	var user model.SysUser
	err := s.DB.WithContext(ctx).Preload("Roles").Preload("Dept").First(&user, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errs.ErrNotFound
	}
	if err != nil {
		return nil, errs.ErrInternal.WithCause(err)
	}
	return &user, nil
}

// UpdateUser 更新用户资料与角色（不含密码）。
func (s *Service) UpdateUser(ctx context.Context, id uint, in UserInput, operator uint) error {
	user, err := s.GetUser(ctx, id)
	if err != nil {
		return err
	}
	if user.IsAdmin() && in.Status == model.StatusDisabled {
		return errs.New(errs.CodeForbidden, "内置管理员不允许停用")
	}
	updates := map[string]any{
		"nickname": in.Nickname, "email": in.Email, "phone": in.Phone,
		"status": in.Status, "dept_id": in.DeptID, "remark": in.Remark,
		"updated_by": operator,
	}
	err = database.Tx(ctx, s.DB, func(tx *gorm.DB) error {
		if err := tx.Model(user).Updates(updates).Error; err != nil {
			return err
		}
		// 内置管理员的角色固定，不随请求变更。
		if user.IsAdmin() {
			return nil
		}
		return s.replaceUserRoles(tx, user, in.RoleIDs)
	})
	if err != nil {
		return errs.ErrInternal.WithCause(err)
	}
	return nil
}

// DeleteUser 删除用户（硬删除，唯一键立即可复用）。
func (s *Service) DeleteUser(ctx context.Context, id uint) error {
	user, err := s.GetUser(ctx, id)
	if err != nil {
		return err
	}
	if user.IsAdmin() {
		return errs.New(errs.CodeForbidden, "内置管理员不允许删除")
	}
	err = database.Tx(ctx, s.DB, func(tx *gorm.DB) error {
		if err := tx.Model(user).Association("Roles").Clear(); err != nil {
			return err
		}
		return tx.Unscoped().Delete(user).Error
	})
	if err != nil {
		return errs.ErrInternal.WithCause(err)
	}
	return nil
}

// ResetPassword 管理员重置用户密码。
func (s *Service) ResetPassword(ctx context.Context, id uint, password string, operator uint) error {
	user, err := s.GetUser(ctx, id)
	if err != nil {
		return err
	}
	hash, err := utils.HashPassword(password)
	if err != nil {
		return errs.ErrInternal.WithCause(err)
	}
	err = s.DB.WithContext(ctx).Model(user).
		Updates(map[string]any{"password": hash, "updated_by": operator}).Error
	if err != nil {
		return errs.ErrInternal.WithCause(err)
	}
	return nil
}

// SetUserStatus 启用/停用用户。
func (s *Service) SetUserStatus(ctx context.Context, id uint, status int8, operator uint) error {
	user, err := s.GetUser(ctx, id)
	if err != nil {
		return err
	}
	if user.IsAdmin() {
		return errs.New(errs.CodeForbidden, "内置管理员不允许停用")
	}
	err = s.DB.WithContext(ctx).Model(user).
		Updates(map[string]any{"status": status, "updated_by": operator}).Error
	if err != nil {
		return errs.ErrInternal.WithCause(err)
	}
	return nil
}

func (s *Service) replaceUserRoles(tx *gorm.DB, user *model.SysUser, roleIDs []uint) error {
	roles := make([]model.SysRole, 0, len(roleIDs))
	if len(roleIDs) > 0 {
		if err := tx.Find(&roles, roleIDs).Error; err != nil {
			return err
		}
	}
	return tx.Model(user).Association("Roles").Replace(roles)
}
