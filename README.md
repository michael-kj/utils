
1. logger 
 - 是对zap进行了封装，使用全局单例Sugar Logger
 - gin automaxprocs gorm的log都替换成了zap
 - 如果用户不自己初始化的话，会自动初始化为info级别到标准输出的logger
 - SetUpLog 不会进行日志切分轮转
 - SetRotateLog v1会按照24小时进行轮转切分，v2按照1GB进行文件切分  目前是hardcode的  暂时没有给配置接口
 
2. prometheus
 - 默认注册了requests_total  request_duration_millisecond response_size_bytes request_size_bytes
 
3. pprof
 - 参见相关路由


### Examples
[具体使用，参考example文件](https://github.com/michael-kj/utils/blob/master/example/main.go)

```go
package main

import (
	"github.com/gin-gonic/gin"
	"github.com/michael-kj/utils"
	"github.com/michael-kj/utils/log"
	"github.com/michael-kj/utils/monitor"
	server "github.com/michael-kj/utils/server"
	"github.com/michael-kj/utils/storage"
	"time"
)

func SayHi(c *gin.Context) {
	log.Logger.Info("hi there,it's global middleware")
}
func MyGroupMiddleware(c *gin.Context) {
	log.Logger.Info("oh!It's my middleware")
}

type WorldService struct {
}

func init() {
	service := WorldService{}
	server.RegisterService(&service)
	//在这里注册自己的服务

}

func (s *WorldService) RegisterRouter() {
	//服务必须实现RegisterRouter，来注册路由

	g := server.GetGlobalGroup()
	// GlobalGroup为根："/"
	my := g.Group("/v2")
	my.GET("/hi", s.Hi)

	my.GET("/panic", func(c *gin.Context) {
		panic("An unexpected error happen!")
	})
}

func (s *WorldService) Hi(c *gin.Context) {
	time.Sleep(1 * time.Second)
	c.JSON(200, "world")
}

type HelloService struct {
}

func init() {
	service := HelloService{}
	server.RegisterService(&service)
	//在这里注册自己的服务

}

func (s *HelloService) Hi(c *gin.Context) {
	c.JSON(200, "hello")
}

func (s *HelloService) RegisterRouter() {
	//服务必须实现RegisterRouter，来注册路由

	g := server.GetGlobalGroup()
	g.Use(MyGroupMiddleware)
	// GlobalGroup为根："/"
	my := g.Group("/v1")
	my.GET("/hi", s.Hi)

	my.GET("/panic", func(c *gin.Context) {
		panic("An unexpected error happen!")
	})
}

func main() {

	utils.SetUpGoMaxProcs()
	// 自动设置cpu个数，主要应用于设置了资源限制的容器应用

	err := log.SetUpLog(log.Config{Format: "console", Level: "debug", Path: "/tmp/a.log", Development: false, DefaultFiled: nil})
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
	//mysqlConfig := storage.MysqlConfig{BaseConfig: storage.BaseConfig{User: "root", Password: "root", Host: "127.0.0.1", Port: 3306, Env: utils.Dev}, Database: "test", MaxIdle: 10, MaxOpen: 20}
	//storage.SetupMysql(mysqlConfig)
	//
	//err = storage.MysqlHealthCheck()
	//if err != nil {
	//	log.Logger.Error(err)
	//}

	//redisConfig := storage.RedisConfig{BaseConfig: storage.BaseConfig{User: "", Password: "", Host: "127.0.0.1", Port: 6379, Env: utils.Dev}, Database: 0, MinIdle: 10, MaxOpen: 20}
	//storage.SetUpRedis(redisConfig)
	//err = storage.Redis.Set(context.Background(), "a", 1, time.Minute).Err()
	//if err != nil {
	//	log.Logger.Error(err)
	//}

	//err = storage.RedisHealthCheck()
	//if err != nil {
	//	log.Logger.Error(err)
	//}

	server.SetGlobalGin(nil, utils.Online)
	// engine 为nil时候会自动初始化全局路由，除了online环境以外，开启debug模式

	r := server.GetGlobalEngine() //获取全局路由engine
	monitor.UsePprof(r)

	g := server.GetGlobalGroup() //获取全局根Group
	g.Use(SayHi)

	p := monitor.NewPrometheus("devops", "cmdb", "/metrics")
	p.Use(g)
	// 中间件是有顺序的  如果使用Prometheus  需要把Prometheus的中间件注册在gin log 之前
	r.Use(server.GinRecover())
	r.Use(server.GinLog())

	server.RunGraceful("127.0.0.1:8081", nil)
	// nil的时候会使用全局路由
	// 打开http://127.0.0.1:8081/api/hi

	storage.CloseStorage()
}

```
