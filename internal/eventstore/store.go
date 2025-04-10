package eventstore

import (
	"context"
	"github.com/uptrace/bun"
	"time"

	"github.com/quyenle-97/init/internal/domain"
)

// EventStore định nghĩa interface cho lưu trữ và truy vấn các sự kiện
type EventStore interface {
	// SaveEvents lưu các sự kiện mới cho một aggregate
	SaveEvents(ctx context.Context, aggregateID string, events []domain.Event) error

	// GetEvents lấy tất cả các sự kiện cho một aggregate
	GetEvents(ctx context.Context, aggregateID string) ([]domain.Event, error)

	// GetEventsByType lấy tất cả các sự kiện của một loại cụ thể
	GetEventsByType(ctx context.Context, eventType domain.EventType) ([]domain.Event, error)

	// GetAllEvents lấy tất cả các sự kiện trong hệ thống, có thể phân trang
	GetAllEvents(ctx context.Context, offset, limit int) ([]domain.Event, error)

	// GetEventStream trả về một kênh để lắng nghe các sự kiện mới
	GetEventStream(ctx context.Context) (<-chan domain.Event, error)
}

// EventSerializer interface để serialize và deserialize các sự kiện
type EventSerializer interface {
	// Serialize chuyển đổi một sự kiện thành dữ liệu nhị phân
	Serialize(event domain.Event) ([]byte, error)

	// Deserialize chuyển đổi dữ liệu nhị phân thành sự kiện
	Deserialize(eventType domain.EventType, data []byte) (domain.Event, error)
}

// EventRecord đại diện cho một bản ghi sự kiện trong cơ sở dữ liệu
type EventRecord struct {
	bun.BaseModel `bun:"table:events,alias:e"`

	ID          string           `bun:"id,pk"`
	AggregateID string           `bun:"aggregate_id,notnull"`
	Type        domain.EventType `json:"type"`
	Version     int              `bun:"version,notnull"`
	Data        []byte           `bun:"data,notnull"`
	Metadata    []byte           `bun:"metadata"`
	Timestamp   int64            `bun:"timestamp,notnull"`
	CreatedAt   time.Time        `bun:"created_at,notnull,default:current_timestamp"`
}
