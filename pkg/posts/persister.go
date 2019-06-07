package posts

import (
	"github.com/joincivil/civil-api-server/pkg/payments"
)

// PostPersister is an interface for CRUD of Posts
type PostPersister interface {
	GetPost(id string) (Post, error)
	CreatePost(post Post) (Post, error)
	SearchPosts(search *SearchInput) (*PostSearchResult, error)
	GetPayments(post Post) ([]payments.Payment, error)
}
