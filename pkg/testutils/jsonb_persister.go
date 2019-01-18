package testutils

import (
	"fmt"

	"github.com/joincivil/civil-api-server/pkg/jsonstore"
)

// InMemoryJSONbPersister is an in memory persister for JSONb for testing.
type InMemoryJSONbPersister struct {
	Store map[string]*jsonstore.JSONb
}

// RetrieveJsonb retrieves JSONb from the inmemory store
func (p *InMemoryJSONbPersister) RetrieveJsonb(id string, hash string) ([]*jsonstore.JSONb, error) {
	jsonb, ok := p.Store[id]
	if !ok {
		return nil, fmt.Errorf("No jsonb found for %v", id)
	}
	return []*jsonstore.JSONb{jsonb}, nil
}

// SaveJsonb stores JSONb to the inmemory store
func (p *InMemoryJSONbPersister) SaveJsonb(jsonb *jsonstore.JSONb) (*jsonstore.JSONb, error) {
	p.Store[jsonb.Key] = jsonb
	return jsonb, nil
}
