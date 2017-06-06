package api2go

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Integration", func() {

	Context("Simple datastructure", func() {
		type Structure struct {
			ID    string
			Value string
		}

		It("Should be marshalled and unmarshalled", func() {
			testdata := []byte(`{"structures":[{"id":"1","value":"Some Contents"}]}`)

			structures := []Structure{}
			err := UnmarshalFromJSON(testdata, &structures)
			Expect(err).ToNot(HaveOccurred())

			json, err := MarshalToJSON(structures)
			Expect(err).ToNot(HaveOccurred())

			Expect(testdata).To(Equal(json))
		})
	})

	Context("Simple datastructure with foreignID", func() {
		type Structure struct {
			ID        string
			Value     string
			ForeignID string
		}

		It("Should be marshalled and unmarshalled", func() {
			testdata := []byte(`{"structures":[{"id": "1", "Value" : "Some Contents"}]}`)

			structures := []Structure{}
			err := UnmarshalFromJSON(testdata, &structures)
			Expect(err).ToNot(HaveOccurred())

			json, err := MarshalToJSON(structures)
			Expect(err).ToNot(HaveOccurred())

			Expect(testdata).To(Equal(json))
		})
	})

})
