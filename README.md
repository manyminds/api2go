# api2go

[![GoDoc](https://godoc.org/github.com/univedo/api2go?status.svg)](https://godoc.org/github.com/univedo/api2go)
[![Build Status](https://travis-ci.org/univedo/api2go.svg?branch=master)](https://travis-ci.org/univedo/api2go)

A [JSON API](http://jsonapi.org) Implementation for Go.

```go
import "github.com/univedo/api2go"
```

## Usage

Take the simple structs:

```go
type Post struct {
	ID       int
	Title    string
	Comments []Comment
}

type Comment struct {
	ID   int
	Text string
}
```

### Marshaling

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
      "title":"Foobar"
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

### Unmarshaling

Recover the structure from above using

```go
var posts []Post
err := api2go.UnmarshalJSON(json, &posts)
```

## Tests

```sh
go test
ginkgo                # Alternative
ginkgo watch -notify  # Watch for changes
```
