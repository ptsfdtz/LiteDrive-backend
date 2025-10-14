package handler

import (
	"net/http"
	"strconv"

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

	// 获取父目录ID（可选）
	parentID := uint(0)
	if pid := c.DefaultPostForm("parent_id", "0"); pid != "0" {
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
	if pid := c.DefaultQuery("parent_id", "0"); pid != "0" {
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
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	folder, err := h.fileService.CreateFolder(uid, req.Name, req.ParentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, folder)
}
