package posts

import (
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/joincivil/civil-api-server/pkg/payments"
)

// TYPES
const (
	TypeBoost               = "boost"
	TypeExternalLink        = "externallink"
	TypeComment             = "comment"
	TypePost                = "posts"
	TypeCommentAnnouncement = "comment_announcement"
	TypeCommentPrompt       = "comment_prompt"
	TypeCommentDefault      = "comment_default"
)

// Post represents a Civil Post
type Post interface {
	GetID() string
	GetPostModel() *PostModel
	GetType() string
	GetChannelID() string
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

// StoryfeedFilter contains fields used to filter storyfeed query
type StoryfeedFilter struct {
	Alg       string
	ChannelID *string
}

// PostModel contains fields common to all types of Posts
type PostModel struct {
	ID           string `gorm:"type:uuid;primary_key"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
	ParentID     *string `gorm:"type:uuid;index:idx_post_parent_id"`
	ChannelID    string  `gorm:"type:uuid;index:idx_post_channel_id"`
	AuthorID     string  `gorm:"type:uuid;index:idx_post_author_id;not null"`
	PostType     string  `gorm:"index:idx_post_type"`
	Reference    *string `gorm:"unique_index:idx_post_reference"`
	Data         postgres.Jsonb
	PostPayments []*payments.PaymentModel `gorm:"polymorphic:Owner;"`
}

// TableName returns the gorm table name for Base
func (PostModel) TableName() string {
	return "posts"
}

// GetID returns the ID of the post
func (p PostModel) GetID() string {
	return p.ID
}

// GetPostModel returns itself in order to implement the Post interface
func (p PostModel) GetPostModel() *PostModel {
	return &p
}

// GetChannelID returns the Channel ID of the post
func (p PostModel) GetChannelID() string {
	return p.ChannelID
}

// Boost is a type of Post that is describes an initiative that can be funded
type Boost struct {
	PostModel    `json:"-"`
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
	return TypeBoost
}

// Comment is a type of Post that contains just type
type Comment struct {
	PostModel   `json:"-"`
	Text        string `json:"text"`
	CommentType string `json:"comment_type"`
}

// GetType returns the post type "Comment"
func (b Comment) GetType() string {
	return TypeComment
}

// ExternalLink is a type of Post that links to another web page
type ExternalLink struct {
	PostModel     `json:"-"`
	URL           string     `json:"url"`
	OpenGraphData []byte     `json:"open_graph_data"`
	DatePosted    time.Time  `json:"date_posted"` // DEPRECATED
	PublishedTime *time.Time `json:"published_time"`
}

// GetType returns the post type "ExternalLink"
func (b ExternalLink) GetType() string {
	return TypeExternalLink
}
