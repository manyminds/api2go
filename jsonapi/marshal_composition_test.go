package jsonapi

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type TaggedPost struct {
	SimplePost
	Tag string
}

var _ = Describe("Embedded struct types", func() {
	created, _ := time.Parse(time.RFC3339, "2014-11-10T16:30:48.823Z")
	post := TaggedPost{
		SimplePost{ID: "first", Title: "First Post", Text: "Lipsum", Created: created},
		"important",
	}
	postMap := map[string]interface{}{
		"type": "taggedPosts",
		"id":   "first",
		"attributes": map[string]interface{}{
			"title":       post.Title,
			"text":        post.Text,
			"size":        0,
			"create-date": created,
			"tag":         post.Tag,
		},
	}

	Context("When marshaling objects with struct composition", func() {
		It("marshals", func() {
			i, err := Marshal(post)
			Expect(err).To(BeNil())
			Expect(i).To(Equal(map[string]interface{}{
				"data": postMap,
			}))
		})
	})
})
