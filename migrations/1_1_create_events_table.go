package migrations

import (
	"context"
	"reflect"
	"time"

	"github.com/uptrace/bun"
)

// EventsTable định nghĩa bảng lưu trữ các sự kiện
type EventsTable struct {
	Version int
}

type EventModel struct {
	bun.BaseModel `bun:"table:events,alias:e"`

	ID          string    `bun:"id,pk"`
	AggregateID string    `bun:"aggregate_id,notnull"`
	Type        string    `bun:"type,notnull"`
	Version     int       `bun:"version,notnull"`
	Data        []byte    `bun:"data,notnull"`
	Metadata    []byte    `bun:"metadata"`
	Timestamp   int64     `bun:"timestamp,notnull"`
	CreatedAt   time.Time `bun:"created_at,notnull,default:current_timestamp"`
}

func (m EventsTable) Up(db *bun.DB) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// Tạo bảng events
	_, err = db.NewCreateTable().
		Model((*EventModel)(nil)).
		IfNotExists().
		Exec(ctx)
	if err != nil {
		return err
	}

	// Tạo index cho tìm kiếm theo aggregate_id
	_, err = db.NewCreateIndex().
		Model((*EventModel)(nil)).
		Index("idx_events_aggregate_id").
		Column("aggregate_id").
		IfNotExists().
		Exec(ctx)
	if err != nil {
		return err
	}

	// Tạo index cho tìm kiếm theo type
	_, err = db.NewCreateIndex().
		Model((*EventModel)(nil)).
		Index("idx_events_type").
		Column("type").
		IfNotExists().
		Exec(ctx)
	if err != nil {
		return err
	}

	// Tạo index cho tìm kiếm theo timestamp
	_, err = db.NewCreateIndex().
		Model((*EventModel)(nil)).
		Index("idx_events_timestamp").
		Column("timestamp").
		IfNotExists().
		Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (m EventsTable) Down(db *bun.DB) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// Xóa các index
	_, err = db.NewDropIndex().
		Model((*EventModel)(nil)).
		Index("idx_events_aggregate_id").
		IfExists().
		Exec(ctx)
	if err != nil {
		return err
	}

	_, err = db.NewDropIndex().
		Model((*EventModel)(nil)).
		Index("idx_events_type").
		IfExists().
		Exec(ctx)
	if err != nil {
		return err
	}

	_, err = db.NewDropIndex().
		Model((*EventModel)(nil)).
		Index("idx_events_timestamp").
		IfExists().
		Exec(ctx)
	if err != nil {
		return err
	}

	// Xóa bảng events
	_, err = db.NewDropTable().
		Model((*EventModel)(nil)).
		IfExists().
		Cascade().
		Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (m EventsTable) GetStructName() string {
	if t := reflect.TypeOf(m); t.Kind() == reflect.Ptr {
		return t.Elem().Name()
	} else {
		return t.Name()
	}
}
