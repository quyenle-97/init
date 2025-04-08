package migrations

import (
	"context"
	"github.com/Minh2009/pv_soa/internal/models"
	"github.com/uptrace/bun"
	"reflect"
	"time"
)

type CreateProductsTable struct {
	Version int
}

func (m CreateProductsTable) Up(db *bun.DB) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	_, err = db.NewCreateTable().
		Model((*models.Product)(nil)).
		IfNotExists().
		WithForeignKeys().
		Exec(ctx)
	if err != nil {
		return err
	}

	_, err = db.NewCreateTable().
		Model((*models.ProductCategory)(nil)).
		IfNotExists().
		WithForeignKeys().
		Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (m CreateProductsTable) Down(db *bun.DB) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	_, err = db.NewDropTable().
		Model((*models.ProductCategory)(nil)).
		IfExists().
		Cascade().
		Exec(ctx)
	if err != nil {
		return err
	}

	_, err = db.NewDropTable().
		Model((*models.Product)(nil)).
		IfExists().
		Cascade().
		Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (m CreateProductsTable) GetStructName() string {
	if t := reflect.TypeOf(m); t.Kind() == reflect.Ptr {
		return t.Elem().Name()
	} else {
		return t.Name()
	}
}
