package postgres

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/joincivil/civil-events-processor/pkg/model"
	"github.com/joincivil/go-common/pkg/bytes"
	"github.com/joincivil/go-common/pkg/numbers"
)

const (
	// ParameterProposalTableBaseName is the table name for this model
	ParameterProposalTableBaseName = "parameter_proposal"
)

// CreateParameterProposalTableQuery returns the query to create this table
func CreateParameterProposalTableQuery(tableName string) string {
	queryString := fmt.Sprintf(`
        CREATE TABLE IF NOT EXISTS %s(
            prop_id TEXT PRIMARY KEY,
            name TEXT,
            value NUMERIC,
            deposit NUMERIC,
            app_expiry INT,
            challenge_id INT,
            proposer TEXT,
            accepted BOOL,
            expired BOOL,
            last_updated_timestamp INT
        );
    `, tableName)
	return queryString
}

// ParameterProposal is postgres definition of model.ParameterProposal
type ParameterProposal struct {
	Name string `db:"name"`

	Value float64 `db:"value"`

	PropID string `db:"prop_id"`

	Deposit float64 `db:"deposit"`

	AppExpiry int64 `db:"app_expiry"`

	ChallengeID int64 `db:"challenge_id"`

	Proposer string `db:"proposer"`

	Accepted bool `db:"accepted"`

	Expired bool `db:"expired"`

	LastUpdatedDateTs int64 `db:"last_updated_timestamp"`
}

// NewParameterProposal is the model definition for parameter_proposal table
func NewParameterProposal(parameterProposal *model.ParameterProposal) *ParameterProposal {
	value := numbers.BigIntToFloat64(parameterProposal.Value())
	propID := bytes.Byte32ToHexString(parameterProposal.PropID())
	deposit := numbers.BigIntToFloat64(parameterProposal.Deposit())
	return &ParameterProposal{
		Name:              parameterProposal.Name(),
		Value:             value,
		PropID:            propID,
		Deposit:           deposit,
		AppExpiry:         parameterProposal.AppExpiry().Int64(),
		ChallengeID:       parameterProposal.ChallengeID().Int64(),
		Proposer:          parameterProposal.Proposer().Hex(),
		Accepted:          parameterProposal.Accepted(),
		Expired:           parameterProposal.Expired(),
		LastUpdatedDateTs: parameterProposal.LastUpdatedDateTs(),
	}
}

// DbToParameterProposalData creates a model.ParameterProposal from postgres ParameterProposal
func (p *ParameterProposal) DbToParameterProposalData() (*model.ParameterProposal, error) {
	value := numbers.Float64ToBigInt(p.Value)
	propID, err := bytes.HexStringToByte32(p.PropID)
	if err != nil {
		return nil, err
	}
	deposit := numbers.Float64ToBigInt(p.Deposit)
	appExpiry := big.NewInt(p.AppExpiry)
	challengeID := big.NewInt(p.ChallengeID)
	proposer := common.HexToAddress(p.Proposer)
	parameterProposalParams := &model.ParameterProposalParams{
		Name:              p.Name,
		Value:             value,
		PropID:            propID,
		Deposit:           deposit,
		AppExpiry:         appExpiry,
		ChallengeID:       challengeID,
		Proposer:          proposer,
		Accepted:          p.Accepted,
		Expired:           p.Expired,
		LastUpdatedDateTs: p.LastUpdatedDateTs,
	}
	return model.NewParameterProposal(parameterProposalParams), nil
}
