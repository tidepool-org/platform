package dataservices

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/logger"
	"github.com/tidepool-org/platform/store"
	"github.com/tidepool-org/platform/user"
	"github.com/tidepool-org/platform/version"
)

const (
	missingPermissionsError = "missing required permissions"
	missingDataError        = "missing data to process"
	gettingDataError        = "there was an error getting your data"

	dataservicesName      = "dataservices"
	dataservicesStoreName = "deviceData"

	minimumSchemaVersion = 0
	maximumSchemaVersion = 99

	useridParamName = "userid"
)

var (
	log           = logger.Log.GetNamed(dataservicesName)
	serviceConfig *dataServiceConfig
)

func main() {

	port, err := config.FromEnv("TIDEPOOL_DATASERVICES_PORT")
	if err != nil {
		log.Fatal(err)
	}

	log.Fatal(NewDataServiceClient().Run(port))
}

type DataServiceClient struct {
	api              *rest.Api
	dataStore        store.Store
	validateToken    user.ChainedMiddleware
	attachPermissons user.ChainedMiddleware
	resolveGroupID   user.ChainedMiddleware
}

type dataServiceConfig struct {
	KeyFile  string `json:"keyFile"`
	CertFile string `json:"certFile"`
}

func NewDataServiceClient() *DataServiceClient {
	log.Info(version.Long())

	userClient := user.NewServicesClient()
	userClient.Start()

	config.FromJSON(&serviceConfig, "dataservices.json")

	return &DataServiceClient{
		api:              rest.NewApi(),
		dataStore:        store.NewMongoStore(dataservicesStoreName),
		validateToken:    user.NewAuthorizationMiddleware(userClient).ValidateToken,
		attachPermissons: user.NewMetadataMiddleware(userClient).GetPermissons,
		resolveGroupID:   user.NewMetadataMiddleware(userClient).GetGroupID,
	}

}

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

	protocol, err := config.FromEnv("TIDEPOOL_DATASERVICES_PROTOCOL")
	if err != nil {
		log.Fatal(err)
	}

	if protocol == "https" {
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

	var datumArray types.DatumArray

	err := r.DecodeJsonPayload(&datumArray)

	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	builder := data.NewTypeBuilder(map[string]interface{}{types.BaseUserIDField.Name: userid, types.BaseGroupIDField.Name: groupID})
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
func process(iter store.Iterator) types.DatumArray {

	var chunk types.Datum
	var all = types.DatumArray{}

	for iter.Next(&chunk) {
		all = append(all, chunk)
	}

	return all
}

func buildQuery(params url.Values) store.Query {

	theTypes := strings.Split(params.Get("type"), ",")
	subTypes := strings.Split(params.Get("subType"), ",")
	start := params.Get("startDate")
	end := params.Get("endDate")

	query := store.Query{}
	if len(theTypes) > 0 && theTypes[0] != "" {
		query[types.BaseTypeField.Name] = map[string]interface{}{store.In: theTypes}
	}
	if len(subTypes) > 0 && subTypes[0] != "" {
		query[types.BaseSubTypeField.Name] = map[string]interface{}{store.In: subTypes}
	}
	if start != "" && end != "" {
		query[types.TimeStringField.Name] = map[string]interface{}{store.GreaterThanEquals: start, store.LessThanEquals: end}
	} else if start != "" {
		query[types.TimeStringField.Name] = map[string]interface{}{store.GreaterThanEquals: start}
	} else if end != "" {
		query[types.TimeStringField.Name] = map[string]interface{}{store.LessThanEquals: end}
	}

	query["_schemaVersion"] = map[string]interface{}{store.GreaterThanEquals: minimumSchemaVersion, store.LessThanEquals: maximumSchemaVersion}

	return query
}

//GetDataset will return the requested users data set if permissons are sufficient
func (client *DataServiceClient) GetDataset(w rest.ResponseWriter, r *rest.Request) {

	log.AddTrace(r.PathParam(useridParamName))

	if !checkPermisson(r, user.Permission{}) {
		rest.Error(w, missingPermissionsError, http.StatusUnauthorized)
		return
	}

	groupID := r.Env[user.GROUPID]

	if groupID == "" {
		rest.Error(w, missingDataError, http.StatusBadRequest)
		return
	}

	var found struct {
		types.DatumArray `json:"Dataset"`
	}

	userid := r.PathParam(useridParamName)
	log.Info(useridParamName, userid)

	iter := client.dataStore.ReadAll(
		store.Field{Name: types.BaseInternalGroupIDField.Name, Value: groupID},
		buildQuery(r.URL.Query()),
		types.InternalFields,
	)
	defer iter.Close()

	found.DatumArray = process(iter)

	w.WriteJson(&found)
	return

}

//GetData will return the requested users data point if permissons are sufficient
func (client *DataServiceClient) GetData(w rest.ResponseWriter, r *rest.Request) {

	log.AddTrace(r.PathParam(useridParamName))

	if checkPermisson(r, user.Permission{}) {
		var foundDatum struct {
			types.Datum `json:"Datum"`
			Errors      string `json:"Errors"`
		}

		userid := r.PathParam(useridParamName)
		datumid := r.PathParam("datumid")

		log.Info("userid and datum", userid, datumid)

		foundDatum.Datum = types.Datum{}
		foundDatum.Errors = ""

		w.WriteJson(&foundDatum)
		return
	}
	rest.Error(w, missingPermissionsError, http.StatusUnauthorized)
	return
}
