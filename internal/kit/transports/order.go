package transports

import (
	"context"
	"encoding/json"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/quyenle-97/init/internal/kit/endpoints"
	"github.com/quyenle-97/init/internal/transforms"
	"net/http"
)

func MakeOrderHandlers(r *mux.Router, ep endpoints.OrderEndpoints, basePath string) {
	validate := validator.New()
	options := []httptransport.ServerOption{
		httptransport.ServerErrorEncoder(encodeError),
	}

	// POST /orders - Create a new order
	r.Methods("POST").Path(basePath + "/orders").Handler(httptransport.NewServer(
		ep.CreateOrder,
		transforms.DecodeCreateOrderRequest,
		encodeResponse,
		options...,
	))

	// GET /orders/{id} - Lấy thông tin chi tiết đơn hàng theo ID
	r.Methods("GET").Path(basePath + "/orders/{id}").Handler(httptransport.NewServer(
		ep.GetOrder,
		transforms.DecodeGetOrderRequest,
		encodeResponse,
		options...,
	))

	// GET /orders - Liệt kê các đơn hàng
	r.Methods("GET").Path(basePath + "/orders").Handler(httptransport.NewServer(
		ep.ListOrders,
		transforms.DecodeListOrdersRequest,
		encodeResponse,
		options...,
	))

	// PUT /orders/{id}/status - Cập nhật trạng thái đơn hàng
	r.Methods("PUT").Path(basePath + "/orders/{id}/status").Handler(httptransport.NewServer(
		ep.UpdateOrderStatus,
		transforms.DecodeUpdateOrderStatusRequest(validate),
		encodeResponse,
		options...,
	))

	// POST /orders/{id}/cancel - Hủy đơn hàng
	r.Methods("POST").Path(basePath + "/orders/{id}/cancel").Handler(httptransport.NewServer(
		ep.CancelOrder,
		transforms.DecodeCancelOrderRequest,
		encodeResponse,
		options...,
	))

	// POST /orders/{id}/notes - Thêm ghi chú cho đơn hàng
	r.Methods("POST").Path(basePath + "/orders/{id}/notes").Handler(httptransport.NewServer(
		ep.AddOrderNote,
		transforms.DecodeAddOrderNoteRequest(validate),
		encodeResponse,
		options...,
	))

	// GET /orders/{id}/history - Lấy lịch sử đơn hàng
	r.Methods("GET").Path(basePath + "/orders/{id}/history").Handler(httptransport.NewServer(
		ep.GetOrderHistory,
		transforms.DecodeGetOrderHistoryRequest,
		encodeResponse,
		options...,
	))

	// GET /orders/tracking/{tracking_number} - Lấy thông tin đơn hàng theo số theo dõi
	r.Methods("GET").Path(basePath + "/orders/tracking/{tracking_number}").Handler(httptransport.NewServer(
		ep.GetOrderByTracking,
		transforms.DecodeGetOrderByTrackingRequest,
		encodeResponse,
		options...,
	))

}

func encodeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusInternalServerError)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}
