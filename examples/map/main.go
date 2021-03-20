package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/aaronland/go-http-tangramjs"
	"github.com/aaronland/go-http-tangramjs/templates/html"
	"html/template"
	"log"
	"net/http"
)

func MapHandler(templates *template.Template) (http.Handler, error) {

	t := templates.Lookup("map")

	if t == nil {
		return nil, errors.New("Missing 'map' template")
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

	host := flag.String("host", "localhost", "...")
	port := flag.Int("port", 8080, "...")
	api_key := flag.String("api-key", "", "...")

	style_url := flag.String("style-url", "/tangram/refill-style.zip", "A valid Leaflet layer tile URL")

	flag.Parse()

	t, err := template.ParseFS(html.FS, "*.html")

	if err != nil {
		log.Fatalf("Failed to parse templates, %v", err)
	}

	mux := http.NewServeMux()

	map_handler, err := MapHandler(t)

	if err != nil {
		log.Fatalf("Failed to create map handler, %v", err)
	}

	tangramjs_opts := tangramjs.DefaultTangramJSOptions()
	tangramjs_opts.Nextzen.APIKey = *api_key
	tangramjs_opts.Nextzen.StyleURL = *style_url

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
