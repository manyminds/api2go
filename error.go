package api2go

import (
	"fmt"
	"log"
	"strconv"
)

//HTTPError is used for errors
type HTTPError struct {
	err    error
	msg    string
	status int
	Errors []Error `json:"errors,omitempty"`
}

//Error can be used for all kind of application errors
//e.g. you would use it to define form errors or any
//other semantical application problems
//for more information see http://jsonapi.org/format/#errors
type Error struct {
	ID     string `json:"id,omitempty"`
	Href   string `json:"href,omitempty"`
	Status string `json:"status,omitempty"`
	Code   string `json:"code,omitempty"`
	Title  string `json:"title,omitempty"`
	Detail string `json:"detail,omitempty"`
	Path   string `json:"path,omitempty"`
}

// GetID returns the ID
func (e Error) GetID() string {
	return e.ID
}

//marshalError marshals all error types
func marshalError(err error, marshaler ContentMarshaler) string {
	httpErr, ok := err.(HTTPError)
	if ok {
		return marshalHTTPError(httpErr, marshaler)
	}

	httpErr = NewHTTPError(err, err.Error(), 500)

	return marshalHTTPError(httpErr, marshaler)
}

//marshalHTTPError marshals an internal httpError
func marshalHTTPError(input HTTPError, marshaler ContentMarshaler) string {
	if len(input.Errors) == 0 {
		input.Errors = []Error{Error{Title: input.msg, Status: strconv.Itoa(input.status)}}
	}

	data, err := marshaler.Marshal(input)

	if err != nil {
		log.Println(err)
		return "{}"
	}

	return string(data)
}

// NewHTTPError creates a new error with message and status code.
// `err` will be logged (but never sent to a client), `msg` will be sent and `status` is the http status code.
// `err` can be nil.
func NewHTTPError(err error, msg string, status int) HTTPError {
	return HTTPError{err: err, msg: msg, status: status}
}

//Error returns a nice string represenation including the status
func (e HTTPError) Error() string {
	msg := fmt.Sprintf("http error (%d) %s and %d more errors", e.status, e.msg, len(e.Errors))
	if e.err != nil {
		msg += ", " + e.err.Error()
	}

	return msg
}
