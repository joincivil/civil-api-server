package testutils

import "github.com/joincivil/civil-api-server/pkg/channels"

// MockChannelHelper is used as a mock for users.UserChannelHelper
type MockChannelHelper struct {
}

// CreateUserChannel mock implementation of CreateUserChannel. no-op.
func (m *MockChannelHelper) CreateUserChannel(userID string) (*channels.Channel, error) {
	return nil, nil
}
