package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/aaronland/go-http-tangramjs"
	"github.com/aaronland/go-http-tangramjs/assets/templates"
	"html/template"
	"log"
	"net/http"
)

type MapVars struct {
	APIKey   string
	StyleURL string
}

func MapHandler(templates *template.Template, map_vars *MapVars) (http.Handler, error) {

	t := templates.Lookup("map")

	if t == nil {
		return nil, errors.New("Missing 'map' template")
	}

	fn := func(rsp http.ResponseWriter, req *http.Request) {

		err := t.Execute(rsp, map_vars)

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
	path_templates := flag.String("templates", "", "An optional string for local templates. This is anything that can be read by the 'templates.ParseGlob' method.")

	flag.Parse()

	t := template.New("example")

	var err error

	if *path_templates != "" {

		t, err = t.ParseGlob(*path_templates)

		if err != nil {
			log.Fatal(err)
		}

	} else {

		for _, name := range templates.AssetNames() {

			body, err := templates.Asset(name)

			if err != nil {
				log.Fatal(err)
			}

			t, err = t.Parse(string(body))

			if err != nil {
				log.Fatal(err)
			}
		}
	}

	mux := http.NewServeMux()

	map_vars := &MapVars{
		APIKey:   *api_key,
		StyleURL: *style_url,
	}

	map_handler, err := MapHandler(t, map_vars)

	if err != nil {
		log.Fatal(err)
	}

	tangramjs_opts := tangramjs.DefaultTangramJSOptions()

	map_handler = tangramjs.AppendResourcesHandler(map_handler, tangramjs_opts)

	mux.Handle("/", map_handler)

	err = tangramjs.AppendAssetHandlers(mux)

	if err != nil {
		log.Fatal(err)
	}

	endpoint := fmt.Sprintf("%s:%d", *host, *port)
	log.Printf("Listening for requests on %s\n", endpoint)

	err = http.ListenAndServe(endpoint, mux)

	if err != nil {
		log.Fatal(err)
	}

}
