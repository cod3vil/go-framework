package service

import (
	"context"
	"mime/multipart"

	"github.com/cod3vil/go-framework/internal/system/model"
	"github.com/cod3vil/go-framework/pkg/errs"
)

// UploadFile 保存上传文件并落库记录。
func (s *Service) UploadFile(ctx context.Context, fh *multipart.FileHeader, operator uint) (*model.SysFile, error) {
	res, err := s.Uploader.Save(ctx, fh)
	if err != nil {
		return nil, errs.New(errs.CodeBadRequest, err.Error())
	}
	file := model.SysFile{
		Name: res.Filename, Key: res.Key, URL: res.URL,
		Ext: res.Ext, Size: res.Size, Storage: s.Config.Upload.Driver,
	}
	file.CreatedBy = operator
	if err := s.DB.WithContext(ctx).Create(&file).Error; err != nil {
		// 落库失败时清理已写入的物理文件，避免孤儿文件。
		_ = s.Uploader.Delete(ctx, res.Key)
		return nil, errs.ErrInternal.WithCause(err)
	}
	return &file, nil
}

// FileQuery 文件列表查询条件。
type FileQuery struct {
	Page     int
	PageSize int
	Name     string
	Ext      string
}

// ListFiles 分页查询文件（按时间倒序）。
func (s *Service) ListFiles(ctx context.Context, q FileQuery) ([]model.SysFile, int64, error) {
	db := s.DB.WithContext(ctx).Model(&model.SysFile{})
	if q.Name != "" {
		db = db.Where("name LIKE ?", "%"+q.Name+"%")
	}
	if q.Ext != "" {
		db = db.Where("ext = ?", q.Ext)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, errs.ErrInternal.WithCause(err)
	}
	var files []model.SysFile
	err := db.Order("id DESC").Offset((q.Page - 1) * q.PageSize).Limit(q.PageSize).Find(&files).Error
	if err != nil {
		return nil, 0, errs.ErrInternal.WithCause(err)
	}
	return files, total, nil
}

// GetFile 查询文件记录。
func (s *Service) GetFile(ctx context.Context, id uint) (*model.SysFile, error) {
	var file model.SysFile
	if err := s.firstByID(ctx, &file, id); err != nil {
		return nil, err
	}
	return &file, nil
}

// DeleteFile 删除文件记录及物理文件。
func (s *Service) DeleteFile(ctx context.Context, id uint) error {
	file, err := s.GetFile(ctx, id)
	if err != nil {
		return err
	}
	if err := s.Uploader.Delete(ctx, file.Key); err != nil {
		s.Logger.Warn("删除物理文件失败: " + err.Error())
	}
	if err := s.DB.WithContext(ctx).Unscoped().Delete(file).Error; err != nil {
		return errs.ErrInternal.WithCause(err)
	}
	return nil
}
