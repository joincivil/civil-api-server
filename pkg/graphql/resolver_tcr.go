package graphql

import (
	context "context"
	"github.com/iancoleman/strcase"
	"strconv"

	"github.com/ethereum/go-ethereum/common"

	crawlutils "github.com/joincivil/civil-events-crawler/pkg/utils"

	model "github.com/joincivil/civil-events-processor/pkg/model"
	putils "github.com/joincivil/civil-events-processor/pkg/utils"

	"github.com/joincivil/civil-api-server/pkg/generated/graphql"
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
func (r *appealResolver) AppealFeePaid(ctx context.Context, obj *model.Appeal) (int, error) {
	return int(obj.AppealFeePaid().Int64()), nil
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

func (r *governanceEventResolver) SenderAddress(ctx context.Context, obj *model.GovernanceEvent) (string, error) {
	return obj.SenderAddress().Hex(), nil
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
	return int(obj.ChallengeID().Int64()), nil
}
func (r *listingResolver) Challenge(ctx context.Context, obj *model.Listing) (*model.Challenge, error) {
	// TODO(IS): add dataloader here
	challengeID := int(obj.ChallengeID().Int64())
	challenge, err := r.challengePersister.ChallengeByChallengeID(challengeID)
	if err != nil {
		return nil, err
	}
	return challenge, nil
}
func (r *listingResolver) PrevChallenge(ctx context.Context, obj *model.Listing) (*model.Challenge, error) {
	// TODO(IS): add dataloader here
	challenges, err := r.challengePersister.ChallengesByListingAddress(obj.ContractAddress())

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
func (r *pollResolver) VotesFor(ctx context.Context, obj *model.Poll) (int, error) {
	return int(obj.VotesFor().Int64()), nil
}
func (r *pollResolver) VotesAgainst(ctx context.Context, obj *model.Poll) (int, error) {
	return int(obj.VotesAgainst().Int64()), nil
}

// QUERIES

func (r *queryResolver) Challenge(ctx context.Context, id int) (*model.Challenge, error) {
	return r.TcrChallenge(ctx, id)
}

func (r *queryResolver) TcrChallenge(ctx context.Context, id int) (*model.Challenge, error) {
	challenge, err := r.challengePersister.ChallengeByChallengeID(id)
	if err != nil {
		return nil, err
	}
	return challenge, nil
}

func (r *queryResolver) GovernanceEvents(ctx context.Context, addr *string, after *string,
	creationDate *graphql.DateRange, first *int) ([]model.GovernanceEvent, error) {
	return r.TcrGovernanceEvents(ctx, addr, after, creationDate, first)
}

func (r *queryResolver) TcrGovernanceEvents(ctx context.Context, addr *string, after *string,
	creationDate *graphql.DateRange, first *int) ([]model.GovernanceEvent, error) {
	criteria := &model.GovernanceEventCriteria{}

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

	events, err := r.govEventPersister.GovernanceEventsByCriteria(criteria)
	if err != nil {
		return nil, err
	}

	modelEvents := make([]model.GovernanceEvent, len(events))
	for index, event := range events {
		modelEvents[index] = *event
	}
	return modelEvents, err
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
	return r.TcrListings(ctx, first, after, whitelistedOnly, rejectedOnly,
		activeChallenge, currentApplication)
}

func (r *queryResolver) TcrListings(ctx context.Context, first *int, after *string,
	whitelistedOnly *bool, rejectedOnly *bool, activeChallenge *bool,
	currentApplication *bool) ([]model.Listing, error) {
	criteria := &model.ListingCriteria{}

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

	listings, err := r.listingPersister.ListingsByCriteria(criteria)
	if err != nil {
		return nil, err
	}

	modelListings := make([]model.Listing, len(listings))
	for index, listing := range listings {
		modelListings[index] = *listing
	}
	return modelListings, nil
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
