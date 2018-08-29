// Package model contains the general data models and interfaces for the Civil processor.
package model // import "github.com/joincivil/civil-events-processor/pkg/model"

import (
	"encoding/json"
	"fmt"
	log "github.com/golang/glog"
	"io"
	"math/big"
	"sort"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
)

// ArticlePayload is the metadata and content data for an article
type ArticlePayload map[string]interface{}

// Hash returns the hash of the article payload.  Hashes
// all the values from the map together as slice of keyvalue pairs.
// Returns a keccak256 hash hex string.
// NOTE(PN): Currently using the contentHash from the newsroom contract
// for the hash.  This is here as an alterative if needed.
func (a ArticlePayload) Hash() string {
	toEncode := make([]string, len(a))
	index := 0
	for key, val := range a {
		hashPart := fmt.Sprintf("%v%v", key, val)
		toEncode[index] = hashPart
		index++
	}
	sort.Strings(toEncode)
	eventBytes, _ := rlp.EncodeToBytes(toEncode) // nolint: gosec
	h := crypto.Keccak256Hash(eventBytes)
	return h.Hex()
}

// ArticlePayloadValue is for serializing to either a string or
// a map of values for GraphQL.  Used as a scalar for GraphQL
type ArticlePayloadValue struct {
	Value interface{}
}

// UnmarshalGQL implements the graphql.Unmarshaler interface
func (a *ArticlePayloadValue) UnmarshalGQL(v interface{}) error {
	a.Value = v
	return nil
}

// MarshalGQL implements the graphql.Marshaler interface
func (a ArticlePayloadValue) MarshalGQL(w io.Writer) {
	switch val := a.Value.(type) {
	case bool:
		_, err := fmt.Fprintf(w, "%v", strconv.FormatBool(val))
		if err != nil {
			log.Errorf("Error writing gql int: err %v", err)
		}
	case float64:
		_, err := fmt.Fprintf(w, "%v", val)
		if err != nil {
			log.Errorf("Error writing gql float: err %v", err)
		}
	case int:
		_, err := fmt.Fprintf(w, "%v", val)
		if err != nil {
			log.Errorf("Error writing gql int: err %v", err)
		}
	case string:
		_, err := fmt.Fprintf(w, "\"%v\"", val)
		if err != nil {
			log.Errorf("Error writing gql string: err %v", err)
		}
	default:
		bytes, err := json.Marshal(val)
		if err != nil {
			log.Errorf("Error marshaling map: err %v", err)
		}
		_, err = fmt.Fprintf(w, "%v", string(bytes))
		if err != nil {
			log.Errorf("Error writing gql map: err %v", err)
		}
	}
}

// NewContentRevision is a convenience function to init a ContentRevision
// struct
func NewContentRevision(listingAddr common.Address, payload ArticlePayload, payloadHash string,
	editorAddress common.Address, contractContentID *big.Int, contractRevisionID *big.Int,
	revisionURI string, revisionDateTs int64) *ContentRevision {
	revision := &ContentRevision{
		listingAddress:     listingAddr,
		payload:            payload,
		payloadHash:        payloadHash,
		editorAddress:      editorAddress,
		contractContentID:  contractContentID,
		contractRevisionID: contractRevisionID,
		revisionURI:        revisionURI,
		revisionDateTs:     revisionDateTs,
	}
	return revision
}

// ContentRevision represents a revision to a content item
type ContentRevision struct {
	listingAddress common.Address

	payload ArticlePayload

	payloadHash string

	editorAddress common.Address

	contractContentID *big.Int

	contractRevisionID *big.Int

	revisionURI string

	revisionDateTs int64
}

// ListingAddress returns the associated listing address
func (c *ContentRevision) ListingAddress() common.Address {
	return c.listingAddress
}

// EditorAddress returns the address of editor who made revision
func (c *ContentRevision) EditorAddress() common.Address {
	return c.editorAddress
}

// Payload returns the ArticlePayload
func (c *ContentRevision) Payload() ArticlePayload {
	return c.payload
}

// PayloadHash returns the hash of the payload
func (c *ContentRevision) PayloadHash() string {
	return c.payloadHash
}

// RevisionURI returns the revision URI
func (c *ContentRevision) RevisionURI() string {
	return c.revisionURI
}

// ContractContentID returns the contract content ID
func (c *ContentRevision) ContractContentID() *big.Int {
	return c.contractContentID
}

// ContractRevisionID returns the contract content revision ID
func (c *ContentRevision) ContractRevisionID() *big.Int {
	return c.contractRevisionID
}

// RevisionDateTs returns the timestamp of the revision
func (c *ContentRevision) RevisionDateTs() int64 {
	return c.revisionDateTs
}
