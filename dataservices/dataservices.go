package main

import (
	"net/http"
	"strings"

	"github.com/tidepool-org/platform/data"
	log "github.com/tidepool-org/platform/logger"
	"github.com/tidepool-org/platform/store"
	"github.com/tidepool-org/platform/user"
	"github.com/tidepool-org/platform/version"

	"github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/ant0ine/go-json-rest/rest"
)

const (
	missingPermissionsError = "missing required permissions"
)

var (
	validateToken user.ChainedMiddleware
	getPermissons user.ChainedMiddleware
	dataStore     store.Store
)

func initMiddleware() {
	userClient := user.NewServicesClient()
	userClient.Start()
	validateToken = user.NewAuthorizationMiddleware(userClient).ValidateToken
	getPermissons = user.NewPermissonsMiddleware(userClient).GetPermissons
}

func main() {
	log.Logging.Info(version.String)

	initMiddleware()

	dataStore = store.NewMongoStore("dataservices")

	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	api.Use(&rest.GzipMiddleware{})

	log.Logging.AddTrace("123-546")

	router, err := rest.MakeRouter(
		rest.Get("/version", getVersion),
		rest.Get("/data/:userid/:datumid", validateToken(getPermissons(getData))),
		rest.Post("/dataset/:userid", validateToken(getPermissons(postDataset))),
		rest.Get("/dataset/:userid", validateToken(getPermissons(getDataset))),
	)
	if err != nil {
		log.Logging.Fatal(err)
	}
	api.SetApp(router)
	//TODO: config for statis port
	log.Logging.Fatal(http.ListenAndServe(":8077", api.MakeHandler()))
}

func checkPermisson(r *rest.Request, expected user.Permission) bool {
	//userid := r.PathParam("userid")
	if permissions := r.Env[user.PERMISSIONS].(*user.UsersPermissions); permissions != nil {

		log.Logging.Info("perms found ", permissions)

		//	perms := permissions[userid]
		//	if perms != nil && perms[""] != nil {
		return true
		//	}
	}
	return false
}

func getVersion(w rest.ResponseWriter, r *rest.Request) {
	w.WriteJson(version.String)
}

func postDataset(w rest.ResponseWriter, r *rest.Request) {

	log.Logging.AddTrace(r.PathParam("userid"))

	if checkPermisson(r, user.Permission{}) {

		var dataSet data.GenericDataset
		var processedDataset struct {
			Dataset []interface{} `json:"Dataset"`
			Errors  string        `json:"Errors"`
		}

		log.Logging.Info("processing")

		err := r.DecodeJsonPayload(&dataSet)

		if err != nil {
			rest.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		data, err := data.NewTypeBuilder().BuildFromDataSet(dataSet)
		processedDataset.Dataset = data

		if err != nil {
			rest.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		//TODO: should this be a bulk insert?
		for i := range data {
			err := dataStore.Save(data[i])
			if err != nil {
				rest.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		w.WriteJson(&processedDataset)
		return
	}
	rest.Error(w, missingPermissionsError, http.StatusUnauthorized)
	return
}

func getDataset(w rest.ResponseWriter, r *rest.Request) {

	log.Logging.AddTrace(r.PathParam("userid"))

	if checkPermisson(r, user.Permission{}) {

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

		var dataSet data.GenericDataset
		err := dataStore.ReadAll(store.IDField{Name: "userId", Value: userid}, &dataSet)

		if err != nil {
			foundDataset.Errors = err.Error()
		}
		foundDataset.GenericDataset = dataSet

		w.WriteJson(&foundDataset)
		return
	}
	rest.Error(w, missingPermissionsError, http.StatusUnauthorized)
	return
}

func getData(w rest.ResponseWriter, r *rest.Request) {

	log.Logging.AddTrace(r.PathParam("userid"))

	if checkPermisson(r, user.Permission{}) {
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
	rest.Error(w, missingPermissionsError, http.StatusUnauthorized)
	return
}
