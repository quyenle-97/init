package endpoints

import (
	"context"
	"github.com/Minh2009/pv_soa/internal/kit/services"
	"github.com/Minh2009/pv_soa/internal/transforms"
	"github.com/Minh2009/pv_soa/pkgs/utils"
	"github.com/go-kit/kit/endpoint"
)

type SupplierEndpoint struct {
	service services.SupplierSvc
}

func NewSupplierEndpoint(s services.SupplierSvc) SupplierEndpoint {
	return SupplierEndpoint{service: s}
}

func (s SupplierEndpoint) Suppliers() endpoint.Endpoint {
	return func(ctx context.Context, r interface{}) (resp interface{}, err error) {
		req := r.(transforms.SuppliersReq)
		results, err := s.service.Suppliers(ctx, req.Search)
		if err != nil {
			return utils.SetDefaultResponse(ctx, utils.Message{Code: 500, Message: err.Error()}), nil
		}
		return utils.SetHttpResponse(ctx, utils.Message{Code: 200, Message: "success"}, results, nil), nil
	}
}

func (s SupplierEndpoint) CreateSupplier() endpoint.Endpoint {
	return func(ctx context.Context, r interface{}) (resp interface{}, err error) {
		req := r.(transforms.SupplierCreateReq)
		result, err := s.service.CreateSupplier(ctx, req)
		if err != nil {
			return utils.SetDefaultResponse(ctx, utils.Message{Code: 500, Message: err.Error()}), nil
		}
		return utils.SetHttpResponse(ctx, utils.Message{Code: 200, Message: "success"}, result, nil), nil
	}
}

func (s SupplierEndpoint) UpdateSupplier() endpoint.Endpoint {
	return func(ctx context.Context, r interface{}) (resp interface{}, err error) {
		req := r.(transforms.SupplierUpdateReq)
		result, err := s.service.UpdateSupplier(ctx, req)
		if err != nil {
			return utils.SetDefaultResponse(ctx, utils.Message{Code: 500, Message: err.Error()}), nil
		}
		return utils.SetHttpResponse(ctx, utils.Message{Code: 200, Message: "success"}, result, nil), nil
	}
}

func (s SupplierEndpoint) DeleteSupplier() endpoint.Endpoint {
	return func(ctx context.Context, r interface{}) (resp interface{}, err error) {
		req := r.(transforms.SupplierReq)
		err = s.service.DeleteSupplier(ctx, req.Id)
		if err != nil {
			return utils.SetDefaultResponse(ctx, utils.Message{Code: 500, Message: err.Error()}), nil
		}
		return utils.SetHttpResponse(ctx, utils.Message{Code: 200, Message: "success"}, nil, nil), nil
	}
}
