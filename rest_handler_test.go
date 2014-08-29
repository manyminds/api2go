package api2go

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("RestHandler", func() {
	Context("when handling requests", func() {
		type Post struct {
			ID    int
			Title string
		}

		var (
			handler http.Handler
			rec     *httptest.ResponseRecorder

			post1    Post
			post1Map map[string]interface{}
		)

		BeforeEach(func() {
			post1 = Post{ID: 1, Title: "Hello, World!"}
			post1Map = map[string]interface{}{
				"id":    "1",
				"title": "Hello, World!",
			}
			rec = httptest.NewRecorder()
			handler = HandlerForResource("posts", func() interface{} {
				return []Post{post1}
			}, func(id string) interface{} {
				switch id {
				case "1":
					return post1
				default:
					panic("unknown id " + id)
				}
			})
		})

		It("GETs collections", func() {
			req, err := http.NewRequest("GET", "/posts", nil)
			Expect(err).To(BeNil())
			handler.ServeHTTP(rec, req)
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
			handler.ServeHTTP(rec, req)
			Expect(rec.Code).To(Equal(http.StatusOK))
			var result map[string]interface{}
			Expect(json.Unmarshal(rec.Body.Bytes(), &result)).To(BeNil())
			Expect(result).To(Equal(map[string]interface{}{
				"posts": []interface{}{post1Map},
			}))
		})
	})
})
