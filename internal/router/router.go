package router

import (
	"strconv"
	"strings"

	"online-disk-server/internal/auth"
	"online-disk-server/internal/config"
	"online-disk-server/internal/database"
	"online-disk-server/internal/handler"
	"online-disk-server/internal/middleware"
	"online-disk-server/internal/model"
	"online-disk-server/internal/service"
	"online-disk-server/internal/storage"

	"github.com/gin-gonic/gin"
)

func Register(r *gin.Engine) {
	// Health check & root
	r.GET("/health", handler.Health)
	r.GET("/", handler.Root)

	// Docs: serve swagger-ui and openapi.yaml
	r.GET("/docs", handler.Docs)
	r.GET("/docs/", handler.Docs)
	r.StaticFile("/docs/openapi.yaml", "api/openapi.yaml")

	// Init DB & migrate
	cfg := config.LoadFromEnv()
	db, err := database.Init(cfg)
	if err == nil {
		_ = db.AutoMigrate(&model.User{}, &model.File{})
	}

	// Init storage
	var stor storage.Storage
	if strings.ToLower(cfg.StorageDriver) == "s3" {
		stor, _ = storage.NewS3Storage(
			strings.TrimPrefix(cfg.S3Endpoint, "http://"),
			cfg.S3AccessKey,
			cfg.S3SecretKey,
			cfg.S3Bucket,
			cfg.S3UseSSL == "true",
		)
	} else {
		stor = storage.NewLocalStorage(cfg.LocalStoragePath)
	}

	// JWT manager and handlers
	exp, _ := strconv.Atoi(cfg.JWTExpireHours)
	jwtm := auth.NewJWTManager(cfg.JWTSecret, exp)
	authHandler := handler.NewAuthHandler(db, jwtm)

	// File service and handler
	fileService := service.NewFileService(db, stor)
	fileHandler := handler.NewFileHandler(fileService)

	// v1 api
	v1 := r.Group("/v1")
	{
		// public auth
		v1.POST("/auth/register", authHandler.Register)
		v1.POST("/auth/login", authHandler.Login)

		// protected routes
		v1auth := v1.Group("")
		v1auth.Use(middleware.AuthRequired(jwtm))
		{
			v1auth.GET("/me", authHandler.Me)

			// file management
			v1auth.POST("/files/upload", fileHandler.Upload)
			v1auth.GET("/files", fileHandler.List)
			v1auth.GET("/files/:id", fileHandler.GetInfo)
			v1auth.GET("/files/:id/download", fileHandler.Download)
			v1auth.DELETE("/files/:id", fileHandler.Delete)

			// folder management
			v1auth.POST("/folders", fileHandler.CreateFolder)
		}
	}
}
