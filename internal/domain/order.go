package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// OrderStatus định nghĩa trạng thái của đơn hàng logistics
type OrderStatus string

const (
	OrderStatusCreated        OrderStatus = "CREATED"
	OrderStatusProcessing     OrderStatus = "PROCESSING"
	OrderStatusInTransit      OrderStatus = "IN_TRANSIT"
	OrderStatusOutForDelivery OrderStatus = "OUT_FOR_DELIVERY"
	OrderStatusDelivered      OrderStatus = "DELIVERED"
	OrderStatusException      OrderStatus = "EXCEPTION"
	OrderStatusCancelled      OrderStatus = "CANCELLED"
)

// Location đại diện cho vị trí địa lý
type Location struct {
	Address   string  `json:"address"`
	City      string  `json:"city"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// Order là aggregate root trong domain model
type Order struct {
	ID              string      `json:"id"`
	CustomerID      string      `json:"customer_id"`
	TrackingNumber  string      `json:"tracking_number"`
	Status          OrderStatus `json:"status"`
	Origin          Location    `json:"origin"`
	Destination     Location    `json:"destination"`
	CurrentLocation *Location   `json:"current_location,omitempty"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
	Items           []OrderItem `json:"items"`
	Notes           []string    `json:"notes"`
	Events          []Event     `json:"-"` // Events không được serialize
}

// OrderItem đại diện cho một mục trong đơn hàng
type OrderItem struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Quantity    int     `json:"quantity"`
	Weight      float64 `json:"weight"`
	Price       float64 `json:"price"`
}

// NewOrder tạo đơn hàng mới
func NewOrder(customerID string, origin, destination Location, items []OrderItem) (*Order, error) {
	if customerID == "" {
		return nil, errors.New("customer ID không được để trống")
	}

	if len(items) == 0 {
		return nil, errors.New("đơn hàng phải có ít nhất một mục")
	}

	now := time.Now()
	orderID := uuid.New().String()
	trackingNumber := generateTrackingNumber()

	order := &Order{
		ID:             orderID,
		CustomerID:     customerID,
		TrackingNumber: trackingNumber,
		Status:         OrderStatusCreated,
		Origin:         origin,
		Destination:    destination,
		CreatedAt:      now,
		UpdatedAt:      now,
		Items:          items,
		Notes:          []string{},
	}

	// Tạo event OrderCreated
	event := NewOrderCreatedEvent(orderID, customerID, trackingNumber, origin, destination, items)
	order.Events = append(order.Events, event)

	return order, nil
}

// UpdateStatus cập nhật trạng thái đơn hàng
func (o *Order) UpdateStatus(newStatus OrderStatus, location *Location, note string) error {
	if o.Status == OrderStatusDelivered || o.Status == OrderStatusCancelled {
		return errors.New("không thể cập nhật trạng thái cho đơn hàng đã hoàn thành hoặc đã hủy")
	}

	oldStatus := o.Status
	o.Status = newStatus
	o.UpdatedAt = time.Now()

	if location != nil {
		o.CurrentLocation = location
	}

	if note != "" {
		o.Notes = append(o.Notes, note)
	}

	// Tạo event OrderStatusUpdated
	event := NewOrderStatusUpdatedEvent(o.ID, oldStatus, newStatus, location, note)
	o.Events = append(o.Events, event)

	return nil
}

// CancelOrder hủy đơn hàng
func (o *Order) CancelOrder(reason string) error {
	if o.Status == OrderStatusDelivered {
		return errors.New("không thể hủy đơn hàng đã giao")
	}

	if o.Status == OrderStatusCancelled {
		return errors.New("đơn hàng đã bị hủy trước đó")
	}

	oldStatus := o.Status
	o.Status = OrderStatusCancelled
	o.UpdatedAt = time.Now()

	if reason != "" {
		o.Notes = append(o.Notes, reason)
	}

	// Tạo event OrderCancelled
	event := NewOrderCancelledEvent(o.ID, oldStatus, reason)
	o.Events = append(o.Events, event)

	return nil
}

// AddNote thêm ghi chú vào đơn hàng
func (o *Order) AddNote(note string) error {
	if note == "" {
		return errors.New("ghi chú không được để trống")
	}

	o.Notes = append(o.Notes, note)
	o.UpdatedAt = time.Now()

	// Tạo event OrderNoteAdded
	event := NewOrderNoteAddedEvent(o.ID, note)
	o.Events = append(o.Events, event)

	return nil
}

// GetUncommittedEvents trả về các sự kiện chưa được commit
func (o *Order) GetUncommittedEvents() []Event {
	return o.Events
}

// ClearUncommittedEvents xóa các sự kiện chưa được commit
func (o *Order) ClearUncommittedEvents() {
	o.Events = []Event{}
}

// generateTrackingNumber tạo số theo dõi đơn hàng
func generateTrackingNumber() string {
	// Trong thực tế, bạn sẽ muốn tạo một số theo dõi duy nhất và có ý nghĩa
	// Đây là một cách đơn giản để minh họa
	return "TRK-" + uuid.New().String()[0:8]
}
