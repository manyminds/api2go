package api2go

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Errors test", func() {
	Context("validate error logic", func() {
		It("can create array tree", func() {
			httpErr := NewHTTPError(errors.New("hi"), "hi", 0)
			for i := 0; i < 20; i++ {
				httpErr.Errors = append(httpErr.Errors, Error{})
			}

			Expect(len(httpErr.Errors)).To(Equal(20))
		})
	})

	Context("Marshalling", func() {
		It("will be marshalled correctly with default error", func() {
			httpErr := errors.New("Invalid use case done")
			result := marshalError(httpErr)
			expected := `{"errors":[{"status":"500","title":"Invalid use case done"}]}`
			Expect(result).To(Equal(expected))
		})

		It("will be marshalled correctly without child errors", func() {
			httpErr := NewHTTPError(errors.New("Bad Request"), "Bad Request", 400)
			result := marshalError(httpErr)
			expected := `{"errors":[{"status":"400","title":"Bad Request"}]}`
			Expect(result).To(Equal(expected))
		})

		It("will be marshalled correctly with child errors", func() {
			httpErr := NewHTTPError(errors.New("Bad Request"), "Bad Request", 500)

			errorOne := Error{
				ID:     "001",
				Href:   "http://bla/blub",
				Status: "500",
				Code:   "001",
				Title:  "Title must not be empty",
				Detail: "Never occures in real life",
				Path:   "#titleField",
			}

			httpErr.Errors = append(httpErr.Errors, errorOne)

			result := marshalError(httpErr)
			expected := `{"errors":[{"id":"001","href":"http://bla/blub","status":"500","code":"001","title":"Title must not be empty","detail":"Never occures in real life","path":"#titleField"}]}`
			Expect(result).To(Equal(expected))
		})
	})
})
