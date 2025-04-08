package rdbms

import (
	"errors"
	"fmt"
	"github.com/Minh2009/pv_soa/pkgs/utils"
	"github.com/uptrace/bun"
)

type DBDriver string

const (
	DBDriverPostgres DBDriver = "pg"
	DBDriverMysql    DBDriver = "mysql"
)

type DB struct {
	DBDriver string `json:"DB_DRIVER"`
	DBHost   string `json:"DB_HOST"`
	DBPort   string `json:"DB_PORT"`
	DBUser   string `json:"DB_USER"`
	DBPass   string `json:"DB_PASS"`
	DBName   string `json:"DB_NAME"`
}

func NewDB(driver string, cfg interface{}, debug bool) (*bun.DB, error) {
	var c DB
	err := utils.BindStruct[DB](cfg, &c)
	if err != nil {
		return nil, err
	}
	cvd := DBDriver(driver)
	switch cvd {
	case DBDriverPostgres:
		return MakePGConnect(c, debug)
	case DBDriverMysql:
		return MakeMysqlConnect(c, debug)
	default:
		return nil, errors.New(fmt.Sprintf("%v driver not suppored", driver))
	}
}
