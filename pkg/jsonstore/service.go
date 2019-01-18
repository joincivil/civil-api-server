package jsonstore

import (
	"fmt"
	"time"
)

// NewJsonbService is a convenience function to init a new JSONb Service struct
func NewJsonbService(jsonbPersister JsonbPersister) *Service {
	return &Service{
		jsonbPersister: jsonbPersister,
	}
}

// Service provide methods for the JSONb store
type Service struct {
	jsonbPersister JsonbPersister
}

// RetrieveJSONb retrieves a JSON blob given an ID in a namespace. Namespace
// is required.
func (s *Service) RetrieveJSONb(id string, namespace string) ([]*JSONb, error) {
	if namespace == "" {
		return nil, fmt.Errorf("Namespace is required")
	}

	key, err := NamespacePlusIDHashKey(namespace, id)
	if err != nil {
		return nil, err
	}

	jsonb, err := s.jsonbPersister.RetrieveJsonb(key, "")
	if err != nil {
		return nil, err
	}

	return jsonb, nil
}

// SaveRawJSONb stores a raw JSON string to a key derived from the ID and namespace.
// Namespace is required.
func (s *Service) SaveRawJSONb(id string, namespace string, jsonStr string) (*JSONb, error) {
	if namespace == "" {
		return nil, fmt.Errorf("Namespace is required")
	}

	key, err := NamespacePlusIDHashKey(namespace, id)
	if err != nil {
		return nil, err
	}

	jsonb := &JSONb{}
	jsonb.Key = key
	jsonb.ID = id
	jsonb.CreatedDate = time.Now().UTC()
	jsonb.LastUpdatedDate = time.Now().UTC()
	jsonb.RawJSON = jsonStr

	return s.SaveJSONb(id, namespace, jsonb)
}

// SaveJSONb stores a JSONb struct to a key derived from the ID and namespace.
// Namespace is required.
func (s *Service) SaveJSONb(id string, namespace string, jsonb *JSONb) (*JSONb, error) {
	err := jsonb.ValidateRawJSON()
	if err != nil {
		return jsonb, err
	}

	err = jsonb.HashIDRawJSON()
	if err != nil {
		return jsonb, err
	}

	err = jsonb.RawJSONToFields()
	if err != nil {
		return jsonb, err
	}

	updatedJSONb, err := s.jsonbPersister.SaveJsonb(jsonb)
	return updatedJSONb, err
}
