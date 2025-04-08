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

func ProductHttpHandler(s services.ProductSvc, log *log.MultiLogger, c cfg.Config) http.Handler {
	pr := mux.NewRouter()

	ep := endpoints.NewProductEndpoint(s)
	options := []httptransport.ServerOption{
		httptransport.ServerErrorHandler(log),
		httptransport.ServerErrorEncoder(utils.EncodeError),
	}

	pr.Methods(http.MethodPost).Path(utils.UrlWithPrefix("products", c.BasePath)).Handler(httptransport.NewServer(
		ep.CreateProduct(),
		transforms.DecodeProductCreateReq,
		utils.EncodeResponseHTTP,
		options...,
	))

	pr.Methods(http.MethodPost).Path(utils.UrlWithPrefix("products/list", c.BasePath)).Handler(httptransport.NewServer(
		ep.Products(),
		transforms.DecodeProductsReq,
		utils.EncodeResponseHTTP,
		options...,
	))

	pr.Methods(http.MethodGet).Path(utils.UrlWithPrefix("products/{uid}", c.BasePath)).Handler(httptransport.NewServer(
		ep.Product(),
		transforms.DecodeProductReq,
		utils.EncodeResponseHTTP,
		options...,
	))

	pr.Methods(http.MethodGet).Path(utils.UrlWithPrefix("products/{uid}/distance", c.BasePath)).Handler(httptransport.NewServer(
		ep.ProductDistance(),
		transforms.DecodeProductDistanceReq,
		utils.EncodeResponseHTTP,
		options...,
	))

	pr.Methods(http.MethodPut).Path(utils.UrlWithPrefix("products/{uid}", c.BasePath)).Handler(httptransport.NewServer(
		ep.UpdateProduct(),
		transforms.DecodeProductUpdateReq,
		utils.EncodeResponseHTTP,
		options...,
	))

	pr.Methods(http.MethodDelete).Path(utils.UrlWithPrefix("products/{uid}", c.BasePath)).Handler(httptransport.NewServer(
		ep.DeleteProduct(),
		transforms.DecodeProductReq,
		utils.EncodeResponseHTTP,
		options...,
	))

	return pr
}
