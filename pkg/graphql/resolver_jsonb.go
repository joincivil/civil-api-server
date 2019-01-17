package graphql

import (
	context "context"
	"fmt"

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

	jsonb, err := r.jsonbService.RetrieveJSONb(idVal, token.Sub)
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

	updatedJSONb, err := r.jsonbService.SaveRawJSONb(input.ID, token.Sub, input.JSONStr)
	return *updatedJSONb, err
}
