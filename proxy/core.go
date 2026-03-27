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

// KeyPool manages a thread-safe round-robin pool of API keys.
type KeyPool struct {
	mu      sync.RWMutex
	keys    []string
	counter uint64
}

// Next returns the next key in round-robin order.
// Returns ("", false) if the pool is empty.
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

// ModelMap manages thread-safe map of original model to target model.
type ModelMap struct {
	mu   sync.RWMutex
	data map[string]string
}

func (m *ModelMap) SetMap(mappings map[string]string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data = mappings
}

func (m *ModelMap) Get(original string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.data == nil {
		return "", false
	}
	
	// Exact match
	if target, exists := m.data[original]; exists {
		return target, true
	}
	
	// Wildcard fallback
	if target, exists := m.data["*"]; exists {
		return target, true
	}
	
	return "", false
}

type Server struct {
	server       *http.Server
	ctx          context.Context
	openaiKeys   KeyPool
	anthropicKeys KeyPool
	generalKeys  KeyPool
	openaiModels ModelMap
	anthropicModels ModelMap
}

// modelRewriteInfo is stored in goproxy ctx.UserData to pass info from request to response handler.
type modelRewriteInfo struct {
	originalModel string
	provider      string // "openai" or "anthropic"
}

func New() *Server {
	return &Server{}
}

func (s *Server) SetContext(ctx context.Context) {
	s.ctx = ctx
}

// SetKeyPools updates the key rotation pools. Can be called while proxy is running.
func (s *Server) SetKeyPools(openai, anthropic, general []string) {
	s.openaiKeys.SetKeys(openai)
	s.anthropicKeys.SetKeys(anthropic)
	s.generalKeys.SetKeys(general)
}

// SetModelMaps updates the model name mappings. Can be called while proxy is running.
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

// maskKey safely returns a masked version of a key for logging.
func maskKey(key string) string {
	if len(key) <= 8 {
		return key + "····"
	}
	return key[:8] + "····"
}

// getKeyForProvider returns the next key for a given provider.
// Falls back to the general pool if the provider-specific pool is empty.
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
	// Fallback to general pool
	return s.generalKeys.Next()
}

func (s *Server) Start(port int, openaiBase string, anthropicBase string, certBytes, keyBytes []byte) error {
	ca, err := tls.X509KeyPair(certBytes, keyBytes)
	if err == nil {
		goproxy.GoproxyCa = ca
		goproxy.OkConnect = &goproxy.ConnectAction{Action: goproxy.ConnectAccept, TLSConfig: goproxy.TLSConfigFromCA(&ca)}
		goproxy.MitmConnect = &goproxy.ConnectAction{Action: goproxy.ConnectMitm, TLSConfig: goproxy.TLSConfigFromCA(&ca)}
		goproxy.HTTPMitmConnect = &goproxy.ConnectAction{Action: goproxy.ConnectHTTPMitm, TLSConfig: goproxy.TLSConfigFromCA(&ca)}
		goproxy.RejectConnect = &goproxy.ConnectAction{Action: goproxy.ConnectReject, TLSConfig: goproxy.TLSConfigFromCA(&ca)}
	} else {
		return fmt.Errorf("failed to load CA: %v", err)
	}

	p := goproxy.NewProxyHttpServer()
	p.Verbose = false

	// MITM Only
	p.OnRequest(goproxy.ReqHostMatches(regexp.MustCompile(`^api\.openai\.com:443$`))).
		HandleConnect(goproxy.AlwaysMitm)
	p.OnRequest(goproxy.ReqHostMatches(regexp.MustCompile(`^api\.anthropic\.com:443$`))).
		HandleConnect(goproxy.AlwaysMitm)

	// Rewrite URLs, rotate keys, and remap models
	p.OnRequest().DoFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		isMethodPost := req.Method == "POST" || req.Method == "PUT" || req.Method == "PATCH"
		
		var bodyBytes []byte
		var bodyParsed map[string]interface{}
		hasBody := false

		// Parse body if there is one, to rewrite the model name
		if isMethodPost && req.Body != nil {
			var err error
			bodyBytes, err = io.ReadAll(req.Body)
			if err == nil && len(bodyBytes) > 0 {
				hasBody = true
				req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // restore for later
				if err := json.Unmarshal(bodyBytes, &bodyParsed); err != nil {
					hasBody = false // ignore non-JSON body
				}
			}
		}

		// Force uncompressed response for easy rewriting
		req.Header.Del("Accept-Encoding")

		// Save original host before rewriting so response handler can determine provider
		req.Header.Set("X-TraeProxy-Original-Host", req.Host)

		if req.Host == "api.openai.com" || req.Host == "api.openai.com:443" {
			if openaiBase != "" {
				if parsed, err := url.Parse(openaiBase); err == nil {
					req.URL.Scheme = parsed.Scheme
					req.URL.Host = parsed.Host
					req.Host = parsed.Host
					s.emitLog("OpenAI", "成功劫持拦截，已转发至 "+parsed.Host, "success")
				}
			}
			// Key rotation: replace Authorization header
			if key, ok := s.getKeyForProvider("openai"); ok {
				req.Header.Set("Authorization", "Bearer "+key)
				s.emitLog("OpenAI", "已轮询替换密钥: "+maskKey(key), "success")
			}
			// Model remapping
			if hasBody {
				if originalModel, ok := bodyParsed["model"].(string); ok {
					if targetModel, mapped := s.openaiModels.Get(originalModel); mapped {
						bodyParsed["model"] = targetModel
						newBodyBytes, err := json.Marshal(bodyParsed)
						if err == nil {
							req.Body = io.NopCloser(bytes.NewBuffer(newBodyBytes))
							req.ContentLength = int64(len(newBodyBytes))
							// Save original model + provider to context for response rewriting
							ctx.UserData = &modelRewriteInfo{originalModel: originalModel, provider: "openai"}
							s.emitLog("Model", "成功重写 OpenAI 模型: "+originalModel+" -> "+targetModel, "success")
						}
					}
				}
			}
		} else if req.Host == "api.anthropic.com" || req.Host == "api.anthropic.com:443" {
			if anthropicBase != "" {
				if parsed, err := url.Parse(anthropicBase); err == nil {
					req.URL.Scheme = parsed.Scheme
					req.URL.Host = parsed.Host
					req.Host = parsed.Host
					s.emitLog("Claude", "成功劫持拦截，已转发至 "+parsed.Host, "success")
				}
			}
			// Key rotation: replace x-api-key header
			if key, ok := s.getKeyForProvider("anthropic"); ok {
				req.Header.Set("x-api-key", key)
				s.emitLog("Claude", "已轮询替换密钥: "+maskKey(key), "success")
			}
			// Model remapping
			if hasBody {
				if originalModel, ok := bodyParsed["model"].(string); ok {
					if targetModel, mapped := s.anthropicModels.Get(originalModel); mapped {
						bodyParsed["model"] = targetModel
						newBodyBytes, err := json.Marshal(bodyParsed)
						if err == nil {
							req.Body = io.NopCloser(bytes.NewBuffer(newBodyBytes))
							req.ContentLength = int64(len(newBodyBytes))
							// Save original model + provider to context for response rewriting
							ctx.UserData = &modelRewriteInfo{originalModel: originalModel, provider: "anthropic"}
							s.emitLog("Model", "成功重写 Claude 模型: "+originalModel+" -> "+targetModel, "success")
						}
					}
				}
			}
		}
		return req, nil
	})

	// Inject models into /v1/models response and rewrite response model back
	p.OnResponse().DoFunc(func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		if resp == nil || resp.Request == nil || resp.Body == nil {
			return resp
		}

		// Rewrite response model name back so Trae doesn't reject it
		if info, ok := ctx.UserData.(*modelRewriteInfo); ok && resp.StatusCode == 200 {
			bodyBytes, err := io.ReadAll(resp.Body)
			if err == nil {
				var bodyParsed map[string]interface{}
				if err := json.Unmarshal(bodyBytes, &bodyParsed); err == nil {
					// Replace the model field in response
					bodyParsed["model"] = info.originalModel
					newBodyBytes, err := json.Marshal(bodyParsed)
					if err == nil {
						resp.Body = io.NopCloser(bytes.NewBuffer(newBodyBytes))
						resp.ContentLength = int64(len(newBodyBytes))
						resp.Header.Set("Content-Length", fmt.Sprintf("%d", len(newBodyBytes)))
					} else {
						resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
					}
				} else {
					resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				}
			}
		}

		// Determine provider from our custom header (set before host rewriting)
		if resp.Request.Method == "GET" && resp.StatusCode == 200 && (strings.HasSuffix(resp.Request.URL.Path, "/v1/models") || strings.HasSuffix(resp.Request.URL.Path, "/models")) {
			originalHost := ""
			if resp.Request != nil {
				originalHost = resp.Request.Header.Get("X-TraeProxy-Original-Host")
			}
			isAnthropic := strings.Contains(originalHost, "anthropic")
			bodyBytes, err := io.ReadAll(resp.Body)
			if err == nil {
				var bodyParsed map[string]interface{}
				if err := json.Unmarshal(bodyBytes, &bodyParsed); err == nil {
					if dataList, ok := bodyParsed["data"].([]interface{}); ok {
						if isAnthropic {
							// Inject Anthropic format models
							s.anthropicModels.mu.RLock()
							for orig := range s.anthropicModels.data {
								if orig != "*" && orig != "" {
									mockObj := map[string]interface{}{
										"id":           orig,
										"type":         "model",
										"display_name": orig,
										"created_at":   "2024-01-01T00:00:00Z",
									}
									dataList = append(dataList, mockObj)
								}
							}
							s.anthropicModels.mu.RUnlock()
						} else {
							// Inject OpenAI format models
							s.openaiModels.mu.RLock()
							for orig := range s.openaiModels.data {
								if orig != "*" && orig != "" {
									mockObj := map[string]interface{}{
										"id":       orig,
										"object":   "model",
										"created":  1686935002,
										"owned_by": "trae-proxy-injected",
									}
									dataList = append(dataList, mockObj)
								}
							}
							s.openaiModels.mu.RUnlock()
						}

						bodyParsed["data"] = dataList
						newBodyBytes, err := json.Marshal(bodyParsed)
						if err == nil {
							resp.Body = io.NopCloser(bytes.NewBuffer(newBodyBytes))
							resp.ContentLength = int64(len(newBodyBytes))
							resp.Header.Set("Content-Length", fmt.Sprintf("%d", len(newBodyBytes)))
						} else {
							resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
						}
					} else {
						resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
					}
				} else {
					resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				}
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
