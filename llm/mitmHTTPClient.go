package llm

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"slices"

	"github.com/DarkCaster/Perpetual/utils"
)

type requestTransformer interface {
	ProcessURL(url string) string
	ProcessBody(body map[string]interface{}) map[string]interface{}
	ProcessHeader(header http.Header) http.Header
}

type bodyValuesRemover struct {
	Path           []string
	ValuesToRemove []string
}

func newTopLevelBodyValuesRemover(valuesToRemove []string) requestTransformer {
	return &bodyValuesRemover{
		Path:           []string{},
		ValuesToRemove: valuesToRemove,
	}
}

func newInnerBodyValuesRemover(path, valuesToRemove []string) requestTransformer {
	return &bodyValuesRemover{
		Path:           path,
		ValuesToRemove: valuesToRemove,
	}
}

func (p *bodyValuesRemover) ProcessBody(body map[string]interface{}) map[string]interface{} {
	current := body
	// Navigate down the path
	for i, key := range p.Path {
		if val, ok := current[key].(map[string]interface{}); ok {
			current = val
		} else if i < len(p.Path)-1 {
			// Path doesn't exist, return original
			return body
		}
	}
	// Remove specified values at current level
	for _, key := range p.ValuesToRemove {
		delete(current, key)
	}
	return body
}

func (p *bodyValuesRemover) ProcessHeader(header http.Header) http.Header {
	// No header modifications for this transformer
	return header
}

func (p *bodyValuesRemover) ProcessURL(url string) string {
	return ""
}

type bodyValuesRenamer struct {
	Path    []string
	OldName string
	NewName string
}

func newTopLevelBodyValueRenamer(oldName, newName string) requestTransformer {
	return &bodyValuesRenamer{
		Path:    []string{},
		OldName: oldName,
		NewName: newName,
	}
}

func newInnerBodyValueRenamer(path []string, oldName, newName string) requestTransformer {
	return &bodyValuesRenamer{
		Path:    path,
		OldName: oldName,
		NewName: newName,
	}
}

func (p *bodyValuesRenamer) ProcessBody(body map[string]interface{}) map[string]interface{} {
	current := body
	// Navigate down the path
	for i, key := range p.Path {
		if val, ok := current[key].(map[string]interface{}); ok {
			current = val
		} else if i < len(p.Path)-1 {
			// Path doesn't exist, return original
			return body
		}
	}
	// Rename specified value at current level
	if value, ok := current[p.OldName]; ok {
		delete(current, p.OldName)
		current[p.NewName] = value
	}
	return body
}

func (p *bodyValuesRenamer) ProcessHeader(header http.Header) http.Header {
	// No header modifications for this transformer
	return header
}

func (p *bodyValuesRenamer) ProcessURL(url string) string {
	return ""
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

func (p *bodyValuesInjector) ProcessURL(url string) string {
	return ""
}

type basicAuthTransformer struct {
	Auth string
}

func newBasicAuthTransformer(auth string) requestTransformer {
	return &basicAuthTransformer{
		Auth: auth,
	}
}

func (p *basicAuthTransformer) ProcessBody(body map[string]interface{}) map[string]interface{} {
	// No body modifications for this transformer
	return body
}

func (p *basicAuthTransformer) ProcessHeader(header http.Header) http.Header {
	// Remove existing Authorization header
	header.Del("Authorization")
	// Add new Authorization header if Auth is not empty
	if len(p.Auth) > 0 {
		encodedAuth := base64.StdEncoding.EncodeToString([]byte(p.Auth))
		header.Set("Authorization", "Basic "+encodedAuth)
	}
	return header
}

func (p *basicAuthTransformer) ProcessURL(url string) string {
	return ""
}

type tokenAuthTransformer struct {
	Token string
}

func newTokenAuthTransformer(token string) requestTransformer {
	return &tokenAuthTransformer{
		Token: token,
	}
}

func (p *tokenAuthTransformer) ProcessBody(body map[string]interface{}) map[string]interface{} {
	// No body modifications for this transformer
	return body
}

func (p *tokenAuthTransformer) ProcessHeader(header http.Header) http.Header {
	// Remove existing Authorization header
	header.Del("Authorization")
	// Add new Authorization header if Token present
	if len(p.Token) > 0 {
		header.Set("Authorization", "Bearer "+p.Token)
	}
	return header
}

func (p *tokenAuthTransformer) ProcessURL(url string) string {
	return ""
}

type systemMessageTransformer struct {
	ChangeTo string
	ExtraAck string
}

func newSystemMessageTransformer(newSystemMessageRole, extraAcknowledge string) requestTransformer {
	return &systemMessageTransformer{
		ChangeTo: newSystemMessageRole,
		ExtraAck: extraAcknowledge,
	}
}

func (p *systemMessageTransformer) ProcessBody(body map[string]interface{}) map[string]interface{} {
	// find messages array to mess with
	iMessages, exist := body["messages"].([]interface{})
	if !exist {
		return body
	}
	// deserialze each message and find system prompt position
	var messages []map[string]interface{}
	sysMsgIdx := -1
	for i, imsg := range iMessages {
		msg := imsg.(map[string]interface{})
		if msg["role"] == "system" {
			sysMsgIdx = i
			//do not add system message if it's new role is empty
			if p.ChangeTo == "" {
				continue
			}
		}
		messages = append(messages, msg)
	}
	if sysMsgIdx < 0 {
		return body
	}
	//only change system message role if it is not empty
	if p.ChangeTo != "" {
		// convert system message into the provided message type
		messages[sysMsgIdx]["role"] = p.ChangeTo
		// insert extra acknowledge message as "assistant" message if needed
		if p.ExtraAck != "" {
			messages = slices.Insert(messages, sysMsgIdx+1, map[string]interface{}{"role": "assistant", "content": p.ExtraAck})
		}
	}
	//set new messages object to body
	body["messages"] = messages
	return body
}

func (p *systemMessageTransformer) ProcessHeader(header http.Header) http.Header {
	// No header modifications for this transformer
	return header
}

func (p *systemMessageTransformer) ProcessURL(url string) string {
	return ""
}

type responseCollector interface {
	CollectResponse(response *http.Response) error
}

type statusCodeCollector struct {
	StatusCode int
}

func newStatusCodeCollector() *statusCodeCollector {
	return &statusCodeCollector{
		StatusCode: 0,
	}
}

func (p *statusCodeCollector) CollectResponse(response *http.Response) error {
	if response == nil {
		p.StatusCode = 0
		return nil
	}
	p.StatusCode = response.StatusCode
	return nil
}

func newMitmHTTPClient(collectors []responseCollector, transformers []requestTransformer) *http.Client {
	return &http.Client{
		Transport: &mitmTransport{
			Transport:    http.DefaultTransport,
			Transformers: transformers,
			Collectors:   collectors,
		},
	}
}

type mitmTransport struct {
	Transport    http.RoundTripper
	Transformers []requestTransformer
	Collectors   []responseCollector
}

func (t *mitmTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Read request body if present
	var bodyData []byte
	// apply request url transformations
	for _, transformer := range t.Transformers {
		urlStr := transformer.ProcessURL(req.URL.String())
		if urlStr == "" {
			continue
		}
		var err error
		req.URL, err = url.Parse(urlStr)
		if err != nil {
			return nil, err
		}
	}
	if req.Body != nil {
		var err error
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
		var err error
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
	response, err := t.Transport.RoundTrip(req)
	cerrs := []error{}
	for _, collector := range t.Collectors {
		cerrs = append(cerrs, collector.CollectResponse(response))
	}
	for _, cerr := range cerrs {
		if cerr != nil {
			return nil, cerr
		}
	}
	return response, err
}
