package jsonapi

// The RouteNamer interface can be optionally implemented to directly return the
// name of route used for the "type" field.
//
// Note: By default the name is guessed from the struct name or from EntityNamer.
type RouteNamer interface {
	GetRouteName() string
}
