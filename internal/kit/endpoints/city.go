package endpoints

import (
	"context"
	"github.com/Minh2009/pv_soa/internal/kit/services"
	"github.com/Minh2009/pv_soa/internal/transforms"
	"github.com/Minh2009/pv_soa/pkgs/utils"
	"github.com/go-kit/kit/endpoint"
)

type CityEndpoint struct {
	service services.CitySvc
}

func NewCityEndpoint(s services.CitySvc) CityEndpoint {
	return CityEndpoint{service: s}
}

func (s CityEndpoint) Cities() endpoint.Endpoint {
	return func(ctx context.Context, r interface{}) (resp interface{}, err error) {
		req := r.(transforms.CitiesReq)
		results, err := s.service.Cities(ctx, req.Search)
		if err != nil {
			return utils.SetDefaultResponse(ctx, utils.Message{Code: 500, Message: err.Error()}), nil
		}
		return utils.SetHttpResponse(ctx, utils.Message{Code: 200, Message: "success"}, results, nil), nil
	}
}

func (s CityEndpoint) CreateCity() endpoint.Endpoint {
	return func(ctx context.Context, r interface{}) (resp interface{}, err error) {
		req := r.(transforms.CityCreateReq)
		result, err := s.service.CreateCity(ctx, req)
		if err != nil {
			return utils.SetDefaultResponse(ctx, utils.Message{Code: 500, Message: err.Error()}), nil
		}
		return utils.SetHttpResponse(ctx, utils.Message{Code: 200, Message: "success"}, result, nil), nil
	}
}

func (s CityEndpoint) UpdateCity() endpoint.Endpoint {
	return func(ctx context.Context, r interface{}) (resp interface{}, err error) {
		req := r.(transforms.CityUpdateReq)
		result, err := s.service.UpdateCity(ctx, req)
		if err != nil {
			return utils.SetDefaultResponse(ctx, utils.Message{Code: 500, Message: err.Error()}), nil
		}
		return utils.SetHttpResponse(ctx, utils.Message{Code: 200, Message: "success"}, result, nil), nil
	}
}

func (s CityEndpoint) DeleteCity() endpoint.Endpoint {
	return func(ctx context.Context, r interface{}) (resp interface{}, err error) {
		req := r.(transforms.CityReq)
		err = s.service.DeleteCity(ctx, req.Id)
		if err != nil {
			return utils.SetDefaultResponse(ctx, utils.Message{Code: 500, Message: err.Error()}), nil
		}
		return utils.SetHttpResponse(ctx, utils.Message{Code: 200, Message: "success"}, nil, nil), nil
	}
}
