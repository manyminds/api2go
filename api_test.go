package api2go

import (
	"encoding/json"
	"io/ioutil"
	"log"

	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type Post struct {
	ID    string
	Title string
}

type fixtureSource struct {
	posts map[string]*Post
}

func (s *fixtureSource) FindAll() (interface{}, error) {
	postsSlice := make([]Post, len(s.posts))
	i := 0
	for _, p := range s.posts {
		postsSlice[i] = *p
		i++
	}
	return postsSlice, nil
}

func (s *fixtureSource) FindOne(id string) (interface{}, error) {
	if p, ok := s.posts[id]; ok {
		return *p, nil
	}
	return nil, NewHTTPError(nil, "post not found", http.StatusNotFound)
}

func (s *fixtureSource) Create(obj interface{}) (string, error) {
	p := obj.(Post)
	maxID := 0
	for k := range s.posts {
		id, _ := strconv.Atoi(k)
		if id > maxID {
			maxID = id
		}
	}
	newID := strconv.Itoa(maxID + 1)
	p.ID = newID
	s.posts[newID] = &p
	return newID, nil
}

func (s *fixtureSource) Delete(id string) error {
	delete(s.posts, id)
	return nil
}

func (s *fixtureSource) Update(obj interface{}) error {
	p := obj.(Post)
	if oldP, ok := s.posts[p.ID]; ok {
		oldP.Title = p.Title
		return nil
	}
	return NewHTTPError(nil, "post not found", http.StatusNotFound)
}

type CustomController struct{}

var controllerErrorText = "exciting error"
var controllerError = NewHTTPError(nil, controllerErrorText, http.StatusInternalServerError)

func (ctrl *CustomController) FindAll(r *http.Request, objs *interface{}) error {
	return controllerError
}

func (ctrl *CustomController) FindOne(r *http.Request, obj *interface{}) error {
	return controllerError
}

func (ctrl *CustomController) Create(r *http.Request, obj *interface{}) error {
	return controllerError
}

func (ctrl *CustomController) Delete(r *http.Request, id string) error {
	return controllerError
}

func (ctrl *CustomController) Update(r *http.Request, obj *interface{}) error {
	return controllerError
}

var _ = Describe("RestHandler", func() {
	Context("when handling requests", func() {

		var (
			source    *fixtureSource
			post1Json map[string]interface{}

			api *API
			rec *httptest.ResponseRecorder
		)

		BeforeEach(func() {
			source = &fixtureSource{map[string]*Post{
				"1": &Post{ID: "1", Title: "Hello, World!"},
			}}

			log.SetOutput(ioutil.Discard)

			post1Json = map[string]interface{}{
				"id":    "1",
				"title": "Hello, World!",
			}

			api = NewAPI("")
			api.AddResource(Post{}, source)

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
				"posts": []interface{}{post1Json},
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
				"posts": []interface{}{post1Json},
			}))
		})

		It("404s", func() {
			req, err := http.NewRequest("GET", "/posts/23", nil)
			Expect(err).To(BeNil())
			api.Handler().ServeHTTP(rec, req)
			Expect(rec.Code).To(Equal(http.StatusNotFound))
			Expect(rec.Body.String()).To(Equal("post not found\n"))
		})

		It("POSTSs new objects", func() {
			reqBody := strings.NewReader(`{"posts": [{"title": "New Post"}]}`)
			req, err := http.NewRequest("POST", "/posts", reqBody)
			Expect(err).To(BeNil())
			api.Handler().ServeHTTP(rec, req)
			Expect(rec.Code).To(Equal(http.StatusCreated))
			Expect(rec.Header().Get("Location")).To(Equal("/posts/2"))
			var result map[string]interface{}
			Expect(json.Unmarshal(rec.Body.Bytes(), &result)).To(BeNil())
			Expect(result).To(Equal(map[string]interface{}{
				"posts": []interface{}{
					map[string]interface{}{
						"id":    "2",
						"title": "New Post",
					},
				},
			}))
		})

		It("POSTSs multiple objects", func() {
			reqBody := strings.NewReader(`{"posts": [{"title": "New Post"}, {"title" : "Second New Post"}]}`)
			req, err := http.NewRequest("POST", "/posts", reqBody)
			Expect(err).To(BeNil())
			api.Handler().ServeHTTP(rec, req)
			Expect(rec.Code).To(Equal(http.StatusInternalServerError))
			Expect(rec.Header().Get("Location")).To(Equal(""))
			Expect(rec.Body.Bytes()).To(BeNil())
		})

		It("PUTSs multiple objects", func() {
			reqBody := strings.NewReader(`{"posts": [{"title": "New Post"}, {"title" : "Second New Post"}]}`)
			req, err := http.NewRequest("PUT", "/posts/1", reqBody)
			Expect(err).To(BeNil())
			api.Handler().ServeHTTP(rec, req)
			Expect(rec.Code).To(Equal(http.StatusInternalServerError))
			Expect(rec.Header().Get("Location")).To(Equal(""))
			Expect(rec.Body.Bytes()).To(BeNil())
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
			Expect(len(source.posts)).To(BeZero())
		})

		It("UPDATEs", func() {
			reqBody := strings.NewReader(`{"posts": [{"id": "1", "title": "New Title"}]}`)
			req, err := http.NewRequest("PUT", "/posts/1", reqBody)
			Expect(err).To(BeNil())
			api.Handler().ServeHTTP(rec, req)
			Expect(rec.Code).To(Equal(http.StatusNoContent))
			Expect(source.posts["1"].Title).To(Equal("New Title"))
		})
	})

	Context("When using custom controller", func() {
		var (
			controller CustomController
			source     *fixtureSource

			api *API
			rec *httptest.ResponseRecorder
		)

		BeforeEach(func() {
			source = &fixtureSource{map[string]*Post{
				"1": &Post{ID: "1", Title: "Hello, World!"},
			}}

			log.SetOutput(ioutil.Discard)

			api = NewAPI("")
			controller = CustomController{}
			api.AddResourceWithController(Post{}, source, &controller)

			rec = httptest.NewRecorder()
		})

		Describe("Controller called for", func() {
			It("FindAll", func() {
				req, err := http.NewRequest("GET", "/posts", nil)
				Expect(err).To(BeNil())
				api.Handler().ServeHTTP(rec, req)
				Expect(rec.Code).To(Equal(http.StatusInternalServerError))
				Expect(strings.TrimSpace(rec.Body.String())).To(Equal(controllerErrorText))
			})

			It("FindOne", func() {
				req, err := http.NewRequest("GET", "/posts/1", nil)
				Expect(err).To(BeNil())
				api.Handler().ServeHTTP(rec, req)
				Expect(rec.Code).To(Equal(http.StatusInternalServerError))
				Expect(strings.TrimSpace(rec.Body.String())).To(Equal(controllerErrorText))
			})

			It("Create", func() {
				reqBody := strings.NewReader(`{"posts": [{"title": "New Post"}]}`)
				req, err := http.NewRequest("POST", "/posts", reqBody)
				Expect(err).To(BeNil())
				api.Handler().ServeHTTP(rec, req)
				Expect(rec.Code).To(Equal(http.StatusInternalServerError))
				Expect(strings.TrimSpace(rec.Body.String())).To(Equal(controllerErrorText))
			})

			It("Delete", func() {
				reqBody := strings.NewReader("")
				req, err := http.NewRequest("DELETE", "/posts/1", reqBody)
				Expect(err).To(BeNil())
				api.Handler().ServeHTTP(rec, req)
				Expect(rec.Code).To(Equal(http.StatusInternalServerError))
				Expect(strings.TrimSpace(rec.Body.String())).To(Equal(controllerErrorText))
			})

			It("Update", func() {
				reqBody := strings.NewReader(`{"posts": [{"id": "1", "title": "New Post"}]}`)
				req, err := http.NewRequest("PUT", "/posts/1", reqBody)
				Expect(err).To(BeNil())
				api.Handler().ServeHTTP(rec, req)
				Expect(rec.Code).To(Equal(http.StatusInternalServerError))
				Expect(strings.TrimSpace(rec.Body.String())).To(Equal(controllerErrorText))
			})
		})
	})

	Context("when prefixing routes", func() {
		It("has correct Location when creating", func() {
			api := NewAPI("v1")
			api.AddResource(Post{}, &fixtureSource{map[string]*Post{}})
			rec := httptest.NewRecorder()
			reqBody := strings.NewReader(`{"posts": [{"title": "New Post"}]}`)
			req, err := http.NewRequest("POST", "/v1/posts", reqBody)
			Expect(err).To(BeNil())
			api.Handler().ServeHTTP(rec, req)
			Expect(rec.Code).To(Equal(http.StatusCreated))
			Expect(rec.Header().Get("Location")).To(Equal("/v1/posts/1"))
		})
	})
})
