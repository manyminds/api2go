package api2go

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/golang/gddo/httputil"
	"github.com/manyminds/api2go/jsonapi"
	"github.com/manyminds/api2go/routing"
)

const (
	codeInvalidQueryFields  = "API2GO_INVALID_FIELD_QUERY_PARAM"
	defaultContentTypHeader = "application/vnd.api+json"
)

var queryFieldsRegex = regexp.MustCompile(`^fields\[(\w+)\]$`)

type response struct {
	Meta   map[string]interface{}
	Data   interface{}
	Status int
}

func (r response) Metadata() map[string]interface{} {
	return r.Meta
}

func (r response) Result() interface{} {
	return r.Data
}

func (r response) StatusCode() int {
	return r.Status
}

type information struct {
	prefix   string
	resolver URLResolver
}

func (i information) GetBaseURL() string {
	return i.resolver.GetBaseURL()
}

func (i information) GetPrefix() string {
	return i.prefix
}

type paginationQueryParams struct {
	number, size, offset, limit string
}

func newPaginationQueryParams(r *http.Request) paginationQueryParams {
	var result paginationQueryParams

	queryParams := r.URL.Query()
	result.number = queryParams.Get("page[number]")
	result.size = queryParams.Get("page[size]")
	result.offset = queryParams.Get("page[offset]")
	result.limit = queryParams.Get("page[limit]")

	return result
}

func (p paginationQueryParams) isValid() bool {
	if p.number == "" && p.size == "" && p.offset == "" && p.limit == "" {
		return false
	}

	if p.number != "" && p.size != "" && p.offset == "" && p.limit == "" {
		return true
	}

	if p.number == "" && p.size == "" && p.offset != "" && p.limit != "" {
		return true
	}

	return false
}

func (p paginationQueryParams) getLinks(r *http.Request, count uint, info information) (result jsonapi.Links, err error) {
	result = jsonapi.Links{}

	params := r.URL.Query()
	prefix := ""
	baseURL := info.GetBaseURL()
	if baseURL != "" {
		prefix = baseURL
	}
	requestURL := fmt.Sprintf("%s%s", prefix, r.URL.Path)

	if p.number != "" {
		// we have number & size params
		var number uint64
		number, err = strconv.ParseUint(p.number, 10, 64)
		if err != nil {
			return
		}

		if p.number != "1" {
			params.Set("page[number]", "1")
			query, _ := url.QueryUnescape(params.Encode())
			result.First = fmt.Sprintf("%s?%s", requestURL, query)

			params.Set("page[number]", strconv.FormatUint(number-1, 10))
			query, _ = url.QueryUnescape(params.Encode())
			result.Previous = fmt.Sprintf("%s?%s", requestURL, query)
		}

		// calculate last page number
		var size uint64
		size, err = strconv.ParseUint(p.size, 10, 64)
		if err != nil {
			return
		}
		totalPages := (uint64(count) / size)
		if (uint64(count) % size) != 0 {
			// there is one more page with some len(items) < size
			totalPages++
		}

		if number != totalPages {
			params.Set("page[number]", strconv.FormatUint(number+1, 10))
			query, _ := url.QueryUnescape(params.Encode())
			result.Next = fmt.Sprintf("%s?%s", requestURL, query)

			params.Set("page[number]", strconv.FormatUint(totalPages, 10))
			query, _ = url.QueryUnescape(params.Encode())
			result.Last = fmt.Sprintf("%s?%s", requestURL, query)
		}
	} else {
		// we have offset & limit params
		var offset, limit uint64
		offset, err = strconv.ParseUint(p.offset, 10, 64)
		if err != nil {
			return
		}
		limit, err = strconv.ParseUint(p.limit, 10, 64)
		if err != nil {
			return
		}

		if p.offset != "0" {
			params.Set("page[offset]", "0")
			query, _ := url.QueryUnescape(params.Encode())
			result.First = fmt.Sprintf("%s?%s", requestURL, query)

			var prevOffset uint64
			if limit > offset {
				prevOffset = 0
			} else {
				prevOffset = offset - limit
			}
			params.Set("page[offset]", strconv.FormatUint(prevOffset, 10))
			query, _ = url.QueryUnescape(params.Encode())
			result.Previous = fmt.Sprintf("%s?%s", requestURL, query)
		}

		// check if there are more entries to be loaded
		if (offset + limit) < uint64(count) {
			params.Set("page[offset]", strconv.FormatUint(offset+limit, 10))
			query, _ := url.QueryUnescape(params.Encode())
			result.Next = fmt.Sprintf("%s?%s", requestURL, query)

			params.Set("page[offset]", strconv.FormatUint(uint64(count)-limit, 10))
			query, _ = url.QueryUnescape(params.Encode())
			result.Last = fmt.Sprintf("%s?%s", requestURL, query)
		}
	}

	return
}

type notAllowedHandler struct {
	marshalers map[string]ContentMarshaler
}

func (n notAllowedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := NewHTTPError(nil, "Method Not Allowed", http.StatusMethodNotAllowed)
	w.WriteHeader(http.StatusMethodNotAllowed)
	handleError(err, w, r, n.marshalers)
}

type resource struct {
	resourceType reflect.Type
	source       CRUD
	name         string
	marshalers   map[string]ContentMarshaler
}

// middlewareChain executes the middleeware chain setup
func (api *API) middlewareChain(c APIContexter, w http.ResponseWriter, r *http.Request) {
	for _, middleware := range api.middlewares {
		middleware(c, w, r)
	}
}

// allocateContext creates a context for the api.contextPool, saving allocations
func (api *API) allocateDefaultContext() APIContexter {
	return &APIContext{}
}

func (api *API) addResource(prototype jsonapi.MarshalIdentifier, source CRUD, marshalers map[string]ContentMarshaler) *resource {
	resourceType := reflect.TypeOf(prototype)
	if resourceType.Kind() != reflect.Struct && resourceType.Kind() != reflect.Ptr {
		panic("pass an empty resource struct or a struct pointer to AddResource!")
	}

	var ptrPrototype interface{}
	var name string

	if resourceType.Kind() == reflect.Struct {
		ptrPrototype = reflect.New(resourceType).Interface()
		name = resourceType.Name()
	} else {
		ptrPrototype = reflect.ValueOf(prototype).Interface()
		name = resourceType.Elem().Name()
	}

	// check if EntityNamer interface is implemented and use that as name
	entityName, ok := prototype.(jsonapi.EntityNamer)
	if ok {
		name = entityName.GetName()
	} else {
		name = jsonapi.Jsonify(jsonapi.Pluralize(name))
	}

	res := resource{
		resourceType: resourceType,
		name:         name,
		source:       source,
		marshalers:   marshalers,
	}

	requestInfo := func(r *http.Request, api *API) *information {
		var info *information
		if resolver, ok := api.info.resolver.(RequestAwareURLResolver); ok {
			resolver.SetRequest(*r)
			info = &information{prefix: api.info.prefix, resolver: resolver}
		} else {
			info = &api.info
		}

		return info
	}

	prefix := strings.Trim(api.info.prefix, "/")
	baseURL := "/" + name
	if prefix != "" {
		baseURL = "/" + prefix + baseURL
	}

	api.router.Handle("OPTIONS", baseURL, func(w http.ResponseWriter, r *http.Request, _ map[string]string) {
		c := api.contextPool.Get().(APIContexter)
		c.Reset()
		api.middlewareChain(c, w, r)
		w.Header().Set("Allow", "GET,POST,PATCH,OPTIONS")
		w.WriteHeader(http.StatusNoContent)
		api.contextPool.Put(c)
	})

	api.router.Handle("OPTIONS", baseURL+"/:id", func(w http.ResponseWriter, r *http.Request, _ map[string]string) {
		c := api.contextPool.Get().(APIContexter)
		c.Reset()
		api.middlewareChain(c, w, r)
		w.Header().Set("Allow", "GET,PATCH,DELETE,OPTIONS")
		w.WriteHeader(http.StatusNoContent)
		api.contextPool.Put(c)
	})

	api.router.Handle("GET", baseURL, func(w http.ResponseWriter, r *http.Request, _ map[string]string) {
		info := requestInfo(r, api)
		c := api.contextPool.Get().(APIContexter)
		c.Reset()
		api.middlewareChain(c, w, r)

		err := res.handleIndex(c, w, r, *info)
		api.contextPool.Put(c)
		if err != nil {
			handleError(err, w, r, marshalers)
		}
	})

	api.router.Handle("GET", baseURL+"/:id", func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		info := requestInfo(r, api)
		c := api.contextPool.Get().(APIContexter)
		c.Reset()
		api.middlewareChain(c, w, r)
		err := res.handleRead(c, w, r, params, *info)
		api.contextPool.Put(c)
		if err != nil {
			handleError(err, w, r, marshalers)
		}
	})

	// generate all routes for linked relations if there are relations
	casted, ok := prototype.(jsonapi.MarshalReferences)
	if ok {
		relations := casted.GetReferences()
		for _, relation := range relations {
			api.router.Handle("GET", baseURL+"/:id/relationships/"+relation.Name, func(relation jsonapi.Reference) routing.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request, params map[string]string) {
					info := requestInfo(r, api)
					c := api.contextPool.Get().(APIContexter)
					c.Reset()
					api.middlewareChain(c, w, r)
					err := res.handleReadRelation(c, w, r, params, *info, relation)
					api.contextPool.Put(c)
					if err != nil {
						handleError(err, w, r, marshalers)
					}
				}
			}(relation))

			api.router.Handle("GET", baseURL+"/:id/"+relation.Name, func(relation jsonapi.Reference) routing.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request, params map[string]string) {
					info := requestInfo(r, api)
					c := api.contextPool.Get().(APIContexter)
					c.Reset()
					api.middlewareChain(c, w, r)
					err := res.handleLinked(c, api, w, r, params, relation, *info)
					api.contextPool.Put(c)
					if err != nil {
						handleError(err, w, r, marshalers)
					}
				}
			}(relation))

			api.router.Handle("PATCH", baseURL+"/:id/relationships/"+relation.Name, func(relation jsonapi.Reference) routing.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request, params map[string]string) {
					c := api.contextPool.Get().(APIContexter)
					c.Reset()
					api.middlewareChain(c, w, r)
					err := res.handleReplaceRelation(c, w, r, params, relation)
					api.contextPool.Put(c)
					if err != nil {
						handleError(err, w, r, marshalers)
					}
				}
			}(relation))

			if _, ok := ptrPrototype.(jsonapi.EditToManyRelations); ok && relation.Name == jsonapi.Pluralize(relation.Name) {
				// generate additional routes to manipulate to-many relationships
				api.router.Handle("POST", baseURL+"/:id/relationships/"+relation.Name, func(relation jsonapi.Reference) routing.HandlerFunc {
					return func(w http.ResponseWriter, r *http.Request, params map[string]string) {
						c := api.contextPool.Get().(APIContexter)
						c.Reset()
						api.middlewareChain(c, w, r)
						err := res.handleAddToManyRelation(c, w, r, params, relation)
						api.contextPool.Put(c)
						if err != nil {
							handleError(err, w, r, marshalers)
						}
					}
				}(relation))

				api.router.Handle("DELETE", baseURL+"/:id/relationships/"+relation.Name, func(relation jsonapi.Reference) routing.HandlerFunc {
					return func(w http.ResponseWriter, r *http.Request, params map[string]string) {
						c := api.contextPool.Get().(APIContexter)
						c.Reset()
						api.middlewareChain(c, w, r)
						err := res.handleDeleteToManyRelation(c, w, r, params, relation)
						api.contextPool.Put(c)
						if err != nil {
							handleError(err, w, r, marshalers)
						}
					}
				}(relation))
			}
		}
	}

	api.router.Handle("POST", baseURL, func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		info := requestInfo(r, api)
		c := api.contextPool.Get().(APIContexter)
		c.Reset()
		api.middlewareChain(c, w, r)
		err := res.handleCreate(c, w, r, info.prefix, *info)
		api.contextPool.Put(c)
		if err != nil {
			handleError(err, w, r, marshalers)
		}
	})

	api.router.Handle("DELETE", baseURL+"/:id", func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		c := api.contextPool.Get().(APIContexter)
		c.Reset()
		api.middlewareChain(c, w, r)
		err := res.handleDelete(c, w, r, params)
		api.contextPool.Put(c)
		if err != nil {
			handleError(err, w, r, marshalers)
		}
	})

	api.router.Handle("PATCH", baseURL+"/:id", func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		c := api.contextPool.Get().(APIContexter)
		c.Reset()
		api.middlewareChain(c, w, r)
		err := res.handleUpdate(c, w, r, params)
		api.contextPool.Put(c)
		if err != nil {
			handleError(err, w, r, marshalers)
		}
	})

	api.resources = append(api.resources, res)

	return &res
}

func buildRequest(c APIContexter, r *http.Request) Request {
	req := Request{PlainRequest: r}
	params := make(map[string][]string)
	for key, values := range r.URL.Query() {
		params[key] = strings.Split(values[0], ",")
	}
	req.QueryParams = params
	req.Header = r.Header
	req.Context = c
	return req
}

func (res *resource) handleIndex(c APIContexter, w http.ResponseWriter, r *http.Request, info information) error {
	pagination := newPaginationQueryParams(r)
	if pagination.isValid() {
		source, ok := res.source.(PaginatedFindAll)
		if !ok {
			return NewHTTPError(nil, "Resource does not implement the PaginatedFindAll interface", http.StatusNotFound)
		}

		count, response, err := source.PaginatedFindAll(buildRequest(c, r))
		if err != nil {
			return err
		}

		paginationLinks, err := pagination.getLinks(r, count, info)
		if err != nil {
			return err
		}

		return respondWithPagination(response, info, http.StatusOK, paginationLinks, w, r, res.marshalers)
	}
	source, ok := res.source.(FindAll)
	if !ok {
		return NewHTTPError(nil, "Resource does not implement the FindAll interface", http.StatusNotFound)
	}

	response, err := source.FindAll(buildRequest(c, r))
	if err != nil {
		return err
	}

	return respondWith(response, info, http.StatusOK, w, r, res.marshalers)
}

func (res *resource) handleRead(c APIContexter, w http.ResponseWriter, r *http.Request, params map[string]string, info information) error {
	id := params["id"]

	response, err := res.source.FindOne(id, buildRequest(c, r))

	if err != nil {
		return err
	}

	return respondWith(response, info, http.StatusOK, w, r, res.marshalers)
}

func (res *resource) handleReadRelation(c APIContexter, w http.ResponseWriter, r *http.Request, params map[string]string, info information, relation jsonapi.Reference) error {
	id := params["id"]

	obj, err := res.source.FindOne(id, buildRequest(c, r))
	if err != nil {
		return err
	}

	document, err := jsonapi.MarshalToStruct(obj.Result(), info)
	if err != nil {
		return err
	}

	rel, ok := document.Data.DataObject.Relationships[relation.Name]
	if !ok {
		return NewHTTPError(nil, fmt.Sprintf("There is no relation with the name %s", relation.Name), http.StatusNotFound)
	}

	meta := obj.Metadata()
	if len(meta) > 0 {
		rel.Meta = meta
	}

	return marshalResponse(rel, w, http.StatusOK, r, res.marshalers)
}

// try to find the referenced resource and call the findAll Method with referencing resource id as param
func (res *resource) handleLinked(c APIContexter, api *API, w http.ResponseWriter, r *http.Request, params map[string]string, linked jsonapi.Reference, info information) error {
	id := params["id"]
	for _, resource := range api.resources {
		if resource.name == linked.Type {
			request := buildRequest(c, r)
			request.QueryParams[res.name+"ID"] = []string{id}
			request.QueryParams[res.name+"Name"] = []string{linked.Name}

			// check for pagination, otherwise normal FindAll
			pagination := newPaginationQueryParams(r)
			if pagination.isValid() {
				source, ok := resource.source.(PaginatedFindAll)
				if !ok {
					return NewHTTPError(nil, "Resource does not implement the PaginatedFindAll interface", http.StatusNotFound)
				}

				var count uint
				count, response, err := source.PaginatedFindAll(request)
				if err != nil {
					return err
				}

				paginationLinks, err := pagination.getLinks(r, count, info)
				if err != nil {
					return err
				}

				return respondWithPagination(response, info, http.StatusOK, paginationLinks, w, r, res.marshalers)
			}

			source, ok := resource.source.(FindAll)
			if !ok {
				return NewHTTPError(nil, "Resource does not implement the FindAll interface", http.StatusNotFound)
			}

			obj, err := source.FindAll(request)
			if err != nil {
				return err
			}
			return respondWith(obj, info, http.StatusOK, w, r, res.marshalers)
		}
	}

	return NewHTTPError(
		errors.New("Not Found"),
		"No resource handler is registered to handle the linked resource "+linked.Name,
		http.StatusNotFound,
	)
}

func (res *resource) handleCreate(c APIContexter, w http.ResponseWriter, r *http.Request, prefix string, info information) error {
	ctx, err := unmarshalRequest(r, res.marshalers)
	if err != nil {
		return err
	}

	// Ok this is weird again, but reflect.New produces a pointer, so we need the pure type without pointer,
	// otherwise we would have a pointer pointer type that we don't want.
	resourceType := res.resourceType
	if resourceType.Kind() == reflect.Ptr {
		resourceType = resourceType.Elem()
	}
	newObj := reflect.New(resourceType).Interface()

	err = jsonapi.Unmarshal(ctx, newObj)
	if err != nil {
		return err
	}

	var response Responder

	if res.resourceType.Kind() == reflect.Struct {
		// we have to dereference the pointer if user wants to use non pointer values
		response, err = res.source.Create(reflect.ValueOf(newObj).Elem().Interface(), buildRequest(c, r))
	} else {
		response, err = res.source.Create(newObj, buildRequest(c, r))
	}
	if err != nil {
		return err
	}

	result, ok := response.Result().(jsonapi.MarshalIdentifier)

	if !ok {
		return fmt.Errorf("Expected one newly created object by resource %s", res.name)
	}

	w.Header().Set("Location", "/"+prefix+"/"+res.name+"/"+result.GetID())

	// handle 200 status codes
	switch response.StatusCode() {
	case http.StatusCreated:
		return respondWith(response, info, http.StatusCreated, w, r, res.marshalers)
	case http.StatusNoContent:
		w.WriteHeader(response.StatusCode())
		return nil
	case http.StatusAccepted:
		w.WriteHeader(response.StatusCode())
		return nil
	default:
		return fmt.Errorf("invalid status code %d from resource %s for method Create", response.StatusCode(), res.name)
	}
}

func (res *resource) handleUpdate(c APIContexter, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	id := params["id"]
	obj, err := res.source.FindOne(id, buildRequest(c, r))
	if err != nil {
		return err
	}

	ctx, err := unmarshalRequest(r, res.marshalers)
	if err != nil {
		return err
	}

	// we have to make the Result to a pointer to unmarshal into it
	updatingObj := reflect.ValueOf(obj.Result())
	if updatingObj.Kind() == reflect.Struct {
		updatingObjPtr := reflect.New(reflect.TypeOf(obj.Result()))
		updatingObjPtr.Elem().Set(updatingObj)
		err = jsonapi.Unmarshal(ctx, updatingObjPtr.Interface())
		updatingObj = updatingObjPtr.Elem()
	} else {
		err = jsonapi.Unmarshal(ctx, updatingObj.Interface())
	}
	if err != nil {
		return NewHTTPError(nil, err.Error(), http.StatusNotAcceptable)
	}

	response, err := res.source.Update(updatingObj.Interface(), buildRequest(c, r))

	if err != nil {
		return err
	}

	switch response.StatusCode() {
	case http.StatusOK:
		updated := response.Result()
		if updated == nil {
			internalResponse, err := res.source.FindOne(id, buildRequest(c, r))
			if err != nil {
				return err
			}
			updated = internalResponse.Result()
			if updated == nil {
				return fmt.Errorf("Expected FindOne to return one object of resource %s", res.name)
			}

			response = internalResponse
		}

		return respondWith(response, information{}, http.StatusOK, w, r, res.marshalers)
	case http.StatusAccepted:
		w.WriteHeader(http.StatusAccepted)
		return nil
	case http.StatusNoContent:
		w.WriteHeader(http.StatusNoContent)
		return nil
	default:
		return fmt.Errorf("invalid status code %d from resource %s for method Update", response.StatusCode(), res.name)
	}
}

func (res *resource) handleReplaceRelation(c APIContexter, w http.ResponseWriter, r *http.Request, params map[string]string, relation jsonapi.Reference) error {
	var (
		err     error
		editObj interface{}
	)

	id := params["id"]

	response, err := res.source.FindOne(id, buildRequest(c, r))
	if err != nil {
		return err
	}

	body, err := unmarshalRequest(r, res.marshalers)
	if err != nil {
		return err
	}

	inc := map[string]interface{}{}
	err = json.Unmarshal(body, &inc)
	if err != nil {
		return err
	}
	data, ok := inc["data"]
	if !ok {
		return errors.New("Invalid object. Need a \"data\" object")
	}

	resType := reflect.TypeOf(response.Result()).Kind()
	if resType == reflect.Struct {
		editObj = getPointerToStruct(response.Result())
	} else {
		editObj = response.Result()
	}

	err = jsonapi.UnmarshalRelationshipsData(editObj, relation.Name, data)
	if err != nil {
		return err
	}

	if resType == reflect.Struct {
		_, err = res.source.Update(reflect.ValueOf(editObj).Elem().Interface(), buildRequest(c, r))
	} else {
		_, err = res.source.Update(editObj, buildRequest(c, r))
	}

	w.WriteHeader(http.StatusNoContent)
	return err
}

func (res *resource) handleAddToManyRelation(c APIContexter, w http.ResponseWriter, r *http.Request, params map[string]string, relation jsonapi.Reference) error {
	var (
		err     error
		editObj interface{}
	)

	id := params["id"]

	response, err := res.source.FindOne(id, buildRequest(c, r))
	if err != nil {
		return err
	}

	body, err := unmarshalRequest(r, res.marshalers)
	if err != nil {
		return err
	}
	inc := map[string]interface{}{}
	err = json.Unmarshal(body, &inc)
	if err != nil {
		return err
	}

	data, ok := inc["data"]
	if !ok {
		return errors.New("Invalid object. Need a \"data\" object")
	}

	newRels, ok := data.([]interface{})
	if !ok {
		return fmt.Errorf("Data must be an array with \"id\" and \"type\" field to add new to-many relationships")
	}

	newIDs := []string{}

	for _, newRel := range newRels {
		casted, ok := newRel.(map[string]interface{})
		if !ok {
			return errors.New("entry in data object invalid")
		}
		newID, ok := casted["id"].(string)
		if !ok {
			return errors.New("no id field found inside data object")
		}

		newIDs = append(newIDs, newID)
	}

	resType := reflect.TypeOf(response.Result()).Kind()
	if resType == reflect.Struct {
		editObj = getPointerToStruct(response.Result())
	} else {
		editObj = response.Result()
	}

	targetObj, ok := editObj.(jsonapi.EditToManyRelations)
	if !ok {
		return errors.New("target struct must implement jsonapi.EditToManyRelations")
	}
	targetObj.AddToManyIDs(relation.Name, newIDs)

	if resType == reflect.Struct {
		_, err = res.source.Update(reflect.ValueOf(targetObj).Elem().Interface(), buildRequest(c, r))
	} else {
		_, err = res.source.Update(targetObj, buildRequest(c, r))
	}

	w.WriteHeader(http.StatusNoContent)

	return err
}

func (res *resource) handleDeleteToManyRelation(c APIContexter, w http.ResponseWriter, r *http.Request, params map[string]string, relation jsonapi.Reference) error {
	var (
		err     error
		editObj interface{}
	)

	id := params["id"]

	response, err := res.source.FindOne(id, buildRequest(c, r))
	if err != nil {
		return err
	}

	body, err := unmarshalRequest(r, res.marshalers)
	if err != nil {
		return err
	}

	inc := map[string]interface{}{}
	err = json.Unmarshal(body, &inc)
	if err != nil {
		return err
	}

	data, ok := inc["data"]
	if !ok {
		return errors.New("Invalid object. Need a \"data\" object")
	}

	newRels, ok := data.([]interface{})
	if !ok {
		return fmt.Errorf("Data must be an array with \"id\" and \"type\" field to add new to-many relationships")
	}

	obsoleteIDs := []string{}

	for _, newRel := range newRels {
		casted, ok := newRel.(map[string]interface{})
		if !ok {
			return errors.New("entry in data object invalid")
		}
		obsoleteID, ok := casted["id"].(string)
		if !ok {
			return errors.New("no id field found inside data object")
		}

		obsoleteIDs = append(obsoleteIDs, obsoleteID)
	}

	resType := reflect.TypeOf(response.Result()).Kind()
	if resType == reflect.Struct {
		editObj = getPointerToStruct(response.Result())
	} else {
		editObj = response.Result()
	}

	targetObj, ok := editObj.(jsonapi.EditToManyRelations)
	if !ok {
		return errors.New("target struct must implement jsonapi.EditToManyRelations")
	}
	targetObj.DeleteToManyIDs(relation.Name, obsoleteIDs)

	if resType == reflect.Struct {
		_, err = res.source.Update(reflect.ValueOf(targetObj).Elem().Interface(), buildRequest(c, r))
	} else {
		_, err = res.source.Update(targetObj, buildRequest(c, r))
	}

	w.WriteHeader(http.StatusNoContent)

	return err
}

// returns a pointer to an interface{} struct
func getPointerToStruct(oldObj interface{}) interface{} {
	resType := reflect.TypeOf(oldObj)
	ptr := reflect.New(resType)
	ptr.Elem().Set(reflect.ValueOf(oldObj))
	return ptr.Interface()
}

func (res *resource) handleDelete(c APIContexter, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	id := params["id"]
	response, err := res.source.Delete(id, buildRequest(c, r))
	if err != nil {
		return err
	}

	switch response.StatusCode() {
	case http.StatusOK:
		data := map[string]interface{}{
			"meta": response.Metadata(),
		}

		return marshalResponse(data, w, http.StatusOK, r, res.marshalers)
	case http.StatusAccepted:
		w.WriteHeader(http.StatusAccepted)
		return nil
	case http.StatusNoContent:
		w.WriteHeader(http.StatusNoContent)
		return nil
	default:
		return fmt.Errorf("invalid status code %d from resource %s for method Delete", response.StatusCode(), res.name)
	}
}

func writeResult(w http.ResponseWriter, data []byte, status int, contentType string) {
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(status)
	w.Write(data)
}

func respondWith(obj Responder, info information, status int, w http.ResponseWriter, r *http.Request, marshalers map[string]ContentMarshaler) error {
	data, err := jsonapi.MarshalToStruct(obj.Result(), info)
	if err != nil {
		return err
	}

	meta := obj.Metadata()
	if len(meta) > 0 {
		data.Meta = meta
	}

	return marshalResponse(data, w, status, r, marshalers)
}

func respondWithPagination(obj Responder, info information, status int, links jsonapi.Links, w http.ResponseWriter, r *http.Request, marshalers map[string]ContentMarshaler) error {
	data, err := jsonapi.MarshalToStruct(obj.Result(), info)
	if err != nil {
		return err
	}

	data.Links = &links
	meta := obj.Metadata()
	if len(meta) > 0 {
		data.Meta = meta
	}

	return marshalResponse(data, w, status, r, marshalers)
}

func unmarshalRequest(r *http.Request, marshalers map[string]ContentMarshaler) ([]byte, error) {
	defer r.Body.Close()
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	// Todo: custom content unmarshaler is broken atm
	//result := map[string]interface{}{}
	//marshaler, _ := selectContentMarshaler(r, marshalers)
	//err = marshaler.Unmarshal(data, &result)
	//if err != nil {
	//return nil, err
	//}
	return data, nil
}

func marshalResponse(resp interface{}, w http.ResponseWriter, status int, r *http.Request, marshalers map[string]ContentMarshaler) error {
	marshaler, contentType := selectContentMarshaler(r, marshalers)
	filtered, err := filterSparseFields(resp, r)
	if err != nil {
		return err
	}
	result, err := marshaler.Marshal(filtered)
	if err != nil {
		return err
	}
	writeResult(w, result, status, contentType)
	return nil
}

func filterSparseFields(resp interface{}, r *http.Request) (interface{}, error) {
	query := r.URL.Query()
	queryParams := parseQueryFields(&query)
	if len(queryParams) < 1 {
		return resp, nil
	}

	if document, ok := resp.(jsonapi.Document); ok {
		wrongFields := map[string][]string{}

		// single entry in data
		data := document.Data.DataObject
		if data != nil {
			errors := replaceAttributes(&queryParams, data)
			for t, v := range errors {
				wrongFields[t] = v
			}
		}

		// data can be a slice too
		datas := document.Data.DataArray
		for index, data := range datas {
			errors := replaceAttributes(&queryParams, &data)
			for t, v := range errors {
				wrongFields[t] = v
			}
			datas[index] = data
		}

		// included slice
		for index, include := range document.Included {
			errors := replaceAttributes(&queryParams, &include)
			for t, v := range errors {
				wrongFields[t] = v
			}
			document.Included[index] = include
		}

		if len(wrongFields) > 0 {
			httpError := NewHTTPError(nil, "Some requested fields were invalid", http.StatusBadRequest)
			for k, v := range wrongFields {
				for _, field := range v {
					httpError.Errors = append(httpError.Errors, Error{
						Status: "Bad Request",
						Code:   codeInvalidQueryFields,
						Title:  fmt.Sprintf(`Field "%s" does not exist for type "%s"`, field, k),
						Detail: "Please make sure you do only request existing fields",
						Source: &ErrorSource{
							Parameter: fmt.Sprintf("fields[%s]", k),
						},
					})
				}
			}
			return nil, httpError
		}
	}
	return resp, nil
}

func parseQueryFields(query *url.Values) (result map[string][]string) {
	result = map[string][]string{}
	for name, param := range *query {
		matches := queryFieldsRegex.FindStringSubmatch(name)
		if len(matches) > 1 {
			match := matches[1]
			result[match] = strings.Split(param[0], ",")
		}
	}

	return
}

func filterAttributes(attributes map[string]interface{}, fields []string) (filteredAttributes map[string]interface{}, wrongFields []string) {
	wrongFields = []string{}
	filteredAttributes = map[string]interface{}{}

	for _, field := range fields {
		if attribute, ok := attributes[field]; ok {
			filteredAttributes[field] = attribute
		} else {
			wrongFields = append(wrongFields, field)
		}
	}

	return
}

func replaceAttributes(query *map[string][]string, entry *jsonapi.Data) map[string][]string {
	fieldType := entry.Type
	attributes := map[string]interface{}{}
	_ = json.Unmarshal(entry.Attributes, &attributes)
	fields := (*query)[fieldType]
	if len(fields) > 0 {
		var wrongFields []string
		attributes, wrongFields = filterAttributes(attributes, fields)
		if len(wrongFields) > 0 {
			return map[string][]string{
				fieldType: wrongFields,
			}
		}
		bytes, _ := json.Marshal(attributes)
		entry.Attributes = bytes
	}

	return nil
}

func selectContentMarshaler(r *http.Request, marshalers map[string]ContentMarshaler) (marshaler ContentMarshaler, contentType string) {
	if _, found := r.Header["Accept"]; found {
		var contentTypes []string
		for ct := range marshalers {
			contentTypes = append(contentTypes, ct)
		}

		contentType = httputil.NegotiateContentType(r, contentTypes, defaultContentTypHeader)
		marshaler = marshalers[contentType]
	} else if contentTypes, found := r.Header["Content-Type"]; found {
		contentType = contentTypes[0]
		marshaler = marshalers[contentType]
	}

	if marshaler == nil {
		contentType = defaultContentTypHeader
		marshaler = JSONContentMarshaler{}
	}

	return
}

func handleError(err error, w http.ResponseWriter, r *http.Request, marshalers map[string]ContentMarshaler) {
	marshaler, contentType := selectContentMarshaler(r, marshalers)

	log.Println(err)
	if e, ok := err.(HTTPError); ok {
		writeResult(w, []byte(marshaler.MarshalError(err)), e.status, contentType)
		return

	}

	writeResult(w, []byte(marshaler.MarshalError(err)), http.StatusInternalServerError, contentType)
}
