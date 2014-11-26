package api2go

import "strconv"

//HTTPError is used for errors
type HTTPError struct {
	err    error
	msg    string
	status int
	Errors map[string]Error `json:"errors,omitempty"`
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

// NewHTTPError creates a new error with message and status code.
// `err` will be logged (but never sent to a client), `msg` will be sent and `status` is the http status code.
// `err` can be nil.
func NewHTTPError(err error, msg string, status int) HTTPError {
	return HTTPError{err: err, msg: msg, status: status, Errors: make(map[string]Error)}
}

//Error returns a nice string represenation including the status
func (e HTTPError) Error() string {
	msg := "http error (" + strconv.Itoa(e.status) + "): " + e.msg
	if e.err != nil {
		msg += ", " + e.err.Error()
	}

	return msg
}
