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

func (s PublishStatus) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

func (s *PublishStatus) UnmarshalText(text []byte) error {
	label := string(text)

	for key, v := range publishStatusValues {
		if v == label {
			*s = PublishStatus(key)
			return nil
		}
	}

	return fmt.Errorf("invalid value `%s`", label)
}

func (s *PublishStatus) UnmarshalJSON(data []byte) error {
	var text string
	if err := json.Unmarshal(data, &text); err != nil {
		return err
	}
	return s.UnmarshalText([]byte(text))
}

type EnumPost struct {
	ID     string        `json:"-"`
	LID    string        `json:"-"`
	Title  string        `json:"title"`
	Status PublishStatus `json:"status"`
}

func (e EnumPost) GetID() Identifier {
	return Identifier{ID: e.ID, LID: e.LID}
}

func (e EnumPost) GetName() string {
	return "enum-posts"
}

func (e *EnumPost) SetID(ID Identifier) error {
	e.ID = ID.ID
	e.LID = ID.LID
	return nil
}

var _ = Describe("Custom enum types", func() {
	status := StatusPublished
	statusValue := "published"
	singleJSON := []byte(`{"data":{"id": "1", "type": "enum-posts", "attributes": {"title":"First Post","status":"published"}}}`)
	firstPost := EnumPost{ID: "1", Title: "First Post", Status: StatusPublished}

	Context("When marshaling objects including enumes", func() {
		singlePost := EnumPost{
			ID:     "1",
			Title:  "First Post",
			Status: StatusPublished,
		}

		It("marshals JSON", func() {
			result, err := Marshal(singlePost)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(MatchJSON(singleJSON))
		})
	})

	Context("When unmarshaling objects including enums", func() {
		It("unmarshals status string values to int enum type", func() {
			var result PublishStatus
			result.UnmarshalText([]byte(statusValue))
			Expect(result).To(Equal(status))
		})

		It("unmarshals single objects into a struct", func() {
			// Todo: Hm, what was that test for? I don't remember, maybe delete it, but now it checks empty jsons and
			// raises an error which is also a good thing
			var post EnumPost
			err := Unmarshal([]byte("{}"), &post)
			Expect(err).To(HaveOccurred())
		})

		It("unmarshals JSON", func() {
			var post EnumPost
			err := Unmarshal(singleJSON, &post)
			Expect(err).ToNot(HaveOccurred())
			Expect(post).To(Equal(firstPost))
		})
	})
})
