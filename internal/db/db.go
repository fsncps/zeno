package db

import (
    "database/sql"
    "fmt"

    _ "github.com/go-sql-driver/mysql"
    "github.com/fsncps/zeno/internal/config"
)

func Connect() *sql.DB {
    cfg := config.Load()

    dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4",
        cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBPort, cfg.DBName)
    db, err := sql.Open("mysql", dsn)
    if err != nil {
        panic(err)
    }
    return db
}

