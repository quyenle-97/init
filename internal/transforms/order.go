package transforms

import (
	"context"
	"encoding/json"
	"fmt"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/quyenle-97/init/internal/domain"
	"net/http"
	"strconv"
)

// CreateOrderRequest Request
type CreateOrderRequest struct {
	CustomerID  string             `json:"customer_id,omitempty"`
	Origin      domain.Location    `json:"origin,omitempty"`
	Destination domain.Location    `json:"destination,omitempty"`
	Items       []domain.OrderItem `json:"items,omitempty"`
}

// CreateOrderResponse Responses
type CreateOrderResponse struct {
	OrderID        string `json:"order_id,omitempty"`
	TrackingNumber string `json:"tracking_number,omitempty"`
}

// DecodeCreateOrderRequest functions
func DecodeCreateOrderRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}
	return req, nil
}

// GetOrderRequest Các request
type GetOrderRequest struct {
	OrderID string
}

type GetOrderResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// DecodeGetOrderRequest xử lý việc giải mã request lấy thông tin đơn hàng
func DecodeGetOrderRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return nil, fmt.Errorf("thiếu tham số id")
	}

	return GetOrderRequest{OrderID: id}, nil
}

// ListOrdersRequest truy vấn danh sách đơn hàng
type ListOrdersRequest struct {
	CustomerID string             `json:"customer_id"`
	Status     domain.OrderStatus `json:"status"`
	Offset     int                `json:"offset"`
	Limit      int                `json:"limit"`
}

type OrderSummaryResponse struct {
	ID             string             `json:"id"`
	TrackingNumber string             `json:"tracking_number"`
	Status         domain.OrderStatus `json:"status"`
	CreatedAt      string             `json:"created_at"`
	UpdatedAt      string             `json:"updated_at"`
}

type ListOrdersResponse struct {
	Items       []OrderSummaryResponse `json:"items"`
	TotalCount  int                    `json:"total_count"`
	CurrentPage int                    `json:"current_page"`
	PageSize    int                    `json:"page_size"`
}

// DecodeListOrdersRequest xử lý việc giải mã request liệt kê đơn hàng
func DecodeListOrdersRequest(_ context.Context, r *http.Request) (interface{}, error) {
	// Lấy tham số từ query string
	q := r.URL.Query()
	customerID := q.Get("customer_id")
	status := q.Get("status")

	// Parse offset và limit
	offset, err := parseIntParam(q.Get("offset"), 0)
	if err != nil {
		return nil, fmt.Errorf("offset không hợp lệ: %w", err)
	}

	limit, err := parseIntParam(q.Get("limit"), 10)
	if err != nil {
		return nil, fmt.Errorf("limit không hợp lệ: %w", err)
	}

	return ListOrdersRequest{
		CustomerID: customerID,
		Status:     domain.OrderStatus(status),
		Offset:     offset,
		Limit:      limit,
	}, nil
}

type UpdateOrderStatusRequest struct {
	OrderID   string             `json:"order_id" validate:"required"`
	NewStatus domain.OrderStatus `json:"new_status" validate:"required"`
	Location  *domain.Location   `json:"location"`
	Note      string             `json:"note"`
}

type UpdateOrderStatusResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// DecodeUpdateOrderStatusRequest xử lý việc giải mã request cập nhật trạng thái
func DecodeUpdateOrderStatusRequest(validate *validator.Validate) httptransport.DecodeRequestFunc {
	return func(_ context.Context, r *http.Request) (interface{}, error) {
		vars := mux.Vars(r)
		id, ok := vars["id"]
		if !ok {
			return nil, fmt.Errorf("thiếu tham số id")
		}

		var req UpdateOrderStatusRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			return nil, fmt.Errorf("không thể decode request: %w", err)
		}

		// Thêm order ID từ URL
		req.OrderID = id

		// Validate request
		if err := validate.Struct(req); err != nil {
			return nil, fmt.Errorf("request không hợp lệ: %w", err)
		}

		return req, nil
	}
}

type CancelOrderRequest struct {
	OrderID string `json:"order_id" validate:"required"`
	Reason  string `json:"reason"`
}

type CancelOrderResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// DecodeCancelOrderRequest xử lý việc giải mã request hủy đơn hàng
func DecodeCancelOrderRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return nil, fmt.Errorf("thiếu tham số id")
	}

	var req CancelOrderRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		// Nếu body rỗng hoặc không phải JSON, tạo một request trống
		req = CancelOrderRequest{}
	}

	// Thêm order ID từ URL
	req.OrderID = id

	return req, nil
}

type AddOrderNoteRequest struct {
	OrderID string `json:"order_id" validate:"required"`
	Note    string `json:"note" validate:"required"`
}

type AddOrderNoteResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// DecodeAddOrderNoteRequest xử lý việc giải mã request thêm ghi chú
func DecodeAddOrderNoteRequest(validate *validator.Validate) httptransport.DecodeRequestFunc {
	return func(_ context.Context, r *http.Request) (interface{}, error) {
		vars := mux.Vars(r)
		id, ok := vars["id"]
		if !ok {
			return nil, fmt.Errorf("thiếu tham số id")
		}

		var req AddOrderNoteRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			return nil, fmt.Errorf("không thể decode request: %w", err)
		}

		// Thêm order ID từ URL
		req.OrderID = id

		// Validate request
		if err := validate.Struct(req); err != nil {
			return nil, fmt.Errorf("request không hợp lệ: %w", err)
		}

		return req, nil
	}
}

// GetOrderHistoryRequest truy vấn lịch sử của đơn hàng
type GetOrderHistoryRequest struct {
	OrderID string `json:"order_id" validate:"required"`
}

// GetOrderHistoryEntryResponse là view model của một mục trong lịch sử đơn hàng
type GetOrderHistoryEntryResponse struct {
	Timestamp  string             `json:"timestamp"`
	EventType  string             `json:"event_type"`
	Status     domain.OrderStatus `json:"status,omitempty"`
	Location   *domain.Location   `json:"location,omitempty"`
	Note       string             `json:"note,omitempty"`
	PrevStatus domain.OrderStatus `json:"prev_status,omitempty"`
}

// GetOrderHistoryResponse là view model của lịch sử đơn hàng
type GetOrderHistoryResponse struct {
	OrderID string                         `json:"order_id"`
	Entries []GetOrderHistoryEntryResponse `json:"entries"`
}

// DecodeGetOrderHistoryRequest xử lý việc giải mã request lấy lịch sử đơn hàng
func DecodeGetOrderHistoryRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return nil, fmt.Errorf("thiếu tham số id")
	}

	return GetOrderHistoryRequest{OrderID: id}, nil
}

// GetOrderByTrackingRequest truy vấn đơn hàng theo số theo dõi
type GetOrderByTrackingRequest struct {
	TrackingNumber string `json:"tracking_number" validate:"required"`
}

// DecodeGetOrderByTrackingRequest xử lý việc giải mã request lấy đơn hàng theo số theo dõi
func DecodeGetOrderByTrackingRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	trackingNumber, ok := vars["tracking_number"]
	if !ok {
		return nil, fmt.Errorf("thiếu tham số tracking_number")
	}

	return GetOrderByTrackingRequest{TrackingNumber: trackingNumber}, nil
}

// parseIntParam chuyển đổi string thành int với giá trị mặc định
func parseIntParam(param string, defaultValue int) (int, error) {
	if param == "" {
		return defaultValue, nil
	}

	value, err := strconv.Atoi(param)
	if err != nil {
		return 0, err
	}

	return value, nil
}
