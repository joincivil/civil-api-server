package discourse_test

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"

	"github.com/joincivil/civil-api-server/pkg/discourse"

	ctime "github.com/joincivil/go-common/pkg/time"
)

func TestAddressConversions(t *testing.T) {
	addr := "0x49fd8f1d3e6f88a4d08cd4a6e445f848e9475caf"
	normalizedAddr := common.HexToAddress(addr).Hex()
	topicID := int64(1010101)
	ts := ctime.CurrentEpochSecsInInt64()

	ldm := &discourse.ListingMap{
		ListingAddress: addr,
		TopicID:        topicID,
		CreatedTs:      ts,
		UpdatedTs:      ts,
	}

	commonAddr := ldm.ListingAddressAsAddr()
	if commonAddr.Hex() != normalizedAddr {
		t.Errorf("Address is incorrect")
	}

	ldm = &discourse.ListingMap{}
	ldm.AddrToListingAddress(commonAddr)
	if commonAddr.Hex() != normalizedAddr {
		t.Errorf("Address is incorrect")
	}
}
