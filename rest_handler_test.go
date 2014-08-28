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
			post1 = Post{Title: "Hello, World!"}
			post1Map = map[string]interface{}{
				"id":    "0",
				"title": "Hello, World!",
			}
			rec = httptest.NewRecorder()
			handler = HandlerForResource("posts", func() interface{} {
				return []Post{post1}
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
	})
})
