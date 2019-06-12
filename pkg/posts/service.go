package posts

// Service provides methods to interact with Posts
type Service struct {
	PostPersister
}

// NewService builds an instance of posts.Service
func NewService(persister PostPersister) *Service {
	return &Service{
		PostPersister: persister,
	}
}
