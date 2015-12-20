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
	var dataSet data.GenericDataset

	err := r.DecodeJsonPayload(&dataSet)

	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteJson(&dataSet)
	return
}

func getDataset(w rest.ResponseWriter, r *rest.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	return
}

func getData(w rest.ResponseWriter, r *rest.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	return
}
