package server

import (
	"github.com/Minh2009/pv_soa/cfg"
	"github.com/Minh2009/pv_soa/internal/kit/services"
	"github.com/Minh2009/pv_soa/internal/kit/transports"
	"github.com/Minh2009/pv_soa/pkgs/log"
	"github.com/Minh2009/pv_soa/pkgs/utils"
	"github.com/oschwald/geoip2-golang"
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

func Routing(c cfg.Config, db *bun.DB, log *log.MultiLogger, cache redis.UniversalClient) *http.ServeMux {
	mux := http.NewServeMux()

	swagHttp := transports.SwaggerHttpHandler(c)
	mux.Handle("/", swagHttp) //don't delete or change this!!
	mux.HandleFunc("/__health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(http.StatusText(http.StatusOK)))
	})

	cateSvc := services.NewCategorySvc(db, log)
	cateTrans := transports.CategoryHttpHandler(cateSvc, log, c)
	mux.Handle(utils.UrlWithPrefix("categories", c.BasePath), cateTrans)
	mux.Handle(utils.UrlWithPrefix("categories/", c.BasePath), cateTrans)

	citySvc := services.NewCitySvc(db, log)
	cityTrans := transports.CityHttpHandler(citySvc, log, c)
	mux.Handle(utils.UrlWithPrefix("cities", c.BasePath), cityTrans)
	mux.Handle(utils.UrlWithPrefix("cities/", c.BasePath), cityTrans)

	supplierSvc := services.NewSupplierSvc(db, log)
	supplierTrans := transports.SupplierHttpHandler(supplierSvc, log, c)
	mux.Handle(utils.UrlWithPrefix("suppliers", c.BasePath), supplierTrans)
	mux.Handle(utils.UrlWithPrefix("suppliers/", c.BasePath), supplierTrans)

	geoIPReader, err := geoip2.Open("GeoLite2-City.mmdb")
	if err != nil {
		panic(err)
	}

	productSvc := services.NewProductSvc(db, log, cateSvc, citySvc, supplierSvc, geoIPReader, cache)
	productTrans := transports.ProductHttpHandler(productSvc, log, c)
	mux.Handle(utils.UrlWithPrefix("products", c.BasePath), productTrans)
	mux.Handle(utils.UrlWithPrefix("products/", c.BasePath), productTrans)

	statisticsSvc := services.NewStatisticsSvc(db, cache, log)
	statisticsTrans := transports.StatisticsHttpHandler(statisticsSvc, log, c)
	mux.Handle(utils.UrlWithPrefix("statistics/", c.BasePath), statisticsTrans)

	return mux
}
