package handler

import (
	"net/http"
	"strconv"

	"online-disk-server/internal/auth"
	"online-disk-server/internal/middleware"
	"online-disk-server/internal/model"
	"online-disk-server/internal/repository"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AuthHandler struct {
	db    *gorm.DB
	users *repository.UserRepository
	jwtm  *auth.JWTManager
}

func NewAuthHandler(db *gorm.DB, jwtm *auth.JWTManager) *AuthHandler {
	return &AuthHandler{db: db, users: repository.NewUserRepository(db), jwtm: jwtm}
}

type registerReq struct {
	Username string `json:"username" binding:"required,min=3,max=64"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6,max=64"`
	Nickname string `json:"nickname"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req registerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	hashed, err := auth.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "hash password failed"})
		return
	}
	u := &model.User{Username: req.Username, Email: req.Email, Password: hashed, Nickname: req.Nickname}
	if err := h.users.Create(u); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": u.ID, "username": u.Username, "email": u.Email, "nickname": u.Nickname})
}

type loginReq struct {
	Username string `json:"username" binding:"required_without=Email"`
	Email    string `json:"email" binding:"required_without=Username"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var u *model.User
	var err error
	if req.Username != "" {
		u, err = h.users.FindByUsername(req.Username)
	} else {
		u, err = h.users.FindByEmail(req.Email)
	}
	if err != nil || auth.CheckPassword(u.Password, req.Password) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	token, err := h.jwtm.Generate(u.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "generate token failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (h *AuthHandler) Me(c *gin.Context) {
	v, exists := c.Get(middleware.CtxUserID)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	uid := uint(0)
	switch t := v.(type) {
	case uint:
		uid = t
	case int:
		uid = uint(t)
	case string:
		if n, err := strconv.Atoi(t); err == nil {
			uid = uint(n)
		}
	}
	if uid == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var u model.User
	if err := h.db.First(&u, uid).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": u.ID, "username": u.Username, "email": u.Email, "nickname": u.Nickname})
}
