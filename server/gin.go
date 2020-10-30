//go:generate statik -src=./static/swagger -f -p statik -dest ./static

package server

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"text/template"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/michael-kj/utils"
	"github.com/michael-kj/utils/log"
	_ "github.com/michael-kj/utils/static/statik"
	"github.com/rakyll/statik/fs"
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

func buildPath(c *gin.Context) string {
	path := c.Request.URL.Path
	if c.Request.URL.RawQuery != "" {
		path = fmt.Sprintf("%s?%s", c.Request.URL.Path, c.Request.URL.RawQuery)
	}
	return path
}
func buildBody(c *gin.Context) string {
	body, err := c.GetRawData()
	if err != nil {
		body = []byte("err when get request body ")
	}
	c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	return string(body)
}

func GinLog(skip func(c *gin.Context) bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if skip(c) {
			c.Next()
		} else {
			start := time.Now()
			path := buildPath(c)
			body := buildBody(c)
			c.Next()

			end := time.Now()
			latency := end.Sub(start)

			if len(c.Errors) > 0 {
				for _, e := range c.Errors.Errors() {
					log.Logger.Desugar().Error(e)
				}
			} else {
				log.Logger.Desugar().Info(path,
					zap.Int("status", c.Writer.Status()),
					zap.String("method", c.Request.Method),
					zap.String("ip", c.ClientIP()),
					zap.String("latency", latency.String()),
					zap.String("body", body),

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
					body := buildBody(c)
					path := buildPath(c)
					log.Logger.Desugar().Error(path,
						zap.Any("error", err),
						zap.Int("status", c.Writer.Status()),
						zap.String("method", c.Request.Method),
						zap.String("ip", c.ClientIP()),
						zap.String("user-agent", c.Request.UserAgent()),
						zap.String("body", body),
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

var Doc = `<!-- HTML for static distribution bundle build -->
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Swagger UI</title>
  <link rel="stylesheet" type="text/css" href="{{.StaticPath}}swagger-ui.css" >
  <link rel="icon" type="image/png" href="{{.StaticPath}}favicon-32x32.png" sizes="32x32" />
  <link rel="icon" type="image/png" href="{{.StaticPath}}favicon-16x16.png" sizes="16x16" />
  <style>
    html
    {
      box-sizing: border-box;
      overflow: -moz-scrollbars-vertical;
      overflow-y: scroll;
    }

    *,
    *:before,
    *:after
    {
      box-sizing: inherit;
    }

    body
    {
      margin:0;
      background: #fafafa;
    }
  </style>
</head>

<body>
<div id="swagger-ui"></div>

<script src="{{.StaticPath}}swagger-ui-bundle.js" charset="UTF-8"> </script>
<script src="{{.StaticPath}}swagger-ui-standalone-preset.js" charset="UTF-8"> </script>
<script>
  window.onload = function() {
    // Begin Swagger UI call region
    const ui = SwaggerUIBundle({
      url: "{{.JsonDocPath}}",
      dom_id: '#swagger-ui',
      deepLinking: true,
      presets: [
        SwaggerUIBundle.presets.apis,
        SwaggerUIStandalonePreset
      ],
      plugins: [
        SwaggerUIBundle.plugins.DownloadUrl
      ],
      layout: "StandaloneLayout"
    })
    // End Swagger UI call region

    window.ui = ui
  }
</script>
</body>
</html>

`

func loadSwaggerFile(router *gin.RouterGroup) {
	statikFS, err := fs.New()
	if err != nil {
		fmt.Println(err)
	}
	router.StaticFS("/static/swagger", statikFS)

}

func RegisterDoc(router *gin.RouterGroup, routerPath string, host string, JsonDocPath string) {
	loadSwaggerFile(router)
	router.GET(routerPath, func(c *gin.Context) {
		tmpl, _ := template.New("docIndex").Parse(Doc)
		err := tmpl.Execute(c.Writer, struct {
			StaticPath  string
			JsonDocPath string
		}{StaticPath: host + "/static/swagger/",
			JsonDocPath: JsonDocPath})
		if err != nil {
			fmt.Println(err)
		}
	})

}
