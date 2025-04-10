package eventstore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/quyenle-97/init/internal/domain"
	"github.com/uptrace/bun"
)

// PostgresEventStore lưu trữ sự kiện sử dụng PostgreSQL
type PostgresEventStore struct {
	db         *bun.DB
	serializer EventSerializer
}

// NewPostgresEventStore tạo một event store mới sử dụng PostgreSQL
func NewPostgresEventStore(db *bun.DB) *PostgresEventStore {
	return &PostgresEventStore{
		db:         db,
		serializer: &JSONEventSerializer{},
	}
}

// SaveEvents lưu danh sách sự kiện vào cơ sở dữ liệu
func (s *PostgresEventStore) SaveEvents(ctx context.Context, aggregateID string, events []domain.Event) error {
	if len(events) == 0 {
		return nil
	}

	// Sử dụng transaction để đảm bảo tất cả hoặc không có sự kiện nào được lưu
	err := s.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		// Kiểm tra phiên bản hiện tại
		var currentVersion int
		err := tx.NewSelect().
			Table("events").
			ColumnExpr("MAX(version)").
			Where("aggregate_id = ?", aggregateID).
			Scan(ctx, &currentVersion)

		if err != nil {
			// Xử lý trường hợp không có sự kiện nào cho aggregate này
			if err.Error() == "sql: no rows in result set" {
				currentVersion = 0
			} else {
				return fmt.Errorf("lỗi khi kiểm tra phiên bản: %w", err)
			}
		}

		// Lưu từng sự kiện
		for i, event := range events {
			// Tính phiên bản mới
			newVersion := currentVersion + i + 1

			// Serialize sự kiện
			data, err := s.serializer.Serialize(event)
			if err != nil {
				return fmt.Errorf("lỗi khi serialize sự kiện: %w", err)
			}

			// Tạo record
			record := EventRecord{
				ID:          event.GetID(),
				AggregateID: event.GetAggregateID(),
				Type:        event.GetType(),
				Version:     newVersion,
				Data:        data,
				Timestamp:   event.GetTimestamp().Unix(),
			}

			// Lưu vào cơ sở dữ liệu
			_, err = tx.NewInsert().
				Model(&record).
				Exec(ctx)

			if err != nil {
				return fmt.Errorf("lỗi khi lưu sự kiện: %w", err)
			}
		}

		return nil
	})

	return err
}

// GetEvents lấy tất cả sự kiện cho một aggregate
func (s *PostgresEventStore) GetEvents(ctx context.Context, aggregateID string) ([]domain.Event, error) {
	var records []EventRecord

	err := s.db.NewSelect().
		Table("events").
		Where("aggregate_id = ?", aggregateID).
		Order("version ASC").
		Scan(ctx, &records)

	if err != nil {
		return nil, fmt.Errorf("lỗi khi truy vấn sự kiện: %w", err)
	}

	events := make([]domain.Event, len(records))
	for i, record := range records {
		event, err := s.serializer.Deserialize(record.Type, record.Data)
		if err != nil {
			return nil, fmt.Errorf("lỗi khi deserialize sự kiện: %w", err)
		}
		events[i] = event
	}

	return events, nil
}

// GetEventsByType lấy tất cả sự kiện của một loại cụ thể
func (s *PostgresEventStore) GetEventsByType(ctx context.Context, eventType domain.EventType) ([]domain.Event, error) {
	var records []EventRecord

	err := s.db.NewSelect().
		Table("events").
		Where("type = ?", eventType).
		Order("timestamp ASC").
		Scan(ctx, &records)

	if err != nil {
		return nil, fmt.Errorf("lỗi khi truy vấn sự kiện theo loại: %w", err)
	}

	events := make([]domain.Event, len(records))
	for i, record := range records {
		event, err := s.serializer.Deserialize(record.Type, record.Data)
		if err != nil {
			return nil, fmt.Errorf("lỗi khi deserialize sự kiện: %w", err)
		}
		events[i] = event
	}

	return events, nil
}

// GetAllEvents lấy tất cả sự kiện với phân trang
func (s *PostgresEventStore) GetAllEvents(ctx context.Context, offset, limit int) ([]domain.Event, error) {
	var records []EventRecord

	query := s.db.NewSelect().
		Table("events").
		Order("timestamp ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Scan(ctx, &records)
	if err != nil {
		return nil, fmt.Errorf("lỗi khi truy vấn tất cả sự kiện: %w", err)
	}

	events := make([]domain.Event, len(records))
	for i, record := range records {
		event, err := s.serializer.Deserialize(record.Type, record.Data)
		if err != nil {
			return nil, fmt.Errorf("lỗi khi deserialize sự kiện: %w", err)
		}
		events[i] = event
	}

	return events, nil
}

// GetEventStream trả về kênh lắng nghe sự kiện mới
// Lưu ý: Đây là một cách đơn giản hóa, trong thực tế sẽ phức tạp hơn
func (s *PostgresEventStore) GetEventStream(ctx context.Context) (<-chan domain.Event, error) {
	eventChan := make(chan domain.Event)

	// Trong một hệ thống thực, bạn có thể sử dụng LISTEN/NOTIFY của PostgreSQL
	// hoặc một hệ thống message broker như Kafka

	go func() {
		defer close(eventChan)

		lastTimestamp := time.Now().Unix()
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// Kiểm tra sự kiện mới mỗi giây
				var records []EventRecord
				err := s.db.NewSelect().
					Table("events").
					Where("timestamp > ?", lastTimestamp).
					Order("timestamp ASC").
					Scan(ctx, &records)

				if err != nil {
					// Log lỗi và tiếp tục
					fmt.Printf("lỗi khi kiểm tra sự kiện mới: %v\n", err)
					continue
				}

				for _, record := range records {
					event, err := s.serializer.Deserialize(record.Type, record.Data)
					if err != nil {
						fmt.Printf("lỗi khi deserialize sự kiện: %v\n", err)
						continue
					}

					// Gửi sự kiện qua kênh
					select {
					case eventChan <- event:
						lastTimestamp = record.Timestamp
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()

	return eventChan, nil
}

// JSONEventSerializer serializer sử dụng JSON
type JSONEventSerializer struct{}

// Serialize một sự kiện thành JSON
func (s *JSONEventSerializer) Serialize(event domain.Event) ([]byte, error) {
	return json.Marshal(event)
}

// Deserialize JSON thành sự kiện
func (s *JSONEventSerializer) Deserialize(eventType domain.EventType, data []byte) (domain.Event, error) {
	var event domain.Event

	switch eventType {
	case domain.OrderCreatedType:
		var e domain.OrderCreatedEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, err
		}
		event = e
	case domain.OrderStatusUpdatedType:
		var e domain.OrderStatusUpdatedEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, err
		}
		event = e
	case domain.OrderCancelledType:
		var e domain.OrderCancelledEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, err
		}
		event = e
	case domain.OrderNoteAddedType:
		var e domain.OrderNoteAddedEvent
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, err
		}
		event = e
	default:
		return nil, errors.New("loại sự kiện không được hỗ trợ")
	}

	return event, nil
}
