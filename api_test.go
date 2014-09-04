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
	findAll func() (interface{}, error)
	findOne func(string) (interface{}, error)
	create  func(interface{}) (string, error)
	delete  func(string) error
	update  func(interface{}) (interface{}, error)
}

func (a *sourceAdapter) FindAll() (interface{}, error)               { return a.findAll() }
func (a *sourceAdapter) FindOne(id string) (interface{}, error)      { return a.findOne(id) }
func (a *sourceAdapter) Create(obj interface{}) (string, error)      { return a.create(obj) }
func (a *sourceAdapter) Delete(id string) error                      { return a.delete(id) }
func (a *sourceAdapter) Update(obj interface{}) (interface{}, error) { return a.update(obj) }

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

			deleted bool
		)

		BeforeEach(func() {
			deleted = false

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
					case "42":
						return Post{ID: 42, Title: "New Post"}, nil
					default:
						panic("unknown id " + id)
					}
				},
				create: func(obj interface{}) (string, error) {
					p := obj.(Post)
					Expect(p.Title).To(Equal("New Post"))
					return "42", nil
				},
				delete: func(id string) error {
					if id != "1" {
						panic("unknown id")
					}
					deleted = true
					return nil
				},
				update: func(obj interface{}) (interface{}, error) {
					p := obj.(Post)
					if p.ID != 1 {
						panic("unknown id")
					}
					post1.Title = p.Title
					return post1, nil
				},
			}

			api = NewAPI()
			api.AddResource(Post{}, adapter)

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
			var result map[string]interface{}
			Expect(json.Unmarshal(rec.Body.Bytes(), &result)).To(BeNil())
			Expect(result).To(Equal(map[string]interface{}{
				"posts": []interface{}{
					map[string]interface{}{
						"id":    "42",
						"title": "New Post",
					},
				},
			}))
		})

		It("OPTIONS on collection route", func() {
			req, err := http.NewRequest("OPTIONS", "/posts", nil)
			api.Handler().ServeHTTP(rec, req)
			Expect(err).To(BeNil())
			Expect(rec.Code).To(Equal(http.StatusNoContent))
			Expect(rec.Header().Get("Allow")).To(Equal("GET,POST,OPTIONS"))
		})

		It("OPTIONS on element route", func() {
			req, err := http.NewRequest("OPTIONS", "/posts/1", nil)
			api.Handler().ServeHTTP(rec, req)
			Expect(err).To(BeNil())
			Expect(rec.Code).To(Equal(http.StatusNoContent))
			Expect(rec.Header().Get("Allow")).To(Equal("GET,PUT,DELETE,OPTIONS"))
		})

		It("DELETEs", func() {
			req, err := http.NewRequest("DELETE", "/posts/1", nil)
			Expect(err).To(BeNil())
			api.Handler().ServeHTTP(rec, req)
			Expect(rec.Code).To(Equal(http.StatusNoContent))
			Expect(deleted).To(BeTrue())
		})

		It("UPDATEs", func() {
			reqBody := strings.NewReader(`{"posts": [{"id": "1", "title": "New Title"}]}`)
			req, err := http.NewRequest("PUT", "/posts/1", reqBody)
			Expect(err).To(BeNil())
			api.Handler().ServeHTTP(rec, req)
			Expect(rec.Code).To(Equal(http.StatusOK))
			var result map[string]interface{}
			Expect(json.Unmarshal(rec.Body.Bytes(), &result)).To(BeNil())
			Expect(result).To(Equal(map[string]interface{}{
				"posts": []interface{}{
					map[string]interface{}{
						"id":    "1",
						"title": "New Title",
					},
				},
			}))
		})
	})
})
