package repository

import (
	"online-disk-server/internal/model"

	"gorm.io/gorm"
)

type FileRepository struct {
	db *gorm.DB
}

func NewFileRepository(db *gorm.DB) *FileRepository {
	return &FileRepository{db: db}
}

func (r *FileRepository) Create(file *model.File) error {
	return r.db.Create(file).Error
}

func (r *FileRepository) FindByIDAndUser(fileID, userID uint) (*model.File, error) {
	var file model.File
	if err := r.db.Where("id = ? AND user_id = ?", fileID, userID).First(&file).Error; err != nil {
		return nil, err
	}
	return &file, nil
}

func (r *FileRepository) FindByUserAndParent(userID, parentID uint, offset, limit int) ([]*model.File, int64, error) {
	var files []*model.File
	var total int64

	query := r.db.Where("user_id = ? AND parent_id = ?", userID, parentID)

	if err := query.Model(&model.File{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset(offset).Limit(limit).Order("is_dir DESC, name ASC").Find(&files).Error; err != nil {
		return nil, 0, err
	}

	return files, total, nil
}

func (r *FileRepository) Delete(fileID, userID uint) error {
	return r.db.Where("id = ? AND user_id = ?", fileID, userID).Delete(&model.File{}).Error
}

func (r *FileRepository) FindByHash(hash string, userID uint) (*model.File, error) {
	var file model.File
	if err := r.db.Where("hash = ? AND user_id = ?", hash, userID).First(&file).Error; err != nil {
		return nil, err
	}
	return &file, nil
}
