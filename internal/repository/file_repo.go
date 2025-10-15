package repository

import (
	"online-disk-server/internal/model"

	"strings"

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

// FindChildFolder 查找指定父目录下的子文件夹
func (r *FileRepository) FindChildFolder(userID, parentID uint, name string) (*model.File, error) {
	var folder model.File
	if err := r.db.Where("user_id = ? AND parent_id = ? AND is_dir = ? AND name = ?", userID, parentID, true, name).First(&folder).Error; err != nil {
		return nil, err
	}
	return &folder, nil
}

// FindOrCreateFolder 在父目录下查找文件夹，不存在则创建
func (r *FileRepository) FindOrCreateFolder(userID, parentID uint, name, fullPath string) (*model.File, error) {
	// 先查
	if f, err := r.FindChildFolder(userID, parentID, name); err == nil {
		return f, nil
	}
	// 创建
	folder := &model.File{
		Name:     name,
		Path:     fullPath,
		Size:     0,
		UserID:   userID,
		ParentID: parentID,
		IsDir:    true,
	}
	if err := r.db.Create(folder).Error; err != nil {
		return nil, err
	}
	return folder, nil
}

// FindDirByPath 递归查找用户某 path 下的目录节点（path 以 / 开头，根为 /，不含文件名）
func (r *FileRepository) FindDirByPath(userID uint, path string) (*model.File, error) {
	path = strings.TrimPrefix(path, "/")
	if path == "" {
		// 根目录
		return &model.File{ID: 0, Name: "/", Path: "/", UserID: userID, ParentID: 0, IsDir: true}, nil
	}
	parts := strings.Split(path, "/")
	var parentID uint = 0
	var folder *model.File
	for _, name := range parts {
		if name == "" || name == "." || name == ".." {
			continue
		}
		f, err := r.FindChildFolder(userID, parentID, name)
		if err != nil {
			return nil, err
		}
		folder = f
		parentID = f.ID
	}
	if folder == nil {
		return nil, gorm.ErrRecordNotFound
	}
	return folder, nil
}
