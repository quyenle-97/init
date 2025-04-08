package main

import (
	"fmt"
	"github.com/Minh2009/pv_soa/cfg"
	"github.com/Minh2009/pv_soa/internal/models"
	"github.com/Minh2009/pv_soa/migrations"
	"github.com/Minh2009/pv_soa/pkgs/rdbms"
	"os"
)

func main() {
	c := cfg.LoadConfig()
	db, err := rdbms.NewDB(c.DBDriver, c.DB, true)
	if err != nil {
		panic(err)
	}
	models.Init(db)

	migration := rdbms.NewMigrationTool(db)
	lists := migrations.MigrationLists()
	var arg string
	if len(os.Args) < 2 {
		arg = "migrate"
	} else {
		arg = os.Args[1]
	}
	switch arg {
	case "migrate":
		migration.Migrate(lists)
		fmt.Printf("Migrate successfully !!! \n")
	case "migrate:rollback":
		migration.MigrateRollback(lists)
		fmt.Printf("Migrate rollback successfully !!! \n")
	case "migrate:reset":
		migration.MigrateReset(lists)
		fmt.Printf("Migrate reset successfully !!! \n")
	}
}
