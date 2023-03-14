package main

import (
	"embed"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/aaronland/go-http-leaflet"
	"github.com/aaronland/go-http-tangramjs"
)

//go:embed *.html
var FS embed.FS

func ExampleHandler(templates *template.Template) (http.Handler, error) {

	t := templates.Lookup("example")

	if t == nil {
		return nil, errors.New("Missing 'example' template")
	}

	fn := func(rsp http.ResponseWriter, req *http.Request) {

		err := t.Execute(rsp, nil)

		if err != nil {
			http.Error(rsp, err.Error(), http.StatusInternalServerError)
		}

		return
	}

	return http.HandlerFunc(fn), nil
}

func main() {

	host := flag.String("host", "localhost", "The host name to listen for requests on.")
	port := flag.Int("port", 8080, "The port number to list for requests on.")
	api_key := flag.String("api-key", "", "A valid Nextzen API key.")

	style_url := flag.String("style-url", "/tangram/refill-style.zip", "A valid Leaflet layer tile URL")

	append_leaflet := flag.Bool("append-leaflet", true, "Automatically append Leafet.js assets and resources.")

	js_eof := flag.Bool("javascript-at-eof", false, "Append JavaScript resources to end of HTML file.")
	rollup_assets := flag.Bool("rollup-assets", false, "Rollup (minify and bundle) JavaScript and CSS assets.")

	flag.Parse()

	logger := log.Default()

	t, err := template.ParseFS(FS, "*.html")

	if err != nil {
		log.Fatalf("Failed to parse templates, %v", err)
	}

	tangramjs_opts := tangramjs.DefaultTangramJSOptions()
	tangramjs_opts.NextzenOptions.APIKey = *api_key
	tangramjs_opts.NextzenOptions.StyleURL = *style_url
	tangramjs_opts.AppendJavaScriptAtEOF = *js_eof
	tangramjs_opts.RollupAssets = *rollup_assets
	tangramjs_opts.Logger = logger

	mux := http.NewServeMux()

	map_handler, err := ExampleHandler(t)

	if err != nil {
		logger.Fatalf("Failed to create map handler, %v", err)
	}

	if !*append_leaflet {

		tangramjs_opts.AppendLeafletResources = false
		tangramjs_opts.AppendLeafletAssets = false

		leaflet_opts := leaflet.DefaultLeafletOptions()
		leaflet_opts.AppendJavaScriptAtEOF = *js_eof
		leaflet_opts.RollupAssets = *rollup_assets

		map_handler = leaflet.AppendResourcesHandler(map_handler, leaflet_opts)

		err = leaflet.AppendAssetHandlers(mux, leaflet_opts)

		if err != nil {
			logger.Fatalf("Failed to append Leaflet asset handlers, %v", err)
		}
	}

	map_handler = tangramjs.AppendResourcesHandler(map_handler, tangramjs_opts)

	mux.Handle("/", map_handler)

	err = tangramjs.AppendAssetHandlers(mux, tangramjs_opts)

	if err != nil {
		logger.Fatalf("Failed to append Tangram asset handlers, %v", err)
	}

	endpoint := fmt.Sprintf("%s:%d", *host, *port)
	logger.Printf("Listening for requests on %s\n", endpoint)

	err = http.ListenAndServe(endpoint, mux)

	if err != nil {
		logger.Fatalf("Failed to serve requests, %v", err)
	}

}
