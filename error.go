package api2go

import (
	"fmt"
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
	ID     string       `json:"id,omitempty"`
	Links  *ErrorLinks  `json:"links,omitempty"`
	Status string       `json:"status,omitempty"`
	Code   string       `json:"code,omitempty"`
	Title  string       `json:"title,omitempty"`
	Detail string       `json:"detail,omitempty"`
	Source *ErrorSource `json:"source,omitempty"`
	Meta   interface{}  `json:"meta,omitempty"`
}

// GetID returns the ID
func (e Error) GetID() string {
	return e.ID
}

//ErrorLinks is used to provide an About URL that leads to
//further details about the particular occurrence of the problem.
//
//for more information see http://jsonapi.org/format/#error-objects
type ErrorLinks struct {
	About string `json:"about,omitempty"`
}

//ErrorSource is used to provide references to the source of an error.
//
//The Pointer is a JSON Pointer to the associated entity in the request
//document.
//The Paramter is a string indicating which query parameter caused the error.
//
//for more information see http://jsonapi.org/format/#error-objects
type ErrorSource struct {
	Pointer   string `json:"pointer,omitempty"`
	Parameter string `json:"parameter,omitempty"`
}

//MarshalError marshals errors recursively in json format.
//it can make use of the jsonapi.HTTPError struct
func (j JSONContentMarshaler) MarshalError(err error) (int, []byte) {
	httpErr, ok := err.(HTTPError)
	if ok {
		return httpErr.status, marshalHTTPError(httpErr, j)
	}

	httpErr = NewHTTPError(err, err.Error(), 500)

	return httpErr.status, marshalHTTPError(httpErr, j)
}

//marshalHTTPError marshals an internal httpError
func marshalHTTPError(input HTTPError, marshaler ContentMarshaler) []byte {
	if len(input.Errors) == 0 {
		code := ""
		if input.err != nil {
			code = input.err.Error()
		}
		input.Errors = []Error{Error{Code: code, Title: input.msg, Status: strconv.Itoa(input.status)}}
	}

	data, err := marshaler.Marshal(input)

	if err != nil {
		return []byte("{}")
	}

	return data
}

// NewHTTPError creates a new error with message and status code.
// `err` will be sent as 'code', `msg` as 'title' and `status` is the http status code.
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
