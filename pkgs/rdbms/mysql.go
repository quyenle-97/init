package rdbms

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/mysqldialect"
	"github.com/uptrace/bun/extra/bundebug"
	"strconv"
	"time"
)

func buildDSN(cfg DB) string {
	params := fmt.Sprintf("timeout=%s&readTimeout=%s&writeTimeout=%s",
		15*time.Second,
		15*time.Second,
		15*time.Second,
	)
	port, err := strconv.Atoi(cfg.DBPort)
	if err != nil {
		panic(err)
	}
	params += "&tls=true"
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s",
		cfg.DBUser, cfg.DBPass, cfg.DBHost, port, cfg.DBName, params)
}

func MakeMysqlConnect(cfg DB, debug bool) (*bun.DB, error) {
	dsn := buildDSN(cfg)
	sqlDB, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db := bun.NewDB(sqlDB, mysqldialect.New())
	if debug {
		db.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}
