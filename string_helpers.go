package api2go

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/gedex/inflector"
)

var (
	underscorizeAbbreveationsRegex = regexp.MustCompile("([A-Z]+)([A-Z][a-z])")
	underscorizeCamelCaseRegex     = regexp.MustCompile("([a-z\\d])([A-Z])")
	camelizeRegex                  = regexp.MustCompile(`_[a-z\d]`)
	camelizeAbbrevs = []string{
		"xml",
		"id",
		"json",
	}
)

// underscorize takes a camel-cased word and transforms it to a underscored version
func underscorize(word string) string {
	word = underscorizeAbbreveationsRegex.ReplaceAllString(word, "${1}_${2}")
	word = underscorizeCamelCaseRegex.ReplaceAllString(word, "${1}_${2}")
	return strings.ToLower(word)
}

//camelize takes a underscored word and transforms it to a camel-cased version
func camelize(word string) string {
	if word == "" {
		return ""
	}
	// Special abbreviations
	for _, v := range camelizeAbbrevs {
		word = strings.Replace(word, v, strings.ToUpper(v), -1)
	}
	// Capitalize first char
	rs := []rune(word)
	rs[0] = unicode.ToUpper(rs[0])
	word = string(rs)
	// Replace rest
	word = camelizeRegex.ReplaceAllStringFunc(word, func(w string) string {
		return strings.ToUpper(w[1:])
	})
	return word
}

// pluralize a noun
func pluralize(word string) string {
	return inflector.Pluralize(word)
}

// singularize a noun
func singularize(word string) string {
	return inflector.Singularize(word)
}
