package utils

import (
	"encoding/json"
	"fmt"
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

func maxprocsLog(format string, v ...interface{}) {
	log.Logger.Info(fmt.Sprintf(format, v))

}

func SetUpGoMaxProcs() {

	maxprocs.Set(maxprocs.Logger(maxprocsLog))

}

type IntSlice []int

func (s IntSlice) Len() int           { return len(s) }
func (s IntSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s IntSlice) Less(i, j int) bool { return s[i] < s[j] }

type StringSlice []string

func (s StringSlice) Len() int           { return len(s) }
func (s StringSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s StringSlice) Less(i, j int) bool { return s[i] < s[j] }
