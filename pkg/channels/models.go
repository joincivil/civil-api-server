package channels

import (
	"time"

	uuid "github.com/satori/go.uuid"
)

// ROLES
const (
	// RoleAdmin has the string for the "admin" role
	RoleAdmin = "admin"
)

// TYPES
const (
	TypeUser     = "user"
	TypeNewsroom = "newsroom"
	TypeGroup    = "group"
)

// CreateChannelInput contains the fields needed to create a channel
type CreateChannelInput struct {
	CreatorUserID string
	ChannelType   string
	Reference     string
	Handle        *string
}

// SetHandleInput comment
type SetHandleInput struct {
	ChannelID string
	Handle    string
}

// Channel is container for Posts
type Channel struct {
	ID              string `gorm:"type:uuid;primary_key"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time
	ChannelType     string  `gorm:"not_null;unique_index:channel_idx_type_reference"` // user, newsroom, group
	Reference       string  `gorm:"not_null;unique_index:channel_idx_type_reference"` // user_id, newsroom smart contract address, group DID
	Handle          *string `gorm:"unique_index:channel_idx_handle"`                  // globally unique identifier for channels
	Members         []ChannelMember
	StripeAccountID string
}

// BeforeCreate is a GORM hook that sets the ID before it its persisted
func (c *Channel) BeforeCreate() (err error) {
	id, err := uuid.NewV4()
	if err != nil {
		return err
	}
	c.ID = id.String()
	return
}

// ChannelMember defines the permissions a User has within a Channel
type ChannelMember struct {
	ID        string `gorm:"type:uuid;primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
	ChannelID string `gorm:"type:uuid;not null;index:idx_chanmember_channel_id;unique_index:idx_channel_user"`
	UserID    string `gorm:"type:uuid;not null;index:idx_chanmember_user_id;unique_index:idx_channel_user"`
	Role      string `gorm:"not null"`
	Channel   *Channel
}

// BeforeCreate is a GORM hook that sets the ID before it its persisted
func (c *ChannelMember) BeforeCreate() (err error) {
	id, err := uuid.NewV4()
	if err != nil {
		return err
	}
	c.ID = id.String()
	return
}
