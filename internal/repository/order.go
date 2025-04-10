package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/quyenle-97/init/internal/domain"
	"github.com/quyenle-97/init/internal/models"
	"github.com/uptrace/bun"
)

type OrderRepository interface {
	// GetByID lấy đơn hàng theo ID
	GetByID(ctx context.Context, id string) (*domain.Order, error)

	// GetByTrackingNumber lấy đơn hàng theo số theo dõi
	GetByTrackingNumber(ctx context.Context, trackingNumber string) (*domain.Order, error)

	// ListOrders lấy danh sách đơn hàng
	ListOrders(ctx context.Context, customerID string, status domain.OrderStatus, offset, limit int) ([]*domain.Order, int, error)

	//// Save lưu đơn hàng mới
	//Save(ctx context.Context, order *models.OrderModel) error

	//// Update cập nhật đơn hàng
	//Update(ctx context.Context, order *models.OrderModel) error

	HandleEvent(event domain.Event) error
}

type orderRepository struct {
	db *bun.DB
}

// NewOrderRepository tạo repository mới
func NewOrderRepository(db *bun.DB) OrderRepository {
	return &orderRepository{
		db: db,
	}
}

// GetByID lấy đơn hàng theo ID
func (r *orderRepository) GetByID(ctx context.Context, id string) (*domain.Order, error) {
	model := &models.OrderModel{}
	err := r.db.NewSelect().
		Model(model).
		Where("id = ?", id).
		Scan(ctx)

	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, errors.New("không tìm thấy đơn hàng")
		}
		return nil, fmt.Errorf("lỗi khi truy vấn đơn hàng: %w", err)
	}

	return r.modelToDomain(model)
}

// GetByTrackingNumber lấy đơn hàng theo số theo dõi
func (r *orderRepository) GetByTrackingNumber(ctx context.Context, trackingNumber string) (*domain.Order, error) {
	model := &models.OrderModel{}
	err := r.db.NewSelect().
		Model(model).
		Where("tracking_number = ?", trackingNumber).
		Scan(ctx)

	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, errors.New("không tìm thấy đơn hàng")
		}
		return nil, fmt.Errorf("lỗi khi truy vấn đơn hàng theo số theo dõi: %w", err)
	}

	return r.modelToDomain(model)
}

// ListOrders lấy danh sách đơn hàng theo các tiêu chí
func (r *orderRepository) ListOrders(ctx context.Context, customerID string, status domain.OrderStatus, offset, limit int) ([]*domain.Order, int, error) {
	query := r.db.NewSelect().
		Model((*models.OrderModel)(nil))

	// Áp dụng các bộ lọc
	if customerID != "" {
		query = query.Where("customer_id = ?", customerID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Đếm tổng số đơn hàng
	//var count int
	//err := query.Clone().Count(ctx, &count)
	//if err != nil {
	//	return nil, 0, fmt.Errorf("lỗi khi đếm đơn hàng: %w", err)
	//}

	// Lấy các đơn hàng với phân trang
	var orderModels []*models.OrderModel
	count, err := query.
		Model(&orderModels).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		ScanAndCount(ctx)

	if err != nil {
		return nil, 0, fmt.Errorf("lỗi khi truy vấn danh sách đơn hàng: %w", err)
	}

	// Chuyển đổi model thành domain
	orders := make([]*domain.Order, len(orderModels))
	for i, model := range orderModels {
		order, err := r.modelToDomain(model)
		if err != nil {
			return nil, 0, err
		}
		orders[i] = order
	}

	return orders, count, nil
}

// HandleEvent xử lý các sự kiện để cập nhật read model
func (r *orderRepository) HandleEvent(event domain.Event) error {
	ctx := context.Background()
	switch e := event.(type) {
	case domain.OrderCreatedEvent:
		return r.handleOrderCreated(ctx, e)
	case domain.OrderStatusUpdatedEvent:
		return r.handleOrderStatusUpdated(ctx, e)
	case domain.OrderCancelledEvent:
		return r.handleOrderCancelled(ctx, e)
	case domain.OrderNoteAddedEvent:
		return r.handleOrderNoteAdded(ctx, e)
	default:

		fmt.Println("12312312312", e)

		return nil // Bỏ qua các sự kiện không quan tâm
	}
}

// handleOrderCreated xử lý sự kiện tạo đơn hàng
func (r *orderRepository) handleOrderCreated(ctx context.Context, event domain.OrderCreatedEvent) error {
	fmt.Println(333333)
	fmt.Println(event)
	// Serialize dữ liệu
	originData, err := json.Marshal(event.Origin)
	if err != nil {
		return fmt.Errorf("lỗi khi serialize origin: %w", err)
	}

	destData, err := json.Marshal(event.Destination)
	if err != nil {
		return fmt.Errorf("lỗi khi serialize destination: %w", err)
	}

	itemsData, err := json.Marshal(event.Items)
	if err != nil {
		return fmt.Errorf("lỗi khi serialize items: %w", err)
	}

	notesData, err := json.Marshal([]string{})
	if err != nil {
		return fmt.Errorf("lỗi khi serialize notes: %w", err)
	}

	// Tạo model đơn hàng mới
	model := models.OrderModel{
		ID:              event.AggregateID,
		CustomerID:      event.CustomerID,
		TrackingNumber:  event.TrackingNumber,
		Status:          domain.OrderStatusCreated,
		OriginData:      originData,
		DestinationData: destData,
		ItemsData:       itemsData,
		NotesData:       notesData,
		CreatedAt:       event.Timestamp,
		UpdatedAt:       event.Timestamp,
	}

	// Lưu vào cơ sở dữ liệu
	_, err = r.db.NewInsert().
		Model(&model).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("lỗi khi lưu đơn hàng mới: %w", err)
	}

	return nil
}

// handleOrderStatusUpdated xử lý sự kiện cập nhật trạng thái đơn hàng
func (r *orderRepository) handleOrderStatusUpdated(ctx context.Context, event domain.OrderStatusUpdatedEvent) error {
	// Lấy thông tin đơn hàng hiện tại
	var model models.OrderModel
	err := r.db.NewSelect().
		Model(&model).
		Where("id = ?", event.AggregateID).
		Scan(ctx)

	if err != nil {
		return fmt.Errorf("lỗi khi tìm đơn hàng: %w", err)
	}

	// Cập nhật trạng thái
	model.Status = event.NewStatus
	model.UpdatedAt = event.Timestamp

	// Cập nhật vị trí hiện tại nếu có
	if event.CurrentLocation != nil {
		currentLocData, err := json.Marshal(event.CurrentLocation)
		if err != nil {
			return fmt.Errorf("lỗi khi serialize current location: %w", err)
		}
		model.CurrentLocData = currentLocData
	}

	// Cập nhật ghi chú nếu có
	if event.Note != "" {
		var notes []string
		err = json.Unmarshal(model.NotesData, &notes)
		if err != nil {
			return fmt.Errorf("lỗi khi deserialize notes: %w", err)
		}

		notes = append(notes, event.Note)
		notesData, err := json.Marshal(notes)
		if err != nil {
			return fmt.Errorf("lỗi khi serialize notes: %w", err)
		}

		model.NotesData = notesData
	}

	// Lưu cập nhật vào cơ sở dữ liệu
	_, err = r.db.NewUpdate().
		Model(&model).
		WherePK().
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("lỗi khi cập nhật đơn hàng: %w", err)
	}

	return nil
}

// handleOrderCancelled xử lý sự kiện hủy đơn hàng
func (r *orderRepository) handleOrderCancelled(ctx context.Context, event domain.OrderCancelledEvent) error {
	// Lấy thông tin đơn hàng hiện tại
	var model models.OrderModel
	err := r.db.NewSelect().
		Model(&model).
		Where("id = ?", event.AggregateID).
		Scan(ctx)

	if err != nil {
		return fmt.Errorf("lỗi khi tìm đơn hàng: %w", err)
	}

	// Cập nhật trạng thái
	model.Status = domain.OrderStatusCancelled
	model.UpdatedAt = event.Timestamp

	// Cập nhật ghi chú nếu có
	if event.Reason != "" {
		var notes []string
		err = json.Unmarshal(model.NotesData, &notes)
		if err != nil {
			return fmt.Errorf("lỗi khi deserialize notes: %w", err)
		}

		notes = append(notes, event.Reason)
		notesData, err := json.Marshal(notes)
		if err != nil {
			return fmt.Errorf("lỗi khi serialize notes: %w", err)
		}

		model.NotesData = notesData
	}

	// Lưu cập nhật vào cơ sở dữ liệu
	_, err = r.db.NewUpdate().
		Model(&model).
		WherePK().
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("lỗi khi cập nhật đơn hàng hủy: %w", err)
	}

	return nil
}

// handleOrderNoteAdded xử lý sự kiện thêm ghi chú
func (r *orderRepository) handleOrderNoteAdded(ctx context.Context, event domain.OrderNoteAddedEvent) error {
	// Lấy thông tin đơn hàng hiện tại
	var model models.OrderModel
	err := r.db.NewSelect().
		Model(&model).
		Where("id = ?", event.AggregateID).
		Scan(ctx)

	if err != nil {
		return fmt.Errorf("lỗi khi tìm đơn hàng: %w", err)
	}

	// Cập nhật ghi chú
	var notes []string
	err = json.Unmarshal(model.NotesData, &notes)
	if err != nil {
		return fmt.Errorf("lỗi khi deserialize notes: %w", err)
	}

	notes = append(notes, event.Note)
	notesData, err := json.Marshal(notes)
	if err != nil {
		return fmt.Errorf("lỗi khi serialize notes: %w", err)
	}

	model.NotesData = notesData
	model.UpdatedAt = event.Timestamp

	// Lưu cập nhật vào cơ sở dữ liệu
	_, err = r.db.NewUpdate().
		Model(&model).
		WherePK().
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("lỗi khi cập nhật ghi chú đơn hàng: %w", err)
	}

	return nil
}

// modelToDomain chuyển đổi model thành domain
func (r *orderRepository) modelToDomain(model *models.OrderModel) (*domain.Order, error) {
	var origin domain.Location
	err := json.Unmarshal(model.OriginData, &origin)
	if err != nil {
		return nil, fmt.Errorf("lỗi khi deserialize origin: %w", err)
	}

	var destination domain.Location
	err = json.Unmarshal(model.DestinationData, &destination)
	if err != nil {
		return nil, fmt.Errorf("lỗi khi deserialize destination: %w", err)
	}

	var items []domain.OrderItem
	err = json.Unmarshal(model.ItemsData, &items)
	if err != nil {
		return nil, fmt.Errorf("lỗi khi deserialize items: %w", err)
	}

	var notes []string
	err = json.Unmarshal(model.NotesData, &notes)
	if err != nil {
		return nil, fmt.Errorf("lỗi khi deserialize notes: %w", err)
	}

	var currentLocation *domain.Location
	if model.CurrentLocData != nil && len(model.CurrentLocData) > 0 {
		currentLocation = &domain.Location{}
		err = json.Unmarshal(model.CurrentLocData, currentLocation)
		if err != nil {
			return nil, fmt.Errorf("lỗi khi deserialize current location: %w", err)
		}
	}

	order := &domain.Order{
		ID:              model.ID,
		CustomerID:      model.CustomerID,
		TrackingNumber:  model.TrackingNumber,
		Status:          model.Status,
		Origin:          origin,
		Destination:     destination,
		CurrentLocation: currentLocation,
		Items:           items,
		Notes:           notes,
		CreatedAt:       model.CreatedAt,
		UpdatedAt:       model.UpdatedAt,
	}

	return order, nil
}
