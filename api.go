package api2go

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"strconv"

	"github.com/golang/gddo/httputil"
	"github.com/julienschmidt/httprouter"
	"github.com/manyminds/api2go/jsonapi"
)

const defaultContentTypHeader = "application/vnd.api+json"

type response struct {
	Meta   map[string]interface{}
	Data   interface{}
	RawData []byte
	Error error
	Status int
	Header http.Header
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
	prefix  string
	baseURL string
}

func (i information) GetBaseURL() string {
	return i.baseURL
}

func (i information) GetPrefix() string {
	return i.prefix
}

type paginationQueryParams struct {
	number, size, offset, limit string
}

func newPaginationQueryParams(r Request) paginationQueryParams {
	var result paginationQueryParams

	params := r.QueryParams

	result.number = params.Get("page[number]")
	result.size = params.Get("page[size]")
	result.offset = params.Get("page[offset]")
	result.limit = params.Get("page[limit]")

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

func (p paginationQueryParams) getLinks(r Request, count uint, info information) (result map[string]string, err error) {
	result = make(map[string]string)

	params := r.QueryParams
	prefix := ""

	baseURL := info.GetBaseURL()
	if baseURL != "" {
		prefix = baseURL
	}
	requestURL := fmt.Sprintf("%s%s", prefix, r.PlainRequest.URL.Path)

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
			result["first"] = fmt.Sprintf("%s?%s", requestURL, query)

			params.Set("page[number]", strconv.FormatUint(number-1, 10))
			query, _ = url.QueryUnescape(params.Encode())
			result["prev"] = fmt.Sprintf("%s?%s", requestURL, query)
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
			result["next"] = fmt.Sprintf("%s?%s", requestURL, query)

			params.Set("page[number]", strconv.FormatUint(totalPages, 10))
			query, _ = url.QueryUnescape(params.Encode())
			result["last"] = fmt.Sprintf("%s?%s", requestURL, query)
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
			result["first"] = fmt.Sprintf("%s?%s", requestURL, query)

			var prevOffset uint64
			if limit > offset {
				prevOffset = 0
			} else {
				prevOffset = offset - limit
			}
			params.Set("page[offset]", strconv.FormatUint(prevOffset, 10))
			query, _ = url.QueryUnescape(params.Encode())
			result["prev"] = fmt.Sprintf("%s?%s", requestURL, query)
		}

		// check if there are more entries to be loaded
		if (offset + limit) < uint64(count) {
			params.Set("page[offset]", strconv.FormatUint(offset+limit, 10))
			query, _ := url.QueryUnescape(params.Encode())
			result["next"] = fmt.Sprintf("%s?%s", requestURL, query)

			params.Set("page[offset]", strconv.FormatUint(uint64(count)-limit, 10))
			query, _ = url.QueryUnescape(params.Encode())
			result["last"] = fmt.Sprintf("%s?%s", requestURL, query)
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
	resourceType   reflect.Type
	source         CRUD
	hooks          CRUDHooks
	name           string
	marshalers     map[string]ContentMarshaler

	api            *API
}

func (res resource) buildRequest(r *http.Request, params httprouter.Params) (*Request, error) {
	req := Request{
		PlainRequest: r,
		Header: r.Header,
		Params: params,
		QueryParams: r.URL.Query(),
		Meta: make(map[string]interface{}),
		Data: nil,
	}

	var body []byte

	if r.Body != nil {
		var err error
		body, err = ioutil.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}
		r.Body.Close()
	}

	if string(body) == "" {
		req.Data = make(map[string]interface{})
	} else {
		data, err := unmarshalRequest(body, r, res.marshalers)
		if err != nil {
			return nil, errors.New("unmarshal_error: " + err.Error())
		}
		req.Data = data
	}

	if meta, ok := req.Data["meta"]; ok {
		req.Meta = meta.(map[string]interface{})
	}

	return &req, nil
}

func (res *resource) getMiddleware(handler func(Request) response) func(http.ResponseWriter, *http.Request, httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {

		var output []byte
		var status int
		contentType := "application/vnd.api+json"
		header := w.Header()

		request, err := res.buildRequest(r, params)

		if err != nil {
			output = []byte(fmt.Sprintf("{\"errors\": [{\"title\": \"%v\"}]}", err))
			status = 500
		} else {
			// Todo: implement aborting + changing status code in hooks.
			if res.hooks != nil {
				res.hooks.BeforeHandle(request)
			}

			resp := handler(*request)
			if res.hooks != nil {
				res.hooks.AfterHandle(request, resp)
			}

			status = resp.Status

			output = resp.RawData
			
			var err error

			if resp.Error != nil {
				log.Printf("api error: %v\n", resp.Error)
				output, contentType, status = res.marshalError(resp.Error, r)
			} else if resp.Data != nil {
				output, contentType, err = res.marshalData(resp.Data, r)
				if err != nil {
					status = 500
					output = []byte(fmt.Sprintf("{\"errors\": [{\"title\": \"%v\"}]}", err))
				}
			}

			if resp.Header != nil {
				for key := range resp.Header {
					header[key] = resp.Header[key]
				}
			}
		}

		header.Set("Content-Type", contentType)
		w.WriteHeader(status)
		w.Write(output)	
	}
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

	// Check if CRUDHooks is implemented.
	hooks, _ := source.(CRUDHooks) // Nil when not implemented.

	res := resource{
		resourceType: resourceType,
		name:         name,
		source:       source,
		hooks:		    hooks,		
		marshalers:   marshalers,
		api:          api,
	}

	api.router.Handle("OPTIONS", api.prefix+name, res.getMiddleware(func(r Request) response {
		return response{
			Status: http.StatusNoContent,
			Header: http.Header{"Allow": []string{"GET,POST,PATCH,OPTIONS"}},
		}
	}))

	api.router.Handle("OPTIONS", api.prefix+name+"/:id", res.getMiddleware(func(r Request) response {
		return response{
			Status: http.StatusNoContent,
			Header: http.Header{"Allow": []string{"GET,PATCH,DELETE,OPTIONS"}},
		}
	}))

	api.router.GET(api.prefix+name, res.getMiddleware(func(r Request) response {
		return res.handleIndex(r)
	}))

	api.router.GET(api.prefix+name+"/:id", res.getMiddleware(func(r Request) response {
		return res.handleRead(r)
	}))

	// generate all routes for linked relations if there are relations
	casted, ok := prototype.(jsonapi.MarshalReferences)
	if ok {
		relations := casted.GetReferences()
		for _, relation := range relations {
			api.router.GET(api.prefix+name+"/:id/relationships/"+relation.Name, func(relation jsonapi.Reference) httprouter.Handle {
				return res.getMiddleware(func(r Request) response {
					return res.handleReadRelation(r, relation)
				})			
			}(relation))

			api.router.GET(api.prefix+name+"/:id/"+relation.Name, func(relation jsonapi.Reference) httprouter.Handle {
				return res.getMiddleware(func(r Request) response {
					return res.handleLinked(r, relation)
				})
			}(relation))

			api.router.PATCH(api.prefix+name+"/:id/relationships/"+relation.Name, func(relation jsonapi.Reference) httprouter.Handle {
				return res.getMiddleware(func(r Request) response {
					return res.handleReplaceRelation(r, relation)
				})
			}(relation))

			if _, ok := ptrPrototype.(jsonapi.EditToManyRelations); ok && relation.Name == jsonapi.Pluralize(relation.Name) {
				// generate additional routes to manipulate to-many relationships
				api.router.POST(api.prefix+name+"/:id/relationships/"+relation.Name, func(relation jsonapi.Reference) httprouter.Handle {
					return res.getMiddleware(func(r Request) response {
						return res.handleAddToManyRelation(r, relation)
					})
				}(relation))

				api.router.DELETE(api.prefix+name+"/:id/relationships/"+relation.Name, func(relation jsonapi.Reference) httprouter.Handle {
					return res.getMiddleware(func(r Request) response {
						return  res.handleDeleteToManyRelation(r, relation)
					})
				}(relation))
			}
		}
	}

	api.router.POST(api.prefix+name, res.getMiddleware(func(r Request) response {
		return res.handleCreate(r)
	}))

	api.router.DELETE(api.prefix+name+"/:id", res.getMiddleware(func(r Request) response {
		return res.handleDelete(r)
	}))

	api.router.PATCH(api.prefix+name+"/:id", res.getMiddleware(func(r Request) response {
		return res.handleUpdate(r)
	}))

	api.resources = append(api.resources, res)

	return &res
}

func (res *resource) marshalData(data interface{}, r *http.Request) ([]byte, string, error) {
	marshaler, contentType := selectContentMarshaler(r, res.marshalers)
	result, err := marshaler.Marshal(data)
	if err != nil {
		return nil, "", err
	}

	return result, contentType, nil
}

func (res *resource) marshalError(data error, r *http.Request) ([]byte, string, int) {
	marshaler, contentType := selectContentMarshaler(r, res.marshalers)

	status, result := marshaler.MarshalError(data)

	return result, contentType, status
}

func (res *resource) handleIndex(r Request) response {
	pagination := newPaginationQueryParams(r)
	if pagination.isValid() {
		source, ok := res.source.(PaginatedFindAll)
		if !ok {
			return response{
				Status: http.StatusNotFound,
				Error: NewHTTPError(nil, "Resource does not implement the PaginatedFindAll interface", http.StatusNotFound),
			}
		}

		count, resp, err := source.PaginatedFindAll(r)
		if err != nil {
			return response{
				Error: err,
			}
		}

		paginationLinks, err := pagination.getLinks(r, count, res.api.info)
		if err != nil {
			return response{Error: err}
		}

		return res.buildResponse(resp, paginationLinks, http.StatusOK)
	}

	source, ok := res.source.(FindAll)
	if !ok {
		return response{
			Error: NewHTTPError(nil, "Resource does not implement the FindAll interface", http.StatusNotFound),
		}
	}

	resp, err := source.FindAll(r)
	if err != nil {
		return response{Error: err}
	}

	return res.buildResponse(resp, nil, http.StatusOK)
}

func (res *resource) handleRead(r Request) response {
	id := r.Params.ByName("id")

	resp, err := res.source.FindOne(id, r)

	if err != nil {
		return response{Error: err}
	}

	return res.buildResponse(resp, nil, http.StatusOK)
}

func (res *resource) handleReadRelation(r Request, relation jsonapi.Reference) response {
	id := r.Params.ByName("id")

	obj, err := res.source.FindOne(id, r)
	if err != nil {
		return response{Error: err}
	}

	internalError := response{
		Error: NewHTTPError(nil, "Internal server error, invalid object structure", http.StatusInternalServerError),
	}

	marshalled, err := jsonapi.MarshalWithURLs(obj.Result(), res.api.info)
	data, ok := marshalled["data"]
	if !ok {
		return internalError
	}
	relationships, ok := data.(map[string]interface{})["relationships"]
	if !ok {
		return internalError
	}

	rel, ok := relationships.(map[string]map[string]interface{})[relation.Name]
	if !ok {
		return response{
			Error: NewHTTPError(nil, fmt.Sprintf("There is no relation with the name %s", relation.Name), http.StatusNotFound),
		}
	}
	links, ok := rel["links"].(map[string]string)
	if !ok {
		return internalError
	}
	self, ok := links["self"]
	if !ok {
		return internalError
	}
	related, ok := links["related"]
	if !ok {
		return internalError
	}
	relationData, ok := rel["data"]
	if !ok {
		return internalError
	}

	result := map[string]interface{}{}
	result["links"] = map[string]interface{}{
		"self":    self,
		"related": related,
	}
	result["data"] = relationData
	meta := obj.Metadata()
	if len(meta) > 0 {
		result["meta"] = meta
	}

	return response{Status: http.StatusOK, Data: result}
}

// try to find the referenced resource and call the findAll Method with referencing resource id as param
func (res *resource) handleLinked(r Request, linked jsonapi.Reference) response {
	id := r.Params.ByName("id")
	for _, resource := range res.api.resources {
		if resource.name == linked.Type {
			request := r
			request.QueryParams[res.name+"ID"] = []string{id}
			request.QueryParams[res.name+"Name"] = []string{linked.Name}

			// check for pagination, otherwise normal FindAll
			pagination := newPaginationQueryParams(r)
			if pagination.isValid() {
				source, ok := resource.source.(PaginatedFindAll)
				if !ok {
					return response{
						Error: NewHTTPError(nil, "Resource does not implement the PaginatedFindAll interface", http.StatusNotFound),
					}
				}

				var count uint
				count, resp, err := source.PaginatedFindAll(request)
				if err != nil {
					return response{Error: err}
				}

				paginationLinks, err := pagination.getLinks(r, count, res.api.info)
				if err != nil {
					return response{Error: err}
				}

				return res.buildResponse(resp, paginationLinks, http.StatusOK)
			}

			source, ok := resource.source.(FindAll)
			if !ok {
				return response{
					Error: NewHTTPError(nil, "Resource does not implement the FindAll interface", http.StatusNotFound),
				}
			}

			obj, err := source.FindAll(request)
			if err != nil {
				return response{Error: err}
			}
			return res.buildResponse(obj, nil, http.StatusOK)
		}
	}

	return response{
		Error: NewHTTPError(
			errors.New("Not Found"), 
			"No resource handler is registered to handle the linked resource " + linked.Name, 
			http.StatusNotFound),
	}
}

func (res *resource) handleCreate(r Request) response {
	ctx := r.Data
	newObjs := reflect.MakeSlice(reflect.SliceOf(res.resourceType), 0, 0)

	structType := res.resourceType
	if structType.Kind() == reflect.Ptr {
		structType = structType.Elem()
	}

	err := jsonapi.UnmarshalInto(ctx, structType, &newObjs)
	if err != nil {
		return response{Error: err}
	}
	if newObjs.Len() != 1 {
		return response{Error: errors.New("expected one object in POST")}
	}

	//TODO create multiple objects not only one.
	newObj := newObjs.Index(0).Interface()

	resp, err := res.source.Create(newObj, r)
	if err != nil {
		return response{Error: err}
	}

	result, ok := resp.Result().(jsonapi.MarshalIdentifier)

	if !ok {
		return response{
			Error:fmt.Errorf("Expected one newly created object by resource %s", res.name),
		}
	}

	// TODO: ADD HEADER!
	//w.Header().Set("Location", prefix+res.name+"/"+result.GetID())
	header := http.Header{
		"Location": []string{res.api.prefix+res.name+"/"+result.GetID()},
	}

	// handle 200 status codes
	switch resp.StatusCode() {
	case http.StatusCreated:
		r := res.buildResponse(response{
			Data: result,
		}, nil, http.StatusCreated)
		r.Header = header
		return r
	case http.StatusNoContent:
		return response{
			Status: resp.StatusCode(),
			Header: header,
		}
	case http.StatusAccepted:
		return response{
			Status: resp.StatusCode(),
			Header: header,
		}
	default:
		err := NewHTTPError(errors.New("invalid_status_code"), 
				fmt.Sprintf("invalid status code %d from resource %s for method Create", resp.StatusCode(), res.name), 500)
		return response{
			Error: err,
		}
	}
}

func (res *resource) handleUpdate(r Request) response {
	obj, err := res.source.FindOne(r.Params.ByName("id"), r)
	if err != nil {
		log.Printf("ERROR!: %v\n", err)
		return response{Error: err}
	}

	ctx := r.Data
	data, ok := ctx["data"]
	if !ok {
		return response{
			Error: NewHTTPError(errors.New("missing_data_key"), "missing mandatory data key.", http.StatusForbidden),
		}
	}

	check, ok := data.(map[string]interface{})
	if !ok {
		return response{
			Error: NewHTTPError(errors.New("invalid_data"), "data must contain an object.", http.StatusForbidden),
		}
	}

	if _, ok := check["id"]; !ok {
		return response{
			Error: NewHTTPError(errors.New("missing_id_key"), "missing mandatory id key.", http.StatusForbidden),
		}
	}
	if _, ok := check["type"]; !ok {
		return response{
			Error: NewHTTPError(errors.New("missing_type_key"), "missing mandatory type key.", http.StatusForbidden),
		}

	}
	updatingObjs := reflect.MakeSlice(reflect.SliceOf(res.resourceType), 1, 1)
	updatingObjs.Index(0).Set(reflect.ValueOf(obj.Result()))

	structType := res.resourceType
	if structType.Kind() == reflect.Ptr {
		structType = structType.Elem()
	}
	err = jsonapi.UnmarshalInto(ctx, structType, &updatingObjs)
	if err != nil {
		log.Printf("unmarshal failed: %v\n", err)
		return response{Error: err}
	}
	if updatingObjs.Len() != 1 {
		return response{Error: errors.New("expected one object")}
	}
	updatingObj := updatingObjs.Index(0).Interface()

	resp, err := res.source.Update(updatingObj, r)

	if err != nil {
		return response{Error: err}
	}
	switch resp.StatusCode() {
	case http.StatusOK:
		updated := resp.Result()
		if updated == nil {
			internalResponse, err := res.source.FindOne(r.Params.ByName("id"), r)
			if err != nil {
				return response{Error: err}
			}
			updated = internalResponse.Result()
			if updated == nil {
				return response{
					Error: fmt.Errorf("Expected FindOne to return one object of resource %s", res.name),
				}
			}

			resp = internalResponse
		}

		return res.buildResponse(response{Data: updated}, nil, http.StatusOK)
	case http.StatusAccepted:
		return response{
			Status: http.StatusAccepted,
		}
	case http.StatusNoContent:
		return response{
			Status: http.StatusNoContent,
		}
	default:
		err := NewHTTPError(errors.New("invalid_status_code"), 
			fmt.Sprintf("invalid status code %d from resource %s for method Update", resp.StatusCode(), res.name), 
			http.StatusInternalServerError)

		return response{
			Error: err,
		}
	}
}

func (res *resource) handleReplaceRelation(r Request, relation jsonapi.Reference) response {
	var (
		err     error
		editObj interface{}
	)

	resp, err := res.source.FindOne(r.Params.ByName("id"), r)
	if err != nil {
		return response{Error: err}
	}

	inc := r.Data
	data, ok := inc["data"]
	if !ok {
		return response{
			Error: NewHTTPError(nil, "Invalid object. Need a \"data\" object", http.StatusInternalServerError),
		}
	}

	resType := reflect.TypeOf(resp.Result()).Kind()
	if resType == reflect.Struct {
		editObj = getPointerToStruct(resp.Result())
	} else {
		editObj = resp.Result()
	}

	err = jsonapi.UnmarshalRelationshipsData(editObj, relation.Name, data)
	if err != nil {
		return response{Error: err}
	}

	if resType == reflect.Struct {
		_, err = res.source.Update(reflect.ValueOf(editObj).Elem().Interface(), r)
	} else {
		_, err = res.source.Update(editObj, r)
	}

	if err != nil {
		return response{Error: err}
	}

	return response{Status: http.StatusNoContent}
}

func (res *resource) handleAddToManyRelation(r Request, relation jsonapi.Reference) response {
	var (
		err     error
		editObj interface{}
	)

	resp, err := res.source.FindOne(r.Params.ByName("id"), r)
	if err != nil {
		return response{Error: err}
	}

	inc := r.Data
	data, ok := inc["data"]
	if !ok {
		return response{
			Error: NewHTTPError(nil, "Invalid object. Need a \"data\" object", http.StatusInternalServerError),
		}
	}

	newRels, ok := data.([]interface{})
	if !ok {
		return response{
			Error: NewHTTPError(nil, 
				"Data must be an array with \"id\" and \"type\" field to add new to-many relationships", 
				http.StatusInternalServerError),
		}
	}

	newIDs := []string{}

	for _, newRel := range newRels {
		casted, ok := newRel.(map[string]interface{})
		if !ok {
			return response{
				Error: NewHTTPError(nil, 
					"entry in data object invalid", 
					http.StatusInternalServerError),
			}
		}
		newID, ok := casted["id"].(string)
		if !ok {
			return response{
				Error: NewHTTPError(nil, 
					"no id field found inside data object", 
					http.StatusInternalServerError),
			}
		}

		newIDs = append(newIDs, newID)
	}

	resType := reflect.TypeOf(resp.Result()).Kind()
	if resType == reflect.Struct {
		editObj = getPointerToStruct(resp.Result())
	} else {
		editObj = resp.Result()
	}

	targetObj, ok := editObj.(jsonapi.EditToManyRelations)
	if !ok {
		return response{
			Error: NewHTTPError(nil, 
				"target struct must implement jsonapi.EditToManyRelations", 
				http.StatusInternalServerError),
		}
	}
	targetObj.AddToManyIDs(relation.Name, newIDs)

	if resType == reflect.Struct {
		_, err = res.source.Update(reflect.ValueOf(targetObj).Elem().Interface(), r)
	} else {
		_, err = res.source.Update(targetObj, r)
	}

	if err != nil {
		return response{Error: err}
	}
	
	return response{Status: http.StatusNoContent}
}

func (res *resource) handleDeleteToManyRelation(r Request, relation jsonapi.Reference) response {
	var (
		err     error
		editObj interface{}
	)
	resp, err := res.source.FindOne(r.Params.ByName("id"), r)
	if err != nil {
		return response{Error: err}
	}

	inc := r.Data
	data, ok := inc["data"]
	if !ok {
		return response{Error: errors.New("Invalid object. Need a \"data\" object")}
	}

	newRels, ok := data.([]interface{})
	if !ok {
		return response{
			Error: errors.New("Data must be an array with \"id\" and \"type\" field to add new to-many relationships"),
		}
	}

	obsoleteIDs := []string{}

	for _, newRel := range newRels {
		casted, ok := newRel.(map[string]interface{})
		if !ok {
			return response{Error: errors.New("entry in data object invalid")}
		}
		obsoleteID, ok := casted["id"].(string)
		if !ok {
			return response{Error: errors.New("no id field found inside data object")}
		}

		obsoleteIDs = append(obsoleteIDs, obsoleteID)
	}

	resType := reflect.TypeOf(resp.Result()).Kind()
	if resType == reflect.Struct {
		editObj = getPointerToStruct(resp.Result())
	} else {
		editObj = resp.Result()
	}

	targetObj, ok := editObj.(jsonapi.EditToManyRelations)
	if !ok {
		return response{Error: errors.New("target struct must implement jsonapi.EditToManyRelations")}
	}
	targetObj.DeleteToManyIDs(relation.Name, obsoleteIDs)

	if resType == reflect.Struct {
		_, err = res.source.Update(reflect.ValueOf(targetObj).Elem().Interface(), r)
	} else {
		_, err = res.source.Update(targetObj, r)
	}

	if err != nil {
		return response{Error: err}
	}
	
	return response{Status: http.StatusNoContent}
}

// returns a pointer to an interface{} struct
func getPointerToStruct(oldObj interface{}) interface{} {
	resType := reflect.TypeOf(oldObj)
	ptr := reflect.New(resType)
	ptr.Elem().Set(reflect.ValueOf(oldObj))
	return ptr.Interface()
}

func (res *resource) handleDelete(r Request) response {
	resp, err := res.source.Delete(r.Params.ByName("id"), r)
	if err != nil {
		return response{Error: err}
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		data := map[string]interface{}{
			"meta": resp.Metadata(),
		}

		return response{
			Status: http.StatusOK,
			Data: data,
		}
	case http.StatusAccepted:
		return response{
			Status: http.StatusAccepted,
		}
	case http.StatusNoContent:
		return response{
			Status: http.StatusNoContent,
		}
	default:
		err := NewHTTPError(errors.New("invalid_status_code"),
			fmt.Sprintf("invalid status code %d from resource %s for method Delete", resp.StatusCode(), res.name),
			http.StatusInternalServerError)
		return response{
			Error: err,
		}
	}
}

func writeResult(w http.ResponseWriter, data []byte, status int, contentType string) {
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(status)
	w.Write(data)
}

func (res *resource) buildResponse(obj Responder, links map[string]string, status int) response {
	data, err := jsonapi.MarshalWithURLs(obj.Result(), res.api.info)
	if err != nil {
		return response{Error: err}
	}
	if links != nil {
		data["links"] = links
	}

	return response{Data: data, Status: status, Meta: obj.Metadata()}
}


func unmarshalRequest(body []byte, r *http.Request, marshalers map[string]ContentMarshaler) (map[string]interface{}, error) {
	result := map[string]interface{}{}
	marshaler, _ := selectContentMarshaler(r, marshalers)
	err := marshaler.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
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

	if _, ok := err.(HTTPError); ok {
		status, result := marshaler.MarshalError(err)
		writeResult(w, result, status, contentType)
		return
	}

	status, result := marshaler.MarshalError(err)
	writeResult(w, result, status, contentType)
}
