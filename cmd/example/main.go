package main

import (
	"embed"
	"errors"
	"flag"
	"fmt"
	"github.com/aaronland/go-http-leaflet"
	"github.com/aaronland/go-http-tangramjs"
	"html/template"
	"log"
	"net/http"
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

	flag.Parse()

	t, err := template.ParseFS(FS, "*.html")

	if err != nil {
		log.Fatalf("Failed to parse templates, %v", err)
	}

	mux := http.NewServeMux()

	map_handler, err := ExampleHandler(t)

	if err != nil {
		log.Fatalf("Failed to create map handler, %v", err)
	}

	if !*append_leaflet {

		tangramjs.APPEND_LEAFLET_RESOURCES = false
		tangramjs.APPEND_LEAFLET_ASSETS = false

		leaflet_opts := leaflet.DefaultLeafletOptions()
		map_handler = leaflet.AppendResourcesHandler(map_handler, leaflet_opts)

		err = leaflet.AppendAssetHandlers(mux)

		if err != nil {
			log.Fatalf("Failed to append Leaflet asset handlers, %v", err)
		}
	}

	tangramjs_opts := tangramjs.DefaultTangramJSOptions()
	tangramjs_opts.NextzenOptions.APIKey = *api_key
	tangramjs_opts.NextzenOptions.StyleURL = *style_url

	map_handler = tangramjs.AppendResourcesHandler(map_handler, tangramjs_opts)

	mux.Handle("/", map_handler)

	err = tangramjs.AppendAssetHandlers(mux)

	if err != nil {
		log.Fatalf("Failed to append Tangram asset handlers, %v", err)
	}

	endpoint := fmt.Sprintf("%s:%d", *host, *port)
	log.Printf("Listening for requests on %s\n", endpoint)

	err = http.ListenAndServe(endpoint, mux)

	if err != nil {
		log.Fatalf("Failed to serve requests, %v", err)
	}

}
