package api2go

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"

	"github.com/julienschmidt/httprouter"
)

// DataSource provides methods needed for CRUD.
type DataSource interface {
	// FindAll returns all objects
	FindAll() (interface{}, error)

	// FindOne returns an object by its ID
	FindOne(ID string) (interface{}, error)

	// Create a new object and return its ID
	Create(interface{}) (string, error)

	// Delete an object
	Delete(id string) error
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
// `resource` should by an empty struct instance such as `Post{}`. The same type will be used for constructing new elements.
func (api *API) AddResource(resource interface{}, source DataSource) {
	resourceType := reflect.TypeOf(resource)
	if resourceType.Kind() != reflect.Struct {
		panic("pass an empty resource struct to AddResource!")
	}
	name := jsonify(pluralize(resourceType.Name()))

	api.router.GET("/"+name, func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		objs, err := source.FindAll()
		if err != nil {
			w.WriteHeader(500)
			log.Println(err)
			return
		}
		json, err := MarshalToJSON(objs)
		if err != nil {
			log.Println(err)
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
			log.Println(err)
			return
		}
		json, err := MarshalToJSON(obj)
		if err != nil {
			w.WriteHeader(500)
			log.Println(err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(json)
	})

	api.router.Handle("OPTIONS", "/"+name, func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		w.Header().Set("Allow", "GET,PUT,DELETE,OPTIONS")
		w.WriteHeader(200)
	})

	api.router.POST("/"+name, func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		defer r.Body.Close()
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(500)
			log.Println(err)
			return
		}

		var ctx unmarshalContext
		err = json.Unmarshal(data, &ctx)
		if err != nil {
			w.WriteHeader(500)
			log.Println(err)
			return
		}

		newObjs, err := unmarshalValue(ctx, resourceType)
		if err != nil {
			w.WriteHeader(500)
			log.Println(err)
			return
		}

		if newObjs.Len() != 1 {
			panic("expected one object in POST")
		}
		id, err := source.Create(newObjs.Index(0).Interface())
		if err != nil {
			w.WriteHeader(500)
			log.Println(err)
			return
		}
		w.Header().Set("Location", "/"+name+"/"+id)

		obj, err := source.FindOne(id)
		if err != nil {
			w.WriteHeader(500)
			log.Println(err)
			return
		}
		data, err = MarshalToJSON(obj)
		if err != nil {
			w.WriteHeader(500)
			log.Println(err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(data)
	})

	api.router.DELETE("/"+name+"/:id", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		err := source.Delete(ps.ByName("id"))
		if err != nil {
			w.WriteHeader(500)
			log.Println(err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}

// Handler returns the http.Handler instance for the API.
func (api *API) Handler() http.Handler {
	return api.router
}
