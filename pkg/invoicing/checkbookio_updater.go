package invoicing

import (
	"time"

	log "github.com/golang/glog"
	"github.com/joincivil/go-common/pkg/email"
)

const (
	callDelayPerInvoiceMillis = 250
)

// NewCheckoutIOUpdater is a convenience func to create a new CheckoutIOUpdater
func NewCheckoutIOUpdater(client *CheckbookIO, persister *PostgresPersister,
	emailer *email.Emailer, runEverySecs time.Duration) *CheckoutIOUpdater {
	return &CheckoutIOUpdater{
		checkbookIOClient: client,
		invoicePersister:  persister,
		emailer:           emailer,
		runEverySecs:      runEverySecs,
		cancelChan:        make(chan bool),
	}
}

// CheckoutIOUpdater is a struct that contains all the code to run a process
// to maintain and update the state of invoices/checks.
type CheckoutIOUpdater struct {
	checkbookIOClient *CheckbookIO
	invoicePersister  *PostgresPersister
	emailer           *email.Emailer
	runEverySecs      time.Duration
	cancelChan        chan bool
}

// Run starts the background goroutine to run the updater
func (c *CheckoutIOUpdater) Run() {
	log.Infof("Running check and update")
	err := c.checkAndUpdate()
	log.Infof("Check and update complete")
	if err != nil {
		log.Error(err.Error())
	}

Loop:
	for {
		select {
		case <-time.After(c.runEverySecs * time.Second):
			log.Infof("Running check and update")
			err := c.checkAndUpdate()
			log.Infof("Check and update complete")
			if err != nil {
				log.Error(err.Error())
			}
		case <-c.cancelChan:
			break Loop
		}
	}
}

func (c *CheckoutIOUpdater) checkAndUpdate() error {

	allInvoices := []*PostgresInvoice{}

	// Check all UNPAID invoices
	// Once they go to another state, should have a check id, so the webhook will work.
	invoices, err := c.invoicePersister.Invoices("", "", InvoiceStatusUnpaid, "")
	if err != nil {
		log.Errorf("Error retrieving invoices from store: err: %v", err)
		return err
	}
	allInvoices = append(allInvoices, invoices...)

	// Check all IN_PROCESS invoices
	// XXX(PN): Added this bc the webhook from checkbook.io doesn't seem to be working
	// so putting this in for now.
	invoices, err = c.invoicePersister.Invoices("", "", InvoiceStatusInProcess, "")
	if err != nil {
		log.Errorf("Error retrieving invoices from store: err: %v", err)
		return err
	}
	allInvoices = append(allInvoices, invoices...)

	log.Infof("Checking %v invoices", len(allInvoices))
	c.updateInvoices(allInvoices)

	return nil
}

func (c *CheckoutIOUpdater) updateInvoices(invoices []*PostgresInvoice) {
	for _, invoice := range invoices {
		if invoice.StopPoll {
			continue
		}
		if invoice.InvoiceID == "" {
			continue
		}

		checkbookInvoice, err := c.checkbookIOClient.GetInvoice(invoice.InvoiceID)
		if err != nil {
			log.Errorf("Error retrieving invoice %v: err: %v", invoice.InvoiceID, err)
			continue
		}
		if invoice.InvoiceID != checkbookInvoice.ID {
			log.Errorf("Invoice IDs don't match up: %v != %v", invoice.InvoiceID, checkbookInvoice.ID)
			continue
		}
		if invoice.InvoiceStatus == checkbookInvoice.Status {
			continue
		}

		updatedFields := []string{}

		nowPaid := false
		if invoice.InvoiceStatus == InvoiceStatusUnpaid ||
			invoice.InvoiceStatus == InvoiceStatusInProcess {
			if checkbookInvoice.Status == InvoiceStatusPaid {
				nowPaid = true
				invoice.CheckStatus = CheckStatusPaid
				updatedFields = append(updatedFields, "CheckStatus")
			}
		}

		invoice.InvoiceStatus = checkbookInvoice.Status
		updatedFields = append(updatedFields, "InvoiceStatus")

		// Update an empty CheckID
		if checkbookInvoice.CheckID != "" && invoice.CheckID == "" {
			invoice.CheckID = checkbookInvoice.CheckID
			updatedFields = append(updatedFields, "CheckID")
		}

		if nowPaid {
			log.Info("Post payment email would be pushed")
			// TODO(PN): Commenting out for now
			// SendPostPaymentEmail(c.emailer, invoice.Email, invoice.Name)
			// log.Infof("Post payment email sent to %v", invoice.Email)

			// invoice.EmailState = EmailStateSentNextSteps
			// updatedFields = append(updatedFields, "EmailState")
		}

		err = c.invoicePersister.UpdateInvoice(invoice, updatedFields)
		if err != nil {
			log.Errorf("Error updated invoice %v to new status %v",
				invoice.InvoiceID, checkbookInvoice.Status)
		}

		log.Infof("Updated invoice %v, %v to status %v", invoice.InvoiceID,
			invoice.Email, checkbookInvoice.Status)

		// Sleep hack so we don't pound the checkbook API
		time.Sleep(callDelayPerInvoiceMillis * time.Millisecond)
	}
}
