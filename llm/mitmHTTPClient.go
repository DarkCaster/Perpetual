package llm

import (
	"bytes"
	"io"
	"net/http"

	"github.com/DarkCaster/Perpetual/utils"
)

func NewMitmHTTPClient(jsonToInject string) *http.Client {
	return &http.Client{ //nolint:gochecknoglobals
		Transport: &mitmTransport{
			Transport:    http.DefaultTransport,
			JsonToInject: jsonToInject,
		},
	}
}

type mitmTransport struct {
	Transport    http.RoundTripper
	JsonToInject string
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
	//TODO: (not now) Combine injected JSON with body
	newBody := /*t.JsonToInject + */ string(bodyData)
	// Create new ReadCloser with modified body
	req.Body = io.NopCloser(bytes.NewReader([]byte(newBody)))
	return t.Transport.RoundTrip(req)
}
