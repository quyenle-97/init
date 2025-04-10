package domain

import (
	"time"
)

// EventType định nghĩa loại sự kiện
type EventType string

const (
	OrderCreatedType       EventType = "ORDER_CREATED"
	OrderStatusUpdatedType EventType = "ORDER_STATUS_UPDATED"
	OrderCancelledType     EventType = "ORDER_CANCELLED"
	OrderNoteAddedType     EventType = "ORDER_NOTE_ADDED"
)

// Event là interface cho tất cả các sự kiện domain
type Event interface {
	GetID() string
	GetAggregateID() string
	GetType() EventType
	GetTimestamp() time.Time
	GetVersion() int
}

// BaseEvent chứa các trường chung cho tất cả các sự kiện
type BaseEvent struct {
	ID          string    `json:"id"`
	AggregateID string    `json:"aggregate_id"`
	Type        EventType `json:"type"`
	Timestamp   time.Time `json:"timestamp"`
	Version     int       `json:"version"`
}

// GetID trả về ID của sự kiện
func (e BaseEvent) GetID() string {
	return e.ID
}

// GetAggregateID trả về ID của aggregate
func (e BaseEvent) GetAggregateID() string {
	return e.AggregateID
}

// GetType trả về loại sự kiện
func (e BaseEvent) GetType() EventType {
	return e.Type
}

// GetTimestamp trả về thời gian của sự kiện
func (e BaseEvent) GetTimestamp() time.Time {
	return e.Timestamp
}

// GetVersion trả về phiên bản của sự kiện
func (e BaseEvent) GetVersion() int {
	return e.Version
}
