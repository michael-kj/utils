package utils

import (
	"encoding/json"
	"fmt"
	"net/http"

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
