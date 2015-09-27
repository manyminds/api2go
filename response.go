package api2go

// The Response struct implements api2go.Responder and can be used as a default
// implementation for your responses
// you can fill the field `Meta` with all the metadata your application needs
// like license, tokens, etc
type Response struct {
	Res  interface{}
	Code int
	Meta map[string]interface{}
}

// Metadata returns additional meta data
func (r Response) Metadata() map[string]interface{} {
	return r.Meta
}

// Result returns the actual payload
func (r Response) Result() interface{} {
	return r.Res
}

// StatusCode sets the return status code
func (r Response) StatusCode() int {
	return r.Code
}
