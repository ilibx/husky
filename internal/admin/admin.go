package admin

import (
	"embed"
	"io/fs"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

//go:embed static/*
var staticFiles embed.FS

type Option struct {
	Mode string // "embedded" or "external"
	URL  string // external URL for "external" mode, e.g. http://localhost:5173
}

// NewHandler returns an HTTP handler for the admin UI.
//
// Modes:
//
//	embedded (default) — serve SPA from embedded static files
//	external          — reverse-proxy to an external dev server (e.g. Vite dev)
func NewHandler(opt Option) http.Handler {
	if opt.Mode == "" {
		opt.Mode = "embedded"
	}

	switch opt.Mode {
	case "external":
		return newProxyHandler(opt.URL)
	default:
		return newEmbeddedHandler()
	}
}

func newEmbeddedHandler() http.Handler {
	sub, err := fs.Sub(staticFiles, "static")
	if err != nil {
		panic("failed to load admin static files: " + err.Error())
	}
	fileServer := http.FileServer(http.FS(sub))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/admin")
		path = strings.TrimPrefix(path, "/")

		if path == "" {
			path = "index.html"
		} else {
			if !strings.Contains(path, ".") {
				path = "index.html"
			}
		}

		r.URL.Path = "/" + path
		fileServer.ServeHTTP(w, r)
	})
}

func newProxyHandler(targetURL string) http.Handler {
	if targetURL == "" {
		targetURL = "http://localhost:5173"
	}

	target, err := url.Parse(targetURL)
	if err != nil {
		panic("invalid admin external URL: " + err.Error())
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.ModifyResponse = func(resp *http.Response) error {
		return nil
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Host = target.Host
		r.URL.Scheme = target.Scheme
		r.Host = target.Host
		proxy.ServeHTTP(w, r)
	})
}

// Handler returns an embedded-only handler (kept for backward compatibility).
func Handler() http.Handler {
	return newEmbeddedHandler()
}
