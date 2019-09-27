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

// SetEmailEnum represents the different Civil channels for confirming an email
type SetEmailEnum string

const (
	// SetEmailEnumDefault is the default channel value
	SetEmailEnumDefault SetEmailEnum = "DEFAULT"
	// SetEmailEnumUser is the user channel value
	SetEmailEnumUser SetEmailEnum = "USER"
	// SetEmailEnumNewsroom is the newsroom channel value
	SetEmailEnumNewsroom SetEmailEnum = "NEWSROOM"
	// SetEmailEnumGroup is the group channel value
	SetEmailEnumGroup SetEmailEnum = "GROUP"
)

// SetEmailResponse is sent when a channel successfully confirms their email address
type SetEmailResponse struct {
	ChannelID string
	UserID    string `json:"uid"`
}

// CreateChannelInput contains the fields needed to create a channel
type CreateChannelInput struct {
	CreatorUserID string
	ChannelType   string
	Reference     string
	Handle        *string
}

// SetHandleInput contains the fields needed to set a channel handle
type SetHandleInput struct {
	ChannelID string
	Handle    string
}

// UserSetHandleInput contains the fields needed to set a user's channel handle
type UserSetHandleInput struct {
	UserID string
	Handle string
}

// SetEmailInput contains the fields needed to set a channel's email address
type SetEmailInput struct {
	ChannelID    string
	EmailAddress string
	AddToMailing bool
}

// SetAvatarInput contains the fields needed to set a channels' avatar
type SetAvatarInput struct {
	ChannelID     string
	AvatarDataURL string
}

// Channel is container for Posts
type Channel struct {
	ID                   string `gorm:"type:uuid;primary_key"`
	CreatedAt            time.Time
	UpdatedAt            time.Time
	DeletedAt            *time.Time
	ChannelType          string  `gorm:"not_null;unique_index:channel_idx_type_reference"` // user, newsroom, group
	Reference            string  `gorm:"not_null;unique_index:channel_idx_type_reference"` // user_id, newsroom smart contract address, group DID
	Handle               *string `gorm:"unique_index:channel_idx_handle"`                  // globally unique identifier for channels
	RawHandle            *string
	Members              []ChannelMember
	StripeAccountID      string
	EmailAddress         string
	AvatarDataURL        string
	Tiny100AvatarDataURL string // avatar data url scaled down to width 100
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
