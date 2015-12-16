package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/version"

	"github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/ant0ine/go-json-rest/rest"
)

func main() {
	fmt.Println(version.String)
	fmt.Println(data.GetData())

	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	router, err := rest.MakeRouter(
		rest.Get("/version", getVersion),
		rest.Get("/data", getData),
		rest.Post("/dataset", postDataset),
		rest.Get("/dataset", getDataset),
	)
	if err != nil {
		log.Fatal(err)
	}
	api.SetApp(router)
	log.Fatal(http.ListenAndServe(":8080", api.MakeHandler()))
}

func getVersion(w rest.ResponseWriter, r *rest.Request) {
	w.WriteJson(version.String)
}

func postDataset(w rest.ResponseWriter, r *rest.Request) {
	var requied map[string]interface{}

	err := r.DecodeJsonPayload(&requied)

	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if requied == nil {
	}
	w.WriteJson(&requied)
}

func getDataset(w rest.ResponseWriter, r *rest.Request) {
	w.WriteJson(version.String)
}

func getData(w rest.ResponseWriter, r *rest.Request) {
	w.WriteJson(version.String)
}
