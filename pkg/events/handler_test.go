package events_test

import (
	"math/big"
	"testing"
	"encoding/json"

	"github.com/ethereum/go-ethereum/common"
	"github.com/joincivil/civil-api-server/pkg/events"
	"github.com/joincivil/civil-api-server/pkg/testutils"
	"github.com/joincivil/civil-api-server/pkg/users"

	"github.com/joincivil/civil-events-processor/pkg/model"
	"github.com/joincivil/civil-events-processor/pkg/processor"
	ptestutils "github.com/joincivil/civil-events-processor/pkg/testutils"
)

func TestCvlTokenTransferEventHandler(t *testing.T) {
	transfer := model.NewTokenTransfer(&model.TokenTransferParams{
		ToAddress:    common.HexToAddress("0x3e39fa983abcd349d95aed608e798817397cf0d1"),
		FromAddress:  common.HexToAddress("0x3e39fa983abcd349d95aed608e798817397cf0d1"),
		Amount:       big.NewInt(10000),
		TransferDate: 1555379455,
		BlockNumber:  3696772,
		TxHash:       common.HexToHash("0x4fa779b4dbf20f8df5b4e523c49920858234172492dc4fb477aee4f7abd67899"),
		TxIndex:      1,
		BlockHash:    common.HexToHash("0x50ba9a0089abe411e11f605d9876eab9554998e5e2f05a7606224cdff55d22f5"),
		Index:        3,
	})

	transfers := map[string][]*model.TokenTransfer{}
	blockData := transfer.BlockData()
	transfers[blockData.TxHash()] = []*model.TokenTransfer{transfer}

	persister := &ptestutils.TestPersister{
		TokenTransfersTxHash: transfers,
	}

	initUsers := map[string]*users.User{
		"1": {
			UID:        "1",
			Email:      "devtests@bar.com",
			EthAddress: "0x3e39fa983abcd349d95aed608e798817397cf0d1",
		},
	}
	userPersister := &testutils.InMemoryUserPersister{UsersInMemory: initUsers}
	controller := &testutils.ControllerUpdaterSpy{}
	userService := users.NewUserService(userPersister, controller)

	pubsubMsg := &processor.PubSubMessage{
		TxHash: "0x4fa779b4dbf20f8df5b4e523c49920858234172492dc4fb477aee4f7abd67899",
	}

	data, err := json.Marshal(pubsubMsg)
	if err != nil {
		t.Errorf("Problem marshalling json: %v", err)
	}
	handler := events.NewCvlTokenTransferEventHandler(
		persister,
		userService,
		[]common.Address{common.HexToAddress("0x3e39fa983abcd349d95aed608e798817397cf0d1")},
		"6933914", // pete test list
	)
	ran, err := handler.Handle(data)
	if err != nil {
		t.Errorf("Should have not returned an error: err: %v", err)
	}
	if !ran {
		t.Errorf("Should have ran the handler")
	}

	name := handler.Name()
	if name != "Transfer" {
		t.Errorf("Should have returned Transfer as name")
	}

}
