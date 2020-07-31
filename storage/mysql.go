package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/michael-kj/utils"
	"github.com/michael-kj/utils/log"
)

var Db *gorm.DB

type Base struct {
	ID int      `json:"id,omitempty" gorm:"primary_key"`
	Db *gorm.DB `json:"-" gorm:"-" swaggerignore:"true"`
}

func (b *Base) DB() *gorm.DB {
	//用于事务支持，提高函数复用，单语句使用全局Db，自动commit
	if b.Db == nil {
		return Db
	} else {
		return b.Db
	}
}

type Time struct {
	CreatedAt *time.Time `json:"createdAt,omitempty" gorm:"column:created_at" swaggerignore:"true"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty" gorm:"column:updated_at" swaggerignore:"true"`
}

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

type GormLogger struct{}

func (*GormLogger) Print(v ...interface{}) {
	switch v[0] {
	case "sql":
		log.Logger.Debug(v)
	case "log":
		log.Logger.Error(v)
	}
}

func SetupMysql(config Config) {

	gorm.DefaultCallback.Create().Remove("gorm:update_time_stamp")
	gorm.DefaultCallback.Update().Remove("gorm:update_time_stamp")

	dbUrl := fmt.Sprintf("%s:%s@(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local", config.User, config.Password, config.Host, config.Port, config.Database)
	db, err := gorm.Open("storage", dbUrl)
	if err != nil {
		log.Logger.Fatalw("err when open connection to storage ", "err", err)
	}

	env := config.Env
	err = utils.CheckEnv(env)
	if err != nil {
		log.Logger.Fatalw("err wrong env value ", "err", err)
	}
	gormLog := GormLogger{}
	db.SetLogger(&gormLog)
	if config.MaxIdle != 0 {
		db.DB().SetMaxIdleConns(config.MaxIdle)
	}
	if config.MaxOpen != 0 {
		db.DB().SetMaxOpenConns(config.MaxIdle)
	}

	if env != utils.Online {
		db.LogMode(true)
	}

	Db = db

	err = MysqlHealthCheck()
	if err != nil {
		log.Logger.Fatalw("err when init storage and  ping storage server", "err", err)
	}
	config.Password = "*******"
	log.Logger.Infow("create storage client success", "config", config)
}

func MysqlHealthCheck() error {
	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	return Db.DB().PingContext(ctx)
}
