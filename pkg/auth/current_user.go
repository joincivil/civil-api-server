package auth

import (
	"github.com/joincivil/civil-api-server/pkg/invoicing"
)

// CurrentUser represents the current API user
type CurrentUser struct {
	Email            string
	invoicePersister *invoicing.PostgresPersister
}

// GetCurrentUser returns the current user
func GetCurrentUser(sub string, invoicePersister *invoicing.PostgresPersister) (*CurrentUser, error) {

	return &CurrentUser{
		Email:            sub,
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
