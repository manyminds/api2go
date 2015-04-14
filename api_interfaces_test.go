package api2go

import (
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type SomeData struct {
	ID   string
	Data string
}

func (s SomeData) GetID() string {
	return s.ID
}

func (s *SomeData) SetID(ID string) error {
	s.ID = ID
	return nil
}

type SomeResource struct{}

func (s SomeResource) FindOne(ID string, req Request) (interface{}, error) {
	return SomeData{ID: "12345", Data: "A Brezzn"}, nil
}

func (s SomeResource) Create(obj interface{}, req Request) (string, error) {
	return "newID", nil
}

func (s SomeResource) Delete(ID string, req Request) error {
	return nil
}

func (s SomeResource) Update(obj interface{}, req Request) error {
	return nil
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
})
