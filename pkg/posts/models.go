package posts

import (
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
)

// Post represents a Civil Post
type Post interface {
	GetBase() *Base
	GetType() string
}

// Pagination is used to pass cursors for pagination
type Pagination struct {
	AfterCursor  string
	BeforeCursor string
}

// PostSearchResult includes SearchResults of Posts
type PostSearchResult struct {
	Posts []Post
	Pagination
}

// Base contains fields common to all types of Posts
type Base struct {
	ID        string `gorm:"type:uuid;primary_key" json:"-"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
	ParentID  *string `gorm:"index:parent"`
	ChannelID string  `gorm:"index:channel"`
	AuthorID  string  `gorm:"index:author"`
	PostType  string  `gorm:"index:type"`
	Data      postgres.Jsonb
}

// GetBase returns itself in order to implement the Post interface
func (p Base) GetBase() *Base {
	return &p
}

// Boost is a type of Post that is describes an initiative that can be funded
type Boost struct {
	Base         `json:"-"`
	Title        string      `json:"text"`
	CurrencyCode string      `json:"currency_code"`
	GoalAmount   float64     `json:"goal_amount"`
	DateEnd      time.Time   `json:"date_end"`
	Why          string      `json:"why,omitempty"`
	What         string      `json:"what,omitempty"`
	About        string      `json:"about,omitempty"`
	Items        []BoostItem `json:"items,omitempty"`
}

// BoostItem describes the items within a boost
type BoostItem struct {
	Item string  `json:"item"`
	Cost float64 `json:"cost"`
}

// GetType returns the post type "Boost"
func (b Boost) GetType() string {
	return "boost"
}

// Comment is a type of Post that contains just type
type Comment struct {
	Base `json:"-"`
	Text string `json:"text"`
}

// GetType returns the post type "Boost"
func (b Comment) GetType() string {
	return "comment"
}

// ExternalLink is a type of Post that links to another web page
type ExternalLink struct {
	Base `json:"-"`
	URL  string `json:"url"`
}

// GetType returns the post type "Boost"
func (b ExternalLink) GetType() string {
	return "externallink"
}
