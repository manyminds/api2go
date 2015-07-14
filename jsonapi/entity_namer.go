package jsonapi

// The EntityNamer interface can be opionally implemented to rename a struct. The name returned by
// GetName will be used for the route generation as well as the "type" field in all responses
type EntityNamer interface {
	GetName() string
}
