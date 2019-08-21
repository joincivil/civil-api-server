package graphql

// NOTE(IS): The constructor method for dataloaders are manually added here.

import (
	"context"
	"net/http"
	"time"

	"github.com/joincivil/civil-api-server/pkg/discourse"

	"github.com/ethereum/go-ethereum/common"

	model "github.com/joincivil/civil-events-processor/pkg/model"

	cstrings "github.com/joincivil/go-common/pkg/strings"
)

type ctxKeyType struct{ name string }

var ctxKey = ctxKeyType{"userCtx"}

type loaders struct {
	listingLoader             *ListingLoader
	challengeLoader           *ChallengeLoader
	challengeAddressesLoader  *ChallengeSliceByAddressesLoader
	appealLoader              *AppealLoader
	discourseListingMapLoader *ListingMapLoader
}

// DataloaderMiddleware defines the listingLoader
func DataloaderMiddleware(g *Resolver, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ldrs := loaders{}

		ldrs.listingLoader = &ListingLoader{
			maxBatch: 100,
			wait:     100 * time.Millisecond,
			fetch: func(keys []string) ([]model.Listing, []error) {
				addresses := cstrings.ListStringToListCommonAddress(keys)
				listings, err := g.listingPersister.ListingsByAddresses(addresses)
				errors := []error{err}
				retlistings := make([]model.Listing, len(listings))
				for ind, listing := range listings {
					retlistings[ind] = *listing
				}
				return retlistings, errors
			},
		}

		ldrs.challengeLoader = &ChallengeLoader{
			maxBatch: 100,
			wait:     100 * time.Millisecond,
			fetch: func(keys []int) ([]model.Challenge, []error) {
				challengeEvents, err := g.challengePersister.ChallengesByChallengeIDs(keys)
				errors := []error{err}
				challenges := make([]model.Challenge, len(challengeEvents))
				for ind, challenge := range challengeEvents {
					challenges[ind] = *challenge
				}
				return challenges, errors
			},
		}

		ldrs.challengeAddressesLoader = &ChallengeSliceByAddressesLoader{
			maxBatch: 100,
			wait:     100 * time.Millisecond,
			fetch: func(keys []string) ([][]*model.Challenge, []error) {
				// This needs a function that returns a slice of slices of challenges for each item
				// in the slice of keys.
				addrKeys := make([]common.Address, len(keys))
				for index, key := range keys {
					addrKeys[index] = common.HexToAddress(key)
				}
				challengeEvents, err := g.challengePersister.ChallengesByListingAddresses(addrKeys)
				errors := []error{err}
				return challengeEvents, errors
			},
		}

		ldrs.appealLoader = &AppealLoader{
			maxBatch: 100,
			wait:     100 * time.Millisecond,
			fetch: func(keys []int) ([]model.Appeal, []error) {
				appeals, err := g.appealPersister.AppealsByChallengeIDs(keys)
				errors := []error{err}
				retAppeals := make([]model.Appeal, len(appeals))
				for ind, appeal := range appeals {
					retAppeals[ind] = *appeal
				}
				return retAppeals, errors
			},
		}

		ldrs.discourseListingMapLoader = &ListingMapLoader{
			maxBatch: 100,
			wait:     100 * time.Millisecond,
			fetch: func(keys []string) ([]discourse.ListingMap, []error) {
				ldms, err := g.discourseListingMapPersister.RetrieveListingMaps(keys)
				errors := []error{err}
				retldms := make([]discourse.ListingMap, len(ldms))
				for ind, ldm := range ldms {
					retldms[ind] = *ldm
				}
				return retldms, errors
			},
		}

		ctx := context.WithValue(r.Context(), ctxKey, ldrs) // nolint: golint
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func ctxLoaders(ctx context.Context) loaders {
	return ctx.Value(ctxKey).(loaders)
}
