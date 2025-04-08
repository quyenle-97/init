package services

import (
	"context"
	"github.com/Minh2009/pv_soa/pkgs/log"
	"github.com/redis/go-redis/v9"
	"github.com/uptrace/bun"
	"strconv"
	"strings"
	"time"
)

type StatisticsSvc interface {
	BySupplier(ctx context.Context) (map[string]int64, error)
	ByCategories(ctx context.Context) (map[string]int64, error)
}

type statisticsSvc struct {
	db     *bun.DB
	cache  redis.UniversalClient
	logger *log.MultiLogger
}

func NewStatisticsSvc(db *bun.DB, cache redis.UniversalClient, logger *log.MultiLogger) StatisticsSvc {
	return &statisticsSvc{
		db:     db,
		cache:  cache,
		logger: logger,
	}
}

func (cv statisticsSvc) ByCategories(ctx context.Context) (map[string]int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rs := make(map[string]int64)
	s := make(map[string]int64)
	var total int64

	var cursor uint64
	for {
		var keys []string
		var err error
		keys, cursor, err = cv.cache.Scan(ctx, cursor, "category_product:*", 100).Result()
		if err != nil {
			continue
		}

		for _, key := range keys {
			cID := strings.Split(key, ":")[1]
			data, err := cv.cache.Get(ctx, key).Result()
			if err != nil {
				continue
			}
			d, er := strconv.ParseInt(data, 10, 64)
			if er != nil {
				continue
			}
			rs[cID] = d
			total = total + d
		}

		if cursor == 0 {
			break
		}
	}
	if total == 0 {
		return s, nil
	}
	for k, v := range rs {
		s[k] = v * 100 / total
	}
	return s, nil
}

func (cv statisticsSvc) BySupplier(ctx context.Context) (map[string]int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rs := make(map[string]int64)
	s := make(map[string]int64)
	var total int64

	var cursor uint64
	for {
		var keys []string
		var err error
		keys, cursor, err = cv.cache.Scan(ctx, cursor, "supplier_product:*", 100).Result()
		if err != nil {
			continue
		}

		for _, key := range keys {
			cID := strings.Split(key, ":")[1]
			data, err := cv.cache.Get(ctx, key).Result()
			if err != nil {
				continue
			}
			d, er := strconv.ParseInt(data, 10, 64)
			if er != nil {
				continue
			}
			rs[cID] = d
			total = total + d
		}

		if cursor == 0 {
			break
		}
	}
	if total == 0 {
		return s, nil
	}
	for k, v := range rs {
		s[k] = v * 100 / total
	}
	return s, nil
}
