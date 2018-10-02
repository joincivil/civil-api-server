package main

// This script finds all the unpaid invoices and sends them a nudge email.

import (
	"flag"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/kelseyhightower/envconfig"

	"github.com/joincivil/civil-api-server/pkg/invoicing"
	"github.com/joincivil/civil-api-server/pkg/utils"
)

const (
	envVarPrefixConfig   = "config"
	nudgeEmailTemplateID = "d-e55578cdaa6b4651abdf2a1f2e0c0cdf"

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

func invoiceLink(invoiceID string) string {
	invoiceURL := fmt.Sprintf("https://checkbook.io/invoice/%v", invoiceID)
	invoiceLink := fmt.Sprintf("<a href=\"%v\"><b>this link</b></a>", invoiceURL)
	return invoiceLink
}

func sendNudgeEmail(emailer *utils.Emailer, email string, name string, invoiceID string) {
	templateData := utils.TemplateData{}
	templateData["name"] = name
	templateData["invoice_link"] = invoiceLink(invoiceID)

	emailReq := &utils.SendTemplateEmailRequest{
		ToName:  name,
		ToEmail: email,
		// FromName:     "The Civil Media Company",
		FromName:     "Christine from Civil",
		FromEmail:    "support@civil.co",
		TemplateID:   nudgeEmailTemplateID,
		TemplateData: templateData,
		AsmGroupID:   7395,
	}
	err := emailer.SendTemplateEmail(emailReq)
	if err != nil {
		fmt.Printf("Error sending referral email: err: %v\n", err)
	}
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
	fmt.Printf("config = %v\n", config)

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

	// Store the most recent invoice for an email address.
	emailToInvoice := map[string]*invoicing.PostgresInvoice{}

	for _, invoice := range invoices {
		if invoice.InvoiceStatus != invoicing.InvoiceStatusUnpaid {
			fmt.Printf("invoice is not unpaid, skipping: %v\n", invoice.Email)
			continue
		}
		if invoice.InvoiceID == "" {
			fmt.Printf("invoice id was empty, skipping: %v\n", invoice.Email)
			continue
		}

		existingInvoice, ok := emailToInvoice[invoice.Email]

		// If invoice already there, figure out which one is the most recent
		// and use that one.
		if ok {
			if existingInvoice.DateCreated < invoice.DateCreated {
				emailToInvoice[invoice.Email] = invoice
			}
			continue
		}

		emailToInvoice[invoice.Email] = invoice
	}

	for _, invoice := range emailToInvoice {
		// Send the nudge email
		sendNudgeEmail(emailer, invoice.Email, invoice.Name, invoice.InvoiceID)
		fmt.Printf("Nudge sent to %v, %v, %v, %v\n", invoice.InvoiceID, invoice.Name, invoice.Email, invoice.Amount)
		time.Sleep(1 * time.Second)
	}
}
