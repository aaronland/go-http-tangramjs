package tangramjs

import (
	"github.com/aaronland/go-http-leaflet"
	"github.com/aaronland/go-http-rewrite"
	_ "log"
	"net/http"
	"path/filepath"
	"strings"
)

type NextzenOptions struct {
	APIKey string
	StyleURL string
}

type TangramJSOptions struct {
	JS      []string
	CSS     []string
	Nextzen *NextzenOptions
	Leaflet *leaflet.LeafletOptions
}

func DefaultTangramJSOptions() *TangramJSOptions {

	leaflet_opts := leaflet.DefaultLeafletOptions()
	nextzen_opts := &NextzenOptions{}
	
	opts := &TangramJSOptions{
		CSS: []string{},
		JS: []string{
			"/javascript/tangram.min.js",
		},
		Leaflet: leaflet_opts,
		Nextzen: nextzen_opts,
	}

	return opts
}

func AppendResourcesHandler(next http.Handler, opts *TangramJSOptions) http.Handler {
	return AppendResourcesHandlerWithPrefix(next, opts, "")
}

func AppendResourcesHandlerWithPrefix(next http.Handler, opts *TangramJSOptions, prefix string) http.Handler {

	js := opts.JS
	css := opts.CSS

	if prefix != "" {

		for i, path := range js {
			js[i] = appendPrefix(prefix, path)
		}

		for i, path := range css {
			css[i] = appendPrefix(prefix, path)
		}
	}

	attrs := map[string]string{
		"nextzen-api-key": opts.Nextzen.APIKey,
		"nextzen-style-url": opts.Nextzen.StyleURL,		
	}
	
	append_opts := &rewrite.AppendResourcesOptions{
		JavaScript:  js,
		Stylesheets: css,
		DataAttributes: attrs,
	}

	next = leaflet.AppendResourcesHandlerWithPrefix(next, opts.Leaflet, prefix)

	return rewrite.AppendResourcesHandler(next, append_opts)
}

func AssetsHandler() (http.Handler, error) {

	fs := assetFS()
	return http.FileServer(fs), nil
}

func AssetsHandlerWithPrefix(prefix string) (http.Handler, error) {

	fs_handler, err := AssetsHandler()

	if err != nil {
		return nil, err
	}

	prefix = strings.TrimRight(prefix, "/")

	if prefix == "" {
		return fs_handler, nil
	}

	rewrite_func := func(req *http.Request) (*http.Request, error) {
		req.URL.Path = strings.Replace(req.URL.Path, prefix, "", 1)
		return req, nil
	}

	rewrite_handler := rewrite.RewriteRequestHandler(fs_handler, rewrite_func)
	return rewrite_handler, nil
}

func AppendAssetHandlers(mux *http.ServeMux) error {
	return AppendAssetHandlersWithPrefix(mux, "")
}

func AppendAssetHandlersWithPrefix(mux *http.ServeMux, prefix string) error {

	err := leaflet.AppendAssetHandlersWithPrefix(mux, prefix)

	if err != nil {
		return err
	}

	asset_handler, err := AssetsHandlerWithPrefix(prefix)

	if err != nil {
		return nil
	}

	for _, path := range AssetNames() {

		path := strings.Replace(path, "static", "", 1)

		if prefix != "" {
			path = appendPrefix(prefix, path)
		}

		mux.Handle(path, asset_handler)
	}

	return nil
}

func appendPrefix(prefix string, path string) string {

	prefix = strings.TrimRight(prefix, "/")

	if prefix != "" {
		path = strings.TrimLeft(path, "/")
		path = filepath.Join(prefix, path)
	}

	return path
}
