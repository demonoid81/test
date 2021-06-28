package utils

import (
	"bytes"
	"io/ioutil"
	"net/http"
)

func SendRequest(url string, method string, body []byte, query map[string]string) ([]byte, error) {

	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	if len(query) > 0 {
		q := req.URL.Query()
		for k, v := range query {
			q.Add(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	rBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return rBody, nil
}
