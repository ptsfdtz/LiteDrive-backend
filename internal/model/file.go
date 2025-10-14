package model

import (
	"time"
)

type File struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 基本信息
	Name     string `gorm:"size:255;not null" json:"name"`
	Path     string `gorm:"size:500;not null" json:"path"`
	Size     int64  `json:"size"`
	MimeType string `gorm:"size:100" json:"mime_type"`
	Hash     string `gorm:"size:64;index" json:"hash"` // MD5 or SHA256

	// 关联
	UserID   uint `gorm:"not null;index" json:"user_id"`
	ParentID uint `gorm:"index" json:"parent_id"` // 0 表示根目录

	// 存储信息
	StoragePath string `gorm:"size:500" json:"storage_path"` // S3 key or local path

	// 状态
	IsDir    bool `gorm:"default:false" json:"is_dir"`
	IsPublic bool `gorm:"default:false" json:"is_public"`
}
