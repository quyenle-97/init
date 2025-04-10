package migrations

import (
	"context"
	"github.com/quyenle-97/init/internal/models"
	"reflect"
	"time"

	"github.com/uptrace/bun"
)

// ProjectionsTable định nghĩa các bảng cho modelss
type ProjectionsTable struct {
	Version int
}

func (m ProjectionsTable) Up(db *bun.DB) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// Tạo bảng orders cho models đơn hàng
	_, err = db.NewCreateTable().
		Model((*models.OrderModel)(nil)).
		IfNotExists().
		Exec(ctx)
	if err != nil {
		return err
	}

	// Tạo các index cho bảng orders
	_, err = db.NewCreateIndex().
		Model((*models.OrderModel)(nil)).
		Index("idx_orders_customer_id").
		Column("customer_id").
		IfNotExists().
		Exec(ctx)
	if err != nil {
		return err
	}

	_, err = db.NewCreateIndex().
		Model((*models.OrderModel)(nil)).
		Index("idx_orders_tracking_number").
		Column("tracking_number").
		Unique().
		IfNotExists().
		Exec(ctx)
	if err != nil {
		return err
	}

	_, err = db.NewCreateIndex().
		Model((*models.OrderModel)(nil)).
		Index("idx_orders_status").
		Column("status").
		IfNotExists().
		Exec(ctx)
	if err != nil {
		return err
	}

	//// Tạo bảng tracking_info cho models theo dõi
	//_, err = db.NewCreateTable().
	//	Model((*models.TrackingModel)(nil)).
	//	IfNotExists().
	//	Exec(ctx)
	//if err != nil {
	//	return err
	//}
	//
	//// Tạo các index cho bảng tracking_info
	//_, err = db.NewCreateIndex().
	//	Model((*models.TrackingModel)(nil)).
	//	Index("idx_tracking_order_id").
	//	Column("order_id").
	//	IfNotExists().
	//	Exec(ctx)
	//if err != nil {
	//	return err
	//}
	//
	//_, err = db.NewCreateIndex().
	//	Model((*models.TrackingModel)(nil)).
	//	Index("idx_tracking_tracking_number").
	//	Column("tracking_number").
	//	Unique().
	//	IfNotExists().
	//	Exec(ctx)
	//if err != nil {
	//	return err
	//}
	//
	//// Tạo bảng tracking_updates cho cập nhật theo dõi
	//_, err = db.NewCreateTable().
	//	Model((*models.TrackingUpdateModel)(nil)).
	//	IfNotExists().
	//	Exec(ctx)
	//if err != nil {
	//	return err
	//}
	//
	//// Tạo index cho bảng tracking_updates
	//_, err = db.NewCreateIndex().
	//	Model((*models.TrackingUpdateModel)(nil)).
	//	Index("idx_tracking_updates_tracking_id").
	//	Column("tracking_id").
	//	IfNotExists().
	//	Exec(ctx)
	//if err != nil {
	//	return err
	//}
	//
	//_, err = db.NewCreateIndex().
	//	Model((*models.TrackingUpdateModel)(nil)).
	//	Index("idx_tracking_updates_timestamp").
	//	Column("timestamp").
	//	IfNotExists().
	//	Exec(ctx)
	//if err != nil {
	//	return err
	//}

	return nil
}

func (m ProjectionsTable) Down(db *bun.DB) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// Xóa các bảng theo thứ tự ngược lại để tránh lỗi khóa ngoại
	//_, err = db.NewDropTable().
	//	Model((*models.TrackingUpdateModel)(nil)).
	//	IfExists().
	//	Cascade().
	//	Exec(ctx)
	//if err != nil {
	//	return err
	//}
	//
	//_, err = db.NewDropTable().
	//	Model((*models.TrackingModel)(nil)).
	//	IfExists().
	//	Cascade().
	//	Exec(ctx)
	//if err != nil {
	//	return err
	//}

	_, err = db.NewDropTable().
		Model((*models.OrderModel)(nil)).
		IfExists().
		Cascade().
		Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (m ProjectionsTable) GetStructName() string {
	if t := reflect.TypeOf(m); t.Kind() == reflect.Ptr {
		return t.Elem().Name()
	} else {
		return t.Name()
	}
}
