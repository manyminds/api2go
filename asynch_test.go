package api2go

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/manyminds/api2go/jsonapi"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type Queue struct {
	ID           string `json:"-"`
	Name         string `json:"name"`
	Status       string `json:"status"`
	requestCount int
}

type seeOtherResponse struct {
	Response
	OtherObject jsonapi.MarshalIdentifier
}

func (s seeOtherResponse) Other() jsonapi.MarshalIdentifier {
	return s.OtherObject
}

func (p *Queue) SetID(s string) error {
	p.ID = s
	return nil
}

func (p Queue) GetID() string {
	return p.ID
}

type PrintJob struct {
	ID      string `json:"-"`
	JobName string `json:"name"`
}

func (p PrintJob) GetName() string {
	return "jobs"
}

func (p *PrintJob) SetID(s string) error {
	p.ID = s
	return nil
}

func (p PrintJob) GetID() string {
	return p.ID
}

type queueSource struct {
	queue map[string]*Queue
}

var queueID = 0
var jobID = 0

func (q *queueSource) Create(obj interface{}, _ Request) (Responder, error) {
	pq, ok := obj.(*Queue)
	if !ok {
		return nil, errors.New("invalid object given")
	}

	queueID++
	pq.ID = fmt.Sprintf("%d", queueID)
	q.queue[pq.ID] = pq
	pq.Status = "Pending request, waiting for process to finish"
	//TODO content location header
	return &Response{Res: pq, Code: http.StatusAccepted}, nil
}

func (q queueSource) Delete(id string, _ Request) (Responder, error) {
	return &Response{Code: 500}, nil
}

func (q queueSource) Update(obj interface{}, _ Request) (Responder, error) {
	return &Response{Code: 500}, nil
}

func (q queueSource) FindAll(_ Request) (Responder, error) {
	return &Response{Code: 500}, nil
}

func (q *queueSource) FindOne(id string, req Request) (Responder, error) {
	if id == "1" {
		item, ok := q.queue[id]
		if !ok {
			return &Response{Code: 404}, nil
		}
		item.requestCount++
		if item.requestCount > 1 {
			return &seeOtherResponse{Response{Code: 303}, PrintJob{ID: "1"}}, nil
		}

		item.Status = "Pending request, waiting for process to finish"

		return &Response{Res: item, Code: http.StatusAccepted}, nil
	}

	return &Response{Code: 404}, nil
}

type jobSource struct {
	jobs map[string]*PrintJob
}

func (js *jobSource) Create(obj interface{}, _ Request) (Responder, error) {
	pq, ok := obj.(PrintJob)
	if !ok {
		return nil, errors.New("invalid object given")
	}

	queueID++
	pq.ID = fmt.Sprintf("%d", queueID)
	js.jobs[pq.ID] = &pq
	return &Response{Res: pq, Code: http.StatusAccepted}, nil
}

func (js jobSource) Delete(id string, _ Request) (Responder, error) {
	return &Response{Code: 500}, nil
}

func (js jobSource) Update(obj interface{}, _ Request) (Responder, error) {
	return &Response{Code: 500}, nil
}

func (js jobSource) FindAll(_ Request) (Responder, error) {
	return &Response{Code: 500}, nil
}

func (js *jobSource) FindOne(id string, req Request) (Responder, error) {
	return &Response{Code: 404}, nil
}

var _ = Describe("AsynchTest", func() {
	var (
		qs  *queueSource
		js  *jobSource
		api *API
		rec *httptest.ResponseRecorder
	)

	BeforeEach(func() {
		api = NewAPIWithBaseURL("v1", "http://localhost")
		rec = httptest.NewRecorder()

		qs = &queueSource{queue: map[string]*Queue{}}
		js = &jobSource{jobs: map[string]*PrintJob{}}
		api.AddResource(&Queue{}, qs)
		api.AddResource(&PrintJob{}, js)
	})

	Context("Test asynchronus handling", func() {
		It("Should process a post ansynchronous", func() {
			postRequest := `
			{
				"data" : 
					{
						"attributes":				
						{
							"name": "invoice_2.pdf"
						},
						"type": "queues"
					}	
			}`

			expected := `{
        "data": {
          "type": "queues",
          "id": "1",
          "attributes": {
            "name": "invoice_2.pdf",
            "status": "Pending request, waiting for process to finish"
          }
        }
      }`

			req, err := http.NewRequest("POST", "/v1/queues", strings.NewReader(postRequest))
			Expect(err).To(BeNil())
			api.Handler().ServeHTTP(rec, req)
			Expect(rec.Code).To(Equal(http.StatusAccepted))
			actual := string(rec.Body.Bytes())
			Expect(actual).ToNot(Equal(""))
			Expect(actual).To(MatchJSON(expected))

			req, err, actual, rec = nil, nil, "", httptest.NewRecorder()
			req, err = http.NewRequest("GET", "/v1/queues/1", nil)
			Expect(err).To(BeNil())
			api.Handler().ServeHTTP(rec, req)
			Expect(rec.Code).To(Equal(http.StatusAccepted))
			actual = string(rec.Body.Bytes())
			Expect(actual).ToNot(Equal(""))
			Expect(actual).To(MatchJSON(expected))

			req, err, actual, rec = nil, nil, "", httptest.NewRecorder()
			req, err = http.NewRequest("GET", "/v1/queues/1", nil)
			Expect(err).To(BeNil())
			api.Handler().ServeHTTP(rec, req)
			Expect(rec.Code).To(Equal(http.StatusSeeOther))
			Expect(rec.Header().Get("Location")).To(Equal("http://localhost/v1/jobs/1"))
		})
	})
})
