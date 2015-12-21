package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/user"
	"github.com/tidepool-org/platform/version"

	"github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/ant0ine/go-json-rest/rest"
)

//TODO: faking it for now
type FakeUserClient struct{}

func (c FakeUserClient) Start() error { return nil }
func (c FakeUserClient) Close()       { return }
func (c FakeUserClient) CheckToken(token string) *user.ClientTokenData {
	fmt.Println("checking token...")
	return &user.ClientTokenData{}
}
func (c FakeUserClient) GetUser(userID, token string) (*user.ClientData, error) {
	return &user.ClientData{}, nil
}

func main() {
	fmt.Println(version.String)
	fmt.Println(data.GetData())

	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	api.Use(&rest.GzipMiddleware{})
	api.Use(&user.AuthorizationMiddleware{Client: &FakeUserClient{}})

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
	var processedDataset struct {
		Dataset []interface{} `json:"Dataset"`
		Errors  string        `json:"Errors"`
	}

	err := r.DecodeJsonPayload(&dataSet)

	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := data.NewTypeBuilder().BuildFromDataSet(dataSet)

	processedDataset.Dataset = data
	processedDataset.Errors = err.Error()

	w.WriteJson(&processedDataset)
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
