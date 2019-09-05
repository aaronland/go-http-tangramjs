package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/aaronland/go-http-leaflet"
	"github.com/aaronland/go-http-leaflet/assets/templates"
	"html/template"
	"log"
	"net/http"
)

type MapVars struct {
	TileURL string
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

	tile_url := flag.String("tile-url", "", "A valid Leaflet layer tile URL")
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

	map_vars := new(MapVars)

	if *tile_url != "" {
		map_vars.TileURL = *tile_url
	}

	mux := http.NewServeMux()

	map_handler, err := MapHandler(t, map_vars)

	if err != nil {
		log.Fatal(err)
	}

	leaflet_opts := leaflet.DefaultLeafletOptions()

	map_handler = leaflet.AppendResourcesHandler(map_handler, leaflet_opts)

	mux.Handle("/", map_handler)

	err = leaflet.AppendAssetHandlers(mux)

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
