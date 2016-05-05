package server

/* CHECKLIST
 * [ ] Uses interfaces as appropriate
 * [ ] Private package variables use underscore prefix
 * [ ] All parameters validated
 * [ ] All errors handled
 * [ ] Reviewed for concurrency safety
 * [ ] Code complete
 * [ ] Full test coverage
 */

import (
	"net/http"
	"time"

	"github.com/ant0ine/go-json-rest/rest"
	graceful "gopkg.in/tylerb/graceful.v1"

	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/dataservices/server/api"
	"github.com/tidepool-org/platform/dataservices/server/api/v1"
	"github.com/tidepool-org/platform/dataservices/server/api/v1/middleware"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/store"
	"github.com/tidepool-org/platform/store/mongo"
	"github.com/tidepool-org/platform/userservices/client"
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
	store            store.Store
	client           client.Client
	api              *rest.Api
	statusMiddleware *rest.StatusMiddleware
}

func New(logger log.Logger) (*Server, error) {

	logger.Debug("Loading dataservices server config")

	dataservicesConfig := &Config{}
	if err := config.Load("dataservices", dataservicesConfig); err != nil {
		return nil, app.ExtError(err, "dataservices", "unable to load config")
	}
	if err := dataservicesConfig.Validate(); err != nil {
		return nil, app.ExtError(err, "dataservices", "config is not valid")
	}

	logger.Debug("Creating dataservices server")

	return &Server{
		logger: logger,
		config: dataservicesConfig,
	}, nil
}

func (s *Server) Close() {
	if s.client != nil {
		s.client.Close()
		s.client = nil
	}
	if s.store != nil {
		s.store.Close()
		s.store = nil
	}
}

func (s *Server) Run() error {
	if err := s.setupStore(); err != nil {
		return err
	}
	if err := s.setupClient(); err != nil {
		return err
	}
	if err := s.setupAPI(); err != nil {
		return err
	}

	return s.serve()
}

func (s *Server) setupStore() error {

	// TODO: Consider alternate data stores

	s.logger.Debug("Loading mongo data store config")

	mongoDataStoreConfig := &mongo.Config{}
	if err := config.Load("data_store", mongoDataStoreConfig); err != nil {
		return app.ExtError(err, "dataservices", "unable to load mongo data store config")
	}
	mongoDataStoreConfig.Collection = "deviceData"

	s.logger.Debug("Creating mongo data store")

	mongoDataStore, err := mongo.New(mongoDataStoreConfig, s.logger)
	if err != nil {
		return app.ExtError(err, "dataservices", "unable to create mongo data store")
	}

	s.store = mongoDataStore

	return nil
}

func (s *Server) setupClient() error {

	s.logger.Debug("Loading userservices client config")

	userservicesClientConfig := &client.Config{}
	if err := config.Load("userservices_client", userservicesClientConfig); err != nil {
		return app.ExtError(err, "dataservices", "unable to load userservices client config")
	}

	s.logger.Debug("Creating userservices client")

	userservicesClient, err := client.NewStandard(userservicesClientConfig, s.logger)
	if err != nil {
		return app.ExtError(err, "dataservices", "unable to create userservices client")
	}

	s.logger.Debug("Starting userservices client")
	if err := userservicesClient.Start(); err != nil {

		return app.ExtError(err, "dataservices", "unable to start userservices client")
	}

	s.client = userservicesClient

	return nil
}

func (s *Server) setupAPI() error {

	s.logger.Debug("Creating API")

	s.api = rest.NewApi()

	if err := s.setupAPIMiddleware(); err != nil {
		return err
	}

	if err := s.setupAPIRouter(); err != nil {
		return err
	}

	return nil
}

func (s *Server) setupAPIMiddleware() error {

	s.logger.Debug("Creating API middleware")

	loggerMiddleware, err := service.NewLoggerMiddleware(s.logger)
	if err != nil {
		return err
	}
	traceMiddleware, err := service.NewTraceMiddleware()
	if err != nil {
		return err
	}
	accessLogMiddleware, err := service.NewAccessLogMiddleware()
	if err != nil {
		return err
	}
	recoverMiddleware, err := service.NewRecoverMiddleware()
	if err != nil {
		return err
	}

	statusMiddleware := &rest.StatusMiddleware{}
	timerMiddleware := &rest.TimerMiddleware{}
	recorderMiddleware := &rest.RecorderMiddleware{}
	gzipMiddleware := &rest.GzipMiddleware{}

	middlewareStack := []rest.Middleware{
		loggerMiddleware,
		traceMiddleware,
		accessLogMiddleware,
		statusMiddleware,
		timerMiddleware,
		recorderMiddleware,
		recoverMiddleware,
		gzipMiddleware,
	}

	s.api.Use(middlewareStack...)

	s.statusMiddleware = statusMiddleware

	// s.validateToken = client.NewAuthorizationMiddlewareOld(s.client).ValidateToken
	// s.attachPermissons = client.NewMetadataMiddleware(s.client).GetPermissons
	// s.resolveGroupID = client.NewMetadataMiddleware(s.client).GetGroupID
	// TODO: Add content checker

	// s.authorizationMiddleware = client.NewAuthorizationMiddleware(s.client)

	return nil
}

func (s *Server) setupAPIRouter() error {

	s.logger.Debug("Creating API router")

	router, err := rest.MakeRouter(
		rest.Get("/status", s.GetStatus),
		rest.Get("/version", s.GetVersion),
		rest.Get("/api/v1/users/:userid/check", s.withContext(middleware.Authenticate(v1.UsersCheck))),
		// rest.Post("/api/v1/users/:userid/datasets", s.validateToken(s.resolveGroupID(s.withContext(v1.DatasetCreate)))),
		// rest.Put("/api/v1/datasets/:datasetid", s.validateToken(s.withContext(v1.DatasetUpdate))),
		// rest.Post("/api/v1/datasets/:datasetid/data", s.validateToken(s.withContext(v1.DatasetDataCreate))),
	)
	if err != nil {
		return app.ExtError(err, "server", "unable to setup router")
	}

	s.api.SetApp(router)

	return nil
}

func (s *Server) serve() error {
	server := &graceful.Server{
		Timeout: time.Duration(s.config.Timeout) * time.Second,
		Server: &http.Server{
			Addr:    s.config.Address,
			Handler: s.api.MakeHandler(),
		},
	}

	var err error
	if s.config.TLS != nil {
		err = server.ListenAndServeTLS(s.config.TLS.CertificateFile, s.config.TLS.KeyFile)
	} else {
		err = server.ListenAndServe()
	}
	return err
}

func (s *Server) withContext(handler api.HandlerFunc) rest.HandlerFunc {
	return api.WithContext(s.store, s.client, handler)
}

// TODO: Fix all this

//checkPermisson will check that we have the expected permisson
// func checkPermisson(r *rest.Request, expected client.Permission) bool {
// 	//TODO: fill in the details
// 	if permissions := r.Env[client.PERMISSIONS].(*client.UsersPermissions); permissions != nil {
// 		return true
// 	}
// 	return false
// }

// //PostDataset will process a posted dataset for the requested user if permissons are sufficient
// func (client *Server) PostDataset(w rest.ResponseWriter, r *rest.Request) {

// 	userid := r.PathParam(useridParamName)

// 	if !checkPermisson(r, client.Permission{}) {
// 		rest.Error(w, missingPermissionsError, http.StatusUnauthorized)
// 		return
// 	}

// 	groupID := r.Env[client.GROUPID]

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
// 		if err = client.store.Save(platformData[i]); err != nil {
// 			rest.Error(w, err.Error(), http.StatusInternalServerError)
// 			return
// 		}
// 	}
// 	w.WriteHeader(http.StatusOK)
// 	return
// }

// //PostBlob will save a posted blob of  data for the requested user if permissons are sufficient
// func (client *Server) PostBlob(w rest.ResponseWriter, r *rest.Request) {

// 	if checkPermisson(r, client.Permission{}) {
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

// 	if !checkPermisson(r, client.Permission{}) {
// 		rest.Error(w, missingPermissionsError, http.StatusUnauthorized)
// 		return
// 	}

// 	groupID := r.Env[client.GROUPID]

// 	if groupID == "" {
// 		rest.Error(w, missingDataError, http.StatusBadRequest)
// 		return
// 	}

// 	var found struct {
// 		types.DatumArray `json:"Dataset"`
// 	}

// 	userid := r.PathParam(useridParamName)
// 	client.logger.WithField(useridParamName, userid).Debug("GetDataset")

// 	iter := client.store.ReadAll(
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

// 	if checkPermisson(r, client.Permission{}) {
// 		var foundDatum struct {
// 			types.Datum `json:"Datum"`
// 			Errors      string `json:"Errors"`
// 		}

// 		userid := r.PathParam(useridParamName)
// 		datumid := r.PathParam("datumid")

// 		client.logger.WithFields(map[string]interface{}{"userid": userid, "datumid": datumid}).Debug("GetData")

// 		foundDatum.Datum = types.Datum{}
// 		foundDatum.Errors = ""

// 		w.WriteJson(&foundDatum)
// 		return
// 	}
// 	rest.Error(w, missingPermissionsError, http.StatusUnauthorized)
// 	return
// }
