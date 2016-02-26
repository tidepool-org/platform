package main

import (
	"net/http"
	"strings"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/logger"
	"github.com/tidepool-org/platform/store"
	"github.com/tidepool-org/platform/user"
	"github.com/tidepool-org/platform/version"

	"github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/ant0ine/go-json-rest/rest"
)

const (
	missingPermissionsError = "missing required permissions"
	missingDataError        = "missing data to process"

	dataservicesName = "dataservices"
	useridParamName  = "userid"
)

var (
	log = logger.Log.GetNamed(dataservicesName)
)

func main() {
	log.Fatal(NewDataServiceClient().Run(":8077"))
}

//DataServiceClient for the data service
type DataServiceClient struct {
	api           *rest.Api
	dataStore     store.Store
	validateToken user.ChainedMiddleware
	getPermissons user.ChainedMiddleware
}

//NewDataServiceClient returns an initialised client
func NewDataServiceClient() *DataServiceClient {
	log.Info(version.String)

	userClient := user.NewServicesClient()
	userClient.Start()

	return &DataServiceClient{
		api:           rest.NewApi(),
		dataStore:     store.NewMongoStore(dataservicesName),
		validateToken: user.NewAuthorizationMiddleware(userClient).ValidateToken,
		getPermissons: user.NewPermissonsMiddleware(userClient).GetPermissons,
	}

}

//Run will run the service
func (client *DataServiceClient) Run(URL string) error {
	client.api.Use(rest.DefaultDevStack...)
	client.api.Use(&rest.GzipMiddleware{})

	router, err := rest.MakeRouter(
		rest.Get("/version", client.GetVersion),
		rest.Get("/data/:userid/:datumid", client.validateToken(client.getPermissons(client.GetData))),
		rest.Post("/dataset/:userid", client.validateToken(client.getPermissons(client.PostDataset))),
		rest.Get("/dataset/:userid", client.validateToken(client.getPermissons(client.GetDataset))),
	)
	if err != nil {
		log.Fatal(err)
	}
	client.api.SetApp(router)

	return http.ListenAndServe(URL, client.api.MakeHandler())
}

//checkPermisson will check that we have the expected permisson
func checkPermisson(r *rest.Request, expected user.Permission) bool {
	//userid := r.PathParam("userid")
	if permissions := r.Env[user.PERMISSIONS].(*user.UsersPermissions); permissions != nil {

		log.Info("perms found ", permissions)

		//	perms := permissions[userid]
		//	if perms != nil && perms[""] != nil {
		return true
		//	}
	}
	return false
}

//GetVersion will return the current API version
func (client *DataServiceClient) GetVersion(w rest.ResponseWriter, r *rest.Request) {
	w.WriteJson(&version.String)
	return
}

//PostDataset will process a posted dataset for the requested user if permissons are sufficient
func (client *DataServiceClient) PostDataset(w rest.ResponseWriter, r *rest.Request) {

	log.AddTrace(r.PathParam(useridParamName))

	if checkPermisson(r, user.Permission{}) {

		if r.ContentLength == 0 {
			rest.Error(w, missingDataError, http.StatusBadRequest)
			return
		}

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

		//TODO: should this be a bulk insert?
		for i := range data {
			err := client.dataStore.Save(data[i])
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

//GetDataset will return the requested users data set if permissons are sufficient
func (client *DataServiceClient) GetDataset(w rest.ResponseWriter, r *rest.Request) {

	log.AddTrace(r.PathParam(useridParamName))

	if checkPermisson(r, user.Permission{}) {

		var foundDataset struct {
			data.GenericDataset `json:"Dataset"`
			Errors              string `json:"Errors"`
		}

		userid := r.PathParam(useridParamName)
		log.Info(useridParamName, userid)

		types := strings.Split(r.URL.Query().Get("type"), ",")
		subTypes := strings.Split(r.URL.Query().Get("subType"), ",")
		start := r.URL.Query().Get("startDate")
		end := r.URL.Query().Get("endDate")

		log.Info("params", types, subTypes, start, end)

		var dataSet data.GenericDataset
		err := client.dataStore.ReadAll(store.IDField{Name: "userId", Value: userid}, &dataSet)

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

//GetData will return the requested users data point if permissons are sufficient
func (client *DataServiceClient) GetData(w rest.ResponseWriter, r *rest.Request) {

	log.AddTrace(r.PathParam(useridParamName))

	if checkPermisson(r, user.Permission{}) {
		var foundDatum struct {
			data.GenericDatam `json:"Datum"`
			Errors            string `json:"Errors"`
		}

		userid := r.PathParam(useridParamName)
		datumid := r.PathParam("datumid")

		log.Info("userid and datum", userid, datumid)

		foundDatum.GenericDatam = data.GenericDatam{}
		foundDatum.Errors = ""

		w.WriteJson(&foundDatum)
		return
	}
	rest.Error(w, missingPermissionsError, http.StatusUnauthorized)
	return
}
