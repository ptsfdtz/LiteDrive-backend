package router

import (
	"online-disk-server/internal/handler"

	"github.com/gin-gonic/gin"
)

func Register(r *gin.Engine) {
	// Health check & root
	r.GET("/health", handler.Health)
	r.GET("/", handler.Root)

	// v1 api group placeholder
	v1 := r.Group("/v1")
	{
		// placeholder route to avoid unused variable warning
		v1.GET("/ping", func(c *gin.Context) { c.JSON(200, gin.H{"message": "pong"}) })
	}
}
