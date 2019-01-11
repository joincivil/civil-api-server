package graphql

import (
	context "context"
	"fmt"
	"time"

	"github.com/joincivil/civil-api-server/pkg/auth"
	"github.com/joincivil/civil-api-server/pkg/generated/graphql"
	"github.com/joincivil/civil-api-server/pkg/jsonstore"
)

func (r *queryResolver) Jsonb(ctx context.Context, id *string) ([]*jsonstore.JSONb, error) {
	// Needs to have a valid auth token
	token := auth.ForContext(ctx)
	if token == nil {
		return nil, fmt.Errorf("Access denied")
	}

	idVal := ""
	if id != nil {
		idVal = *id
	}

	key, err := jsonstore.TokenPlusIDHashKey(token, idVal)
	if err != nil {
		return nil, err
	}

	jsonb, err := r.jsonbPersister.RetrieveJsonb(key, "")
	if err != nil {
		return nil, err
	}
	return jsonb, nil
}

func (r *mutationResolver) JsonbSave(ctx context.Context, input graphql.JsonbInput) (
	jsonstore.JSONb, error) {
	// Needs to have a valid auth token
	token := auth.ForContext(ctx)
	if token == nil {
		return jsonstore.JSONb{}, fmt.Errorf("Access denied")
	}

	jsonb := jsonstore.JSONb{}
	key, err := jsonstore.TokenPlusIDHashKey(token, input.ID)
	if err != nil {
		return jsonb, err
	}

	jsonb.Key = key
	jsonb.ID = input.ID
	jsonb.CreatedDate = time.Now().UTC()
	jsonb.LastUpdatedDate = time.Now().UTC()
	jsonb.RawJSON = input.JSONStr

	err = jsonb.ValidateRawJSON()
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

	updatedJSONb, err := r.jsonbPersister.SaveJsonb(&jsonb)
	return *updatedJSONb, err
}
