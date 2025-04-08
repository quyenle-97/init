package cache

import (
	"context"
	"fmt"
	"github.com/Minh2009/pv_soa/pkgs/log"
	"github.com/redis/go-redis/v9"
)

type RConfig struct {
	Host    string   `json:"REDIS_HOST"` // redis host
	Port    int      `json:"REDIS_PORT"`
	Pass    string   `json:"REDIS_PASS"`    // redis pass
	Index   int      `json:"REDIS_INDEX"`   // redis index
	Addr    []string `json:"REDIS_ADDR"`    // redis addr
	Cluster bool     `json:"REDIS_CLUSTER"` // redis cluster
}

func NewRedis(c RConfig, log *log.MultiLogger) (redis.UniversalClient, error) {
	//var c RConfig
	//err := utils.BindStruct[RConfig](cfg, &c)
	//if err != nil {
	//	panic(err)
	//}
	addr := c.Addr
	if len(addr) == 0 {
		addr = []string{fmt.Sprintf("%s:%d", c.Host, c.Port)}
	}
	opts := redis.UniversalOptions{
		Addrs:    addr,
		Password: c.Pass,
		DB:       c.Index,
	}
	cluster := c.Cluster || len(addr) > 1

	var client redis.UniversalClient
	if cluster {
		client = redis.NewClusterClient(opts.Cluster())
	} else {
		client = redis.NewUniversalClient(&opts)
	}

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %v", err)
	}

	return client, nil
}
