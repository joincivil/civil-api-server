package posts

import (
	"github.com/dyatlov/go-htmlinfo/htmlinfo"
	"github.com/joincivil/civil-api-server/pkg/channels"
	"github.com/joincivil/civil-api-server/pkg/newsrooms"
	"github.com/joincivil/civil-events-processor/pkg/utils"
	"net/http"
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

// CreateExternalLinkEmbedded creates a new Post, with business logic ensuring posts are correct, and follow certain rules
func (s *Service) CreateExternalLinkEmbedded(post Post) (Post, error) {
	base, err := PostInterfaceToBase(post)
	if err != nil {
		return nil, err
	}
	postType := base.PostType
	if postType == TypeExternalLink {
		externalLink, err := s.getExternalLink(post)
		if err != nil {
			return nil, err
		}

		return s.PostPersister.CreatePost(externalLink.ChannelID, *externalLink)
	}
	return nil, ErrorNotImplemented
}

// CreatePost creates a new Post, with business logic ensuring posts are correct, and follow certain rules
func (s *Service) CreatePost(authorID string, post Post) (Post, error) {
	base, err := PostInterfaceToBase(post)
	if err != nil {
		return nil, err
	}
	postType := base.PostType
	if postType == TypeBoost {
		return s.PostPersister.CreatePost(authorID, post)
	} else if postType == TypeExternalLink {
		externalLink, err := s.getExternalLink(post)
		if err != nil {
			return nil, err
		}

		return s.PostPersister.CreatePost(authorID, *externalLink)
	}
	return nil, nil
}

func (s *Service) getExternalLink(post Post) (*ExternalLink, error) {

	base, err := PostInterfaceToBase(post)
	if err != nil {
		return nil, err
	}
	channel, err := s.channelService.GetChannel(base.ChannelID)
	if err != nil {
		return nil, err
	}
	channelType := channel.ChannelType
	if channelType == channels.TypeNewsroom {
		externalLink := post.(ExternalLink)
		cleanedSubmittedURL, err := utils.CleanURL(externalLink.URL)
		if err != nil {
			return nil, err
		}

		newsroom, err := s.newsroomService.GetNewsroomByAddress(channel.Reference)
		if err != nil {
			return nil, err
		}
		cleanedChannelURL, err := utils.CleanURL(newsroom.Charter.NewsroomURL)
		if err != nil {
			return nil, err
		}

		if cleanedChannelURL != cleanedSubmittedURL {
			return nil, ErrorBadURLSubmitted
		}

		resp, err := http.Get(externalLink.URL)

		if err != nil {
			return nil, err
		}

		defer resp.Body.Close() // nolint: errcheck

		htmlInfo := htmlinfo.NewHTMLInfo()

		err = htmlInfo.Parse(resp.Body, &(externalLink.URL), nil)
		if err != nil {
			return nil, err
		}

		ref := TypeExternalLink + "+" + channel.Reference + "+" + htmlInfo.CanonicalURL
		externalLink.Reference = &ref

		ogJSON, err := htmlInfo.OGInfo.ToJSON()
		if err != nil {
			return nil, err
		}
		externalLink.OpenGraphData = ogJSON
		return &externalLink, nil
	}
	return nil, ErrorNotImplemented
}
