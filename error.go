package api2go

import "strconv"

type httpError struct {
	err    error
	msg    string
	status int
}

// NewHTTPError creates a new error with message and status code.
// `err` will be logged (but never sent to a client), `msg` will be sent and `status` is the http status code.
// `err` can be nil.
func NewHTTPError(err error, msg string, status int) error {
	return httpError{err, msg, status}
}

func (e httpError) Error() string {
	msg := "http error (" + strconv.Itoa(e.status) + "): " + e.msg
	if e.err != nil {
		msg += ", " + e.err.Error()
	}
	return msg
}
