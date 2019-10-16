package posts

// PostPersister is an interface for CRUD of Posts
type PostPersister interface {
	GetPost(id string) (Post, error)
	CreatePost(authorID string, post Post) (Post, error)
	EditPost(requestorUserID string, postID string, patch Post) (Post, error)
	DeletePost(requestorUserID string, id string) error
	SearchPosts(search *SearchInput) (*PostSearchResult, error)
	SearchPostsMostRecentPerChannel(search *SearchInput) (*PostSearchResult, error)
}
