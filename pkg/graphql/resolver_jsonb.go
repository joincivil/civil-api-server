package graphql

import (
	context "context"

	log "github.com/golang/glog"
	"github.com/joincivil/civil-api-server/pkg/auth"
	"github.com/joincivil/civil-api-server/pkg/generated/graphql"
	"github.com/joincivil/civil-api-server/pkg/jsonstore"
)

func (r *queryResolver) Jsonb(ctx context.Context, id *string) (*jsonstore.JSONb, error) {
	// Needs to have a valid auth token
	token := auth.ForContext(ctx)
	if token == nil {
		return nil, ErrAccessDenied
	}

	idVal := ""
	if id != nil {
		idVal = *id
	}

	jsonb, err := r.jsonbService.RetrieveJSONb(
		idVal,
		jsonstore.DefaultJsonbGraphqlNs,
		token.Sub, // token sub as salt
	)
	if err != nil {
		return nil, err
	}
	if len(jsonb) > 1 {
		log.Errorf("Warning: More than one JSONb for id: %v", idVal)
	}

	return jsonb[0], nil
}

func (r *mutationResolver) JsonbSave(ctx context.Context, input graphql.JsonbInput) (
	jsonstore.JSONb, error) {
	// Needs to have a valid auth token
	token := auth.ForContext(ctx)
	if token == nil {
		return jsonstore.JSONb{}, ErrAccessDenied
	}

	updatedJSONb, err := r.jsonbService.SaveRawJSONb(
		input.ID,
		jsonstore.DefaultJsonbGraphqlNs,
		token.Sub, // token sub as salt
		input.JSONStr,
		&token.Sub,
	)
	return *updatedJSONb, err
}
