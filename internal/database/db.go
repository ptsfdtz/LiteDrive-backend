package database

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"online-disk-server/internal/config"

	sqlite "github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

func Init(cfg *config.Config) (*gorm.DB, error) {
	if db != nil {
		return db, nil
	}

	var (
		d   *gorm.DB
		err error
	)

	gcfg := &gorm.Config{Logger: logger.Default.LogMode(logger.Warn)}

	switch cfg.DatabaseDriver {
	case "sqlite":
		// ensure sqlite db directory exists if DSN points to a file path
		ensureSqliteDir(cfg.DatabaseDSN)
		d, err = gorm.Open(sqlite.Open(cfg.DatabaseDSN), gcfg)
	case "postgres", "postgresql":
		d, err = gorm.Open(postgres.Open(cfg.DatabaseDSN), gcfg)
	case "mysql":
		d, err = gorm.Open(mysql.Open(cfg.DatabaseDSN), gcfg)
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", cfg.DatabaseDriver)
	}
	if err != nil {
		return nil, err
	}

	db = d
	return db, nil
}

func DB() *gorm.DB { return db }

func ensureSqliteDir(dsn string) {
	// common forms: file:./data/online_disk.db?params or ./data/online_disk.db
	path := dsn
	if after, ok := strings.CutPrefix(path, "file:"); ok {
		path = after
	}
	if i := strings.Index(path, "?"); i >= 0 {
		path = path[:i]
	}
	if path == "" {
		return
	}
	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		_ = os.MkdirAll(dir, 0o755)
	}
}
