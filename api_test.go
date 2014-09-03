package api2go

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type sourceAdapter struct {
	findAll   func() (interface{}, error)
	findOne   func(string) (interface{}, error)
	lNewSlice func() interface{}
	create    func(interface{}) (string, error)
}

func (a *sourceAdapter) FindAll() (interface{}, error)          { return a.findAll() }
func (a *sourceAdapter) FindOne(id string) (interface{}, error) { return a.findOne(id) }
func (a *sourceAdapter) NewSlice() interface{}                  { return a.lNewSlice() }
func (a *sourceAdapter) Create(obj interface{}) (string, error) { return a.create(obj) }

var _ = Describe("RestHandler", func() {
	Context("when handling requests", func() {
		type Post struct {
			ID    int
			Title string
		}

		var (
			post1    Post
			post1Map map[string]interface{}

			api *API
			rec *httptest.ResponseRecorder
		)

		BeforeEach(func() {
			post1 = Post{ID: 1, Title: "Hello, World!"}
			post1Map = map[string]interface{}{
				"id":    "1",
				"title": "Hello, World!",
			}

			adapter := &sourceAdapter{
				findAll: func() (interface{}, error) {
					return []Post{post1}, nil
				},
				findOne: func(id string) (interface{}, error) {
					switch id {
					case "1":
						return post1, nil
					default:
						panic("unknown id " + id)
					}
				},
				lNewSlice: func() interface{} {
					return &[]Post{}
				},
				create: func(obj interface{}) (string, error) {
					p := obj.(Post)
					Expect(p.Title).To(Equal("New Post"))
					return "42", nil
				},
			}

			api = NewAPI()
			api.AddResource("posts", adapter)

			rec = httptest.NewRecorder()
		})

		It("GETs collections", func() {
			req, err := http.NewRequest("GET", "/posts", nil)
			Expect(err).To(BeNil())
			api.Handler().ServeHTTP(rec, req)
			Expect(rec.Code).To(Equal(http.StatusOK))
			var result map[string]interface{}
			Expect(json.Unmarshal(rec.Body.Bytes(), &result)).To(BeNil())
			Expect(result).To(Equal(map[string]interface{}{
				"posts": []interface{}{post1Map},
			}))
		})

		It("GETs single objects", func() {
			req, err := http.NewRequest("GET", "/posts/1", nil)
			Expect(err).To(BeNil())
			api.Handler().ServeHTTP(rec, req)
			Expect(rec.Code).To(Equal(http.StatusOK))
			var result map[string]interface{}
			Expect(json.Unmarshal(rec.Body.Bytes(), &result)).To(BeNil())
			Expect(result).To(Equal(map[string]interface{}{
				"posts": []interface{}{post1Map},
			}))
		})

		It("POSTSs new objects", func() {
			reqBody := strings.NewReader(`{"posts": [{"title": "New Post"}]}`)
			req, err := http.NewRequest("POST", "/posts", reqBody)
			Expect(err).To(BeNil())
			api.Handler().ServeHTTP(rec, req)
			Expect(rec.Code).To(Equal(http.StatusCreated))
			Expect(rec.Header().Get("Location")).To(Equal("/posts/42"))
		})
	})
})
