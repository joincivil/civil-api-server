package graphql

import (
	context "context"
	"strconv"

	"github.com/iancoleman/strcase"

	eventModel "github.com/joincivil/civil-events-processor/pkg/model"
	newsroomModel "github.com/joincivil/civil-api-server/pkg/nrsignup"
	"github.com/joincivil/go-common/pkg/eth"
	"github.com/joincivil/civil-api-server/pkg/auth"

	"github.com/joincivil/civil-api-server/pkg/generated/graphql"
)

// ContentRevision is the resolver for the ContentRevision type
func (r *Resolver) ContentRevision() graphql.ContentRevisionResolver {
	return &contentRevisionResolver{r}
}

// TYPE RESOLVERS

type contentRevisionResolver struct{ *Resolver }

func (r *contentRevisionResolver) ListingAddress(ctx context.Context, obj *eventModel.ContentRevision) (string, error) {
	return r.Resolver.DetermineAddrCase(obj.ListingAddress().Hex()), nil
}
func (r *contentRevisionResolver) Payload(ctx context.Context, obj *eventModel.ContentRevision) ([]graphql.ArticlePayload, error) {
	data := []graphql.ArticlePayload{}
	for key, val := range obj.Payload() {
		meta := graphql.ArticlePayload{
			// Make the key lower camel case for consistency with GraphQL field names
			Key:   strcase.ToLowerCamel(key),
			Value: eventModel.ArticlePayloadValue{Value: val},
		}
		data = append(data, meta)
	}
	return data, nil
}
func (r *contentRevisionResolver) EditorAddress(ctx context.Context, obj *eventModel.ContentRevision) (string, error) {
	return r.Resolver.DetermineAddrCase(obj.ListingAddress().Hex()), nil
}
func (r *contentRevisionResolver) ContractContentID(ctx context.Context, obj *eventModel.ContentRevision) (int, error) {
	bigInt := obj.ContractContentID()
	return int(bigInt.Int64()), nil
}
func (r *contentRevisionResolver) ContractRevisionID(ctx context.Context, obj *eventModel.ContentRevision) (int, error) {
	bigInt := obj.ContractContentID()
	return int(bigInt.Int64()), nil
}
func (r *contentRevisionResolver) RevisionDate(ctx context.Context, obj *eventModel.ContentRevision) (int, error) {
	return int(obj.RevisionDateTs()), nil
}

// QUERIES

func (r *queryResolver) Articles(ctx context.Context, addr *string, first *int,
	after *string, lowercaseAddr *bool) ([]eventModel.ContentRevision, error) {
	r.Resolver.lowercaseAddr = lowercaseAddr
	return r.NewsroomArticles(ctx, addr, first, after, lowercaseAddr)
}

func (r *queryResolver) NewsroomArticles(ctx context.Context, addr *string, first *int,
	after *string, lowercaseAddr *bool) ([]eventModel.ContentRevision, error) {
	r.Resolver.lowercaseAddr = lowercaseAddr
	criteria := &eventModel.ContentRevisionCriteria{
		LatestOnly: true,
	}
	if addr != nil && *addr != "" {
		criteria.ListingAddress = eth.NormalizeEthAddress(*addr)
	}
	if after != nil && *after != "" {
		afterInt, err := strconv.Atoi(*after)
		if err != nil {
			return nil, err
		}
		criteria.Offset = afterInt
	}
	if first != nil {
		criteria.Count = *first
	}

	revisions, err := r.revisionPersister.ContentRevisionsByCriteria(criteria)
	if err != nil {
		return nil, err
	}

	modelRevisions := make([]eventModel.ContentRevision, len(revisions))
	for index, revision := range revisions {
		modelRevisions[index] = *revision
	}
	return modelRevisions, nil
}

func (r *queryResolver) Newsroom(ctx context.Context) (*newsroomModel.SignupUserJSONData, error) {
	token := auth.ForContext(ctx)
	if token == nil {
		return nil, ErrAccessDenied
	}

	newsroom, err := r.nrsignupService.RetrieveUserJSONData(token.Sub)

	if err != nil {
		return nil, err
	}

	return newsroom, nil
}
