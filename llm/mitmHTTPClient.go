package llm

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/DarkCaster/Perpetual/utils"
)

type requestTransformer interface {
	ProcessBody(body map[string]interface{}) map[string]interface{}
	ProcessHeader(header http.Header) http.Header
}

type bodyValuesInjector struct {
	ValuesToInject map[string]interface{}
}

func newTopLevelBodyValuesInjector(valuesToInject map[string]interface{}) requestTransformer {
	return &bodyValuesInjector{
		ValuesToInject: valuesToInject,
	}
}

func (p *bodyValuesInjector) ProcessBody(body map[string]interface{}) map[string]interface{} {
	// inject new values to body json at the top level
	for name, value := range p.ValuesToInject {
		body[name] = value
	}
	return body
}

func (p *bodyValuesInjector) ProcessHeader(header http.Header) http.Header {
	// No header modifications for this transformer
	return header
}

func newMitmHTTPClient(transformers ...requestTransformer) *http.Client {
	return &http.Client{
		Transport: &mitmTransport{
			Transport:    http.DefaultTransport,
			Transformers: transformers,
		},
	}
}

type mitmTransport struct {
	Transport    http.RoundTripper
	Transformers []requestTransformer
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
	// apply request transformations
	for _, transformer := range t.Transformers {
		bodyObj = transformer.ProcessBody(bodyObj)
		req.Header = transformer.ProcessHeader(req.Header)
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
