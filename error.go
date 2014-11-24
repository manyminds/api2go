package api2go

import "strconv"

type httpError struct {
	err         error
	msg         string
	status      int
	errors      []httpError
	errorsCount int
}

// NewHTTPError creates a new error with message and status code.
// `err` will be logged (but never sent to a client), `msg` will be sent and `status` is the http status code.
// `err` can be nil.
func NewHTTPError(err error, msg string, status int) error {
	var errors []httpError
	return httpError{err, msg, status, errors, 0}
}

//AddHTTPError adds an additional error to this http error
func (e *httpError) AddHTTPError(err httpError) {
	if e.errors == nil {
		e.errors = make([]httpError, 0, 10)
	}

	e.errors = append(e.errors, err)
	e.errorsCount++
}

func (e httpError) Error() string {
	msg := "http error (" + strconv.Itoa(e.status) + "): " + e.msg
	if e.err != nil {
		msg += ", " + e.err.Error()
	}
	return msg
}
