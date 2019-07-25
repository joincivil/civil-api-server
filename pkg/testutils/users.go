package testutils

import "github.com/joincivil/civil-api-server/pkg/channels"

// MockChannelHelper COMMENT!!!
type MockChannelHelper struct {
}

// CreateUserChannel comment
func (m *MockChannelHelper) CreateUserChannel(userID string) (*channels.Channel, error) {
	return nil, nil
}
