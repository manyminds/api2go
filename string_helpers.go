package api2go

import (
	"bytes"
	"regexp"

	"github.com/gedex/inflector"
)

var (
	underscorizeAbbreveationsRegex = regexp.MustCompile("([A-Z]+)([A-Z][a-z])")
	underscorizeCamelCaseRegex     = regexp.MustCompile("([a-z\\d])([A-Z])")
)

// Underscorize takes a camel-cased word and transforms it to a underscored version
func Underscorize(word string) string {
	word = underscorizeAbbreveationsRegex.ReplaceAllString(word, "${1}_${2}")
	word = underscorizeCamelCaseRegex.ReplaceAllString(word, "${1}_${2}")
	return string(bytes.ToLower([]byte(word)))
}

// Pluralize a noun
func Pluralize(word string) string {
	return inflector.Pluralize(word)
}
