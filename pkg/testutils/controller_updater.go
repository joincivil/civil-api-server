package testutils

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/joincivil/civil-api-server/pkg/tokencontroller"
)

// ControllerUpdaterSpy is used as a mock for users.TokenControllerUpdater
type ControllerUpdaterSpy struct {
	Calls int
}

// AddToCivilians increments calls and returns a hash
func (c *ControllerUpdaterSpy) AddToCivilians(addr common.Address) (common.Hash, error) {
	c.Calls++

	if addr == common.HexToAddress("0x001") {
		return common.HexToHash("0xf00"), tokencontroller.ErrAlreadyOnList
	}
	return common.HexToHash("0xf00"), nil
}
