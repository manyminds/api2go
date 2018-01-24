package api2go

import (
	"net/http"
	"net/http/httptest"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type ObjectInitializerResource struct{}

func (s ObjectInitializerResource) InitializeObject(obj interface{}) {
	if post, ok := obj.(*Post); ok {
		post.Title = "New Title"
	}
}

func (s ObjectInitializerResource) FindOne(ID string, req Request) (Responder, error) {
	return nil, nil
}

func (s ObjectInitializerResource) Create(obj interface{}, req Request) (Responder, error) {
	return &Response{Res: obj, Code: http.StatusCreated}, nil
}

func (s ObjectInitializerResource) Delete(ID string, req Request) (Responder, error) {
	return nil, nil
}

func (s ObjectInitializerResource) Update(obj interface{}, req Request) (Responder, error) {
	return nil, nil
}

var _ = Describe("Test resource implementing the ObjectInitializer interface", func() {
	var (
		api  *API
		rec  *httptest.ResponseRecorder
		body *strings.Reader
	)
	BeforeEach(func() {
		api = NewAPIWithRouting(testPrefix, NewStaticResolver(""), newTestRouter())
		api.AddResource(Post{}, ObjectInitializerResource{})
		rec = httptest.NewRecorder()
		body = strings.NewReader(`
		{
			"data": {
				"attributes": {},
				"id": "blubb",
				"type": "posts"
			}
		}
		`)
	})

	It("Create", func() {
		req, err := http.NewRequest("POST", "/v1/posts", body)
		Expect(err).ToNot(HaveOccurred())
		api.Handler().ServeHTTP(rec, req)

		Expect(rec.Body.String()).To(MatchJSON(`
		{
        	"data": {
          		"type": "posts",
          		"id": "blubb",
          		"attributes": {
					"title": "New Title",
            		"value": null
          		},
          		"relationships": {
					"author": {
						"links": {
							"self": "/v1/posts/blubb/relationships/author",
							"related": "/v1/posts/blubb/author"
						},
						"data": null
					},
					"bananas": {
						"links": {
							"self": "/v1/posts/blubb/relationships/bananas",
							"related": "/v1/posts/blubb/bananas"
						},
						"data": []
					},
					"comments": {
						"links": {
							"self": "/v1/posts/blubb/relationships/comments",
							"related": "/v1/posts/blubb/comments"
						},
						"data": []
					}
          		}
        	}
     	}
		`))
		Expect(rec.Code).To(Equal(http.StatusCreated))
	})
})
