package jsonstore

import (
	"fmt"
)

// InMemoryJSONbPersister is an in memory persister for JSONb for testing.
type InMemoryJSONbPersister struct {
	Store map[string]*JSONb
}

// RetrieveJsonb retrieves JSONb from the in memory store
func (p *InMemoryJSONbPersister) RetrieveJsonb(id string, hash string) ([]*JSONb, error) {
	jsonb, ok := p.Store[id]
	if !ok {
		return nil, fmt.Errorf("No jsonb found for %v", id)
	}
	return []*JSONb{jsonb}, nil
}

// SaveJsonb stores JSONb to the inmemory store
func (p *InMemoryJSONbPersister) SaveJsonb(jsonb *JSONb) (*JSONb, error) {
	p.Store[jsonb.Key] = jsonb
	return jsonb, nil
}
