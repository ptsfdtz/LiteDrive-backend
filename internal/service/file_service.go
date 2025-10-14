package service

import (
	"crypto/md5"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"

	"online-disk-server/internal/model"
	"online-disk-server/internal/repository"
	"online-disk-server/internal/storage"

	"gorm.io/gorm"
)

type FileService struct {
	db       *gorm.DB
	fileRepo *repository.FileRepository
	storage  storage.Storage
}

func NewFileService(db *gorm.DB, storage storage.Storage) *FileService {
	return &FileService{
		db:       db,
		fileRepo: repository.NewFileRepository(db),
		storage:  storage,
	}
}

// UploadFile 上传文件
func (s *FileService) UploadFile(userID uint, file *multipart.FileHeader, parentID uint) (*model.File, error) {
	// 打开上传的文件
	src, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	// 计算文件哈希
	hash := md5.New()
	size, err := io.Copy(hash, src)
	if err != nil {
		return nil, err
	}
	hashStr := fmt.Sprintf("%x", hash.Sum(nil))

	// 重新定位到文件开始
	src.Seek(0, io.SeekStart)

	// 生成存储路径
	ext := filepath.Ext(file.Filename)
	storagePath := fmt.Sprintf("files/%d/%s%s", userID, hashStr, ext)

	// 上传到存储
	if err := s.storage.Upload(storagePath, src, size); err != nil {
		return nil, err
	}

	// 创建文件记录
	fileModel := &model.File{
		Name:        file.Filename,
		Path:        "/" + strings.TrimPrefix(file.Filename, "/"),
		Size:        size,
		MimeType:    file.Header.Get("Content-Type"),
		Hash:        hashStr,
		UserID:      userID,
		ParentID:    parentID,
		StoragePath: storagePath,
		IsDir:       false,
	}

	if err := s.fileRepo.Create(fileModel); err != nil {
		// 如果数据库失败，清理已上传的文件
		s.storage.Delete(storagePath)
		return nil, err
	}

	return fileModel, nil
}

// GetFile 获取文件信息
func (s *FileService) GetFile(userID, fileID uint) (*model.File, error) {
	return s.fileRepo.FindByIDAndUser(fileID, userID)
}

// DownloadFile 下载文件
func (s *FileService) DownloadFile(userID, fileID uint) (io.ReadCloser, *model.File, error) {
	file, err := s.fileRepo.FindByIDAndUser(fileID, userID)
	if err != nil {
		return nil, nil, err
	}

	reader, err := s.storage.Download(file.StoragePath)
	if err != nil {
		return nil, nil, err
	}

	return reader, file, nil
}

// ListFiles 列出文件
func (s *FileService) ListFiles(userID, parentID uint, page, limit int) ([]*model.File, int64, error) {
	offset := (page - 1) * limit
	return s.fileRepo.FindByUserAndParent(userID, parentID, offset, limit)
}

// DeleteFile 删除文件
func (s *FileService) DeleteFile(userID, fileID uint) error {
	file, err := s.fileRepo.FindByIDAndUser(fileID, userID)
	if err != nil {
		return err
	}

	// 删除存储中的文件
	if err := s.storage.Delete(file.StoragePath); err != nil {
		// 即使存储删除失败，也继续删除数据库记录
		// 可以记录日志用于后续清理
	}

	// 删除数据库记录
	return s.fileRepo.Delete(fileID, userID)
}

// CreateFolder 创建文件夹
func (s *FileService) CreateFolder(userID uint, name string, parentID uint) (*model.File, error) {
	folder := &model.File{
		Name:     name,
		Path:     "/" + strings.TrimPrefix(name, "/"),
		Size:     0,
		UserID:   userID,
		ParentID: parentID,
		IsDir:    true,
	}

	if err := s.fileRepo.Create(folder); err != nil {
		return nil, err
	}

	return folder, nil
}
