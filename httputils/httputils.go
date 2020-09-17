package httputils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/mercari/testdeck/constants"
	"github.com/pkg/errors"
	"net/http"
	"net/url"
)

/*
httputils.go: Helper methods for testing HTTP endpoints
*/

// Converts a x-www-form-urlencoded form into a bytes buffer
func ToBufferArray(form url.Values) *bytes.Buffer {
	var b bytes.Buffer
	b.Write([]byte(form.Encode()))
	return &b
}

// Creates an HTTP request with the specified headers, http data, etc.
func SendHTTPRequest(method string, url string, body *bytes.Buffer, headers map[string]string, host ...string) (*http.Response, *bytes.Buffer, error) {

	client := &http.Client{
		Timeout: constants.DefaultHttpTimeout,
	}
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, nil, err
	}

	// if host param was passed in, add it to the request
	if len(host) > 0 {
		req.Host = host[0]
	}

	// add headers if specified
	for key, element := range headers {
		req.Header.Add(key, element)
	}

	// send request
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}

	defer resp.Body.Close()

	// export json response
	responseBody := &bytes.Buffer{}
	_, err = responseBody.ReadFrom(resp.Body)
	if err != nil {
		return resp, nil, err
	}

	return resp, responseBody, nil
}

// Returns the value of a specified field in the json data
func GetJSONField(fieldName string, jsonBody *bytes.Buffer) (string, error) {
	body := make(map[string]interface{})
	_ = json.Unmarshal(jsonBody.Bytes(), &body)
	dataObject := body["data"]

	if dataObject != nil {
		// return value in nested json http if it exists
		if body, _ := dataObject.(map[string]interface{}); body != nil {
			value, ok := body[fieldName].(string)
			if ok {
				return value, nil
			}
		}
	} else {
		// otherwise return value from http
		value, ok := body[fieldName].(string)
		if ok {
			return value, nil
		}
	}

	return "", errors.New("JSON does not contain the specified field")
}

// Connect to a debugging proxy (Burp Suite, Charles, etc.)
func ConnectToProxy(ip string) error {
	proxyUrl, err := url.Parse(ip)
	if err != nil {
		return err
	}
	http.DefaultTransport = &http.Transport{Proxy: http.ProxyURL(proxyUrl)}
	fmt.Println("Connected to debugging proxy...")
	return nil
}
