package rdbms

import (
	"database/sql"
	"fmt"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
	"strconv"
	"time"
)

func MakePGConnect(cfg DB, debug bool) (*bun.DB, error) {
	port, err := strconv.Atoi(cfg.DBPort)
	if err != nil {
		panic(err)
	}
	opts := []pgdriver.Option{
		pgdriver.WithAddr(fmt.Sprintf("%s:%d", cfg.DBHost, port)),
		pgdriver.WithUser(cfg.DBUser),
		pgdriver.WithPassword(cfg.DBPass),
		pgdriver.WithDatabase(cfg.DBName),
		pgdriver.WithInsecure(true),
		pgdriver.WithTimeout(15 * time.Second),
		pgdriver.WithDialTimeout(15 * time.Second),
		pgdriver.WithReadTimeout(15 * time.Second),
		pgdriver.WithWriteTimeout(15 * time.Second),
	}

	pgConn := pgdriver.NewConnector(opts...)
	sqlPG := sql.OpenDB(pgConn)
	database := bun.NewDB(sqlPG, pgdialect.New())
	if debug {
		database.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))
	}
	if err = database.Ping(); err != nil {
		return nil, err
	}
	return database, nil
}
