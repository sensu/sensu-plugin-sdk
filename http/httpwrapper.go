package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

type HttpWrapper struct {
	httpClient *http.Client
	username   string
	password   string
}

func NewHttpWrapper(timeout uint64, proxy string, username string, password string) (*HttpWrapper, error) {
	var httpClient *http.Client

	if len(proxy) > 0 {
		proxyUrl, err := url.Parse(proxy)
		if err != nil {
			return nil, fmt.Errorf("error parsing proxy host %s: %s", proxy, err)
		}
		httpClient = &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(proxyUrl),
			},
			Timeout: time.Second * time.Duration(timeout),
		}
	} else {
		httpClient = &http.Client{
			Timeout: time.Second * time.Duration(timeout),
		}
	}

	return &HttpWrapper{
		httpClient,
		username,
		password,
	}, nil
}

// ExecuteRequest executes an HTTP request and parses the result JSON document into the result interface
// method specifies the HTTP method to use
// url specifies the URL to call
// body interface is serialized into a JSON string and sent in the HTTP body
// result will contain the HTTP call result deserialized from its JSON string
func (httpWrapper *HttpWrapper) ExecuteRequest(method string, url string, body interface{}, result interface{}) (int, string, error) {

	var bodyBytes []byte
	var bodyReader io.Reader
	var err error

	if body != nil {
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return 0, "", fmt.Errorf("error marshalling body to json: %s", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	request, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return 0, "", fmt.Errorf("error building request: %s", err)
	}

	request.Header.Add("Content-Type", "application/json")
	request.SetBasicAuth(httpWrapper.username, httpWrapper.password)

	response, err := httpWrapper.httpClient.Do(request)
	if err != nil {
		return 0, "", fmt.Errorf("http error: %s", err)
	}

	defer response.Body.Close()
	buf, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return 0, "", fmt.Errorf("error reading response body: %s", err)
	}

	log.Println("HTTP response status code:", response.StatusCode)

	resultJson := string(buf)
	log.Print("HTTP response JSON:", resultJson)
	if result != nil {
		err = json.Unmarshal(buf, result)
		if err != nil {
			return response.StatusCode, resultJson, fmt.Errorf("error unmarshalling json: %s", err)
		}
	}

	return response.StatusCode, resultJson, nil
}
