package endpoints

import (
	"context"
	"github.com/Minh2009/pv_soa/internal/kit/services"
	"github.com/Minh2009/pv_soa/internal/transforms"
	"github.com/Minh2009/pv_soa/pkgs/utils"
	"github.com/go-kit/kit/endpoint"
)

type ProductEndpoint struct {
	service services.ProductSvc
}

func NewProductEndpoint(s services.ProductSvc) ProductEndpoint {
	return ProductEndpoint{service: s}
}

func (s ProductEndpoint) Products() endpoint.Endpoint {
	return func(ctx context.Context, r interface{}) (resp interface{}, err error) {
		req := r.(transforms.ProductsReq)
		results, offset, err := s.service.Products(ctx, req)
		if err != nil {
			return utils.SetDefaultResponse(ctx, utils.Message{Code: 500, Message: err.Error()}), nil
		}
		return utils.SetHttpResponse(ctx, utils.Message{Code: 200, Message: "success"}, results, &utils.Pagination{
			Offset: offset, Limit: req.GetLimit(),
		}), nil
	}
}

func (s ProductEndpoint) Product() endpoint.Endpoint {
	return func(ctx context.Context, r interface{}) (resp interface{}, err error) {
		req := r.(transforms.ProductReq)
		rs, err := s.service.Product(ctx, req.Id)
		if err != nil {
			return utils.SetDefaultResponse(ctx, utils.Message{Code: 500, Message: err.Error()}), nil
		}
		return utils.SetHttpResponse(ctx, utils.Message{Code: 200, Message: "success"}, rs, nil), nil
	}
}

func (s ProductEndpoint) ProductDistance() endpoint.Endpoint {
	return func(ctx context.Context, r interface{}) (resp interface{}, err error) {
		req := r.(transforms.ProductDistanceReq)
		rs, err := s.service.ProductDistance(ctx, req.IP, req.Id)
		if err != nil {
			return utils.SetDefaultResponse(ctx, utils.Message{Code: 500, Message: err.Error()}), nil
		}
		return utils.SetHttpResponse(ctx, utils.Message{Code: 200, Message: "success"}, rs, nil), nil
	}
}

func (s ProductEndpoint) CreateProduct() endpoint.Endpoint {
	return func(ctx context.Context, r interface{}) (resp interface{}, err error) {
		req := r.(transforms.ProductCreateReq)
		result, err := s.service.CreateProduct(ctx, req)
		if err != nil {
			return utils.SetDefaultResponse(ctx, utils.Message{Code: 500, Message: err.Error()}), nil
		}
		return utils.SetHttpResponse(ctx, utils.Message{Code: 200, Message: "success"}, result, nil), nil
	}
}

func (s ProductEndpoint) UpdateProduct() endpoint.Endpoint {
	return func(ctx context.Context, r interface{}) (resp interface{}, err error) {
		req := r.(transforms.ProductUpdateReq)
		result, err := s.service.UpdateProduct(ctx, req)
		if err != nil {
			return utils.SetDefaultResponse(ctx, utils.Message{Code: 500, Message: err.Error()}), nil
		}
		return utils.SetHttpResponse(ctx, utils.Message{Code: 200, Message: "success"}, result, nil), nil
	}
}

func (s ProductEndpoint) DeleteProduct() endpoint.Endpoint {
	return func(ctx context.Context, r interface{}) (resp interface{}, err error) {
		req := r.(transforms.ProductReq)
		err = s.service.DeleteProduct(ctx, req.Id)
		if err != nil {
			return utils.SetDefaultResponse(ctx, utils.Message{Code: 500, Message: err.Error()}), nil
		}
		return utils.SetHttpResponse(ctx, utils.Message{Code: 200, Message: "success"}, nil, nil), nil
	}
}
