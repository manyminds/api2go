package api2go

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"

	"github.com/manyminds/api2go/jsonapi"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type SomeData struct {
	ID         string `json:"-"`
	Data       string `json:"data"`
	CustomerID string `json:"customerId"`
}

func (s SomeData) GetID() string {
	return s.ID
}

func (s *SomeData) SetID(ID string) error {
	s.ID = ID
	return nil
}

type SomeResource struct{}

func (s SomeResource) FindOne(ID string, req Request) (Responder, error) {
	return &Response{Res: SomeData{ID: "12345", Data: "A Brezzn"}}, nil
}

func (s SomeResource) Create(obj interface{}, req Request) (Responder, error) {
	incoming := obj.(SomeData)
	switch incoming.ID {
	case "":
		return &Response{Res: SomeData{ID: "12345", Data: "A Brezzn"}, Code: http.StatusCreated}, nil
	case "accept":
		return &Response{Res: SomeData{ID: "someID"}, Code: http.StatusAccepted}, nil
	case "forbidden":
		return &Response{}, NewHTTPError(nil, "Forbidden", http.StatusForbidden)
	case "conflict":
		return &Response{}, NewHTTPError(nil, "Conflict", http.StatusConflict)
	case "invalid":
		return &Response{Res: SomeData{}, Code: http.StatusTeapot}, nil
	default:
		return &Response{Res: SomeData{ID: incoming.ID}, Code: http.StatusNoContent}, nil
	}
}

func (s SomeResource) Delete(ID string, req Request) (Responder, error) {
	switch ID {
	case "200":
		return &Response{Code: http.StatusOK, Meta: map[string]interface{}{"some": "cool stuff"}}, nil
	case "202":
		return &Response{Code: http.StatusAccepted}, nil
	default:
		return &Response{Code: http.StatusNoContent}, nil
	}
}

func (s SomeResource) Update(obj interface{}, req Request) (Responder, error) {
	incoming := obj.(SomeData)
	switch incoming.Data {
	case "override me":
		return &Response{Code: http.StatusOK}, nil
	case "delayed":
		return &Response{Code: http.StatusAccepted}, nil
	case "new value":
		return &Response{Code: http.StatusNoContent}, nil
	case "fail":
		return &Response{}, NewHTTPError(nil, "Fail", http.StatusForbidden)
	case "invalid":
		return &Response{Code: http.StatusTeapot}, nil
	default:
		return &Response{Code: http.StatusNoContent}, nil
	}
}

var _ = Describe("Test interface api type casting", func() {
	var (
		api *API
		rec *httptest.ResponseRecorder
	)
	BeforeEach(func() {
		api = NewAPI("v1")
		api.AddResource(SomeData{}, SomeResource{})
		rec = httptest.NewRecorder()
	})

	It("FindAll returns 404 for simple CRUD", func() {
		req, err := http.NewRequest("GET", "/v1/someDatas", nil)
		Expect(err).ToNot(HaveOccurred())
		api.Handler().ServeHTTP(rec, req)
		Expect(rec.Code).To(Equal(http.StatusNotFound))
	})

	It("Works for a normal FindOne", func() {
		req, err := http.NewRequest("GET", "/v1/someDatas/12345", nil)
		Expect(err).ToNot(HaveOccurred())
		api.Handler().ServeHTTP(rec, req)
		Expect(rec.Code).To(Equal(http.StatusOK))
	})

	It("Post works with lowercase renaming", func() {
		reqBody := strings.NewReader(`{"data": {"attributes":{"customerId": "2" }, "type": "someDatas"}}`)
		req, err := http.NewRequest("POST", "/v1/someDatas", reqBody)
		Expect(err).To(BeNil())
		api.Handler().ServeHTTP(rec, req)
		Expect(rec.Code).To(Equal(http.StatusCreated))
	})
})

var _ = Describe("Test return code behavior", func() {
	var (
		api                *API
		rec                *httptest.ResponseRecorder
		payload, payloadID SomeData
	)

	BeforeEach(func() {
		api = NewAPI("v1")
		api.AddResource(SomeData{}, SomeResource{})
		rec = httptest.NewRecorder()
		payloadID = SomeData{ID: "12345", Data: "A Brezzn"}
		payload = SomeData{Data: "A Brezzn"}
	})

	Context("Create", func() {
		post := func(payload SomeData) {
			m, err := jsonapi.Marshal(payload)
			Expect(err).ToNot(HaveOccurred())
			req, err := http.NewRequest("POST", "/v1/someDatas", strings.NewReader(string(m)))
			Expect(err).ToNot(HaveOccurred())
			api.Handler().ServeHTTP(rec, req)
		}

		It("returns object with 201 created", func() {
			post(payload)
			Expect(rec.Code).To(Equal(http.StatusCreated))
			var actual SomeData
			err := jsonapi.Unmarshal(rec.Body.Bytes(), &actual)
			Expect(err).ToNot(HaveOccurred())
			Expect(payloadID).To(Equal(actual))
		})

		It("return no content 204 with client side generated id", func() {
			post(payloadID)
			Expect(rec.Code).To(Equal(http.StatusNoContent))
			Expect(rec.Body.String()).To(BeEmpty())
		})

		It("return accepted and no content", func() {
			post(SomeData{ID: "accept", Data: "nothing"})
			Expect(rec.Code).To(Equal(http.StatusAccepted))
			Expect(rec.Body.String()).To(BeEmpty())
		})

		It("does not accept invalid return codes", func() {
			post(SomeData{ID: "invalid"})
			Expect(rec.Code).To(Equal(http.StatusInternalServerError))
			var err HTTPError
			json.Unmarshal(rec.Body.Bytes(), &err)
			Expect(err.Errors[0]).To(Equal(Error{
				Title:  "invalid status code 418 from resource someDatas for method Create",
				Status: strconv.Itoa(http.StatusInternalServerError)}))
		})

		It("handles forbidden 403 error", func() {
			post(SomeData{ID: "forbidden", Data: "i am so forbidden"})
			Expect(rec.Code).To(Equal(http.StatusForbidden))
			var err HTTPError
			json.Unmarshal(rec.Body.Bytes(), &err)
			Expect(err.Errors[0]).To(Equal(Error{Title: "Forbidden", Status: strconv.Itoa(http.StatusForbidden)}))
		})

		It("handles 409 conflict error", func() {
			post(SomeData{ID: "conflict", Data: "no force push here"})
			Expect(rec.Code).To(Equal(http.StatusConflict))
			var err HTTPError
			json.Unmarshal(rec.Body.Bytes(), &err)
			Expect(err.Errors[0]).To(Equal(Error{Title: "Conflict", Status: strconv.Itoa(http.StatusConflict)}))
		})
	})

	Context("Update", func() {
		patch := func(payload SomeData) {
			m, err := jsonapi.Marshal(payload)
			Expect(err).ToNot(HaveOccurred())
			req, err := http.NewRequest("PATCH", "/v1/someDatas/12345", strings.NewReader(string(m)))
			Expect(err).ToNot(HaveOccurred())
			api.Handler().ServeHTTP(rec, req)
		}

		It("returns 200 ok if the server modified a field", func() {
			patch(SomeData{ID: "12345", Data: "override me"})
			Expect(rec.Code).To(Equal(http.StatusOK))
			var actual SomeData
			err := jsonapi.Unmarshal(rec.Body.Bytes(), &actual)
			Expect(err).ToNot(HaveOccurred())
			Expect(payloadID).To(Equal(actual))
		})

		It("returns 202 Accepted if update is delayed", func() {
			patch(SomeData{ID: "12345", Data: "delayed"})
			Expect(rec.Code).To(Equal(http.StatusAccepted))
			Expect(rec.Body.String()).To(BeEmpty())
		})

		It("returns 204 No Content if update was accepted", func() {
			patch(SomeData{ID: "12345", Data: "new value"})
			Expect(rec.Code).To(Equal(http.StatusNoContent))
			Expect(rec.Body.String()).To(BeEmpty())
		})

		It("does not accept invalid return codes", func() {
			patch(SomeData{ID: "12345", Data: "invalid"})
			Expect(rec.Code).To(Equal(http.StatusInternalServerError))
			var err HTTPError
			json.Unmarshal(rec.Body.Bytes(), &err)
			Expect(err.Errors[0]).To(Equal(Error{
				Title:  "invalid status code 418 from resource someDatas for method Update",
				Status: strconv.Itoa(http.StatusInternalServerError)}))
		})

		// We do not check everything again like in Create, because it's always the same handleError
		// method that get's called.
		It("handles error cases", func() {
			patch(SomeData{ID: "12345", Data: "fail"})
			Expect(rec.Code).To(Equal(http.StatusForbidden), "we do not allow failes here!")
			var err HTTPError
			json.Unmarshal(rec.Body.Bytes(), &err)
			Expect(err.Errors[0]).To(Equal(Error{Title: "Fail", Status: strconv.Itoa(http.StatusForbidden)}))
		})

	})

	Context("Delete", func() {
		delete := func(ID string) {
			req, err := http.NewRequest("DELETE", "/v1/someDatas/"+ID, nil)
			Expect(err).ToNot(HaveOccurred())
			api.Handler().ServeHTTP(rec, req)
		}

		It("returns 200 ok if there is some meta data", func() {
			delete("200")
			Expect(rec.Code).To(Equal(http.StatusOK))
			var response map[string]interface{}
			json.Unmarshal(rec.Body.Bytes(), &response)
			Expect(response).To(Equal(map[string]interface{}{
				"meta": map[string]interface{}{
					"some": "cool stuff",
				},
			}))
		})

		It("returns 202 accepted if deletion is delayed", func() {
			delete("202")
			Expect(rec.Code).To(Equal(http.StatusAccepted))
			Expect(rec.Body.String()).To(BeEmpty())
		})

		It("return 204 No Content if deletion just worked", func() {
			delete("204")
			Expect(rec.Code).To(Equal(http.StatusNoContent))
			Expect(rec.Body.String()).To(BeEmpty())
		})
	})
})
