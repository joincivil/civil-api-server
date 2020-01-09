package channels

import (
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"strings"
)

// DBPersister implements the Persister interface using GORM
type DBPersister struct {
	db *gorm.DB
}

// NewDBPersister instantiates a new DBPersister
func NewDBPersister(db *gorm.DB) *DBPersister {
	return &DBPersister{
		db,
	}
}

// CreateChannel saves a new Channel to the database
func (p *DBPersister) CreateChannel(input CreateChannelInput) (*Channel, error) {
	var normalizedHandle *string
	var err error
	if input.Handle != nil {
		normalized, err := NormalizeHandle(*(input.Handle))
		if err != nil {
			return nil, err
		}
		normalizedHandle = &normalized

		// make sure there is not a channel with this handle
		ch, err := p.GetChannelByHandle(normalized)
		if err != nil && err != ErrorNotFound {
			return nil, err
		}
		if ch != nil {
			return nil, ErrorNotUnique
		}

	}

	// make sure there is not a channel with this reference
	ch, err := p.GetChannelByReference(input.ChannelType, input.Reference)
	if err != nil && err != ErrorNotFound {
		return nil, err
	}
	if ch != nil {
		return nil, ErrorNotUnique
	}

	tx := p.db.Begin()

	c := &Channel{
		ChannelType: input.ChannelType,
		Reference:   input.Reference,
		Handle:      normalizedHandle,
		RawHandle:   input.Handle,
	}

	if err := tx.Create(c).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	_, err = p.createChannelMemberWithTx(input.CreatorUserID, c, tx)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	tx.Commit()

	return c, nil
}

// CreateChannelMember creates a channel member for the given channel and user id
func (p *DBPersister) CreateChannelMember(channel *Channel, userID string) (*ChannelMember, error) {
	tx := p.db.Begin()
	member, err := p.createChannelMemberWithTx(userID, channel, tx)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()
	return member, nil
}

// DeleteChannelMember deletes the channel member for the given channel and user id
// this uses the `Unscoped` Delete function to remove entry from DB, rather than just
// setting deleted_at
func (p *DBPersister) DeleteChannelMember(channel *Channel, userID string) error {
	member, err := p.GetChannelMember(channel.ID, userID)
	if err != nil {
		return err
	}

	err = p.db.Unscoped().Delete(member).Error

	if err != nil {
		return err
	}

	return nil
}

func (p *DBPersister) createChannelMemberWithTx(userID string, c *Channel, tx *gorm.DB) (*ChannelMember, error) {
	id := uuid.NewV4()
	member := &ChannelMember{
		ID:     id.String(),
		UserID: userID,
		Role:   RoleAdmin,
	}

	tx.Model(c).Association("Members").Append(member)

	if tx.Error != nil {
		return nil, tx.Error
	}

	return member, nil
}

// GetChannel retrieves a Channel with the provided ID
func (p *DBPersister) GetChannel(id string) (*Channel, error) {
	c := &Channel{}

	if p.db.Where(&Channel{ID: id}).Preload("Members").First(c).RecordNotFound() {
		return nil, ErrorNotFound
	}

	return c, nil
}

// GetChannelByReference retrieves a Channel by type+reference
func (p *DBPersister) GetChannelByReference(channelType string, reference string) (*Channel, error) {
	c := &Channel{}

	if p.db.Where(&Channel{
		ChannelType: channelType,
		Reference:   strings.ToLower(reference),
	}).First(c).RecordNotFound() {
		return nil, ErrorNotFound
	}

	return c, nil
}

// GetChannelByHandle retrieves a Channel with the provided handle
func (p *DBPersister) GetChannelByHandle(handle string) (*Channel, error) {
	c := &Channel{}

	if p.db.Where(&Channel{
		Handle: &handle,
	}).First(c).RecordNotFound() {
		return nil, ErrorNotFound
	}

	return c, nil
}

// GetChannelMembers retrieves all the members of a channel given an id
func (p *DBPersister) GetChannelMembers(channelID string) ([]*ChannelMember, error) {
	var c []*ChannelMember

	if err := p.db.Where(&ChannelMember{
		ChannelID: channelID,
	}).Preload("Channel").Find(&c).Error; err != nil {
		return nil, ErrorNotFound
	}

	return c, nil
}

// GetChannelMember retrieves the channel member for a channel id and user id
func (p *DBPersister) GetChannelMember(channelID string, userID string) (*ChannelMember, error) {
	c := &ChannelMember{}

	if err := p.db.Where(&ChannelMember{
		ChannelID: channelID,
		UserID:    userID,
	}).Preload("Channel").First(c).Error; err != nil {
		return nil, ErrorNotFound
	}

	return c, nil
}

// GetUserChannels returns the channel a user belongs to
func (p *DBPersister) GetUserChannels(userID string) ([]*ChannelMember, error) {
	var c []*ChannelMember

	if err := p.db.Where(&ChannelMember{
		UserID: userID,
	}).Preload("Channel").Find(&c).Error; err != nil {
		return nil, ErrorNotFound
	}

	return c, nil
}

// SetHandle updates the handle for the channel, ensuring that it is unique
func (p *DBPersister) SetHandle(userID string, channelID string, handle string) (*Channel, error) {
	// get channel
	ch, err := p.GetChannel(channelID)
	if err != nil {
		return nil, errors.Wrap(err, "error setting handle, could not get channel")
	}

	// make sure the user requesting is an admin
	err = p.requireAdmin(userID, channelID)
	if err == ErrorUnauthorized {
		return nil, ErrorUnauthorized
	} else if err != nil {
		return nil, errors.Wrap(err, "error setting handle, not an admin")
	}

	return p.setHandle(ch, handle)
}

// SetNewsroomHandleOnAccepted updates the handle for the newsroom channel, ensuring that it is unique
func (p *DBPersister) SetNewsroomHandleOnAccepted(channelID string, handle string) (*Channel, error) {
	// get channel
	ch, err := p.GetChannel(channelID)
	if err != nil {
		return nil, errors.Wrap(err, "error setting handle, could not get channel")
	}

	return p.setHandle(ch, handle)
}

func (p *DBPersister) setHandle(ch *Channel, handle string) (*Channel, error) {

	normalizedHandle, err := NormalizeHandle(handle)
	if err != nil {
		return nil, err
	}

	// make sure there is not a channel with this handle
	ch2, err := p.GetChannelByHandle(normalizedHandle)
	if err != nil && err != ErrorNotFound {
		return nil, err
	}
	if ch2 != nil {
		return nil, ErrorNotUnique
	}

	err = p.db.Model(ch).Update(Channel{Handle: &normalizedHandle, RawHandle: &handle}).Error
	if err != nil {
		return nil, errors.Wrap(err, "error setting handle")
	}

	return ch, nil
}

// ClearNewsroomHandleOnRemoved clears a newsroom's handle
func (p *DBPersister) ClearNewsroomHandleOnRemoved(channelID string) (*Channel, error) {
	// get channel
	ch, err := p.GetChannel(channelID)
	if err != nil {
		return nil, errors.Wrap(err, "error setting handle, could not get channel")
	}

	emptyHandle := ""
	err = p.db.Model(ch).Update(Channel{Handle: &emptyHandle, RawHandle: &emptyHandle}).Error
	if err != nil {
		return nil, errors.Wrap(err, "error setting handle")
	}

	return ch, nil
}

// SetEmailAddress updates the email address for the channel
func (p *DBPersister) SetEmailAddress(userID string, channelID string, emailAddress string) (*Channel, error) {
	// get channel
	ch, err := p.GetChannel(channelID)
	if err != nil {
		return nil, errors.Wrap(err, "error setting email, could not get channel")
	}

	// make sure the user requesting is an admin
	err = p.requireAdmin(userID, channelID)
	if err == ErrorUnauthorized {
		return nil, ErrorUnauthorized
	} else if err != nil {
		return nil, errors.Wrap(err, "error setting email, not an admin")
	}

	err = p.db.Model(ch).Update(Channel{EmailAddress: emailAddress}).Error
	if err != nil {
		return nil, errors.Wrap(err, "error setting email")
	}

	return ch, nil
}

// SetAvatarDataURL updates the avatar data url for the channel
func (p *DBPersister) SetAvatarDataURL(userID string, channelID string, avatarDataURL string) (*Channel, error) {
	// get channel
	ch, err := p.GetChannel(channelID)
	if err != nil {
		return nil, errors.Wrap(err, "error setting avatar data url, could not get channel")
	}

	// make sure the user requesting is an admin
	err = p.requireAdmin(userID, channelID)
	if err == ErrorUnauthorized {
		return nil, ErrorUnauthorized
	} else if err != nil {
		return nil, errors.Wrap(err, "error setting avatar data url, not an admin")
	}

	err = p.db.Model(ch).Update(Channel{AvatarDataURL: avatarDataURL}).Error
	if err != nil {
		return nil, errors.Wrap(err, "error setting email")
	}

	return ch, nil
}

// SetTiny72AvatarDataURL updates the tiny72 avatar data url for the channel
func (p *DBPersister) SetTiny72AvatarDataURL(userID string, channelID string, avatarDataURL string) error {
	// get channel
	ch, err := p.GetChannel(channelID)
	if err != nil {
		return errors.Wrap(err, "error setting avatar data url, could not get channel")
	}

	// make sure the user requesting is an admin
	err = p.requireAdmin(userID, channelID)
	if err == ErrorUnauthorized {
		return ErrorUnauthorized
	} else if err != nil {
		return errors.Wrap(err, "error setting avatar data url, not an admin")
	}

	err = p.db.Model(ch).Update(Channel{Tiny72AvatarDataURL: avatarDataURL}).Error
	if err != nil {
		return errors.Wrap(err, "error setting email")
	}

	return nil
}

// SetStripeCustomerID updates the stripe customer id for the channel
func (p *DBPersister) SetStripeCustomerID(userID string, channelID string, stripeCustomerID string) (*Channel, error) {
	// get channel
	ch, err := p.GetChannel(channelID)
	if err != nil {
		return nil, errors.Wrap(err, "error setting stripe customer id")
	}

	// make sure the user requesting is an admin
	err = p.requireAdmin(userID, channelID)
	if err == ErrorUnauthorized {
		return nil, ErrorUnauthorized
	} else if err != nil {
		return nil, errors.Wrap(err, "error setting stripe customer id")
	}

	// update the stripe account id
	err = p.db.Model(ch).Update(Channel{StripeCustomerID: stripeCustomerID}).Error
	if err != nil {
		return nil, errors.Wrap(err, "error setting stripe customer id")
	}

	return ch, nil
}

// ClearStripeCustomerID clears the stripe customer id for the channel
func (p *DBPersister) ClearStripeCustomerID(userID string, channelID string) (*Channel, error) {
	// get channel
	ch, err := p.GetChannel(channelID)
	if err != nil {
		return nil, errors.Wrap(err, "error setting stripe customer id")
	}

	// make sure the user requesting is an admin
	err = p.requireAdmin(userID, channelID)
	if err == ErrorUnauthorized {
		return nil, ErrorUnauthorized
	} else if err != nil {
		return nil, errors.Wrap(err, "error setting stripe customer id")
	}

	// clears the stripe account id
	err = p.db.Model(ch).Update(Channel{StripeCustomerID: ""}).Error
	if err != nil {
		return nil, errors.Wrap(err, "error setting stripe customer id")
	}

	return ch, nil
}

// SetStripeAccountID updates the stripe account id for the channel
func (p *DBPersister) SetStripeAccountID(userID string, channelID string, stripeAccountID string) (*Channel, error) {
	// get channel
	ch, err := p.GetChannel(channelID)
	if err != nil {
		return nil, errors.Wrap(err, "error setting stripe id")
	}

	// make sure the user requesting is an admin
	err = p.requireAdmin(userID, channelID)
	if err == ErrorUnauthorized {
		return nil, ErrorUnauthorized
	} else if err != nil {
		return nil, errors.Wrap(err, "error setting stripe id")
	}

	// update the stripe account id
	err = p.db.Model(ch).Update(Channel{StripeAccountID: stripeAccountID}).Error
	if err != nil {
		return nil, errors.Wrap(err, "error setting stripe account id")
	}

	return ch, nil
}

// IsChannelAdmin returns whether the userID
func (p *DBPersister) IsChannelAdmin(userID string, channelID string) (bool, error) {

	var c = &ChannelMember{}
	err := p.db.Where(&ChannelMember{
		ChannelID: channelID,
		UserID:    userID,
	}).First(c).Error
	if gorm.IsRecordNotFoundError(err) {
		return false, nil
	} else if err != nil {
		return false, ErrorNotFound
	}

	return true, nil
}

func (p *DBPersister) requireAdmin(userID string, channelID string) error {
	// make sure the user requesting is an admin
	isAdmin, err := p.IsChannelAdmin(userID, channelID)
	if err != nil {
		return err
	}
	if !isAdmin {
		return ErrorUnauthorized
	}

	return nil
}
