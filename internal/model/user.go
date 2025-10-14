package model

import (
	"time"
)

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Username string `gorm:"uniqueIndex;size:64" json:"username"`
	Email    string `gorm:"uniqueIndex;size:128" json:"email"`
	Password string `json:"-"` // hashed
	Nickname string `gorm:"size:64" json:"nickname"`
}
