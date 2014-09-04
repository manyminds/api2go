package api2go

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
	"unicode"

	"github.com/gedex/inflector"
)

// commonInitialisms, taken from
// https://github.com/golang/lint/blob/3d26dc39376c307203d3a221bada26816b3073cf/lint.go#L482
var commonInitialisms = map[string]bool{
	"API":   true,
	"ASCII": true,
	"CPU":   true,
	"CSS":   true,
	"DNS":   true,
	"EOF":   true,
	"GUID":  true,
	"HTML":  true,
	"HTTP":  true,
	"HTTPS": true,
	"ID":    true,
	"IP":    true,
	"JSON":  true,
	"LHS":   true,
	"QPS":   true,
	"RAM":   true,
	"RHS":   true,
	"RPC":   true,
	"SLA":   true,
	"SMTP":  true,
	"SSH":   true,
	"TLS":   true,
	"TTL":   true,
	"UI":    true,
	"UID":   true,
	"UUID":  true,
	"URI":   true,
	"URL":   true,
	"UTF8":  true,
	"VM":    true,
	"XML":   true,
}

// dejsonify returns a go struct key name from a JSON key name
func dejsonify(s string) string {
	if s == "" {
		return ""
	}
	if upper := strings.ToUpper(s); commonInitialisms[upper] {
		return upper
	}
	rs := []rune(s)
	rs[0] = unicode.ToUpper(rs[0])
	return string(rs)
}

// jsonify returns a JSON formatted key name from a go struct field name
func jsonify(s string) string {
	if s == "" {
		return ""
	}
	if commonInitialisms[s] {
		return strings.ToLower(s)
	}
	rs := []rune(s)
	rs[0] = unicode.ToLower(rs[0])
	return string(rs)
}

// pluralize a noun
func pluralize(word string) string {
	return inflector.Pluralize(word)
}

// singularize a noun
func singularize(word string) string {
	return inflector.Singularize(word)
}

func idFromObject(obj reflect.Value) (string, error) {
	idField := obj.FieldByName("ID")
	if !idField.IsValid() {
		return "", errors.New("expected 'ID' field in struct")
	}
	return idFromValue(idField)
}

func idFromValue(v reflect.Value) (string, error) {
	switch v.Kind() {
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

func setObjectID(obj reflect.Value, idInterface interface{}) error {
	field := obj.FieldByName("ID")
	if !field.IsValid() {
		return errors.New("expected struct to have field 'ID'")
	}
	return setIDValue(field, idInterface)
}

func setIDValue(val reflect.Value, idInterface interface{}) error {
	id, ok := idInterface.(string)
	if !ok {
		return errors.New("expected ID to be string in json")
	}
	switch val.Kind() {
	case reflect.String:
		val.Set(reflect.ValueOf(id))

	case reflect.Int:
		intID, err := strconv.Atoi(id)
		if err != nil {
			return err
		}
		val.Set(reflect.ValueOf(intID))

	default:
		return errors.New("expected ID to be of type int or string in struct")
	}

	return nil
}
