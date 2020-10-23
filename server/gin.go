package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/michael-kj/utils"
	"github.com/michael-kj/utils/log"
	"go.uber.org/zap"
)

var e *gin.Engine
var serviceRegister []GinServiceInterface
var NotRegisteredErr = errors.New("router group not registered")
var gs = routerGroups{groups: map[string]*gin.RouterGroup{}}

type routerGroups struct {
	groups map[string]*gin.RouterGroup
	lock   sync.RWMutex
}

type GinServiceInterface interface {
	RegisterRouter()
}

func initRouter() {
	for _, service := range serviceRegister {
		service.RegisterRouter()
	}
}

func RegisterService(service GinServiceInterface) {
	serviceRegister = append(serviceRegister, service)
}

func GinLog(skip func(c *gin.Context) bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := fmt.Sprintf("%s?%s", c.Request.URL.Path, c.Request.URL.RawQuery)

		c.Next()
		if !skip(c) {

			end := time.Now()
			latency := end.Sub(start)

			if len(c.Errors) > 0 {
				for _, e := range c.Errors.Errors() {
					log.Logger.Desugar().Error(e)
				}
			} else {
				Cyan := 36
				data := fmt.Sprintf("\x1b[%dm%s\x1b[0m", uint8(Cyan), path)
				body, err := c.GetRawData()
				if err != nil {
					body = []byte("err when get request body ")
				}
				log.Logger.Desugar().Info(data,
					zap.Int("status", c.Writer.Status()),
					zap.String("method", c.Request.Method),
					zap.String("ip", c.ClientIP()),
					zap.String("time", end.Format(time.RFC3339)),
					zap.String("latency", latency.String()),
					zap.String("body", string(body)),

					//zap.String("user-agent", c.Request.UserAgent()),
					//zap.Any("header",c.Request.Header),
				)
			}
		}
	}
}

func GinRecover() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Check for a broken connection, as it is not really a
				// condition that warrants a panic f trace.
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				if brokenPipe {
					log.Logger.Desugar().Error("broken connection", zap.Any("err", err))
				} else {
					log.Logger.Desugar().Error(c.Request.URL.Path,
						zap.Any("error", err),
						//zap.Int("status", c.Writer.Status()),
						zap.String("method", c.Request.Method),
						zap.String("path", c.Request.URL.Path),
						zap.String("query", c.Request.URL.RawQuery),
						zap.String("ip", c.ClientIP()),
						zap.String("user-agent", c.Request.UserAgent()),
					)
				}

				// If the connection is dead, we can't write a status to it.
				if brokenPipe {
					c.Error(err.(error)) // nolint: errcheck
					c.Abort()
				} else {
					c.AbortWithStatus(http.StatusInternalServerError)
				}
			}
		}()
		c.Next()

	}
}

func SetGlobalGin(engine *gin.Engine, env utils.Env) {
	// 如果使用自定义的engine的话，自己处理gin.SetMode，server.SetMode必须放在gin.New初始化之前
	if engine == nil {
		SetGinMode(env)
		e = gin.New()
	} else {
		e = engine
	}
}

func GetGlobalEngine() *gin.Engine {
	return e

}

func GetRegisteredGroup(path string) (*gin.RouterGroup, error) {
	if path == "/" {
		return &e.RouterGroup, nil
	}

	gs.lock.RLock()
	defer gs.lock.RUnlock()
	g, ok := gs.groups[path]
	if !ok {
		return nil, NotRegisteredErr
	}
	return g, nil

}

func RegisteredGroup(path string, baseGroup *gin.RouterGroup) {
	if baseGroup == nil {
		baseGroup = &e.RouterGroup
	}
	gs.lock.Lock()
	gs.groups[path] = baseGroup.Group(path)
	gs.lock.Unlock()
}

func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method

		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token, Authorization, Token")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")

		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		c.Next()
	}
}

func RunGraceful(addr string, engine http.Handler) {
	if engine == nil {
		engine = GetGlobalEngine()
	}
	initRouter()
	srv := &http.Server{
		Addr:    addr,
		Handler: engine,
	}
	go func() {
		log.Logger.Infow("Server start", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Logger.Fatalw("start service failed", "err", err)
		}
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Logger.Infow("Shutting down server...")

	var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Logger.Fatalw("Server forced to shutdown", "err", err)
	}

	log.Logger.Infow("Server stopped")
}

func SetGinMode(env utils.Env) {
	switch env {
	case utils.Online:
		gin.SetMode(gin.ReleaseMode)
	default:
		gin.SetMode(gin.DebugMode)

	}
}
