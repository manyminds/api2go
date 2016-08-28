package jsonapi

import "testing"

func BenchmarkMarshal(b *testing.B) {
	post := &Post{
		ID: 1,
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
		ID: 1,
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
