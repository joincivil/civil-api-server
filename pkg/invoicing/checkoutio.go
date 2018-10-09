package invoicing

import (
	"bytes"
	"encoding/json"
	"fmt"
	log "github.com/golang/glog"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (

	// ProdCheckbookIOBaseURL is the production checkbook IO URL
	ProdCheckbookIOBaseURL = "https://checkbook.io/v3"

	// SandboxCheckbookIOBaseURL is the sandbox checkbook IO URL
	SandboxCheckbookIOBaseURL = "https://sandbox.checkbook.io/v3"

	// InvoiceStatusUnpaid is the checkbook.io status for unpaid invoices
	InvoiceStatusUnpaid = "unpaid"

	// InvoiceStatusPaid is the checkbook.io status for paid invoices
	InvoiceStatusPaid = "paid"

	// InvoiceStatusInProcess is the checkbook.io status for in process invoices
	InvoiceStatusInProcess = "in_process"

	// InvoiceStatusCanceled is the checkbook.io status for canceled invoices
	InvoiceStatusCanceled = "canceled"

	// InvoiceStatusOverdue is the checkbook.io status for overdue invoices
	InvoiceStatusOverdue = "overdue"

	// CheckStatusUnpaid is the checkbook.io status for unpaid checks
	CheckStatusUnpaid = "unpaid"

	// CheckStatusInProcess is the checkbook.io status for in process checks
	CheckStatusInProcess = "in_process"

	// CheckStatusPaid is the checkbook.io status for paid checks
	CheckStatusPaid = "paid"

	// CheckStatusVoid is the checkbook.io status for void checks
	CheckStatusVoid = "void"

	// CheckStatusFailed is the checkbook.io status for failed checks
	CheckStatusFailed = "failed"
)

// NewCheckbookIO is a convenience function to create a new CheckbookIO struct
func NewCheckbookIO(baseURL string, key string, secret string, test bool) *CheckbookIO {
	return &CheckbookIO{
		key:     key,
		secret:  secret,
		baseURL: baseURL,
		test:    test,
	}
}

// InvoicesResponse is the response from checkbook.io on lists of
// invoices. Includes some pagination information from them.
type InvoicesResponse struct {
	Invoices []*InvoiceResponse `json:"invoices"`
	Page     int                `json:"page"`
	Pages    int                `json:"pages"`
	Total    int                `json:"total"`
}

// InvoiceResponse is the response from Checkbook.io for the RequestInvoice
// query
type InvoiceResponse struct {
	Amount      float64     `json:"amount"`
	Date        string      `json:"date"`
	Description string      `json:"description"`
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Number      string      `json:"-"`
	NumberInf   interface{} `json:"number"`
	Recipient   string      `json:"recipient"`
	Status      string      `json:"status"`
	CheckID     string      `json:"check_id"`
}

// CheckbookIO is a wrapper around the CheckbookIO API
type CheckbookIO struct {
	key     string
	secret  string
	baseURL string
	test    bool
}

func (i *InvoiceResponse) normalizeData() {
	// Converting the NumberInf fields into the Number field as
	// a string value
	switch val := i.NumberInf.(type) {
	case string:
		i.Number = val
	case int:
		i.Number = strconv.Itoa(val)
	case float64:
		i.Number = strconv.FormatFloat(val, 'f', -1, 64)
	}

	// Normalize the status string (lower, replace spaces with _)
	i.Status = strings.ToLower(i.Status)
	i.Status = strings.Replace(i.Status, " ", "_", -1)
}

// GetInvoices returns the list of invoices for the given key/secret account
func (c *CheckbookIO) GetInvoices(status string, pageNum int) (*InvoicesResponse, error) {
	vals := &url.Values{}
	if pageNum > 0 {
		vals.Add("page", fmt.Sprintf("%v", pageNum))
	}
	if status != "" {
		// Request requires uppercase version of the status, but trying to
		// keep everything else lowercase for consistency.
		vals.Add("status", strings.ToUpper(status))
	}

	endpoint := "invoice"
	bys, err := c.sendRequest(endpoint, http.MethodGet, vals, nil)
	if err != nil {
		return nil, err
	}

	invoicesResp := &InvoicesResponse{}
	err = json.Unmarshal(bys, invoicesResp)
	if err != nil {
		return nil, err
	}

	for _, invoice := range invoicesResp.Invoices {
		invoice.normalizeData()
	}

	return invoicesResp, nil
}

// GetInvoice returns the invoice by id
func (c *CheckbookIO) GetInvoice(id string) (*InvoiceResponse, error) {
	endpoint := fmt.Sprintf("invoice/%v", id)
	bys, err := c.sendRequest(endpoint, http.MethodGet, nil, nil)
	if err != nil {
		return nil, err
	}

	invoiceResp := &InvoiceResponse{}
	err = json.Unmarshal(bys, invoiceResp)
	if err != nil {
		return nil, err
	}
	invoiceResp.normalizeData()
	return invoiceResp, nil
}

// RequestInvoiceParams is the data passed to RequestInvoice
type RequestInvoiceParams struct {
	Recipient   string  `json:"recipient"` // The email of the recipient
	Name        string  `json:"name"`
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
	Number      string  `json:"number,omitempty"`
	Account     string  `json:"account,omitempty"`
	Attachment  string  `json:"attachment,omitempty"`
}

// Used for testing purposes so it doesn't hit the checkbook.io services
// and actually deliver an invoice. Helps us get around the "no account" issue
// in the checkbook.io sandbox.
// XXX(PN): Temporary, remove later
func testInvoiceResponse(r *RequestInvoiceParams) *InvoiceResponse {
	return &InvoiceResponse{
		Amount:      r.Amount,
		Date:        "",
		Description: r.Description,
		ID:          "",
		Name:        r.Name,
		Number:      "",
		NumberInf:   "",
		Recipient:   r.Recipient,
		Status:      InvoiceStatusUnpaid,
		CheckID:     "",
	}

}

// RequestInvoice sends an invoice to the recipient as given in the params.  Returns
// data on the new invoice from checkbook.io
func (c *CheckbookIO) RequestInvoice(params *RequestInvoiceParams) (*InvoiceResponse, error) {
	if c.test {
		log.Infof("Returning test request response")
		return testInvoiceResponse(params), nil
	}
	endpoint := "invoice"

	bys, err := c.sendRequest(endpoint, http.MethodPost, nil, params)
	if err != nil {
		return nil, err
	}

	invoiceResp := &InvoiceResponse{}
	err = json.Unmarshal(bys, invoiceResp)
	if err != nil {
		return nil, err
	}
	invoiceResp.normalizeData()
	return invoiceResp, nil
}

func (c *CheckbookIO) sendRequest(endpointName string, method string, params *url.Values,
	payload interface{}) ([]byte, error) {
	client := &http.Client{}
	var req *http.Request
	var err error

	url := fmt.Sprintf("%v/%v", c.baseURL, endpointName)

	if method == http.MethodPost {
		req, err = c.buildPostPutRequest(method, url, payload)
	} else {
		req, err = c.buildGetDeleteRequest(method, url, params)
	}

	if err != nil {
		return nil, err
	}

	// Add the authorization header
	keySecret := fmt.Sprintf("%v:%v", c.key, c.secret)
	req.Header.Add("Authorization", keySecret)

	// Make the request
	rsp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer rsp.Body.Close() // nolint: errcheck
	rspBodyData, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}
	log.Infof("respBody = %v", string(rspBodyData))
	if rsp.StatusCode != 200 && rsp.StatusCode != 201 {
		return nil, fmt.Errorf("Request failed: %v, %v", rsp.StatusCode, string(rspBodyData))
	}

	return rspBodyData, nil
}

func (c *CheckbookIO) buildPostPutRequest(method string, url string,
	payload interface{}) (*http.Request, error) {
	var reqBody *bytes.Buffer
	// If there was a payload struct to marshal into payload string
	if payload != nil {
		payloadData, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBufferString(string(payloadData))
		// log.Infof("reqBody = %v", reqBody.String())
	}

	// Build a new request
	req, err := http.NewRequest(
		method,
		url,
		reqBody,
	)
	if err != nil {
		return nil, err
	}
	// log.Infof("%v", req.URL.String())
	return req, nil
}

func (c *CheckbookIO) buildGetDeleteRequest(method string, url string,
	params *url.Values) (*http.Request, error) {
	req, err := http.NewRequest(
		method,
		url,
		nil,
	)
	if err != nil {
		return nil, err
	}
	if params != nil {
		req.URL.RawQuery = params.Encode()
	}
	// log.Infof("%v", req.URL.String())
	return req, nil
}
