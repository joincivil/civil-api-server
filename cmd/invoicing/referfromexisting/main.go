package main

// This script sends referral emails to all invoices that don't have a
// referral code yet.  Will generate and update the invoice with a new generated
// referral code.

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/kelseyhightower/envconfig"

	"github.com/joincivil/civil-api-server/pkg/invoicing"
	// "github.com/joincivil/civil-api-server/pkg/utils"
)

const (
	envVarPrefixConfig = "config"

	usageListFormat = `The command is configured via environment vars only. The following environment variables can be used:
{{range .}}
{{usage_key .}}
  description: {{usage_description .}}
  type:        {{usage_type .}}
  default:     {{usage_default .}}
  required:    {{usage_required .}}
{{end}}
`
)

// Config is the struct that handles configuration for this script
type Config struct {
	PostgresAddress string `split_words:"true"`
	PostgresPort    int    `split_words:"true"`
	PostgresUser    string `split_words:"true"`
	PostgresPw      string `split_words:"true"`
	PostgresDbName  string `split_words:"true"`
	SendgridKey     string `split_words:"true"`
}

// OutputUsage prints the usage string to os.Stdout
func (c *Config) OutputUsage() {
	tabs := tabwriter.NewWriter(os.Stdout, 1, 0, 4, ' ', 0)
	_ = envconfig.Usagef(envVarPrefixConfig, c, tabs, usageListFormat) // nolint: gosec
	_ = tabs.Flush()                                                   // nolint: gosec
}

func main() {
	config := &Config{}
	flag.Usage = func() {
		config.OutputUsage()
		os.Exit(0)
	}
	flag.Parse()

	err := envconfig.Process(envVarPrefixConfig, config)
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

	// emailer := utils.NewEmailer(config.SendgridKey)

	for _, invoice := range invoices {
		// If code already exists, skip that user
		if invoice.ReferralCode != "" {
			fmt.Printf("referral code found, skipping: %v\n", invoice.Email)
			continue
		}
		if invoice.InvoiceStatus == invoicing.InvoiceStatusCanceled {
			fmt.Printf("invoice was cancelled, skipping: %v\n", invoice.Email)
			continue
		}
		// May have already sent the referral email, so skip it
		if invoice.EmailState >= invoicing.EmailStateSentReferral {
			fmt.Printf("email does not need to be sent, skipping: %v\n", invoice.Email)
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
		// invoicing.SendReferralProgramEmail(emailer, req, referralCode)

		// Update the invoice w new referral code
		invoice.ReferralCode = referralCode
		updatedFields := []string{"ReferralCode"}

		// Update the email state
		invoice.EmailState = invoicing.EmailStateSentReferral
		updatedFields = append(updatedFields, "EmailState")

		err = persister.UpdateInvoice(invoice, updatedFields)
		if err != nil {
			fmt.Printf("error updating with referral data: err: %v", err)
		}
	}
}
