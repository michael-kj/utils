package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/michael-kj/utils/log"
)

var Redis *redis.Client

func SetUpRedis(config RedisConfig) {
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)

	rdb := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     config.Password,
		DB:           config.Database,
		PoolSize:     config.MaxOpen,
		MinIdleConns: config.MinIdle,
	})
	Redis = rdb
	err := RedisHealthCheck()
	if err != nil {
		config.Password = "******"
		log.Logger.Fatalw("err when connect redis ", "err", err, "config", config)
	}
}

func RedisHealthCheck() error {
	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	_, err := Redis.Ping(ctx).Result()
	return err

}
