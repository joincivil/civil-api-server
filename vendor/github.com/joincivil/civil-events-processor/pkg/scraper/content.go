package scraper

import (
	"errors"

	"github.com/joincivil/civil-events-processor/pkg/model"
)

// ContentScraper is a struct that encapsulates scraping content off the web
// Used to retrieve and store newsroom content
type ContentScraper struct {
}

// ScrapeContent scrapes the content at the given URI and returns it as a
// ScraperContent struct.
func (c *ContentScraper) ScrapeContent(uri string) (*model.ScraperContent, error) {
	return &model.ScraperContent{}, errors.New("Not implemented yet")
}
