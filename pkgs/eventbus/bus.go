package eventbus

import (
	"github.com/quyenle-97/init/internal/domain"
)

// EventHandler là interface cho các handler xử lý sự kiện
type EventHandler interface {
	// HandleEvent xử lý một sự kiện
	HandleEvent(event domain.Event) error
}

// EventBus là interface cho việc phát và đăng ký xử lý sự kiện
type EventBus interface {
	// Publish phát một sự kiện tới tất cả các handler đã đăng ký
	Publish(event domain.Event) error

	// Subscribe đăng ký một handler cho một hoặc nhiều loại sự kiện
	Subscribe(handler EventHandler, eventTypes ...domain.EventType) error

	// Unsubscribe hủy đăng ký một handler
	Unsubscribe(handler EventHandler, eventTypes ...domain.EventType) error
}

// InMemoryEventBus là triển khai đơn giản của EventBus trong bộ nhớ
type InMemoryEventBus struct {
	handlers map[domain.EventType][]EventHandler
}

// NewInMemoryEventBus tạo một event bus mới trong bộ nhớ
func NewInMemoryEventBus() *InMemoryEventBus {
	return &InMemoryEventBus{
		handlers: make(map[domain.EventType][]EventHandler),
	}
}

// Publish phát một sự kiện tới tất cả các handler đã đăng ký
func (b *InMemoryEventBus) Publish(event domain.Event) error {
	eventType := event.GetType()
	handlers := b.handlers[eventType]
	for _, handler := range handlers {
		if err := handler.HandleEvent(event); err != nil {
			// Trong triển khai thực tế, chúng ta sẽ muốn log lỗi và tiếp tục
			// thay vì dừng lại hoàn toàn
			return err
		}
	}

	return nil
}

// Subscribe đăng ký một handler cho một hoặc nhiều loại sự kiện
func (b *InMemoryEventBus) Subscribe(handler EventHandler, eventTypes ...domain.EventType) error {
	// Nếu không có loại sự kiện nào được chỉ định, đăng ký cho tất cả các loại
	if len(eventTypes) == 0 {
		eventTypes = []domain.EventType{
			domain.OrderCreatedType,
			domain.OrderStatusUpdatedType,
			domain.OrderCancelledType,
			domain.OrderNoteAddedType,
		}
	}

	for _, eventType := range eventTypes {
		if b.handlers[eventType] == nil {
			b.handlers[eventType] = []EventHandler{}
		}
		b.handlers[eventType] = append(b.handlers[eventType], handler)
	}

	return nil
}

// Unsubscribe hủy đăng ký một handler
func (b *InMemoryEventBus) Unsubscribe(handler EventHandler, eventTypes ...domain.EventType) error {
	// Nếu không có loại sự kiện nào được chỉ định, hủy đăng ký khỏi tất cả các loại
	if len(eventTypes) == 0 {
		eventTypes = []domain.EventType{
			domain.OrderCreatedType,
			domain.OrderStatusUpdatedType,
			domain.OrderCancelledType,
			domain.OrderNoteAddedType,
		}
	}

	for _, eventType := range eventTypes {
		if handlers, ok := b.handlers[eventType]; ok {
			updatedHandlers := []EventHandler{}
			for _, h := range handlers {
				// So sánh con trỏ của handler, chỉ giữ lại các handler khác
				if h != handler {
					updatedHandlers = append(updatedHandlers, h)
				}
			}
			b.handlers[eventType] = updatedHandlers
		}
	}

	return nil
}
