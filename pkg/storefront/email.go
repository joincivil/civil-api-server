package storefront

import (
	"strings"

	log "github.com/golang/glog"
	"github.com/joincivil/civil-api-server/pkg/users"
	"github.com/joincivil/go-common/pkg/email"
)

// ServiceEmailLists is an interface to handle the management of email addresses to
// email lists
type ServiceEmailLists interface {
	PurchaseCompleteAddToMembersList(user *users.User)
	PurchaseCancelRemoveFromAbandonedList(user *users.User)
}

// NewMailchimpServiceEmailLists is a convenience function that returns
// a new MailchimpServiceEmailLists
func NewMailchimpServiceEmailLists(mailchimpAPI *email.MailchimpAPI) *MailchimpServiceEmailLists {
	return &MailchimpServiceEmailLists{
		mailchimpAPI: mailchimpAPI,
	}
}

// MailchimpServiceEmailLists implements ServiceEmailLists on top of Mailchimp
type MailchimpServiceEmailLists struct {
	mailchimpAPI *email.MailchimpAPI
}

// PurchaseCompleteAddToMembersList adds email to the members list and unsubscribes the user
// from the abandoned list if applicable.
func (s *MailchimpServiceEmailLists) PurchaseCompleteAddToMembersList(user *users.User) {
	if user.Email == "" {
		log.Infof("No email found for user: %v", user.UID)
		return
	}

	// Check to see if user email is on member list
	onMemberList, err := s.mailchimpAPI.IsSubscribedToList(mailchimpMemberListID, user.Email)
	if err != nil {
		log.Errorf("Error checking if is subbed to the members list: err: %v", err)
	}

	if !onMemberList {
		// Add user to the Mailchimp members list for marketing/growth team
		err = s.mailchimpAPI.SubscribeToList(mailchimpMemberListID, user.Email)
		if err != nil {
			if !strings.Contains(err.Error(), mailchimpAlreadySubErrSubstring) {
				log.Errorf("Error subscribing to the members list: err: %v", err)
			}
		}
	}

	// Make sure it is unsubscribed from the abandoned list
	err = s.mailchimpAPI.UnsubscribeFromList(mailchimpAbandonedListID, user.Email, false)
	if err != nil {
		if !strings.Contains(err.Error(), mailchimpNotSubscribedErrSubstring) {
			log.Errorf("Error unsubscribing to the abandoned list: err: %v", err)
		}
	}
}

// PurchaseCancelRemoveFromAbandonedList adds the user to the mailchimp abandoned list
// if the user is not on the member's list
func (s *MailchimpServiceEmailLists) PurchaseCancelRemoveFromAbandonedList(user *users.User) {
	if user.Email == "" {
		log.Infof("No email found for user: %v", user.UID)
		return
	}

	// Check to see if user email is on member list
	onMemberList, err := s.mailchimpAPI.IsSubscribedToList(mailchimpMemberListID, user.Email)
	if err != nil {
		log.Errorf("Error checking if is subbed to the members list: err: %v", err)
	}
	// if it is not on the members list, then sent to abandoned list
	if !onMemberList {
		// Make sure it is subscribed to the abandoned list
		err = s.mailchimpAPI.SubscribeToList(mailchimpAbandonedListID, user.Email)
		if err != nil {
			if !strings.Contains(err.Error(), mailchimpAlreadySubErrSubstring) {
				log.Errorf("Error subscribed to the abandoned list: err: %v", err)
			}
		}
	}
}
