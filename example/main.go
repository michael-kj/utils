package main

import (
	"github.com/gin-gonic/gin"
	"github.com/michael-kj/utils"
	"github.com/michael-kj/utils/log"
	"github.com/michael-kj/utils/storage"
)

type HelloService struct {
}

func init() {
	service := HelloService{}
	utils.RegisterService(&service)

}

func (s *HelloService) Hi(c *gin.Context) {
	c.JSON(200, "hi")
}

func (s *HelloService) RegisterRouter() {
	g := utils.GetGlobalGroup()
	// GlobalGroup为根："/"
	my := g.Group("/api")
	my.GET("/hi", s.Hi)

	my.GET("/panic", func(c *gin.Context) {
		panic("An unexpected error happen!")
	})
}

func main() {

	utils.SetUpGoMaxProcs()

	err := log.SetUpLog(log.Config{Format: "console", Level: "debug", Path: "/tmp/a.log", Development: true, DefaultFiled: nil})
	//err := log.SetRotateLog(log.Config{"console", "info", "/tmp/a.log", true, nil}, "v2")
	// SetUpLog创建的全局日志不会做切分轮转，SetRotateLog v1会按照24小时进行轮转切分，v2按照1GB进行文件切分
	if err != nil {
		log.Logger.Error(err.Error())
	}

	log.Logger.Info("info")
	log.Logger.Warn("warn")
	log.Logger.Debug("debug")
	log.Logger.Infow("info", "key", "value")
	// SetupMysql,SetUpRedis会检查mysql,redis链接，如果失败会os.exist(1)
	//mysqlConfig:=storage.Config{User: "root",Password: "root",Database: "test",Host: "127.0.0.1",Port: 3306,Env: utils.Dev,MaxIdle: 10,MaxOpen: 20}
	//storage.SetupMysql(mysqlConfig)

	//err=storage.MysqlHealthCheck()
	//if err != nil {
	//	log.Logger.Error(err)
	//}
	//redisConfig:=storage.Config{User: "root",Password: "root",Database: "0",Host: "127.0.0.1",Port: 6379,Env: utils.Dev,MaxIdle: 10,MaxOpen: 20}
	//storage.SetUpRedis(redisConfig)

	//err=storage.RedisHealthCheck()
	//if err != nil {
	//	log.Logger.Error(err)
	//}

	utils.SetGlobalGin(nil, utils.Dev)
	// engine 为nil时候会自动初始化全局路由，除了online环境以外，开启debug模式

	//r:=utils.GetGlobalEngine()   获取全局路由

	utils.RunGraceful("127.0.0.1:8081", nil)
	// nil的时候会使用全局路由

	storage.CloseStorage()
}
