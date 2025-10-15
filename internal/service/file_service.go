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
	baseName := filepath.Base(file.Filename) // 避免包含相对路径
	ext := filepath.Ext(baseName)
	storagePath := fmt.Sprintf("files/%d/%s%s", userID, hashStr, ext)

	// 上传到存储
	if err := s.storage.Upload(storagePath, src, size); err != nil {
		return nil, err
	}

	// 创建文件记录
	fileModel := &model.File{
		Name:        baseName,
		Path:        "/" + strings.TrimPrefix(baseName, "/"),
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
	parentPath := ""
	if parentID != 0 {
		parent, err := s.fileRepo.FindByIDAndUser(parentID, userID)
		if err != nil {
			return nil, err
		}
		parentPath = parent.Path
	}
	var fullPath string
	if parentPath == "" || parentPath == "/" {
		fullPath = "/" + strings.TrimPrefix(name, "/")
	} else {
		fullPath = parentPath
		if !strings.HasSuffix(fullPath, "/") {
			fullPath += "/"
		}
		fullPath += strings.TrimPrefix(name, "/")
	}
	folder := &model.File{
		Name:     name,
		Path:     fullPath,
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

// UploadFilesWithRelativePaths 批量上传文件（带相对路径），自动创建中间文件夹
// 参数：
// - files: 来自 multipart 的文件列表
// - relPaths: 与 files 按索引一一对应的相对路径（例如 "sub/dir/file.txt"），可为空字符串
// - parentID: 作为起始父目录ID（0 为根）
func (s *FileService) UploadFilesWithRelativePaths(userID uint, files []*multipart.FileHeader, relPaths []string, parentID uint) ([]*model.File, error) {
	result := make([]*model.File, 0, len(files))

	for i, fh := range files {
		rel := ""
		if i < len(relPaths) {
			rel = relPaths[i]
		}

		// 解析相对路径中的目录和文件名
		rel = filepath.ToSlash(strings.TrimPrefix(rel, "/"))
		dirPart := filepath.Dir(rel)
		baseName := filepath.Base(rel)
		if baseName == "." || baseName == "" || baseName == ".." {
			// fallback 到上传文件名
			baseName = filepath.Base(fh.Filename)
		}

		// 逐级确保目录存在
		currentParent := parentID
		if dirPart != "." && dirPart != "" {
			parts := strings.Split(dirPart, "/")
			var fullPathBuilder strings.Builder
			fullPathBuilder.WriteString("/")
			for idx, p := range parts {
				if p == "" || p == "." || p == ".." {
					continue
				}
				if idx > 0 {
					fullPathBuilder.WriteString("/")
				}
				fullPathBuilder.WriteString(p)

				fullPath := fullPathBuilder.String()
				folder, err := s.fileRepo.FindOrCreateFolder(userID, currentParent, p, fullPath)
				if err != nil {
					return nil, err
				}
				currentParent = folder.ID
			}
		}

		// 为当前文件构造一个新的 FileHeader，保持原内容，但名称使用 baseName
		// Gin 的 FileHeader 中 Filename 影响我们保存的显示名称
		// 直接修改 fh.Filename 不会影响底层内容
		originalName := fh.Filename
		fh.Filename = baseName
		f, err := s.UploadFile(userID, fh, currentParent)
		// 恢复原始名字以避免副作用（尽管生命周期仅此处）
		fh.Filename = originalName
		if err != nil {
			return nil, err
		}
		result = append(result, f)
	}
	return result, nil
}

// FindDirByPath 递归查找 path 下的目录节点
func (s *FileService) FindDirByPath(userID uint, path string) (*model.File, error) {
	return s.fileRepo.FindDirByPath(userID, path)
}

// FindOrCreateFolder 递归查找或创建目录
func (s *FileService) FindOrCreateFolder(userID, parentID uint, name, fullPath string) (*model.File, error) {
	return s.fileRepo.FindOrCreateFolder(userID, parentID, name, fullPath)
}
