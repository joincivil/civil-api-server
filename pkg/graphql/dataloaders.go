package graphql

// NOTE(IS): The constructor method for dataloaders are manually added here. Only listing loader for now

import (
	"context"
	model "github.com/joincivil/civil-events-processor/pkg/model"
	"github.com/joincivil/civil-events-processor/pkg/persistence/postgres"
	"net/http"
	"time"
)

type ctxKeyType struct{ name string }

var ctxKey = ctxKeyType{"userCtx"}

type loaders struct {
	listingLoader   *ListingLoader
	challengeLoader *GovernanceEventLoader
}

// DataloaderMiddleware defines the listingLoader
func DataloaderMiddleware(g *Resolver, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ldrs := loaders{}

		ldrs.listingLoader = &ListingLoader{
			maxBatch: 100,
			wait:     100 * time.Millisecond,
			fetch: func(keys []string) ([]*model.Listing, []error) {
				addresses := postgres.ListStringToListCommonAddress(keys)
				listings, err := g.listingPersister.ListingsByAddresses(addresses)
				errors := []error{err}
				return listings, errors
			},
		}

		ldrs.challengeLoader = &GovernanceEventLoader{
			maxBatch: 100,
			wait:     100 * time.Millisecond,
			fetch: func(keys []int) ([]*model.GovernanceEvent, []error) {
				challengeEvents, err := g.govEventPersister.ChallengesByIDs(keys)
				errors := []error{err}
				return challengeEvents, errors
			},
		}

		ctx := context.WithValue(r.Context(), ctxKey, ldrs) // nolint: golint
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

// func getListingLoader(ctx context.Context) *ListingLoader {
// 	return ctx.Value(listingLoaderKey).(*ListingLoader)
// }
func ctxLoaders(ctx context.Context) loaders {
	return ctx.Value(ctxKey).(loaders)
}
