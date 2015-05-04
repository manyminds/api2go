# api2go

[![Join the chat at https://gitter.im/manyminds/api2go](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/manyminds/api2go?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

[![GoDoc](https://godoc.org/github.com/manyminds/api2go?status.svg)](https://godoc.org/github.com/manyminds/api2go)
[![Build Status](https://travis-ci.org/manyminds/api2go.svg?branch=master)](https://travis-ci.org/manyminds/api2go)

A [JSON API](http://jsonapi.org) Implementation for Go, to be used e.g. as server for [Ember Data](https://github.com/emberjs/data).

```go
import "github.com/manyminds/api2go"
```

**api2go works, but we're still working on some rough edges. Things might change. Open an issue and join in!**

**we moved the project from the univedo organization to manyminds. If you upgrade, please fix your import paths**

Note: if you only need the marshaling functionality, you can install the subpackage via
 ```go
go get github.com/manyminds/api2go/jsonapi
```

## TOC
- [Examples](#examples)
- [Interfaces to implement](#interfaces-to-implement)
  - [MarshalIdentifier](#marshalidentifier)
  - [UnmarshalIdentifier](#unmarshalidentifier)
  - [Marshalling with References to other structs](#marshalling-with-references-to-other-structs)
  - [Unmarshalling with references to other structs](#unmarshalling-with-references-to-other-structs)
- [Ignoring fields](#ignoring-fields)
- [Manual marshaling / unmarshaling](#manual-marshaling--unmarshaling)
- [SQL Null-Types](#sql-null-types)
- [Building a REST API](#building-a-rest-api)
  - [Query Params](#query-params)
  - [Using Pagination](#using-pagination)
  - [Fetching related IDs](#fetching-related-ids)
  - [Fetching related resources](#fetching-related-resources)
- [Tests](#tests)

## Examples

Examples can be found [here](https://github.com/manyminds/api2go/blob/master/examples/crud_example.go).

## Interfaces to implement
For the following query and result examples, imagine the following 2 structs which represent a posts and
comments that belong with a has-many relation to the post.

```go
type Post struct {
  ID          int
  Title       string
  Comments    []Comment `json:"-"` // this will be ignored by the api2go marshaller
  CommentsIDs []int     `json:"-"` // it's only useful for our internal relationship handling
}

type Comment struct {
  ID   int
  Text string `jsonapi:"name=content"`
}
```

You must at least implement one interface for api2go to work, which is the one for marshalling/unmarshalling the primary `ID` of the struct
that you want to marshal/unmarshal. This is because of the huge variety of types that you could  use for the primary ID. For example a string,
a UUID or a BSON Object for MongoDB etc...

If the struct already has a field named `ID`, or `Id`, it will be ignored automatically. If your ID field has a different name, please use the
json ignore tag. Api2go will use the `GetID` method that you implemented for your struct to fetch the ID of the struct.

In order to use different internal names for elements, you can specify a jsonapi tag. The api will marshal results now with the name in the tag.
Create/Update/Delete works accordingly, but will fallback to the internal value as well if possible.

### MarshalIdentifier
```go
type MarshalIdentifier interface {
	GetID() string
}
```

Implement this interface to marshal a struct.

### UnmarshalIdentifier
```go
type UnmarshalIdentifier interface {
	SetID(string) error
}
```

This is the corresponding interface to MarshalIdentifier. Implement this interface in order to unmarshal incoming json into
a struct.

### Marshalling with References to other structs
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
result inside the `included` object.

We choose to do this because it gives you better flexibility and eliminates the conventions in the previous versions of api2go. **You can
now choose how you internally manage relations.** So, there are no limits regarding the use of ORMs.

### Unmarshalling with references to other structs
Incoming jsons can also contain reference IDs. In order to unmarshal them correctly, you have to implement the following interfaces. If you only have to-one
relationships, the `UnmarshalToOneRelations` interface is enough. 

```go
// UnmarshalToOneRelations must be implemented to unmarshal to-one relations
type UnmarshalToOneRelations interface {
	SetToOneReferenceID(name, ID string) error
}

// UnmarshalToManyRelations must be implemented to unmarshal to-many relations
type UnmarshalToManyRelations interface {
	SetToManyReferenceIDs(name string, IDs []string) error
}
```

**If you need to know more about how to use the interfaces, look at our tests or at the example project.**

## Ignoring fields
api2go ignores all fields that are marked with the `json"-"` ignore tag. This is useful if your struct has some more
fields which are only used internally to manage relations or data that needs to stay private, like a password field.

## Manual marshaling / unmarshaling
Please keep in mind that this only works if you implemented the previously mentioned interfaces. Manual marshalling and
unmarshalling makes sense, if you do not want to use our API that automatically generates all the necessary routes for you. You
can directly use our sub-package `github.com/manyminds/api2go/jsonapi`

```go
comment1 = Comment{ID: 1, Text: "First!"}
comment2 = Comment{ID: 2, Text: "Second!"}
post = Post{ID: 1, Title: "Foobar", Comments: []Comment{comment1, comment2}}

json, err := jsonapi.MarshalToJSON(post)
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
          "linkage": [
            {
              "id": "1",
              "type": "comments"
            },
            {
              "id": "2",
              "type": "comments"
            }
          ],
        }
      },
      "title": "Foobar"
    }
  ],
  "included": [
    {
      "id": "1",
      "type": "comments",
      "text": "First!"
    },
    {
      "id": "2",
      "type": "comments",
      "text": "Second!"
    }
  ]
}
```

You can also use `jsonapi.MarshalToJSONWithURLs` to automatically generate URLs for the rest endpoints that have a
version and BaseURL prefix. This will generate the same routes that our API uses. This adds `self` and `related` fields
for relations inside the `links` object.

Recover the structure from above using

```go
var posts []Post
err := jsonapi.UnmarshalFromJSON(json, &posts)
// posts[0] == Post{ID: 1, Title: "Foobar", CommentsIDs: []int{1, 2}}
```

## SQL Null-Types
When using a SQL Database it is most likely you want to use the special SQL-Types from the `database/sql` package. These are

- sql.NullBool
- sql.NullFloat64
- sql.NullInt64
- sql.NullString

The Problem is, that they internally manage the `null` value behavior by using a custom struct. In order to Marshal und Unmarshal
these values, it is required to implement the `json.Marshaller` and `json.Unmarshaller` interfaces of the go standard library.

But you dont have to do this by yourself! There already is a library that did the work for you. We recommend that you use the types
of this library: http://gopkg.in/guregu/null.v2/zero

## Building a REST API

First, write an implementation of `api2go.CRUD`. You have to implement at least these 4 methods:

```go
type fixtureSource struct {}

func (s *fixtureSource) FindOne(ID string, r api2go.Request) (interface{}, error) {
  // Return a single post by ID as Post
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

To fetch all objects of a specific resource you can choose to implement one or both of the following
interfaces:

```go
type FindAll interface {
	// FindAll returns all objects
	FindAll(req Request) (interface{}, error)
}

type PaginatedFindAll interface {
	PaginatedFindAll(req Request) (obj interface{}, totalCount uint, err error)
}
```

`FindAll` returns everything. You could limit the results only by using Query Params which are described [here](#query-params)

`PaginatedFindAll` can also use Query Params, but in addition to that it does not need to send all objects at once and can split
up the result with pagination. You have to return the total number of found objects in order to let our API automatically generate
pagination links. More about pagination is described [here](#using-pagination)

You can then create an API:

```go
api := api2go.NewAPI("v1")
api.AddResource(Post{}, &PostsSource{})
http.ListenAndServe(":8080", api.Handler())
```

Instead of `api2go.NewAPI` you can also use `api2go.NewAPIWithBaseURL("v1", "http://yourdomain.com")` to prefix all
automatically generated routes with your domain and protocoll.

This generates the standard endpoints:

```
OPTIONS /v1/posts
OPTIONS /v1/posts/<id>
GET     /v1/posts
POST    /v1/posts
GET     /v1/posts/<id>
PATCH   /v1/posts/<id>
DELETE  /v1/posts/<id>
GET     /v1/posts/<id>/comments            // fetch referenced comments of a post
GET     /v1/posts/<id>/links/comments      // fetch IDs of the referenced comments only
PATCH   /v1/posts/<id>/links/comments      // replace all related comments

// These 2 routes are only created for to-many relations that implement EditToManyRelations interface
POST    /v1/posts/<id>/links/comments      // Add a new comment reference, only for to-many relations
DELETE  /v1/posts/<id>/links/comments      // Delete a comment reference, only for to-many relations
```

For the last two generated routes, it is necessary to implement the `jsonapi.EditToManyRelations` interface.

```go
type EditToManyRelations interface {
	AddToManyIDs(name string, IDs []string) error
	DeleteToManyIDs(name string, IDs []string) error
}
```

All PATCH, POST and DELETE routes do a `FindOne` and update the values/relations in the previously found struct. This
struct will then be passed on to the `Update` method of a resource struct. So you get all these routes "for free" and just
have to implement the CRUD Update method.

### Query Params
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

### Using Pagination
Api2go can automatically generate the required links for pagination. Currently there are 2 combinations of query
parameters supported:

- page[number], page[size]
- page[offset], page[limit]

Pagination is optional. If you want to support pagination, you have to implement the `PaginatedFindAll` method
in you resource struct. For an example, you best look into our example project.

Example request

```
GET /v0/users?page[number]=2&page[size]=2
```

would return a json with the top level links object

```json
{
  "links": {
    "first": "http://localhost:31415/v0/users?page[number]=1&page[size]=2",
    "last": "http://localhost:31415/v0/users?page[number]=5&page[size]=2",
    "next": "http://localhost:31415/v0/users?page[number]=3&page[size]=2",
    "prev": "http://localhost:31415/v0/users?page[number]=1&page[size]=2"
  },
  "data": [...]
}
```

### Fetching related IDs
The IDs of a relationship can be fetched by following the `self` link of a relationship object in the `links` object
of a result. For the posts and comments example you could use the following generated URL:

```
GET /v1/posts/1/links/comments
```

This would return all comments that are currently referenced by post with ID 1. For example:

```json
{
  "links": {
    "self": "/v1/posts/1/links/comments",
    "related": "/v1/posts/1/comments"
  },
  "data": [
    {
      "type": "comments",
      "id": "1"
    },
    {
      "type":"comments",
      "id": "2"
    }
  ]
}
```

### Fetching related resources
Api2go always creates a `related` field for elements in the `links` object of the result. This is like it's
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
          "related": "/v1/posts/1/comments",
          "self": "/v1/posts/1/links/comments",
          "linkage": [
            {
              "id": "1",
              "type": "comments"
            },
            {
              "id": "2",
              "type": "comments"
            }
          ]
        }
      }
    }
  ]
}
```

If a client requests this `related` url, the `FindAll` method of the comments resource will be called with a query
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
