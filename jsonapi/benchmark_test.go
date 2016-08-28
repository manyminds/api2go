package jsonapi

import (
	"database/sql"
	"testing"
)

func BenchmarkMarshal(b *testing.B) {
	post := &Post{
		ID:    1,
		Title: "Title",
	}

	for i := 0; i < b.N; i++ {
		_, err := Marshal(post)
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkUnmarshal(b *testing.B) {
	post := &Post{
		ID:    1,
		Title: "Title",
	}

	data, err := Marshal(post)
	if err != nil {
		panic(err)
	}

	for i := 0; i < b.N; i++ {
		err = Unmarshal(data, &Post{})
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkMarshalSlice(b *testing.B) {
	post := []Post{
		{
			ID:    1,
			Title: "Title",
		},
		{
			ID:    2,
			Title: "Title",
		},
	}

	for i := 0; i < b.N; i++ {
		_, err := Marshal(post)
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkUnmarshalSlice(b *testing.B) {
	post := []*Post{
		{
			ID:    1,
			Title: "Title",
		},
		{
			ID:    2,
			Title: "Title",
		},
	}

	data, err := Marshal(post)
	if err != nil {
		panic(err)
	}

	for i := 0; i < b.N; i++ {
		var posts []Post
		err = Unmarshal(data, &posts)
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkMarshalWithRelationships(b *testing.B) {
	post := &Post{
		ID:          1,
		Title:       "Title",
		AuthorID:    sql.NullInt64{Valid: true, Int64: 1},
		CommentsIDs: []int{1, 2, 3, 4, 5, 6, 7, 8, 9},
	}

	for i := 0; i < b.N; i++ {
		_, err := Marshal(post)
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkUnmarshalWithRelationships(b *testing.B) {
	post := &Post{
		ID:          1,
		Title:       "Title",
		AuthorID:    sql.NullInt64{Valid: true, Int64: 1},
		CommentsIDs: []int{1, 2, 3, 4, 5, 6, 7, 8, 9},
	}

	data, err := Marshal(post)
	if err != nil {
		panic(err)
	}

	for i := 0; i < b.N; i++ {
		err = Unmarshal(data, &Post{})
		if err != nil {
			panic(err)
		}
	}
}
