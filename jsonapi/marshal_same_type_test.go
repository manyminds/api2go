package jsonapi

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type Node struct {
	ID                string   `json:"-"`
	Content           string   `json:"content"`
	MotherID          string   `json:"-"`
	ChildIDs          []string `json:"-"`
	AbandonedChildIDs []string `json:"-"`
}

func (n *Node) GetID() string {
	return n.ID
}

func (n *Node) GetReferences() []Reference {
	return []Reference{
		{
			Type: "nodes",
			Name: "mother-node",
		},
		{
			Type: "nodes",
			Name: "child-nodes",
		},
		{
			Type: "nodes",
			Name: "abandoned-child-nodes",
		},
	}
}

func (n *Node) GetReferencedIDs() []ReferenceID {
	result := []ReferenceID{}

	if n.MotherID != "" {
		result = append(result, ReferenceID{Type: "nodes", Name: "mother-node", ID: n.MotherID})
	}

	for _, referenceID := range n.ChildIDs {
		result = append(result, ReferenceID{Type: "nodes", Name: "child-nodes", ID: referenceID})
	}

	for _, referenceID := range n.AbandonedChildIDs {
		result = append(result, ReferenceID{Type: "nodes", Name: "abandoned-child-nodes", ID: referenceID})
	}

	return result
}

var _ = Describe("Marshalling with the same reference type", func() {
	var theNode Node

	BeforeEach(func() {
		theNode = Node{
			ID:                "super",
			Content:           "I am the Super Node",
			MotherID:          "1337",
			ChildIDs:          []string{"666", "42"},
			AbandonedChildIDs: []string{"2", "1"},
		}
	})

	It("marshals all the relationships of the same type", func() {
		i, err := Marshal(&theNode)
		Expect(err).To(BeNil())
		Expect(i).To(MatchJSON(`
		{
			"data": {
				"type": "nodes",
				"id": "super",
				"attributes": {
					"content": "I am the Super Node"
				},
				"relationships": {
					"abandoned-child-nodes": {
						"data": [
						{
							"type": "nodes",
							"id": "2"
						},
						{
							"type": "nodes",
							"id": "1"
						}
						]
					},
					"child-nodes": {
						"data": [
						{
							"type": "nodes",
							"id": "666"
						},
						{
							"type": "nodes",
							"id": "42"
						}
						]
					},
					"mother-node": {
						"data": {
							"type": "nodes",
							"id": "1337"
						}
					}
				}
			}
		}
		`))
	})
})
