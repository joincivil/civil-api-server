package scraper

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/joincivil/civil-events-processor/pkg/model"
	"github.com/joincivil/civil-events-processor/pkg/utils"
)

// CharterIPFSScraper scrapes content from an IPFS link for a Civil charter.
type CharterIPFSScraper struct {
}

// ScrapeContent scrapes the IPFS charter content at the given URI and returns it as a
// ScraperContent struct.
// TODO(PN): The right way to do this is to return a Charter object.  Move from
// api-server to processor.  For now, just return a generic ScraperContent with payload
// in data.
func (c *CharterIPFSScraper) ScrapeContent(uri string) (*model.ScraperContent, error) {
	bys, err := utils.RetrieveIPFSLink(uri)
	if err != nil {
		return nil, err
	}
	if len(bys) == 0 {
		return nil, fmt.Errorf("No []byte returned")
	}

	var data map[string]interface{}
	err = json.Unmarshal(bys, &data)
	if err != nil {
		return nil, err
	}

	return model.NewScraperContent("", "", uri, "", data), nil
}

// ContentScraper is a struct that encapsulates scraping content off the web
// Used to retrieve and store newsroom content
type ContentScraper struct {
}

// ScrapeContent scrapes the content at the given URI and returns it as a
// ScraperContent struct.
func (c *ContentScraper) ScrapeContent(uri string) (*model.ScraperContent, error) {
	return &model.ScraperContent{}, errors.New("Not implemented yet")
}
