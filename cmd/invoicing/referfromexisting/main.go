package main

// This script sends referral emails to all invoices that don't have a
// referral code yet.  Will generate and update the invoice with a new generated
// referral code.

import (
	"fmt"
	"strings"

	"github.com/kelseyhightower/envconfig"

	"github.com/joincivil/civil-api-server/pkg/invoicing"
	"github.com/joincivil/civil-api-server/pkg/utils"
)

// Config is the struct that handles configuration for this script
type Config struct {
	PostgresAddress string
	PostgresPort    int
	PostgresUser    string
	PostgresPw      string
	PostgresDbName  string
	SendgridKey     string
}

func main() {
	config := &Config{}
	err := envconfig.Process("config", config)
	if err != nil {
		fmt.Printf("Failed to grab config envars: err: %v", err)
		return
	}

	persister, err := invoicing.NewPostgresPersister(
		config.PostgresAddress,
		config.PostgresPort,
		config.PostgresUser,
		config.PostgresPw,
		config.PostgresDbName,
	)
	if err != nil {
		fmt.Printf("error generating persister: err: %v", err)
		return
	}

	invoices, err := persister.Invoices("", "", "", "")
	if err != nil {
		fmt.Printf("error getting invoices: err: %v", err)
		return
	}

	emailer := utils.NewEmailer(config.SendgridKey)

	for _, invoice := range invoices {
		// If code already exists, skip that user
		if invoice.ReferralCode != "" {
			// fmt.Printf("referral code found, skipping: %v\n", invoice.Email)
			continue
		}
		if invoice.InvoiceStatus == invoicing.InvoiceStatusCanceled {
			// fmt.Printf("invoice was cancelled, skipping: %v\n", invoice.Email)
			continue
		}

		// Generate a new code for a user
		referralCode, err := invoicing.GenerateReferralCode()
		if err != nil {
			fmt.Printf("error generating referral code: err: %v\n", err)
			continue
		}

		// Pull out the first name
		var name string
		nameSplit := strings.Split(invoice.Name, " ")

		if len(nameSplit) > 0 {
			name = nameSplit[0]
		} else {
			name = invoice.Name
		}

		// Send the referral email
		req := &invoicing.Request{
			FirstName: name,
			Email:     invoice.Email,
		}

		fmt.Printf("info = %v, %v, %v\n", req.FirstName, req.Email, referralCode)
		invoicing.SendReferralProgramEmail(emailer, req, referralCode)

		// Update the invoice w new referral code
		invoice.ReferralCode = referralCode
		updatedFields := []string{"ReferralCode"}
		err = persister.UpdateInvoice(invoice, updatedFields)
		if err != nil {
			fmt.Printf("error updating referral code: err: %v", err)
		}
	}
}
