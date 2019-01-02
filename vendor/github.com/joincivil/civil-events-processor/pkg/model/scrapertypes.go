// Package model contains the general data models and interfaces for the Civil processor.
package model // import "github.com/joincivil/civil-events-processor/pkg/model"

import (
	"encoding/json"

	log "github.com/golang/glog"
)

// MetadataScraper is the interface for implementations of metadata scraper
// Provides a generic interface for writing implementations of fetching metadata
// from non-Civil sources.
type MetadataScraper interface {
	ScrapeMetadata(uri string) (*ScraperContentMetadata, error)
}

// CivilMetadataScraper is the interface for implementations of Civil-specific metadata scraper
type CivilMetadataScraper interface {
	ScrapeCivilMetadata(uri string) (*ScraperCivilMetadata, error)
}

// ContentScraper is the interface for implementations of content scraper
type ContentScraper interface {
	ScrapeContent(uri string) (*ScraperContent, error)
}

// ScraperContentMetadata represents metadata for the scraped content
// Potentially retrieved from a different location than the content and generally
// in JSON format
type ScraperContentMetadata map[string]interface{}

// NewScraperCivilMetadata is a convenience function to create a new
// ScraperCivilMetadata struct. Should use this to ensure the
// internal struct is initialized.
func NewScraperCivilMetadata() *ScraperCivilMetadata {
	return &ScraperCivilMetadata{
		metadata: &scraperCivilMetadata{},
	}
}

// ScraperCivilMetadata represents metadata specifically from the Civil
// article content API
type ScraperCivilMetadata struct {
	metadata *scraperCivilMetadata
}

// MarshalJSON converts this struct into a []byte
func (s *ScraperCivilMetadata) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.metadata)
}

// UnmarshalJSON populates the struct with given []byte.
func (s *ScraperCivilMetadata) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, s.metadata)
}

// Title returns the title field in the metadata
func (s *ScraperCivilMetadata) Title() string {
	return s.metadata.Title
}

// RevisionContentHash returns the revision content hash field in the metadata
func (s *ScraperCivilMetadata) RevisionContentHash() string {
	return s.metadata.RevisionContentHash
}

// RevisionContentURL returns the revision content url field in the metadata
func (s *ScraperCivilMetadata) RevisionContentURL() string {
	return s.metadata.RevisionContentURL
}

// CanonicalURL returns the canonical url field in the metadata
func (s *ScraperCivilMetadata) CanonicalURL() string {
	return s.metadata.CanonicalURL
}

// Slug returns the slug field in the metadata
func (s *ScraperCivilMetadata) Slug() string {
	return s.metadata.Slug
}

// Description returns the description field in the metadata
func (s *ScraperCivilMetadata) Description() string {
	return s.metadata.Description
}

// PrimaryTag returns the primary tag field in the metadata
func (s *ScraperCivilMetadata) PrimaryTag() string {
	return s.metadata.PrimaryTag
}

// RevisionDate returns the revision date field in the metadata
func (s *ScraperCivilMetadata) RevisionDate() string {
	return s.metadata.RevisionDate
}

// OriginalPublishDate returns the original published date field in the metadata
func (s *ScraperCivilMetadata) OriginalPublishDate() string {
	return s.metadata.OriginalPublishDate
}

// Opinion returns the opinion field in the metadata
func (s *ScraperCivilMetadata) Opinion() bool {
	return s.metadata.Opinion
}

// SchemaVersion returns the schema version field in the metadata
func (s *ScraperCivilMetadata) SchemaVersion() string {
	return s.metadata.SchemaVersion
}

// Contributors returns the schema contributors field in the metadata
func (s *ScraperCivilMetadata) Contributors() []*ScraperCivilMetadataContributor {
	contributors := []*ScraperCivilMetadataContributor{}
	for _, contributor := range s.metadata.Contributors {
		contributors = append(contributors, &ScraperCivilMetadataContributor{contributor: contributor})
	}
	return contributors
}

// Images returns the images field in the metadata
func (s *ScraperCivilMetadata) Images() []*ScraperCivilMetadataImage {
	images := []*ScraperCivilMetadataImage{}
	for _, image := range s.metadata.Images {
		images = append(images, &ScraperCivilMetadataImage{image: image})
	}
	return images
}

// CredibilityIndicators returns the credibility indicators field in the metadata
func (s *ScraperCivilMetadata) CredibilityIndicators() *ScraperCivilMetadataCredibility {
	switch val := s.metadata.CredIndicators.(type) {
	case map[string]interface{}:
		bys, err := json.Marshal(val)
		if err == nil {
			md := &scraperCivilMetadataCredibility{}
			err := json.Unmarshal(bys, md)
			if err == nil {
				return &ScraperCivilMetadataCredibility{cred: md}
			}
		}
	}
	log.Errorf("Returning an empty set of credibility indicators")
	return &ScraperCivilMetadataCredibility{cred: &scraperCivilMetadataCredibility{}}
}

type scraperCivilMetadata struct {
	Title               string                             `json:"title"`
	RevisionContentHash string                             `json:"revisionContentHash"`
	RevisionContentURL  string                             `json:"revisionContentUrl"`
	CanonicalURL        string                             `json:"canonicalUrl"`
	Slug                string                             `json:"slug"`
	Description         string                             `json:"description"`
	Contributors        []*scraperCivilMetadataContributor `json:"contributors"`
	Images              []*scraperCivilMetadataImage       `json:"images"`
	Tags                []string                           `json:"tags"`
	PrimaryTag          string                             `json:"primaryTag"`
	RevisionDate        string                             `json:"revisionDate"`
	OriginalPublishDate string                             `json:"originalPublishDate"`
	CredIndicators      interface{}                        `json:"credibilityIndicators"`
	Opinion             bool                               `json:"opinion"`
	SchemaVersion       string                             `json:"civilSchemaVersion"`
}

// ScraperCivilMetadataContributor represents a contributor in the Civil article metadata
type ScraperCivilMetadataContributor struct {
	contributor *scraperCivilMetadataContributor
}

// Role returns the role for this contributor
func (s *ScraperCivilMetadataContributor) Role() string {
	return s.contributor.Role
}

// Name returns the name for this contributor
func (s *ScraperCivilMetadataContributor) Name() string {
	return s.contributor.Name
}

// Address returns the address for this contributor
func (s *ScraperCivilMetadataContributor) Address() string {
	return s.contributor.Address
}

// Signature returns the signature for this contributor
func (s *ScraperCivilMetadataContributor) Signature() string {
	return s.contributor.Signature
}

type scraperCivilMetadataContributor struct {
	Role      string `json:"role"`
	Name      string `json:"name"`
	Address   string `json:"address"`
	Signature string `json:"signature"`
}

// ScraperCivilMetadataImage represents an image in the Civil article metadata
type ScraperCivilMetadataImage struct {
	image *scraperCivilMetadataImage
}

// URL returns the url for this image
func (s *ScraperCivilMetadataImage) URL() string {
	return s.image.URL
}

// Hash returns the hash for this image
func (s *ScraperCivilMetadataImage) Hash() string {
	return s.image.Hash
}

// Height returns the height of this image
func (s *ScraperCivilMetadataImage) Height() int {
	return s.image.Height
}

// Width returns the width of this image
func (s *ScraperCivilMetadataImage) Width() int {
	return s.image.Width
}

type scraperCivilMetadataImage struct {
	URL    string `json:"url"`
	Hash   string `json:"hash"`
	Height int    `json:"h"`
	Width  int    `json:"w"`
}

// ScraperCivilMetadataCredibility represents a credibility indicator from the
// Civil article metadata
type ScraperCivilMetadataCredibility struct {
	cred *scraperCivilMetadataCredibility
}

// OriginalReporting returns the value in the original reporting field for
// credibility
func (c *ScraperCivilMetadataCredibility) OriginalReporting() bool {
	switch t := c.cred.OriginalReporting.(type) {
	case string:
		if t == "true" || t == "1" {
			return true
		}
	case bool:
		return t
	}
	return false
}

// OnTheGround returns the value of the on the ground field for credibility
func (c *ScraperCivilMetadataCredibility) OnTheGround() bool {
	switch t := c.cred.OnTheGround.(type) {
	case string:
		if t == "true" || t == "1" {
			return true
		}
	case bool:
		return t
	}
	return false
}

// SourcesCited returns the value of the sources cited field for credibility
func (c *ScraperCivilMetadataCredibility) SourcesCited() bool {
	switch t := c.cred.SourcesCited.(type) {
	case string:
		if t == "true" || t == "1" {
			return true
		}
	case bool:
		return t
	}
	return false
}

// SubjectSpecialist returns the value of the subject specialist field for credibility
func (c *ScraperCivilMetadataCredibility) SubjectSpecialist() bool {
	switch t := c.cred.SubjectSpecialist.(type) {
	case string:
		if t == "true" || t == "1" {
			return true
		}
	case bool:
		return t
	}
	return false
}

type scraperCivilMetadataCredibility struct {
	OriginalReporting interface{} `json:"original_reporting"`
	OnTheGround       interface{} `json:"on_the_ground"`
	SourcesCited      interface{} `json:"sources_cited"`
	SubjectSpecialist interface{} `json:"subject_specialist"`
}

// NewScraperContent is a convenience function to init a new ScraperContent struct
func NewScraperContent(text string, html string, uri string, author string,
	data map[string]interface{}) *ScraperContent {
	return &ScraperContent{
		text:   text,
		html:   html,
		uri:    uri,
		author: author,
		data:   data,
	}
}

// ScraperContent represents the scraped content data
type ScraperContent struct {
	text   string
	html   string
	uri    string
	author string
	data   map[string]interface{}
}

// URI returns the URI to the content
func (c *ScraperContent) URI() string {
	return c.uri
}

// HTML returns the raw HTML of the content
func (c *ScraperContent) HTML() string {
	return c.html
}

// Text returns the plain text of the content
func (c *ScraperContent) Text() string {
	return c.text
}

// Author returns the plain text name of the author if found.
func (c *ScraperContent) Author() string {
	return c.author
}

// Data returns any additional data for the content
func (c *ScraperContent) Data() map[string]interface{} {
	return c.data
}
