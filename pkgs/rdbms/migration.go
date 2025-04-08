package rdbms

import (
	"context"
	"github.com/Minh2009/pv_soa/pkgs/utils"
	"github.com/uptrace/bun"
	"log"
	"time"
)

type MFile interface {
	Up(db *bun.DB) error
	Down(db *bun.DB) error
	GetStructName() string
}

type Migration struct {
	bun.BaseModel `bun:"table:migrations"`

	ID        int64     `bun:"id,pk,autoincrement"`
	Name      string    `bun:"name,notnull,unique"`
	Version   int       `bun:"version"`
	CreatedAt time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp"`
}

type MTool struct {
	db *bun.DB
}

func NewMigrationTool(db *bun.DB) MTool {
	return MTool{db: db}
}

func (t MTool) init() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	_, err := t.db.NewCreateTable().
		Model((*Migration)(nil)).
		IfNotExists().
		WithForeignKeys().
		Exec(ctx)
	if err != nil {
		log.Print(err)
	}
}

func (t MTool) lastVersion() (version int) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	var migration Migration
	err := t.db.NewSelect().
		Model(&migration).
		Order("id DESC").
		Limit(1).
		Scan(ctx)
	if err != nil {
		return 0
	}
	return migration.Version
}

func (t MTool) migrated() (migrated []string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	err = t.db.NewSelect().
		Model((*Migration)(nil)).
		Column("name").
		Scan(ctx, &migrated)
	if err != nil {
		return nil, err
	}
	return migrated, nil
}

func (t MTool) Migrate(list []MFile) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	t.init()
	lastVersion := t.lastVersion()
	migrated, err := t.migrated()
	if err != nil {
		panic(err)
	}
	for _, m := range list {
		if !utils.Contains(migrated, m.GetStructName()) {
			err = m.Up(t.db)
			if err != nil {
				panic(err)
			} else {
				ms := Migration{
					Version: lastVersion + 1,
					Name:    m.GetStructName(),
				}
				_, err = t.db.NewInsert().Model(&ms).Exec(ctx)
				if err != nil {
					panic(err)
				}
			}
		}
	}
}

func (t MTool) MigrateRollback(list []MFile) {
	ctxSelect, cancelSelect := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancelSelect()
	ctxDelete, cancelDelete := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancelDelete()
	lastVersion := t.lastVersion()
	var msR []string
	_, err := t.db.NewSelect().
		Model((*Migration)(nil)).
		Column("name").
		Where("version = ?", lastVersion).
		Order("id desc").
		Exec(ctxSelect, &msR)
	if err != nil {
		panic(err)
	}
	migrationList := list
	utils.Reverse(migrationList)
	for _, m := range list {
		if utils.Contains(msR, m.GetStructName()) {
			err = m.Down(t.db)
			if err != nil {
				panic(err)
			}
		}
	}
	_, err = t.db.NewDelete().Model((*Migration)(nil)).
		Where("version = ?", lastVersion).
		Exec(ctxDelete)
	if err != nil {
		panic(err)
	}
}

func (t MTool) MigrateReset(list []MFile) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	migrationList := make([]MFile, len(list), len(list))
	copy(migrationList, list)
	utils.Reverse(migrationList)
	for _, m := range migrationList {
		err := m.Down(t.db)
		if err != nil {
			panic(err)
		}
	}
	_, err := t.db.NewDropTable().
		Model((*Migration)(nil)).
		IfExists().
		Cascade().
		Exec(ctx)
	if err != nil {
		panic(err)
	}
	t.init()
	t.Migrate(list)
}
