package transforms

import (
	"context"
	"encoding/json"
	"github.com/Minh2009/pv_soa/pkgs/utils"
	"github.com/gorilla/mux"
	"net/http"
)

type CitiesReq struct {
	Search string `json:"search,omitempty"`
}

func DecodeCitiesReq(_ context.Context, r *http.Request) (interface{}, error) {
	var req CitiesReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}
	return req, nil
}

type CityReq struct {
	Id string `json:"id" validate:"required"`
}

func DecodeCityReq(_ context.Context, r *http.Request) (interface{}, error) {
	uid := mux.Vars(r)["uid"]
	if uid == "" {
		return nil, utils.Message{Code: 422, Message: "invalid uid"}
	}
	return CityReq{Id: uid}, nil
}

type CityCreateReq struct {
	Name      string  `json:"name" validate:"required"`
	Latitude  float64 `json:"latitude" validate:"required"`
	Longitude float64 `json:"longitude" validate:"required"`
}

func DecodeCityCreateReq(_ context.Context, r *http.Request) (interface{}, error) {
	var req CityCreateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}
	err := utils.Validate(req)
	if err != nil {
		return nil, err
	}
	return req, nil
}

type CityUpdateReq struct {
	CityReq
	CityCreateReq
}

func DecodeCityUpdateReq(_ context.Context, r *http.Request) (interface{}, error) {
	var req CityUpdateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}
	req.Id = mux.Vars(r)["uid"]
	err := utils.Validate(req.CityReq)
	if err != nil {
		return nil, err
	}
	return req, nil
}
