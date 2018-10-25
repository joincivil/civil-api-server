//go:generate gorunpkg github.com/99designs/gqlgen

// NOTE(PN): gqlgen does not update this file if major updates to the schema are made.
// To completely update, need to move this file and run gqlgen again and replace
// the code.  Fixed when gqlgen matures a bit more?

package graphql

import (
	context "context"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/iancoleman/strcase"

	crawlutils "github.com/joincivil/civil-events-crawler/pkg/utils"

	model "github.com/joincivil/civil-events-processor/pkg/model"
	putils "github.com/joincivil/civil-events-processor/pkg/utils"

	"github.com/joincivil/civil-api-server/pkg/generated/graphql"
	"github.com/joincivil/civil-api-server/pkg/invoicing"
	"github.com/joincivil/civil-api-server/pkg/kyc"
	"github.com/joincivil/civil-api-server/pkg/tokenfoundry"
	"github.com/joincivil/civil-api-server/pkg/users"
)

// ResolverConfig is the config params for the Resolver
type ResolverConfig struct {
	InvoicePersister    *invoicing.PostgresPersister
	ListingPersister    model.ListingPersister
	GovEventPersister   model.GovernanceEventPersister
	RevisionPersister   model.ContentRevisionPersister
	ChallengePersister  model.ChallengePersister
	AppealPersister     model.AppealPersister
	PollPersister       model.PollPersister
	UserPersister       users.UserPersister
	OnfidoAPI           *kyc.OnfidoAPI
	OnfidoTokenReferrer string
	TokenFoundry        *tokenfoundry.API
	UserService         *users.UserService
}

// NewResolver is a convenience function to init a Resolver struct
func NewResolver(config *ResolverConfig) *Resolver {
	return &Resolver{
		invoicePersister:    config.InvoicePersister,
		listingPersister:    config.ListingPersister,
		revisionPersister:   config.RevisionPersister,
		govEventPersister:   config.GovEventPersister,
		challengePersister:  config.ChallengePersister,
		appealPersister:     config.AppealPersister,
		pollPersister:       config.PollPersister,
		userPersister:       config.UserPersister,
		onfidoAPI:           config.OnfidoAPI,
		onfidoTokenReferrer: config.OnfidoTokenReferrer,
		tokenFoundry:        config.TokenFoundry,
		userService:         config.UserService,
	}
}

// Resolver is the main resolver for the GraphQL endpoint
type Resolver struct {
	invoicePersister    *invoicing.PostgresPersister
	listingPersister    model.ListingPersister
	revisionPersister   model.ContentRevisionPersister
	govEventPersister   model.GovernanceEventPersister
	challengePersister  model.ChallengePersister
	appealPersister     model.AppealPersister
	pollPersister       model.PollPersister
	userPersister       users.UserPersister
	onfidoAPI           *kyc.OnfidoAPI
	onfidoTokenReferrer string
	tokenFoundry        *tokenfoundry.API
	userService         *users.UserService
}

// ContentRevision is the resolver for the ContentRevision type
func (r *Resolver) ContentRevision() graphql.ContentRevisionResolver {
	return &contentRevisionResolver{r}
}

// GovernanceEvent is the resolver for the GovernanceEvent type
func (r *Resolver) GovernanceEvent() graphql.GovernanceEventResolver {
	return &governanceEventResolver{r}
}

// Listing is the resolver for the Listing type
func (r *Resolver) Listing() graphql.ListingResolver {
	return &listingResolver{r}
}

// Charter is the resolver for the Charter type
func (r *Resolver) Charter() graphql.CharterResolver {
	return &charterResolver{r}
}

// Challenge is the resolver for the Challenge type
func (r *Resolver) Challenge() graphql.ChallengeResolver {
	return &challengeResolver{r}
}

// Appeal is the resolver for the Appeal type
func (r *Resolver) Appeal() graphql.AppealResolver {
	return &appealResolver{r}
}

// Poll is the resolver for the Poll type
func (r *Resolver) Poll() graphql.PollResolver {
	return &pollResolver{r}
}

// Query is the resolver for the Query type
func (r *Resolver) Query() graphql.QueryResolver {
	return &queryResolver{r}
}

// Mutation is the resolver for the Mutation type
func (r *Resolver) Mutation() graphql.MutationResolver {
	return &mutationResolver{r}
}

type contentRevisionResolver struct{ *Resolver }

func (r *contentRevisionResolver) ListingAddress(ctx context.Context, obj *model.ContentRevision) (string, error) {
	return obj.ListingAddress().Hex(), nil
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
	return obj.ListingAddress().Hex(), nil
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

type queryResolver struct{ *Resolver }

func (r *queryResolver) Listings(ctx context.Context, first *int, after *string,
	whitelistedOnly *bool, rejectedOnly *bool, activeChallengeOnly *bool,
	currentApplicationOnly *bool) ([]model.Listing, error) {
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
	}
	if rejectedOnly != nil {
		criteria.RejectedOnly = *rejectedOnly
	}
	if activeChallengeOnly != nil {
		criteria.ActiveChallenge = *activeChallengeOnly
	}
	if currentApplicationOnly != nil {
		criteria.CurrentApplication = *currentApplicationOnly
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
	address := common.HexToAddress(addr)
	listing, err := r.listingPersister.ListingByAddress(address)
	if err != nil {
		return nil, err
	}
	return listing, nil
}
func (r *queryResolver) GovernanceEvents(ctx context.Context, addr *string, creationDate *graphql.DateRange,
	first *int, after *string) ([]model.GovernanceEvent, error) {
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
func (r *queryResolver) GovernanceEventsTxHash(ctx context.Context, txString string) ([]model.GovernanceEvent, error) {
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
func (r *queryResolver) Articles(ctx context.Context, addr *string, first *int,
	after *string) ([]model.ContentRevision, error) {
	criteria := &model.ContentRevisionCriteria{
		LatestOnly: true,
	}
	if addr != nil && *addr != "" {
		criteria.ListingAddress = crawlutils.NormalizeEthAddress(*addr)
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
func (r *queryResolver) Challenge(ctx context.Context, id int) (*model.Challenge, error) {
	challenge, err := r.challengePersister.ChallengeByChallengeID(id)
	if err != nil {
		return nil, err
	}
	return challenge, nil
}

type mutationResolver struct{ *Resolver }
