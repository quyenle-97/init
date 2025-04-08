package endpoints

import (
	"context"
	"github.com/Minh2009/pv_soa/internal/kit/services"
	"github.com/Minh2009/pv_soa/internal/transforms"
	"github.com/Minh2009/pv_soa/pkgs/utils"
	"github.com/go-kit/kit/endpoint"
)

type CategoryEndpoint struct {
	service services.CategorySvc
}

func NewCategoryEndpoint(s services.CategorySvc) CategoryEndpoint {
	return CategoryEndpoint{service: s}
}

func (s CategoryEndpoint) Categories() endpoint.Endpoint {
	return func(ctx context.Context, r interface{}) (resp interface{}, err error) {
		req := r.(transforms.CategoriesReq)
		results, err := s.service.Categories(ctx, req.Search, req.Status)
		if err != nil {
			return utils.SetDefaultResponse(ctx, utils.Message{Code: 500, Message: err.Error()}), nil
		}
		return utils.SetHttpResponse(ctx, utils.Message{Code: 200, Message: "success"}, results, nil), nil
	}
}

func (s CategoryEndpoint) CreateCategory() endpoint.Endpoint {
	return func(ctx context.Context, r interface{}) (resp interface{}, err error) {
		req := r.(transforms.CateCreateReq)
		result, err := s.service.CreateCategory(ctx, req.Name)
		if err != nil {
			return utils.SetDefaultResponse(ctx, utils.Message{Code: 500, Message: err.Error()}), nil
		}
		return utils.SetHttpResponse(ctx, utils.Message{Code: 200, Message: "success"}, result, nil), nil
	}
}

func (s CategoryEndpoint) UpdateCategory() endpoint.Endpoint {
	return func(ctx context.Context, r interface{}) (resp interface{}, err error) {
		req := r.(transforms.CateUpdateReq)
		result, err := s.service.UpdateCategory(ctx, req.Id, req.Name, req.Status)
		if err != nil {
			return utils.SetDefaultResponse(ctx, utils.Message{Code: 500, Message: err.Error()}), nil
		}
		return utils.SetHttpResponse(ctx, utils.Message{Code: 200, Message: "success"}, result, nil), nil
	}
}

func (s CategoryEndpoint) DeleteCategory() endpoint.Endpoint {
	return func(ctx context.Context, r interface{}) (resp interface{}, err error) {
		req := r.(transforms.CategoryReq)
		err = s.service.DeleteCategory(ctx, req.Id)
		if err != nil {
			return utils.SetDefaultResponse(ctx, utils.Message{Code: 500, Message: err.Error()}), nil
		}
		return utils.SetHttpResponse(ctx, utils.Message{Code: 200, Message: "success"}, nil, nil), nil
	}
}
