package channels

// Persister defines the methods needed to persister Channels
type Persister interface {
	CreateChannel(input CreateChannelInput) (*Channel, error)
	GetChannel(id string) (*Channel, error)
	GetChannelByReference(channelType string, reference string) (*Channel, error)
	GetChannelByHandle(handle string) (*Channel, error)
}
