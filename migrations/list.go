package migrations

import "github.com/quyenle-97/init/pkgs/rdbms"

func MigrationLists() []rdbms.MFile {
	return []rdbms.MFile{
		EventsTable{},
		ProjectionsTable{},
	}
}
