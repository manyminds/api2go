package api2go

import (
	"database/sql"
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
	if obj.Kind() == reflect.Ptr {
		obj = obj.Elem()
	}
	idField := obj.FieldByName("ID")
	if !idField.IsValid() {
		return "", errors.New("expected 'ID' field in struct")
	}
	return idFromValue(idField)
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
		intID, err := strconv.ParseInt(id, 10, 0)
		if err != nil {
			return err
		}
		val.Set(reflect.ValueOf(int(intID)))

	case reflect.Int8:
		intID, err := strconv.ParseInt(id, 10, 8)
		if err != nil {
			return err
		}
		val.Set(reflect.ValueOf(int8(intID)))

	case reflect.Int16:
		intID, err := strconv.ParseInt(id, 10, 16)
		if err != nil {
			return err
		}
		val.Set(reflect.ValueOf(int16(intID)))

	case reflect.Int32:
		intID, err := strconv.ParseInt(id, 10, 32)
		if err != nil {
			return err
		}
		val.Set(reflect.ValueOf(int32(intID)))

	case reflect.Int64:
		intID, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			return err
		}
		val.Set(reflect.ValueOf(intID))

	case reflect.Uint:
		intID, err := strconv.ParseInt(id, 10, 0)
		if err != nil {
			return err
		}
		val.Set(reflect.ValueOf(uint(intID)))

	case reflect.Uint8:
		intID, err := strconv.ParseUint(id, 10, 8)
		if err != nil {
			return err
		}
		val.Set(reflect.ValueOf(uint8(intID)))

	case reflect.Uint16:
		intID, err := strconv.ParseUint(id, 10, 16)
		if err != nil {
			return err
		}
		val.Set(reflect.ValueOf(uint16(intID)))

	case reflect.Uint32:
		intID, err := strconv.ParseUint(id, 10, 32)
		if err != nil {
			return err
		}
		val.Set(reflect.ValueOf(uint32(intID)))

	case reflect.Uint64:
		intID, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			return err
		}
		val.Set(reflect.ValueOf(uint64(intID)))

	default:
		return errors.New("expected ID to be of type int or string in struct")
	}

	return nil
}
