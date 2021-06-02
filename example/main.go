package main

import (
	"math/rand"
	"time"

	"go.uber.org/zap/zapcore"

	"github.com/gin-gonic/gin"
	"github.com/michael-kj/utils"
	"github.com/michael-kj/utils/log"
	"github.com/michael-kj/utils/monitor"
	server "github.com/michael-kj/utils/server"
	"github.com/michael-kj/utils/storage"
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
	g, _ := server.GetRegisteredGroup("/api/v1")
	my := g.Group("/world")
	my.GET("/", s.Hi)

	my.GET("/panic", func(c *gin.Context) {
		panic("An unexpected error happen!")
	})
}

func (s *WorldService) Hi(c *gin.Context) {
	time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
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

	g, _ := server.GetRegisteredGroup("/api/v1")
	my := g.Group("/hi")
	my.Use(MyGroupMiddleware)

	my.GET("/", s.Hi)

	my.GET("/panic", func(c *gin.Context) {
		panic("An unexpected error happen!")
	})
}
func skipLog(c *gin.Context) bool {
	path := c.Request.URL.Path
	return path == "/health_check" || path == "/metrics"
}

type Person struct {
	Age  int
	Name string
}

func main() {
	//log.SetUpLog(log.Config{Format: "json", Level: "debug", Path: "", Development: false, DefaultFiled: nil})
	err := log.SetRotateLog(log.Config{"json", "info", "/tmp/zz.log", true, nil}, "time")
	// SetUpLog创建的全局日志不会做切分轮转，SetRotateLog time会按照24小时进行轮转切分，chunk按照1GB进行文件切分
	if err != nil {
		panic(err.Error())
	}
	utils.SetUpGoMaxProcs()
	// 自动设置cpu个数，主要应用于设置了资源限制的容器应用

	log.Logger.Info("info")
	log.Logger.Warn("warn")
	log.Logger.Debug("debug")
	log.Logger.Error("err")
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

	// 注意中间件是有顺序的

	rootGroup, _ := server.GetRegisteredGroup("/")
	rootGroup.Use(server.GinRecover())
	rootGroup.Use(server.GinLog(skipLog))

	rootGroup.Use(SayHi)
	p := monitor.NewPrometheus("devops", "cmdb", "/metrics")
	rootGroup.GET("/log", func(c *gin.Context) {
		levelStr := c.DefaultQuery("level", "")
		var level zapcore.Level
		level.Set(levelStr)
		log.Level.SetLevel(level)
		c.JSON(200, gin.H{"status": "ok"})

	})

	server.RegisteredGroup("/api/v1", rootGroup)
	v1Group, _ := server.GetRegisteredGroup("/api/v1")

	p.Use(v1Group)

	monitor.UsePprof(v1Group)

	v1Group.GET("/ping", func(c *gin.Context) {
		log.Logger.Info("info")
		log.Logger.Warn("warn")
		log.Logger.Error("err")
		log.Logger.Debug("debug")
		log.Logger.Infow("info", "key", "value")
		c.JSON(200, gin.H{"status": "ok"})

	})

	v1Group.POST("/pong", func(c *gin.Context) {
		p := Person{}
		c.ShouldBindJSON(&p)
		c.JSON(200, gin.H{"person": p})

	})

	v1Group.POST("/panic", func(c *gin.Context) {
		panic("aaaa")

	})

	server.RunGraceful("127.0.0.1:8081", nil)
	// nil的时候会使用全局路由
	// 打开http://127.0.0.1:8081/api/v1/hi

	storage.CloseStorage()
}
