package postgres

import (
	"fmt"
	"github.com/joincivil/civil-events-processor/pkg/model"
	"math/big"
)

const (
	defaultPollTableName = "poll"
)

// CreatePollTableQuery returns the query to create the poll table
func CreatePollTableQuery() string {
	return CreatePollTableQueryString(defaultPollTableName)
}

// CreatePollTableQueryString returns the query to create this table
func CreatePollTableQueryString(tableName string) string {
	queryString := fmt.Sprintf(`
        CREATE TABLE IF NOT EXISTS %s(
            poll_id INT PRIMARY KEY,
            commit_end_date INT,
            reveal_end_date INT,
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
	return CreatePollTableIndicesString(defaultPollTableName)
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

	CommitEndDate int64 `db:"commit_end_date"`

	RevealEndDate int64 `db:"reveal_end_date"`

	VoteQuorum uint64 `db:"vote_quorum"`

	VotesFor uint64 `db:"votes_for"`

	VotesAgainst uint64 `db:"votes_against"`

	LastUpdatedDateTs int64 `db:"last_updated_timestamp"`
}

// NewPoll creates a new poll
func NewPoll(pollData *model.Poll) *Poll {
	poll := &Poll{}
	poll.PollID = pollData.PollID().Uint64()
	poll.CommitEndDate = pollData.CommitEndDate().Int64()
	poll.RevealEndDate = pollData.RevealEndDate().Int64()
	poll.VoteQuorum = pollData.VoteQuorum().Uint64()
	poll.VotesFor = pollData.VotesFor().Uint64()
	poll.VotesAgainst = pollData.VotesAgainst().Uint64()
	poll.LastUpdatedDateTs = pollData.LastUpdatedDateTs()
	return poll
}

// DbToPollData converts a db poll to a model poll
func (p *Poll) DbToPollData() *model.Poll {
	return model.NewPoll(
		new(big.Int).SetUint64(p.PollID),
		big.NewInt(p.CommitEndDate),
		big.NewInt(p.RevealEndDate),
		new(big.Int).SetUint64(p.VoteQuorum),
		new(big.Int).SetUint64(p.VotesFor),
		new(big.Int).SetUint64(p.VotesAgainst),
		p.LastUpdatedDateTs,
	)
}
