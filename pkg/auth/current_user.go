package auth

import (
	"github.com/joincivil/civil-api-server/pkg/invoicing"
	"github.com/joincivil/civil-api-server/pkg/tokenfoundry"
)

// CurrentUser represents the current API user
type CurrentUser struct {
	Email            string
	tokenFoundry     *tokenfoundry.API
	invoicePersister *invoicing.PostgresPersister
}

// GetCurrentUser returns the current user
func GetCurrentUser(sub string, invoicePersister *invoicing.PostgresPersister, tokenFoundry *tokenfoundry.API) (*CurrentUser, error) {

	return &CurrentUser{
		Email:            sub,
		tokenFoundry:     tokenFoundry,
		invoicePersister: invoicePersister,
	}, nil
}

// Invoices returns a list of `invoicing.PostgresInvoice` created by the user
func (u *CurrentUser) Invoices() ([]*invoicing.PostgresInvoice, error) {

	invoices, error := u.invoicePersister.Invoices("", u.Email, "", "")
	if error != nil {
		return nil, error
	}

	return invoices, nil
}

// IsTokenFoundryRegistered determines if the CurrentUser is registered to buy CVL on TokenFoundry
func (u *CurrentUser) IsTokenFoundryRegistered() (bool, error) {

	return u.tokenFoundry.GetKYCStatus(u.Email)
}
