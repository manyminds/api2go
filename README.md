# api2go

[![GoDoc](https://godoc.org/github.com/univedo/api2go?status.svg)](https://godoc.org/github.com/univedo/api2go)
[![Build Status](https://travis-ci.org/univedo/api2go.svg?branch=master)](https://travis-ci.org/univedo/api2go)

A [JSON API](http://jsonapi.org) Implementation for Go, to be used e.g. as server for [Ember Data](https://github.com/emberjs/data).

```go
import "github.com/univedo/api2go"
```

**api2go works, but we're still working on some rough edges. Things might change. Open an issue and join in!**

** we are currently re-implementing a lot of stuff in a cleaner way with interfaces and upgrading to the RC3 (Final) Standard of jsonapi.org**

Note: if you only need the marshaling functionality, you can install the subpackage via
 ```go
go get github.com/univedo/api2go/jsonapi
```

## Usage

Take the simple structs:

```go
type Post struct {
  ID          int
  Title       string
  Comments    []Comment `json:"-"` // this will be ignored by the api2go marshaller
  CommentsIDs []int     `json:"-"` // it's only useful for our internal relationship handling
}

type Comment struct {
  ID   int
  Text string
}
```

### Interfaces to implement
You must at least implement one interface for api2go to work, which is the one for marshalling/unmarshalling the primary `ID` of the struct
that you want to marshal/unmarshal. This is because of the huge variety of types that you could  use for the primary ID. For example a string,
a UUID or a BSON Object for MongoDB etc...

If the struct already has a field named `ID`, or `Id`, it will be ignored automatically. If your ID field has a different name, please use the
json ignore tag.

#### MarshalIdentifier
```go
type MarshalIdentifier interface {
	GetID() string
}
```

Implement this interface to marshal a struct.

#### UnmarshalIdentifier
```go
type UnmarshalIdentifier interface {
	SetID(string) error
}
```

This is the corresponding interface to MarshalIdentifier. Implement this interface in order to unmarshal incoming json into
a struct.

#### Marshalling with References to other structs
For relationships to work, there are 3 Interfaces that you can use:

```go
type MarshalReferences interface {
	GetReferences() []Reference
}

// MarshalLinkedRelations must be implemented if there are references and the reference IDs should be included
type MarshalLinkedRelations interface {
	MarshalReferences
	MarshalIdentifier
	GetReferencedIDs() []ReferenceID
}

// MarshalIncludedRelations must be implemented if referenced structs should be included
type MarshalIncludedRelations interface {
	MarshalReferences
	MarshalIdentifier
	GetReferencedStructs() []MarshalIdentifier
}
```

Here, you can choose what you want to implement too, but, you must at least implement `MarshalReferences` and `MarshalLinkedRelations`.

`MarshalReferences` must be implemented in order for api2go to know which relations are possible for your struct.

`MarshalLinkedRelations` must be implemented to retrieve the `IDs` of the relations that are connected to this struct. This method
could also return an empty array, if there are currently no relations. This is why there is the `MarshalReferences` interface, so that api2go
knows what is possible, even if nothing is referenced at the time.

In addition to that, you can implement `MarshalIncludedRelations` which exports the complete referenced structs and embeds them in the json
result inside the `linked` object.

We choose to do this because it gives you better flexibility and eliminates the conventions in the previous versions of api2go. **You can
now choose how you internally manage relations.** So, there are no limits regarding the use of ORMs.

#### Unmarshalling with references to other structs
Incoming jsons can also contain reference IDs. In order to unmarshal them correctly, you have to implement the following interface

```go
type UnmarshalLinkedRelations interface {
	SetReferencedIDs([]ReferenceID) error
}
```

**If you need to know more about how to use the interfaces, look at our tests or at the example project.**

### Ignoring fields
api2go ignores all fields that are marked with the `json"-"` ignore tag. This is useful if your struct has some more
fields which are only used internally to manage relations or data that needs to stay private, like a password field.

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
  "data": [
    {
      "id": "1",
      "type": "posts",
      "links": {
        "comments": {
          "ids": ["1", "2"],
          "type": "comments",
          "resource": "/posts/1/comments"
        }
      },
      "title": "Foobar"
    }
  ],
  "linked": [
    {"id": "1", "type": "comments", "text": "First!"},
    {"id": "2", "type": "comments", "text": "Second!"}
  ]
}
```

Recover the structure from above using

```go
var posts []Post
err := api2go.UnmarshalFromJSON(json, &posts)
// posts[0] == Post{ID: 1, Title: "Foobar", CommentsIDs: []int{1, 2}}
```

Note that when unmarshaling, api2go will always fill the `CommentsIDs` field, never the `Comments` field.

### Building a REST API

First, write an implementation of `api2go.DataSource`. You have to implement 5 methods:

```go
type fixtureSource struct {}

func (s *fixtureSource) FindAll(r api2go.Request) (interface{}, error) {
  // Return a slice of all posts as []Post
}

func (s *fixtureSource) FindOne(ID string, r api2go.Request) (interface{}, error) {
  // Return a single post by ID as Post
}

func (s *fixtureSource) FindMultiple(IDs []string, r api2go.Request) (interface{}, error) {
  // Return multiple posts by ID as []Post
  // For example for Requests like GET /posts/1,2,3
}

func (s *fixtureSource) Create(obj interface{}, r api2go.Request) (string, error) {
  // Save the new Post in `obj` and return its ID.
}

func (s *fixtureSource) Delete(id string, r api2go.Request) error {
  // Delete a post
}

func (s *fixtureSource) Update(obj interface{}, r api2go.Request) error {
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
GET     /v1/posts/<id>,<id>,...
GET     /v1/posts/<id>/comments
```

#### Query Params
To support all the features mentioned in the `Fetching Resources` section of Jsonapi:
http://jsonapi.org/format/#fetching

If you want to support any parameters mentioned there, you can access them in your Resource
via the `api2go.Request` Parameter. This currently supports `QueryParams` which holds
all query parameters as `map[string][]string` unfiltered. So you can use it for:
  * Filtering
  * Inclusion of Linked Resources
  * Sparse Fieldsets
  * Sorting
  * Aything else you want to do that is not in the official Jsonapi Spec

```go
type fixtureSource struct {}

func (s *fixtureSource) FindAll(req api2go.Request) (interface{}, error) {
  for key, values range req.QueryParams {
    ...
  }
  ...
}
```

If there are multiple values, you have to separate them with a comma. api2go automatically
slices the values for you.

```
Example Request
GET /people?fields=id,name,age

req.QueryParams["fields"] contains values: ["id", "name", "age"]
```

### Loading related resources
Api2go always creates a `resource` property for elements in the `links` property of the result. This is like it's
specified on jsonapi.org. Post example:

```json
{
  "data": [
    {
      "id": "1",
      "type": "posts",
      "title": "Foobar",
      "links": {
        "comments": {
          "resource": "/v1/posts/1/comments",
          "ids": ["1", "2"],
          "type": "comments",
        }
      }
    }
  ]
}
```

If a client requests this `resource` url, the `FindAll` method of the comments resource will be called with a query
parameter `postsID`.

So if you implement the `FindAll` method, do not forget to check for all possible query Parameters. This means you have
to check all your other structs and if it references the one for that you are implementing `FindAll`, check for the
query Paramter and only return comments that belong to it. In this example, return the comments for the Post.


## Tests

```sh
go test
ginkgo                # Alternative
ginkgo watch -notify  # Watch for changes
```
