package postgres

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/joincivil/civil-events-processor/pkg/model"
	"github.com/joincivil/go-common/pkg/numbers"
)

const (
	defaultAppealTableName = "appeal"
)

// CreateAppealTableQuery returns the query to create the appeal table
func CreateAppealTableQuery() string {
	return CreateAppealTableQueryString(defaultAppealTableName)
}

// CreateAppealTableQueryString returns the query to create this table
func CreateAppealTableQueryString(tableName string) string {
	queryString := fmt.Sprintf(`
        CREATE TABLE IF NOT EXISTS %s(
            original_challenge_id INT PRIMARY KEY,
            requester TEXT,
            appeal_fee_paid NUMERIC,
            appeal_phase_expiry INT,
            appeal_granted BOOLEAN,
            appeal_open_to_challenge_expiry INT,
            statement TEXT,
            appeal_challenge_id INT,
            last_updated_timestamp INT,
            appeal_granted_statement_uri TEXT
        );
    `, tableName)
	return queryString
}

// AppealTableIndices returns the query to create indices for this table
func AppealTableIndices() string {
	return CreateAppealTableIndicesString(defaultAppealTableName)
}

// CreateAppealTableIndicesString returns the query to create indices this table
func CreateAppealTableIndicesString(tableName string) string {
	// queryString := fmt.Sprintf(`
	// `, tableName)
	// return queryString
	return ""
}

// Appeal is model for appeal object
type Appeal struct {
	OriginalChallengeID int64 `db:"original_challenge_id"`

	Requester string `db:"requester"`

	AppealFeePaid float64 `db:"appeal_fee_paid"`

	AppealPhaseExpiry int64 `db:"appeal_phase_expiry"`

	AppealGranted bool `db:"appeal_granted"`

	AppealOpenToChallengeExpiry int64 `db:"appeal_open_to_challenge_expiry"`

	AppealGrantedStatementURI string `db:"appeal_granted_statement_uri"`

	Statement string `db:"statement"`

	AppealChallengeID uint64 `db:"appeal_challenge_id"`

	LastUpdatedDateTs int64 `db:"last_updated_timestamp"`
}

// NewAppeal creates a new appeal
func NewAppeal(appealData *model.Appeal) *Appeal {
	appeal := &Appeal{}
	appeal.OriginalChallengeID = appealData.OriginalChallengeID().Int64()
	appeal.Requester = appealData.Requester().Hex()
	appeal.AppealFeePaid = numbers.BigIntToFloat64(appealData.AppealFeePaid())
	appeal.AppealPhaseExpiry = appealData.AppealPhaseExpiry().Int64()
	appeal.AppealGranted = appealData.AppealGranted()
	appeal.LastUpdatedDateTs = appealData.LastUpdatedDateTs()
	appeal.Statement = appealData.Statement()
	appeal.AppealGrantedStatementURI = appealData.AppealGrantedStatementURI()
	// NOTE(IS): following fields can be nil so set to 0
	if appealData.AppealOpenToChallengeExpiry() != nil {
		appeal.AppealOpenToChallengeExpiry = appealData.AppealOpenToChallengeExpiry().Int64()
	} else {
		appeal.AppealOpenToChallengeExpiry = int64(0)
	}
	if appealData.AppealChallengeID() != nil {
		appeal.AppealChallengeID = appealData.AppealChallengeID().Uint64()
	} else {
		appeal.AppealChallengeID = uint64(0)
	}

	return appeal
}

// DbToAppealData creates a model.Appeal from postgres.Appeal
func (a *Appeal) DbToAppealData() *model.Appeal {
	appeal := model.NewAppeal(
		big.NewInt(a.OriginalChallengeID),
		common.HexToAddress(a.Requester),
		numbers.Float64ToBigInt(a.AppealFeePaid),
		big.NewInt(a.AppealPhaseExpiry),
		a.AppealGranted,
		a.Statement,
		a.LastUpdatedDateTs,
		a.AppealGrantedStatementURI,
	)

	appeal.SetAppealChallengeID(new(big.Int).SetUint64(a.AppealChallengeID))
	appeal.SetAppealOpenToChallengeExpiry(big.NewInt(a.AppealOpenToChallengeExpiry))
	return appeal
}
