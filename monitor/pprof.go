package monitor

import (
	"net/http"
	"net/http/pprof"

	"github.com/gin-gonic/gin"
)

func UsePprof(s gin.IRoutes) {
	s.GET("/index", pprofHandler(pprof.Index))
	s.GET("/cmdline", pprofHandler(pprof.Cmdline))
	s.GET("/profile", pprofHandler(pprof.Profile))
	s.POST("/symbol", pprofHandler(pprof.Symbol))
	s.GET("/symbol", pprofHandler(pprof.Symbol))
	s.GET("/trace", pprofHandler(pprof.Trace))
	s.GET("/allocs", pprofHandler(pprof.Handler("allocs").ServeHTTP))
	s.GET("/block", pprofHandler(pprof.Handler("block").ServeHTTP))
	s.GET("/goroutine", pprofHandler(pprof.Handler("goroutine").ServeHTTP))
	s.GET("/heap", pprofHandler(pprof.Handler("heap").ServeHTTP))
	s.GET("/mutex", pprofHandler(pprof.Handler("mutex").ServeHTTP))
	s.GET("/threadcreate", pprofHandler(pprof.Handler("threadcreate").ServeHTTP))

}

func pprofHandler(h http.HandlerFunc) gin.HandlerFunc {
	handler := http.HandlerFunc(h)
	return func(c *gin.Context) {
		handler.ServeHTTP(c.Writer, c.Request)
	}
}
