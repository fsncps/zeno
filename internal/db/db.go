package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/fsncps/zeno/internal/config"
	_ "github.com/go-sql-driver/mysql"
)

func Connect(ctx context.Context) (*sql.DB, error) {
	cfg := config.Load()

	// Short, explicit timeouts to avoid “hangs”
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci&timeout=3s&readTimeout=3s&writeTimeout=3s",
		cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBPort, cfg.DBName,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	// conservative pool settings
	db.SetMaxOpenConns(8)
	db.SetMaxIdleConns(8)
	db.SetConnMaxIdleTime(5 * time.Minute)
	db.SetConnMaxLifetime(30 * time.Minute)

	// enforce connectability now
	cctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if err := db.PingContext(cctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}
