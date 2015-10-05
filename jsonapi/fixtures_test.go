package jsonapi

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"gopkg.in/guregu/null.v2/zero"
)

type Magic struct {
	ID MagicID `jsonapi:"-"`
}

func (m Magic) GetID() string {
	return m.ID.String()
}

type MagicID string

func (m MagicID) String() string {
	return "This should be visible"
}

type Comment struct {
	ID   int `jsonapi:"-"`
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
	ID       int `jsonapi:"-"`
	Name     string
	Password string `jsonapi:"-"`
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
	ID          string `jsonapi:"-"`
	Title, Text string
	Internal    string `jsonapi:"-"`
	Size        int
	Created     time.Time  `jsonapi:"name=create-date"`
	Updated     *time.Time `jsonapi:"name=update-date"`
}

func (s SimplePost) GetID() string {
	return s.ID
}

func (s *SimplePost) SetID(ID string) error {
	s.ID = ID

	return nil
}

type Post struct {
	ID            int `jsonapi:"-"`
	Title         string
	Comments      []Comment     `jsonapi:"-"`
	CommentsIDs   []int         `jsonapi:"-"`
	CommentsEmpty bool          `jsonapi:"-"`
	Author        *User         `jsonapi:"-"`
	AuthorID      sql.NullInt64 `jsonapi:"-"`
	AuthorEmpty   bool          `jsonapi:"-"`
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
			Type:        "comments",
			Name:        "comments",
			IsNotLoaded: c.CommentsEmpty,
		},
		{
			Type:        "users",
			Name:        "author",
			IsNotLoaded: c.AuthorEmpty,
		},
	}
}

func (c *Post) SetToOneReferenceID(name, ID string) error {
	if name == "author" {
		intID, err := strconv.ParseInt(ID, 10, 64)
		if err != nil {
			return err
		}
		c.AuthorID = sql.NullInt64{Valid: true, Int64: intID}

		return nil
	}

	return errors.New("There is no to-one relationship named " + name)
}

func (c *Post) SetToManyReferenceIDs(name string, IDs []string) error {
	if name == "comments" {
		commentsIDs := []int{}
		for _, ID := range IDs {
			intID, err := strconv.ParseInt(ID, 10, 64)
			if err != nil {
				return err
			}
			commentsIDs = append(commentsIDs, int(intID))
		}
		c.CommentsIDs = commentsIDs

		return nil
	}

	return errors.New("There is no to-many relationship named " + name)
}

func (c *Post) SetReferencedIDs(ids []ReferenceID) error {
	for _, reference := range ids {
		intID, err := strconv.ParseInt(reference.ID, 10, 64)
		if err != nil {
			return err
		}

		switch reference.Name {
		case "comments":
			c.CommentsIDs = append(c.CommentsIDs, int(intID))
		case "author":
			c.AuthorID = sql.NullInt64{Valid: true, Int64: intID}
		}
	}

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

type AnotherPost struct {
	ID       int   `jsonapi:"-"`
	AuthorID int   `jsonapi:"-"`
	Author   *User `jsonapi:"-"`
}

func (p AnotherPost) GetID() string {
	return fmt.Sprintf("%d", p.ID)
}

func (p AnotherPost) GetReferences() []Reference {
	return []Reference{
		{
			Type: "users",
			Name: "author",
		},
	}
}

func (p AnotherPost) GetReferencedIDs() []ReferenceID {
	result := []ReferenceID{}

	if p.AuthorID != 0 {
		result = append(result, ReferenceID{ID: fmt.Sprintf("%d", p.AuthorID), Name: "author", Type: "users"})
	}

	return result
}

type ZeroPost struct {
	ID    string `jsonapi:"-"`
	Title string
	Value zero.Float
}

func (z ZeroPost) GetID() string {
	return z.ID
}

type ZeroPostPointer struct {
	ID    string `jsonapi:"-"`
	Title string
	Value *zero.Float
}

func (z ZeroPostPointer) GetID() string {
	return z.ID
}

type Question struct {
	ID                  string `jsonapi:"-"`
	Text                string
	InspiringQuestionID sql.NullString `jsonapi:"-"`
	InspiringQuestion   *Question      `jsonapi:"-"`
}

func (q Question) GetID() string {
	return q.ID
}

func (q Question) GetReferences() []Reference {
	return []Reference{
		{
			Type: "questions",
			Name: "inspiringQuestion",
		},
	}
}

func (q Question) GetReferencedIDs() []ReferenceID {
	result := []ReferenceID{}
	if q.InspiringQuestionID.Valid {
		result = append(result, ReferenceID{ID: q.InspiringQuestionID.String, Name: "inspiringQuestion", Type: "questions"})
	}

	return result
}

func (q Question) GetReferencedStructs() []MarshalIdentifier {
	result := []MarshalIdentifier{}
	if q.InspiringQuestion != nil {
		result = append(result, *q.InspiringQuestion)
	}

	return result
}

type Identity struct {
	ID     int64    `jsonapi:"-"`
	Scopes []string `json:"scopes"`
}

func (i Identity) GetID() string {
	return fmt.Sprintf("%d", i.ID)
}

func (i *Identity) SetID(ID string) error {
	var err error
	i.ID, err = strconv.ParseInt(ID, 10, 64)
	return err
}

type Unicorn struct {
	UnicornID int64    `json:"unicorn_id"` //Annotations are ignored
	Scopes    []string `json:"scopes"`
}

func (u Unicorn) GetID() string {
	return "magicalUnicorn"
}

type NumberPost struct {
	ID             string `jsonapi:"-"`
	Title          string
	Number         int64
	UnsignedNumber uint64
}

func (n *NumberPost) SetID(ID string) error {
	n.ID = ID

	return nil
}

type SQLNullPost struct {
	ID     string `jsonapi:"-"`
	Title  zero.String
	Likes  zero.Int
	Rating zero.Float
	IsCool zero.Bool
}

func (s SQLNullPost) GetID() string {
	return s.ID
}

func (s *SQLNullPost) SetID(ID string) error {
	s.ID = ID
	return nil
}

func (s SQLNullPost) GetName() string {
	return "sqlNullPosts"
}

type RenamedComment struct {
	Data string
}

func (r RenamedComment) GetID() string {
	return "666"
}

func (r RenamedComment) GetName() string {
	return "renamed-comments"
}

type CompleteServerInformation struct{}

const completePrefix = "http://my.domain/v1"
const baseURL = "http://my.domain"
const prefix = "v1"

func (i CompleteServerInformation) GetBaseURL() string {
	return baseURL
}

func (i CompleteServerInformation) GetPrefix() string {
	return prefix
}

type BaseURLServerInformation struct{}

func (i BaseURLServerInformation) GetBaseURL() string {
	return baseURL
}

func (i BaseURLServerInformation) GetPrefix() string {
	return ""
}

type PrefixServerInformation struct{}

func (i PrefixServerInformation) GetBaseURL() string {
	return ""
}

func (i PrefixServerInformation) GetPrefix() string {
	return prefix
}
