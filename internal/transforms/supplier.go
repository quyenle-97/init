package transforms

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/Minh2009/pv_soa/pkgs/utils"
	"github.com/gorilla/mux"
	"net/http"
)

type SuppliersReq struct {
	Search string `json:"search,omitempty"`
}

func DecodeSuppliersReq(_ context.Context, r *http.Request) (interface{}, error) {
	var req SuppliersReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}
	return req, nil
}

type SupplierReq struct {
	Id string `json:"id" validate:"required"`
}

func DecodeSupplierReq(_ context.Context, r *http.Request) (interface{}, error) {
	uid := mux.Vars(r)["uid"]
	if uid == "" {
		return nil, errors.New("invalid uid")
	}
	return SupplierReq{Id: uid}, nil
}

type SupplierCreateReq struct {
	Name string `json:"name" validate:"required"`
}

func DecodeSupplierCreateReq(_ context.Context, r *http.Request) (interface{}, error) {
	var req SupplierCreateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}
	err := utils.Validate(req)
	if err != nil {
		return nil, err
	}
	return req, nil
}

func DecodeReq(_ context.Context, r *http.Request) (interface{}, error) {
	return nil, nil
}

type SupplierUpdateReq struct {
	SupplierReq
	SupplierCreateReq
}

func DecodeSupplierUpdateReq(_ context.Context, r *http.Request) (interface{}, error) {
	var req SupplierUpdateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}
	req.Id = mux.Vars(r)["uid"]
	err := utils.Validate(req.SupplierReq)
	if err != nil {
		return nil, err
	}
	return req, nil
}
