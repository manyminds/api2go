package jsonapi

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

// ConventionWrapper restores api2go compatibilty
// to the old version.
// simply let your struct be wrapped around
// ConventionWrapper and enjoy the old lazy
// way to use api2go.
// Deprecated will be removed in further
// releases. Its strongly advised to invest
// the time to implement api2gos interfaces.
type ConventionWrapper struct {
	Item interface{}
}

const (
	// NoIDFieldPresent will be returned by GetID if nothing could be found
	NoIDFieldPresent = "NoIDFieldPresent"
)

// GetID will search your struct via reflection
// and return the value within an field called
// ID, this field must be one of the types:
// int* uint* string
// will return -1 if the field was not found
func (c ConventionWrapper) GetID() string {
	val := reflect.ValueOf(c.Item)
	valType := val.Type()

	for i := 0; i < val.NumField(); i++ {
		tag := valType.Field(i).Tag.Get("json")
		if tag == "-" {
			continue
		}

		field := val.Field(i)
		keyName := Jsonify(valType.Field(i).Name)

		if keyName == "id" {
			id, _ := idFromValue(field)
			return id
		}
	}

	return NoIDFieldPresent
}

// GetReferences will dynamically search for TODO
func (c ConventionWrapper) GetReferences() []Reference {

	return []Reference{}
}

// GetReferencedIDs will dynamicaly search for TODO
func (c ConventionWrapper) GetReferencedIDs() []ReferenceID {

	return []ReferenceID{}
}

// GetReferencedStructs will dynamically search for TODO
func (c ConventionWrapper) GetReferencedStructs() []MarshalIdentifier {
	return []MarshalIdentifier{}
}

func idFromValue(v reflect.Value) (string, error) {
	kind := v.Kind()
	if kind == reflect.Struct {
		if sv, err := extractIDFromSqlStruct(v); err == nil {
			v = sv
			kind = v.Kind()
		} else {
			return "", err
		}
	} else if v.CanInterface() {
		x := v.Interface()

		switch x := x.(type) {
		case fmt.Stringer:
			return x.String(), nil
		}
	}

	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(v.Uint(), 10), nil
	case reflect.String:
		return v.String(), nil
	default:
		return "", errors.New("need int or string as type of ID")
	}
}

func extractIDFromSqlStruct(v reflect.Value) (reflect.Value, error) {
	i := v.Interface()
	switch value := i.(type) {
	case sql.NullInt64:
		if value.Valid {
			return reflect.ValueOf(value.Int64), nil
		}
	case sql.NullFloat64:
		if value.Valid {
			return reflect.ValueOf(value.Float64), nil
		}
	case sql.NullString:
		if value.Valid {
			return reflect.ValueOf(value.String), nil
		}
	default:
		return reflect.ValueOf(""), errors.New("invalid type, allowed sql/database types are sql.NullInt64, sql.NullFloat64, sql.NullString")
	}

	return reflect.ValueOf(""), nil
}
