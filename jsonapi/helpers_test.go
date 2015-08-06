package jsonapi

import (
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("StringHelpers", func() {
	Context("json funcs", func() {
		It("dejsonifies", func() {
			Expect(Dejsonify("Post")).To(Equal("Post"))
			Expect(Dejsonify("post")).To(Equal("Post"))
			Expect(Dejsonify("id")).To(Equal("ID"))
			Expect(Dejsonify("")).To(Equal(""))
		})

		It("jsonifies", func() {
			Expect(Jsonify("Post")).To(Equal("post"))
			Expect(Jsonify("post")).To(Equal("post"))
			Expect(Jsonify("ID")).To(Equal("id"))
			Expect(Jsonify("")).To(Equal(""))
		})

		It("Pluralizes", func() {
			Expect(Pluralize("post")).To(Equal("posts"))
			Expect(Pluralize("posts")).To(Equal("posts"))
			Expect(Pluralize("category")).To(Equal("categories"))
		})

		It("singularizes", func() {
			Expect(Singularize("posts")).To(Equal("post"))
			Expect(Singularize("post")).To(Equal("post"))
			Expect(Singularize("categories")).To(Equal("category"))
		})
	})

	Context("Reflect funcs", func() {
		type Element struct {
			Name string `jsonapi:"name=actress;body=hot;chest=awesome;character"`
		}

		It("tests for existance of settings", func() {
			element := Element{Name: "Scarlett Johansson"}
			testField := reflect.ValueOf(element).Type().Field(0)
			Expect(GetTagValueByName(testField, "name")).To(Equal("actress"))
			Expect(GetTagValueByName(testField, "body")).To(Equal("hot"))
			Expect(GetTagValueByName(testField, "chest")).To(Equal("awesome"))
			Expect(GetTagValueByName(testField, "character")).To(Equal("character"))
		})

		It("tests for non existing settings", func() {
			element := Element{Name: "Jennifer Lawrence"}
			testField := reflect.ValueOf(element).Type().Field(0)
			Expect(GetTagValueByName(testField, "talent")).To(Equal(""))
		})
	})
})
