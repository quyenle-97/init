package main

import (
	"fmt"
	"github.com/quyenle-97/init/cfg"
	"github.com/quyenle-97/init/internal/models"
	"github.com/quyenle-97/init/migrations"
	"github.com/quyenle-97/init/pkgs/rdbms"
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
