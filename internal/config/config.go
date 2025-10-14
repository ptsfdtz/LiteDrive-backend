package config

import (
    "os"
)

type Config struct {
    AppName   string
    GinMode   string
    HTTPAddr  string

    DatabaseDriver string
    DatabaseDSN    string

    StorageDriver    string
    LocalStoragePath string
    S3Endpoint       string
    S3Bucket         string
    S3AccessKey      string
    S3SecretKey      string
    S3UseSSL         string

    JWTSecret      string
    JWTExpireHours string
}

func getenv(key, def string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return def
}

func LoadFromEnv() *Config {
    return &Config{
        AppName:         getenv("APP_NAME", "online-disk-server"),
        GinMode:         getenv("GIN_MODE", "debug"),
        HTTPAddr:        getenv("HTTP_ADDR", "0.0.0.0:8080"),
        DatabaseDriver:  getenv("DATABASE_DRIVER", "sqlite"),
        DatabaseDSN:     getenv("DATABASE_DSN", "file:./data/online_disk.db?cache=shared&_foreign_keys=on"),
        StorageDriver:   getenv("STORAGE_DRIVER", "local"),
        LocalStoragePath:getenv("LOCAL_STORAGE_PATH", "./data/files"),
        S3Endpoint:      getenv("S3_ENDPOINT", ""),
        S3Bucket:        getenv("S3_BUCKET", ""),
        S3AccessKey:     getenv("S3_ACCESS_KEY", ""),
        S3SecretKey:     getenv("S3_SECRET_KEY", ""),
        S3UseSSL:        getenv("S3_USE_SSL", "false"),
        JWTSecret:       getenv("JWT_SECRET", "please_change_me"),
        JWTExpireHours:  getenv("JWT_EXPIRE_HOURS", "72"),
    }
}
