package posts

// PostPersister is an interface for CRUD of Users
type PostPersister interface {
	GetPost(id string) (Post, error)
	CreatePost(post Post) (Post, error)
	SearchPosts(search *SearchInput) (*PostSearchResult, error)
}
