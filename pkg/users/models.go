package users

import (
	"errors"

	"github.com/lib/pq"

	"github.com/ethereum/go-ethereum/common"
	"github.com/joincivil/civil-api-server/pkg/channels"
	cpostgres "github.com/joincivil/go-common/pkg/persistence/postgres"
	uuid "github.com/satori/go.uuid"
)

const (
	// UserKycStatusInProgress is a user with an in progress KYC
	UserKycStatusInProgress = "in_progress"
	// UserKycStatusPassed is a user with passed KYC
	UserKycStatusPassed = "passed"
	// UserKycStatusFailed is a user with failed KYC
	UserKycStatusFailed = "failed"
	// UserKycStatusNeedsReview is a user that needs additional human review
	UserKycStatusNeedsReview = "needs_review"
)

// User represents a Civil User
type User struct {
	UID                  string                 `db:"uid"`
	Email                string                 `db:"email"`
	EthAddress           string                 `db:"eth_address"`
	QuizPayload          cpostgres.JsonbPayload `db:"quiz_payload"`
	QuizStatus           string                 `db:"quiz_status"`
	DateCreated          int64                  `db:"date_created"`
	DateUpdated          int64                  `db:"date_updated"`
	AppReferral          string                 `db:"app_refer"`
	NewsroomStep         int                    `db:"nr_step"`
	NewsroomFurthestStep int                    `db:"nr_far_step"`
	NewsroomLastSeen     int64                  `db:"nr_last_seen"`

	// PurchaseTxHashes is a list of txhashes of all the token purchases for
	// this user
	PurchaseTxHashes pq.StringArray `db:"purchase_txhashes"`

	// CivilianWhitelistTxID is the txHash of the whitelisted transaction
	// which unlocks the user's tokens
	CivilianWhitelistTxID string `db:"civilian_whitelist_tx_id"`

	// AssocNewsroomAddr is a list of newsroom addresses of which the user is associated
	AssocNewsoomAddr pq.StringArray `db:"assoc_nr_addr"`

	// UserChannelEmailPromptSeen whether or not the user has seen the prompt to set their user channel email address
	UserChannelEmailPromptSeen bool `db:"uc_email_prompt_seen"`
}

// TokenControllerUpdater describes methods that the user service will use to manage the whitelists a user is a member of
type TokenControllerUpdater interface {
	AddToCivilians(addr common.Address) (common.Hash, error)
}

// UserChannelHelper describes the methods that the user service will use to create channels for users
type UserChannelHelper interface {
	CreateUserChannel(userID string) (*channels.Channel, error)
}

// GenerateUID generates and set the UID field for the user.  Will only
// generate a new one if UID field is empty.
func (u *User) GenerateUID() error {
	if u.UID != "" {
		return errors.New("Already has a UID")
	}
	code, err := uuid.NewV4()
	if err != nil {
		return err
	}
	u.UID = code.String()
	return nil
}
