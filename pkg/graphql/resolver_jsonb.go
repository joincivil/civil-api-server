package graphql

import (
	context "context"
	"time"

	"github.com/joincivil/civil-api-server/pkg/generated/graphql"
	"github.com/joincivil/civil-api-server/pkg/jsonstore"
)

func (r *queryResolver) Jsonb(ctx context.Context, id *string) ([]*jsonstore.JSONb, error) {
	idVal := ""
	if id != nil {
		idVal = *id
	}
	jsonb, err := r.jsonbPersister.RetrieveJsonb(idVal, "")
	if err != nil {
		return nil, err
	}
	return jsonb, nil
}

func (r *mutationResolver) JsonbSave(ctx context.Context, input graphql.JsonbInput) (jsonstore.JSONb, error) {
	jsonb := jsonstore.JSONb{}
	jsonb.ID = input.ID
	jsonb.CreatedDate = time.Now().UTC()
	jsonb.LastUpdatedDate = time.Now().UTC()
	jsonb.RawJSON = input.JSONStr

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

	updatedJSONb, err := r.jsonbPersister.SaveJsonb(&jsonb)
	return *updatedJSONb, err
}
