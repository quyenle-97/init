package services

import (
	"context"
	"errors"
	"github.com/Minh2009/pv_soa/internal/models"
	"github.com/Minh2009/pv_soa/internal/transforms"
	"github.com/Minh2009/pv_soa/pkgs/log"
	"github.com/Minh2009/pv_soa/pkgs/utils"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"time"
)

type SupplierSvc interface {
	Suppliers(ctx context.Context, search string) ([]models.Supplier, error)
	Supplier(ctx context.Context, id string) (models.Supplier, error)
	SuppliersByIds(ctx context.Context, ids []string) ([]models.Supplier, error)
	CreateSupplier(ctx context.Context, req transforms.SupplierCreateReq) (models.Supplier, error)
	UpdateSupplier(ctx context.Context, req transforms.SupplierUpdateReq) (models.Supplier, error)
	DeleteSupplier(ctx context.Context, id string) error
}

type supplierSvc struct {
	db     *bun.DB
	logger *log.MultiLogger
}

func NewSupplierSvc(db *bun.DB, logger *log.MultiLogger) SupplierSvc {
	return &supplierSvc{
		db:     db,
		logger: logger,
	}
}

func (cv supplierSvc) Suppliers(ctx context.Context, search string) ([]models.Supplier, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var rs []models.Supplier
	q := cv.db.NewSelect().Model(&rs)
	if search != "" {
		q = q.Where("name ILIKE ?", "%"+search+"%")
	}
	err := q.Scan(ctx)
	if err != nil {
		cv.logger.Errorf("failed to fetch all Suppliers: %v", err)
		return nil, err
	}
	return rs, nil
}

func (cv supplierSvc) Supplier(ctx context.Context, id string) (models.Supplier, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	uid, err := uuid.Parse(id)
	if err != nil {
		cv.logger.Errorf("Invalid supplier ID format: %v", err)
		return models.Supplier{}, err
	}

	rs := models.Supplier{
		ID: uid,
	}
	err = cv.db.NewSelect().
		Model(&rs).
		WherePK().
		Scan(ctx)
	if err != nil {
		cv.logger.Errorf("failed to fetch all supplier: %v", err)
		return models.Supplier{}, err
	}
	return rs, nil
}

func (cv supplierSvc) SuppliersByIds(ctx context.Context, ids []string) ([]models.Supplier, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if len(ids) == 0 {
		return []models.Supplier{}, nil
	}

	var uuids []uuid.UUID
	for _, id := range ids {
		cid, err := uuid.Parse(id)
		if err == nil {
			uuids = append(uuids, cid)
		}
	}

	if len(uuids) == 0 {
		return []models.Supplier{}, nil
	}

	var rs []models.Supplier
	err := cv.db.NewSelect().
		Model(&rs).
		Where("id IN (?)", bun.In(uuids)).
		Scan(ctx)
	if err != nil {
		cv.logger.Errorf("failed to fetch all Suppliers: %v", err)
		return nil, err
	}
	return rs, nil
}

func (cv supplierSvc) CreateSupplier(ctx context.Context, req transforms.SupplierCreateReq) (models.Supplier, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	supplier := models.Supplier{
		Name: req.Name,
	}
	_, err := cv.db.NewInsert().Model(&supplier).Exec(ctx)
	if err != nil {
		if utils.IsUniqueViolation(err) {
			cv.logger.Error("Failed to create supplier: supplier already exists")
			return models.Supplier{}, errors.New("failed to create supplier: supplier already exists")
		}
		cv.logger.Errorf("Failed to create supplier: %v", err)
		return models.Supplier{}, err
	}
	return supplier, nil
}

func (cv supplierSvc) UpdateSupplier(ctx context.Context, req transforms.SupplierUpdateReq) (models.Supplier, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	supplier := models.Supplier{
		ID:   uuid.MustParse(req.Id),
		Name: req.Name,
	}
	eff, err := cv.db.NewUpdate().
		Model(&supplier).
		WherePK().
		OmitZero().
		Returning("*").
		Exec(ctx)
	if err != nil {
		if utils.IsUniqueViolation(err) {
			cv.logger.Error("Failed to update supplier: supplier name already exists")
			return models.Supplier{}, errors.New("failed to update supplier: supplier name already exists")
		}
		cv.logger.Errorf("Failed to update supplier: %v", err)
		return models.Supplier{}, err
	}
	up, err := eff.RowsAffected()
	if err != nil {
		cv.logger.Errorf("Failed to update supplier: %v", err)
		return models.Supplier{}, errors.New("supplier not found")
	}
	if up == 0 {
		cv.logger.Error("Failed to update supplier: supplier not found")
		return models.Supplier{}, errors.New("supplier not found")
	}
	return supplier, nil
}

func (cv supplierSvc) DeleteSupplier(ctx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	supplier := models.Supplier{
		ID: uuid.MustParse(id),
	}

	eff, err := cv.db.NewDelete().
		Model(&supplier).
		WherePK().
		Exec(ctx)
	if err != nil {
		cv.logger.Errorf("failed to delete supplier: %v", err)
		return err
	}
	del, err := eff.RowsAffected()
	if err != nil {
		cv.logger.Errorf("failed to delete supplier: %v", err)
		return errors.New("supplier not found")
	}
	if del == 0 {
		cv.logger.Errorf("failed to delete supplier: supplier not found")
		return errors.New("supplier not found")
	}
	return nil
}
