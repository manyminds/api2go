package api2go

import (
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
