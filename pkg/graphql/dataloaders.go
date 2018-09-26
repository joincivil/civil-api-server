package graphql

// NOTE(IS): The constructor method for dataloaders are manually added here. Only listing loader for now

import (
	"context"
	model "github.com/joincivil/civil-events-processor/pkg/model"
	"github.com/joincivil/civil-events-processor/pkg/persistence/postgres"
	"net/http"
	"time"
)

const listingLoaderKey = "listingloader"

// DataloaderMiddleware defines the listingLoader
func DataloaderMiddleware(g *Resolver, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		listingLoader := ListingLoader{
			maxBatch: 100,
			wait:     2 * time.Millisecond,
			fetch: func(keys []string) ([]*model.Listing, []error) {
				addresses := postgres.ListStringToListCommonAddress(keys)
				listings, err := g.listingPersister.ListingsByAddresses(addresses)
				errors := []error{err}
				return listings, errors
			},
		}
		ctx := context.WithValue(r.Context(), listingLoaderKey, &listingLoader) // nolint: golint
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func getListingLoader(ctx context.Context) *ListingLoader {
	return ctx.Value(listingLoaderKey).(*ListingLoader)
}
