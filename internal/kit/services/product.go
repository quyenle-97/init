package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/oschwald/geoip2-golang"
	"github.com/redis/go-redis/v9"
	"github.com/uptrace/bun"
	"math"
	"net"
	"time"

	"github.com/Minh2009/pv_soa/internal/models"
	"github.com/Minh2009/pv_soa/internal/transforms"
	"github.com/Minh2009/pv_soa/pkgs/log"
	"github.com/Minh2009/pv_soa/pkgs/utils"
)

type ProductSvc interface {
	Products(ctx context.Context, req transforms.ProductsReq) ([]models.Product, int, error)
	Product(ctx context.Context, id string) (models.Product, error)
	ProductDistance(ctx context.Context, ip, id string) (float64, error)
	CreateProduct(ctx context.Context, req transforms.ProductCreateReq) (models.Product, error)
	UpdateProduct(ctx context.Context, req transforms.ProductUpdateReq) (models.Product, error)
	DeleteProduct(ctx context.Context, id string) error
}

type productSvc struct {
	db          *bun.DB
	logger      *log.MultiLogger
	categorySvc CategorySvc
	supplierSvc SupplierSvc
	citySvc     CitySvc
	geoIPReader *geoip2.Reader
	cache       redis.UniversalClient
}

func NewProductSvc(db *bun.DB, logger *log.MultiLogger, cateSvc CategorySvc, citySvc CitySvc, supplierSvc SupplierSvc, geoIPReader *geoip2.Reader, cache redis.UniversalClient) ProductSvc {
	return &productSvc{
		db:          db,
		logger:      logger,
		categorySvc: cateSvc,
		citySvc:     citySvc,
		supplierSvc: supplierSvc,
		geoIPReader: geoIPReader,
		cache:       cache,
	}
}

func (cv productSvc) Products(ctx context.Context, req transforms.ProductsReq) ([]models.Product, int, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var cateUuids []uuid.UUID
	var cityUuids []uuid.UUID

	wg := utils.NewWgGroup()
	wg.Go(func() error {
		if len(req.Categories) > 0 {
			exists, err := cv.categorySvc.CategoriesByIds(ctx, req.Categories)
			if err != nil {
				return err
			}
			for _, cat := range exists {
				cateUuids = append(cateUuids, cat.ID)
			}
		}
		return nil
	})
	wg.Go(func() error {
		if len(req.Cities) > 0 {
			exists, err := cv.citySvc.CitiesByIds(ctx, req.Cities)
			if err != nil {
				return err
			}
			for _, city := range exists {
				cityUuids = append(cityUuids, city.ID)
			}
		}
		return nil
	})
	err := wg.Wait()
	if err != nil {
		return nil, 0, err
	}

	var rs []models.Product
	q := cv.db.NewSelect().
		Relation("Categories").
		Relation("City").
		Relation("Supplier").
		Model(&rs)
	if len(req.Reference) > 0 {
		q.Where("reference_id In (?)", bun.In(req.Reference))
	}
	if len(req.Names) > 0 {
		q.Where("name In (?)", bun.In(req.Names))
	}
	//if req.AddFrom != nil {
	//	q.Where("created_time >= ?", req.AddFrom.UTC())
	//}
	//if req.AddTo != nil {
	//	q.Where("created_time <= ?", req.AddTo.UTC())
	//}
	if len(req.Status) > 0 {
		var st []models.ProductStatus
		valid := []models.ProductStatus{models.Available, models.OutOfStock, models.OnOrder}
		for _, status := range req.Status {
			if utils.Contains(valid, models.ProductStatus(status)) && !utils.Contains(st, models.ProductStatus(status)) {
				st = append(st, models.ProductStatus(status))
			}
		}
		if len(st) > 0 {
			q.Where("status IN (?)", bun.In(st))
		}
	}
	if len(cateUuids) > 0 {
		q = q.Where("EXISTS (SELECT * FROM pc WHERE pc.product_id = mh_product.id AND pc.category_id IN (?))",
			bun.In(cateUuids))
	}
	if len(cityUuids) > 0 {
		q = q.Where("city_id IN (?)", bun.In(cityUuids))
	}
	err = q.Order("created_time desc").
		Limit(req.GetLimit()).
		Offset(req.GetOffset()).
		Scan(ctx, &rs)
	return rs, req.GetOffset() + req.GetOffset(), nil
}

func (cv productSvc) Product(ctx context.Context, id string) (models.Product, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	uid, err := uuid.Parse(id)
	if err != nil {
		cv.logger.Errorf("invalid product id: %s", id)
		return models.Product{}, fmt.Errorf("invalid product id: %s", id)
	}
	product := models.Product{ID: uid}
	err = cv.db.NewSelect().
		Model(&product).
		Relation("Categories").
		Relation("City").
		Relation("Supplier").
		WherePK().
		Scan(ctx)
	if err != nil {
		return models.Product{}, err
	}
	return product, nil
}

func (cv productSvc) calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371.0 // Earth radius in kilometers

	// Convert degrees to radians
	lat1Rad := lat1 * math.Pi / 180
	lon1Rad := lon1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	lon2Rad := lon2 * math.Pi / 180

	// Haversine formula
	dLat := lat2Rad - lat1Rad
	dLon := lon2Rad - lon1Rad
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	distance := earthRadius * c

	return distance
}

func (cv productSvc) getLocationByIP(ipAddr string) (models.Location, error) {
	ip := net.ParseIP(ipAddr)
	record, err := cv.geoIPReader.City(ip)
	if err != nil {
		return models.Location{}, err
	}

	return models.Location{
		Latitude:  record.Location.Latitude,
		Longitude: record.Location.Longitude,
	}, nil
}

func (cv productSvc) ProductDistance(ctx context.Context, ip, id string) (float64, error) {
	product, err := cv.Product(ctx, id)
	if err != nil {
		return 0, err
	}
	userLocation, err := cv.getLocationByIP(ip)
	if err != nil {
		cv.logger.Errorf("failed to get location by ip: %s, err: %v", ip, err)
		return 0, fmt.Errorf("failed to get location by ip: %s", ip)
	}
	distance := cv.calculateDistance(
		userLocation.Latitude, userLocation.Longitude,
		product.City.Latitude, product.City.Longitude,
	)
	return math.Round(distance*100) / 100, nil
}

const (
	CategoryProductKey = "category_product:%s"
	SupplierProductKey = "supplier_product:%s"
)

func (cv productSvc) CreateProduct(ctx context.Context, req transforms.ProductCreateReq) (models.Product, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	product := models.Product{
		Name:   req.Name,
		Status: models.ProductStatus(req.Status),
	}
	if req.Price != nil {
		product.Price = *req.Price
	}
	if req.Quantity != nil {
		product.Quantity = *req.Quantity
	}

	var err error
	var city models.City
	var supplier models.Supplier
	var existPCs []models.Category

	wg := utils.NewWgGroup()
	wg.Go(func() error {
		city, err = cv.citySvc.City(ctx, req.CityId)
		if err != nil {
			return err
		}
		return nil
	})
	wg.Go(func() error {
		supplier, err = cv.supplierSvc.Supplier(ctx, req.SupplierId)
		if err != nil {
			return err
		}
		return nil
	})
	wg.Go(func() error {
		existPCs, err = cv.categorySvc.CategoriesByIds(ctx, req.Categories)
		if err != nil {
			return err
		}
		if len(existPCs) == 0 {
			return errors.New("don't have any existed product categories")
		}
		return nil
	})
	err = wg.Wait()
	if err != nil {
		return models.Product{}, err
	}

	product.CityId = city.ID
	product.SupplierId = supplier.ID

	err = cv.db.RunInTx(ctx, &sql.TxOptions{}, func(ctxD context.Context, tx bun.Tx) error {
		_, err = tx.NewInsert().
			Model(&product).
			Returning("*").
			Exec(ctxD)
		if err != nil {
			cv.logger.Errorf("Failed to create product: %v", err)
			return err
		}
		var mid []models.ProductCategory
		for _, cat := range existPCs {
			mid = append(mid, models.ProductCategory{
				ProductID:  product.ID,
				CategoryID: cat.ID,
			})
		}
		_, err = tx.NewInsert().
			Model(&mid).
			Exec(ctxD)
		if err != nil {
			cv.logger.Errorf("Failed to create product category relation: %v", err)
			return err
		}
		return nil
	})
	if err != nil {
		return models.Product{}, err
	}
	product.Categories = existPCs
	product.City = city
	product.Supplier = supplier
	go func() {
		defer utils.Recovery()
		ctxD := context.Background()
		pipe := cv.cache.Pipeline()
		for _, cat := range existPCs {
			pipe.Incr(ctxD, fmt.Sprintf(CategoryProductKey, cat.ID.String()))
		}
		pipe.Incr(ctxD, fmt.Sprintf(SupplierProductKey, supplier.ID.String()))
		_, err = pipe.Exec(ctxD)
		if err != nil {
			cv.logger.Errorf("Failed to create product category relation: %v", err)
		}
	}()
	return product, nil
}

func (cv productSvc) diffCategories(existCt []models.Category, newCt []models.Category) (toRemove []models.Category, toInsert []models.Category) {
	// Create maps for faster lookup
	existMap := make(map[string]models.Category)
	newMap := make(map[string]models.Category)

	// Populate the maps with category names as keys
	for _, cat := range existCt {
		existMap[cat.ID.String()] = cat
	}

	for _, cat := range newCt {
		newMap[cat.ID.String()] = cat
	}

	// Find categories to remove (in existCt but not in newCt)
	for _, cat := range existCt {
		if _, exists := newMap[cat.ID.String()]; !exists {
			toRemove = append(toRemove, cat)
		}
	}

	// Find categories to insert (in newCt but not in existCt)
	for _, cat := range newCt {
		if _, exists := existMap[cat.ID.String()]; !exists {
			toInsert = append(toInsert, cat)
		}
	}

	return toRemove, toInsert
}

func (cv productSvc) UpdateProduct(ctx context.Context, req transforms.ProductUpdateReq) (models.Product, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	product, err := cv.Product(ctx, req.ProductReq.Id)
	if err != nil {
		return models.Product{}, err
	}

	if req.Name != "" {
		product.Name = req.Name
	}

	if req.Status != 0 {
		product.Status = models.ProductStatus(req.Status)
	}

	if req.Price != nil {
		product.Price = *req.Price
	}

	supplierOldId := product.SupplierId
	var city *models.City
	var supplier *models.Supplier
	var existPCs []models.Category
	var toRemove []models.Category
	var toInsert []models.Category
	wg := utils.NewWgGroup()
	wg.Go(func() error {
		if req.CityId != "" && req.CityId != product.CityId.String() {
			c, err := cv.citySvc.City(ctx, req.CityId)
			if err != nil {
				cv.logger.Errorf("Failed to get city: %v", err)
				return err
			}
			city = &c
		}
		return nil
	})
	wg.Go(func() error {
		if req.SupplierId != "" && req.SupplierId != product.SupplierId.String() {
			c, err := cv.supplierSvc.Supplier(ctx, req.CityId)
			if err != nil {
				cv.logger.Errorf("Failed to get city: %v", err)
				return err
			}
			supplier = &c
		}
		return nil
	})
	wg.Go(func() error {
		if len(req.Categories) > 0 {
			existPCs, err = cv.categorySvc.CategoriesByIds(ctx, req.Categories)
			if err != nil {
				cv.logger.Errorf("Failed to get categories: %v", err)
				return err
			}
			if len(existPCs) > 0 {
				toRemove, toInsert = cv.diffCategories(product.Categories, existPCs)
			}
		}
		return nil
	})
	err = wg.Wait()
	if err != nil {
		return models.Product{}, err
	}

	if city != nil {
		product.CityId = city.ID
	}
	if supplier != nil {
		product.SupplierId = supplier.ID
	}

	err = cv.db.RunInTx(ctx, &sql.TxOptions{}, func(ctxD context.Context, tx bun.Tx) error {
		_, err = tx.NewUpdate().
			Model(&product).
			Returning("*").
			Exec(ctxD)
		if err != nil {
			cv.logger.Errorf("Failed to update product: %v", err)
			return err
		}
		if len(toRemove) > 0 {
			var remove []models.ProductCategory
			for _, cat := range toRemove {
				remove = append(remove, models.ProductCategory{
					ProductID:  product.ID,
					CategoryID: cat.ID,
				})
			}
			_, err = tx.NewDelete().
				Model(&remove).
				WherePK().
				Exec(ctxD)
			if err != nil {
				cv.logger.Errorf("Failed to delete product category relation: %v", err)
				return err
			}
		}
		if len(toInsert) > 0 {
			var insert []models.ProductCategory
			for _, cat := range toInsert {
				insert = append(insert, models.ProductCategory{
					ProductID:  product.ID,
					CategoryID: cat.ID,
				})
			}
			_, err = tx.NewInsert().
				Model(&insert).
				Exec(ctxD)
			if err != nil {
				cv.logger.Errorf("Failed to create product category relation: %v", err)
				return err
			}
		}
		return nil
	})
	if err != nil {
		return models.Product{}, err
	}
	if len(existPCs) > 0 {
		product.Categories = existPCs
	}
	if city != nil {
		product.City = *city
	}
	if city != nil {
		product.Supplier = *supplier
	}
	go func() {
		defer utils.Recovery()
		pipe := cv.cache.Pipeline()
		if supplier != nil {
			pipe.Decr(ctx, fmt.Sprintf(SupplierProductKey, supplierOldId))
			pipe.Incr(ctx, fmt.Sprintf(SupplierProductKey, supplier.ID.String()))
		}
		for _, cat := range toRemove {
			pipe.Decr(ctx, fmt.Sprintf(CategoryProductKey, cat.ID.String()))
		}
		for _, cat := range toInsert {
			pipe.Incr(ctx, fmt.Sprintf(CategoryProductKey, cat.ID.String()))
		}
		_, err = pipe.Exec(context.Background())
		if err != nil {
			panic(err)
		}
	}()
	return product, nil
}

func (cv productSvc) DeleteProduct(ctx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	product, err := cv.Product(ctx, id)
	if err != nil {
		cv.logger.Errorf("Failed to delete product: %v", err)
		return err
	}
	err = cv.db.RunInTx(ctx, &sql.TxOptions{}, func(ctxD context.Context, tx bun.Tx) error {
		if len(product.Categories) > 0 {
			var remove []models.ProductCategory
			for _, cat := range product.Categories {
				remove = append(remove, models.ProductCategory{
					ProductID:  product.ID,
					CategoryID: cat.ID,
				})
			}
			_, err = tx.NewDelete().
				Model(&remove).
				WherePK().
				Exec(ctxD)
			if err != nil {
				cv.logger.Errorf("Failed to delete product category relation: %v", err)
				return err
			}
		}
		_, err = tx.NewDelete().
			Model(&product).
			WherePK().
			Exec(ctxD)
		if err != nil {
			cv.logger.Errorf("Failed to delete product: %v", err)
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
