package api2go

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type Queue struct {
	ID   string `json:"-"`
	Name string `json:"name"`
}

func (p *Queue) SetID(s string) error {
	p.ID = s
	return nil
}

func (p Queue) GetID() string {
	return p.ID
}

type PrintJob struct {
	ID   string `json:"-"`
	Name string
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

	//TODO content location header
	/*
	 *HTTP/1.1 202 Accepted
	 *Content-Type: application/vnd.api+json
	 *Content-Location: https://example.com/photos/queue-jobs/5234
	 *
	 *{
	 *  "data": {
	 *    "type": "queue-jobs",
	 *    "id": "5234",
	 *    "attributes": {
	 *      "status": "Pending request, waiting other process"
	 *    },
	 *    "links": {
	 *      "self": "/photos/queue-jobs/5234"
	 *    }
	 *  }
	 *}
	 */
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

func (q queueSource) FindOne(id string, req Request) (Responder, error) {
	return &Response{}, nil
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

func (js jobSource) FindOne(id string, req Request) (Responder, error) {
	return &Response{}, nil
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
			}	
			`
			req, err := http.NewRequest("POST", "/v1/queues", strings.NewReader(postRequest))
			Expect(err).To(BeNil())
			api.Handler().ServeHTTP(rec, req)
			Expect(rec.Code).To(Equal(http.StatusAccepted))
		})
	})
})
