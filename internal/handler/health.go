package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func Root(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"name": "online-disk-server", "status": "running"})
}
