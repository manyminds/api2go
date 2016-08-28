package jsonapi

import "testing"

func BenchmarkMarshal(b *testing.B) {
	post := &Post{
		ID: 1,
		Title: "Title",
	}

	for i := 0; i < b.N; i++ {
		Marshal(post)
	}
}

func BenchmarkUnmarshal(b *testing.B) {
	post := &Post{
		ID: 1,
		Title: "Title",
	}

	data, _ := Marshal(post)

	for i := 0; i < b.N; i++ {
		Unmarshal(data, &Post{})
	}
}
