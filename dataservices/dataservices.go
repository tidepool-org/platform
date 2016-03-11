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
	//TODO: from config
	log.Fatal(NewDataServiceClient().Run(":8077"))
}

//DataServiceClient for the data service
type DataServiceClient struct {
	api              *rest.Api
	dataStore        store.Store
	validateToken    user.ChainedMiddleware
	attachPermissons user.ChainedMiddleware
	resolveGroupID   user.ChainedMiddleware
}

//NewDataServiceClient returns an initialised client
func NewDataServiceClient() *DataServiceClient {
	log.Info(version.String)

	userClient := user.NewServicesClient()
	userClient.Start()

	return &DataServiceClient{
		api: rest.NewApi(),
		//TODO: from config
		dataStore:        store.NewMongoStore("deviceData"),
		validateToken:    user.NewAuthorizationMiddleware(userClient).ValidateToken,
		attachPermissons: user.NewMetadataMiddleware(userClient).GetPermissons,
		resolveGroupID:   user.NewMetadataMiddleware(userClient).GetGroupID,
	}

}

//Run will run the service
func (client *DataServiceClient) Run(URL string) error {
	client.api.Use(rest.DefaultDevStack...)
	client.api.Use(&rest.GzipMiddleware{})

	router, err := rest.MakeRouter(
		rest.Get("/version", client.GetVersion),
		rest.Get("/data/:userid/:datumid", client.validateToken(client.resolveGroupID((client.GetData)))),
		rest.Post("/dataset/:userid", client.validateToken(client.attachPermissons(client.resolveGroupID(client.PostDataset)))),
		rest.Get("/dataset/:userid", client.validateToken(client.attachPermissons(client.resolveGroupID(client.GetDataset)))),
		rest.Post("/blob/:userid", client.validateToken(client.attachPermissons(client.resolveGroupID(client.PostBlob)))),
	)
	if err != nil {
		log.Fatal(err)
	}
	client.api.SetApp(router)

	return http.ListenAndServe(URL, client.api.MakeHandler())
}

//checkPermisson will check that we have the expected permisson
func checkPermisson(r *rest.Request, expected user.Permission) bool {
	//TODO: fill in the details
	if permissions := r.Env[user.PERMISSIONS].(*user.UsersPermissions); permissions != nil {
		return true
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

	userid := r.PathParam(useridParamName)

	log.AddTrace(userid)

	if checkPermisson(r, user.Permission{}) {

		groupID := r.Env[user.GROUPID]

		if r.ContentLength == 0 || groupID == "" {
			rest.Error(w, missingDataError, http.StatusBadRequest)
			return
		}

		var dataSet data.Dataset
		var processedDataset struct {
			Dataset []interface{} `json:"Dataset"`
			Errors  string        `json:"Errors"`
		}

		err := r.DecodeJsonPayload(&dataSet)

		if err != nil {
			rest.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		platformData, err := data.NewTypeBuilder(map[string]interface{}{data.UserIDField: userid, data.GroupIDField: groupID}).BuildFromDataSet(dataSet)
		processedDataset.Dataset = platformData
		processedDataset.Errors = err.Error()

		//TODO: should this be a bulk insert?
		for i := range platformData {
			if err = client.dataStore.Save(platformData[i]); err != nil {
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

//PostBlob will save a posted blob of  data for the requested user if permissons are sufficient
func (client *DataServiceClient) PostBlob(w rest.ResponseWriter, r *rest.Request) {

	userid := r.PathParam(useridParamName)
	log.AddTrace(userid)

	if checkPermisson(r, user.Permission{}) {
		w.WriteHeader(http.StatusNotImplemented)
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
			data.Dataset `json:"Dataset"`
			Errors       string `json:"Errors"`
		}

		userid := r.PathParam(useridParamName)
		log.Info(useridParamName, userid)

		types := strings.Split(r.URL.Query().Get("type"), ",")
		subTypes := strings.Split(r.URL.Query().Get("subType"), ",")
		start := r.URL.Query().Get("startDate")
		end := r.URL.Query().Get("endDate")

		log.Info("params", types, subTypes, start, end)

		var dataSet data.Dataset
		err := client.dataStore.ReadAll(store.IDField{Name: data.UserIDField, Value: userid}, &dataSet)

		if err != nil {
			foundDataset.Errors = err.Error()
		}
		foundDataset.Dataset = dataSet

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
			data.Datum `json:"Datum"`
			Errors     string `json:"Errors"`
		}

		userid := r.PathParam(useridParamName)
		datumid := r.PathParam("datumid")

		log.Info("userid and datum", userid, datumid)

		foundDatum.Datum = data.Datum{}
		foundDatum.Errors = ""

		w.WriteJson(&foundDatum)
		return
	}
	rest.Error(w, missingPermissionsError, http.StatusUnauthorized)
	return
}
