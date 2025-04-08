package transforms

import (
	"context"
	"encoding/json"
	"github.com/Minh2009/pv_soa/internal/models"
	"github.com/Minh2009/pv_soa/pkgs/utils"
	"github.com/gorilla/mux"
	"net/http"
)

type CateCreateReq struct {
	Name string `json:"name" validate:"required"`
}

func DecodeCateCreateReq(_ context.Context, r *http.Request) (interface{}, error) {
	var req CateCreateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}
	err := utils.Validate(req)
	if err != nil {
		return nil, err
	}
	return req, nil
}

type CategoriesReq struct {
	Search string `json:"search,omitempty"`
	Status int    `json:"status,omitempty"`
}

func DecodeCategoriesReq(_ context.Context, r *http.Request) (interface{}, error) {
	var req CategoriesReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}
	valid := []models.ModelStatus{models.MSDeActive, models.MSActive}
	if !utils.Contains(valid, models.ModelStatus(req.Status)) {
		req.Status = 0
	}
	return req, nil
}

type CateUpdateReq struct {
	Id     string `json:"id" validate:"required"`
	Name   string `json:"name"`
	Status int    `json:"status"`
}

func DecodeCateUpdateReq(_ context.Context, r *http.Request) (interface{}, error) {
	var req CateUpdateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}
	uid := mux.Vars(r)["uid"]
	req.Id = uid
	err := utils.Validate(req)
	if err != nil {
		return nil, err
	}
	valid := []models.ModelStatus{models.MSDeActive, models.MSActive}
	if !utils.Contains(valid, models.ModelStatus(req.Status)) {
		req.Status = 0
	}
	return req, nil
}

type CategoryReq struct {
	Id string `json:"id" validate:"required"`
}

func DecodeCategoryReq(_ context.Context, r *http.Request) (interface{}, error) {
	var req CategoryReq
	uid := mux.Vars(r)["uid"]
	req.Id = uid
	err := utils.Validate(req)
	if err != nil {
		return nil, err
	}
	return req, nil
}
