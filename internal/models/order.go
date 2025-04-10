package models

import (
	"github.com/quyenle-97/init/internal/domain"
	"github.com/uptrace/bun"
	"time"
)

type OrderModel struct {
	bun.BaseModel `bun:"table:orders,alias:o"`

	ID              string             `bun:"id,pk"`
	CustomerID      string             `bun:"customer_id,notnull"`
	TrackingNumber  string             `bun:"tracking_number,notnull,unique"`
	Status          domain.OrderStatus `bun:"status,notnull"`
	OriginData      []byte             `bun:"origin_data,notnull"`
	DestinationData []byte             `bun:"destination_data,notnull"`
	CurrentLocData  []byte             `bun:"current_location_data"`
	ItemsData       []byte             `bun:"items_data,notnull"`
	NotesData       []byte             `bun:"notes_data,notnull"`
	CreatedAt       time.Time          `bun:"created_at,notnull"`
	UpdatedAt       time.Time          `bun:"updated_at,notnull"`
}
