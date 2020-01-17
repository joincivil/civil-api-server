package posts

import (
	"github.com/dyatlov/go-htmlinfo/htmlinfo"
	"github.com/joincivil/civil-api-server/pkg/channels"
	"github.com/joincivil/civil-api-server/pkg/newsrooms"
	"github.com/joincivil/civil-events-processor/pkg/utils"
	"golang.org/x/net/html"
	"net/http"
	"strings"
	"time"
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

// GetPostByReferenceSafe returns a post associated with the provided reference
// cleans reference before checking to avoid "http://" vs "https://" issue
func (s *Service) GetPostByReferenceSafe(reference string) (Post, error) {
	cleanReference := strings.Replace(reference, TypeExternalLink+"+http://", TypeExternalLink+"+", 1)
	cleanReference = strings.Replace(cleanReference, TypeExternalLink+"+https://", TypeExternalLink+"+", 1)

	post, err := s.GetPostByReference(cleanReference)

	// TODO: remove this once existing posts are updated
	// backwards compatibility
	if post == nil || err != nil {
		post, err = s.GetPostByReference(reference)
	}

	return post, err
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

		// Just get this twice since we need 2 readers. Trying to duplicate readers is complicated and doesn't play nicely with htmlInfo parsing
		resp2, err := http.Get(externalLink.URL)
		if err != nil {
			return nil, err
		}
		defer resp2.Body.Close() // nolint: errcheck

		htmlInfo := htmlinfo.NewHTMLInfo()

		err = htmlInfo.Parse(resp.Body, &(externalLink.URL), nil)
		if err != nil {
			return nil, err
		}

		var refURL string
		if htmlInfo.CanonicalURL != "" {
			refURL = htmlInfo.CanonicalURL
		} else if htmlInfo.OGInfo.URL != "" {
			refURL = htmlInfo.OGInfo.URL
		} else {
			return nil, ErrorNoReferenceURLFound
		}

		refURL = strings.Replace(refURL, "https://", "", 1)
		refURL = strings.Replace(refURL, "http://", "", 1)

		ref := TypeExternalLink + "+" + refURL
		externalLink.Reference = &ref
		didFindCivilPublishedTime := false
		z := html.NewTokenizer(resp2.Body)
		timePublished, err := getPublishedTime(z)
		if err != nil {
			return nil, err
		}
		if timePublished != nil {
			externalLink.PublishedTime = timePublished
			didFindCivilPublishedTime = true
		}

		if htmlInfo.OGInfo != nil {
			ogJSON, err := htmlInfo.OGInfo.ToJSON()
			if err != nil {
				return nil, err
			}
			externalLink.OpenGraphData = ogJSON

			if !didFindCivilPublishedTime && htmlInfo.OGInfo.Article != nil && htmlInfo.OGInfo.Article.PublishedTime != nil {
				time := htmlInfo.OGInfo.Article.PublishedTime
				externalLink.PublishedTime = time
			}
		}

		return &externalLink, nil
	}
	return nil, ErrorNotImplemented
}

func getPublishedTime(z *html.Tokenizer) (*time.Time, error) {
	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			return nil, nil
		case html.StartTagToken, html.SelfClosingTagToken:
			t := z.Token()
			if t.Data == "meta" {
				civilPublishedTime, ok := extractMetaProperty(t, "civil:published_time")
				if ok {
					timePublished, err := time.Parse("2006-01-02T15:04:05-0700", civilPublishedTime)
					if err != nil {
						return nil, err
					}
					return &timePublished, nil
				}
			}
		}
	}
}

func extractMetaProperty(t html.Token, prop string) (content string, ok bool) {
	for _, attr := range t.Attr {
		if attr.Key == "property" && attr.Val == prop {
			ok = true
		}

		if attr.Key == "content" {
			content = attr.Val
		}
	}

	return
}
