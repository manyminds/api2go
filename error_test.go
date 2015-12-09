package api2go

import (
	"errors"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type ErrorMarshaler struct{}

func (e ErrorMarshaler) Marshal(i interface{}) ([]byte, error) {
	return []byte{}, errors.New("this will always fail")
}
func (e ErrorMarshaler) Unmarshal(data []byte, i interface{}) error {
	return nil
}

func (e ErrorMarshaler) MarshalError(error) string {
	return ""
}

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
			httpErr := NewHTTPError(nil, "Invalid use case done", http.StatusInternalServerError)
			result := marshalHTTPError(httpErr)
			expected := `{"errors":[{"status":"500","title":"Invalid use case done"}]}`
			Expect(result).To(Equal(expected))
		})

		It("will be marshalled correctly without child errors", func() {
			httpErr := NewHTTPError(errors.New("Bad Request"), "Bad Request", 400)
			result := marshalHTTPError(httpErr)
			expected := `{"errors":[{"status":"400","title":"Bad Request"}]}`
			Expect(result).To(Equal(expected))
		})

		It("will be marshalled correctly with child errors", func() {
			httpErr := NewHTTPError(errors.New("Bad Request"), "Bad Request", 500)

			errorOne := Error{
				ID: "001",
				Links: &ErrorLinks{
					About: "http://bla/blub",
				},
				Status: "500",
				Code:   "001",
				Title:  "Title must not be empty",
				Detail: "Never occures in real life",
				Source: &ErrorSource{
					Pointer: "#titleField",
				},
				Meta: map[string]interface{}{
					"creator": "api2go",
				},
			}

			httpErr.Errors = append(httpErr.Errors, errorOne)

			result := marshalHTTPError(httpErr)
			expected := `{"errors":[{"id":"001","links":{"about":"http://bla/blub"},"status":"500","code":"001","title":"Title must not be empty","detail":"Never occures in real life","source":{"pointer":"#titleField"},"meta":{"creator":"api2go"}}]}`
			Expect(result).To(Equal(expected))
		})

		It("will be marshalled correctly with child errors without links or source", func() {
			httpErr := NewHTTPError(errors.New("Bad Request"), "Bad Request", 500)

			errorOne := Error{
				ID:     "001",
				Status: "500",
				Code:   "001",
				Title:  "Title must not be empty",
				Detail: "Never occures in real life",
				Meta: map[string]interface{}{
					"creator": "api2go",
				},
			}

			httpErr.Errors = append(httpErr.Errors, errorOne)

			result := marshalHTTPError(httpErr)
			expected := `{"errors":[{"id":"001","status":"500","code":"001","title":"Title must not be empty","detail":"Never occures in real life","meta":{"creator":"api2go"}}]}`
			Expect(result).To(Equal(expected))
		})
	})
})
