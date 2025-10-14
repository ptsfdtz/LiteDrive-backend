package router

import (
	"strconv"

	"online-disk-server/internal/auth"
	"online-disk-server/internal/config"
	"online-disk-server/internal/database"
	"online-disk-server/internal/handler"
	"online-disk-server/internal/middleware"
	"online-disk-server/internal/model"

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
		_ = db.AutoMigrate(&model.User{})
	}

	// JWT manager and handler
	exp, _ := strconv.Atoi(cfg.JWTExpireHours)
	jwtm := auth.NewJWTManager(cfg.JWTSecret, exp)
	authHandler := handler.NewAuthHandler(db, jwtm)

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
		}
	}
}
