package storage

import (
	"github.com/michael-kj/utils"
	"github.com/michael-kj/utils/log"
)

type BaseConfig struct {
	User     string    `json:"user"`
	Password string    `json:"password"`
	Host     string    `json:"host"`
	Port     int       `json:"port"`
	Env      utils.Env `json:"env"`
}

type MysqlConfig struct {
	BaseConfig
	Database string `json:"database"`
	MaxIdle  int    `json:"maxIdle,omitempty"`
	MaxOpen  int    `json:"maxOpen,omitempty"`
}

type RedisConfig struct {
	BaseConfig
	Database int `json:"database"`
	MinIdle  int `json:"minIdle,omitempty"`
	MaxOpen  int `json:"maxOpen,omitempty"`
}

func CloseStorage() {
	if Db != nil {
		err := Db.Close()
		if err != nil {
			log.Logger.Warnw("err when shutdown mysql connection", "err", err)

		}
		log.Logger.Infow("mysql client closed")

	}
	if Redis != nil {
		err := Redis.Close()
		if err != nil {
			log.Logger.Warnw("err when shutdown redis connection", "err", err)

		}
		log.Logger.Infow("redis client closed")

	}
	log.Logger.Infow("all storage closed")

}
