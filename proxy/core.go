package proxy

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
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

type Server struct {
	server       *http.Server
	ctx          context.Context
	openaiKeys   KeyPool
	anthropicKeys KeyPool
	generalKeys  KeyPool
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

	// Rewrite URLs and rotate keys
	p.OnRequest().DoFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
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
		}
		return req, nil
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
