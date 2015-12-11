package api2go

import (
	"net/http"
	"net/http/httptest"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type BaguetteTaste struct {
	ID    string `json:"-"`
	Taste string `json:"taste"`
}

func (s BaguetteTaste) GetID() string {
	return s.ID
}

func (s *BaguetteTaste) SetID(ID string) error {
	s.ID = ID
	return nil
}

func (s BaguetteTaste) GetName() string {
	return "baguette-tastes"
}

type BaguetteResource struct{}

func (s BaguetteResource) FindOne(ID string, req Request) (Responder, error) {
	return &Response{Res: BaguetteTaste{ID: "blubb", Taste: "Very Bad"}}, nil
}

func (s BaguetteResource) FindAll(req Request) (Responder, error) {
	return &Response{Res: []BaguetteTaste{
		{
			ID:    "1",
			Taste: "Very Good",
		},
		{
			ID:    "2",
			Taste: "Very Bad",
		},
	}}, nil
}

func (s BaguetteResource) Create(obj interface{}, req Request) (Responder, error) {
	e := obj.(BaguetteTaste)
	e.ID = "newID"
	return &Response{
		Res:  e,
		Code: http.StatusCreated,
	}, nil
}

func (s BaguetteResource) Delete(ID string, req Request) (Responder, error) {
	return &Response{
		Res:  BaguetteTaste{ID: ID},
		Code: http.StatusNoContent,
	}, nil
}

func (s BaguetteResource) Update(obj interface{}, req Request) (Responder, error) {
	return &Response{
		Res:  obj,
		Code: http.StatusNoContent,
	}, nil
}

var _ = Describe("Test route renaming with EntityNamer interface", func() {
	var (
		api  *API
		rec  *httptest.ResponseRecorder
		body *strings.Reader
	)
	BeforeEach(func() {
		api = NewAPI("v1")
		api.AddResource(BaguetteTaste{}, BaguetteResource{})
		rec = httptest.NewRecorder()
		body = strings.NewReader(`
		{
			"data": {
				"attributes": {
					"taste": "smells awful"
				},
				"id": "blubb",
				"type": "baguette-tastes"
			}
		}
		`)
	})

	// check that renaming works, we do not test every single route here, the name variable is used
	// for each route, we just check the 5 basic ones. Marshalling and Unmarshalling is tested with
	// this again too.
	It("FindAll returns 200", func() {
		req, err := http.NewRequest("GET", "/v1/baguette-tastes", nil)
		Expect(err).ToNot(HaveOccurred())
		api.Handler().ServeHTTP(rec, req)
		Expect(rec.Code).To(Equal(http.StatusOK))
	})

	It("FindOne", func() {
		req, err := http.NewRequest("GET", "/v1/baguette-tastes/12345", nil)
		Expect(err).ToNot(HaveOccurred())
		api.Handler().ServeHTTP(rec, req)
		Expect(rec.Code).To(Equal(http.StatusOK))
	})

	It("Delete", func() {
		req, err := http.NewRequest("DELETE", "/v1/baguette-tastes/12345", nil)
		Expect(err).ToNot(HaveOccurred())
		api.Handler().ServeHTTP(rec, req)
		Expect(rec.Code).To(Equal(http.StatusNoContent))
	})

	It("Create", func() {
		req, err := http.NewRequest("POST", "/v1/baguette-tastes", body)
		Expect(err).ToNot(HaveOccurred())
		api.Handler().ServeHTTP(rec, req)
		// the response is always the one record returned by FindOne, the implementation does not
		// check the ID here and returns something new ...
		Expect(rec.Body.String()).To(MatchJSON(`
		{
			"data": {
				"attributes": {
					"taste": "smells awful"
				},
				"id": "newID",
				"type": "baguette-tastes"
			}
		}
		`))
		Expect(rec.Code).To(Equal(http.StatusCreated))
	})

	It("Update", func() {
		req, err := http.NewRequest("PATCH", "/v1/baguette-tastes/blubb", body)
		Expect(err).ToNot(HaveOccurred())
		api.Handler().ServeHTTP(rec, req)
		Expect(rec.Body.String()).To(Equal(""))
		Expect(rec.Code).To(Equal(http.StatusNoContent))
	})
})
