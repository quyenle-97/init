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

func CityHttpHandler(s services.CitySvc, log *log.MultiLogger, c cfg.Config) http.Handler {
	pr := mux.NewRouter()

	ep := endpoints.NewCityEndpoint(s)
	options := []httptransport.ServerOption{
		httptransport.ServerErrorHandler(log),
		httptransport.ServerErrorEncoder(utils.EncodeError),
	}

	pr.Methods(http.MethodPost).Path(utils.UrlWithPrefix("cities", c.BasePath)).Handler(httptransport.NewServer(
		ep.CreateCity(),
		transforms.DecodeCityCreateReq,
		utils.EncodeResponseHTTP,
		options...,
	))

	pr.Methods(http.MethodPost).Path(utils.UrlWithPrefix("cities/list", c.BasePath)).Handler(httptransport.NewServer(
		ep.Cities(),
		transforms.DecodeCitiesReq,
		utils.EncodeResponseHTTP,
		options...,
	))

	pr.Methods(http.MethodPut).Path(utils.UrlWithPrefix("cities/{uid}", c.BasePath)).Handler(httptransport.NewServer(
		ep.UpdateCity(),
		transforms.DecodeCityUpdateReq,
		utils.EncodeResponseHTTP,
		options...,
	))

	pr.Methods(http.MethodDelete).Path(utils.UrlWithPrefix("cities/{uid}", c.BasePath)).Handler(httptransport.NewServer(
		ep.DeleteCity(),
		transforms.DecodeCityReq,
		utils.EncodeResponseHTTP,
		options...,
	))

	return pr
}
