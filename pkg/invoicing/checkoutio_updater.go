package invoicing

import (
	log "github.com/golang/glog"
	"time"
)

// NewCheckoutIOUpdater is a convenience func to create a new CheckoutIOUpdater
func NewCheckoutIOUpdater(client *CheckbookIO, persister *PostgresPersister,
	runEverySecs time.Duration) *CheckoutIOUpdater {
	return &CheckoutIOUpdater{
		checkbookIOClient: client,
		invoicePersister:  persister,
		runEverySecs:      runEverySecs,
		cancelChan:        make(chan bool),
	}
}

// CheckoutIOUpdater is a struct that contains all the code to run a process
// to maintain and update the state of invoices/checks.
type CheckoutIOUpdater struct {
	checkbookIOClient *CheckbookIO
	invoicePersister  *PostgresPersister
	runEverySecs      time.Duration
	cancelChan        chan bool
}

// Run starts the background goroutine to run the updater
func (c *CheckoutIOUpdater) Run() {
	log.Infof("Running check and update")
	err := c.checkAndUpdate()
	if err != nil {
		log.Error(err.Error())
	}
	log.Infof("Check and update complete")

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
	// Check all UNPAID invoices
	// Once they go to another state, should have a check id, so the webhook will work.
	invoices, err := c.invoicePersister.Invoices("", "", InvoiceStatusUnpaid, "")
	if err != nil {
		log.Errorf("Error retrieving invoices from store: err: %v", err)
		return err
	}

	for _, invoice := range invoices {
		if invoice.StopPoll {
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

		nowPaid := false
		if invoice.InvoiceStatus == InvoiceStatusUnpaid ||
			invoice.InvoiceStatus == InvoiceStatusInProcess {
			if checkbookInvoice.Status == InvoiceStatusPaid {
				nowPaid = true
			}
		}

		invoice.InvoiceStatus = checkbookInvoice.Status
		updatedFields := []string{"InvoiceStatus"}

		// Update an empty CheckID
		if checkbookInvoice.CheckID != "" && invoice.CheckID == "" {
			invoice.CheckID = checkbookInvoice.CheckID
			updatedFields = append(updatedFields, "CheckID")
		}

		err = c.invoicePersister.UpdateInvoice(invoice, updatedFields)
		if err != nil {
			log.Errorf("Error updated invoice %v to new status %v", checkbookInvoice.Status)
		}

		if nowPaid {
			// Push a message to pubsub
			log.Infof("Invoice was just paid, so push message to pubsub")
		}

		log.Infof("Updated invoice %v, %v to status %v", invoice.InvoiceID, invoice.Email, checkbookInvoice.Status)
	}

	return nil
}
