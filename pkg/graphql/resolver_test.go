package graphql_test

import (
	"context"
	"fmt"
	"math/big"
	"math/rand"
	"testing"

	"github.com/ethereum/go-ethereum/common"

	graphqlgen "github.com/joincivil/civil-api-server/pkg/generated/graphql"
	"github.com/joincivil/civil-api-server/pkg/graphql"

	pmodel "github.com/joincivil/civil-events-processor/pkg/model"

	cstrings "github.com/joincivil/go-common/pkg/strings"
	ctime "github.com/joincivil/go-common/pkg/time"
)

type testListingPersister struct {
	Listings []*pmodel.Listing
}

func (t *testListingPersister) ListingsByCriteria(criteria *pmodel.ListingCriteria) ([]*pmodel.Listing, error) {
	offset := 0
	if criteria.Offset != 0 {
		offset = criteria.Offset
	}

	count := 20
	if criteria.Count != 0 {
		count = criteria.Count
	}

	start := offset
	end := offset + count
	if end >= len(t.Listings) {
		end = len(t.Listings)
	}

	results := t.Listings[start:end]
	return results, nil
}

func (t *testListingPersister) ListingsByAddresses(addresses []common.Address) ([]*pmodel.Listing, error) {
	return t.Listings, nil
}

func (t *testListingPersister) ListingByAddress(address common.Address) (*pmodel.Listing, error) {
	return t.Listings[0], nil
}

func (t *testListingPersister) CreateListing(listing *pmodel.Listing) error {
	return nil
}

func (t *testListingPersister) UpdateListing(listing *pmodel.Listing, updatedFields []string) error {
	return nil
}

func (t *testListingPersister) DeleteListing(listing *pmodel.Listing) error {
	return nil
}

type testGovernanceEventPersister struct {
	Events []*pmodel.GovernanceEvent
}

//GovernanceEventsByTxHash gets governance events based on txhash
func (t *testGovernanceEventPersister) GovernanceEventsByTxHash(txHash common.Hash) ([]*pmodel.GovernanceEvent, error) {
	return t.Events, nil
}

// GovernanceEventsByCriteria retrieves governance events based on criteria
func (t *testGovernanceEventPersister) GovernanceEventsByCriteria(criteria *pmodel.GovernanceEventCriteria) ([]*pmodel.GovernanceEvent, error) {
	offset := 0
	if criteria.Offset != 0 {
		offset = criteria.Offset
	}

	count := 20
	if criteria.Count != 0 {
		count = criteria.Count
	}

	start := offset
	end := offset + count
	if end >= len(t.Events) {
		end = len(t.Events)
	}

	results := t.Events[start:end]
	return results, nil
}

// GovernanceEventsByListingAddress retrieves governance events based on listing address
func (t *testGovernanceEventPersister) GovernanceEventsByListingAddress(address common.Address) ([]*pmodel.GovernanceEvent, error) {
	return t.Events, nil
}

// GovernanceEventByChallengeID retrieves challenge by challengeID
func (t *testGovernanceEventPersister) GovernanceEventByChallengeID(challengeID int) (*pmodel.GovernanceEvent, error) {
	return &pmodel.GovernanceEvent{}, nil
}

// GovernanceEventsByChallengeIDs retrieves challenges by challengeIDs
func (t *testGovernanceEventPersister) GovernanceEventsByChallengeIDs(challengeIDs []int) ([]*pmodel.GovernanceEvent, error) {
	return t.Events, nil
}

// CreateGovernanceEvent creates a new governance event
func (t *testGovernanceEventPersister) CreateGovernanceEvent(govEvent *pmodel.GovernanceEvent) error {
	return nil
}

// UpdateGovernanceEvent updates fields on an existing governance event
func (t *testGovernanceEventPersister) UpdateGovernanceEvent(govEvent *pmodel.GovernanceEvent, updatedFields []string) error {
	return nil
}

// DeleteGovernanceEvent removes a governance event
func (t *testGovernanceEventPersister) DeleteGovernanceEvent(govEvent *pmodel.GovernanceEvent) error {
	return nil
}

func getTestListings(t *testing.T) []*pmodel.Listing {
	listings := []*pmodel.Listing{}
	for i := 0; i < 54; i++ {
		rand1, err := cstrings.RandomHexStr(32)
		if err != nil {
			t.Logf("Error getting random hex str: err: %v", err)
		}
		rand2, err := cstrings.RandomHexStr(32)
		if err != nil {
			t.Logf("Error getting random hex str: err: %v", err)
		}
		rand3, err := cstrings.RandomHexStr(32)
		if err != nil {
			t.Logf("Error getting random hex str: err: %v", err)
		}
		rand4, err := cstrings.RandomHexStr(32)
		if err != nil {
			t.Logf("Error getting random hex str: err: %v", err)
		}

		listing := pmodel.NewListing(&pmodel.NewListingParams{
			Name:                 fmt.Sprintf("listing%v", i),
			ContractAddress:      common.HexToAddress(rand1),
			Whitelisted:          (rand.Intn(10) <= 5),
			LastState:            pmodel.GovernanceStateAppWhitelisted,
			URL:                  "",
			Charter:              &pmodel.Charter{},
			Owner:                common.HexToAddress(rand2),
			OwnerAddresses:       []common.Address{common.HexToAddress(rand3)},
			ContributorAddresses: []common.Address{common.HexToAddress(rand4)},
			CreatedDateTs:        1543339458,
			ApplicationDateTs:    1543339458,
			ApprovalDateTs:       1543339458,
			LastUpdatedDateTs:    1543339458,
			AppExpiry:            big.NewInt(10),
			UnstakedDeposit:      big.NewInt(10000),
			ChallengeID:          big.NewInt(100),
		})
		listings = append(listings, listing)
	}
	return listings
}

// Generates a random list of gov events for a single address
func getTestGovEvents(t *testing.T) ([]*pmodel.GovernanceEvent, common.Address) {
	listingAddress1, err := cstrings.RandomHexStr(32)
	if err != nil {
		t.Logf("Error getting random hex str: err: %v", err)
	}

	addr := common.HexToAddress(listingAddress1)
	numEvents := 10
	events := []*pmodel.GovernanceEvent{}

	for i := 0; i < numEvents; i++ {
		rand3, err := cstrings.RandomHexStr(32)
		if err != nil {
			t.Logf("Error getting random hex str: err: %v", err)
		}
		rand4, err := cstrings.RandomHexStr(32)
		if err != nil {
			t.Logf("Error getting random hex str: err: %v", err)
		}
		rand5, err := cstrings.RandomHexStr(32)
		if err != nil {
			t.Logf("Error getting random hex str: err: %v", err)
		}

		govStates := []string{
			"Application",
			"ApplicationWhitelisted",
			"Challenge",
		}

		event := pmodel.NewGovernanceEvent(
			addr,
			pmodel.Metadata{},
			govStates[rand.Intn(len(govStates))],
			ctime.CurrentEpochSecsInInt64(),
			ctime.CurrentEpochSecsInInt64(),
			rand3,
			uint64(1000),
			common.HexToHash(rand4),
			uint(i),
			common.HexToHash(rand5),
			uint(i),
		)

		events = append(events, event)
	}
	return events, addr
}

func initResolver(t *testing.T) (*graphql.Resolver, common.Address) {
	listingPersister := &testListingPersister{
		Listings: getTestListings(t),
	}
	events, govEventAddr := getTestGovEvents(t)
	govEventsPersister := &testGovernanceEventPersister{
		Events: events,
	}
	resolver := graphql.NewResolver(&graphql.ResolverConfig{
		InvoicePersister:    nil,
		ListingPersister:    listingPersister,
		GovEventPersister:   govEventsPersister,
		RevisionPersister:   nil,
		ChallengePersister:  nil,
		AppealPersister:     nil,
		PollPersister:       nil,
		OnfidoAPI:           nil,
		OnfidoTokenReferrer: "",
		TokenFoundry:        nil,
		UserService:         nil,
	})
	return resolver, govEventAddr
}

func TestResolverTcrListings(t *testing.T) {
	resolver, _ := initResolver(t)

	queries := resolver.Query()

	first := 54
	cursor, err := queries.TcrListings(context.Background(), &first, nil, nil, nil, nil, nil)
	if err != nil {
		t.Errorf("Should have gotten a cursor: err: %v", err)
	}
	if len(cursor.Edges) != 54 {
		t.Errorf("Should have gotten 54 listings: len: %v", len(cursor.Edges))
	}
	if cursor.PageInfo.HasNextPage {
		t.Errorf("Should have had hasNextPage equal to false: is: %v", cursor.PageInfo.HasNextPage)
	}

	for _, edge := range cursor.Edges {
		if edge.Cursor == "" {
			t.Errorf("Should have included a non empty cursor value: cursor: %v", edge.Cursor)
		}
		if edge.Node.CreatedDateTs() == 0 {
			t.Errorf("Should have included a non empty created date: date: %v", edge.Node.CreatedDateTs())
		}
	}
}

func TestResolverTcrListingsPagination(t *testing.T) {
	resolver, _ := initResolver(t)

	queries := resolver.Query()
	// Get 10 at a time
	first := 10
	// No initial cursor
	var after string
	allEdges := []*graphqlgen.ListingEdge{}

Loop:
	for {
		cursor, err := queries.TcrListings(context.Background(), &first, &after, nil, nil, nil, nil)
		if err != nil {
			t.Errorf("Should have gotten a cursor: err: %v", err)
		}
		if len(cursor.Edges) > 10 {
			t.Errorf("Should have gotten 10 or less results per query: len: %v", len(cursor.Edges))
		}
		if *cursor.PageInfo.EndCursor == "" {
			t.Errorf("Should have gotten an end cursor")
		}
		allEdges = append(allEdges, cursor.Edges...)

		after = *cursor.PageInfo.EndCursor
		if !cursor.PageInfo.HasNextPage {
			break Loop
		}
	}

	if len(allEdges) != 54 {
		t.Errorf("Should have gotten 54 items in the listings: len: %v", len(allEdges))
	}
}

func TestResolverTcrGovernanceEvents(t *testing.T) {
	resolver, govEventAddr := initResolver(t)

	queries := resolver.Query()

	first := 10
	addr := govEventAddr.Hex()
	cursor, err := queries.TcrGovernanceEvents(
		context.Background(),
		&addr,
		nil,
		nil,
		&first,
	)
	if err != nil {
		t.Errorf("Should have gotten a cursor: err: %v", err)
	}
	if len(cursor.Edges) != 10 {
		t.Errorf("Should have gotten 10 listings: len: %v", len(cursor.Edges))
	}
	if cursor.PageInfo.HasNextPage {
		t.Errorf("Should have had hasNextPage equal to false: is: %v", cursor.PageInfo.HasNextPage)
	}

	for _, edge := range cursor.Edges {
		if edge.Cursor == "" {
			t.Errorf("Should have included a non empty cursor value: cursor: %v", edge.Cursor)
		}
		if edge.Node.CreationDateTs() == 0 {
			t.Errorf("Should have included a non empty created date: date: %v", edge.Node.CreationDateTs())
		}
	}
}

func TestResolverTcrGovernanceEventsPagination(t *testing.T) {
	resolver, govEventAddr := initResolver(t)

	queries := resolver.Query()
	addr := govEventAddr.Hex()
	// Get 10 at a time
	first := 2
	// No initial cursor
	var after string
	allEdges := []*graphqlgen.GovernanceEventEdge{}

Loop:
	for {
		cursor, err := queries.TcrGovernanceEvents(
			context.Background(),
			&addr,
			&after,
			nil,
			&first,
		)
		if err != nil {
			t.Errorf("Should have gotten a cursor: err: %v", err)
		}
		if len(cursor.Edges) != 2 {
			t.Errorf("Should have gotten 2 or less results per query: len: %v", len(cursor.Edges))
		}
		if *cursor.PageInfo.EndCursor == "" {
			t.Errorf("Should have gotten an end cursor")
		}
		allEdges = append(allEdges, cursor.Edges...)

		after = *cursor.PageInfo.EndCursor
		if !cursor.PageInfo.HasNextPage {
			break Loop
		}
	}

	if len(allEdges) != 10 {
		t.Errorf("Should have gotten 54 items in the listings: len: %v", len(allEdges))
	}
}
