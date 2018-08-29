package scraper

import (
	"io/ioutil"
	"net/http"
	"time"

	"github.com/joincivil/civil-events-processor/pkg/model"
)

const (
	timeoutSecs = 2
)

// CivilMetadataScraper is a struct that encapsulates scraping the Civil article content API
// metadata. Implements the CivilMetadataScraper interface.
type CivilMetadataScraper struct{}

// ScrapeCivilMetadata scrapes the metadata from the Civil article content API at
// the given URI.
func (m *CivilMetadataScraper) ScrapeCivilMetadata(uri string) (*model.ScraperCivilMetadata, error) {
	timeout := timeoutSecs * time.Second
	client := http.Client{
		Timeout: timeout,
	}
	resp, err := client.Get(uri)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	metadata := model.NewScraperCivilMetadata()
	err = metadata.UnmarshalJSON(body)
	if err != nil {
		return nil, err
	}
	return metadata, nil
}
