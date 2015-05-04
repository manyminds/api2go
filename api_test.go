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
	"github.com/manyminds/api2go/jsonapi"
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

func (p *Post) SetToOneReferenceID(name, ID string) error {
	if name == "author" {
		if ID == "" {
			p.Author = nil
		} else {
			p.Author = &User{ID: ID}
		}

		return nil
	}

	return errors.New("There is no to-one relationship with the name " + name)
}

func (p *Post) SetToManyReferenceIDs(name string, IDs []string) error {
	if name == "comments" {
		comments := []Comment{}
		for _, ID := range IDs {
			comments = append(comments, Comment{ID: ID})
		}
		p.Comments = comments
	}

	return errors.New("There is no to-many relationship with the name " + name)
}

func (p *Post) AddToManyIDs(name string, IDs []string) error {
	if name == "comments" {
		for _, ID := range IDs {
			p.Comments = append(p.Comments, Comment{ID: ID})
		}
	}

	return errors.New("There is no to-manyrelationship with the name " + name)
}

func (p *Post) DeleteToManyIDs(name string, IDs []string) error {
	if name == "comments" {
		for _, ID := range IDs {
			// find and delete the comment with ID
			for pos, comment := range p.Comments {
				if comment.GetID() == ID {
					p.Comments = append(p.Comments[:pos], p.Comments[pos+1:]...)
				}
			}
		}
	}

	return errors.New("There is no to-manyrelationship with the name " + name)
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

// this does not read the query parameters, which you would do to limit the result in real world usage
func (s *fixtureSource) PaginatedFindAll(req Request) (interface{}, uint, error) {
	postsSlice := []Post{}

	for _, post := range s.posts {
		postsSlice = append(postsSlice, *post)
	}

	return postsSlice, uint(len(s.posts)), nil
}

func (s *fixtureSource) FindOne(id string, req Request) (interface{}, error) {
	if p, ok := s.posts[id]; ok {
		return *p, nil
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
		oldP.Author = p.Author
		oldP.Comments = p.Comments
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
						"linkage": map[string]interface{}{
							"id":   "1",
							"type": "users",
						},
						"self":    "/v1/posts/1/links/author",
						"related": "/v1/posts/1/author",
					},
					"comments": map[string]interface{}{
						"linkage": []map[string]interface{}{
							map[string]interface{}{
								"id":   "1",
								"type": "comments",
							},
						},
						"self":    "/v1/posts/1/links/comments",
						"related": "/v1/posts/1/comments",
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
						"linkage": nil,
						"self":    "/v1/posts/2/links/author",
						"related": "/v1/posts/2/author",
					},
					"comments": map[string]interface{}{
						"linkage": []interface{}{},
						"self":    "/v1/posts/2/links/comments",
						"related": "/v1/posts/2/comments",
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
						"linkage": nil,
						"self":    "/v1/posts/3/links/author",
						"related": "/v1/posts/3/author",
					},
					"comments": map[string]interface{}{
						"linkage": []interface{}{},
						"self":    "/v1/posts/3/links/comments",
						"related": "/v1/posts/3/comments",
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
				"data":     []map[string]interface{}{post1Json, post2Json, post3Json},
				"included": post1LinkedJSON,
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
				"data":     post1Json,
				"included": post1LinkedJSON,
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

		It("GETs relationship data from relationship url for to-many", func() {
			req, err := http.NewRequest("GET", "/v1/posts/1/links/comments", nil)
			Expect(err).ToNot(HaveOccurred())
			api.Handler().ServeHTTP(rec, req)
			Expect(rec.Code).To(Equal(http.StatusOK))
			Expect(rec.Body.Bytes()).To(MatchJSON(`{"data": [{"id": "1", "type": "comments"}], "links": {"self": "/v1/posts/1/links/comments", "related": "/v1/posts/1/comments"}}`))
		})

		It("GETs relationship data from relationship url for to-one", func() {
			req, err := http.NewRequest("GET", "/v1/posts/1/links/author", nil)
			Expect(err).ToNot(HaveOccurred())
			api.Handler().ServeHTTP(rec, req)
			Expect(rec.Code).To(Equal(http.StatusOK))
			Expect(rec.Body.Bytes()).To(MatchJSON(`{"data": {"id": "1", "type": "users"}, "links": {"self": "/v1/posts/1/links/author", "related": "/v1/posts/1/author"}}`))
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
							"linkage": nil,
							"self":    "/v1/posts/4/links/author",
							"related": "/v1/posts/4/author",
						},
						"comments": map[string]interface{}{
							"linkage": []interface{}{},
							"self":    "/v1/posts/4/links/comments",
							"related": "/v1/posts/4/comments",
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
		It("OPTIONS on collection route", func() {
			req, err := http.NewRequest("OPTIONS", "/v1/posts", nil)
			api.Handler().ServeHTTP(rec, req)
			Expect(err).To(BeNil())
			Expect(rec.Code).To(Equal(http.StatusNoContent))
			Expect(rec.Header().Get("Allow")).To(Equal("GET,POST,PATCH,OPTIONS"))
		})

		It("OPTIONS on element route", func() {
			req, err := http.NewRequest("OPTIONS", "/v1/posts/1", nil)
			api.Handler().ServeHTTP(rec, req)
			Expect(err).To(BeNil())
			Expect(rec.Code).To(Equal(http.StatusNoContent))
			Expect(rec.Header().Get("Allow")).To(Equal("GET,PATCH,DELETE,OPTIONS"))
		})

		It("DELETEs", func() {
			req, err := http.NewRequest("DELETE", "/v1/posts/1", nil)
			Expect(err).To(BeNil())
			api.Handler().ServeHTTP(rec, req)
			Expect(rec.Code).To(Equal(http.StatusNoContent))
			Expect(len(source.posts)).To(Equal(2))
		})

		It("patch must contain type and id but does not have type", func() {
			reqBody := strings.NewReader(`{"data": {"title": "New Title", "id": "id"}}`)
			req, err := http.NewRequest("PATCH", "/v1/posts/1", reqBody)
			Expect(err).To(BeNil())
			api.Handler().ServeHTTP(rec, req)
			Expect(rec.Code).To(Equal(http.StatusForbidden))
			Expect(string(rec.Body.Bytes())).To(MatchJSON(`{"errors":[{"status":"403","title":"missing mandatory type key."}]}`))
		})

		It("patch must contain type and id but does not have id", func() {
			reqBody := strings.NewReader(`{"data": {"title": "New Title", "type": "posts"}}`)
			req, err := http.NewRequest("PATCH", "/v1/posts/1", reqBody)
			Expect(err).To(BeNil())
			api.Handler().ServeHTTP(rec, req)
			Expect(rec.Code).To(Equal(http.StatusForbidden))
			Expect(string(rec.Body.Bytes())).To(MatchJSON(`{"errors":[{"status":"403","title":"missing mandatory id key."}]}`))
		})

		Context("Updating", func() {
			doRequest := func(payload, url, method string) {
				reqBody := strings.NewReader(payload)
				req, err := http.NewRequest(method, url, reqBody)
				Expect(err).To(BeNil())
				api.Handler().ServeHTTP(rec, req)
				Expect(rec.Body.String()).To(Equal(""))
				Expect(rec.Code).To(Equal(http.StatusNoContent))
			}

			It("UPDATEs", func() {
				target := source.posts["1"]
				target.Value = null.FloatFrom(2)
				doRequest(`{"data": {"id": "1", "title": "New Title", "type": "posts"}}`, "/v1/posts/1", "PATCH")
				Expect(source.posts["1"].Title).To(Equal("New Title"))
				Expect(target.Title).To(Equal("New Title"))
				Expect(target.Value).To(Equal(null.FloatFrom(2)))
			})

			It("Patch updates to-one relationships", func() {
				target := source.posts["1"]
				doRequest(`{
				"data": {
					"type": "posts",
					"id": "1",
					"links": {
						"author": {
							"linkage": {
								"type": "users",
								"id": "2"
							}
						}
					}
				}
			}
			`, "/v1/posts/1", "PATCH")
				Expect(target.Author.GetID()).To(Equal("2"))
			})

			It("Patch can delete to-one relationships", func() {
				target := source.posts["1"]
				doRequest(`{
				"data": {
					"type": "posts",
					"id": "1",
					"links": {
						"author": {
							"linkage": null
						}
					}
				}
			}
			`, "/v1/posts/1", "PATCH")
				Expect(target.Author).To(BeNil())
			})

			It("Patch updates to-many relationships", func() {
				target := source.posts["1"]
				doRequest(`{
				"data": {
					"type": "posts",
					"id": "1",
					"links": {
						"comments": {
							"linkage": [
								{
									"type": "comments",
									"id": "2"
								}
							]
						}
					}
				}
			}
			`, "/v1/posts/1", "PATCH")
				Expect(target.Comments[0].GetID()).To(Equal("2"))
			})

			It("Patch can delete to-many relationships", func() {
				target := source.posts["1"]
				doRequest(`{
				"data": {
					"type": "posts",
					"id": "1",
					"links": {
						"comments": {
							"linkage": []
						}
					}
				}
			}
			`, "/v1/posts/1", "PATCH")
				Expect(target.Comments).To(HaveLen(0))
			})

			It("Relationship PATCH route updates to-one", func() {
				doRequest(`{
				"data": {
					"type": "users",
					"id": "2"
				}
			}`, "/v1/posts/1/links/author", "PATCH")
				target := source.posts["1"]
				Expect(target.Author.GetID()).To(Equal("2"))
			})

			It("Relationship PATCH route updates to-many", func() {
				doRequest(`{
				"data": [{
					"type": "comments",
					"id": "2"
				}]
			}`, "/v1/posts/1/links/comments", "PATCH")
				target := source.posts["1"]
				Expect(target.Comments).To(HaveLen(1))
				Expect(target.Comments[0].GetID()).To(Equal("2"))
			})

			It("Relationship POST route adds to-many elements", func() {
				doRequest(`{
				"data": [{
					"type": "comments",
					"id": "2"
				}]
			}`, "/v1/posts/1/links/comments", "POST")
				target := source.posts["1"]
				Expect(target.Comments).To(HaveLen(2))
			})

			It("Relationship DELETE route deletes to-many elements", func() {
				doRequest(`{
				"data": [{
					"type": "comments",
					"id": "1"
				}]
			}`, "/v1/posts/1/links/comments", "DELETE")
				target := source.posts["1"]
				Expect(target.Comments).To(HaveLen(0))
			})
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

	Context("Extracting query parameters with complete BaseURL API", func() {
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
						"linkage": nil,
						"self":    "http://localhost:1337/v0/posts/1/links/author",
						"related": "http://localhost:1337/v0/posts/1/author",
					},
					"comments": map[string]interface{}{
						"linkage": []interface{}{},
						"self":    "http://localhost:1337/v0/posts/1/links/comments",
						"related": "http://localhost:1337/v0/posts/1/comments",
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
						"linkage": nil,
						"self":    "http://localhost:1337/v0/posts/2/links/author",
						"related": "http://localhost:1337/v0/posts/2/author",
					},
					"comments": map[string]interface{}{
						"linkage": []interface{}{},
						"self":    "http://localhost:1337/v0/posts/2/links/comments",
						"related": "http://localhost:1337/v0/posts/2/comments",
					},
				},
			}

			api = NewAPIWithBaseURL("v0", "http://localhost:1337")

			api.AddResource(Post{}, source)

			rec = httptest.NewRecorder()
		})

		It("FindAll returns 2 posts if no limit was set", func() {
			req, err := http.NewRequest("GET", "/v0/posts", nil)
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
			req, err := http.NewRequest("GET", "/v0/posts?limit=1", nil)
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
			req, err := http.NewRequest("GET", "/v0/posts?sort=title,date", nil)
			Expect(err).To(BeNil())

			api2goReq := buildRequest(req)
			Expect(api2goReq.QueryParams).To(Equal(map[string][]string{"sort": []string{"title", "date"}}))
		})
	})

	Context("When using pagination", func() {
		var (
			api    *API
			rec    *httptest.ResponseRecorder
			source *fixtureSource
		)

		BeforeEach(func() {
			source = &fixtureSource{map[string]*Post{
				"1": &Post{ID: "1", Title: "Hello, World!"},
				"2": &Post{ID: "2", Title: "Hello, World!"},
				"3": &Post{ID: "3", Title: "Hello, World!"},
				"4": &Post{ID: "4", Title: "Hello, World!"},
				"5": &Post{ID: "5", Title: "Hello, World!"},
				"6": &Post{ID: "6", Title: "Hello, World!"},
				"7": &Post{ID: "7", Title: "Hello, World!"},
			}}

			api = NewAPI("v1")
			api.AddResource(Post{}, source)

			rec = httptest.NewRecorder()
		})

		// helper function that does a request and returns relevant pagination urls out of the response body
		doRequest := func(URL string) map[string]string {
			req, err := http.NewRequest("GET", URL, nil)
			Expect(err).ToNot(HaveOccurred())
			api.Handler().ServeHTTP(rec, req)
			Expect(rec.Code).To(Equal(http.StatusOK))
			var response map[string]interface{}
			Expect(json.Unmarshal(rec.Body.Bytes(), &response)).To(BeNil())

			result := map[string]string{}
			if links, ok := response["links"].(map[string]interface{}); ok {
				if first, ok := links["first"]; ok {
					result["first"] = first.(string)
					Expect(err).ToNot(HaveOccurred())
				}
				if next, ok := links["next"]; ok {
					result["next"] = next.(string)
					Expect(err).ToNot(HaveOccurred())
				}
				if prev, ok := links["prev"]; ok {
					result["prev"] = prev.(string)
					Expect(err).ToNot(HaveOccurred())
				}
				if last, ok := links["last"]; ok {
					result["last"] = last.(string)
					Expect(err).ToNot(HaveOccurred())
				}
			}

			return result
		}

		Context("number & size links", func() {
			It("No prev and first on first page, size = 1", func() {
				links := doRequest("/v1/posts?page[number]=1&page[size]=1")
				Expect(links).To(HaveLen(2))
				Expect(links["next"]).To(Equal("/v1/posts?page[number]=2&page[size]=1"))
				Expect(links["last"]).To(Equal("/v1/posts?page[number]=7&page[size]=1"))
			})

			It("No prev and first on first page, size = 2", func() {
				links := doRequest("/v1/posts?page[number]=1&page[size]=2")
				Expect(links).To(HaveLen(2))
				Expect(links["next"]).To(Equal("/v1/posts?page[number]=2&page[size]=2"))
				Expect(links["last"]).To(Equal("/v1/posts?page[number]=4&page[size]=2"))
			})

			It("All links on page 2, size = 2", func() {
				links := doRequest("/v1/posts?page[number]=2&page[size]=2")
				Expect(links).To(HaveLen(4))
				Expect(links["first"]).To(Equal("/v1/posts?page[number]=1&page[size]=2"))
				Expect(links["prev"]).To(Equal("/v1/posts?page[number]=1&page[size]=2"))
				Expect(links["next"]).To(Equal("/v1/posts?page[number]=3&page[size]=2"))
				Expect(links["last"]).To(Equal("/v1/posts?page[number]=4&page[size]=2"))
			})

			It("No next and last on last page, size = 2", func() {
				links := doRequest("/v1/posts?page[number]=4&page[size]=2")
				Expect(links).To(HaveLen(2))
				Expect(links["prev"]).To(Equal("/v1/posts?page[number]=3&page[size]=2"))
				Expect(links["first"]).To(Equal("/v1/posts?page[number]=1&page[size]=2"))
			})

			It("Does not generate links if results fit on one page", func() {
				links := doRequest("/v1/posts?page[number]=1&page[size]=10")
				Expect(links).To(HaveLen(0))
			})
		})

		// If the combination of parameters is invalid, no links are generated and the normal FindAll method get's called
		Context("invalid parameter combinations", func() {
			It("all 4 of them", func() {
				links := doRequest("/v1/posts?page[number]=1&page[size]=1&page[offset]=1&page[limit]=1")
				Expect(links).To(HaveLen(0))
			})

			It("number only", func() {
				links := doRequest("/v1/posts?page[number]=1")
				Expect(links).To(HaveLen(0))
			})

			It("size only", func() {
				links := doRequest("/v1/posts?page[size]=1")
				Expect(links).To(HaveLen(0))
			})

			It("offset only", func() {
				links := doRequest("/v1/posts?page[offset]=1")
				Expect(links).To(HaveLen(0))
			})

			It("limit only", func() {
				links := doRequest("/v1/posts?page[limit]=1")
				Expect(links).To(HaveLen(0))
			})

			It("number, size & offset", func() {
				links := doRequest("/v1/posts?page[number]=1&page[size]=1&page[offset]=1")
				Expect(links).To(HaveLen(0))
			})

			It("number, size & limit", func() {
				links := doRequest("/v1/posts?page[number]=1&page[size]=1&page[limit]=1")
				Expect(links).To(HaveLen(0))
			})

			It("limit, offset & number", func() {
				links := doRequest("/v1/posts?page[limit]=1&page[offset]=1&page[number]=1")
				Expect(links).To(HaveLen(0))
			})

			It("limit, offset & size", func() {
				links := doRequest("/v1/posts?page[limit]=1&page[offset]=1&page[size]=1")
				Expect(links).To(HaveLen(0))
			})
		})

		Context("offset & limit links", func() {
			It("No prev and first on offset = 0, limit = 1", func() {
				links := doRequest("/v1/posts?page[offset]=0&page[limit]=1")
				Expect(links).To(HaveLen(2))
				Expect(links["next"]).To(Equal("/v1/posts?page[limit]=1&page[offset]=1"))
				Expect(links["last"]).To(Equal("/v1/posts?page[limit]=1&page[offset]=6"))
			})

			It("No prev and first on offset = 0, limit = 2", func() {
				links := doRequest("/v1/posts?page[offset]=0&page[limit]=2")
				Expect(links).To(HaveLen(2))
				Expect(links["next"]).To(Equal("/v1/posts?page[limit]=2&page[offset]=2"))
				Expect(links["last"]).To(Equal("/v1/posts?page[limit]=2&page[offset]=5"))
			})

			It("All links on offset = 2, limit = 2", func() {
				links := doRequest("/v1/posts?page[offset]=2&page[limit]=2")
				Expect(links).To(HaveLen(4))
				Expect(links["first"]).To(Equal("/v1/posts?page[limit]=2&page[offset]=0"))
				Expect(links["prev"]).To(Equal("/v1/posts?page[limit]=2&page[offset]=0"))
				Expect(links["next"]).To(Equal("/v1/posts?page[limit]=2&page[offset]=4"))
				Expect(links["last"]).To(Equal("/v1/posts?page[limit]=2&page[offset]=5"))
			})

			It("No next and last on offset = 5, limit = 2", func() {
				links := doRequest("/v1/posts?page[offset]=5&page[limit]=2")
				Expect(links).To(HaveLen(2))
				Expect(links["prev"]).To(Equal("/v1/posts?page[limit]=2&page[offset]=3"))
				Expect(links["first"]).To(Equal("/v1/posts?page[limit]=2&page[offset]=0"))
			})

			It("Does not generate links if results fit on one page", func() {
				links := doRequest("/v1/posts?page[offset]=0&page[limit]=10")
				Expect(links).To(HaveLen(0))
			})
		})
	})
})
