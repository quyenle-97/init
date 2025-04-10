package services

import (
	"context"
	"fmt"
	"github.com/quyenle-97/init/internal/domain"
	"github.com/quyenle-97/init/internal/eventstore"
	"github.com/quyenle-97/init/internal/repository"
	"github.com/quyenle-97/init/pkgs/eventbus"
)

// OrderService định nghĩa các thao tác có thể thực hiện với đơn hàng
type OrderService interface {
	// Command side (write)
	CreateOrder(ctx context.Context, customerID string, origin, destination domain.Location, items []domain.OrderItem) (string, string, error)
	UpdateOrderStatus(ctx context.Context, orderID string, newStatus domain.OrderStatus, location *domain.Location, note string) error
	CancelOrder(ctx context.Context, orderID string, reason string) error
	AddOrderNote(ctx context.Context, orderID string, note string) error

	// Query side (read)
	GetOrder(ctx context.Context, orderID string) (*domain.Order, error)
	GetOrderByTracking(ctx context.Context, trackingNumber string) (*domain.Order, error)
	ListOrders(ctx context.Context, customerID string, status domain.OrderStatus, offset, limit int) ([]*domain.Order, int, error)
	GetOrderHistory(ctx context.Context, orderID string) ([]domain.Event, error)
}

// orderService triển khai OrderService
type orderService struct {
	eventStore eventstore.EventStore
	orderRepo  repository.OrderRepository
	eventBus   eventbus.EventBus
}

// NewOrderService tạo một instance mới của OrderService
func NewOrderService(
	eventStore eventstore.EventStore,
	orderRepo repository.OrderRepository,
	eventBus eventbus.EventBus,
) OrderService {
	return &orderService{
		eventStore: eventStore,
		orderRepo:  orderRepo,
		eventBus:   eventBus,
	}
}

// CreateOrder triển khai phương thức tạo đơn hàng mới
func (s *orderService) CreateOrder(
	ctx context.Context,
	customerID string,
	origin domain.Location,
	destination domain.Location,
	items []domain.OrderItem,
) (string, string, error) {
	// Tạo đơn hàng mới (command handling)
	order, err := domain.NewOrder(customerID, origin, destination, items)
	if err != nil {
		return "", "", fmt.Errorf("không thể tạo đơn hàng: %w", err)
	}

	// Lưu sự kiện vào event store
	events := order.GetUncommittedEvents()
	err = s.eventStore.SaveEvents(ctx, order.ID, events)
	if err != nil {
		return "", "", fmt.Errorf("lỗi khi lưu sự kiện: %w", err)
	}

	// Phát các sự kiện tới event bus để cập nhật read models
	for _, event := range events {
		fmt.Println("events", event.GetType())

		if err := s.eventBus.Publish(event); err != nil {
			// Chỉ log lỗi, không fail request
			fmt.Printf("Lỗi khi phát sự kiện: %v\n", err)
		}
	}

	// Xóa các sự kiện đã xử lý
	order.ClearUncommittedEvents()

	return order.ID, order.TrackingNumber, nil
}

// GetOrder lấy thông tin đơn hàng theo ID
func (s *orderService) GetOrder(ctx context.Context, orderID string) (*domain.Order, error) {
	return s.orderRepo.GetByID(ctx, orderID)
}

// GetOrderByTracking lấy đơn hàng theo số theo dõi
func (s *orderService) GetOrderByTracking(ctx context.Context, trackingNumber string) (*domain.Order, error) {
	return s.orderRepo.GetByTrackingNumber(ctx, trackingNumber)
}

// ListOrders lấy danh sách đơn hàng
func (s *orderService) ListOrders(
	ctx context.Context,
	customerID string,
	status domain.OrderStatus,
	offset,
	limit int,
) ([]*domain.Order, int, error) {
	return s.orderRepo.ListOrders(ctx, customerID, status, offset, limit)
}

// UpdateOrderStatus cập nhật trạng thái đơn hàng
func (s *orderService) UpdateOrderStatus(
	ctx context.Context,
	orderID string,
	newStatus domain.OrderStatus,
	location *domain.Location,
	note string,
) error {
	// Lấy các sự kiện của đơn hàng
	events, err := s.eventStore.GetEvents(ctx, orderID)
	if err != nil {
		return fmt.Errorf("lỗi khi lấy lịch sử sự kiện: %w", err)
	}

	if len(events) == 0 {
		return fmt.Errorf("không tìm thấy đơn hàng")
	}

	// Xây dựng lại trạng thái đơn hàng từ các sự kiện
	order := domain.RebuildFromEvents(events)
	if order == nil {
		return fmt.Errorf("không thể xây dựng lại đơn hàng từ sự kiện")
	}

	// Cập nhật trạng thái
	err = order.UpdateStatus(newStatus, location, note)
	if err != nil {
		return fmt.Errorf("không thể cập nhật trạng thái đơn hàng: %w", err)
	}

	// Lưu các sự kiện mới
	err = s.eventStore.SaveEvents(ctx, order.ID, order.GetUncommittedEvents())
	if err != nil {
		return fmt.Errorf("lỗi khi lưu sự kiện: %w", err)
	}

	// Phát các sự kiện
	for _, event := range order.GetUncommittedEvents() {
		if err := s.eventBus.Publish(event); err != nil {
			fmt.Printf("lỗi khi phát sự kiện: %v\n", err)
		}
	}

	// Xóa các sự kiện đã xử lý
	order.ClearUncommittedEvents()

	return nil
}

// CancelOrder hủy đơn hàng
func (s *orderService) CancelOrder(ctx context.Context, orderID string, reason string) error {
	// Lấy các sự kiện của đơn hàng
	events, err := s.eventStore.GetEvents(ctx, orderID)
	if err != nil {
		return fmt.Errorf("lỗi khi lấy lịch sử sự kiện: %w", err)
	}

	if len(events) == 0 {
		return fmt.Errorf("không tìm thấy đơn hàng")
	}

	// Xây dựng lại trạng thái đơn hàng từ các sự kiện
	order := domain.RebuildFromEvents(events)
	if order == nil {
		return fmt.Errorf("không thể xây dựng lại đơn hàng từ sự kiện")
	}

	// Hủy đơn hàng
	err = order.CancelOrder(reason)
	if err != nil {
		return fmt.Errorf("không thể hủy đơn hàng: %w", err)
	}

	// Lưu các sự kiện mới
	err = s.eventStore.SaveEvents(ctx, order.ID, order.GetUncommittedEvents())
	if err != nil {
		return fmt.Errorf("lỗi khi lưu sự kiện: %w", err)
	}

	// Phát các sự kiện
	for _, event := range order.GetUncommittedEvents() {
		if err := s.eventBus.Publish(event); err != nil {
			fmt.Printf("lỗi khi phát sự kiện: %v\n", err)
		}
	}

	// Xóa các sự kiện đã xử lý
	order.ClearUncommittedEvents()

	return nil
}

// AddOrderNote thêm ghi chú vào đơn hàng
func (s *orderService) AddOrderNote(ctx context.Context, orderID string, note string) error {
	// Lấy các sự kiện của đơn hàng
	events, err := s.eventStore.GetEvents(ctx, orderID)
	if err != nil {
		return fmt.Errorf("lỗi khi lấy lịch sử sự kiện: %w", err)
	}

	if len(events) == 0 {
		return fmt.Errorf("không tìm thấy đơn hàng")
	}

	// Xây dựng lại trạng thái đơn hàng từ các sự kiện
	order := domain.RebuildFromEvents(events)
	if order == nil {
		return fmt.Errorf("không thể xây dựng lại đơn hàng từ sự kiện")
	}

	// Thêm ghi chú
	err = order.AddNote(note)
	if err != nil {
		return fmt.Errorf("không thể thêm ghi chú: %w", err)
	}

	// Lưu các sự kiện mới
	err = s.eventStore.SaveEvents(ctx, order.ID, order.GetUncommittedEvents())
	if err != nil {
		return fmt.Errorf("lỗi khi lưu sự kiện: %w", err)
	}

	// Phát các sự kiện
	for _, event := range order.GetUncommittedEvents() {
		if err := s.eventBus.Publish(event); err != nil {
			fmt.Printf("lỗi khi phát sự kiện: %v\n", err)
		}
	}

	// Xóa các sự kiện đã xử lý
	order.ClearUncommittedEvents()

	return nil
}

// GetOrderHistory lấy lịch sử đơn hàng
func (s *orderService) GetOrderHistory(ctx context.Context, orderID string) ([]domain.Event, error) {
	// Lấy các sự kiện của đơn hàng
	events, err := s.eventStore.GetEvents(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("lỗi khi lấy lịch sử sự kiện: %w", err)
	}

	if len(events) == 0 {
		return nil, fmt.Errorf("không tìm thấy đơn hàng")
	}

	return events, nil
}
