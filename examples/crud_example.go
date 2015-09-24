/*
Package examples shows how to implement a basic CRUD for two data structures with the api2go server functionality.
To play with this example server you can run some of the following curl requests

In order to demonstrate dynamic baseurl handling for requests, apply the --header="REQUEST_URI:https://www.your.domain.example.com" parameter to any of the commands.

Create a new user:
	curl -X POST http://localhost:31415/v0/users -d '{"data" : [{"type" : "users" , "attributes": {"user-name" : "marvin"}}]}'

List users:
	curl -X GET http://localhost:31415/v0/users

List paginated users:
	curl -X GET 'http://localhost:31415/v0/users?page\[offset\]=0&page\[limit\]=2'
OR
	curl -X GET 'http://localhost:31415/v0/users?page\[number\]=1&page\[size\]=2'

Update:
	curl -vX PATCH http://localhost:31415/v0/users/1 -d '{ "data" : {"type" : "users", "id": "1", "attributes": {"user-name" : "better marvin"}}}'

Delete:
	curl -vX DELETE http://localhost:31415/v0/users/2

Create a chocolate with the name sweet
	curl -X POST http://localhost:31415/v0/chocolates -d '{"data" : [{"type" : "chocolates" , "attributes": {"name" : "Ritter Sport", "taste": "Very Good"}}]}'

Create a user with a sweet
	curl -X POST http://localhost:31415/v0/users -d '{"data" : [{"type" : "users" , "attributes": {"user-name" : "marvin"}, "relationships": {"sweets": {"data": [{"type": "chocolates", "id": "1"}]}}}]}'

List a users sweets
	curl -X GET http://localhost:31415/v0/users/1/sweets

Replace a users sweets
	curl -X PATCH http://localhost:31415/v0/users/1/relationships/sweets -d '{"data" : [{"type": "chocolates", "id": "2"}]}'

Add a sweet
	curl -X POST http://localhost:31415/v0/users/1/relationships/sweets -d '{"data" : [{"type": "chocolates", "id": "2"}]}'

Remove a sweet
	curl -X DELETE http://localhost:31415/v0/users/1/relationships/sweets -d '{"data" : [{"type": "chocolates", "id": "2"}]}'
*/
package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/manyminds/api2go"
	"github.com/manyminds/api2go/examples/model"
	"github.com/manyminds/api2go/examples/resolver"
	"github.com/manyminds/api2go/examples/resource"
	"github.com/manyminds/api2go/examples/storage"
)

// PrettyJSONContentMarshaler for JSON in a human readable format
type PrettyJSONContentMarshaler struct{}

// Marshal marshals to pretty JSON
func (m PrettyJSONContentMarshaler) Marshal(i interface{}) ([]byte, error) {
	return json.MarshalIndent(i, "", "    ")
}

// Unmarshal the JSON
func (m PrettyJSONContentMarshaler) Unmarshal(data []byte, i interface{}) error {
	return json.Unmarshal(data, i)
}

// MarshalError to configure error marshaling
func (m PrettyJSONContentMarshaler) MarshalError(err error) string {
	jsonmarshaler := api2go.JSONContentMarshaler{}
	return jsonmarshaler.MarshalError(err)
}

func main() {
	marshalers := map[string]api2go.ContentMarshaler{
		"application/vnd.api+json": PrettyJSONContentMarshaler{},
	}

	port := 31415
	api := api2go.NewAPIWithMarshalling("v0", &resolver.RequestURL{Port: port}, marshalers)
	userStorage := storage.NewUserStorage()
	chocStorage := storage.NewChocolateStorage()
	api.AddResource(model.User{}, resource.UserResource{ChocStorage: chocStorage, UserStorage: userStorage})
	api.AddResource(model.Chocolate{}, resource.ChocolateResource{ChocStorage: chocStorage, UserStorage: userStorage})

	fmt.Printf("Listening on :%d", port)
	handler := api.Handler().(*httprouter.Router)
	// It is also possible to get the instance of julienschmidt/httprouter and add more custom routes!
	handler.GET("/hello-world", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		fmt.Fprint(w, "Hello World!\n")
	})

	http.ListenAndServe(fmt.Sprintf(":%d", port), handler)
}
