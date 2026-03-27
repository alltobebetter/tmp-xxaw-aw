package proxy

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/elazarl/goproxy"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// ---------------------------------------------------------------------------
// Thread-safe data structures
// ---------------------------------------------------------------------------

// KeyPool manages a thread-safe round-robin pool of API keys.
type KeyPool struct {
	mu      sync.RWMutex
	keys    []string
	counter uint64
}

// Next returns the next key in round-robin order.
func (p *KeyPool) Next() (string, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if len(p.keys) == 0 {
		return "", false
	}
	idx := atomic.AddUint64(&p.counter, 1) - 1
	return p.keys[idx%uint64(len(p.keys))], true
}

// SetKeys replaces the key pool atomically.
func (p *KeyPool) SetKeys(keys []string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.keys = keys
}

// ModelMap manages a thread-safe mapping from original model name to target model name.
type ModelMap struct {
	mu   sync.RWMutex
	data map[string]string
}

func (m *ModelMap) SetMap(mappings map[string]string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data = mappings
}

// Get returns the mapped model name. Supports exact match and wildcard (*) fallback.
func (m *ModelMap) Get(original string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.data == nil {
		return "", false
	}
	if target, exists := m.data[original]; exists {
		return target, true
	}
	return "", false
}

// InjectableModels returns all non-wildcard model names for injection into /v1/models.
func (m *ModelMap) InjectableModels() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []string
	for orig := range m.data {
		if orig != "" {
			result = append(result, orig)
		}
	}
	return result
}

// ---------------------------------------------------------------------------
// Provider abstraction
// ---------------------------------------------------------------------------

// providerConfig holds provider-specific configuration for the proxy intercept logic.
type providerConfig struct {
	name      string   // display name for logs, e.g. "OpenAI", "Claude"
	hosts     []string // e.g. ["api.openai.com", "api.openai.com:443"]
	base      string   // user-configured base URL
	keyPool   *KeyPool
	keyHeader string // "Authorization" or "x-api-key"
	keyPrefix string // "Bearer " or ""
	modelMap  *ModelMap
	modelFmt  string // "openai" or "anthropic" — determines /v1/models response format
}

// requestInfo is stored in ctx.UserData to pass data from OnRequest to OnResponse.
type requestInfo struct {
	provider      string // "openai" or "anthropic"
	originalModel string // model name before rewriting (empty if no rewrite happened)
}

// ---------------------------------------------------------------------------
// Server
// ---------------------------------------------------------------------------

type Server struct {
	server          *http.Server
	ctx             context.Context
	openaiKeys      KeyPool
	anthropicKeys   KeyPool
	generalKeys     KeyPool
	openaiModels    ModelMap
	anthropicModels ModelMap
}

func New() *Server {
	return &Server{}
}

func (s *Server) SetContext(ctx context.Context) {
	s.ctx = ctx
}

func (s *Server) SetKeyPools(openai, anthropic, general []string) {
	s.openaiKeys.SetKeys(openai)
	s.anthropicKeys.SetKeys(anthropic)
	s.generalKeys.SetKeys(general)
}

func (s *Server) SetModelMaps(openai, anthropic map[string]string) {
	s.openaiModels.SetMap(openai)
	s.anthropicModels.SetMap(anthropic)
}

func (s *Server) emitLog(prefix, msg, msgType string) {
	if s.ctx != nil {
		runtime.EventsEmit(s.ctx, "proxy_log", map[string]string{
			"prefix": prefix,
			"msg":    msg,
			"type":   msgType,
		})
	}
}

func maskKey(key string) string {
	if len(key) <= 8 {
		return key + "····"
	}
	return key[:8] + "····"
}

func (s *Server) getKeyForProvider(provider string) (string, bool) {
	switch provider {
	case "openai":
		if key, ok := s.openaiKeys.Next(); ok {
			return key, true
		}
	case "anthropic":
		if key, ok := s.anthropicKeys.Next(); ok {
			return key, true
		}
	}
	return s.generalKeys.Next()
}

// ---------------------------------------------------------------------------
// Unified request/response handling
// ---------------------------------------------------------------------------

// matchProvider checks if the request matches any provider and returns its config.
func (s *Server) matchProvider(host string, providers []providerConfig) *providerConfig {
	for i := range providers {
		for _, h := range providers[i].hosts {
			if host == h {
				return &providers[i]
			}
		}
	}
	return nil
}

// handleProviderRequest performs URL rewriting, key rotation, and model remapping for a matched provider.
func (s *Server) handleProviderRequest(prov *providerConfig, req *http.Request, ctx *goproxy.ProxyCtx) {
	// URL rewriting
	if prov.base != "" {
		if parsed, err := url.Parse(prov.base); err == nil {
			req.URL.Scheme = parsed.Scheme
			req.URL.Host = parsed.Host
			req.Host = parsed.Host
			s.emitLog(prov.name, "成功劫持拦截，已转发至 "+parsed.Host, "success")
		}
	}

	// Key rotation
	if key, ok := s.getKeyForProvider(prov.modelFmt); ok {
		if prov.keyPrefix != "" {
			req.Header.Set(prov.keyHeader, prov.keyPrefix+key)
		} else {
			req.Header.Set(prov.keyHeader, key)
		}
		s.emitLog(prov.name, "已轮询替换密钥: "+maskKey(key), "success")
	}

	// Model remapping — only parse body for POST-like methods
	if req.Method == "POST" || req.Method == "PUT" || req.Method == "PATCH" {
		if req.Body != nil {
			bodyBytes, err := io.ReadAll(req.Body)
			if err == nil && len(bodyBytes) > 0 {
				var bodyParsed map[string]interface{}
				if json.Unmarshal(bodyBytes, &bodyParsed) == nil {
					if originalModel, ok := bodyParsed["model"].(string); ok {
						if targetModel, mapped := prov.modelMap.Get(originalModel); mapped {
							bodyParsed["model"] = targetModel
							newBodyBytes, err := json.Marshal(bodyParsed)
							if err == nil {
								req.Body = io.NopCloser(bytes.NewBuffer(newBodyBytes))
								req.ContentLength = int64(len(newBodyBytes))
								// Store rewrite info for response handler
								ctx.UserData = &requestInfo{
									provider:      prov.modelFmt,
									originalModel: originalModel,
								}
								s.emitLog("Model", "成功重写 "+prov.name+" 模型: "+originalModel+" -> "+targetModel, "success")
								return
							}
						}
					}
				}
				// No rewrite happened — restore body
				req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
		}
	}

	// Store provider info even if no model rewrite (for /v1/models injection)
	if ctx.UserData == nil {
		ctx.UserData = &requestInfo{provider: prov.modelFmt}
	}
}

// rewriteResponseBody rewrites the model field in a non-streaming JSON response back to the original name.
func rewriteResponseBody(resp *http.Response, originalModel string) {
	contentType := resp.Header.Get("Content-Type")

	// ── Streaming (SSE) responses ──
	// Don't buffer; wrap the body in a streaming replacer.
	if strings.Contains(contentType, "text/event-stream") {
		// We can't predict the exact target model name in the response,
		// so for SSE we skip response model rewriting.
		// The request-side rewrite is sufficient for Trae to work.
		return
	}

	// ── Non-streaming JSON responses ──
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var bodyParsed map[string]interface{}
	if json.Unmarshal(bodyBytes, &bodyParsed) != nil {
		resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		return
	}

	bodyParsed["model"] = originalModel
	newBodyBytes, err := json.Marshal(bodyParsed)
	if err != nil {
		resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		return
	}

	resp.Body = io.NopCloser(bytes.NewBuffer(newBodyBytes))
	resp.ContentLength = int64(len(newBodyBytes))
	resp.Header.Set("Content-Length", fmt.Sprintf("%d", len(newBodyBytes)))
}

// injectModelsIntoList appends custom model entries to a GET /v1/models response.
func injectModelsIntoList(resp *http.Response, modelMap *ModelMap, format string) {
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var bodyParsed map[string]interface{}
	if json.Unmarshal(bodyBytes, &bodyParsed) != nil {
		resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		return
	}

	dataList, ok := bodyParsed["data"].([]interface{})
	if !ok {
		resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		return
	}

	for _, modelName := range modelMap.InjectableModels() {
		var mockObj map[string]interface{}
		if format == "anthropic" {
			mockObj = map[string]interface{}{
				"id":           modelName,
				"type":         "model",
				"display_name": modelName,
				"created_at":   "2024-01-01T00:00:00Z",
			}
		} else {
			mockObj = map[string]interface{}{
				"id":       modelName,
				"object":   "model",
				"created":  1686935002,
				"owned_by": "trae-proxy-injected",
			}
		}
		dataList = append(dataList, mockObj)
	}

	bodyParsed["data"] = dataList
	newBodyBytes, err := json.Marshal(bodyParsed)
	if err != nil {
		resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		return
	}

	resp.Body = io.NopCloser(bytes.NewBuffer(newBodyBytes))
	resp.ContentLength = int64(len(newBodyBytes))
	resp.Header.Set("Content-Length", fmt.Sprintf("%d", len(newBodyBytes)))
}

// ---------------------------------------------------------------------------
// Start / Stop
// ---------------------------------------------------------------------------

func (s *Server) Start(port int, openaiBase string, anthropicBase string, certBytes, keyBytes []byte) error {
	ca, err := tls.X509KeyPair(certBytes, keyBytes)
	if err != nil {
		return fmt.Errorf("failed to load CA: %v", err)
	}

	goproxy.GoproxyCa = ca
	goproxy.OkConnect = &goproxy.ConnectAction{Action: goproxy.ConnectAccept, TLSConfig: goproxy.TLSConfigFromCA(&ca)}
	goproxy.MitmConnect = &goproxy.ConnectAction{Action: goproxy.ConnectMitm, TLSConfig: goproxy.TLSConfigFromCA(&ca)}
	goproxy.HTTPMitmConnect = &goproxy.ConnectAction{Action: goproxy.ConnectHTTPMitm, TLSConfig: goproxy.TLSConfigFromCA(&ca)}
	goproxy.RejectConnect = &goproxy.ConnectAction{Action: goproxy.ConnectReject, TLSConfig: goproxy.TLSConfigFromCA(&ca)}

	p := goproxy.NewProxyHttpServer()
	p.Verbose = false

	// Build provider configs
	providers := []providerConfig{
		{
			name:      "OpenAI",
			hosts:     []string{"api.openai.com", "api.openai.com:443"},
			base:      openaiBase,
			keyPool:   &s.openaiKeys,
			keyHeader: "Authorization",
			keyPrefix: "Bearer ",
			modelMap:  &s.openaiModels,
			modelFmt:  "openai",
		},
		{
			name:      "Claude",
			hosts:     []string{"api.anthropic.com", "api.anthropic.com:443"},
			base:      anthropicBase,
			keyPool:   &s.anthropicKeys,
			keyHeader: "x-api-key",
			keyPrefix: "",
			modelMap:  &s.anthropicModels,
			modelFmt:  "anthropic",
		},
	}

	// MITM intercept for target domains
	p.OnRequest(goproxy.ReqHostMatches(regexp.MustCompile(`^api\.openai\.com:443$`))).
		HandleConnect(goproxy.AlwaysMitm)
	p.OnRequest(goproxy.ReqHostMatches(regexp.MustCompile(`^api\.anthropic\.com:443$`))).
		HandleConnect(goproxy.AlwaysMitm)

	// ── OnRequest: URL rewrite, key rotation, model remapping ──
	p.OnRequest().DoFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		prov := s.matchProvider(req.Host, providers)
		if prov == nil {
			return req, nil // Not a target host, pass through
		}

		// Only strip Accept-Encoding for GET /models so we can parse the model list.
		// Leave it intact for all other requests to preserve streaming & compression.
		isModelsRequest := req.Method == "GET" && (strings.HasSuffix(req.URL.Path, "/v1/models") || strings.HasSuffix(req.URL.Path, "/models"))
		if isModelsRequest {
			req.Header.Del("Accept-Encoding")
		}

		s.handleProviderRequest(prov, req, ctx)
		return req, nil
	})

	// ── OnResponse: model name back-rewrite & model list injection ──
	p.OnResponse().DoFunc(func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		if resp == nil || resp.Request == nil || resp.Body == nil {
			return resp
		}

		info, _ := ctx.UserData.(*requestInfo)
		if info == nil {
			return resp
		}

		// Rewrite response model name back to original (for non-streaming responses)
		if info.originalModel != "" && resp.StatusCode == 200 {
			rewriteResponseBody(resp, info.originalModel)
		}

		// Inject custom models into GET /v1/models response
		if resp.Request.Method == "GET" && resp.StatusCode == 200 &&
			(strings.HasSuffix(resp.Request.URL.Path, "/v1/models") || strings.HasSuffix(resp.Request.URL.Path, "/models")) {

			var modelMap *ModelMap
			switch info.provider {
			case "openai":
				modelMap = &s.openaiModels
			case "anthropic":
				modelMap = &s.anthropicModels
			}
			if modelMap != nil {
				injectModelsIntoList(resp, modelMap, info.provider)
			}
		}

		return resp
	})

	s.server = &http.Server{
		Addr:    fmt.Sprintf("127.0.0.1:%d", port),
		Handler: p,
	}

	go func() {
		_ = s.server.ListenAndServe()
	}()
	return nil
}

func (s *Server) Stop() error {
	if s.server != nil {
		return s.server.Shutdown(context.Background())
	}
	return nil
}
