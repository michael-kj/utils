package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kirinlabs/HttpRequest"
	"github.com/michael-kj/utils/log"
	"go.uber.org/automaxprocs/maxprocs"
)

func Post(url string, request, response interface{}, header map[string]string) error {
	req := HttpRequest.NewRequest()
	if header != nil {
		req.SetHeaders(header)
	}
	jsonBytes, err := json.Marshal(request)
	if err != nil {
		return err
	}

	res, err := req.JSON().Post(url, string(jsonBytes))
	if err != nil {
		return err
	}

	body, _ := res.Body()
	err = json.Unmarshal(body, response)
	if err != nil {
		return err
	}
	return nil

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

func maxprocsLog(format string, v ...interface{}) {
	log.Logger.Info(fmt.Sprintf(format, v))

}

func SetUpGoMaxProcs() {

	maxprocs.Set(maxprocs.Logger(maxprocsLog))

}

func RunGraceful(addr string, handler http.Handler) {
	srv := &http.Server{
		Addr:    addr,
		Handler: handler,
	}
	go func() {
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

func SetGinMode(env Env) {
	switch env {
	case Online:
		gin.SetMode(gin.ReleaseMode)
	default:
		gin.SetMode(gin.DebugMode)

	}
}
