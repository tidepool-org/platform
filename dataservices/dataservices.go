package main

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/tidepool-org/platform/config"
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
	gettingDataError        = "there was an error getting your data"

	dataservicesName = "dataservices"
	useridParamName  = "userid"
)

var (
	log           = logger.Log.GetNamed(dataservicesName)
	serviceConfig *dataServiceConfig
)

func main() {
	log.Fatal(NewDataServiceClient().Run(serviceConfig.Port))
}

//DataServiceClient for the data service
type DataServiceClient struct {
	api              *rest.Api
	dataStore        store.Store
	validateToken    user.ChainedMiddleware
	attachPermissons user.ChainedMiddleware
	resolveGroupID   user.ChainedMiddleware
}

type dataServiceConfig struct {
	Port          string `json:"port"`
	Protocol      string `json:"protocol"`
	KeyFile       string `json:"keyFile"`
	CertFile      string `json:"certFile"`
	SchemaVersion struct {
		Minimum int
		Maximum int
	} `json:"schemaVersion"`
	StoreName string `json:"storeName"`
}

//NewDataServiceClient returns an initialised client
func NewDataServiceClient() *DataServiceClient {
	log.Info(version.Long())

	userClient := user.NewServicesClient()
	userClient.Start()

	config.FromJSON(&serviceConfig, "dataservices.json")

	return &DataServiceClient{
		api:              rest.NewApi(),
		dataStore:        store.NewMongoStore(serviceConfig.StoreName),
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

	if serviceConfig.Protocol == "https" {
		return http.ListenAndServeTLS(URL, serviceConfig.CertFile, serviceConfig.KeyFile, client.api.MakeHandler())
	}
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
	w.WriteJson(version.Long())
	return
}

//PostDataset will process a posted dataset for the requested user if permissons are sufficient
func (client *DataServiceClient) PostDataset(w rest.ResponseWriter, r *rest.Request) {

	userid := r.PathParam(useridParamName)

	log.AddTrace(userid)

	if !checkPermisson(r, user.Permission{}) {
		rest.Error(w, missingPermissionsError, http.StatusUnauthorized)
		return
	}

	groupID := r.Env[user.GROUPID]

	if r.ContentLength == 0 || groupID == "" {
		rest.Error(w, missingDataError, http.StatusBadRequest)
		return
	}

	var datumArray data.DatumArray

	err := r.DecodeJsonPayload(&datumArray)

	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	builder := data.NewTypeBuilder(map[string]interface{}{data.UserIDField: userid, data.GroupIDField: groupID})
	platformData, platformErrors := builder.BuildFromDatumArray(datumArray)

	if platformErrors != nil && platformErrors.HasErrors() {
		w.WriteHeader(http.StatusBadRequest)
		w.WriteJson(&platformErrors.Errors)
		return
	}

	//TODO: should this be a bulk insert?
	for i := range platformData {
		if err = client.dataStore.Save(platformData[i]); err != nil {
			rest.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	w.WriteHeader(http.StatusOK)
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

//process the found data and send the appropriate response
func process(iter store.Iterator) data.DatumArray {

	var chunk data.Datum
	var all = data.DatumArray{}

	for iter.Next(&chunk) {
		all = append(all, chunk)
	}

	return all
}

func buildQuery(params url.Values) store.Query {

	types := strings.Split(params.Get("type"), ",")
	subTypes := strings.Split(params.Get("subType"), ",")
	start := params.Get("startDate")
	end := params.Get("endDate")

	query := store.Query{}
	if len(types) > 0 && types[0] != "" {
		query[data.TypeField] = map[string]interface{}{store.In: types}
	}
	if len(subTypes) > 0 && subTypes[0] != "" {
		query[data.SubTypeField] = map[string]interface{}{store.In: subTypes}
	}
	if start != "" && end != "" {
		query[data.TimeField] = map[string]interface{}{store.GreaterThanEquals: start, store.LessThanEquals: end}
	} else if start != "" {
		query[data.TimeField] = map[string]interface{}{store.GreaterThanEquals: start}
	} else if end != "" {
		query[data.TimeField] = map[string]interface{}{store.LessThanEquals: end}
	}
	return query
}

//GetDataset will return the requested users data set if permissons are sufficient
func (client *DataServiceClient) GetDataset(w rest.ResponseWriter, r *rest.Request) {

	log.AddTrace(r.PathParam(useridParamName))

	if checkPermisson(r, user.Permission{}) {

		groupID := r.Env[user.GROUPID]

		if groupID == "" {
			rest.Error(w, missingDataError, http.StatusBadRequest)
			return
		}

		var found struct {
			data.DatumArray `json:"Dataset"`
			Errors          string `json:"Errors"`
		}

		userid := r.PathParam(useridParamName)
		log.Info(useridParamName, userid)

		iter := client.dataStore.ReadAll(
			store.Field{Name: data.InternalGroupIDField, Value: groupID},
			buildQuery(r.URL.Query()),
			data.InternalFields,
		)
		defer iter.Close()

		found.DatumArray = process(iter)

		w.WriteJson(&found)
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
