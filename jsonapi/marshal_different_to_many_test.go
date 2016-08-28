package jsonapi

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type ManyParent struct {
	ID      string `json:"-"`
	Content string `json:"content"`
}

func (m ManyParent) GetID() string {
	return m.ID
}

func (m ManyParent) GetReferences() []Reference {
	return []Reference{
		{
			Type: "childs",
			Name: "childs",
		},
	}
}

func (m ManyParent) GetReferencedIDs() []ReferenceID {
	return []ReferenceID{
		{
			Type: "childs",
			Name: "childs",
			ID:   "one",
		},
		{
			Type: "other-childs",
			Name: "childs",
			ID:   "two",
		},
	}
}

var _ = Describe("Marshalling toMany relations with the same name and different types", func() {
	var toMarshal ManyParent

	BeforeEach(func() {
		toMarshal = ManyParent{
			ID:      "one",
			Content: "test",
		}
	})

	It("marshals toMany relationships with different type and same name", func() {
		result, err := Marshal(toMarshal)
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(MatchJSON(`{
      		"data": {
        		"attributes": {
          			"content": "test"
        		},
        		"id": "one",
        		"relationships": {
          			"childs": {
            			"data": [
            				{
              					"id": "one",
              					"type": "childs"
            				},
            				{
              					"id": "two",
              					"type": "other-childs"
            				}
            			]
          			}
        		},
        		"type": "manyParents"
      		}
    	}`))
	})
})
