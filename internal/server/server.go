package server

import (
    "log"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"

    "online-disk-server/internal/config"
    "online-disk-server/internal/router"
)

func Run() error {
    cfg := config.LoadFromEnv()

    // Set Gin mode
    gin.SetMode(cfg.GinMode)

    r := gin.New()
    r.Use(gin.Recovery())
    r.Use(requestLogger())
    r.Use(cors())

    // Register routes
    router.Register(r)

    srv := &http.Server{
        Addr:              cfg.HTTPAddr,
        Handler:           r,
        ReadHeaderTimeout: 10 * time.Second,
    }

    log.Printf("starting %s at %s (mode=%s)", cfg.AppName, cfg.HTTPAddr, cfg.GinMode)
    return srv.ListenAndServe()
}

// Simple request logger middleware
func requestLogger() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        c.Next()
        latency := time.Since(start)
        log.Printf("%s %s -> %d (%s)", c.Request.Method, c.Request.URL.Path, c.Writer.Status(), latency)
    }
}

// Basic CORS middleware (allow all for dev)
func cors() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
        c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        if c.Request.Method == http.MethodOptions {
            c.AbortWithStatus(http.StatusNoContent)
            return
        }
        c.Next()
    }
}
