package api2go

import "strconv"

type httpError struct {
	err    error
	msg    string
	status int
	errors []APIError
}

//APIError can be used for
type APIError struct {
	ID     uint64
	Href   string
	Status string
	Code   string
	Title  string
	Detail string
	Path   string
}

// NewHTTPError creates a new error with message and status code.
// `err` will be logged (but never sent to a client), `msg` will be sent and `status` is the http status code.
// `err` can be nil.
func NewHTTPError(err error, msg string, status int) error {
	var errors []APIError
	return httpError{err, msg, status, errors}
}

//AddAPIError adds an additional json api error
func (e *httpError) AddAPIError(err APIError) {
	e.errors = append(e.errors, err)
}

//Error returns a nice string represenation including the status
func (e httpError) Error() string {
	msg := "http error (" + strconv.Itoa(e.status) + "): " + e.msg
	if e.err != nil {
		msg += ", " + e.err.Error()
	}
	return msg
}
