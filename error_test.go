package api2go

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Errors test", func() {
	Context("validate error logic", func() {
		It("can create array tree", func() {
			err := NewHTTPError(errors.New("hi"), "hi", 0)
			httpErr, ok := err.(httpError)
			for i := 0; i < 20; i++ {
				httpErr.AddError(errors.New("Some error"), "Invalid error error")
			}
			Expect(ok).To(Equal(true))
			Expect(len(httpErr.errors)).To(Equal(20))
		})
	})
})
