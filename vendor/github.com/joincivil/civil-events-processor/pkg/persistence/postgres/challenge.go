package postgres

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/joincivil/civil-events-processor/pkg/model"
	"github.com/joincivil/go-common/pkg/numbers"
)

const (
	// ChallengeTableBaseName is the type of table this code defines
	ChallengeTableBaseName = "challenge"
)

// CreateChallengeTableQuery returns the query to create this table
func CreateChallengeTableQuery(tableName string) string {
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
            request_appeal_expiry NUMERIC,
            challenge_type TEXT,
            last_updated_timestamp INT
        );
    `, tableName)
	return queryString
}

// CreateChallengeTableIndicesQuery returns the query to create indices this table
func CreateChallengeTableIndicesQuery(tableName string) string {
	queryString := fmt.Sprintf(`
		CREATE INDEX IF NOT EXISTS challenge_addr_idx ON %s (listing_address);
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

	RequestAppealExpiry int64 `db:"request_appeal_expiry"`

	ChallengeType string `db:"challenge_type"`

	LastUpdatedDateTs int64 `db:"last_updated_timestamp"`
}

// NewChallenge creates a new postgres challenge
func NewChallenge(challengeData *model.Challenge) *Challenge {
	challenge := &Challenge{}
	challenge.ChallengeID = challengeData.ChallengeID().Uint64()
	challenge.ListingAddress = challengeData.ListingAddress().Hex()
	challenge.Statement = challengeData.Statement()
	challenge.Challenger = challengeData.Challenger().Hex()
	challenge.Resolved = challengeData.Resolved()
	challenge.LastUpdatedDateTs = challengeData.LastUpdatedDateTs()
	if challengeData.RewardPool() != nil {
		challenge.RewardPool = numbers.BigIntToFloat64(challengeData.RewardPool())
	} else {
		challenge.RewardPool = 0
	}
	if challengeData.Stake() != nil {
		challenge.Stake = numbers.BigIntToFloat64(challengeData.Stake())
	} else {
		challenge.Stake = 0
	}
	if challengeData.TotalTokens() != nil {
		challenge.TotalTokens = numbers.BigIntToFloat64(challengeData.TotalTokens())
	} else {
		challenge.TotalTokens = 0
	}
	if challengeData.RequestAppealExpiry() != nil {
		challenge.RequestAppealExpiry = challengeData.RequestAppealExpiry().Int64()
	} else {
		challenge.RequestAppealExpiry = 0
	}
	challenge.ChallengeType = challengeData.ChallengeType()
	return challenge
}

// DbToChallengeData creates a model.Challenge from postgres.Challenge
func (c *Challenge) DbToChallengeData() *model.Challenge {
	challengeID := new(big.Int).SetUint64(c.ChallengeID)
	listingAddress := common.HexToAddress(c.ListingAddress)
	rewardPool := numbers.Float64ToBigInt(c.RewardPool)
	challenger := common.HexToAddress(c.Challenger)
	stake := numbers.Float64ToBigInt(c.Stake)
	totalTokens := numbers.Float64ToBigInt(c.TotalTokens)
	return model.NewChallenge(challengeID, listingAddress, c.Statement, rewardPool, challenger, c.Resolved,
		stake, totalTokens, big.NewInt(c.RequestAppealExpiry), c.ChallengeType, c.LastUpdatedDateTs)
}
