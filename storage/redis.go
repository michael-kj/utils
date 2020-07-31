package storage

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/michael-kj/utils/log"
)

var Redis *redis.Client

func SetUpRedis(config Config) {
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	database, err := strconv.Atoi(config.Database)
	if err != nil {
		log.Logger.Fatalw("wrong redis database setting", "err", err, "database", config.Database)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: config.Password,
		DB:       database,
	})
	Redis = rdb
	err = RedisHealthCheck()
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
