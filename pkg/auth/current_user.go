package auth

import (
	"github.com/joincivil/civil-api-server/pkg/invoicing"
	"github.com/joincivil/civil-api-server/pkg/kyc"
	"github.com/joincivil/civil-api-server/pkg/tokenfoundry"
	pmodels "github.com/joincivil/civil-events-processor/pkg/model"
)

// CurrentUser represents the current API user
type CurrentUser struct {
	Email            string
	tokenFoundry     *tokenfoundry.API
	invoicePersister *invoicing.PostgresPersister
	kycPersister     *kyc.PostgresPersister
}

// GetCurrentUser returns the current user
func GetCurrentUser(sub string, invoicePersister *invoicing.PostgresPersister,
	kycPersister *kyc.PostgresPersister, tokenFoundry *tokenfoundry.API) (*CurrentUser, error) {

	return &CurrentUser{
		Email:            sub,
		tokenFoundry:     tokenFoundry,
		invoicePersister: invoicePersister,
		kycPersister:     kycPersister,
	}, nil
}

// Invoices returns a list of `invoicing.PostgresInvoice` created by the user
func (u *CurrentUser) Invoices() ([]*invoicing.PostgresInvoice, error) {

	invoices, err := u.invoicePersister.Invoices("", u.Email, "", "")
	if err != nil {
		return nil, err
	}

	return invoices, nil
}

// EthAddress returns the eth address of the current user from the kyc user data
func (u *CurrentUser) EthAddress() (string, error) {
	user, err := u.kycPersister.User(&kyc.UserCriteria{
		Email: u.Email,
	})
	if err != nil {
		// Return an empty eth address if no results found.
		if err == pmodels.ErrPersisterNoResults {
			return "", nil
		}
		return "", err
	}
	return user.EthAddress, nil
}

// IsTokenFoundryRegistered determines if the CurrentUser is registered to buy CVL on TokenFoundry
func (u *CurrentUser) IsTokenFoundryRegistered() (bool, error) {

	return u.tokenFoundry.GetKYCStatus(u.Email)
}
