package posts

import (
	"github.com/goware/urlx"
	"github.com/joincivil/civil-api-server/pkg/channels"
	"github.com/joincivil/civil-api-server/pkg/newsrooms"
)

// Service provides methods to interact with Posts
type Service struct {
	PostPersister
	channelService  *channels.Service
	newsroomService newsrooms.Service
}

// NewService builds an instance of posts.Service
func NewService(persister PostPersister, channelSer *channels.Service, newsroomSer newsrooms.Service) *Service {
	return &Service{
		PostPersister:   persister,
		channelService:  channelSer,
		newsroomService: newsroomSer,
	}
}

// CreatePost creates a new Post, with business logic ensuring posts are correct, and follow certain rules
func (s *Service) CreatePost(authorID string, post Post) (Post, error) {
	base, err := PostInterfaceToBase(post)
	if err != nil {
		return nil, err
	}
	postType := base.PostType
	if postType == "boost" {
		return s.PostPersister.CreatePost(authorID, post)
	} else if postType == "externallink" {

		channel, err := s.channelService.GetChannel(base.ChannelID)
		if err != nil {
			return nil, err
		}
		channelType := channel.ChannelType
		if channelType == "newsroom" {
			externallink := post.(ExternalLink)
			submittedURL, err := urlx.Parse(externallink.URL)
			if err != nil {
				return nil, err
			}
			submittedHost, _, err := urlx.SplitHostPort(submittedURL)
			if err != nil {
				return nil, err
			}

			newsroom, err := s.newsroomService.GetNewsroomByAddress(channel.Reference)
			if err != nil {
				return nil, err
			}
			channelURL, err := urlx.Parse(newsroom.Charter.NewsroomURL)
			if err != nil {
				return nil, err
			}
			channelHost, _, err := urlx.SplitHostPort(channelURL)
			if err != nil {
				return nil, err
			}

			if channelHost != submittedHost {
				return nil, ErrorBadURLSubmitted
			}
			return s.PostPersister.CreatePost(authorID, post)
		}
		return nil, ErrorNotImplemented
	}
	return nil, nil
}
