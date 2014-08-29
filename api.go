package api2go

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// DataSource provides methods needed for CRUD.
type DataSource interface {
	// FindAll returns all objects
	FindAll() (interface{}, error)

	// FindOne returns an object by its ID
	FindOne(ID string) (interface{}, error)
}

// API is a REST JSONAPI.
type API struct {
	router *httprouter.Router
}

// NewAPI returns an initialized API instance
func NewAPI() *API {
	api := new(API)
	api.router = httprouter.New()
	return api
}

// AddResource registers a data source for the given resource
func (api *API) AddResource(name string, source DataSource) {
	api.router.GET("/"+name, func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		objs, err := source.FindAll()
		if err != nil {
			w.WriteHeader(500)
			return
		}
		json, err := MarshalToJSON(objs)
		if err != nil {
			w.WriteHeader(500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(json)
	})

	api.router.GET("/"+name+"/:id", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		obj, err := source.FindOne(ps.ByName("id"))
		if err != nil {
			w.WriteHeader(500)
			return
		}
		json, err := MarshalToJSON(obj)
		if err != nil {
			w.WriteHeader(500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(json)
	})
}

// Handler returns the http.Handler instance for the API.
func (api *API) Handler() http.Handler {
	return api.router
}
