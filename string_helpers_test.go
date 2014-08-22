package api2go_test

import (
	"github.com/univedo/api2go"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("StringHelpers", func() {
	It("underscorizes", func() {
		Expect(api2go.Underscorize("Post")).To(Equal("post"))
		Expect(api2go.Underscorize("post")).To(Equal("post"))
		Expect(api2go.Underscorize("SimplePost")).To(Equal("simple_post"))
		Expect(api2go.Underscorize("simple_post")).To(Equal("simple_post"))
		Expect(api2go.Underscorize("XML")).To(Equal("xml"))
		Expect(api2go.Underscorize("XMLPost")).To(Equal("xml_post"))
	})

	It("pluralizes", func() {
		Expect(api2go.Pluralize("post")).To(Equal("posts"))
		Expect(api2go.Pluralize("posts")).To(Equal("posts"))
		Expect(api2go.Pluralize("category")).To(Equal("categories"))
	})
})
