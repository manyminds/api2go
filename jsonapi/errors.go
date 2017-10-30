package jsonapi

import "errors"

//Error for all errors within this package
type Error interface {
	error
}

//MarshalError interface
type MarshalError interface {
	Error
}

//UnmarshalError interface
type UnmarshalError interface {
	Error
}

// Marshal errors
var (
	// ErrInvalidMarshalType is returned if marshal is called with an invalid type
	ErrInvalidMarshalType MarshalError = errors.New("Marshal only accepts slice, struct or ptr types")
	// ErrNonHomogenousSlice if any element within a slice does not implement marshal identifier
	ErrNonHomogenousSlice MarshalError = errors.New("all elements within the slice must implement api2go.MarshalIdentifier")
	//ErrNonNilElement marshaling a nil type is not possible
	ErrNonNilElement MarshalError = errors.New("MarshalIdentifier must not be nil")
)

// Unmarshal errors
var (
	// ErrNilTarget if element will be unmarshalled into a nil element
	ErrNilTarget UnmarshalError = errors.New("target must not be nil")
	// ErrNonPointerTarget
	ErrNonPointerTarget UnmarshalError = errors.New("target must be a ptr")
)

// Semantic errors
var (
	ErrInvalidUnmarshalType UnmarshalError = errors.New("target must implement UnmarshalIdentifier interface")
	ErrEmptySource          UnmarshalError = errors.New(`Source JSON is empty and has no "attributes" payload object`)
	ErrMissingType          UnmarshalError = errors.New("invalid record, no type was specified")
	ErrInvalidStruct        UnmarshalError = errors.New("existing structs must implement interface MarshalIdentifier")
)
