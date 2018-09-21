// +build integration

package invoicing_test

import (
	// "fmt"
	"testing"

	"github.com/joincivil/civil-api-server/pkg/invoicing"
)

const (
	testKey    = "b6e5468631bbcba13ac3ba282300ad8d"
	testSecret = "ffc1ca5d1b75841aeda0126d7dfc62a5"
)

func TestCheckbookIOGetInvoices(t *testing.T) {
	cbio := invoicing.NewCheckbookIO(invoicing.SandboxCheckbookIOBaseURL, testKey, testSecret)
	// cbio := invoicing.NewCheckbookIO(invoicing.ProdCheckbookIOBaseURL, testKey, testSecret)
	invoices, err := cbio.GetInvoices("", 0)
	if err != nil {
		t.Errorf("Should not have failed when getting invoices: err: %v", err)
	}

	index := 0
	for _, invoice := range invoices.Invoices {
		if index >= 3 {
			break
		}
		if invoice.ID != "" && invoice.Number == "" {
			t.Error("Number field should not be empty")
		}
		inv, err := cbio.GetInvoice(invoice.ID)
		if err != nil {
			t.Errorf("Should not have failed when getting invoice: err: %v", err)
		}
		if inv.ID != "" && inv.Number == "" {
			t.Error("Number field should not be empty")
		}
		index++
	}

	_, err = cbio.GetInvoices(invoicing.InvoiceStatusUnpaid, 0)
	if err != nil {
		t.Errorf("Should not have failed when getting unpaid invoices: err: %v", err)
	}

	_, err = cbio.GetInvoices(invoicing.InvoiceStatusInProcess, 0)
	if err != nil {
		t.Errorf("Should not have failed when getting inprocess invoices: err: %v", err)
	}

	_, err = cbio.GetInvoices("BADSTATUS", 0)
	if err == nil {
		t.Errorf("Should have failed when getting invoices: err: %v", err)
	}

	_, err = cbio.GetInvoices("", 3)
	if err != nil {
		t.Errorf("Should not have failed when getting inprocess invoices: err: %v", err)
	}

	_, err = cbio.GetInvoices(invoicing.InvoiceStatusInProcess, 3)
	if err != nil {
		t.Errorf("Should not have failed when getting inprocess invoices: err: %v", err)
	}
}

// func TestCheckbookIORequestInvoice(t *testing.T) {
// 	// cbio := invoicing.NewCheckbookIO(invoicing.SandboxCheckbookIOBaseURL, testKey, testSecret)
// 	cbio := invoicing.NewCheckbookIO(invoicing.ProdCheckbookIOBaseURL, testKey, testSecret)
// 	newInvoice := &invoicing.RequestInvoiceParams{
// 		Recipient:   "peter@civil.co",
// 		Name:        "Peter Ng",
// 		Amount:      1.00,
// 		Description: "Give me CVL",
// 	}
// 	_, err := cbio.RequestInvoice(newInvoice)
// 	if err != nil {
// 		t.Errorf("Should not have failed when requesting a new invoice: err: %v", err)
// 	}
// }
