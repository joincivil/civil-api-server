package users

import (
	"errors"

	crawlerpg "github.com/joincivil/civil-events-crawler/pkg/persistence/postgres"
	uuid "github.com/satori/go.uuid"
)

// User represents a Civil User
type User struct {
	UID               string                 `db:"uid"`
	Email             string                 `db:"email"`
	EthAddress        string                 `db:"eth_address"`
	OnfidoApplicantID string                 `db:"onfido_applicant_id"`
	KycStatus         string                 `db:"kyc_status"`
	QuizPayload       crawlerpg.JsonbPayload `db:"quiz_payload"`
	QuizStatus        string                 `db:"quiz_status"`
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
