package postgres

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/joincivil/civil-events-processor/pkg/model"
	"math/big"
)

// CreateAppealTableQuery returns the query to create the governance_event table
func CreateAppealTableQuery() string {
	return CreateAppealTableQueryString("appeal")
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
            last_updated_timestamp INT
        );
    `, tableName)
	return queryString
}

// Appeal is model for appeal object
type Appeal struct {
	OriginalChallengeID int64 `db:"original_challenge_id"`

	Requester string `db:"requester"`

	AppealFeePaid uint64 `db:"appeal_fee_paid"`

	AppealPhaseExpiry int64 `db:"appeal_phase_expiry"`

	AppealGranted bool `db:"appeal_granted"`

	AppealOpenToChallengeExpiry int64 `db:"appeal_open_to_challenge_expiry"`

	Statement string `db:"statement"`

	AppealChallengeID uint64 `db:"appeal_challenge_id"`

	LastUpdatedDateTs int64 `db:"last_updated_timestamp"`
}

// NewAppeal creates a new appeal
func NewAppeal(appealData *model.Appeal) *Appeal {
	appeal := &Appeal{}
	appeal.OriginalChallengeID = appealData.OriginalChallengeID().Int64()
	appeal.Requester = appealData.Requester().Hex()
	appeal.AppealFeePaid = appealData.AppealFeePaid().Uint64()
	appeal.AppealPhaseExpiry = appealData.AppealPhaseExpiry().Int64()
	appeal.AppealGranted = appealData.AppealGranted()
	appeal.LastUpdatedDateTs = appealData.LastUpdatedDateTs()
	appeal.Statement = appealData.Statement()

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
		new(big.Int).SetUint64(a.AppealFeePaid),
		big.NewInt(a.AppealPhaseExpiry),
		a.AppealGranted,
		a.Statement,
		a.LastUpdatedDateTs,
	)

	appeal.SetAppealChallengeID(new(big.Int).SetUint64(a.AppealChallengeID))
	appeal.SetAppealOpenToChallengeExpiry(big.NewInt(a.AppealOpenToChallengeExpiry))
	return appeal
}
