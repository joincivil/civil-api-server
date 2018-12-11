package graphql

import (
	context "context"
	"fmt"
	"strconv"

	"github.com/iancoleman/strcase"

	"github.com/ethereum/go-ethereum/common"

	crawlutils "github.com/joincivil/civil-events-crawler/pkg/utils"

	model "github.com/joincivil/civil-events-processor/pkg/model"
	putils "github.com/joincivil/civil-events-processor/pkg/utils"

	"github.com/joincivil/civil-api-server/pkg/generated/graphql"
)

const (
	defaultCriteriaCount = 25
)

// Appeal is the resolver for the Appeal type
func (r *Resolver) Appeal() graphql.AppealResolver {
	return &appealResolver{r}
}

// Charter is the resolver for the Charter type
func (r *Resolver) Charter() graphql.CharterResolver {
	return &charterResolver{r}
}

// Challenge is the resolver for the Challenge type
func (r *Resolver) Challenge() graphql.ChallengeResolver {
	return &challengeResolver{r}
}

// GovernanceEvent is the resolver for the GovernanceEvent type
func (r *Resolver) GovernanceEvent() graphql.GovernanceEventResolver {
	return &governanceEventResolver{r}
}

// Listing is the resolver for the Listing type
func (r *Resolver) Listing() graphql.ListingResolver {
	return &listingResolver{r}
}

// Poll is the resolver for the Poll type
func (r *Resolver) Poll() graphql.PollResolver {
	return &pollResolver{r}
}

// TYPE RESOLVERS

type appealResolver struct{ *Resolver }

func (r *appealResolver) Requester(ctx context.Context, obj *model.Appeal) (string, error) {
	return obj.Requester().Hex(), nil
}
func (r *appealResolver) AppealFeePaid(ctx context.Context, obj *model.Appeal) (string, error) {
	return obj.AppealFeePaid().String(), nil
}
func (r *appealResolver) AppealPhaseExpiry(ctx context.Context, obj *model.Appeal) (int, error) {
	return int(obj.AppealPhaseExpiry().Int64()), nil
}
func (r *appealResolver) AppealOpenToChallengeExpiry(ctx context.Context, obj *model.Appeal) (int, error) {
	return int(obj.AppealOpenToChallengeExpiry().Int64()), nil
}
func (r *appealResolver) AppealChallengeID(ctx context.Context, obj *model.Appeal) (int, error) {
	return int(obj.AppealChallengeID().Int64()), nil
}
func (r *appealResolver) AppealChallenge(ctx context.Context, obj *model.Appeal) (*model.Challenge, error) {
	if obj.AppealChallengeID() == nil {
		return nil, nil
	}
	challengeID := int(obj.AppealChallengeID().Int64())
	challenge, err := r.challengePersister.ChallengeByChallengeID(challengeID)
	if err != nil {
		return nil, err
	}
	return challenge, nil
}

type challengeResolver struct{ *Resolver }

func (r *challengeResolver) ChallengeID(ctx context.Context, obj *model.Challenge) (int, error) {
	return int(obj.ChallengeID().Uint64()), nil
}
func (r *challengeResolver) ListingAddress(ctx context.Context, obj *model.Challenge) (string, error) {
	return obj.ListingAddress().Hex(), nil
}
func (r *challengeResolver) RewardPool(ctx context.Context, obj *model.Challenge) (string, error) {
	rewardPool := obj.RewardPool()
	if rewardPool != nil {
		return rewardPool.String(), nil
	}
	return "", nil
}
func (r *challengeResolver) Challenger(ctx context.Context, obj *model.Challenge) (string, error) {
	return obj.Challenger().Hex(), nil
}
func (r *challengeResolver) Stake(ctx context.Context, obj *model.Challenge) (string, error) {
	stake := obj.Stake()
	if stake != nil {
		return stake.String(), nil
	}
	return "", nil

}
func (r *challengeResolver) TotalTokens(ctx context.Context, obj *model.Challenge) (string, error) {
	totalTokens := obj.TotalTokens()
	if totalTokens != nil {
		return totalTokens.String(), nil
	}
	return "", nil
}
func (r *challengeResolver) RequestAppealExpiry(ctx context.Context, obj *model.Challenge) (int, error) {
	requestAppealExpiry := obj.RequestAppealExpiry()
	if requestAppealExpiry != nil {
		return int(requestAppealExpiry.Uint64()), nil
	}
	return 0, nil
}
func (r *challengeResolver) LastUpdatedDateTs(ctx context.Context, obj *model.Challenge) (int, error) {
	return int(obj.LastUpdatedDateTs()), nil
}
func (r *challengeResolver) Poll(ctx context.Context, obj *model.Challenge) (*model.Poll, error) {
	challengeID := int(obj.ChallengeID().Int64())

	poll, err := r.pollPersister.PollByPollID(challengeID)
	if err != nil {
		return nil, err
	}

	return poll, nil
}
func (r *challengeResolver) Appeal(ctx context.Context, obj *model.Challenge) (*model.Appeal, error) {
	challengeID := int(obj.ChallengeID().Int64())

	appeal, err := r.appealPersister.AppealByChallengeID(challengeID)
	if err != nil {
		return nil, err
	}

	return appeal, nil
}

type charterResolver struct{ *Resolver }

func (r *charterResolver) ContentID(ctx context.Context, obj *model.Charter) (int, error) {
	return int(obj.ContentID().Int64()), nil
}
func (r *charterResolver) RevisionID(ctx context.Context, obj *model.Charter) (int, error) {
	return int(obj.RevisionID().Int64()), nil
}
func (r *charterResolver) Signature(ctx context.Context, obj *model.Charter) (string, error) {
	return string(obj.Signature()), nil
}
func (r *charterResolver) Author(ctx context.Context, obj *model.Charter) (string, error) {
	return obj.Author().Hex(), nil
}
func (r *charterResolver) ContentHash(ctx context.Context, obj *model.Charter) (string, error) {
	return putils.Byte32ToHexString(obj.ContentHash()), nil
}
func (r *charterResolver) Timestamp(ctx context.Context, obj *model.Charter) (int, error) {
	return int(obj.Timestamp().Int64()), nil
}

type governanceEventResolver struct{ *Resolver }

func (r *governanceEventResolver) ListingAddress(ctx context.Context, obj *model.GovernanceEvent) (string, error) {
	return obj.ListingAddress().Hex(), nil
}

func (r *governanceEventResolver) Metadata(ctx context.Context, obj *model.GovernanceEvent) ([]graphql.Metadata, error) {
	data := make([]graphql.Metadata, len(obj.Metadata()))
	index := 0
	for key, val := range obj.Metadata() {
		meta := graphql.Metadata{}

		// Make the key lower camel case for consistency with GraphQL field names
		meta.Key = strcase.ToLowerCamel(key)

		switch v := val.(type) {
		case int:
			meta.Value = strconv.Itoa(v)
		case int64:
			meta.Value = strconv.FormatInt(v, 10)
		case float64:
			meta.Value = strconv.FormatFloat(v, 'f', -1, 64)
		case bool:
			meta.Value = strconv.FormatBool(v)
		default:
			meta.Value = v.(string)
		}

		data[index] = meta
		index++
	}
	return data, nil
}
func (r *governanceEventResolver) BlockData(ctx context.Context, obj *model.GovernanceEvent) (graphql.BlockData, error) {
	modelBlockData := obj.BlockData()
	blockData := graphql.BlockData{}
	blockData.BlockNumber = int(modelBlockData.BlockNumber())
	blockData.TxHash = modelBlockData.TxHash()
	blockData.TxIndex = int(modelBlockData.TxIndex())
	blockData.BlockHash = modelBlockData.BlockHash()
	blockData.Index = int(modelBlockData.Index())
	return blockData, nil
}
func (r *governanceEventResolver) CreationDate(ctx context.Context, obj *model.GovernanceEvent) (int, error) {
	return int(obj.CreationDateTs()), nil
}
func (r *governanceEventResolver) LastUpdatedDate(ctx context.Context, obj *model.GovernanceEvent) (int, error) {
	return int(obj.LastUpdatedDateTs()), nil
}
func (r *governanceEventResolver) Listing(ctx context.Context, obj *model.GovernanceEvent) (model.Listing, error) {
	loaders := ctxLoaders(ctx)
	listingAddress := obj.ListingAddress().Hex()
	listing, err := loaders.listingLoader.Load(listingAddress)
	if err != nil {
		return model.Listing{}, err
	}
	if listing == nil {
		return model.Listing{}, nil
	}
	return *listing, nil
}

type listingResolver struct{ *Resolver }

func (r *listingResolver) ContractAddress(ctx context.Context, obj *model.Listing) (string, error) {
	return obj.ContractAddress().Hex(), nil
}
func (r *listingResolver) LastGovState(ctx context.Context, obj *model.Listing) (string, error) {
	return obj.LastGovernanceStateString(), nil
}
func (r *listingResolver) OwnerAddresses(ctx context.Context, obj *model.Listing) ([]string, error) {
	addrs := obj.OwnerAddresses()
	ownerAddrs := make([]string, len(addrs))
	for index, addr := range addrs {
		ownerAddrs[index] = addr.Hex()
	}
	return ownerAddrs, nil
}
func (r *listingResolver) Owner(ctx context.Context, obj *model.Listing) (string, error) {
	return obj.Owner().Hex(), nil
}
func (r *listingResolver) ContributorAddresses(ctx context.Context, obj *model.Listing) ([]string, error) {
	addrs := obj.ContributorAddresses()
	ownerAddrs := make([]string, len(addrs))
	for index, addr := range addrs {
		ownerAddrs[index] = addr.Hex()
	}
	return ownerAddrs, nil
}
func (r *listingResolver) CreatedDate(ctx context.Context, obj *model.Listing) (int, error) {
	return int(obj.CreatedDateTs()), nil
}
func (r *listingResolver) ApplicationDate(ctx context.Context, obj *model.Listing) (*int, error) {
	if obj.ApplicationDateTs() == 0 {
		return nil, nil
	}
	timeObj := int(obj.ApplicationDateTs())
	return &timeObj, nil
}
func (r *listingResolver) ApprovalDate(ctx context.Context, obj *model.Listing) (*int, error) {
	if obj.ApprovalDateTs() == 0 {
		return nil, nil
	}
	timeObj := int(obj.ApprovalDateTs())
	return &timeObj, nil
}
func (r *listingResolver) LastUpdatedDate(ctx context.Context, obj *model.Listing) (int, error) {
	return int(obj.LastUpdatedDateTs()), nil
}
func (r *listingResolver) AppExpiry(ctx context.Context, obj *model.Listing) (int, error) {
	return int(obj.AppExpiry().Int64()), nil
}
func (r *listingResolver) UnstakedDeposit(ctx context.Context, obj *model.Listing) (string, error) {
	return obj.UnstakedDeposit().String(), nil
}
func (r *listingResolver) ChallengeID(ctx context.Context, obj *model.Listing) (int, error) {
	challengeID := obj.ChallengeID().Int64()
	// Currently saving listings who haven't yet been challenged with challengeID=-1. Handling
	// this to return 0 matching with on-chain data
	if challengeID < 0 {
		return 0, nil
	}
	return int(challengeID), nil
}
func (r *listingResolver) Challenge(ctx context.Context, obj *model.Listing) (*model.Challenge, error) {
	loaders := ctxLoaders(ctx)
	challengeID := int(obj.ChallengeID().Int64())
	challenge, err := loaders.challengeLoader.Load(challengeID)
	if err != nil {
		return nil, err
	}
	return challenge, nil
}
func (r *listingResolver) PrevChallenge(ctx context.Context, obj *model.Listing) (*model.Challenge, error) {
	loaders := ctxLoaders(ctx)
	listingAddress := obj.ContractAddress().Hex()
	challenges, err := loaders.challengeAddressesLoader.Load(listingAddress)

	// No 0 challenges found, return nil
	if err == model.ErrPersisterNoResults || challenges == nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// TODO(PN): Move this value to somewhere common, like in processor code
	resolvedChallengeID := int64(0)
	totalChallenges := len(challenges)

	// If only 1 challenge, check to see if it is resolved and the current challenge id is 0.
	// If it is, then is it the "previous" challenge.
	// If it is not, return nil as there is no previous challenge.
	if totalChallenges == 1 {
		challenge := challenges[0]
		if obj.ChallengeID().Int64() == resolvedChallengeID && challenge.Resolved() {
			return challenge, nil
		}
		return nil, nil
	}

	// If there is more than 1 challenge, then determine which is the "previous challenge"
	latestChallenge := challenges[totalChallenges-1]
	nextLatestChallenge := challenges[totalChallenges-2]
	if obj.ChallengeID().Int64() == resolvedChallengeID && latestChallenge.Resolved() {
		return latestChallenge, nil
	}
	return nextLatestChallenge, nil
}

type pollResolver struct{ *Resolver }

func (r *pollResolver) CommitEndDate(ctx context.Context, obj *model.Poll) (int, error) {
	return int(obj.CommitEndDate().Int64()), nil
}
func (r *pollResolver) RevealEndDate(ctx context.Context, obj *model.Poll) (int, error) {
	return int(obj.RevealEndDate().Int64()), nil
}
func (r *pollResolver) VoteQuorum(ctx context.Context, obj *model.Poll) (int, error) {
	return int(obj.VoteQuorum().Int64()), nil
}
func (r *pollResolver) VotesFor(ctx context.Context, obj *model.Poll) (string, error) {
	return obj.VotesFor().String(), nil
}
func (r *pollResolver) VotesAgainst(ctx context.Context, obj *model.Poll) (string, error) {
	return obj.VotesAgainst().String(), nil
}

// QUERIES

func (r *queryResolver) Challenge(ctx context.Context, id int) (*model.Challenge, error) {
	return r.TcrChallenge(ctx, id)
}

func (r *queryResolver) TcrChallenge(ctx context.Context, id int) (*model.Challenge, error) {
	loaders := ctxLoaders(ctx)
	challenge, err := loaders.challengeLoader.Load(id)
	if err != nil {
		return nil, err
	}
	return challenge, nil
}

func (r *queryResolver) GovernanceEvents(ctx context.Context, addr *string, after *string,
	creationDate *graphql.DateRange, first *int) ([]model.GovernanceEvent, error) {
	resultCursor, err := r.TcrGovernanceEvents(ctx, addr, after, creationDate, first)
	if err != nil {
		return []model.GovernanceEvent{}, nil
	}

	results := make([]model.GovernanceEvent, len(resultCursor.Edges))
	for index, edge := range resultCursor.Edges {
		results[index] = edge.Node
	}
	return results, nil
}

func (r *queryResolver) TcrGovernanceEvents(ctx context.Context, addr *string, after *string,
	creationDate *graphql.DateRange, first *int) (*graphql.GovernanceEventResultCursor, error) {
	var err error

	// The default pagination is by offset
	// Only support sorted offset until we need other types
	cursor := defaultPaginationCursor
	criteria := &model.GovernanceEventCriteria{}

	// Figure out the pagination index start point if given
	if after != nil && *after != "" {
		criteria.Offset, cursor, err = r.paginationOffsetFromCursor(cursor, after)
		if err != nil {
			return nil, err
		}
	}

	criteria.Count = r.criteriaCount(first)

	if addr != nil && *addr != "" {
		criteria.ListingAddress = crawlutils.NormalizeEthAddress(*addr)
	}
	if creationDate != nil {
		if creationDate.Gt != nil {
			criteria.CreatedFromTs = int64(*creationDate.Gt)
		}
		if creationDate.Lt != nil {
			criteria.CreatedBeforeTs = int64(*creationDate.Lt)
		}
	}

	allEvents, err := r.govEventPersister.GovernanceEventsByCriteria(criteria)
	if err != nil {
		return nil, err
	}

	// Figure out the listings to return and if there is another page
	// to query for.
	events, hasNextPage := r.govEventsReturnGovEvents(allEvents, criteria)

	// Build edges
	// Only support sorted offset until we need other types
	edges := r.govEventsBuildEdges(events, cursor)

	// Figure out the endCursor value
	endCursor := r.govEventsEndCursor(edges)

	return &graphql.GovernanceEventResultCursor{
		Edges: edges,
		PageInfo: graphql.PageInfo{
			EndCursor:   endCursor,
			HasNextPage: hasNextPage,
		},
	}, err
}

func (r *queryResolver) govEventsReturnGovEvents(allEvents []*model.GovernanceEvent,
	criteria *model.GovernanceEventCriteria) ([]*model.GovernanceEvent, bool) {
	allEventsLen := len(allEvents)

	// Figure out the hasNextPage value
	// If we received all the events for the criteria.Count, that means there
	// are more beyond the requested number of events.  This saves us an extra query.
	hasNextPage := false
	var events []*model.GovernanceEvent

	// Figure out the "true" events we want to return.
	// If the events actually equals what we requested, then we have more results
	// and hasNextPage should be true
	if allEventsLen == criteria.Count {
		hasNextPage = true
		events = allEvents[:allEventsLen-1]
	} else {
		events = allEvents
	}
	return events, hasNextPage
}

func (r *queryResolver) govEventsBuildEdges(events []*model.GovernanceEvent,
	cursor *paginationCursor) []*graphql.GovernanceEventEdge {

	edges := make([]*graphql.GovernanceEventEdge, len(events))

	// Build edges
	// Only support sorted offset until we need other types
	for index, event := range events {
		cv := cursor.ValueInt()
		newCursor := &paginationCursor{
			typeName: cursor.typeName,
			value:    fmt.Sprintf("%v", cv+index),
		}
		edges[index] = &graphql.GovernanceEventEdge{
			Cursor: newCursor.Encode(),
			Node:   *event,
		}
	}
	return edges
}

func (r *queryResolver) govEventsEndCursor(edges []*graphql.GovernanceEventEdge) *string {
	var endCursor *string
	if len(edges) > 0 {
		endCursor = &(edges[len(edges)-1]).Cursor
	}
	return endCursor
}

func (r *queryResolver) GovernanceEventsTxHash(ctx context.Context,
	txString string) ([]model.GovernanceEvent, error) {
	return r.TcrGovernanceEventsTxHash(ctx, txString)
}

func (r *queryResolver) TcrGovernanceEventsTxHash(ctx context.Context,
	txString string) ([]model.GovernanceEvent, error) {
	txHash := common.HexToHash(txString)
	events, err := r.govEventPersister.GovernanceEventsByTxHash(txHash)
	if err != nil {
		return nil, err
	}
	modelEvents := make([]model.GovernanceEvent, len(events))
	for index, event := range events {
		modelEvents[index] = *event
	}
	return modelEvents, err
}

func (r *queryResolver) Listings(ctx context.Context, first *int, after *string,
	whitelistedOnly *bool, rejectedOnly *bool, activeChallenge *bool,
	currentApplication *bool) ([]model.Listing, error) {

	resultCursor, err := r.TcrListings(ctx, first, after, whitelistedOnly, rejectedOnly,
		activeChallenge, currentApplication)
	if err != nil {
		return []model.Listing{}, err
	}

	results := make([]model.Listing, len(resultCursor.Edges))
	for index, edge := range resultCursor.Edges {
		results[index] = edge.Node
	}
	return results, nil
}

func (r *queryResolver) TcrListings(ctx context.Context, first *int, after *string,
	whitelistedOnly *bool, rejectedOnly *bool, activeChallenge *bool,
	currentApplication *bool) (*graphql.ListingResultCursor, error) {
	var err error

	// The default pagination is by offset
	// Only support sorted offset until we need other types
	cursor := defaultPaginationCursor
	criteria := &model.ListingCriteria{}

	// Figure out the pagination index start point if given
	if after != nil && *after != "" {
		criteria.Offset, cursor, err = r.paginationOffsetFromCursor(cursor, after)
		if err != nil {
			return nil, err
		}
	}

	criteria.Count = r.criteriaCount(first)

	if whitelistedOnly != nil {
		criteria.WhitelistedOnly = *whitelistedOnly
	} else if rejectedOnly != nil {
		criteria.RejectedOnly = *rejectedOnly
	} else if activeChallenge != nil && currentApplication != nil {
		criteria.ActiveChallenge = *activeChallenge
		criteria.CurrentApplication = *currentApplication
	} else if activeChallenge != nil {
		criteria.ActiveChallenge = *activeChallenge
	} else if currentApplication != nil {
		criteria.CurrentApplication = *currentApplication
	}

	allListings, err := r.listingPersister.ListingsByCriteria(criteria)
	if err != nil {
		return nil, err
	}

	// Figure out the listings to return and if there is another page
	// to query for.
	listings, hasNextPage := r.listingsReturnListings(allListings, criteria)

	// Build edges
	// Only support sorted offset until we need other types
	modelEdges := r.listingsBuildEdges(listings, cursor)

	// Figure out the endCursor value
	endCursor := r.listingsEndCursor(modelEdges)

	return &graphql.ListingResultCursor{
		Edges: modelEdges,
		PageInfo: graphql.PageInfo{
			EndCursor:   endCursor,
			HasNextPage: hasNextPage,
		},
	}, nil
}

func (r *queryResolver) listingsReturnListings(allListings []*model.Listing,
	criteria *model.ListingCriteria) ([]*model.Listing, bool) {
	allListingsLen := len(allListings)

	// Figure out the hasNextPage value
	// If we received all the listings for the criteria.Count, that means there
	// are more beyond the requested number of listings.  This saves us an extra query.
	hasNextPage := false
	var listings []*model.Listing

	// Figure out the "true" listings we want to return.
	// If the listings actually equals what we requested, then we have more results
	// and hasNextPage should be true
	if allListingsLen == criteria.Count {
		hasNextPage = true
		listings = allListings[:allListingsLen-1]
	} else {
		listings = allListings
	}
	return listings, hasNextPage
}

func (r *queryResolver) listingsBuildEdges(listings []*model.Listing,
	cursor *paginationCursor) []*graphql.ListingEdge {

	modelEdges := make([]*graphql.ListingEdge, len(listings))

	// Build edges
	// Only support sorted offset until we need other types
	for index, listing := range listings {
		cv := cursor.ValueInt()
		newCursor := &paginationCursor{
			typeName: cursor.typeName,
			value:    fmt.Sprintf("%v", cv+index),
		}
		modelEdges[index] = &graphql.ListingEdge{
			Cursor: newCursor.Encode(),
			Node:   *listing,
		}
	}
	return modelEdges
}

func (r *queryResolver) listingsEndCursor(modelEdges []*graphql.ListingEdge) *string {
	var endCursor *string
	if len(modelEdges) > 0 {
		endCursor = &(modelEdges[len(modelEdges)-1]).Cursor
	}
	return endCursor
}

func (r *queryResolver) Listing(ctx context.Context, addr string) (*model.Listing, error) {
	return r.TcrListing(ctx, addr)
}

func (r *queryResolver) TcrListing(ctx context.Context, addr string) (*model.Listing, error) {
	address := common.HexToAddress(addr)
	listing, err := r.listingPersister.ListingByAddress(address)
	if err != nil {
		return nil, err
	}
	return listing, nil
}

func (r *queryResolver) paginationOffsetFromCursor(cursor *paginationCursor,
	after *string) (int, *paginationCursor, error) {
	afterCursor, err := decodeToPaginationCursor(*after)
	if err != nil {
		return 0, nil, err
	}

	startOffset := 0

	if afterCursor.typeName == cursorTypeOffset {
		cursorIntValue := afterCursor.ValueInt()
		// Increment the offset and get the next item
		startOffset = cursorIntValue + 1
		afterCursor.ValueFromInt(cursorIntValue + 1)
		cursor = afterCursor
	}

	return startOffset, cursor, nil
}

func (r *queryResolver) criteriaCount(first *int) int {
	// Default count value
	criteriaCount := defaultCriteriaCount
	if first != nil {
		criteriaCount = *first
	}

	// Add 1 to all of these to see if there are additional items
	// If we see items beyond what we truly requested, then that warrants
	// another query by the caller.
	criteriaCount++
	return criteriaCount
}
