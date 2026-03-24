package proxy

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"regexp"

	"github.com/elazarl/goproxy"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type Server struct {
	server *http.Server
	ctx    context.Context
}

func New() *Server {
	return &Server{}
}

func (s *Server) SetContext(ctx context.Context) {
	s.ctx = ctx
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

	// Rewrite URLs - check against internal decrypted HTTPS requests by Host header
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
		} else if req.Host == "api.anthropic.com" || req.Host == "api.anthropic.com:443" {
			if anthropicBase != "" {
				if parsed, err := url.Parse(anthropicBase); err == nil {
					req.URL.Scheme = parsed.Scheme
					req.URL.Host = parsed.Host
					req.Host = parsed.Host
					s.emitLog("Claude", "成功劫持拦截，已转发至 "+parsed.Host, "success")
				}
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
