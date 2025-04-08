package services

import (
	"context"
	"errors"
	"fmt"
	"github.com/Minh2009/pv_soa/internal/models"
	"github.com/Minh2009/pv_soa/pkgs/log"
	"github.com/Minh2009/pv_soa/pkgs/utils"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"time"
)

type CategorySvc interface {
	Categories(ctx context.Context, search string, status int) ([]models.Category, error)
	CategoriesByIds(ctx context.Context, ids []string) ([]models.Category, error)
	CreateCategory(ctx context.Context, name string) (models.Category, error)
	UpdateCategory(ctx context.Context, id, name string, status int) (models.Category, error)
	DeleteCategory(ctx context.Context, id string) error
}

type categorySvc struct {
	db     *bun.DB
	logger *log.MultiLogger
}

func NewCategorySvc(db *bun.DB, logger *log.MultiLogger) CategorySvc {
	return &categorySvc{
		db:     db,
		logger: logger,
	}
}

func (cv categorySvc) Categories(ctx context.Context, search string, status int) ([]models.Category, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var rs []models.Category
	q := cv.db.NewSelect().Model(&rs)
	if search != "" {
		q = q.Where("name ILIKE ?", "%"+search+"%")
	}
	if status != 0 {
		q = q.Where("status = ?", models.ModelStatus(status))
	}
	err := q.Scan(ctx)
	if err != nil {
		cv.logger.Errorf("failed to fetch all categories: %v", err)
		return nil, err
	}
	return rs, nil
}

func (cv categorySvc) CategoriesByIds(ctx context.Context, ids []string) ([]models.Category, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if len(ids) == 0 {
		return []models.Category{}, nil
	}

	var uuids []uuid.UUID
	for _, id := range ids {
		cid, err := uuid.Parse(id)
		if err == nil {
			uuids = append(uuids, cid)
		}
	}

	if len(uuids) == 0 {
		return []models.Category{}, nil
	}

	var rs []models.Category
	err := cv.db.NewSelect().
		Model(&rs).
		Where("id IN (?)", bun.In(uuids)).
		Scan(ctx)
	if err != nil {
		cv.logger.Errorf("failed to fetch all categories: %v", err)
		return nil, err
	}
	return rs, nil
}

func (cv categorySvc) CreateCategory(ctx context.Context, name string) (models.Category, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	category := models.Category{
		Name:   name,
		Status: models.MSActive,
	}
	_, err := cv.db.NewInsert().Model(&category).Exec(ctx)
	if err != nil {
		if utils.IsUniqueViolation(err) {
			cv.logger.Error("Failed to create category: category already exists")
			return models.Category{}, errors.New("failed to create category: category already exists")
		}
		cv.logger.Errorf("Failed to create category: %v", err)
		return models.Category{}, err
	}
	return category, nil
}

func (cv categorySvc) UpdateCategory(ctx context.Context, id, name string, status int) (models.Category, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	uid, err := uuid.Parse(id)
	if err != nil {
		cv.logger.Errorf("failed to parse id: %v", err)
		return models.Category{}, fmt.Errorf("id invalid: %v", err)
	}

	category := models.Category{
		ID:   uid,
		Name: name,
	}
	if status != 0 {
		category.Status = models.ModelStatus(status)
	}
	eff, err := cv.db.NewUpdate().
		Model(&category).
		WherePK().
		Exec(ctx)
	if err != nil {
		if utils.IsUniqueViolation(err) {
			cv.logger.Error("Failed to update category: category name already exists")
			return models.Category{}, errors.New("failed to update category: category name already exists")
		}
		cv.logger.Errorf("Failed to update category: %v", err)
		return models.Category{}, err
	}
	up, err := eff.RowsAffected()
	if err != nil {
		cv.logger.Errorf("Failed to update category: %v", err)
		return models.Category{}, errors.New("category not found")
	}
	if up == 0 {
		cv.logger.Error("Failed to update category: category not found")
		return models.Category{}, errors.New("category not found")
	}
	return category, nil
}

func (cv categorySvc) DeleteCategory(ctx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	category := models.Category{
		ID: uuid.MustParse(id),
	}

	eff, err := cv.db.NewDelete().
		Model(&category).
		WherePK().
		Exec(ctx)
	if err != nil {
		cv.logger.Errorf("failed to delete category: %v", err)
		return err
	}
	del, err := eff.RowsAffected()
	if err != nil {
		cv.logger.Errorf("failed to delete category: %v", err)
		return errors.New("category not found")
	}
	if del == 0 {
		cv.logger.Errorf("failed to delete category: category not found")
		return errors.New("category not found")
	}
	return nil
}
