package migrations

import "github.com/Minh2009/pv_soa/pkgs/rdbms"

func MigrationLists() []rdbms.MFile {
	return []rdbms.MFile{
		CreateCategoriesTable{},
		CreateCitiesTable{},
		CreateSupplierTable{},
		CreateProductsTable{},
	}
}
