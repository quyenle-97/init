package models

import (
	"context"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"time"
)

type Supplier struct {
	bun.BaseModel `bun:"table:mh_suppliers"`

	ID   uuid.UUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()"`
	Name string    `json:"name" bun:"name,notnull,unique"`

	CreatedTime time.Time  `json:"created_time" bun:"created_time,notnull,default:current_timestamp"`
	UpdatedTime time.Time  `json:"updated_time" bun:"updated_time,notnull,default:current_timestamp"`
	DeletedTime *time.Time `json:"deleted_time" bun:"deleted_time,soft_delete"`
}

var _ bun.BeforeAppendModelHook = (*Supplier)(nil)

func (m *Supplier) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.UpdateQuery:
		m.UpdatedTime = time.Now()
	}
	return nil
}
