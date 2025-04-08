package migrations

import (
	"context"
	"github.com/Minh2009/pv_soa/internal/models"
	"github.com/uptrace/bun"
	"reflect"
	"time"
)

type CreateSupplierTable struct {
	Version int
}

func (m CreateSupplierTable) Up(db *bun.DB) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	_, err = db.NewCreateTable().
		Model((*models.Supplier)(nil)).
		IfNotExists().
		WithForeignKeys().
		Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (m CreateSupplierTable) Down(db *bun.DB) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	_, err = db.NewDropTable().
		Model((*models.Supplier)(nil)).
		IfExists().
		Cascade().
		Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (m CreateSupplierTable) GetStructName() string {
	if t := reflect.TypeOf(m); t.Kind() == reflect.Ptr {
		return t.Elem().Name()
	} else {
		return t.Name()
	}
}
