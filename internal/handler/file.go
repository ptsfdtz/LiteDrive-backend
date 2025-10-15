package handler

import (
	"net/http"
	"strconv"
	"strings"

	"online-disk-server/internal/middleware"
	"online-disk-server/internal/service"

	"github.com/gin-gonic/gin"
)

type FileHandler struct {
	fileService *service.FileService
}

func NewFileHandler(fileService *service.FileService) *FileHandler {
	return &FileHandler{fileService: fileService}
}

// Upload 文件上传
func (h *FileHandler) Upload(c *gin.Context) {
	// 获取用户ID
	userID, exists := c.Get(middleware.CtxUserID)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	uid := userID.(uint)

	// 优先支持 path 参数
	parentID := uint(0)
	path := c.DefaultPostForm("path", c.DefaultQuery("path", ""))
	if path != "" {
		// 递归查找/创建目录，每一级都用 FindOrCreateFolder，保证 parent_id 递归正确
		parts := strings.Split(strings.Trim(path, "/"), "/")
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
			folder, err := h.fileService.FindOrCreateFolder(uid, parentID, p, fullPath)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "create/find folder failed: " + err.Error()})
				return
			}
			parentID = folder.ID
		}
	} else if pid := c.DefaultPostForm("parent_id", c.DefaultQuery("parent_id", "0")); pid != "0" {
		if id, err := strconv.ParseUint(pid, 10, 32); err == nil {
			parentID = uint(id)
		}
	}

	// 获取上传文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no file uploaded"})
		return
	}

	// 上传文件
	uploadedFile, err := h.fileService.UploadFile(uid, file, parentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, uploadedFile)
}

// BatchUpload 批量上传文件，支持相对路径（来自前端 webkitRelativePath）
func (h *FileHandler) BatchUpload(c *gin.Context) {
	userID, exists := c.Get(middleware.CtxUserID)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	uid := userID.(uint)

	parentID := uint(0)
	if pid := c.DefaultPostForm("parent_id", "0"); pid != "0" {
		if id, err := strconv.ParseUint(pid, 10, 32); err == nil {
			parentID = uint(id)
		}
	}

	// 解析 multipart 表单
	if err := c.Request.ParseMultipartForm(64 << 20); err != nil { // 64MB 内存阈值
		// 即使超出会使用临时文件，不影响整体
	}

	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid multipart form"})
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		// 兼容部分前端字段名为 file
		files = form.File["file"]
	}
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no files uploaded"})
		return
	}

	relPaths := form.Value["relative_paths"]
	// 若前端按每个文件提供 single relative_paths 字段，也尝试读取
	if len(relPaths) == 0 {
		relPaths = form.Value["relative_path"]
	}

	uploaded, err := h.fileService.UploadFilesWithRelativePaths(uid, files, relPaths, parentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"files": uploaded, "count": len(uploaded)})
}

// Download 文件下载
func (h *FileHandler) Download(c *gin.Context) {
	userID, exists := c.Get(middleware.CtxUserID)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	uid := userID.(uint)
	fileIDStr := c.Param("id")
	fileID, err := strconv.ParseUint(fileIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file id"})
		return
	}

	reader, file, err := h.fileService.DownloadFile(uid, uint(fileID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}
	defer reader.Close()

	c.Header("Content-Disposition", "attachment; filename="+file.Name)
	c.Header("Content-Type", file.MimeType)
	c.Header("Content-Length", strconv.FormatInt(file.Size, 10))

	c.DataFromReader(http.StatusOK, file.Size, file.MimeType, reader, nil)
}

// GetInfo 获取文件信息
func (h *FileHandler) GetInfo(c *gin.Context) {
	userID, exists := c.Get(middleware.CtxUserID)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	uid := userID.(uint)
	fileIDStr := c.Param("id")
	fileID, err := strconv.ParseUint(fileIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file id"})
		return
	}

	file, err := h.fileService.GetFile(uid, uint(fileID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	c.JSON(http.StatusOK, file)
}

// List 文件列表
func (h *FileHandler) List(c *gin.Context) {
	userID, exists := c.Get(middleware.CtxUserID)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	uid := userID.(uint)
	parentID := uint(0)
	path := c.Query("path")
	if path != "" {
		dir, err := h.fileService.FindDirByPath(uid, path)
		if err != nil {
			// 路径不存在，返回空列表
			c.JSON(http.StatusOK, gin.H{
				"files": []interface{}{},
				"total": 0,
				"page":  1,
				"limit": 20,
			})
			return
		}
		parentID = dir.ID
	} else if pid := c.DefaultQuery("parent_id", "0"); pid != "0" {
		if id, err := strconv.ParseUint(pid, 10, 32); err == nil {
			parentID = uint(id)
		}
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	files, total, err := h.fileService.ListFiles(uid, parentID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"files": files,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// Delete 删除文件
func (h *FileHandler) Delete(c *gin.Context) {
	userID, exists := c.Get(middleware.CtxUserID)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	uid := userID.(uint)
	fileIDStr := c.Param("id")
	fileID, err := strconv.ParseUint(fileIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file id"})
		return
	}

	if err := h.fileService.DeleteFile(uid, uint(fileID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "file deleted successfully"})
}

// CreateFolder 创建文件夹
func (h *FileHandler) CreateFolder(c *gin.Context) {
	userID, exists := c.Get(middleware.CtxUserID)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	uid := userID.(uint)

	var req struct {
		Name     string `json:"name" binding:"required"`
		ParentID uint   `json:"parent_id"`
		Path     string `json:"path"` // 新增 path 字段
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	parentID := req.ParentID
	if req.Path != "" {
		parts := strings.Split(strings.Trim(req.Path, "/"), "/")
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
			folder, err := h.fileService.FindOrCreateFolder(uid, parentID, p, fullPath)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "create/find parent folder failed: " + err.Error()})
				return
			}
			parentID = folder.ID
		}
	}

	folder, err := h.fileService.CreateFolder(uid, req.Name, parentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, folder)
}
