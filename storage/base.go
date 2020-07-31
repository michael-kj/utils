package storage

import (
	"github.com/michael-kj/utils"
	"github.com/michael-kj/utils/log"
)

type Config struct {
	User     string
	Password string
	Host     string
	Port     int
	Database string
	Env      utils.Env
	MaxIdle  int
	MaxOpen  int
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
