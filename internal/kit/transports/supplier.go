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

func SupplierHttpHandler(s services.SupplierSvc, log *log.MultiLogger, c cfg.Config) http.Handler {
	pr := mux.NewRouter()

	ep := endpoints.NewSupplierEndpoint(s)
	options := []httptransport.ServerOption{
		httptransport.ServerErrorHandler(log),
		httptransport.ServerErrorEncoder(utils.EncodeError),
	}

	pr.Methods(http.MethodPost).Path(utils.UrlWithPrefix("suppliers", c.BasePath)).Handler(httptransport.NewServer(
		ep.CreateSupplier(),
		transforms.DecodeSupplierCreateReq,
		utils.EncodeResponseHTTP,
		options...,
	))

	pr.Methods(http.MethodPost).Path(utils.UrlWithPrefix("suppliers/list", c.BasePath)).Handler(httptransport.NewServer(
		ep.Suppliers(),
		transforms.DecodeSuppliersReq,
		utils.EncodeResponseHTTP,
		options...,
	))

	pr.Methods(http.MethodPut).Path(utils.UrlWithPrefix("suppliers/{uid}", c.BasePath)).Handler(httptransport.NewServer(
		ep.UpdateSupplier(),
		transforms.DecodeSupplierUpdateReq,
		utils.EncodeResponseHTTP,
		options...,
	))

	pr.Methods(http.MethodDelete).Path(utils.UrlWithPrefix("suppliers/{uid}", c.BasePath)).Handler(httptransport.NewServer(
		ep.DeleteSupplier(),
		transforms.DecodeSupplierReq,
		utils.EncodeResponseHTTP,
		options...,
	))

	return pr
}
