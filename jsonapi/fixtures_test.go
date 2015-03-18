package jsonapi

import (
	"database/sql"
	"fmt"
	"strconv"
)

type Magic struct {
	ID MagicID
}

func (m Magic) GetID() string {
	return m.ID.String()
}

type MagicID string

func (m MagicID) String() string {
	return "This should be visible"
}

type Comment struct {
	ID   int `json:"-"`
	Text string
}

func (c Comment) GetID() string {
	return fmt.Sprintf("%d", c.ID)
}

func (c *Comment) SetID(stringID string) error {
	id, err := strconv.Atoi(stringID)
	if err != nil {
		return err
	}

	c.ID = id

	return nil
}

type User struct {
	ID       int
	Name     string
	Password string `json:"-"`
}

func (u User) GetID() string {
	return fmt.Sprintf("%d", u.ID)
}

func (u *User) SetID(stringID string) error {
	id, err := strconv.Atoi(stringID)
	if err != nil {
		return err
	}

	u.ID = id

	return nil
}

type SimplePost struct {
	ID, Title, Text string
}

func (s SimplePost) GetID() string {
	return s.ID
}

type Post struct {
	ID          int
	Title       string
	Comments    []Comment     `json:"-"`
	CommentsIDs []int         `json:"-"`
	Author      *User         `json:"-"`
	AuthorID    sql.NullInt64 `json:"-"`
}

func (c Post) GetID() string {
	return fmt.Sprintf("%d", c.ID)
}

func (c *Post) SetID(stringID string) error {
	id, err := strconv.Atoi(stringID)
	if err != nil {
		return err
	}

	c.ID = id

	return nil
}

func (c Post) GetReferences() []Reference {
	return []Reference{
		{
			Type: "comments",
			Name: "comments",
		},
		{
			Type: "users",
			Name: "author",
		},
	}
}

func (c *Post) SetReferencedIDs(ids []ReferenceID) error {
	return nil
}

func (c Post) GetReferencedIDs() []ReferenceID {
	result := []ReferenceID{}

	if c.Author != nil {
		authorID := ReferenceID{Type: "users", Name: "author", ID: c.Author.GetID()}
		result = append(result, authorID)
	} else if c.AuthorID.Valid {
		authorID := ReferenceID{Type: "users", Name: "author", ID: fmt.Sprintf("%d", c.AuthorID.Int64)}
		result = append(result, authorID)
	}

	if len(c.Comments) > 0 {
		for _, comment := range c.Comments {
			result = append(result, ReferenceID{Type: "comments", Name: "comments", ID: comment.GetID()})
		}
	} else if len(c.CommentsIDs) > 0 {
		for _, commentID := range c.CommentsIDs {
			result = append(result, ReferenceID{Type: "comments", Name: "comments", ID: fmt.Sprintf("%d", commentID)})
		}
	}

	return result
}

func (c Post) GetReferencedStructs() []MarshalIdentifier {
	result := []MarshalIdentifier{}
	if c.Author != nil {
		result = append(result, c.Author)
	}
	for key := range c.Comments {
		result = append(result, c.Comments[key])
	}
	return result
}

func (c *Post) SetReferencedStructs(references []UnmarshalIdentifier) error {
	return nil
}
