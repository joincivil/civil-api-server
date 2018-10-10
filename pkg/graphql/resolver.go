//go:generate gorunpkg github.com/99designs/gqlgen

// NOTE(PN): gqlgen does not update this file if major updates to the schema are made.
// To completely update, need to move this file and run gqlgen again and replace
// the code.  Fixed when gqlgen matures a bit more?

package graphql

import (
	context "context"
	"fmt"

	"github.com/joincivil/civil-api-server/pkg/tokenfoundry"

	"github.com/joincivil/civil-api-server/pkg/invoicing"
	// "fmt"
	"strconv"
	time "time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/iancoleman/strcase"

	model "github.com/joincivil/civil-events-processor/pkg/model"
	"github.com/joincivil/civil-events-processor/pkg/utils"

	"github.com/joincivil/civil-api-server/pkg/auth"
	graphql "github.com/joincivil/civil-api-server/pkg/generated/graphql"
	kyc "github.com/joincivil/civil-api-server/pkg/kyc"
)

// ResolverConfig is the config params for the Resolver
type ResolverConfig struct {
	InvoicePersister    *invoicing.PostgresPersister
	ListingPersister    model.ListingPersister
	GovEventPersister   model.GovernanceEventPersister
	RevisionPersister   model.ContentRevisionPersister
	OnfidoAPI           *kyc.OnfidoAPI
	OnfidoTokenReferrer string
	TokenFoundry        *tokenfoundry.API
}

// NewResolver is a convenience function to init a Resolver struct
func NewResolver(config *ResolverConfig) *Resolver {
	return &Resolver{
		listingPersister:    config.ListingPersister,
		revisionPersister:   config.RevisionPersister,
		govEventPersister:   config.GovEventPersister,
		invoicePersister:    config.InvoicePersister,
		onfidoAPI:           config.OnfidoAPI,
		onfidoTokenReferrer: config.OnfidoTokenReferrer,
		tokenFoundry:        config.TokenFoundry,
	}
}

// Resolver is the main resolver for the GraphQL endpoint
type Resolver struct {
	invoicePersister    *invoicing.PostgresPersister
	listingPersister    model.ListingPersister
	revisionPersister   model.ContentRevisionPersister
	govEventPersister   model.GovernanceEventPersister
	onfidoAPI           *kyc.OnfidoAPI
	onfidoTokenReferrer string
	tokenFoundry        *tokenfoundry.API
}

// ContentRevision is the resolver for the ContentRevision type
func (r *Resolver) ContentRevision() graphql.ContentRevisionResolver {
	return &contentRevisionResolver{r}
}

// GovernanceEvent is the resolver for the GovernanceEvent type
func (r *Resolver) GovernanceEvent() graphql.GovernanceEventResolver {
	return &governanceEventResolver{r}
}

// Listing is the resolver for the listingtype
func (r *Resolver) Listing() graphql.ListingResolver {
	return &listingResolver{r}
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
func (r *contentRevisionResolver) RevisionDate(ctx context.Context, obj *model.ContentRevision) (time.Time, error) {
	return utils.SecsFromEpochToTime(obj.RevisionDateTs()), nil
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
func (r *governanceEventResolver) CreationDate(ctx context.Context, obj *model.GovernanceEvent) (time.Time, error) {
	return utils.SecsFromEpochToTime(obj.CreationDateTs()), nil
}
func (r *governanceEventResolver) LastUpdatedDate(ctx context.Context, obj *model.GovernanceEvent) (time.Time, error) {
	return utils.SecsFromEpochToTime(obj.LastUpdatedDateTs()), nil
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
func (r *listingResolver) CreatedDate(ctx context.Context, obj *model.Listing) (time.Time, error) {
	return utils.SecsFromEpochToTime(obj.CreatedDateTs()), nil
}
func (r *listingResolver) ApplicationDate(ctx context.Context, obj *model.Listing) (*time.Time, error) {
	if obj.ApplicationDateTs() == 0 {
		return nil, nil
	}
	timeObj := utils.SecsFromEpochToTime(obj.ApplicationDateTs())
	return &timeObj, nil
}
func (r *listingResolver) ApprovalDate(ctx context.Context, obj *model.Listing) (*time.Time, error) {
	if obj.ApprovalDateTs() == 0 {
		return nil, nil
	}
	timeObj := utils.SecsFromEpochToTime(obj.ApprovalDateTs())
	return &timeObj, nil
}
func (r *listingResolver) LastUpdatedDate(ctx context.Context, obj *model.Listing) (time.Time, error) {
	return utils.SecsFromEpochToTime(obj.LastUpdatedDateTs()), nil
}
func (r *listingResolver) AppExpiry(ctx context.Context, obj *model.Listing) (time.Time, error) {
	return utils.SecsFromEpochToTime(obj.AppExpiry().Int64()), nil
}
func (r *listingResolver) UnstakedDeposit(ctx context.Context, obj *model.Listing) (string, error) {
	return obj.UnstakedDeposit().String(), nil
}
func (r *listingResolver) ChallengeID(ctx context.Context, obj *model.Listing) (int, error) {
	return int(obj.ChallengeID().Int64()), nil
}
func (r *listingResolver) Challenge(ctx context.Context, obj *model.Listing) (*model.GovernanceEvent, error) {
	loaders := ctxLoaders(ctx)
	challengeID := int(obj.ChallengeID().Int64())
	challenge, err := loaders.challengeLoader.Load(challengeID)
	if err != nil {
		return nil, err
	}
	return challenge, nil
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) Listings(ctx context.Context, whitelistedOnly *bool, first *int,
	after *string) ([]model.Listing, error) {
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
		criteria.ListingAddress = *addr
	}
	if creationDate != nil {
		if creationDate.Gt != nil {
			criteria.CreatedFromTs = utils.TimeToSecsFromEpoch(creationDate.Gt)
		}
		if creationDate.Lt != nil {
			criteria.CreatedBeforeTs = utils.TimeToSecsFromEpoch(creationDate.Lt)
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
		criteria.ListingAddress = *addr
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

func (r *queryResolver) CurrentUser(ctx context.Context) (*auth.CurrentUser, error) {
	token := auth.ForContext(ctx)
	if token == nil {
		return nil, fmt.Errorf("Access denied")
	}
	return auth.GetCurrentUser(token.Sub, r.invoicePersister, r.tokenFoundry)
}

type mutationResolver struct{ *Resolver }

func (r *mutationResolver) KycCreateApplicant(ctx context.Context, applicant graphql.KycCreateApplicantInput) (*string, error) {
	newAddress := kyc.Address{}
	if applicant.AptNumber != nil {
		newAddress.FlatNumber = *applicant.AptNumber
	}
	if applicant.BuildingNumber != nil {
		newAddress.BuildingNumber = *applicant.BuildingNumber
	}
	if applicant.Street != nil {
		newAddress.Street = *applicant.Street
	}
	if applicant.City != nil {
		newAddress.Town = *applicant.City
	}
	if applicant.State != nil {
		newAddress.State = *applicant.State
	}
	if applicant.Zipcode != nil {
		newAddress.Postcode = *applicant.Zipcode
	}
	if applicant.CountryOfResidence != nil {
		newAddress.Country = *applicant.CountryOfResidence
	}

	newApplicant := &kyc.Applicant{}
	newApplicant.Addresses = []kyc.Address{newAddress}
	newApplicant.FirstName = applicant.FirstName
	newApplicant.LastName = applicant.LastName

	if applicant.MiddleName != nil {
		newApplicant.MiddleName = *applicant.MiddleName
	}
	if applicant.Email != nil {
		newApplicant.Email = *applicant.Email
	}
	if applicant.DateOfBirth != nil {
		newApplicant.Dob = *applicant.DateOfBirth
	}
	if applicant.CountryOfResidence != nil {
		newApplicant.Country = *applicant.CountryOfResidence
	}

	returnedApplicant, err := r.onfidoAPI.CreateApplicant(newApplicant)
	if err != nil {
		return nil, err
	}

	return &returnedApplicant.ID, nil
}
func (r *mutationResolver) KycGenerateSdkToken(ctx context.Context, applicantID string) (*string, error) {
	token, err := r.onfidoAPI.GenerateSDKToken(applicantID, r.onfidoTokenReferrer)
	if err != nil {
		return nil, err
	}

	return &token, err
}
func (r *mutationResolver) KycCreateCheck(ctx context.Context, applicantID string, facialVariant *string) (*string, error) {
	var rep *kyc.Report
	if facialVariant != nil && *facialVariant == kyc.ReportVariantFacialSimilarityVideo {
		rep = kyc.FacialSimilarityVideoReport
	} else {
		rep = kyc.FacialSimilarityStandardReport
	}
	newCheck := &kyc.Check{
		Type: kyc.CheckTypeExpress,
		Reports: []kyc.Report{
			// *kyc.IdentityKycReport,
			*kyc.DocumentReport,
			*rep,
			// *kyc.WatchlistKycReport,
		},
	}

	returnedCheck, err := r.onfidoAPI.CreateCheck(applicantID, newCheck)
	if err != nil {
		return nil, err
	}

	return &returnedCheck.ID, nil
}
