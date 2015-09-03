package jsonapi

import (
	"encoding/json"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type PublishStatus int

const (
	StatusUnpublished PublishStatus = iota
	StatusPublished
)

var publishStatusValues = []string{
	StatusUnpublished: "unpublished",
	StatusPublished:   "published",
}

func (s PublishStatus) String() string {
	if s < 0 || int(s) >= len(publishStatusValues) {
		panic("value out of range")
	}
	return publishStatusValues[s]
}

// MarshalText implements the TextMarshaler interface.
func (s PublishStatus) MarshalText() (text []byte, err error) {
	return []byte(s.String()), nil
}

// UnmarshalText implements the TextUnmarshaler interface.
func (s *PublishStatus) UnmarshalText(text []byte) error {
	var label string
	json.Unmarshal(text, &label)

	for key, v := range publishStatusValues {
		if v == label {
			*s = PublishStatus(key)
			return nil
		}
	}

	return fmt.Errorf("invalid value `%s`", label)
}

func (s *PublishStatus) UnmarshalJSON(data []byte) error {
	return s.UnmarshalText(data)
}

type EnumPost struct {
	ID     string `json:"-"`
	Title  string
	Status PublishStatus
}

func (e EnumPost) GetID() string {
	return e.ID
}

func (e *EnumPost) SetID(ID string) error {
	e.ID = ID

	return nil
}

var _ = Describe("Unmarshal", func() {
	Context("When unmarshaling objects including enums", func() {
		status := StatusPublished
		statusValue := "\"published\""
		singleJSON := []byte(`{"data":{"id": "1", "type": "enumPosts", "attributes": {"title":"First Post","status":"published"}}}`)
		firstPost := EnumPost{ID: "1", Title: "First Post", Status: StatusPublished}
		singlePostMap := map[string]interface{}{
			"data": map[string]interface{}{
				"id":   "1",
				"type": "enumPosts",
				"attributes": map[string]interface{}{
					"title":  firstPost.Title,
					"status": StatusPublished,
				},
			},
		}

		It("unmarshals status string values to int enum type", func() {
			var result PublishStatus
			result.UnmarshalText([]byte(statusValue))
			Expect(result).To(Equal(status))
		})

		It("unmarshals single objects into a struct", func() {
			var post EnumPost
			err := Unmarshal(singlePostMap, &post)
			Expect(err).To(BeNil())
			Expect(post).To(Equal(firstPost))
		})

		It("unmarshals JSON", func() {
			var posts []EnumPost
			err := UnmarshalFromJSON(singleJSON, &posts)
			Expect(err).To(BeNil())
			Expect(posts).To(Equal([]EnumPost{firstPost}))
		})
	})
})
