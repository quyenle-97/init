package server

import (
	"github.com/gorilla/mux"
	"github.com/quyenle-97/init/cfg"
	"github.com/quyenle-97/init/internal/domain"
	"github.com/quyenle-97/init/internal/eventstore"
	"github.com/quyenle-97/init/internal/kit/endpoints"
	"github.com/quyenle-97/init/internal/kit/services"
	"github.com/quyenle-97/init/internal/kit/transports"
	"github.com/quyenle-97/init/internal/repository"
	"github.com/quyenle-97/init/pkgs/eventbus"
	"github.com/quyenle-97/init/pkgs/log"
	"github.com/uptrace/bun"
	"net/http"
)

// SetupLogisticsRoutes cấu hình các route liên quan đến logistics
func SetupLogisticsRoutes(r *mux.Router, db *bun.DB, logger *log.MultiLogger, c cfg.Config) *mux.Router {
	// Khởi tạo event store
	eventStore := eventstore.NewPostgresEventStore(db)

	// Khởi tạo event bus
	bus := eventbus.NewInMemoryEventBus()

	// Khởi tạo order repository (kết hợp cả repository và projection)
	orderRepo := repository.NewOrderRepository(db)
	//trackingProjection := projection.NewPostgresTrackingProjection(db)

	// Đăng ký các projections với event bus
	bus.Subscribe(orderRepo, domain.OrderCreatedType, domain.OrderStatusUpdatedType,
		domain.OrderCancelledType, domain.OrderNoteAddedType)

	// Khởi tạo service
	orderService := services.NewOrderService(eventStore, orderRepo, bus)

	// Khởi tạo endpoints
	orderEndpoints := endpoints.NewOrderEndpoints(orderService)

	// Đăng ký HTTP handlers
	transports.MakeOrderHandlers(r, orderEndpoints, c.BasePath+"logistics")

	// Tạo subrouter cho logistics API

	r.HandleFunc("/__health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	return r
}
