package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/uptrace/bun"
	"time"
)

type ProductStatus int

const (
	Available  ProductStatus = 1
	OutOfStock ProductStatus = 2
	OnOrder    ProductStatus = 3
)

var lastIndex *string

type Product struct {
	bun.BaseModel `bun:"table:mh_products"`

	ID        uuid.UUID       `json:"id" bun:"id,pk,type:uuid,default:uuid_generate_v4()"`
	Reference string          `json:"reference" bun:"reference,notnull,unique"`
	Name      string          `json:"name" bun:"name,notnull"`
	Status    ProductStatus   `json:"status" bun:"status"`
	Price     decimal.Decimal `json:"price" bun:"price"`
	Quantity  int64           `json:"quantity" bun:"quantity"`

	Categories []Category `json:"categories" bun:"m2m:mh_product_categories,join:Product=Category"`

	CityId uuid.UUID `json:"city_id" bun:"city_id,notnull,type:uuid"`
	City   City      `json:"city" bun:"rel:belongs-to,join:city_id=id"`

	SupplierId uuid.UUID `json:"supplier_id" bun:"supplier_id,notnull,type:uuid"`
	Supplier   Supplier  `json:"supplier" bun:"rel:belongs-to,join:supplier_id=id"`

	CreatedTime time.Time  `json:"created_time" bun:"created_time,notnull,default:current_timestamp"`
	UpdatedTime time.Time  `json:"updated_time" bun:"updated_time,notnull,default:current_timestamp"`
	DeletedTime *time.Time `json:"deleted_time" bun:"deleted_time,soft_delete"`
}

var _ bun.BeforeAppendModelHook = (*Product)(nil)

func generateProductCode(ctx context.Context, db *bun.DB) (string, error) {
	// Format current date as YYYYMM
	datePrefix := time.Now().Format("200601") // Go's date format: 2006 is year, 01 is month

	// Get the current max number for this month's prefix
	var maxCode string
	if lastIndex == nil {
		err := db.NewSelect().
			ColumnExpr("COALESCE(MAX(reference), 'PROD-"+datePrefix+"-000')").
			TableExpr("mh_products").
			Where("reference LIKE ?", "PROD-"+datePrefix+"-%").
			Scan(ctx, &maxCode)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return "0", fmt.Errorf("failed to get max product code: %w", err)
		}
		if errors.Is(err, sql.ErrNoRows) {
			maxCode = fmt.Sprintf("PROD-%s-%03d", datePrefix, 0)
		}
		lastIndex = &maxCode
	} else {
		maxCode = *lastIndex
	}

	// Extract the numeric part (last 3 digits)
	var numericPart int
	if len(maxCode) >= 12 { // PROD-YYYYMM-XXX format has at least 12 chars
		_, err := fmt.Sscanf(maxCode[len(maxCode)-3:], "%03d", &numericPart)
		if err != nil {
			numericPart = 0 // Default to 0 if parsing fails
		}
	}
	//// Increment the number and format the new code
	newCode := fmt.Sprintf("PROD-%s-%03d", datePrefix, numericPart+1)
	return newCode, nil
}

func (m *Product) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch q := query.(type) {
	case *bun.InsertQuery:
		code, err := generateProductCode(ctx, q.DB())
		if err != nil {
			return fmt.Errorf("failed to generate product code: %w", err)
		}
		m.Reference = code
	case *bun.UpdateQuery:
		m.UpdatedTime = time.Now()
	}
	return nil
}

type ProductCategory struct {
	bun.BaseModel `bun:"table:mh_product_categories,alias:pc"`

	ProductID uuid.UUID `bun:"product_id,pk,notnull,type:uuid"`
	Product   Product   `bun:"rel:belongs-to,join:product_id=id"`

	CategoryID uuid.UUID `bun:"category_id,pk,notnull,type:uuid"`
	Category   Category  `bun:"rel:belongs-to,join:category_id=id"`
}
