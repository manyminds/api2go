package api2go

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// IndexFunc should return a slice of all existing objects
type IndexFunc func() interface{}

// GetFunc should return the object matching with the given id
type GetFunc func(id string) interface{}

// HandlerForResource returns a http.Handler for the given resource
func HandlerForResource(name string, indexFunc IndexFunc, getFunc GetFunc) http.Handler {
	router := httprouter.New()

	router.GET("/"+name, func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		objs := indexFunc()
		json, err := MarshalToJSON(objs)
		if err != nil {
			w.WriteHeader(500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(json)
	})

	router.GET("/"+name+"/:id", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		obj := getFunc(ps.ByName("id"))
		json, err := MarshalToJSON(obj)
		if err != nil {
			w.WriteHeader(500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(json)
	})

	return router
}
