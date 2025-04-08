package models

import (
	"context"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"time"
)

type City struct {
	bun.BaseModel `bun:"table:mh_cities"`

	ID        uuid.UUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()"`
	Name      string    `json:"name" bun:"name,notnull,unique"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`

	CreatedTime time.Time  `json:"created_time" bun:"created_time,notnull,default:current_timestamp"`
	UpdatedTime time.Time  `json:"updated_time" bun:"updated_time,notnull,default:current_timestamp"`
	DeletedTime *time.Time `json:"deleted_time" bun:"deleted_time,soft_delete"`
}

var _ bun.BeforeAppendModelHook = (*City)(nil)

func (m *City) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.UpdateQuery:
		m.UpdatedTime = time.Now()
	}
	return nil
}
