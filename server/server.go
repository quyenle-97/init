package server

import (
	"github.com/gorilla/mux"
	"github.com/quyenle-97/init/cfg"
	"github.com/quyenle-97/init/pkgs/log"
	"github.com/quyenle-97/init/pkgs/utils"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/uptrace/bun"
	"net/http"
	"runtime/debug"
)

type Middleware func(http.Handler) http.Handler

func MuxRecovery(logger *log.MultiLogger) Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					utils.ResponseWriter(w, http.StatusInternalServerError, utils.SetDefaultResponse(req.Context(), utils.Message{Code: 500}))
					logger.Log(logrus.ErrorLevel, "panic", err)
					debug.PrintStack()
				}
			}()
			h.ServeHTTP(w, req)
		})
	}
}

func Adapt(h http.Handler, middlewares ...Middleware) http.Handler {
	for _, middleware := range middlewares {
		h = middleware(h)
	}
	return h
}

func TraceIdentifier() Middleware {
	return utils.TraceIdentifierMiddleware
}

func AppMiddleware(handler http.Handler, lg *log.MultiLogger) http.Handler {
	return Adapt(
		handler,
		MuxRecovery(lg),
		TraceIdentifier(),
	)
}

func Routing(c cfg.Config, db *bun.DB, log *log.MultiLogger, cache redis.UniversalClient) *mux.Router {

	//// Thiết lập các route cho logistics
	//// Sử dụng gorilla/mux router để xử lý các route logistics
	r := mux.NewRouter()
	//// Kết hợp các handler từ logistics với các handler hiện tại
	SetupLogisticsRoutes(r, db, log, c)

	return r
}

// getRoutes là một hàm helper để lấy tất cả các route từ mux
func getRoutes(mux *mux.Router) map[string]http.Handler {
	routes := make(map[string]http.Handler)

	// Đây là một cách đơn giản hóa, thực tế bạn cần triển khai logic
	// để lấy tất cả các route từ mux

	// Trong thực tế, nếu bạn sử dụng gorilla/mux, bạn có thể lấy các route thông qua
	// r.Walk() để lấy tất cả các route và bổ sung vào mux chính

	return routes
}
