package jsonstore

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/golang/glog"
	"io"
	"strconv"
	"time"

	processorutils "github.com/joincivil/civil-events-processor/pkg/utils"
)

var (
	// ErrNoJsonbFound indicates when a JSONb is not found for an ID
	ErrNoJsonbFound = errors.New("No jsonb found")
)

// JSONFieldValue represents a value in a JsonField
type JSONFieldValue struct {
	Value interface{}
}

// UnmarshalGQL implements the graphql.Unmarshaler interface
func (a *JSONFieldValue) UnmarshalGQL(v interface{}) error {
	a.Value = v
	return nil
}

// MarshalGQL implements the graphql.Marshaler interface
func (a JSONFieldValue) MarshalGQL(w io.Writer) {
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

// JSONField represents a field in a json
type JSONField struct {
	Key   string          `json:"key"`
	Value *JSONFieldValue `json:"value"`
}

// JSONb represents a JSON blob to be stored and retrieved
type JSONb struct {
	ID          string       `json:"id"`
	Hash        string       `json:"hash"`
	CreatedDate time.Time    `json:"createdDate"`
	RawJSON     string       `json:"rawJson"`
	JSON        []*JSONField `json:"json"`
}

// ValidateRawJSON validates the JSON string to ensure it is of the correct
// JSON format
func (j *JSONb) ValidateRawJSON() error {
	var js json.RawMessage
	err := json.Unmarshal([]byte(j.RawJSON), &js)
	if err != nil {
		return fmt.Errorf("Invalid Raw JSON: err: %v", err)
	}
	return nil
}

// RawJSONToFields converts the RawJSON string into the JSONField slice
func (j *JSONb) RawJSONToFields() error {
	if j.RawJSON == "" {
		return errors.New("No RawJSON to hash")
	}
	err := j.ValidateRawJSON()
	if err != nil {
		return err
	}
	js := map[string]interface{}{}
	err = json.Unmarshal([]byte(j.RawJSON), &js)
	if err != nil {
		return err
	}
	jsonFields := []*JSONField{}
	for key, val := range js {
		jsonFields = append(jsonFields, &JSONField{
			Key: key,
			Value: &JSONFieldValue{
				Value: val,
			},
		})
	}
	j.JSON = jsonFields
	return nil
}

// HashIDRawJSON takes the RawJSON and ID, creates a hash of it, and sets
// it as the Hash field
func (j *JSONb) HashIDRawJSON() error {
	if j.RawJSON == "" {
		return errors.New("No RawJSON to hash")
	}
	err := j.ValidateRawJSON()
	if err != nil {
		return err
	}
	createDateSecs := processorutils.TimeToSecsFromEpoch(&j.CreatedDate)
	strToHash := fmt.Sprintf("%v|%v|%v", j.ID, j.RawJSON, createDateSecs)
	h := sha256.New()
	_, err = h.Write([]byte(strToHash))
	if err != nil {
		return err
	}
	j.Hash = hex.EncodeToString(h.Sum(nil))
	return nil
}

// JsonbPersister defines the interface for a Jsonb persister.
type JsonbPersister interface {

	// RetrieveJsonb retrieves a slice of populated Jsonb objects or
	// an error.  If no results, returns ErrNoJsonbFound.
	RetrieveJsonb(id string, hash string) ([]*JSONb, error)

	// SaveJsonb saves a populated Jsonb object or returns an error
	SaveJsonb(jsonb *JSONb) error
}
