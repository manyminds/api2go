//go:build gingonic && !gorillamux && !echo
// +build gingonic,!gorillamux,!echo

package routing_test

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"unsafe"

	"github.com/gin-gonic/gin"
	"github.com/manyminds/api2go"
	"github.com/manyminds/api2go/examples/model"
	"github.com/manyminds/api2go/examples/resource"
	"github.com/manyminds/api2go/examples/storage"
	"github.com/manyminds/api2go/routing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("api2go with gingonic router adapter", func() {
	var (
		router       routing.Routeable
		gg           *gin.Engine
		api          *api2go.API
		rec          *httptest.ResponseRecorder
		contextKey   = "userID"
		contextValue *string
		apiContext   api2go.APIContext
		userStorage  *storage.UserStorage
	)

	BeforeSuite(func() {
		gin.SetMode(gin.ReleaseMode)
		gg = gin.Default()
		router = routing.Gin(gg)
		api = api2go.NewAPIWithRouting(
			"api",
			api2go.NewStaticResolver("/"),
			router,
		)

		// Define the ApiContext to allow for access.
		apiContext = api2go.APIContext{}
		api.SetContextAllocator(func(*api2go.API) api2go.APIContexter {
			return &apiContext
		})

		userStorage = storage.NewUserStorage()
		chocStorage := storage.NewChocolateStorage()
		api.AddResource(model.User{}, resource.UserResource{ChocStorage: chocStorage, UserStorage: userStorage})

		gg.Use(func(c *gin.Context) {
			if contextValue != nil {
				c.Set(contextKey, *contextValue)
			}
		})

		api.AddResource(model.Chocolate{}, resource.ChocolateResource{ChocStorage: chocStorage, UserStorage: userStorage})
	})

	BeforeEach(func() {
		log.SetOutput(ioutil.Discard)
		rec = httptest.NewRecorder()
	})

	Context("CRUD Tests", func() {
		It("will create a new user", func() {
			reqBody := strings.NewReader(`{"data": {"attributes": {"user-name": "Sansa Stark"}, "id": "1", "type": "users"}}`)
			req, err := http.NewRequest("POST", "/api/users", reqBody)
			Expect(err).To(BeNil())
			gg.ServeHTTP(rec, req)
			Expect(rec.Code).To(Equal(http.StatusCreated))
		})

		It("will find her", func() {
			expectedUser := `
			{
				"data":
				{
					"attributes":{
						"user-name":"Sansa Stark"
					},
					"id":"1",
					"relationships":{
						"sweets":{
							"data":[],"links":{"related":"/api/users/1/sweets","self":"/api/users/1/relationships/sweets"}
						}
					},"type":"users"
				},
				"meta":
				{
					"author":"The api2go examples crew","license":"wtfpl","license-url":"http://www.wtfpl.net"
				}
			}`

			req, err := http.NewRequest("GET", "/api/users/1", nil)
			Expect(err).To(BeNil())
			gg.ServeHTTP(rec, req)
			Expect(rec.Code).To(Equal(http.StatusOK))
			Expect(string(rec.Body.Bytes())).To(MatchJSON((expectedUser)))
		})

		It("can call handle", func() {
			handler := api.Handler()
			_, ok := handler.(http.Handler)
			Expect(ok).To(Equal(true))
		})

		It("update the username", func() {
			reqBody := strings.NewReader(`{"data": {"id": "1", "attributes": {"user-name": "Alayne"}, "type" : "users"}}`)
			req, err := http.NewRequest("PATCH", "/api/users/1", reqBody)
			Expect(err).To(BeNil())
			gg.ServeHTTP(rec, req)
			Expect(rec.Code).To(Equal(http.StatusNoContent))
		})

		It("will find her once again", func() {
			expectedUser := `
			{
				"data":
				{
					"attributes":{
						"user-name":"Alayne"
					},
					"id":"1",
					"relationships":{
						"sweets":{
							"data":[],"links":{"related":"/api/users/1/sweets","self":"/api/users/1/relationships/sweets"}
						}
					},"type":"users"
				},
				"meta":
				{
					"author":"The api2go examples crew","license":"wtfpl","license-url":"http://www.wtfpl.net"
				}
			}`

			req, err := http.NewRequest("GET", "/api/users/1", nil)
			Expect(err).To(BeNil())
			gg.ServeHTTP(rec, req)
			Expect(rec.Code).To(Equal(http.StatusOK))
			Expect(string(rec.Body.Bytes())).To(MatchJSON((expectedUser)))
		})

		It("will delete her", func() {
			req, err := http.NewRequest("DELETE", "/api/users/1", nil)
			Expect(err).To(BeNil())
			gg.ServeHTTP(rec, req)
			Expect(rec.Code).To(Equal(http.StatusNoContent))
		})

		It("won't find her anymore", func() {
			expected := `{"errors":[{"status":"404","title":"http error (404) User for id 1 not found and 0 more errors, User for id 1 not found"}]}`
			req, err := http.NewRequest("GET", "/api/users/1", nil)
			Expect(err).To(BeNil())
			gg.ServeHTTP(rec, req)
			Expect(rec.Code).To(Equal(http.StatusNotFound))
			Expect(string(rec.Body.Bytes())).To(MatchJSON(expected))
		})
	})

	Context("PaginatedFindAll Test", func() {
		It("will create links data without double slashes", func() {

			userStorage.Insert(model.User{ID: "1", Username: "Bender Bending Rodriguez"})
			userStorage.Insert(model.User{ID: "2", Username: "Calculon"})

			req, err := http.NewRequest("GET", "/api/users?page[offset]=0&page[limit]=1", nil)

			Expect(err).To(BeNil())

			gg.ServeHTTP(rec, req)

			expectedResult := `
			{
				"links": {
					"last": "/api/users?page[limit]=1\u0026page[offset]=1",
					"next": "/api/users?page[limit]=1\u0026page[offset]=1"
				},
				"data": [
					{
						"type": "users",
						"id": "2",
						"attributes": {
							"user-name": "Bender Bending Rodriguez"
						},
						"relationships": {
							"sweets": {
								"links": {
									"related": "/api/users/2/sweets",
									"self": "/api/users/2/relationships/sweets"
								},
								"data": []
							}
						}
					}
				],
				"meta": {
					"author": "The api2go examples crew",
					"license": "wtfpl",
					"license-url": "http://www.wtfpl.net"
				}
			}`

			Expect(string(rec.Body.Bytes())).To(MatchJSON(expectedResult))
		})
	})

	Context("Gin Context Key Copy Tests", func() {
		BeforeEach(func() {
			contextValue = nil
		})

		It("context value is present for chocolate resource", func() {
			tempVal := "1"
			contextValue = &tempVal
			expected := `{"data":[],"meta":{"author": "The api2go examples crew", "license": "wtfpl", "license-url": "http://www.wtfpl.net"}}`
			req, err := http.NewRequest("GET", "/api/chocolates", strings.NewReader(""))
			Expect(err).To(BeNil())
			gg.ServeHTTP(rec, req)
			Expect(rec.Code).To(Equal(http.StatusOK))
			Expect(string(rec.Body.Bytes())).To(MatchJSON(expected))

			rawKeys := reflect.ValueOf(&apiContext).Elem().Field(0)
			keys := reflect.NewAt(rawKeys.Type(), unsafe.Pointer(rawKeys.UnsafeAddr())).Elem().Interface().(map[string]interface{})

			Expect(keys).To(Equal(map[string]interface{}{contextKey: *contextValue}))
		})

		It("context value is not present for chocolate resource", func() {
			expected := `{"data":[],"meta":{"author": "The api2go examples crew", "license": "wtfpl", "license-url": "http://www.wtfpl.net"}}`
			req, err := http.NewRequest("GET", "/api/chocolates", strings.NewReader(""))
			Expect(err).To(BeNil())
			gg.ServeHTTP(rec, req)
			Expect(rec.Code).To(Equal(http.StatusOK))
			Expect(string(rec.Body.Bytes())).To(MatchJSON(expected))

			rawKeys := reflect.ValueOf(&apiContext).Elem().Field(0)
			keys := reflect.NewAt(rawKeys.Type(), unsafe.Pointer(rawKeys.UnsafeAddr())).Elem().Interface().(map[string]interface{})

			Expect(keys).To(BeNil())
		})
	})
})
