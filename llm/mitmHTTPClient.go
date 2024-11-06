package llm

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/DarkCaster/Perpetual/utils"
)

func NewMitmHTTPClient(valuesToInject map[string]interface{}) *http.Client {
	return &http.Client{
		Transport: &mitmTransport{
			Transport:      http.DefaultTransport,
			ValuesToInject: valuesToInject,
		},
	}
}

type mitmTransport struct {
	Transport      http.RoundTripper
	ValuesToInject map[string]interface{}
}

func (t *mitmTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Read request body if present
	var bodyData []byte
	var err error
	if req.Body != nil {
		bodyData, err = io.ReadAll(req.Body)
		req.Body.Close()
		if err != nil {
			return nil, err
		}
	} else {
		return t.Transport.RoundTrip(req)
	}
	// Check and convert body to string
	if len(bodyData) > 0 {
		if err = utils.CheckUTF8(bodyData); err != nil {
			return nil, err
		}
	} else {
		return t.Transport.RoundTrip(req)
	}
	bodyObj := map[string]interface{}{}
	if err := json.Unmarshal(bodyData, &bodyObj); err != nil {
		return nil, err
	}
	// inject new values to body json at the top level
	for name, value := range t.ValuesToInject {
		bodyObj[name] = value
	}
	// convert modified body back into JSON
	newBody, err := json.Marshal(bodyObj)
	if err != nil {
		return nil, err
	}
	// Create new ReadCloser with modified body
	req.Body = io.NopCloser(bytes.NewReader(newBody))
	req.ContentLength = int64(len(newBody))
	// Perform actual http request with new body
	return t.Transport.RoundTrip(req)
}
