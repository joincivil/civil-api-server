package scraper

import (
	"github.com/joincivil/civil-events-processor/pkg/model"
)

// NullScraper is a struct that encapsulates a version of the scrapers
// that perform no action and returns empty data.
// Used for testing or for configurations where no scraping is required.
type NullScraper struct{}

// ScrapeCivilMetadata scrapes the metadata from the Civil article content API at
// the given URI.
func (n *NullScraper) ScrapeCivilMetadata(uri string) (*model.ScraperCivilMetadata, error) {
	return &model.ScraperCivilMetadata{}, nil
}

// ScrapeMetadata scrapes the metadata at the given URI.
func (n *NullScraper) ScrapeMetadata(uri string) (*model.ScraperContentMetadata, error) {
	return &model.ScraperContentMetadata{}, nil
}

// ScrapeContent scrapes the content from a give URI
func (n *NullScraper) ScrapeContent(uri string) (*model.ScraperContent, error) {
	return &model.ScraperContent{}, nil
}
