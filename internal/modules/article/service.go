package article

import (
	"context"
	"errors"

	"github.com/cod3vil/go-framework/pkg/errs"
	"gorm.io/gorm"
)

// 业务错误码：业务模块从 10xxx 起分配自己的一段（见 pkg/errs 约定）。
const (
	codeArticleNotFound = 10001
)

// Service 文章业务逻辑。
type Service struct {
	db *gorm.DB
}

// NewService 创建服务。
func NewService(db *gorm.DB) *Service { return &Service{db: db} }

// Query 列表查询条件。
type Query struct {
	Page     int
	PageSize int
	Title    string
	Status   int8
}

// List 分页查询文章。
func (s *Service) List(ctx context.Context, q Query) ([]Article, int64, error) {
	db := s.db.WithContext(ctx).Model(&Article{})
	if q.Title != "" {
		db = db.Where("title LIKE ?", "%"+q.Title+"%")
	}
	if q.Status != 0 {
		db = db.Where("status = ?", q.Status)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, errs.ErrInternal.WithCause(err)
	}
	var list []Article
	err := db.Order("id DESC").Offset((q.Page - 1) * q.PageSize).Limit(q.PageSize).Find(&list).Error
	if err != nil {
		return nil, 0, errs.ErrInternal.WithCause(err)
	}
	return list, total, nil
}

// Get 查询文章详情并自增浏览量。
func (s *Service) Get(ctx context.Context, id uint) (*Article, error) {
	var a Article
	err := s.db.WithContext(ctx).First(&a, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errs.New(codeArticleNotFound, "文章不存在")
	}
	if err != nil {
		return nil, errs.ErrInternal.WithCause(err)
	}
	// 自增浏览量，并在返回值中体现本次访问后的最新值。
	s.db.WithContext(ctx).Model(&a).UpdateColumn("views", gorm.Expr("views + 1"))
	a.Views++
	return &a, nil
}

// Input 创建/更新参数。
type Input struct {
	Title   string
	Author  string
	Content string
	Status  int8
}

// Create 新建文章。
func (s *Service) Create(ctx context.Context, in Input, operator uint) (*Article, error) {
	a := Article{
		Title: in.Title, Author: in.Author, Content: in.Content,
		Status: in.Status, CreatedBy: operator,
	}
	if err := s.db.WithContext(ctx).Create(&a).Error; err != nil {
		return nil, errs.ErrInternal.WithCause(err)
	}
	return &a, nil
}

// Update 更新文章。
func (s *Service) Update(ctx context.Context, id uint, in Input) error {
	res := s.db.WithContext(ctx).Model(&Article{}).Where("id = ?", id).Updates(map[string]any{
		"title": in.Title, "author": in.Author, "content": in.Content, "status": in.Status,
	})
	if res.Error != nil {
		return errs.ErrInternal.WithCause(res.Error)
	}
	if res.RowsAffected == 0 {
		return errs.New(codeArticleNotFound, "文章不存在")
	}
	return nil
}

// Delete 删除文章。
func (s *Service) Delete(ctx context.Context, id uint) error {
	res := s.db.WithContext(ctx).Delete(&Article{}, id)
	if res.Error != nil {
		return errs.ErrInternal.WithCause(res.Error)
	}
	if res.RowsAffected == 0 {
		return errs.New(codeArticleNotFound, "文章不存在")
	}
	return nil
}
