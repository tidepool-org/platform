package main

import (
	"net/http"
	"strings"

	"github.com/tidepool-org/platform/data"
	log "github.com/tidepool-org/platform/logger"
	"github.com/tidepool-org/platform/user"
	"github.com/tidepool-org/platform/version"

	"github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/ant0ine/go-json-rest/rest"
)

const (
	error_userid_required  = "userid required"
	error_datumid_required = "datumid required"
)

var authorizationMiddleware *user.AuthorizationMiddleware

func initAuthorizationMiddleware() {
	userClient := user.NewUserServicesClient()
	userClient.Start()
	authorizationMiddleware = user.NewAuthorizationMiddleware(userClient)
}

func main() {
	log.Logging.Info(version.String)
	log.Logging.Info(data.GetData())

	initAuthorizationMiddleware()

	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	api.Use(&rest.GzipMiddleware{})

	router, err := rest.MakeRouter(
		rest.Get("/version", getVersion),
		rest.Get("/data/:userid/:datumid", authorizationMiddleware.MiddlewareFunc(getData)),
		rest.Post("/dataset/:userid", authorizationMiddleware.MiddlewareFunc(postDataset)),
		rest.Get("/dataset/:userid", authorizationMiddleware.MiddlewareFunc(getDataset)),
	)
	if err != nil {
		log.Logging.Fatal(err)
	}
	api.SetApp(router)
	//TODO: config
	log.Logging.Fatal(http.ListenAndServe(":8077", api.MakeHandler()))
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

	userid := r.PathParam("userid")
	log.Logging.Info("userid", userid)

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

	var foundDataset struct {
		data.GenericDataset `json:"Dataset"`
		Errors              string `json:"Errors"`
	}

	userid := r.PathParam("userid")
	log.Logging.Info("userid", userid)

	types := strings.Split(r.URL.Query().Get("type"), ",")
	subTypes := strings.Split(r.URL.Query().Get("subType"), ",")
	start := r.URL.Query().Get("startDate")
	end := r.URL.Query().Get("endDate")

	log.Logging.Info("params", types, subTypes, start, end)

	foundDataset.GenericDataset = data.GenericDataset{}
	foundDataset.Errors = ""

	w.WriteJson(&foundDataset)
	return

}

func getData(w rest.ResponseWriter, r *rest.Request) {
	var foundDatum struct {
		data.GenericDatam `json:"Datum"`
		Errors            string `json:"Errors"`
	}

	userid := r.PathParam("userid")
	datumid := r.PathParam("datumid")

	log.Logging.Info("userid and datum", userid, datumid)

	foundDatum.GenericDatam = data.GenericDatam{}
	foundDatum.Errors = ""

	w.WriteJson(&foundDatum)
	return
}
