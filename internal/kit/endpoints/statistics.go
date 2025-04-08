package endpoints

import (
	"context"
	"github.com/Minh2009/pv_soa/internal/kit/services"
	"github.com/Minh2009/pv_soa/pkgs/utils"
	"github.com/go-kit/kit/endpoint"
)

type StatisticsEndpoint struct {
	service services.StatisticsSvc
}

func NewStatisticsEndpoint(s services.StatisticsSvc) StatisticsEndpoint {
	return StatisticsEndpoint{service: s}
}

func (s StatisticsEndpoint) ByCategories() endpoint.Endpoint {
	return func(ctx context.Context, r interface{}) (resp interface{}, err error) {
		results, err := s.service.ByCategories(ctx)
		if err != nil {
			return utils.SetDefaultResponse(ctx, utils.Message{Code: 500, Message: err.Error()}), nil
		}
		return utils.SetHttpResponse(ctx, utils.Message{Code: 200, Message: "success"}, results, nil), nil
	}
}

func (s StatisticsEndpoint) BySupplier() endpoint.Endpoint {
	return func(ctx context.Context, r interface{}) (resp interface{}, err error) {
		result, err := s.service.BySupplier(ctx)
		if err != nil {
			return utils.SetDefaultResponse(ctx, utils.Message{Code: 500, Message: err.Error()}), nil
		}
		return utils.SetHttpResponse(ctx, utils.Message{Code: 200, Message: "success"}, result, nil), nil
	}
}
