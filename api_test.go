package api2go

import (
	"encoding/json"
	"errors"
	"fmt"

	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/univedo/api2go/jsonapi"
	"gopkg.in/guregu/null.v2"
)

type Post struct {
	ID       string
	Title    string
	Value    null.Float
	Author   *User     `json:"-"`
	Comments []Comment `json:"-"`
}

func (p Post) GetID() string {
	return p.ID
}

func (p *Post) SetID(ID string) error {
	p.ID = ID
	return nil
}

func (p Post) GetReferences() []jsonapi.Reference {
	return []jsonapi.Reference{
		{
			Name: "author",
			Type: "users",
		},
		{
			Name: "comments",
			Type: "comments",
		},
	}
}

func (p Post) GetReferencedIDs() []jsonapi.ReferenceID {
	result := []jsonapi.ReferenceID{}
	if p.Author != nil {
		result = append(result, jsonapi.ReferenceID{ID: p.Author.GetID(), Name: "author", Type: "users"})
	}
	for _, comment := range p.Comments {
		result = append(result, jsonapi.ReferenceID{ID: comment.GetID(), Name: "comments", Type: "comments"})
	}

	return result
}

func (p *Post) SetReferencedIDs(IDs []jsonapi.ReferenceID) error {
	return nil
}

func (p Post) GetReferencedStructs() []jsonapi.MarshalIdentifier {
	result := []jsonapi.MarshalIdentifier{}
	if p.Author != nil {
		result = append(result, *p.Author)
	}
	for key := range p.Comments {
		result = append(result, p.Comments[key])
	}

	return result
}

type Comment struct {
	ID    string
	Value string
}

func (c Comment) GetID() string {
	return c.ID
}

type User struct {
	ID   string
	Name string
}

func (u User) GetID() string {
	return u.ID
}

type fixtureSource struct {
	posts map[string]*Post
}

func (s *fixtureSource) FindAll(req Request) (interface{}, error) {
	var (
		postsSlice []Post
	)

	if limit, ok := req.QueryParams["limit"]; ok {
		if l, err := strconv.ParseInt(limit[0], 10, 64); err == nil {
			postsSlice = make([]Post, l)
			length := len(s.posts)
			for i := 0; i < length; i++ {
				postsSlice[i] = *s.posts[strconv.Itoa(i+1)]
				if i+1 >= int(l) {
					break
				}
			}
		} else {
			fmt.Println("Error casting to int", err)
			return nil, err
		}
	} else {
		postsSlice = make([]Post, len(s.posts))
		length := len(s.posts)
		for i := 0; i < length; i++ {
			postsSlice[i] = *s.posts[strconv.Itoa(i+1)]
		}
	}

	return postsSlice, nil
}

func (s *fixtureSource) FindOne(id string, req Request) (interface{}, error) {
	if p, ok := s.posts[id]; ok {
		return *p, nil
	}
	return nil, NewHTTPError(nil, "post not found", http.StatusNotFound)
}

func (s *fixtureSource) FindMultiple(IDs []string, req Request) (interface{}, error) {
	var posts []Post

	for _, id := range IDs {
		if p, ok := s.posts[id]; ok {
			posts = append(posts, *p)
		}
	}

	if len(posts) > 0 {
		return posts, nil
	}

	return nil, NewHTTPError(nil, "post not found", http.StatusNotFound)
}

func (s *fixtureSource) Create(obj interface{}, req Request) (string, error) {
	p := obj.(Post)

	if p.Title == "" {
		err := NewHTTPError(errors.New("Bad request."), "Bad Request", http.StatusBadRequest)
		err.Errors = append(err.Errors, Error{ID: "SomeErrorID", Path: "Title"})
		return "", err
	}

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

func (s *fixtureSource) Delete(id string, req Request) error {
	delete(s.posts, id)
	return nil
}

func (s *fixtureSource) Update(obj interface{}, req Request) error {
	p := obj.(Post)
	if oldP, ok := s.posts[p.ID]; ok {
		oldP.Title = p.Title
		return nil
	}
	return NewHTTPError(nil, "post not found", http.StatusNotFound)
}

type userSource struct{}

func (s *userSource) FindAll(req Request) (interface{}, error) {
	postsIDs, ok := req.QueryParams["postsID"]
	if ok {
		if postsIDs[0] == "1" {
			return User{ID: "1", Name: "Dieter"}, nil
		}
	}

	return []User{}, errors.New("Did not receive query parameter")
}

func (s *userSource) FindOne(id string, req Request) (interface{}, error) {
	return nil, nil
}

func (s *userSource) FindMultiple(IDs []string, req Request) (interface{}, error) {
	return nil, nil
}

func (s *userSource) Create(obj interface{}, req Request) (string, error) {
	return "", nil
}

func (s *userSource) Delete(id string, req Request) error {
	return nil
}

func (s *userSource) Update(obj interface{}, req Request) error {
	return NewHTTPError(nil, "user not found", http.StatusNotFound)
}

type commentSource struct{}

func (s *commentSource) FindAll(req Request) (interface{}, error) {
	postsIDs, ok := req.QueryParams["postsID"]
	if ok {
		if postsIDs[0] == "1" {
			return []Comment{Comment{
				ID:    "1",
				Value: "This is a stupid post!",
			}}, nil
		}
	}

	return []Comment{}, errors.New("Did not receive query parameter")
}

func (s *commentSource) FindOne(id string, req Request) (interface{}, error) {
	return nil, nil
}

func (s *commentSource) FindMultiple(IDs []string, req Request) (interface{}, error) {
	return nil, nil
}

func (s *commentSource) Create(obj interface{}, req Request) (string, error) {
	return "", nil
}

func (s *commentSource) Delete(id string, req Request) error {
	return nil
}

func (s *commentSource) Update(obj interface{}, req Request) error {
	return NewHTTPError(nil, "comment not found", http.StatusNotFound)
}

var _ = Describe("RestHandler", func() {
	Context("when handling requests", func() {

		var (
			source          *fixtureSource
			post1Json       map[string]interface{}
			post1LinkedJSON []map[string]interface{}
			post2Json       map[string]interface{}
			post3Json       map[string]interface{}

			api *API
			rec *httptest.ResponseRecorder
		)

		BeforeEach(func() {
			source = &fixtureSource{map[string]*Post{
				"1": &Post{
					ID:    "1",
					Title: "Hello, World!",
					Author: &User{
						ID:   "1",
						Name: "Dieter",
					},
					Comments: []Comment{Comment{
						ID:    "1",
						Value: "This is a stupid post!",
					}},
				},
				"2": &Post{ID: "2", Title: "I am NR. 2"},
				"3": &Post{ID: "3", Title: "I am NR. 3"},
			}}

			post1Json = map[string]interface{}{
				"id":    "1",
				"type":  "posts",
				"title": "Hello, World!",
				"value": nil,
				"links": map[string]interface{}{
					"author": map[string]interface{}{
						"id":   "1",
						"type": "users",
						//"resource": "/v1/posts/1/author",
					},
					"comments": map[string]interface{}{
						"ids":  []string{"1"},
						"type": "comments",
						//"resource": "/v1/posts/1/comments",
					},
				},
			}

			post1LinkedJSON = []map[string]interface{}{
				map[string]interface{}{
					"id":   "1",
					"name": "Dieter",
					"type": "users",
				},
				map[string]interface{}{
					"id":    "1",
					"type":  "comments",
					"value": "This is a stupid post!",
				},
			}

			post2Json = map[string]interface{}{
				"id":    "2",
				"type":  "posts",
				"title": "I am NR. 2",
				"value": nil,
				"links": map[string]interface{}{
					"author": map[string]interface{}{
						"type": "users",
						//"resource": "/v1/posts/2/author",
					},
					"comments": map[string]interface{}{
						"type": "comments",
						//"resource": "/v1/posts/2/comments",
					},
				},
			}

			post3Json = map[string]interface{}{
				"id":    "3",
				"type":  "posts",
				"title": "I am NR. 3",
				"value": nil,
				"links": map[string]interface{}{
					"author": map[string]interface{}{
						"type": "users",
						//"resource": "/v1/posts/3/author",
					},
					"comments": map[string]interface{}{
						"type": "comments",
						//"resource": "/v1/posts/3/comments",
					},
				},
			}

			api = NewAPI("v1")
			api.AddResource(Post{}, source)
			api.AddResource(User{}, &userSource{})
			api.AddResource(Comment{}, &commentSource{})

			rec = httptest.NewRecorder()
		})

		It("GETs collections", func() {
			req, err := http.NewRequest("GET", "/v1/posts", nil)
			Expect(err).To(BeNil())
			api.Handler().ServeHTTP(rec, req)
			Expect(rec.Code).To(Equal(http.StatusOK))
			expected, err := json.Marshal(map[string]interface{}{
				"data":   []map[string]interface{}{post1Json, post2Json, post3Json},
				"linked": post1LinkedJSON,
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(rec.Body.Bytes()).To(MatchJSON(expected))
		})

		It("GETs single objects", func() {
			req, err := http.NewRequest("GET", "/v1/posts/1", nil)
			Expect(err).To(BeNil())
			api.Handler().ServeHTTP(rec, req)
			Expect(rec.Code).To(Equal(http.StatusOK))
			expected, err := json.Marshal(map[string]interface{}{
				"data":   post1Json,
				"linked": post1LinkedJSON,
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(rec.Body.Bytes()).To(MatchJSON(expected))
		})

		It("GETs multiple objects", func() {
			req, err := http.NewRequest("GET", "/v1/posts/1,2", nil)
			Expect(err).To(BeNil())
			api.Handler().ServeHTTP(rec, req)
			Expect(rec.Code).To(Equal(http.StatusOK))
			expected, err := json.Marshal(map[string]interface{}{
				"data":   []interface{}{post1Json, post2Json},
				"linked": post1LinkedJSON,
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(rec.Body.Bytes()).To(MatchJSON(expected))
		})

		It("GETs related struct from resource url", func() {
			req, err := http.NewRequest("GET", "/v1/posts/1/author", nil)
			Expect(err).ToNot(HaveOccurred())
			api.Handler().ServeHTTP(rec, req)
			Expect(rec.Code).To(Equal(http.StatusOK))
			Expect(rec.Body.Bytes()).To(MatchJSON(`{"data": {"id": "1", "name": "Dieter", "type": "users"}}`))
		})

		It("GETs related structs from resource url", func() {
			req, err := http.NewRequest("GET", "/v1/posts/1/comments", nil)
			Expect(err).ToNot(HaveOccurred())
			api.Handler().ServeHTTP(rec, req)
			Expect(rec.Code).To(Equal(http.StatusOK))
			Expect(rec.Body.Bytes()).To(MatchJSON(`{"data": [{"id": "1", "value": "This is a stupid post!", "type": "comments"}]}`))
		})

		It("Gets 404 if a related struct was not found", func() {
			req, err := http.NewRequest("GET", "/v1/posts/1/unicorns", nil)
			Expect(err).ToNot(HaveOccurred())
			api.Handler().ServeHTTP(rec, req)
			Expect(rec.Code).To(Equal(http.StatusNotFound))
			Expect(rec.Body.Bytes()).ToNot(BeEmpty())
		})

		It("404s", func() {
			req, err := http.NewRequest("GET", "/v1/posts/23", nil)
			Expect(err).To(BeNil())
			api.Handler().ServeHTTP(rec, req)
			Expect(rec.Code).To(Equal(http.StatusNotFound))
			errorJSON := []byte(`{"errors":[{"status":"404","title":"post not found"}]}`)
			Expect(rec.Body.Bytes()).To(MatchJSON(errorJSON))
		})

		It("POSTSs new objects", func() {
			reqBody := strings.NewReader(`{"data": [{"title": "New Post", "type": "posts"}]}`)
			req, err := http.NewRequest("POST", "/v1/posts", reqBody)
			Expect(err).To(BeNil())
			api.Handler().ServeHTTP(rec, req)
			Expect(rec.Code).To(Equal(http.StatusCreated))
			Expect(rec.Header().Get("Location")).To(Equal("/v1/posts/4"))
			var result map[string]interface{}
			Expect(json.Unmarshal(rec.Body.Bytes(), &result)).To(BeNil())
			Expect(result).To(Equal(map[string]interface{}{
				"data": map[string]interface{}{
					"id":    "4",
					"type":  "posts",
					"title": "New Post",
					"value": nil,
					"links": map[string]interface{}{
						"author": map[string]interface{}{
							"type": "users",
							//"resource": "/v1/posts/4/author",
						},
						"comments": map[string]interface{}{
							"type": "comments",
							//"resource": "/v1/posts/4/comments",
						},
					},
				},
			}))
		})

		It("POSTSs new objects with trailing slash automatic redirect enabled", func() {
			reqBody := strings.NewReader(`{"data": [{"title": "New Post", "type": "posts"}]}`)
			req, err := http.NewRequest("POST", "/v1/posts/", reqBody)
			Expect(err).To(BeNil())
			api.SetRedirectTrailingSlash(true)
			api.Handler().ServeHTTP(rec, req)
			Expect(rec.Code).To(Equal(http.StatusTemporaryRedirect))
		})

		It("POSTSs with client id", func() {
			reqBody := strings.NewReader(`{"data": [{"id" : "100", "title": "New Post", "type": "posts"}]}`)
			req, err := http.NewRequest("POST", "/v1/posts", reqBody)
			Expect(err).To(BeNil())
			api.Handler().ServeHTTP(rec, req)
			Expect(rec.Code).To(Equal(http.StatusForbidden))
			Expect(rec.Body).To(ContainSubstring("Client generated IDs are not supported."))
		})

		It("POSTSs new objects with trailing slash automatic redirect disabled", func() {
			reqBody := strings.NewReader(`{"data": [{"title": "New Post", "type": "posts"}]}`)
			req, err := http.NewRequest("POST", "/v1/posts/", reqBody)
			Expect(err).To(BeNil())
			api.SetRedirectTrailingSlash(false)
			api.Handler().ServeHTTP(rec, req)
			Expect(rec.Code).To(Equal(http.StatusNotFound))
		})

		It("POSTSs multiple objects", func() {
			reqBody := strings.NewReader(`{"posts": [{"title": "New Post"}, {"title" : "Second New Post"}]}`)
			req, err := http.NewRequest("POST", "/v1/posts", reqBody)
			Expect(err).To(BeNil())
			api.Handler().ServeHTTP(rec, req)
			Expect(rec.Code).To(Equal(http.StatusInternalServerError))
			Expect(rec.Header().Get("Location")).To(Equal(""))
			Expect(rec.Body.Bytes()).ToNot(HaveLen(0))
		})

		It("PUTSs multiple objects", func() {
			reqBody := strings.NewReader(`{"posts": [{"title": "New Post"}, {"title" : "Second New Post"}]}`)
			req, err := http.NewRequest("PUT", "/v1/posts/1", reqBody)
			Expect(err).To(BeNil())
			api.Handler().ServeHTTP(rec, req)
			Expect(rec.Code).To(Equal(http.StatusInternalServerError))
			Expect(rec.Header().Get("Location")).To(Equal(""))
			Expect(rec.Body.Bytes()).ToNot(HaveLen(0))
		})

		It("OPTIONS on collection route", func() {
			req, err := http.NewRequest("OPTIONS", "/v1/posts", nil)
			api.Handler().ServeHTTP(rec, req)
			Expect(err).To(BeNil())
			Expect(rec.Code).To(Equal(http.StatusNoContent))
			Expect(rec.Header().Get("Allow")).To(Equal("GET,POST,OPTIONS"))
		})

		It("OPTIONS on element route", func() {
			req, err := http.NewRequest("OPTIONS", "/v1/posts/1", nil)
			api.Handler().ServeHTTP(rec, req)
			Expect(err).To(BeNil())
			Expect(rec.Code).To(Equal(http.StatusNoContent))
			Expect(rec.Header().Get("Allow")).To(Equal("GET,PUT,DELETE,OPTIONS"))
		})

		It("DELETEs", func() {
			req, err := http.NewRequest("DELETE", "/v1/posts/1", nil)
			Expect(err).To(BeNil())
			api.Handler().ServeHTTP(rec, req)
			Expect(rec.Code).To(Equal(http.StatusNoContent))
			Expect(len(source.posts)).To(Equal(2))
		})

		It("UPDATEs", func() {
			reqBody := strings.NewReader(`{"data": {"id": "1", "title": "New Title", "type": "posts"}}`)
			req, err := http.NewRequest("PUT", "/v1/posts/1", reqBody)
			Expect(err).To(BeNil())
			api.Handler().ServeHTTP(rec, req)
			Expect(rec.Code).To(Equal(http.StatusNoContent))
			Expect(source.posts["1"].Title).To(Equal("New Title"))
		})

		It("UPDATEs as array", func() {
			reqBody := strings.NewReader(`{"data": [{"id": "1", "title": "New Title", "type": "posts"}]}`)
			req, err := http.NewRequest("PUT", "/v1/posts/1", reqBody)
			Expect(err).To(BeNil())
			api.Handler().ServeHTTP(rec, req)
			Expect(rec.Code).To(Equal(http.StatusNoContent))
			Expect(source.posts["1"].Title).To(Equal("New Title"))
		})
	})

	Context("marshal errors correctly", func() {
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

			post1Json = map[string]interface{}{
				"id":    "1",
				"title": "Hello, World!",
				"value": nil,
			}

			api = NewAPI("")
			api.AddResource(Post{}, source)

			rec = httptest.NewRecorder()
		})

		It("POSTSs new objects", func() {
			reqBody := strings.NewReader(`{"data": [{"title": "", "type": "posts"}]}`)
			req, err := http.NewRequest("POST", "/posts", reqBody)
			Expect(err).To(BeNil())
			api.Handler().ServeHTTP(rec, req)
			Expect(rec.Code).To(Equal(http.StatusBadRequest))
			expected := `{"errors":[{"id":"SomeErrorID","path":"Title"}]}`
			actual := strings.TrimSpace(string(rec.Body.Bytes()))
			Expect(actual).To(Equal(expected))
		})
	})

	Context("Extracting query parameters", func() {
		var (
			source    *fixtureSource
			post1JSON map[string]interface{}
			post2JSON map[string]interface{}

			api *API
			rec *httptest.ResponseRecorder
		)

		BeforeEach(func() {
			source = &fixtureSource{map[string]*Post{
				"1": &Post{ID: "1", Title: "Hello, World!"},
				"2": &Post{ID: "2", Title: "Hello, from second Post!"},
			}}

			post1JSON = map[string]interface{}{
				"id":    "1",
				"type":  "posts",
				"title": "Hello, World!",
				"value": nil,
				"links": map[string]interface{}{
					"author": map[string]interface{}{
						"type": "users",
						//"resource": "/posts/1/author",
					},
					"comments": map[string]interface{}{
						"type": "comments",
						//"resource": "/posts/1/comments",
					},
				},
			}

			post2JSON = map[string]interface{}{
				"id":    "2",
				"type":  "posts",
				"title": "Hello, from second Post!",
				"value": nil,
				"links": map[string]interface{}{
					"author": map[string]interface{}{
						"type": "users",
						//"resource": "/posts/2/author",
					},
					"comments": map[string]interface{}{
						"type": "comments",
						//"resource": "/posts/2/comments",
					},
				},
			}

			api = NewAPI("")
			api.AddResource(Post{}, source)

			rec = httptest.NewRecorder()
		})

		It("FindAll returns 2 posts if no limit was set", func() {
			req, err := http.NewRequest("GET", "/posts", nil)
			Expect(err).To(BeNil())
			api.Handler().ServeHTTP(rec, req)
			Expect(rec.Code).To(Equal(http.StatusOK))
			var result map[string]interface{}
			Expect(json.Unmarshal(rec.Body.Bytes(), &result)).To(BeNil())
			Expect(result).To(Equal(map[string]interface{}{
				"data": []interface{}{post1JSON, post2JSON},
			}))
		})

		It("FindAll returns 1 post with limit 1", func() {
			req, err := http.NewRequest("GET", "/posts?limit=1", nil)
			Expect(err).To(BeNil())
			api.Handler().ServeHTTP(rec, req)
			Expect(rec.Code).To(Equal(http.StatusOK))
			var result map[string]interface{}
			Expect(json.Unmarshal(rec.Body.Bytes(), &result)).To(BeNil())
			Expect(result).To(Equal(map[string]interface{}{
				"data": []interface{}{post1JSON},
			}))
		})

		It("Extracts multiple parameters correctly", func() {
			req, err := http.NewRequest("GET", "/posts?sort=title,date", nil)
			Expect(err).To(BeNil())

			api2goReq := buildRequest(req)
			Expect(api2goReq.QueryParams).To(Equal(map[string][]string{"sort": []string{"title", "date"}}))
		})
	})
})
