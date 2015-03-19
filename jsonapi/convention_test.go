package jsonapi_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/univedo/api2go/jsonapi"
)

type Oldschool struct {
	ID string
}

var _ = Describe("Convention", func() {
	Context("test backwards compatibility for primary ID", func() {
		It("Should not getID via magic if no ID field is present", func() {
			item := &ConventionWrapper{Item: Oldschool{ID: "someID"}}
			Expect(item.GetID()).To(Equal("someID"))
		})

		It("Should not getID via magic if no ID field is present", func() {
			item := &ConventionWrapper{Item: Oldschool{ID: "someID"}}
			Expect(item.GetID()).To(Equal("someID"))
		})

		It("Should getID via magic", func() {
			type Stupid struct {
				EiDi string
			}
			item := &ConventionWrapper{Item: Stupid{EiDi: "someID"}}
			Expect(item.GetID()).To(Equal(NoIDFieldPresent))
		})
	})

	Context("test backwards compatibility for referenced ids", func() {

	})
})
