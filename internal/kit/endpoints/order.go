package endpoints

import (
	"context"
	"errors"
	"github.com/go-kit/kit/endpoint"
	"github.com/quyenle-97/init/internal/domain"
	"github.com/quyenle-97/init/internal/kit/services"
	"github.com/quyenle-97/init/internal/transforms"
	"time"
)

type OrderEndpoints struct {
	CreateOrder        endpoint.Endpoint
	GetOrder           endpoint.Endpoint
	ListOrders         endpoint.Endpoint
	UpdateOrderStatus  endpoint.Endpoint
	CancelOrder        endpoint.Endpoint
	AddOrderNote       endpoint.Endpoint
	GetOrderHistory    endpoint.Endpoint
	GetOrderByTracking endpoint.Endpoint
}

// NewOrderEndpoints tạo các endpoints cho order service
func NewOrderEndpoints(s services.OrderService) OrderEndpoints {
	return OrderEndpoints{
		CreateOrder:        makeCreateOrderEndpoint(s),
		GetOrder:           makeGetOrderEndpoint(s),
		ListOrders:         makeListOrdersEndpoint(s),
		UpdateOrderStatus:  makeUpdateOrderStatusEndpoint(s),
		CancelOrder:        makeCancelOrderEndpoint(s),
		AddOrderNote:       makeAddOrderNoteEndpoint(s),
		GetOrderHistory:    makeGetOrderHistoryEndpoint(s),
		GetOrderByTracking: makeGetOrderByTrackingEndpoint(s),
	}
}

// Triển khai từng endpoint...
func makeCreateOrderEndpoint(s services.OrderService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(transforms.CreateOrderRequest)
		orderID, trackingNumber, err := s.CreateOrder(ctx, req.CustomerID, req.Origin, req.Destination, req.Items)
		if err != nil {
			return nil, errors.New("Lỗi khi tạo đơn hàng: " + err.Error())
		}

		return transforms.CreateOrderResponse{
			OrderID:        orderID,
			TrackingNumber: trackingNumber,
		}, nil
	}
}

func makeGetOrderEndpoint(s services.OrderService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(transforms.GetOrderRequest)
		order, err := s.GetOrder(ctx, req.OrderID)

		if err != nil {
			return nil, errors.New("Lỗi khi lấy thông tin đơn hàng: " + err.Error())
		}

		return order, nil
	}
}

func makeListOrdersEndpoint(s services.OrderService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(transforms.ListOrdersRequest)
		orders, total, err := s.ListOrders(ctx, req.CustomerID, req.Status, req.Offset, req.Limit)
		if err != nil {
			return nil, errors.New("Lỗi khi lấy danh sách đơn hàng: " + err.Error())
		}

		// Tạo view model
		response := &transforms.ListOrdersResponse{
			Items:       make([]transforms.OrderSummaryResponse, len(orders)),
			TotalCount:  total,
			CurrentPage: req.Offset/req.Limit + 1,
			PageSize:    req.Limit,
		}

		// Map từng đơn hàng sang summary view model
		for i, order := range orders {
			response.Items[i] = transforms.OrderSummaryResponse{
				ID:             order.ID,
				TrackingNumber: order.TrackingNumber,
				Status:         order.Status,
				CreatedAt:      order.CreatedAt.Format(time.RFC3339),
				UpdatedAt:      order.UpdatedAt.Format(time.RFC3339),
			}
		}
		return response, nil
	}
}

func makeUpdateOrderStatusEndpoint(s services.OrderService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(transforms.UpdateOrderStatusRequest)
		err := s.UpdateOrderStatus(ctx, req.OrderID, req.NewStatus, req.Location, req.Note)
		if err != nil {
			return nil, errors.New("Lỗi khi cập nhật trạng thái đơn hàng: " + err.Error())
		}

		return transforms.UpdateOrderStatusResponse{
			Status:  "success",
			Message: "Trạng thái đơn hàng đã được cập nhật thành công",
		}, nil
	}
}

func makeCancelOrderEndpoint(s services.OrderService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(transforms.CancelOrderRequest)
		err := s.CancelOrder(ctx, req.OrderID, req.Reason)

		if err != nil {
			return nil, errors.New("Lỗi khi hủy đơn hàng: " + err.Error())
		}

		return transforms.CancelOrderResponse{
			Status:  "success",
			Message: "Đơn hàng đã được hủy thành công",
		}, nil
	}
}

func makeAddOrderNoteEndpoint(s services.OrderService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(transforms.AddOrderNoteRequest)
		err := s.AddOrderNote(ctx, req.OrderID, req.Note)

		if err != nil {
			return nil, errors.New("Lỗi khi thêm ghi chú: " + err.Error())
		}

		return transforms.CancelOrderResponse{
			Status:  "success",
			Message: "Ghi chú đã được thêm thành công",
		}, nil
	}
}

func makeGetOrderHistoryEndpoint(s services.OrderService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(transforms.GetOrderHistoryRequest)
		events, err := s.GetOrderHistory(ctx, req.OrderID)
		if err != nil {
			return nil, errors.New("Lỗi khi lấy lịch sử đơn hàng: " + err.Error())
		}

		response := &transforms.GetOrderHistoryResponse{
			OrderID: req.OrderID,
			Entries: make([]transforms.GetOrderHistoryEntryResponse, 0, len(events)),
		}

		// Chuyển đổi mỗi sự kiện thành một mục lịch sử
		for _, event := range events {
			entry := transforms.GetOrderHistoryEntryResponse{
				Timestamp: event.GetTimestamp().Format(time.RFC3339),
				EventType: string(event.GetType()),
			}

			// Xử lý từng loại sự kiện cụ thể
			switch e := event.(type) {
			case domain.OrderCreatedEvent:
				entry.Status = domain.OrderStatusCreated
				entry.Note = "Đơn hàng được tạo"
			case domain.OrderStatusUpdatedEvent:
				entry.Status = e.NewStatus
				entry.PrevStatus = e.OldStatus
				entry.Location = e.CurrentLocation
				entry.Note = e.Note
			case domain.OrderCancelledEvent:
				entry.Status = domain.OrderStatusCancelled
				entry.PrevStatus = e.PreviousStatus
				entry.Note = e.Reason
			case domain.OrderNoteAddedEvent:
				entry.Note = e.Note
			}

			response.Entries = append(response.Entries, entry)
		}

		return response, nil
	}
}

func makeGetOrderByTrackingEndpoint(s services.OrderService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(transforms.GetOrderByTrackingRequest)
		order, err := s.GetOrderByTracking(ctx, req.TrackingNumber)

		if err != nil {
			return nil, errors.New("Lỗi khi lấy thông tin đơn hàng: " + err.Error())
		}

		return order, nil
	}
}
