package transports

import (
	"github.com/Minh2009/pv_soa/cfg"
	"github.com/Minh2009/pv_soa/internal/kit/endpoints"
	"github.com/Minh2009/pv_soa/internal/kit/services"
	"github.com/Minh2009/pv_soa/internal/transforms"
	"github.com/Minh2009/pv_soa/pkgs/log"
	"github.com/Minh2009/pv_soa/pkgs/utils"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"net/http"
)

func StatisticsHttpHandler(s services.StatisticsSvc, log *log.MultiLogger, c cfg.Config) http.Handler {
	pr := mux.NewRouter()

	ep := endpoints.NewStatisticsEndpoint(s)
	options := []httptransport.ServerOption{
		httptransport.ServerErrorHandler(log),
		httptransport.ServerErrorEncoder(utils.EncodeError),
	}

	pr.Methods(http.MethodGet).Path(utils.UrlWithPrefix("statistics/products-per-category", c.BasePath)).Handler(httptransport.NewServer(
		ep.ByCategories(),
		transforms.DecodeReq,
		utils.EncodeResponseHTTP,
		options...,
	))

	pr.Methods(http.MethodGet).Path(utils.UrlWithPrefix("statistics/products-per-supplier", c.BasePath)).Handler(httptransport.NewServer(
		ep.BySupplier(),
		transforms.DecodeReq,
		utils.EncodeResponseHTTP,
		options...,
	))

	return pr
}
