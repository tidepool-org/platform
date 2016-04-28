package dataservices

import "github.com/ant0ine/go-json-rest/rest"

type Status struct {
	DataStore interface{}
	Server    interface{}
}

func (s *Server) GetStatus(response rest.ResponseWriter, request *rest.Request) {
	status := &Status{
		DataStore: s.dataStore.GetStatus(),
		Server:    s.statusMiddleware.GetStatus(),
	}
	response.WriteJson(status)
}
