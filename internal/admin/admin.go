package admin

import (
	"embed"
	"io/fs"
	"net/http"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
)

//go:embed static/*
var staticFiles embed.FS

type Option struct {
	BasePath string
}

func NewHandler(opt Option) gin.HandlerFunc {
	sub, err := fs.Sub(staticFiles, "static")
	if err != nil {
		panic("failed to load admin static files: " + err.Error())
	}

	return func(c *gin.Context) {
		p := c.Param("path")

		// NoRoute 没有 *path 参数，从 URL 里取
		if p == "" {
			p = strings.TrimPrefix(c.Request.URL.Path, opt.BasePath)
		}
		p = strings.TrimPrefix(p, "/")

		if p == "" || !strings.Contains(p, ".") {
			p = "index.html"
		}

		data, err := fs.ReadFile(sub, p)
		if err != nil {
			c.String(http.StatusNotFound, "404 page not found")
			return
		}

		switch path.Ext(p) {
		case ".css":
			c.Data(http.StatusOK, "text/css; charset=utf-8", data)
		case ".js":
			c.Data(http.StatusOK, "application/javascript", data)
		case ".html":
			c.Data(http.StatusOK, "text/html; charset=utf-8", data)
		case ".png":
			c.Data(http.StatusOK, "image/png", data)
		case ".jpg", ".jpeg":
			c.Data(http.StatusOK, "image/jpeg", data)
		case ".svg":
			c.Data(http.StatusOK, "image/svg+xml", data)
		case ".ico":
			c.Data(http.StatusOK, "image/x-icon", data)
		case ".woff2":
			c.Data(http.StatusOK, "font/woff2", data)
		case ".woff":
			c.Data(http.StatusOK, "font/woff", data)
		default:
			c.Data(http.StatusOK, "text/plain; charset=utf-8", data)
		}
	}
}
