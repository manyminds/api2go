package api2go

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"strings"

	"github.com/julienschmidt/httprouter"
)

// DataSource provides methods needed for CRUD.
type DataSource interface {
	// FindAll returns all objects
	FindAll(req Request) (interface{}, error)

	// FindOne returns an object by its ID
	FindOne(ID string, req Request) (interface{}, error)

	// FindMultiple returns all objects for the specified IDs
	FindMultiple(IDs []string, req Request) (interface{}, error)

	// Create a new object and return its ID
	Create(interface{}) (string, error)

	// Delete an object
	Delete(id string) error

	// Update an object
	Update(obj interface{}) error
}

// Controller provides more customization of each route.
// You can define a controller for every DataSource if needed
type Controller interface {
	// FindAll gets called after resource was called
	FindAll(r *http.Request, objs *interface{}) error

	// FindOne gets called after resource was called
	FindOne(r *http.Request, obj *interface{}) error

	// Create gets called before resource was called
	Create(r *http.Request, obj *interface{}) error

	// Delete gets called before resource was called
	Delete(r *http.Request, id string) error

	// Update gets called before resource was called
	Update(r *http.Request, obj *interface{}) error
}

// API is a REST JSONAPI.
type API struct {
	router *httprouter.Router
	// Route prefix, including slashes
	prefix    string
	resources []resource
}

// NewAPI returns an initialized API instance
// `prefix` is added in front of all endpoints.
func NewAPI(prefix string) *API {
	// Add initial and trailing slash to prefix
	prefix = strings.Trim(prefix, "/")
	if len(prefix) > 0 {
		prefix = "/" + prefix + "/"
	} else {
		prefix = "/"
	}

	return &API{
		router: httprouter.New(),
		prefix: prefix,
	}
}

// Request holds additional information for FindOne and Find Requests
type Request struct {
	QueryParams map[string][]string
	Header      http.Header
}

type resource struct {
	resourceType reflect.Type
	source       DataSource
	name         string
	controller   Controller
}

func (api *API) addResource(prototype interface{}, source DataSource) *resource {
	resourceType := reflect.TypeOf(prototype)
	if resourceType.Kind() != reflect.Struct {
		panic("pass an empty resource struct to AddResource!")
	}

	name := jsonify(pluralize(resourceType.Name()))
	res := resource{
		resourceType: resourceType,
		name:         name,
		source:       source,
	}

	api.router.Handle("OPTIONS", api.prefix+name, func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		w.Header().Set("Allow", "GET,POST,OPTIONS")
		w.WriteHeader(http.StatusNoContent)
	})

	api.router.Handle("OPTIONS", api.prefix+name+"/:id", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		w.Header().Set("Allow", "GET,PUT,DELETE,OPTIONS")
		w.WriteHeader(http.StatusNoContent)
	})

	api.router.GET(api.prefix+name, func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		err := res.handleIndex(w, r, api.prefix)
		if err != nil {
			handleError(err, w)
		}
	})

	api.router.GET(api.prefix+name+"/:id", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		err := res.handleRead(w, r, ps, api.prefix)
		if err != nil {
			handleError(err, w)
		}
	})

	api.router.GET(api.prefix+name+"/:id/:linked", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		err := res.handleLinked(api, w, r, ps, api.prefix)
		if err != nil {
			handleError(err, w)
		}
	})

	api.router.POST(api.prefix+name, func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		err := res.handleCreate(w, r, api.prefix)
		if err != nil {
			handleError(err, w)
		}
	})

	api.router.DELETE(api.prefix+name+"/:id", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		err := res.handleDelete(w, r, ps)
		if err != nil {
			handleError(err, w)
		}
	})

	api.router.PUT(api.prefix+name+"/:id", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		err := res.handleUpdate(w, r, ps)
		if err != nil {
			handleError(err, w)
		}
	})

	api.resources = append(api.resources, res)

	return &res
}

// AddResource registers a data source for the given resource
// `resource` should by an empty struct instance such as `Post{}`. The same type will be used for constructing new elements.
func (api *API) AddResource(prototype interface{}, source DataSource) {
	api.addResource(prototype, source)
}

// AddResourceWithController does the same as `AddResource` but also couples a custom `Controller`
// Use this controller to implement access control and other things that depend on the request
func (api *API) AddResourceWithController(prototype interface{}, source DataSource, controller Controller) {
	res := api.addResource(prototype, source)
	res.controller = controller
}

func buildRequest(r *http.Request) Request {
	req := Request{}
	params := make(map[string][]string)
	for key, values := range r.URL.Query() {
		params[key] = strings.Split(values[0], ",")
	}
	req.QueryParams = params
	req.Header = r.Header
	return req
}

func (res *resource) handleIndex(w http.ResponseWriter, r *http.Request, prefix string) error {
	objs, err := res.source.FindAll(buildRequest(r))
	if err != nil {
		return err
	}

	if res.controller != nil {
		if err := res.controller.FindAll(r, &objs); err != nil {
			return err
		}
	}
	return respondWith(objs, prefix, http.StatusOK, w)
}

func (res *resource) handleRead(w http.ResponseWriter, r *http.Request, ps httprouter.Params, prefix string) error {
	ids := strings.Split(ps.ByName("id"), ",")

	var (
		obj interface{}
		err error
	)

	if len(ids) == 1 {
		obj, err = res.source.FindOne(ids[0], buildRequest(r))
	} else {
		obj, err = res.source.FindMultiple(ids, buildRequest(r))
	}

	if err != nil {
		return err
	}

	if res.controller != nil {
		if err := res.controller.FindOne(r, &obj); err != nil {
			return err
		}
	}
	return respondWith(obj, prefix, http.StatusOK, w)
}

// try to find the referenced resource and call the findAll Method with referencing resource id as param
func (res *resource) handleLinked(api *API, w http.ResponseWriter, r *http.Request, ps httprouter.Params, prefix string) error {
	id := ps.ByName("id")
	linked := ps.ByName("linked")

	// Iterate over all struct fields and determine the type of linked
	for i := 0; i < res.resourceType.NumField(); i++ {
		field := res.resourceType.Field(i)
		fieldName := jsonify(field.Name)
		kind := field.Type.Kind()
		if (kind == reflect.Ptr || kind == reflect.Slice) && fieldName == linked {
			// Check if there is a resource for this type
			fieldType := pluralize(jsonify(field.Type.Elem().Name()))
			for _, resource := range api.resources {
				if resource.name == fieldType {
					request := Request{
						Header: r.Header,
						QueryParams: map[string][]string{
							res.name + "ID": []string{id},
						},
					}
					obj, err := resource.source.FindAll(request)
					if err != nil {
						return err
					}
					return respondWith(obj, prefix, http.StatusOK, w)
				}
			}
		}
	}

	err := Error{
		Status: string(http.StatusNotFound),
		Title:  "Not Found",
		Detail: "No resource handler is registered to handle the linked resource " + linked,
	}
	return respondWith(err, prefix, http.StatusNotFound, w)
}

func (res *resource) handleCreate(w http.ResponseWriter, r *http.Request, prefix string) error {
	ctx, err := unmarshalJSONRequest(r)
	if err != nil {
		return err
	}
	newObjs := reflect.MakeSlice(reflect.SliceOf(res.resourceType), 0, 0)
	err = unmarshalInto(ctx, res.resourceType, &newObjs)
	if err != nil {
		return err
	}
	if newObjs.Len() != 1 {
		return errors.New("expected one object in POST")
	}

	newObj := newObjs.Index(0).Interface()

	if res.controller != nil {
		if err := res.controller.Create(r, &newObj); err != nil {
			return err
		}
	}

	id, err := res.source.Create(newObj)
	if err != nil {
		return err
	}
	w.Header().Set("Location", prefix+res.name+"/"+id)

	obj, err := res.source.FindOne(id, buildRequest(r))
	if err != nil {
		return err
	}

	return respondWith(obj, prefix, http.StatusCreated, w)
}

func (res *resource) handleUpdate(w http.ResponseWriter, r *http.Request, ps httprouter.Params) error {
	obj, err := res.source.FindOne(ps.ByName("id"), buildRequest(r))
	if err != nil {
		return err
	}
	ctx, err := unmarshalJSONRequest(r)
	if err != nil {
		return err
	}
	updatingObjs := reflect.MakeSlice(reflect.SliceOf(res.resourceType), 1, 1)
	updatingObjs.Index(0).Set(reflect.ValueOf(obj))
	err = unmarshalInto(ctx, res.resourceType, &updatingObjs)
	if err != nil {
		return err
	}
	if updatingObjs.Len() != 1 {
		return errors.New("expected one object in PUT")
	}

	updatingObj := updatingObjs.Index(0).Interface()
	if res.controller != nil {
		if err := res.controller.Update(r, &updatingObj); err != nil {
			return err
		}
	}

	if err := res.source.Update(updatingObj); err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (res *resource) handleDelete(w http.ResponseWriter, r *http.Request, ps httprouter.Params) error {
	id := ps.ByName("id")
	if res.controller != nil {
		if err := res.controller.Delete(r, id); err != nil {
			return err
		}
	}

	err := res.source.Delete(ps.ByName("id"))
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func respondWith(obj interface{}, prefix string, status int, w http.ResponseWriter) error {
	data, err := MarshalToJSONPrefix(obj, prefix)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/vnd.api+json")
	w.WriteHeader(status)
	w.Write(data)
	return nil
}

func unmarshalJSONRequest(r *http.Request) (map[string]interface{}, error) {
	defer r.Body.Close()
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	result := map[string]interface{}{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func handleError(err error, w http.ResponseWriter) {
	log.Println(err)
	if e, ok := err.(HTTPError); ok {
		http.Error(w, marshalError(e), e.status)
		return
	}

	w.WriteHeader(500)
}

// Handler returns the http.Handler instance for the API.
func (api *API) Handler() http.Handler {
	return api.router
}
