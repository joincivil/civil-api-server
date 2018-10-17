package postgres

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	crawlerpostgres "github.com/joincivil/civil-events-crawler/pkg/persistence/postgres"
	"github.com/joincivil/civil-events-processor/pkg/model"
	"math/big"
)

// CreateChallengeTableQuery returns the query to create the governance_event table
func CreateChallengeTableQuery() string {
	return CreateChallengeTableQueryString("challenge")
}

// CreateChallengeTableQueryString returns the query to create this table
func CreateChallengeTableQueryString(tableName string) string {
	queryString := fmt.Sprintf(`
        CREATE TABLE IF NOT EXISTS %s(
            challenge_id INT PRIMARY KEY,
            listing_address TEXT,
            statement TEXT,
            reward_pool NUMERIC,
            challenger TEXT,
            resolved BOOL,
            stake NUMERIC,
            total_tokens NUMERIC,
            poll JSONB,
            request_appeal_expiry NUMERIC,
            appeal JSONB,
            last_updated_timestamp INT
        );
    `, tableName)
	return queryString
}

// Challenge is postgres definition of model.Challenge
type Challenge struct {
	ChallengeID uint64 `db:"challenge_id"`

	ListingAddress string `db:"listing_address"`

	Statement string `db:"statement"`

	RewardPool float64 `db:"reward_pool"`

	Challenger string `db:"challenger"`

	Resolved bool `db:"resolved"`

	Stake float64 `db:"stake"`

	TotalTokens float64 `db:"total_tokens"`

	Poll crawlerpostgres.JsonbPayload `db:"poll"`

	RequestAppealExpiry int64 `db:"request_appeal_expiry"`

	Appeal crawlerpostgres.JsonbPayload `db:"appeal"`

	LastUpdatedDateTs int64 `db:"last_updated_timestamp"`
}

// NewChallenge creates a new postgres challenge
// NOTE(IS): When values aren't available they are nil, so convert these to 0 (?)
func NewChallenge(challengeData *model.Challenge) *Challenge {
	challenge := &Challenge{}
	challenge.ChallengeID = challengeData.ChallengeID().Uint64()
	challenge.ListingAddress = challengeData.ListingAddress().Hex()
	challenge.Statement = challengeData.Statement()
	challenge.Challenger = challengeData.Challenger().Hex()
	challenge.Resolved = challengeData.Resolved()
	challenge.LastUpdatedDateTs = challengeData.LastUpdatedDateTs()
	if challengeData.RewardPool() != nil {
		challenge.RewardPool = BigIntToFloat64(challengeData.RewardPool())
	} else {
		challenge.RewardPool = 0
	}
	if challengeData.Stake() != nil {
		challenge.Stake = BigIntToFloat64(challengeData.Stake())
	} else {
		challenge.Stake = 0
	}
	if challengeData.TotalTokens() != nil {
		challenge.TotalTokens = BigIntToFloat64(challengeData.TotalTokens())
	} else {
		challenge.TotalTokens = 0
	}
	if challengeData.RequestAppealExpiry() != nil {
		challenge.RequestAppealExpiry = challengeData.RequestAppealExpiry().Int64()
	} else {
		challenge.RequestAppealExpiry = 0
	}
	challenge.Poll = make(crawlerpostgres.JsonbPayload)
	challenge.fillPoll(challengeData.Poll())
	challenge.Appeal = make(crawlerpostgres.JsonbPayload)
	challenge.fillAppeal(challengeData.Appeal())

	return challenge
}

// DbToChallengeData creates a model.Challenge from postgres.Challenge
// NOTE: jsonb payloads are stored in DB as map[string]interface{}, Postgres converts some fields, see notes in function.
func (c *Challenge) DbToChallengeData() *model.Challenge {
	var challengeID big.Int
	challengeID.SetUint64(c.ChallengeID)
	listingAddress := common.HexToAddress(c.ListingAddress)
	rewardPool := Float64ToBigInt(c.RewardPool)
	challenger := common.HexToAddress(c.Challenger)
	stake := Float64ToBigInt(c.Stake)
	totalTokens := Float64ToBigInt(c.TotalTokens)
	poll := c.createPollData()
	// appeal := c.createAppealData()
	return model.NewChallenge(&challengeID, listingAddress, c.Statement, rewardPool, challenger, c.Resolved,
		stake, totalTokens, poll, big.NewInt(c.RequestAppealExpiry), c.LastUpdatedDateTs)
}

func (c *Challenge) fillPoll(poll *model.Poll) {
	if poll != nil {
		c.Poll["commitEndDate"] = poll.CommitEndDate()
		c.Poll["revealEndDate"] = poll.RevealEndDate()
		c.Poll["voteQuorum"] = poll.VoteQuorum()
		c.Poll["votesFor"] = poll.VotesFor()
		c.Poll["votesAgainst"] = poll.VotesAgainst()
	} else {
		c.Poll["commitEndDate"] = int64(0)
		c.Poll["revealEndDate"] = int64(0)
		c.Poll["voteQuorum"] = uint64(0)
		c.Poll["votesFor"] = uint64(0)
		c.Poll["votesAgainst"] = uint64(0)
	}

}

// NOTE(IS): for now this is empty
// TODO(IS): fill this out
func (c *Challenge) fillAppeal(appeal model.Appeal) {

}

func (c *Challenge) createPollData() *model.Poll {
	// NOTE(IS): I think all these vals will be stored in DB as float64. Will check
	commitEndDate := int64(c.Poll["commitEndDate"].(float64))
	revealEndDate := int64(c.Poll["revealEndDate"].(float64))
	voteQuorum := uint64(c.Poll["voteQuorum"].(float64))
	votesFor := uint64(c.Poll["votesFor"].(float64))
	votesAgainst := uint64(c.Poll["votesAgainst"].(float64))
	return model.NewPoll(commitEndDate, revealEndDate, voteQuorum, votesFor, votesAgainst)
}
