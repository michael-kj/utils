package utils

import (
	"encoding/json"

	"github.com/kirinlabs/HttpRequest"
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
