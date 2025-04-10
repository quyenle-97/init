package domain

import (
	"github.com/google/uuid"
	"time"
)

// OrderCreatedEvent là sự kiện khi đơn hàng được tạo
type OrderCreatedEvent struct {
	BaseEvent
	CustomerID     string      `json:"customer_id"`
	TrackingNumber string      `json:"tracking_number"`
	Origin         Location    `json:"origin"`
	Destination    Location    `json:"destination"`
	Items          []OrderItem `json:"items"`
}

// NewOrderCreatedEvent tạo một OrderCreatedEvent mới
func NewOrderCreatedEvent(orderID, customerID, trackingNumber string, origin, destination Location, items []OrderItem) OrderCreatedEvent {
	return OrderCreatedEvent{
		BaseEvent: BaseEvent{
			ID:          uuid.New().String(),
			AggregateID: orderID,
			Type:        OrderCreatedType,
			Timestamp:   time.Now(),
			Version:     1,
		},
		CustomerID:     customerID,
		TrackingNumber: trackingNumber,
		Origin:         origin,
		Destination:    destination,
		Items:          items,
	}
}

// OrderStatusUpdatedEvent là sự kiện khi trạng thái đơn hàng thay đổi
type OrderStatusUpdatedEvent struct {
	BaseEvent
	OldStatus       OrderStatus `json:"old_status"`
	NewStatus       OrderStatus `json:"new_status"`
	CurrentLocation *Location   `json:"current_location,omitempty"`
	Note            string      `json:"note,omitempty"`
}

// NewOrderStatusUpdatedEvent tạo một OrderStatusUpdatedEvent mới
func NewOrderStatusUpdatedEvent(orderID string, oldStatus, newStatus OrderStatus, location *Location, note string) OrderStatusUpdatedEvent {
	return OrderStatusUpdatedEvent{
		BaseEvent: BaseEvent{
			ID:          uuid.New().String(),
			AggregateID: orderID,
			Type:        OrderStatusUpdatedType,
			Timestamp:   time.Now(),
			Version:     1, // Trong thực tế, phiên bản nên tăng dần
		},
		OldStatus:       oldStatus,
		NewStatus:       newStatus,
		CurrentLocation: location,
		Note:            note,
	}
}

// OrderCancelledEvent là sự kiện khi đơn hàng bị hủy
type OrderCancelledEvent struct {
	BaseEvent
	PreviousStatus OrderStatus `json:"previous_status"`
	Reason         string      `json:"reason,omitempty"`
}

// NewOrderCancelledEvent tạo một OrderCancelledEvent mới
func NewOrderCancelledEvent(orderID string, previousStatus OrderStatus, reason string) OrderCancelledEvent {
	return OrderCancelledEvent{
		BaseEvent: BaseEvent{
			ID:          uuid.New().String(),
			AggregateID: orderID,
			Type:        OrderCancelledType,
			Timestamp:   time.Now(),
			Version:     1, // Trong thực tế, phiên bản nên tăng dần
		},
		PreviousStatus: previousStatus,
		Reason:         reason,
	}
}

// OrderNoteAddedEvent là sự kiện khi ghi chú được thêm vào đơn hàng
type OrderNoteAddedEvent struct {
	BaseEvent
	Note string `json:"note"`
}

// NewOrderNoteAddedEvent tạo một OrderNoteAddedEvent mới
func NewOrderNoteAddedEvent(orderID string, note string) OrderNoteAddedEvent {
	return OrderNoteAddedEvent{
		BaseEvent: BaseEvent{
			ID:          uuid.New().String(),
			AggregateID: orderID,
			Type:        OrderNoteAddedType,
			Timestamp:   time.Now(),
			Version:     1, // Trong thực tế, phiên bản nên tăng dần
		},
		Note: note,
	}
}
