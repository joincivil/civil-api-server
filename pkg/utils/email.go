package utils

import (
	"fmt"
	"strings"

	cemail "github.com/joincivil/go-common/pkg/email"
)

const (
	// Civil Media Company Newsletter List ID in Mailchimp
	newsletterMailchimpListID = "dfcc3f1333"
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
