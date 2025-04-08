package models

import (
	"github.com/uptrace/bun"
)

type ModelStatus int

const (
	MSActive   ModelStatus = 1
	MSDeActive ModelStatus = 2
)

type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type GEO struct {
	Name      string  `json:"name"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

func Init(db *bun.DB) {
	db.RegisterModel((*ProductCategory)(nil))
}
