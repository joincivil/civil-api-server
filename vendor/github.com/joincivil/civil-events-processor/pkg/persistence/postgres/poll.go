package postgres

import (
	"fmt"
	"math/big"

	"github.com/joincivil/civil-events-processor/pkg/model"
	"github.com/joincivil/go-common/pkg/numbers"
)

const (
	// PollTableName is the name of this table
	PollTableName = "poll"
)

// CreatePollTableQuery returns the query to create the poll table
func CreatePollTableQuery() string {
	return CreatePollTableQueryString(PollTableName)
}

// CreatePollTableQueryString returns the query to create this table
func CreatePollTableQueryString(tableName string) string {
	queryString := fmt.Sprintf(`
        CREATE TABLE IF NOT EXISTS %s(
            poll_id INT PRIMARY KEY,
            poll_type TEXT,
            commit_end_date INT,
            reveal_end_date INT,
            is_passed BOOL,
            vote_quorum NUMERIC,
            votes_for NUMERIC,
            votes_against NUMERIC,
            last_updated_timestamp INT
        );
    `, tableName)
	return queryString
}

// PollTableIndices returns the query to create indices for this table
func PollTableIndices() string {
	return CreatePollTableIndicesString(PollTableName)
}

// CreatePollTableIndicesString returns the query to create indices for this table
func CreatePollTableIndicesString(tableName string) string {
	// queryString := fmt.Sprintf(`
	// `, tableName)
	// return queryString
	return ""
}

// Poll is model for poll object
type Poll struct {
	PollID uint64 `db:"poll_id"`

	PollType string `db:"poll_type"`

	CommitEndDate int64 `db:"commit_end_date"`

	RevealEndDate int64 `db:"reveal_end_date"`

	IsPassed bool `db:"is_passed"`

	VoteQuorum uint64 `db:"vote_quorum"`

	VotesFor float64 `db:"votes_for"`

	VotesAgainst float64 `db:"votes_against"`

	LastUpdatedDateTs int64 `db:"last_updated_timestamp"`
}

// NewPoll creates a new poll
func NewPoll(pollData *model.Poll) *Poll {
	poll := &Poll{}
	poll.PollID = pollData.PollID().Uint64()
	poll.PollType = pollData.PollType()
	poll.CommitEndDate = pollData.CommitEndDate().Int64()
	poll.RevealEndDate = pollData.RevealEndDate().Int64()
	poll.IsPassed = pollData.IsPassed()
	poll.VoteQuorum = pollData.VoteQuorum().Uint64()
	poll.VotesFor = numbers.BigIntToFloat64(pollData.VotesFor())
	poll.VotesAgainst = numbers.BigIntToFloat64(pollData.VotesAgainst())
	poll.LastUpdatedDateTs = pollData.LastUpdatedDateTs()
	return poll
}

// DbToPollData converts a db poll to a model poll
func (p *Poll) DbToPollData() *model.Poll {
	poll := model.NewPoll(
		new(big.Int).SetUint64(p.PollID),
		big.NewInt(p.CommitEndDate),
		big.NewInt(p.RevealEndDate),
		new(big.Int).SetUint64(p.VoteQuorum),
		numbers.Float64ToBigInt(p.VotesFor),
		numbers.Float64ToBigInt(p.VotesAgainst),
		p.LastUpdatedDateTs,
	)
	poll.SetPollType(p.PollType)
	poll.SetIsPassed(p.IsPassed)
	return poll
}
