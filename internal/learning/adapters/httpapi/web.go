package httpapi

import (
	"embed"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

//go:embed web
var webFiles embed.FS

func WebApp() http.Handler {
	dist, err := fs.Sub(webFiles, "web/dist")
	if err == nil {
		_, err = fs.Stat(dist, "index.html")
	}
	if err != nil {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "Frontend not built. Run make build.", http.StatusServiceUnavailable)
		})
	}
	files := http.FileServer(http.FS(dist))
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if strings.HasPrefix(request.URL.Path, "/api/") {
			http.NotFound(response, request)
			return
		}
		requested := strings.TrimPrefix(path.Clean(request.URL.Path), "/")
		if requested == "." || requested == "" {
			requested = "index.html"
		}
		if _, err := fs.Stat(dist, requested); err != nil {
			request.URL.Path = "/"
		}
		files.ServeHTTP(response, request)
	})
}
