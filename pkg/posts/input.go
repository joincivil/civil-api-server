package posts

import "time"

// Paging provides fields needed for pagination
type Paging struct {
	AfterCursor  *string
	BeforeCursor *string
	Limit        *int
	Order        *string
}

// SearchInput provides fields to filter and search for posts
type SearchInput struct {
	PostType     string
	ChannelID    string
	AuthorID     string
	CreatedAfter time.Time
	Paging
}
