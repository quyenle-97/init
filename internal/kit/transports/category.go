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

func CategoryHttpHandler(s services.CategorySvc, log *log.MultiLogger, c cfg.Config) http.Handler {
	pr := mux.NewRouter()

	ep := endpoints.NewCategoryEndpoint(s)
	options := []httptransport.ServerOption{
		httptransport.ServerErrorHandler(log),
		httptransport.ServerErrorEncoder(utils.EncodeError),
	}

	pr.Methods(http.MethodPost).Path(utils.UrlWithPrefix("categories", c.BasePath)).Handler(httptransport.NewServer(
		ep.CreateCategory(),
		transforms.DecodeCateCreateReq,
		utils.EncodeResponseHTTP,
		options...,
	))

	pr.Methods(http.MethodPost).Path(utils.UrlWithPrefix("categories/list", c.BasePath)).Handler(httptransport.NewServer(
		ep.Categories(),
		transforms.DecodeCategoriesReq,
		utils.EncodeResponseHTTP,
		options...,
	))
	pr.Methods(http.MethodPut).Path(utils.UrlWithPrefix("categories/{uid}", c.BasePath)).Handler(httptransport.NewServer(
		ep.UpdateCategory(),
		transforms.DecodeCateUpdateReq,
		utils.EncodeResponseHTTP,
		options...,
	))

	pr.Methods(http.MethodDelete).Path(utils.UrlWithPrefix("categories/{uid}", c.BasePath)).Handler(httptransport.NewServer(
		ep.DeleteCategory(),
		transforms.DecodeCategoryReq,
		utils.EncodeResponseHTTP,
		options...,
	))

	return pr
}
