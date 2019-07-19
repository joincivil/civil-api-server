package channels

import (
	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
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
	tx := p.db.Begin()
	c := &Channel{
		ChannelType: input.ChannelType,
		Reference:   input.Reference,
		Handle:      input.Handle,
	}

	if err := tx.Create(c).Error; err != nil {
		return nil, err
	}

	id, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}
	member := &ChannelMember{
		ID:     id.String(),
		UserID: input.CreatorUserID,
		Role:   RoleAdmin,
	}

	tx.Model(c).Association("Members").Append(member)

	tx.Commit()

	if tx.Error != nil {
		return nil, tx.Error
	}

	return c, nil
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
		Reference:   reference,
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
