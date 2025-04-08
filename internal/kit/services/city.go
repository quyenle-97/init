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

type CitySvc interface {
	Cities(ctx context.Context, search string) ([]models.City, error)
	City(ctx context.Context, id string) (models.City, error)
	CitiesByIds(ctx context.Context, ids []string) ([]models.City, error)
	CreateCity(ctx context.Context, req transforms.CityCreateReq) (models.City, error)
	UpdateCity(ctx context.Context, req transforms.CityUpdateReq) (models.City, error)
	DeleteCity(ctx context.Context, id string) error
}

type citySvc struct {
	db     *bun.DB
	logger *log.MultiLogger
}

func NewCitySvc(db *bun.DB, logger *log.MultiLogger) CitySvc {
	return &citySvc{
		db:     db,
		logger: logger,
	}
}

func (cv citySvc) Cities(ctx context.Context, search string) ([]models.City, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var rs []models.City
	q := cv.db.NewSelect().Model(&rs)
	if search != "" {
		q = q.Where("name ILIKE ?", "%"+search+"%")
	}
	err := q.Scan(ctx)
	if err != nil {
		cv.logger.Errorf("failed to fetch all cities: %v", err)
		return nil, err
	}
	return rs, nil
}

func (cv citySvc) City(ctx context.Context, id string) (models.City, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	uid, err := uuid.Parse(id)
	if err != nil {
		cv.logger.Errorf("Invalid city ID format: %v", err)
		return models.City{}, err
	}

	rs := models.City{
		ID: uid,
	}
	err = cv.db.NewSelect().
		Model(&rs).
		WherePK().
		Scan(ctx)
	if err != nil {
		cv.logger.Errorf("failed to fetch all city: %v", err)
		return models.City{}, err
	}
	return rs, nil
}

func (cv citySvc) CitiesByIds(ctx context.Context, ids []string) ([]models.City, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if len(ids) == 0 {
		return []models.City{}, nil
	}

	var uuids []uuid.UUID
	for _, id := range ids {
		cid, err := uuid.Parse(id)
		if err == nil {
			uuids = append(uuids, cid)
		}
	}

	if len(uuids) == 0 {
		return []models.City{}, nil
	}

	var rs []models.City
	err := cv.db.NewSelect().
		Model(&rs).
		Where("id IN (?)", bun.In(uuids)).
		Scan(ctx)
	if err != nil {
		cv.logger.Errorf("failed to fetch all cities: %v", err)
		return nil, err
	}
	return rs, nil
}

func (cv citySvc) CreateCity(ctx context.Context, req transforms.CityCreateReq) (models.City, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	city := models.City{
		Name:      req.Name,
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
	}
	_, err := cv.db.NewInsert().Model(&city).Exec(ctx)
	if err != nil {
		if utils.IsUniqueViolation(err) {
			cv.logger.Error("Failed to create city: city already exists")
			return models.City{}, errors.New("failed to create city: city already exists")
		}
		cv.logger.Errorf("Failed to create city: %v", err)
		return models.City{}, err
	}
	return city, nil
}

func (cv citySvc) UpdateCity(ctx context.Context, req transforms.CityUpdateReq) (models.City, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	supplier := models.City{
		ID:        uuid.MustParse(req.Id),
		Name:      req.Name,
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
	}
	eff, err := cv.db.NewUpdate().
		Model(&supplier).
		WherePK().
		OmitZero().
		Returning("*").
		Exec(ctx)
	if err != nil {
		if utils.IsUniqueViolation(err) {
			cv.logger.Error("Failed to update city: city name already exists")
			return models.City{}, errors.New("failed to update city: city name already exists")
		}
		cv.logger.Errorf("Failed to update city: %v", err)
		return models.City{}, err
	}
	up, err := eff.RowsAffected()
	if err != nil {
		cv.logger.Errorf("Failed to update city: %v", err)
		return models.City{}, errors.New("city not found")
	}
	if up == 0 {
		cv.logger.Error("Failed to update city: city not found")
		return models.City{}, errors.New("city not found")
	}
	return supplier, nil
}

func (cv citySvc) DeleteCity(ctx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	city := models.City{
		ID: uuid.MustParse(id),
	}

	eff, err := cv.db.NewDelete().
		Model(&city).
		WherePK().
		Exec(ctx)
	if err != nil {
		cv.logger.Errorf("failed to delete city: %v", err)
		return err
	}
	del, err := eff.RowsAffected()
	if err != nil {
		cv.logger.Errorf("failed to delete city: %v", err)
		return errors.New("city not found")
	}
	if del == 0 {
		cv.logger.Errorf("failed to delete city: city not found")
		return errors.New("city not found")
	}
	return nil
}
