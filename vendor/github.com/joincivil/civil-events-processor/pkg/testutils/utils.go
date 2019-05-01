package testutils

import (
	"encoding/json"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	cpersist "github.com/joincivil/go-common/pkg/persistence"

	"github.com/joincivil/civil-events-processor/pkg/model"
)

const (
	// https://civil-develop.go-vip.co/crawler-pod/wp-json/civil-newsroom-protocol/v1/revisions/11
	testCivilMetadata = `{"title":"This is a test post","revisionContentHash":"0x9e4acfe532c8458abfc1f1d30c4eaf986fee52cf1f65c9548f1dc437fb6dfd38","revisionContentUrl":"https:\/\/civil-develop.go-vip.co\/crawler-pod\/wp-json\/civil-newsroom-protocol\/v1\/revisions-content\/0x9e4acfe532c8458abfc1f1d30c4eaf986fee52cf1f65c9548f1dc437fb6dfd38\/","canonicalUrl":"https:\/\/civil-develop.go-vip.co\/crawler-pod\/2018\/07\/25\/this-is-a-test-post\/","slug":"this-is-a-test-post","description":"I'm being described","authors":[{"byline":"Walker Flynn"}],"images":[{"url":"https:\/\/civil-develop.go-vip.co\/crawler-pod\/wp-content\/uploads\/sites\/20\/2018\/07\/Messages-Image3453599984.png","hash":"0x72ca80ed96a2b1ca20bf758a2142a678c0bc316e597161d0572af378e52b2e80","h":960,"w":697}],"tags":["news"],"primaryTag":"news","revisionDate":"2018-07-25 17:17:20","originalPublishDate":"2018-07-25 17:17:07","credibilityIndicators":{"original_reporting":"1","on_the_ground":false,"sources_cited":"1","subject_specialist":false},"opinion":false,"civilSchemaVersion":"1.0.0"}`
)

//TestPersister is a test persister
type TestPersister struct {
	Listings             map[string]*model.Listing
	Revisions            map[string][]*model.ContentRevision
	GovEvents            map[string][]*model.GovernanceEvent
	Challenges           map[int]*model.Challenge
	Appeals              map[int]*model.Appeal
	Polls                map[int]*model.Poll
	TokenTransfers       map[string][]*model.TokenTransfer
	TokenTransfersTxHash map[string][]*model.TokenTransfer
	ParameterProposal    map[[32]byte]*model.ParameterProposal
	UserChallengeData    map[int]map[string]*model.UserChallengeData
	Timestamp            int64
	EventHashes          []string
}

func indexAddressInSlice(slice []common.Address, target common.Address) int {
	// if address in slice return idx, else return -1
	for idx, addr := range slice {
		if target.Hex() == addr.Hex() {
			return idx
		}
	}
	return -1
}

// Close does nothing here
func (t *TestPersister) Close() error {
	return nil
}

// ListingsByCriteria returns a slice of Listings based on ListingCriteria
func (t *TestPersister) ListingsByCriteria(criteria *model.ListingCriteria) ([]*model.Listing, error) {
	listings := make([]*model.Listing, len(t.Listings))
	index := 0
	for _, listing := range t.Listings {
		listings[index] = listing
		index++
	}
	return listings, nil
}

// ListingsByAddresses returns a slice of Listings based on addresses
func (t *TestPersister) ListingsByAddresses(addresses []common.Address) ([]*model.Listing, error) {
	results := []*model.Listing{}
	for _, address := range addresses {
		listing, err := t.ListingByAddress(address)
		if err == nil {
			results = append(results, listing)
		}
	}
	return results, nil
}

// ListingByAddress retrieves a listing based on address
func (t *TestPersister) ListingByAddress(address common.Address) (*model.Listing, error) {
	listing := t.Listings[address.Hex()]
	if listing == nil {
		return nil, cpersist.ErrPersisterNoResults
	}
	return listing, nil
}

// CreateListing creates a new listing
func (t *TestPersister) CreateListing(listing *model.Listing) error {
	addressHex := listing.ContractAddress().Hex()
	if t.Listings == nil {
		t.Listings = map[string]*model.Listing{}
	}
	t.Listings[addressHex] = listing
	return nil
}

// UpdateListing updates fields on an existing listing
func (t *TestPersister) UpdateListing(listing *model.Listing, updatedFields []string) error {
	addressHex := listing.ContractAddress().Hex()
	if t.Listings == nil {
		t.Listings = map[string]*model.Listing{}
	}
	t.Listings[addressHex] = listing
	return nil
}

// DeleteListing removes a listing
func (t *TestPersister) DeleteListing(listing *model.Listing) error {
	addressHex := listing.ContractAddress().Hex()
	if t.Listings == nil {
		t.Listings = map[string]*model.Listing{}
	}
	delete(t.Listings, addressHex)
	return nil
}

// ContentRevisionsByCriteria retrieves content revisions by ContentRevisionCriteria
func (t *TestPersister) ContentRevisionsByCriteria(criteria *model.ContentRevisionCriteria) (
	[]*model.ContentRevision, error) {
	revisions := make([]*model.ContentRevision, len(t.Revisions))
	index := 0
	for _, contentRevisions := range t.Revisions {
		revisions[index] = contentRevisions[len(contentRevisions)-1]
		index++
	}
	return revisions, nil
}

// ContentRevisions retrieves content revisions
func (t *TestPersister) ContentRevisions(address common.Address,
	contentID *big.Int) ([]*model.ContentRevision, error) {
	addressHex := address.Hex()
	addrRevs, ok := t.Revisions[addressHex]
	if !ok {
		return []*model.ContentRevision{}, nil
	}
	contentRevisions := []*model.ContentRevision{}
	for _, rev := range addrRevs {
		if rev.ContractContentID() == contentID {
			contentRevisions = append(contentRevisions, rev)
		}
	}

	return contentRevisions, nil
}

// ContentRevision retrieves content revisions
func (t *TestPersister) ContentRevision(address common.Address, contentID *big.Int,
	revisionID *big.Int) (*model.ContentRevision, error) {
	contentRevisions, err := t.ContentRevisions(address, contentID)
	if err != nil {
		return nil, nil
	}
	for _, rev := range contentRevisions {
		if rev.ContractRevisionID() == revisionID {
			return rev, nil
		}
	}
	return nil, nil
}

// CreateContentRevision creates a new content item
func (t *TestPersister) CreateContentRevision(revision *model.ContentRevision) error {
	addressHex := revision.ListingAddress().Hex()
	addrRevs, ok := t.Revisions[addressHex]
	if !ok {
		t.Revisions = map[string][]*model.ContentRevision{}
		t.Revisions[addressHex] = []*model.ContentRevision{revision}
		return nil
	}
	addrRevs = append(addrRevs, revision)
	t.Revisions[addressHex] = addrRevs
	return nil
}

// UpdateContentRevision updates fields on an existing content item
func (t *TestPersister) UpdateContentRevision(revision *model.ContentRevision, updatedFields []string) error {
	addressHex := revision.ListingAddress().Hex()
	addrRevs, ok := t.Revisions[addressHex]
	if !ok {
		t.Revisions = map[string][]*model.ContentRevision{}
		t.Revisions[addressHex] = []*model.ContentRevision{revision}
		return nil
	}
	for index, rev := range addrRevs {
		if rev.ContractContentID() == revision.ContractContentID() &&
			rev.ContractRevisionID() == revision.ContractRevisionID() {
			addrRevs[index] = revision
		}
	}
	return nil
}

// DeleteContentRevision removes a content item
func (t *TestPersister) DeleteContentRevision(revision *model.ContentRevision) error {
	contentRevisions, err := t.ContentRevisions(
		revision.ListingAddress(),
		revision.ContractContentID(),
	)
	if err != nil {
		return nil
	}
	revisionID := revision.ContractRevisionID()
	updateRevs := []*model.ContentRevision{}
	for _, rev := range contentRevisions {
		if rev.ContractRevisionID() != revisionID {
			updateRevs = append(updateRevs, rev)
		}
	}
	t.Revisions[revision.ListingAddress().Hex()] = updateRevs
	return nil
}

// GovernanceEventsByCriteria retrieves content revisions by GovernanceEventCriteria
func (t *TestPersister) GovernanceEventsByCriteria(criteria *model.GovernanceEventCriteria) (
	[]*model.GovernanceEvent, error) {
	// This is more of a placeholder
	events := make([]*model.GovernanceEvent, len(t.GovEvents))
	index := 0
	for _, event := range t.GovEvents {
		events[index] = event[len(event)-1]
		index++
	}
	return events, nil
}

// GovernanceEventByChallengeID retrieves challenge by challengeID
func (t *TestPersister) GovernanceEventByChallengeID(challengeID int) (*model.GovernanceEvent, error) {
	// NOTE(IS): Placeholder for now
	govEvent := &model.GovernanceEvent{}
	return govEvent, nil
}

// GovernanceEventsByChallengeIDs retrieves challenges by challengeIDs
func (t *TestPersister) GovernanceEventsByChallengeIDs(challengeIDs []int) ([]*model.GovernanceEvent, error) {
	// NOTE(IS): Placeholder for now
	govEvents := []*model.GovernanceEvent{}
	return govEvents, nil
}

// GovernanceEventsByTxHash gets governance events based on txhash
func (t *TestPersister) GovernanceEventsByTxHash(txHash common.Hash) ([]*model.GovernanceEvent, error) {
	// NOTE(IS): Placeholder for now
	govEvents := []*model.GovernanceEvent{}
	return govEvents, nil
}

// GovernanceEventsByListingAddress retrieves governance events based on criteria
func (t *TestPersister) GovernanceEventsByListingAddress(address common.Address) ([]*model.GovernanceEvent, error) {
	addressHex := address.Hex()
	govEvents := t.GovEvents[addressHex]
	return govEvents, nil
}

// CreateGovernanceEvent creates a new governance event
func (t *TestPersister) CreateGovernanceEvent(govEvent *model.GovernanceEvent) error {
	addressHex := govEvent.ListingAddress().Hex()
	events, ok := t.GovEvents[addressHex]
	if !ok {
		t.GovEvents = map[string][]*model.GovernanceEvent{}
		t.GovEvents[addressHex] = []*model.GovernanceEvent{govEvent}
		return nil
	}
	events = append(events, govEvent)
	t.GovEvents[addressHex] = events
	return nil
}

// UpdateGovernanceEvent updates fields on an existing governance event
func (t *TestPersister) UpdateGovernanceEvent(govEvent *model.GovernanceEvent, updatedFields []string) error {
	addressHex := govEvent.ListingAddress().Hex()
	events, ok := t.GovEvents[addressHex]
	if !ok {
		t.GovEvents[addressHex] = []*model.GovernanceEvent{govEvent}
		return nil
	}
	for index, event := range events {
		if event.GovernanceEventType() == govEvent.GovernanceEventType() &&
			event.CreationDateTs() == govEvent.CreationDateTs() {
			events[index] = govEvent
		}
	}
	return nil
}

// DeleteGovernanceEvent removes a governance event
func (t *TestPersister) DeleteGovernanceEvent(govEvent *model.GovernanceEvent) error {
	addressHex := govEvent.ListingAddress().Hex()
	events, ok := t.GovEvents[addressHex]
	if !ok {
		t.GovEvents[addressHex] = []*model.GovernanceEvent{govEvent}
		return nil
	}
	updatedEvents := []*model.GovernanceEvent{}
	for _, event := range events {
		if event.GovernanceEventType() != govEvent.GovernanceEventType() ||
			event.CreationDateTs() != govEvent.CreationDateTs() {
			updatedEvents = append(updatedEvents, event)
		}
	}
	t.GovEvents[addressHex] = updatedEvents
	return nil
}

// ChallengeByChallengeID gets a challenge by challengeID
func (t *TestPersister) ChallengeByChallengeID(challengeID int) (*model.Challenge, error) {
	challenge := t.Challenges[challengeID]
	if challenge == nil {
		return nil, cpersist.ErrPersisterNoResults
	}
	return challenge, nil
}

// ChallengesByChallengeIDs returns a slice of challenges based on challenge IDs
func (t *TestPersister) ChallengesByChallengeIDs(challengeIDs []int) ([]*model.Challenge, error) {
	results := []*model.Challenge{}
	for _, challengeID := range challengeIDs {
		challenge, err := t.ChallengeByChallengeID(challengeID)
		if err == nil {
			results = append(results, challenge)
		}
	}
	return results, nil
}

// ChallengesByListingAddress gets a list of challenges by listing
func (t *TestPersister) ChallengesByListingAddress(addr common.Address) ([]*model.Challenge, error) {
	challenges := []*model.Challenge{}
	for _, chal := range t.Challenges {
		listingAddress := chal.ListingAddress()
		if listingAddress.Hex() == addr.Hex() {
			challenges = append(challenges, chal)
		}
	}
	return challenges, nil
}

// ChallengesByListingAddresses gets a list of challenges by listing addresses
func (t *TestPersister) ChallengesByListingAddresses(addr []common.Address) ([][]*model.Challenge, error) {
	challenges := make([][]*model.Challenge, len(addr))
	for _, chal := range t.Challenges {
		listingAddress := chal.ListingAddress()
		addrIdx := indexAddressInSlice(addr, listingAddress)
		if addrIdx != -1 {
			challenges[addrIdx] = append(challenges[addrIdx], chal)
		}
	}
	return challenges, nil
}

// CreateChallenge creates a new challenge
func (t *TestPersister) CreateChallenge(challenge *model.Challenge) error {
	challengeID := int(challenge.ChallengeID().Int64())
	if t.Challenges == nil {
		t.Challenges = map[int]*model.Challenge{}
	}
	t.Challenges[challengeID] = challenge
	return nil
}

// UpdateChallenge updates a challenge
func (t *TestPersister) UpdateChallenge(challenge *model.Challenge, updatedFields []string) error {
	challengeID := int(challenge.ChallengeID().Int64())
	if t.Challenges == nil {
		t.Challenges = map[int]*model.Challenge{}
	}
	t.Challenges[challengeID] = challenge
	return nil
}

// PollByPollID gets a poll by pollID
func (t *TestPersister) PollByPollID(pollID int) (*model.Poll, error) {
	poll := t.Polls[pollID]
	if poll == nil {
		return nil, cpersist.ErrPersisterNoResults
	}
	return poll, nil
}

// PollsByPollIDs returns a slice of polls based on poll IDs
func (t *TestPersister) PollsByPollIDs(pollIDs []int) ([]*model.Poll, error) {
	results := []*model.Poll{}
	for _, pollID := range pollIDs {
		poll, err := t.PollByPollID(pollID)
		if err == nil {
			results = append(results, poll)
		}
	}
	return results, nil
}

// CreatePoll creates a new poll
func (t *TestPersister) CreatePoll(poll *model.Poll) error {
	pollID := int(poll.PollID().Int64())
	if t.Polls == nil {
		t.Polls = map[int]*model.Poll{}
	}
	t.Polls[pollID] = poll
	return nil
}

// UpdatePoll updates a poll
func (t *TestPersister) UpdatePoll(poll *model.Poll, updatedFields []string) error {
	pollID := int(poll.PollID().Int64())
	if t.Polls == nil {
		t.Polls = map[int]*model.Poll{}
	}
	t.Polls[pollID] = poll
	return nil
}

// AppealByChallengeID gets an appeal by challengeID
func (t *TestPersister) AppealByChallengeID(challengeID int) (*model.Appeal, error) {
	appeal := t.Appeals[challengeID]
	if appeal == nil {
		return nil, cpersist.ErrPersisterNoResults
	}
	return appeal, nil
}

// AppealByAppealChallengeID gets an appeal by appealChallengeID
func (t *TestPersister) AppealByAppealChallengeID(challengeID int) (*model.Appeal, error) {
	for _, appeal := range t.Appeals {
		if int(appeal.OriginalChallengeID().Int64()) == challengeID {
			return appeal, nil
		}
	}
	return nil, cpersist.ErrPersisterNoResults
}

// AppealsByChallengeIDs returns a slice of appeals based on challenge IDs
func (t *TestPersister) AppealsByChallengeIDs(challengeIDs []int) ([]*model.Appeal, error) {
	results := []*model.Appeal{}
	for _, challengeID := range challengeIDs {
		appeal, err := t.AppealByChallengeID(challengeID)
		if err == nil {
			results = append(results, appeal)
		}
	}
	return results, nil
}

// CreateAppeal creates a new appeal
func (t *TestPersister) CreateAppeal(appeal *model.Appeal) error {
	challengeID := int(appeal.OriginalChallengeID().Int64())
	if t.Appeals == nil {
		t.Appeals = map[int]*model.Appeal{}
	}
	t.Appeals[challengeID] = appeal
	return nil
}

// UpdateAppeal updates an appeal
func (t *TestPersister) UpdateAppeal(appeal *model.Appeal, updatedFields []string) error {
	challengeID := int(appeal.OriginalChallengeID().Int64())
	if t.Appeals == nil {
		t.Appeals = map[int]*model.Appeal{}
	}
	t.Appeals[challengeID] = appeal
	return nil
}

// TimestampOfLastEventForCron gets last timestamp
func (t *TestPersister) TimestampOfLastEventForCron() (int64, error) {
	return t.Timestamp, nil
}

// UpdateTimestampForCron updates timestamp for cron
func (t *TestPersister) UpdateTimestampForCron(timestamp int64) error {
	t.Timestamp = timestamp
	return nil
}

// EventHashesOfLastTimestampForCron returns the event hashes processed for the last timestamp from cron
func (t *TestPersister) EventHashesOfLastTimestampForCron() ([]string, error) {
	return t.EventHashes, nil
}

// UpdateEventHashesForCron updates the eventHashes saved in cron table
func (t *TestPersister) UpdateEventHashesForCron(eventHashes []string) error {
	t.EventHashes = eventHashes
	return nil
}

// TokenTransfersByTxHash gets a list of token transfers by TxHash
func (t *TestPersister) TokenTransfersByTxHash(txHash common.Hash) (
	[]*model.TokenTransfer, error) {
	purchases, ok := t.TokenTransfersTxHash[txHash.Hex()]
	if !ok {
		return nil, cpersist.ErrPersisterNoResults
	}
	return purchases, nil
}

// TokenTransfersByToAddress gets a list of token transfers by purchaser address
func (t *TestPersister) TokenTransfersByToAddress(addr common.Address) (
	[]*model.TokenTransfer, error) {
	purchases, ok := t.TokenTransfers[addr.Hex()]
	if !ok {
		return nil, cpersist.ErrPersisterNoResults
	}
	return purchases, nil
}

// CreateTokenTransfer creates a new token transfer
func (t *TestPersister) CreateTokenTransfer(purchase *model.TokenTransfer) error {
	addr := purchase.ToAddress().Hex()
	blockData := purchase.BlockData()
	txHash := blockData.TxHash()

	if t.TokenTransfers == nil {
		t.TokenTransfers = map[string][]*model.TokenTransfer{}
	}
	purchases, ok := t.TokenTransfers[addr]
	if !ok {
		t.TokenTransfers[addr] = []*model.TokenTransfer{purchase}
	} else {
		appendedSlice := append(purchases, purchase)
		t.TokenTransfers[addr] = appendedSlice
	}

	if t.TokenTransfersTxHash == nil {
		t.TokenTransfersTxHash = map[string][]*model.TokenTransfer{}
	}
	purchases, ok = t.TokenTransfersTxHash[txHash]
	if !ok {
		t.TokenTransfersTxHash[txHash] = []*model.TokenTransfer{purchase}
	} else {
		appendedSlice := append(purchases, purchase)
		t.TokenTransfersTxHash[txHash] = appendedSlice
	}
	return nil
}

// CreateParameterProposal creates a new parameter proposal
func (t *TestPersister) CreateParameterProposal(paramProposal *model.ParameterProposal) error {
	propID := paramProposal.PropID()
	if t.ParameterProposal == nil {
		t.ParameterProposal = map[[32]byte]*model.ParameterProposal{}
	}
	t.ParameterProposal[propID] = paramProposal
	return nil
}

// ParamProposalByPropID gets a parameter proposal from persistence using propID
func (t *TestPersister) ParamProposalByPropID(propID [32]byte) (*model.ParameterProposal, error) {
	paramProposal, ok := t.ParameterProposal[propID]
	if !ok {
		return nil, cpersist.ErrPersisterNoResults
	}
	return paramProposal, nil
}

// ParamProposalByName gets parameter proposals by name from persistence
func (t *TestPersister) ParamProposalByName(name string, active bool) ([]*model.ParameterProposal, error) {
	proposals := []*model.ParameterProposal{}

	for _, prop := range t.ParameterProposal {
		if name != prop.Name() {
			continue
		}
		if active && prop.Expired() {
			continue
		}
		proposals = append(proposals, prop)

	}
	return proposals, nil
}

// UpdateParamProposal updates parameter propsal in table
func (t *TestPersister) UpdateParamProposal(paramProposal *model.ParameterProposal, updatedFields []string) error {
	propID := paramProposal.PropID()
	t.ParameterProposal[propID] = paramProposal
	return nil
}

// CreateUserChallengeData creates a new UserChallengeData
func (t *TestPersister) CreateUserChallengeData(userChallengeData *model.UserChallengeData) error {
	pollID := int(userChallengeData.PollID().Int64())
	address := userChallengeData.UserAddress().Hex()
	if t.UserChallengeData == nil {
		t.UserChallengeData = map[int]map[string]*model.UserChallengeData{}
	}
	if t.UserChallengeData[pollID] == nil {
		t.UserChallengeData[pollID] = map[string]*model.UserChallengeData{}
	}
	t.UserChallengeData[pollID][address] = userChallengeData
	return nil
}

// UserChallengeDataByCriteria retrieves UserChallengeData based on criteria
func (t *TestPersister) UserChallengeDataByCriteria(criteria *model.UserChallengeDataCriteria) ([]*model.UserChallengeData, error) {
	address := criteria.UserAddress
	pollID := int(criteria.PollID)
	if pollID != 0 && t.UserChallengeData[pollID] == nil {
		return []*model.UserChallengeData{}, nil
	}
	if address != "" && t.UserChallengeData[pollID][address] == nil {
		return []*model.UserChallengeData{}, nil
	}

	return []*model.UserChallengeData{t.UserChallengeData[pollID][address]}, nil
}

// UpdateUserChallengeData updates UserChallengeData in table
func (t *TestPersister) UpdateUserChallengeData(userChallengeData *model.UserChallengeData,
	updatedFields []string, updateWithUserAddress bool) error {
	pollID := int(userChallengeData.PollID().Int64())
	if updateWithUserAddress {
		address := userChallengeData.UserAddress().Hex()
		if t.UserChallengeData[pollID] == nil {
			t.UserChallengeData[pollID] = map[string]*model.UserChallengeData{}
		}
		t.UserChallengeData[pollID][address] = userChallengeData
	} else {
		// NOTE(IS): should go through update fields in updatedfields,
		// but assuming the only usecase is to update pollispassed
		allUSD := t.UserChallengeData[pollID]
		for pollID, user := range allUSD {
			user.SetPollIsPassed(userChallengeData.PollIsPassed())
			allUSD[pollID] = user

		}
	}

	return nil
}

// TestScraper is a testscraper
type TestScraper struct{}

// ScrapeContent scrapes content
func (t *TestScraper) ScrapeContent(uri string) (*model.ScraperContent, error) {
	return &model.ScraperContent{}, nil
}

// ScrapeCivilMetadata scrapes civilmetadata
func (t *TestScraper) ScrapeCivilMetadata(uri string) (*model.ScraperCivilMetadata, error) {
	metadata := model.NewScraperCivilMetadata()
	err := json.Unmarshal([]byte(testCivilMetadata), metadata)
	if err != nil {
		return nil, err
	}
	return metadata, nil
}

// ScrapeMetadata scrapes metadata
func (t *TestScraper) ScrapeMetadata(uri string) (*model.ScraperContentMetadata, error) {
	return &model.ScraperContentMetadata{}, nil
}
