package main_test

import (
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/manyminds/api2go"
	"github.com/manyminds/api2go/examples/model"
	"github.com/manyminds/api2go/examples/resource"
	"github.com/manyminds/api2go/examples/storage"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// there are a lot of functions because each test can be run individually and sets up the complete
// environment. That is because we run all the specs randomized.
var _ = Describe("CrudExample", func() {
	var rec *httptest.ResponseRecorder

	BeforeEach(func() {
		api = api2go.NewAPIWithBaseURL("v0", "http://localhost:31415")
		userStorage := storage.NewUserStorage()
		chocStorage := storage.NewChocolateStorage()
		api.AddResource(model.User{}, resource.UserResource{ChocStorage: chocStorage, UserStorage: userStorage})
		api.AddResource(model.Chocolate{}, resource.ChocolateResource{ChocStorage: chocStorage, UserStorage: userStorage})
		rec = httptest.NewRecorder()
	})

	var createUser = func() {
		rec = httptest.NewRecorder()
		req, err := http.NewRequest("POST", "/v0/users", strings.NewReader(`
		{
			"data": {
				"type": "users",
				"attributes": {
					"user-name": "marvin"
				}
			}
		}
		`))
		Expect(err).ToNot(HaveOccurred())
		api.Handler().ServeHTTP(rec, req)
		Expect(rec.Code).To(Equal(http.StatusCreated))
		Expect(rec.Body.String()).To(MatchJSON(`
		{
			"meta": {
				"author": "The api2go examples crew",
				"license": "wtfpl",
				"license-url": "http://www.wtfpl.net"
			},
			"data": {
				"id": "1",
				"type": "users",
				"attributes": {
					"user-name": "marvin"
				},
				"relationships": {
					"sweets": {
						"data": [],
						"links": {
							"related": "http://localhost:31415/v0/users/1/sweets",
							"self": "http://localhost:31415/v0/users/1/relationships/sweets"
						}
					}
				}
			}
		}
		`))
	}

	It("Creates a new user", func() {
		createUser()
	})

	var createChocolate = func() {
		rec = httptest.NewRecorder()
		req, err := http.NewRequest("POST", "/v0/chocolates", strings.NewReader(`
		{
			"data": {
				"type": "chocolates",
				"attributes": {
					"name": "Ritter Sport",
					"taste": "Very Good"
				}
			}
		}
		`))
		Expect(err).ToNot(HaveOccurred())
		api.Handler().ServeHTTP(rec, req)
		Expect(rec.Code).To(Equal(http.StatusCreated))
		Expect(rec.Body.String()).To(MatchJSON(`
		{
			"meta": {
				"author": "The api2go examples crew",
				"license": "wtfpl",
				"license-url": "http://www.wtfpl.net"
			},
			"data": {
				"id": "1",
				"type": "chocolates",
				"attributes": {
					"name": "Ritter Sport",
					"taste": "Very Good"
				}
			}
		}
		`))
	}

	It("Creates a new chocolate", func() {
		createChocolate()
	})

	var replaceSweets = func() {
		rec = httptest.NewRecorder()
		By("Replacing sweets relationship with PATCH")

		req, err := http.NewRequest("PATCH", "/v0/users/1/relationships/sweets", strings.NewReader(`
		{
			"data": [{
				"type": "chocolates",
				"id": "1"
			}]
		}
		`))
		Expect(err).ToNot(HaveOccurred())
		api.Handler().ServeHTTP(rec, req)
		Expect(rec.Code).To(Equal(http.StatusNoContent))

		By("Loading the user from the backend, it should have the relationship")

		rec = httptest.NewRecorder()
		req, err = http.NewRequest("GET", "/v0/users/1", nil)
		api.Handler().ServeHTTP(rec, req)
		Expect(err).ToNot(HaveOccurred())
		Expect(rec.Body.String()).To(MatchJSON(`
		{
			"meta": {
				"author": "The api2go examples crew",
				"license": "wtfpl",
				"license-url": "http://www.wtfpl.net"
			},
			"data": {
				"attributes": {
					"user-name": "marvin"
				},
				"id": "1",
				"relationships": {
					"sweets": {
						"data": [
							{
								"id": "1",
								"type": "chocolates"
							}
						],
						"links": {
							"related": "http://localhost:31415/v0/users/1/sweets",
							"self": "http://localhost:31415/v0/users/1/relationships/sweets"
						}
					}
				},
				"type": "users"
			},
			"included": [
				{
					"attributes": {
						"name": "Ritter Sport",
						"taste": "Very Good"
					},
					"id": "1",
					"type": "chocolates"
				}
			]
		}
		`))
	}

	It("Replaces users sweets", func() {
		createUser()
		createChocolate()
		replaceSweets()
	})

	It("Deletes a users sweet", func() {
		createUser()
		createChocolate()
		replaceSweets()
		rec = httptest.NewRecorder()

		By("Deleting the users only sweet with ID 1")

		req, err := http.NewRequest("DELETE", "/v0/users/1/relationships/sweets", strings.NewReader(`
		{
			"data": [{
				"type": "chocolates",
				"id": "1"
			}]
		}
		`))
		Expect(err).ToNot(HaveOccurred())
		api.Handler().ServeHTTP(rec, req)
		Expect(rec.Code).To(Equal(http.StatusNoContent))

		By("Loading the user from the backend, it should not have any relations")

		rec = httptest.NewRecorder()
		req, err = http.NewRequest("GET", "/v0/users/1", nil)
		api.Handler().ServeHTTP(rec, req)
		Expect(err).ToNot(HaveOccurred())
		Expect(rec.Body.String()).To(MatchJSON(`
		{
			"meta": {
				"author": "The api2go examples crew",
				"license": "wtfpl",
				"license-url": "http://www.wtfpl.net"
			},
			"data": {
				"attributes": {
					"user-name": "marvin"
				},
				"id": "1",
				"relationships": {
					"sweets": {
						"data": [],
						"links": {
							"related": "http://localhost:31415/v0/users/1/sweets",
							"self": "http://localhost:31415/v0/users/1/relationships/sweets"
						}
					}
				},
				"type": "users"
			}
		}
		`))
	})

	It("Adds a users sweet", func() {
		createUser()
		createChocolate()
		rec = httptest.NewRecorder()

		By("Adding a sweet with POST")

		req, err := http.NewRequest("POST", "/v0/users/1/relationships/sweets", strings.NewReader(`
		{
			"data": [{
				"type": "chocolates",
				"id": "1"
			}]
		}
		`))
		Expect(err).ToNot(HaveOccurred())
		api.Handler().ServeHTTP(rec, req)
		Expect(rec.Code).To(Equal(http.StatusNoContent))

		By("Loading the user from the backend, it should have the relationship")

		rec = httptest.NewRecorder()
		req, err = http.NewRequest("GET", "/v0/users/1", nil)
		api.Handler().ServeHTTP(rec, req)
		Expect(err).ToNot(HaveOccurred())
		Expect(rec.Body.String()).To(MatchJSON(`
		{
			"meta": {
				"author": "The api2go examples crew",
				"license": "wtfpl",
				"license-url": "http://www.wtfpl.net"
			},
			"data": {
				"attributes": {
					"user-name": "marvin"
				},
				"id": "1",
				"relationships": {
					"sweets": {
						"data": [
							{
								"id": "1",
								"type": "chocolates"
							}
						],
						"links": {
							"related": "http://localhost:31415/v0/users/1/sweets",
							"self": "http://localhost:31415/v0/users/1/relationships/sweets"
						}
					}
				},
				"type": "users"
			},
			"included": [
				{
					"attributes": {
						"name": "Ritter Sport",
						"taste": "Very Good"
					},
					"id": "1",
					"type": "chocolates"
				}
			]
		}
		`))
	})

	Describe("Load sweets of a user directly", func() {
		BeforeEach(func() {
			createUser()
			createChocolate()
			replaceSweets()
			rec = httptest.NewRecorder()

			// add another sweet so we have 2, only 1 is connected with the user
			req, err := http.NewRequest("POST", "/v0/chocolates", strings.NewReader(`
			{
				"data": {
					"type": "chocolates",
					"attributes": {
						"name": "Black Chocolate",
						"taste": "Bitter"
					}
				}
			}
			`))
			Expect(err).ToNot(HaveOccurred())
			api.Handler().ServeHTTP(rec, req)
			Expect(rec.Code).To(Equal(http.StatusCreated))
			Expect(rec.Body.String()).To(MatchJSON(`
			{
				"meta": {
					"author": "The api2go examples crew",
					"license": "wtfpl",
					"license-url": "http://www.wtfpl.net"
				},
				"data": {
					"id": "2",
					"type": "chocolates",
					"attributes": {
						"name": "Black Chocolate",
						"taste": "Bitter"
					}
				}
			}
			`))

			rec = httptest.NewRecorder()
		})

		It("There are 2 chocolates in the datastorage now", func() {
			req, err := http.NewRequest("GET", "/v0/chocolates", nil)
			Expect(err).ToNot(HaveOccurred())
			api.Handler().ServeHTTP(rec, req)
			Expect(rec.Code).To(Equal(http.StatusOK))
			Expect(rec.Body.String()).To(MatchJSON(`
			{
				"meta": {
					"author": "The api2go examples crew",
					"license": "wtfpl",
					"license-url": "http://www.wtfpl.net"
				},
				"data": [
					{
						"attributes": {
							"name": "Ritter Sport",
							"taste": "Very Good"
						},
						"id": "1",
						"type": "chocolates"
					},
					{
						"attributes": {
							"name": "Black Chocolate",
							"taste": "Bitter"
						},
						"id": "2",
						"type": "chocolates"
					}
				]
			}
			`))
		})

		It("The user only has the previously connected sweet", func() {
			req, err := http.NewRequest("GET", "/v0/users/1", nil)
			Expect(err).ToNot(HaveOccurred())
			api.Handler().ServeHTTP(rec, req)
			Expect(rec.Code).To(Equal(http.StatusOK))
			Expect(rec.Body.String()).To(MatchJSON(`
			{
				"meta": {
					"author": "The api2go examples crew",
					"license": "wtfpl",
					"license-url": "http://www.wtfpl.net"
				},
				"data": {
					"attributes": {
						"user-name": "marvin"
					},
					"id": "1",
					"relationships": {
						"sweets": {
							"data": [
								{
									"id": "1",
									"type": "chocolates"
								}
							],
							"links": {
								"related": "http://localhost:31415/v0/users/1/sweets",
								"self": "http://localhost:31415/v0/users/1/relationships/sweets"
							}
						}
					},
					"type": "users"
				},
				"included": [
					{
						"attributes": {
							"name": "Ritter Sport",
							"taste": "Very Good"
						},
						"id": "1",
						"type": "chocolates"
					}
				]
			}
			`))
		})

		It("The relationship route works too", func() {
			req, err := http.NewRequest("GET", "/v0/users/1/relationships/sweets", nil)
			Expect(err).ToNot(HaveOccurred())
			api.Handler().ServeHTTP(rec, req)
			Expect(rec.Code).To(Equal(http.StatusOK))
			Expect(rec.Body.String()).To(MatchJSON(`
			{
				"meta": {
					"author": "The api2go examples crew",
					"license": "wtfpl",
					"license-url": "http://www.wtfpl.net"
				},
				"data": [
					{
						"id": "1",
						"type": "chocolates"
					}
				],
				"links": {
					"related": "http://localhost:31415/v0/users/1/sweets",
					"self": "http://localhost:31415/v0/users/1/relationships/sweets"
				}
			}
			`))
		})
	})
})
