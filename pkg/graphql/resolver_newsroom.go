package graphql

import (
	context "context"
	"strconv"

	"github.com/iancoleman/strcase"

	model "github.com/joincivil/civil-events-processor/pkg/model"
	"github.com/joincivil/go-common/pkg/eth"

	"github.com/joincivil/civil-api-server/pkg/generated/graphql"
)

// ContentRevision is the resolver for the ContentRevision type
func (r *Resolver) ContentRevision() graphql.ContentRevisionResolver {
	return &contentRevisionResolver{r}
}

// TYPE RESOLVERS

type contentRevisionResolver struct{ *Resolver }

func (r *contentRevisionResolver) ListingAddress(ctx context.Context, obj *model.ContentRevision) (string, error) {
	return r.Resolver.DetermineAddrCase(obj.ListingAddress().Hex()), nil
}
func (r *contentRevisionResolver) Payload(ctx context.Context, obj *model.ContentRevision) ([]graphql.ArticlePayload, error) {
	data := []graphql.ArticlePayload{}
	for key, val := range obj.Payload() {
		meta := graphql.ArticlePayload{
			// Make the key lower camel case for consistency with GraphQL field names
			Key:   strcase.ToLowerCamel(key),
			Value: model.ArticlePayloadValue{Value: val},
		}
		data = append(data, meta)
	}
	return data, nil
}
func (r *contentRevisionResolver) EditorAddress(ctx context.Context, obj *model.ContentRevision) (string, error) {
	return r.Resolver.DetermineAddrCase(obj.ListingAddress().Hex()), nil
}
func (r *contentRevisionResolver) ContractContentID(ctx context.Context, obj *model.ContentRevision) (int, error) {
	bigInt := obj.ContractContentID()
	return int(bigInt.Int64()), nil
}
func (r *contentRevisionResolver) ContractRevisionID(ctx context.Context, obj *model.ContentRevision) (int, error) {
	bigInt := obj.ContractContentID()
	return int(bigInt.Int64()), nil
}
func (r *contentRevisionResolver) RevisionDate(ctx context.Context, obj *model.ContentRevision) (int, error) {
	return int(obj.RevisionDateTs()), nil
}

// QUERIES

func (r *queryResolver) Articles(ctx context.Context, addr *string, first *int,
	after *string, lowercaseAddr *bool) ([]model.ContentRevision, error) {
	r.Resolver.lowercaseAddr = lowercaseAddr
	return r.NewsroomArticles(ctx, addr, first, after, lowercaseAddr)
}

func (r *queryResolver) NewsroomArticles(ctx context.Context, addr *string, first *int,
	after *string, lowercaseAddr *bool) ([]model.ContentRevision, error) {
	r.Resolver.lowercaseAddr = lowercaseAddr
	criteria := &model.ContentRevisionCriteria{
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

	modelRevisions := make([]model.ContentRevision, len(revisions))
	for index, revision := range revisions {
		modelRevisions[index] = *revision
	}
	return modelRevisions, nil
}
