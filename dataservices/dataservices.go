package dataservices

import (
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/store"
	"github.com/tidepool-org/platform/store/mongo"
	"github.com/tidepool-org/platform/user"
)

const (
	missingPermissionsError = "missing required permissions"
	missingDataError        = "missing data to process"
	gettingDataError        = "there was an error getting your data"

	dataservicesStoreName = "deviceData"

	//TODO: this will removed when updated store is integrated
	minimumSchemaVersion = 0
	currentSchemaVersion = 10

	useridParamName = "userid"
)

type Server struct {
	logger           log.Logger
	config           *Config
	dataStore        store.Store
	api              *rest.Api
	statusMiddleware *rest.StatusMiddleware
	validateToken    user.ChainedMiddleware
	attachPermissons user.ChainedMiddleware
	resolveGroupID   user.ChainedMiddleware
}

type TLS struct {
	CertificateFile string `json:"certificateFile"`
	KeyFile         string `json:"keyFile"`
}

type Config struct {
	Address string `json:"address"`
	TLS     *TLS   `json:"tls"`
}

func (c *Config) Validate() error {
	if c.Address == "" {
		return app.Error("dataservices", "address is not specified")
	}
	if c.TLS != nil {
		if c.TLS.CertificateFile == "" {
			return app.Error("dataservices", "tls certificate file is not specified")
		}
		if c.TLS.KeyFile == "" {
			return app.Error("dataservices", "tls key file is not specified")
		}
	}
	return nil
}

func NewServer(logger log.Logger) (*Server, error) {
	logger.Info("Starting data services server")

	dataservicesConfig := &Config{}
	if err := config.Load("dataservices", dataservicesConfig); err != nil {
		return nil, app.ExtError(err, "dataservices", "unable to load config")
	}
	if err := dataservicesConfig.Validate(); err != nil {
		return nil, app.ExtError(err, "dataservices", "config is not valid")
	}

	dataStoreConfig := &mongo.Config{}
	if err := config.Load("data_store", dataStoreConfig); err != nil {
		return nil, app.ExtError(err, "dataservices", "unable to load data store config")
	}
	dataStoreConfig.Collection = "deviceData"

	dataStore, err := mongo.NewStore(dataStoreConfig, logger)
	if err != nil {
		return nil, app.ExtError(err, "dataservices", "unable to create data store")
	}

	return &Server{
		logger:    logger,
		config:    dataservicesConfig,
		dataStore: dataStore,
		api:       rest.NewApi(),
	}, nil
}

func (s *Server) Close() {
	if s.dataStore != nil {
		s.dataStore.Close()
		s.dataStore = nil
	}
}

func (s *Server) Run() error {
	if err := s.setupMiddleware(); err != nil {
		return err
	}
	if err := s.setupRouter(); err != nil {
		return err
	}
	return s.serve()
}

func (s *Server) setupMiddleware() error {
	userClient, err := user.NewServicesClient(s.logger)
	if err != nil {
		return app.ExtError(err, "dataservices", "unable to create services client")
	}
	if err := userClient.Start(); err != nil {
		return app.ExtError(err, "dataservices", "unable to start services client")
	}

	s.validateToken = user.NewAuthorizationMiddleware(userClient).ValidateToken
	s.attachPermissons = user.NewMetadataMiddleware(userClient).GetPermissons
	s.resolveGroupID = user.NewMetadataMiddleware(userClient).GetGroupID
	// TODO: Add content checker

	s.statusMiddleware = &rest.StatusMiddleware{}

	s.api.Use(&service.LoggerMiddleware{Logger: s.logger})
	s.api.Use(&service.AccessLogMiddleware{})
	s.api.Use(s.statusMiddleware)
	s.api.Use(&rest.TimerMiddleware{})
	s.api.Use(&rest.RecorderMiddleware{})
	s.api.Use(&service.RecoverMiddleware{})
	s.api.Use(&rest.GzipMiddleware{})
	// s.api.Use(&rest.CorsMiddleware{})	// TODO: Need configuration

	return nil
}

func (s *Server) setupRouter() error {
	router, err := rest.MakeRouter(
		rest.Get("/version", s.GetVersion),
		rest.Get("/status", s.GetStatus),
		// rest.Get("/data/:userid/:datumid", s.validateToken(s.resolveGroupID((s.GetData)))),
		// rest.Post("/dataset/:userid", s.validateToken(s.attachPermissons(s.resolveGroupID(s.PostDataset)))),
		// rest.Get("/dataset/:userid", s.validateToken(s.attachPermissons(s.resolveGroupID(s.GetDataset)))),
		// rest.Post("/blob/:userid", s.validateToken(s.attachPermissons(s.resolveGroupID(s.PostBlob)))),

		rest.Post("/api/v1/users/:userid/datasets", s.validateToken(s.resolveGroupID(s.withContext(s.DatasetCreate)))),
		rest.Put("/api/v1/datasets/:datasetid", s.validateToken(s.withContext(s.DatasetUpdate))),
		rest.Post("/api/v1/datasets/:datasetid/data", s.validateToken(s.withContext(s.DatasetDataCreate))),
		// TODO: POST /api/v1/users/:userid/data - all 3 above at once
	)
	if err != nil {
		return app.ExtError(err, "server", "unable to setup router")
	}

	s.api.SetApp(router)
	return nil
}

func (s *Server) serve() (err error) {
	if s.config.TLS != nil {
		err = http.ListenAndServeTLS(s.config.Address, s.config.TLS.CertificateFile, s.config.TLS.KeyFile, s.api.MakeHandler())
	} else {
		err = http.ListenAndServe(s.config.Address, s.api.MakeHandler())
	}
	return err
}

func (s *Server) withContext(handler HandlerFunc) rest.HandlerFunc {
	return WithContext(s.dataStore, handler)
}

// TODO: Fix all this

//checkPermisson will check that we have the expected permisson
// func checkPermisson(r *rest.Request, expected user.Permission) bool {
// 	//TODO: fill in the details
// 	if permissions := r.Env[user.PERMISSIONS].(*user.UsersPermissions); permissions != nil {
// 		return true
// 	}
// 	return false
// }

// //PostDataset will process a posted dataset for the requested user if permissons are sufficient
// func (client *Server) PostDataset(w rest.ResponseWriter, r *rest.Request) {

// 	userid := r.PathParam(useridParamName)

// 	if !checkPermisson(r, user.Permission{}) {
// 		rest.Error(w, missingPermissionsError, http.StatusUnauthorized)
// 		return
// 	}

// 	groupID := r.Env[user.GROUPID]

// 	if r.ContentLength == 0 || groupID == "" {
// 		rest.Error(w, missingDataError, http.StatusBadRequest)
// 		return
// 	}

// 	var datumArray types.DatumArray

// 	err := r.DecodeJsonPayload(&datumArray)

// 	if err != nil {
// 		rest.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	builder := data.NewTypeBuilder(map[string]interface{}{types.BaseUserIDField.Name: userid, types.BaseGroupIDField.Name: groupID})
// 	platformData, platformErrors := builder.BuildFromDatumArray(datumArray)

// 	if platformErrors != nil && platformErrors.HasErrors() {
// 		w.WriteHeader(http.StatusBadRequest)
// 		w.WriteJson(&platformErrors.Errors)
// 		return
// 	}

// 	//TODO: should this be a bulk insert?
// 	for i := range platformData {
// 		if err = client.dataStore.Save(platformData[i]); err != nil {
// 			rest.Error(w, err.Error(), http.StatusInternalServerError)
// 			return
// 		}
// 	}
// 	w.WriteHeader(http.StatusOK)
// 	return
// }

// //PostBlob will save a posted blob of  data for the requested user if permissons are sufficient
// func (client *Server) PostBlob(w rest.ResponseWriter, r *rest.Request) {

// 	if checkPermisson(r, user.Permission{}) {
// 		w.WriteHeader(http.StatusNotImplemented)
// 		return
// 	}
// 	rest.Error(w, missingPermissionsError, http.StatusUnauthorized)
// 	return

// }

//process the found data and send the appropriate response
// func process(iter store.Iterator) types.DatumArray {

// 	var chunk types.Datum
// 	var all = types.DatumArray{}

// 	for iter.Next(&chunk) {
// 		all = append(all, chunk)
// 	}

// 	return all
// }

// func buildQuery(params url.Values) store.Query {

// 	theTypes := strings.Split(params.Get("type"), ",")
// 	subTypes := strings.Split(params.Get("subType"), ",")
// 	start := params.Get("startDate")
// 	end := params.Get("endDate")

// 	query := store.Query{}
// 	if len(theTypes) > 0 && theTypes[0] != "" {
// 		query[types.BaseTypeField.Name] = map[string]interface{}{store.In: theTypes}
// 	}
// 	if len(subTypes) > 0 && subTypes[0] != "" {
// 		query[types.BaseSubTypeField.Name] = map[string]interface{}{store.In: subTypes}
// 	}
// 	if start != "" && end != "" {
// 		query[types.TimeStringField.Name] = map[string]interface{}{store.GreaterThanEquals: start, store.LessThanEquals: end}
// 	} else if start != "" {
// 		query[types.TimeStringField.Name] = map[string]interface{}{store.GreaterThanEquals: start}
// 	} else if end != "" {
// 		query[types.TimeStringField.Name] = map[string]interface{}{store.LessThanEquals: end}
// 	}

//	query["_schemaVersion"] = map[string]interface{}{store.GreaterThanEquals: minimumSchemaVersion, store.LessThanEquals: currentSchemaVersion}

// 	return query
// }

// //GetDataset will return the requested users data set if permissons are sufficient
// func (client *Server) GetDataset(w rest.ResponseWriter, r *rest.Request) {

// 	if !checkPermisson(r, user.Permission{}) {
// 		rest.Error(w, missingPermissionsError, http.StatusUnauthorized)
// 		return
// 	}

// 	groupID := r.Env[user.GROUPID]

// 	if groupID == "" {
// 		rest.Error(w, missingDataError, http.StatusBadRequest)
// 		return
// 	}

// 	var found struct {
// 		types.DatumArray `json:"Dataset"`
// 	}

// 	userid := r.PathParam(useridParamName)
// 	client.logger.WithField(useridParamName, userid).Info("GetDataset")

// 	iter := client.dataStore.ReadAll(
// 		store.Field{Name: types.BaseGroupIDField.Name, Value: groupID},
// 		buildQuery(r.URL.Query()),
// 		[]string{},
// 		types.InternalFields,
// 	)
// 	defer iter.Close()

// 	found.DatumArray = process(iter)

// 	w.WriteJson(&found)
// 	return

// }

// //GetData will return the requested users data point if permissons are sufficient
// func (client *Server) GetData(w rest.ResponseWriter, r *rest.Request) {

// 	if checkPermisson(r, user.Permission{}) {
// 		var foundDatum struct {
// 			types.Datum `json:"Datum"`
// 			Errors      string `json:"Errors"`
// 		}

// 		userid := r.PathParam(useridParamName)
// 		datumid := r.PathParam("datumid")

// 		client.logger.WithFields(map[string]interface{}{"userid": userid, "datumid": datumid}).Info("GetData")

// 		foundDatum.Datum = types.Datum{}
// 		foundDatum.Errors = ""

// 		w.WriteJson(&foundDatum)
// 		return
// 	}
// 	rest.Error(w, missingPermissionsError, http.StatusUnauthorized)
// 	return
// }
