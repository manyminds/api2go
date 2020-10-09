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
	ID MagicID `json:"-"`
}

func (m Magic) GetID() string {
	return m.ID.String()
}

func (m Magic) GetLID() string {
	return ""
}

func (m Magic) GetName() string {
	return "magics"
}

type MagicID string

func (m MagicID) String() string {
	return "This should be visible"
}

type Comment struct {
	ID               int       `json:"-"`
	LID              int       `json:"-"`
	Text             string    `json:"text"`
	SubComments      []Comment `json:"-"`
	SubCommentsEmpty bool      `json:"-"`
}

func (c Comment) GetID() string {
	return fmt.Sprintf("%d", c.ID)
}

func (c Comment) GetLID() string {
	if c.LID == 0 {
		return ""
	}
	return fmt.Sprintf("%d", c.LID)
}

func (c Comment) GetName() string {
	return "comments"
}

func (c *Comment) SetID(stringID string) error {
	id, err := strconv.Atoi(stringID)
	if err != nil {
		return err
	}

	c.ID = id

	return nil
}

func (c *Comment) SetLID(stringID string) error {
	id, err := strconv.Atoi(stringID)
	if err != nil {
		return err
	}

	c.LID = id

	return nil
}

func (c Comment) GetReferences() []Reference {
	return []Reference{
		{
			Type:        "comments",
			Name:        "comments",
			IsNotLoaded: c.SubCommentsEmpty,
		},
	}
}

func (c Comment) GetReferencedIDs() []ReferenceID {
	result := []ReferenceID{}

	for _, comment := range c.SubComments {
		commentID := ReferenceID{Type: "comments", Name: "comments", ID: comment.GetID(), LID: comment.GetLID()}
		result = append(result, commentID)
	}

	return result
}

func (c Comment) GetReferencedStructs() []MarshalIdentifier {
	result := []MarshalIdentifier{}

	for _, comment := range c.SubComments {
		result = append(result, comment)
	}

	return result
}

type User struct {
	ID       int    `json:"-"`
	LID      int    `json:"-"`
	Name     string `json:"name"`
	Password string `json:"-"`
}

func (u User) GetID() string {
	return fmt.Sprintf("%d", u.ID)
}

func (u User) GetLID() string {
	if u.LID == 0 {
		return ""
	}
	return fmt.Sprintf("%d", u.LID)
}

func (u User) GetName() string {
	return "users"
}

func (u *User) SetID(stringID string) error {
	id, err := strconv.Atoi(stringID)
	if err != nil {
		return err
	}

	u.ID = id

	return nil
}

func (u *User) SetLID(stringID string) error {
	lid, err := strconv.Atoi(stringID)
	if err != nil {
		return err
	}

	u.LID = lid

	return nil
}

type SimplePost struct {
	ID        string    `json:"-"`
	LID       string    `json:"-"`
	Title     string    `json:"title"`
	Text      string    `json:"text"`
	Internal  string    `json:"-"`
	Size      int       `json:"size"`
	Created   time.Time `json:"created-date"`
	Updated   time.Time `json:"updated-date"`
	topSecret string
}

func (s SimplePost) GetID() string {
	return s.ID
}

func (s SimplePost) GetLID() string {
	return s.LID
}

func (s SimplePost) GetName() string {
	return "simple-posts"
}

func (s *SimplePost) SetID(ID string) error {
	s.ID = ID

	return nil
}

func (s *SimplePost) SetLID(ID string) error {
	s.LID = ID

	return nil
}

type ErrorIDPost struct {
	Error error
}

func (s ErrorIDPost) GetID() string {
	return ""
}

func (s *ErrorIDPost) SetID(ID string) error {
	return s.Error
}

func (s *ErrorIDPost) SetLID(ID string) error {
	return s.Error
}

type Post struct {
	ID            int           `json:"-"`
	LID           int           `json:"-"`
	Title         string        `json:"title"`
	Comments      []Comment     `json:"-"`
	CommentsIDs   []int         `json:"-"`
	CommentsLIDs  []int         `json:"-"`
	CommentsEmpty bool          `json:"-"`
	Author        *User         `json:"-"`
	AuthorID      sql.NullInt64 `json:"-"`
	AuthorLID     sql.NullInt64 `json:"-"`
	AuthorEmpty   bool          `json:"-"`
}

func (c Post) GetID() string {
	return fmt.Sprintf("%d", c.ID)
}

func (c Post) GetLID() string {
	if c.LID == 0 {
		return ""
	}
	return fmt.Sprintf("%d", c.LID)
}

func (c Post) GetName() string {
	return "posts"
}

func (c *Post) SetID(stringID string) error {
	if stringID == "" {
		return nil
	}
	var err error
	c.ID, err = strconv.Atoi(stringID)
	if err != nil {
		return err
	}

	return nil
}

func (c *Post) SetLID(stringID string) error {
	if stringID == "" {
		return nil
	}
	var err error
	c.LID, err = strconv.Atoi(stringID)
	if err != nil {
		return err
	}


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

func (c *Post) SetToOneReferenceID(name string, ID *RelationshipData) error {
	if name == "author" {
		// Ignore empty author relationships
		if ID != nil {
			if ID.ID != "" {
				intID, err := strconv.ParseInt(ID.ID, 10, 64)
				if err != nil {
					return err
				}
				c.AuthorID = sql.NullInt64{Valid: true, Int64: intID}
			}
			if ID.LID != "" {
				intLID, err := strconv.ParseInt(ID.LID, 10, 64)
				if err != nil {
					return err
				}
				c.AuthorLID = sql.NullInt64{Valid: true, Int64: intLID}
			}
		}

		return nil
	}

	return errors.New("There is no to-one relationship named " + name)
}

func (c *Post) SetToManyReferenceIDs(name string, IDs []RelationshipData) error {
	if name == "comments" {
		var commentsIDs []int
		var commentsLIDs []int

		for _, ID := range IDs {
			if ID.ID != "" {
				intID, err := strconv.ParseInt(ID.ID, 10, 64)
				if err != nil {
					return err
				}

				commentsIDs = append(commentsIDs, int(intID))
			}
			if ID.LID != "" {
				intLID, err := strconv.ParseInt(ID.LID, 10, 64)
				if err != nil {
					return err
				}

				commentsLIDs = append(commentsLIDs, int(intLID))
			}
		}

		c.CommentsIDs = commentsIDs
		c.CommentsLIDs = commentsLIDs

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
		authorID := ReferenceID{Type: "users", Name: "author", ID: c.Author.GetID(), LID: c.Author.GetLID()}
		result = append(result, authorID)
	} else if c.AuthorID.Valid {
		authorID := ReferenceID{Type: "users", Name: "author", ID: fmt.Sprintf("%d", c.AuthorID.Int64)}
		result = append(result, authorID)
	}

	if len(c.Comments) > 0 {
		for _, comment := range c.Comments {
			result = append(result, ReferenceID{Type: "comments", Name: "comments", ID: comment.GetID(), LID: comment.GetLID()})
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
	ID       int   `json:"-"`
	LID      int   `json:"-"`
	AuthorID int   `json:"-"`
	Author   *User `json:"-"`
}

func (p AnotherPost) GetID() string {
	return fmt.Sprintf("%d", p.ID)
}

func (p AnotherPost) GetLID() string {
	if p.LID == 0 {
		return ""
	}
	return fmt.Sprintf("%d", p.LID)
}

func (p AnotherPost) GetName() string {
	return "another-posts"
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
	ID    string     `json:"-"`
	LID   string     `json:"-"`
	Title string     `json:"title"`
	Value zero.Float `json:"value"`
}

func (z ZeroPost) GetID() string {
	return z.ID
}

func (z ZeroPost) GetLID() string {
	return z.LID
}

func (z ZeroPost) GetName() string {
	return "zero-posts"
}

type ZeroPostPointer struct {
	ID    string      `json:"-"`
	LID   string      `json:"-"`
	Title string      `json:"title"`
	Value *zero.Float `json:"value"`
}

func (z ZeroPostPointer) GetID() string {
	return z.ID
}

func (z ZeroPostPointer) GetLID() string {
	return z.LID
}

func (z ZeroPostPointer) GetName() string {
	return "zero-post-pointers"
}

type Question struct {
	ID                  string         `json:"-"`
	Text                string         `json:"text"`
	InspiringQuestionID sql.NullString `json:"-"`
	InspiringQuestion   *Question      `json:"-"`
}

func (q Question) GetID() string {
	return q.ID
}

func (q Question) GetLID() string {
	return ""
}

func (q Question) GetName() string {
	return "questions"
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
	ID     int64    `json:"-"`
	LID    int64    `json:"-"`
	Scopes []string `json:"scopes"`
}

func (i Identity) GetID() string {
	return fmt.Sprintf("%d", i.ID)
}

func (i Identity) GetLID() string {
	if i.LID == 0 {
		return ""
	}
	return fmt.Sprintf("%d", i.LID)
}

func (i Identity) GetName() string {
	return "identities"
}

func (i *Identity) SetID(ID string) error {
	if ID == "" {
		return nil
	}
	var err error
	i.ID, err = strconv.ParseInt(ID, 10, 64)
	return err
}

func (i *Identity) SetLID(ID string) error {
	if ID == "" {
		return nil
	}
	var err error
	i.LID, err = strconv.ParseInt(ID, 10, 64)
	return err
}

type Unicorn struct {
	UnicornID int64    `json:"unicorn_id"` // annotations are ignored
	Scopes    []string `json:"scopes"`
}

func (u Unicorn) GetID() string {
	return "magicalUnicorn"
}

func (u Unicorn) GetLID() string {
	return ""
}

func (u Unicorn) GetName() string {
	return "unicorns"
}

type NumberPost struct {
	ID             string `json:"-"`
	LID            string `json:"-"`
	Title          string
	Number         int64
	UnsignedNumber uint64
}

func (n *NumberPost) SetID(ID string) error {
	n.ID = ID
	return nil
}

func (n *NumberPost) SetLID(ID string) error {
	n.LID = ID
	return nil
}

type SQLNullPost struct {
	ID     string      `json:"-"`
	LID    string      `json:"-"`
	Title  zero.String `json:"title"`
	Likes  zero.Int    `json:"likes"`
	Rating zero.Float  `json:"rating"`
	IsCool zero.Bool   `json:"isCool"`
	Today  zero.Time   `json:"today"`
}

func (s SQLNullPost) GetID() string {
	return s.ID
}

func (s SQLNullPost) GetLID() string {
	return s.LID
}

func (s *SQLNullPost) SetID(ID string) error {
	s.ID = ID
	return nil
}

func (s *SQLNullPost) SetLID(ID string) error {
	s.LID = ID
	return nil
}

type RenamedPostWithEmbedding struct {
	Embedded SQLNullPost
	ID       string `json:"-"`
	LID      string `json:"-"`
	Another  string `json:"another"`
	Field    string `json:"foo"`
	Other    string `json:"bar-bar"`
	Ignored  string `json:"-"`
}

func (p *RenamedPostWithEmbedding) SetID(ID string) error {
	p.ID = ID
	return nil
}

func (p *RenamedPostWithEmbedding) SetLID(ID string) error {
	p.LID = ID
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

type CustomLinksPost struct{}

func (n CustomLinksPost) GetID() string {
	return "someID"
}

func (n CustomLinksPost) GetLID() string {
	return ""
}

func (n *CustomLinksPost) SetID(ID string) error {
	return nil
}

func (n *CustomLinksPost) SetLID(ID string) error {
	return nil
}

func (n CustomLinksPost) GetName() string {
	return "posts"
}

func (n CustomLinksPost) GetCustomLinks(base string) Links {
	return Links{
		"nothingInHere": Link{},
		"someLink":      Link{Href: base + `/someLink`},
		"otherLink": Link{
			Href: base + `/otherLink`,
			Meta: Meta{
				"method": "GET",
			},
		},
	}
}

type CustomResourceMetaPost struct{}

func (n CustomResourceMetaPost) GetID() string {
	return "someID"
}

func (n CustomResourceMetaPost) GetLID() string {
	return ""
}

func (n *CustomResourceMetaPost) SetID(ID string) error {
	return nil
}

func (n *CustomResourceMetaPost) SetLID(ID string) error {
	return nil
}

func (n CustomResourceMetaPost) GetName() string {
	return "posts"
}

func (n CustomResourceMetaPost) Meta() Meta {
	return Meta{"access_count": 15}
}

type CustomMetaPost struct{}

func (n CustomMetaPost) GetID() string {
	return "someID"
}

func (n CustomMetaPost) GetLID() string {
	return ""
}

func (n *CustomMetaPost) SetID(ID string) error {
	return nil
}

func (n *CustomMetaPost) SetLID(ID string) error {
	return nil
}

func (n CustomMetaPost) GetName() string {
	return "posts"
}

func (n CustomMetaPost) GetReferences() []Reference {
	return []Reference{
		{
			Type:        "users",
			Name:        "author",
			IsNotLoaded: true,
		},
	}
}

func (n CustomMetaPost) GetReferencedIDs() []ReferenceID {
	return nil
}

func (n CustomMetaPost) GetCustomMeta(linkURL string) map[string]Meta {
	meta := map[string]Meta{
		"author": {
			"someMetaKey":      "someMetaValue",
			"someOtherMetaKey": "someOtherMetaValue",
		},
	}
	return meta
}

type NoRelationshipPosts struct{}

func (n NoRelationshipPosts) GetID() string {
	return "someID"
}

func (n *NoRelationshipPosts) SetID(ID string) error {
	return nil
}

func (n *NoRelationshipPosts) SetLID(ID string) error {
	return nil
}

func (n NoRelationshipPosts) GetName() string {
	return "posts"
}

type ErrorRelationshipPosts struct{}

func (e ErrorRelationshipPosts) GetID() string {
	return "errorID"
}

func (e *ErrorRelationshipPosts) SetID(ID string) error {
	return nil
}

func (e *ErrorRelationshipPosts) SetLID(ID string) error {
	return nil
}

func (e ErrorRelationshipPosts) GetName() string {
	return "posts"
}

func (e ErrorRelationshipPosts) SetToOneReferenceID(name string, ID *RelationshipData) error {
	return errors.New("this never works")
}

func (e ErrorRelationshipPosts) SetToManyReferenceIDs(name string, IDs []RelationshipData) error {
	return errors.New("this also never works")
}

type Image struct {
	ID    string      `json:"-"`
	LID   string      `json:"-"`
	Ports []ImagePort `json:"image-ports"`
}

func (i Image) GetID() string {
	return i.ID
}

func (i Image) GetLID() string {
	return i.LID
}

func (i Image) GetName() string {
	return "images"
}

func (i *Image) SetID(ID string) error {
	i.ID = ID
	return nil
}

func (i *Image) SetLID(ID string) error {
	i.LID = ID
	return nil
}

type ImagePort struct {
	Protocol string `json:"protocol"`
	Number   int    `json:"number"`
}

type Article struct {
	IDs          []string         `json:"-"`
	Type         string           `json:"-"`
	Name         string           `json:"-"`
	Relationship RelationshipType `json:"-"`
}

func (a Article) GetID() string {
	return "id"
}

func (a Article) GetLID() string {
	return ""
}

func (a Article) GetName() string {
	return "articles"
}

func (a Article) GetReferences() []Reference {
	return []Reference{{Type: a.Type, Name: a.Name, Relationship: a.Relationship}}
}

func (a Article) GetReferencedIDs() []ReferenceID {
	referenceIDs := []ReferenceID{}

	for _, id := range a.IDs {
		referenceIDs = append(referenceIDs, ReferenceID{ID: id, Type: a.Type, Name: a.Name, Relationship: a.Relationship})
	}

	return referenceIDs
}

type DeepDedendencies struct {
	ID            string             `json:"-"`
	Relationships []DeepDedendencies `json:"-"`
}

func (d DeepDedendencies) GetID() string {
	return d.ID
}

func (d DeepDedendencies) GetLID() string {
	return ""
}

func (DeepDedendencies) GetName() string {
	return "deep"
}

func (d DeepDedendencies) GetReferences() []Reference {
	return []Reference{{Type: "deep", Name: "deps"}}
}

func (d DeepDedendencies) GetReferencedIDs() []ReferenceID {
	references := make([]ReferenceID, 0, len(d.Relationships))

	for _, r := range d.Relationships {
		references = append(references, ReferenceID{ID: r.ID, Type: "deep", Name: "deps"})
	}

	return references
}

func (d DeepDedendencies) GetReferencedStructs() []MarshalIdentifier {
	var structs []MarshalIdentifier

	for _, r := range d.Relationships {
		structs = append(structs, r)
		structs = append(structs, r.GetReferencedStructs()...)
	}

	return structs
}
