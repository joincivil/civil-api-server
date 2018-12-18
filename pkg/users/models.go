package users

import (
	"errors"

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
	UID               string                 `db:"uid"`
	Email             string                 `db:"email"`
	EthAddress        string                 `db:"eth_address"`
	OnfidoApplicantID string                 `db:"onfido_applicant_id"`
	OnfidoCheckID     string                 `db:"onfido_check_id"`
	KycStatus         string                 `db:"kyc_status"`
	QuizPayload       cpostgres.JsonbPayload `db:"quiz_payload"`
	QuizStatus        string                 `db:"quiz_status"`
	NewsroomData      cpostgres.JsonbPayload `db:"newsroom_data"`
	DateCreated       int64                  `db:"date_created"`
	DateUpdated       int64                  `db:"date_updated"`
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
