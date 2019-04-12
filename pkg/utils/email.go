package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	cemail "github.com/joincivil/go-common/pkg/email"

	chttp "github.com/joincivil/go-common/pkg/http"
)

const (
	// Civil Media Company Newsletter List ID in Mailchimp
	newsletterMailchimpListID = "dfcc3f1333"

	addToSendgridURL = "https://us-central1-civil-media.cloudfunctions.net/addToSendgrid"
)

// AddToCivilCompanyNewsletterList adds an email address to our main email list in Mailchimp.
func AddToCivilCompanyNewsletterList(lists cemail.ListMemberManager, emailAddress string,
	tags []cemail.Tag) error {
	if lists.ServiceName() != cemail.ServiceNameMailchimp {
		return fmt.Errorf(
			"Civil newsletter is served from Mailchimp: service: %v",
			lists.ServiceName(),
		)
	}

	emailLower := strings.ToLower(emailAddress)

	subbed, err := lists.IsSubscribedToList(
		newsletterMailchimpListID,
		emailLower,
	)
	if err != nil {
		return err
	}

	if !subbed {
		err = lists.SubscribeToList(
			newsletterMailchimpListID,
			emailLower,
			&cemail.SubscriptionParams{
				Tags: tags,
			},
		)

		if err != nil {
			return err
		}
	}

	return nil
}

type regAlertsResult struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// AddToRegistryAlertsList adds an email address to the dapp registry emails
// Calls the addToSendgridList Google function
func AddToRegistryAlertsList(emailAddress string, listID string) error {
	helper := chttp.RestHelper{}

	params := &url.Values{}
	params.Add("email", emailAddress)
	params.Add("list_id", listID)

	res, err := helper.SendRequestToURL(addToSendgridURL, http.MethodGet, params, nil)
	if err != nil {
		return err
	}

	result := &regAlertsResult{}
	err = json.Unmarshal(res, result)
	if err != nil {
		return err
	}
	if result.Status != "OK" {
		return errors.New(result.Message)
	}
	return nil
}
