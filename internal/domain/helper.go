package domain

import "fmt"

// RebuildFromEvents xây dựng lại đơn hàng từ chuỗi các sự kiện
func RebuildFromEvents(listEvents []Event) *Order {
	if len(listEvents) == 0 {
		return nil
	}

	var order *Order
	fmt.Println(122123123)

	for _, event := range listEvents {
		switch e := event.(type) {
		case OrderCreatedEvent:
			// Tạo đơn hàng mới từ sự kiện đầu tiên phải là OrderCreated
			order = &Order{
				ID:             e.AggregateID,
				CustomerID:     e.CustomerID,
				TrackingNumber: e.TrackingNumber,
				Status:         OrderStatusCreated,
				Origin:         e.Origin,
				Destination:    e.Destination,
				Items:          e.Items,
				CreatedAt:      e.Timestamp,
				UpdatedAt:      e.Timestamp,
				Notes:          []string{},
			}
		case OrderStatusUpdatedEvent:
			if order == nil {
				continue
			}
			order.Status = e.NewStatus
			order.UpdatedAt = e.Timestamp
			if e.CurrentLocation != nil {
				order.CurrentLocation = e.CurrentLocation
			}
			if e.Note != "" {
				order.Notes = append(order.Notes, e.Note)
			}
		case OrderCancelledEvent:
			if order == nil {
				continue
			}
			order.Status = OrderStatusCancelled
			order.UpdatedAt = e.Timestamp
			if e.Reason != "" {
				order.Notes = append(order.Notes, e.Reason)
			}
		case OrderNoteAddedEvent:
			if order == nil {
				continue
			}
			order.Notes = append(order.Notes, e.Note)
			order.UpdatedAt = e.Timestamp
		}
	}

	return order
}
