package jsonapi

import (
	"reflect"
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
	"JWT":   true,
}

// Dejsonify returns a go struct key name from a JSON key name
func Dejsonify(s string) string {
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

// Jsonify returns a JSON formatted key name from a go struct field name
func Jsonify(s string) string {
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

// Pluralize a noun
func Pluralize(word string) string {
	return inflector.Pluralize(word)
}

// Singularize a noun
func Singularize(word string) string {
	return inflector.Singularize(word)
}

// GetTagValueByName returns one api2go setting.
// settings must be of the format `jsonapi:"name=newName,body=newbody"
func GetTagValueByName(tfield reflect.StructField, name string) string {
	str := tfield.Tag.Get("jsonapi")
	if str == "" {
		return ""
	}

	tags := strings.Split(str, ";")
	setting := map[string]string{}
	for _, value := range tags {
		v := strings.Split(value, "=")
		k := strings.TrimSpace(strings.ToLower(v[0]))
		if len(v) == 2 {
			setting[k] = v[1]
		} else {
			setting[k] = k
		}
	}

	return setting[strings.ToLower(name)]
}
