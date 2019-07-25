package channels

// Persister defines the methods needed to persister Channels
type Persister interface {
	CreateChannel(input CreateChannelInput) (*Channel, error)
	GetChannel(id string) (*Channel, error)
	GetChannelByReference(channelType string, reference string) (*Channel, error)
	GetChannelByHandle(handle string) (*Channel, error)
	GetUserChannels(userID string) ([]*ChannelMember, error)
	IsChannelAdmin(userID string, channelID string) (bool, error)
	GetChannelMembers(channelID string) ([]*ChannelMember, error)
	SetHandle(userID string, channelID string, handle string, nonNormalizedHandle string) (*Channel, error)
	SetStripeAccountID(userID string, channelID string, stripeAccountID string) (*Channel, error)
}
