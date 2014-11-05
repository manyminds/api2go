# api2go

[![GoDoc](https://godoc.org/github.com/univedo/api2go?status.svg)](https://godoc.org/github.com/univedo/api2go)
[![Build Status](https://travis-ci.org/univedo/api2go.svg?branch=master)](https://travis-ci.org/univedo/api2go)

A [JSON API](http://jsonapi.org) Implementation for Go, to be used e.g. as server for [Ember Data](https://github.com/emberjs/data).

```go
import "github.com/univedo/api2go"
```

**api2go works, but we're still working on some rough edges. Things might change. Open an issue and join in!  **

## Usage

Take the simple structs:

```go
type Post struct {
  ID          int
  Title       string
  Comments    []Comment
  CommentsIDs []int
}

type Comment struct {
  ID   int
  Text string
}
```

### Building a REST API

First, write an implementation of `api2go.DataSource`. You have to implement 5 methods:

```go
type fixtureSource struct {}

func (s *fixtureSource) FindAll() (interface{}, error) {
  // Return a slice of all posts as []Post
}

func (s *fixtureSource) FindOne(id string) (interface{}, error) {
  // Return a single post by ID as Post
}

func (s *fixtureSource) Create(obj interface{}) (string, error) {
  // Save the new Post in `obj` and return its ID.
}

func (s *fixtureSource) Delete(id string) error {
  // Delete a post
}

func (s *fixtureSource) Update(obj interface{}) error {
  // Apply the new values in the Post in `obj`
}
```

As an example, check out the implementation of `fixtureSource` in [api_test.go](/api_test.go).

You can then create an API:

```go
api := api2go.NewAPI("v1")
api.AddResource(Post{}, &PostsSource{})
http.ListenAndServe(":8080", api.Handler())
```

This generates the standard endpoints:

```
OPTIONS /v1/posts
OPTIONS /v1/posts/<id>
GET     /v1/posts
POST    /v1/posts
GET     /v1/posts/<id>
PUT     /v1/posts/<id>
DELETE  /v1/posts/<id>
```

### Use Custom Controllers

By using the `api2go.DataSource` and registering it with `AddResource`,
api2go will do everything for you automatically and you cannot change it. This
means that you cannot access the request, perform some user authorization and so on...

In order to register a Controller for a DataSource, implement the `api2go.Controller` interface:

```go
type Controller interface {
  // FindAll get's called after resource was called
  FindAll(r *http.Request, objs *interface{}) error

  // FindOne get's called after resource was called
  FindOne(r *http.Request, obj *interface{}) error

  // Create get's called before resource was called
  Create(r *http.Request, obj *interface{}) error

  // Delete get's called before resource was called
  Delete(r *http.Request, id string) error

  // Update get's called before resource was called
  Update(r *http.Request, obj *interface{}) error
}
```

Now, you can access the request and for example perform some user authorization by reading the
`Authorization` header or some cookies. In addition, you also have the object out of your database, in
case you need that too.

To deny access you just return a new `httpError` with `api2go.NewHTTPError`

```go
...
func (c *yourController) FindAll(r *http.Request, objs *interface{}) error {
  // do some authorization stuff
  return api2go.NewHTTPError(someError, "Access denied", 403)
}
...
```

Register your Controller with the DataSource together

```go
api := api2go.NewAPI("v1")
api.AddResourceWithController(Post{}, &PostsSource{}, &YourController{})
http.ListenAndServe(":8080", api.Handler())
```

### Manual marshaling / unmarshaling

```go
comment1 = Comment{ID: 1, Text: "First!"}
comment2 = Comment{ID: 2, Text: "Second!"}
post = Post{ID: 1, Title: "Foobar", Comments: []Comment{comment1, comment2}}

json, err := api2go.MarshalJSON(post)
```

will yield

```json
{
  "posts": [
    {
      "id": "1",
      "links": {"comments": ["1", "2"]},
      "title": "Foobar"
    }
  ],
  "linked": {
    "comments": [
      {"id": "1", "text": "First!"},
      {"id": "2", "text": "Second!"}
    ]
  }
}
```

Recover the structure from above using

```go
var posts []Post
err := api2go.UnmarshalFromJSON(json, &posts)
// posts[0] == Post{ID: 1, Title: "Foobar", CommentsIDs: []int{1, 2}}
```

Note that when unmarshaling, api2go will always fill the `CommentsIDs` field, never the `Comments` field.

## Conventions

Structs MUST have:

- A field called `ID` that is either a `string` or `int`.

Structs MAY have:

- Fields with struct-slices, e.g. `Comments []Comment`. They will be serialized as links (using the field name) and the linked structs embedded.
- Fields with `int` / `string` slices, e.g. `CommentsIDs`. They will be serialized as links (using the field name minus an optional `IDs` suffix), but not embedded.
- Fields of struct type, e.g. `Author Person`. They will be serialized as a single link (using the field name) and the linked struct embedded.
- Fields of `int` / `string` type, ending in `ID`, e.g. `AuthorID`. They will be serialized as a single link (using the field name minus the `ID` suffix), but not embedded.

## Tests

```sh
go test
ginkgo                # Alternative
ginkgo watch -notify  # Watch for changes
```
