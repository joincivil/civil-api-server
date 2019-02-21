package testutils

import "github.com/ethereum/go-ethereum/common"

// ControllerUpdaterSpy is used as a mock for users.TokenControllerUpdater
type ControllerUpdaterSpy struct {
	Calls int
}

// AddToCivilians increments calls and returns a hash
func (c *ControllerUpdaterSpy) AddToCivilians(addr common.Address) (common.Hash, error) {
	c.Calls++
	return common.HexToHash("0xf00"), nil
}
