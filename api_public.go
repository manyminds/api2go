package api2go

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/manyminds/api2go/jsonapi"
)

// API is a REST JSONAPI.
type API struct {
	router *httprouter.Router
	// Route prefix, including slashes
	prefix     string
	info       information
	resources  []resource
	marshalers map[string]ContentMarshaler
}

// Handler returns the http.Handler instance for the API.
func (api *API) Handler() http.Handler {
	return api.router
}

// Router can be used instead of Handler() to get the instance of julienschmidt httprouter.
func (api *API) Router() *httprouter.Router {
	return api.router
}

// AddResource registers a data source for the given resource
// At least the CRUD interface must be implemented, all the other interfaces are optional.
// `resource` should be either an empty struct instance such as `Post{}` or a pointer to
// a struct such as `&Post{}`. The same type will be used for constructing new elements.
func (api *API) AddResource(prototype jsonapi.MarshalIdentifier, source CRUD) {
	api.addResource(prototype, source, api.marshalers)
}

// Request contains additional information for FindOne and Find Requests
type Request struct {
	PlainRequest *http.Request
	QueryParams  map[string][]string
	Header       http.Header
}

//SetRedirectTrailingSlash enables 307 redirects on urls ending with /
//when disabled, an URL ending with / will 404
func (api *API) SetRedirectTrailingSlash(enabled bool) {
	if api.router == nil {
		panic("router must not be nil")
	}

	api.router.RedirectTrailingSlash = enabled
}

// NewAPIWithMarshalers does the same as NewAPIWithBaseURL with the addition
// of a set of marshalers that provide a way to interact with clients that
// use a serialization format other than JSON. The marshalers map is indexed
// by the MIME content type to use for a given request-response pair. If the
// client provides an Accept header the server will respond using the client's
// preferred content type, otherwise it will respond using whatever content
// type the client provided in its Content-Type request header.
func NewAPIWithMarshalers(prefix string, baseURL string, marshalers map[string]ContentMarshaler) *API {
	if len(marshalers) == 0 {
		panic("marshaler map must not be empty")
	}

	// Add initial and trailing slash to prefix
	prefixSlashes := strings.Trim(prefix, "/")
	if len(prefixSlashes) > 0 {
		prefixSlashes = "/" + prefixSlashes + "/"
	} else {
		prefixSlashes = "/"
	}

	router := httprouter.New()
	router.MethodNotAllowed = notAllowedHandler{marshalers: marshalers}

	info := information{prefix: prefix, baseURL: baseURL}

	return &API{
		router:     router,
		prefix:     prefixSlashes,
		info:       info,
		marshalers: marshalers,
	}
}

// NewAPI returns an initialized API instance
// `prefix` is added in front of all endpoints.
func NewAPI(prefix string) *API {
	return NewAPIWithMarshalers(prefix, "", DefaultContentMarshalers)
}

// NewAPIWithBaseURL does the same as NewAPI with the addition of
// a baseURL which get's added in front of all generated URLs.
// For example http://localhost/v1/myResource/abc instead of /v1/myResource/abc
func NewAPIWithBaseURL(prefix string, baseURL string) *API {
	return NewAPIWithMarshalers(prefix, baseURL, DefaultContentMarshalers)
}

// DefaultContentMarshalers is the default set of content marshalers for an API.
// Currently this means handling application/vnd.api+json content type bodies
// using the standard encoding/json package.
var DefaultContentMarshalers = map[string]ContentMarshaler{
	defaultContentTypHeader: JSONContentMarshaler{},
}

// JSONContentMarshaler uses the standard encoding/json package for
// decoding requests and encoding responses in JSON format.
type JSONContentMarshaler struct {
}

// Marshal marshals with default JSON
func (m JSONContentMarshaler) Marshal(i interface{}) ([]byte, error) {
	return json.Marshal(i)
}

// Unmarshal with default JSON
func (m JSONContentMarshaler) Unmarshal(data []byte, i interface{}) error {
	return json.Unmarshal(data, i)
}
