package tokencontroller

import (
	"errors"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/joincivil/go-common/pkg/eth"
	"github.com/joincivil/go-common/pkg/generated/contract"
)

var (
	// ErrAlreadyOnList is thrown when trying to add an address to a list of which it is already on
	ErrAlreadyOnList = errors.New("address is already on list")
	// ErrNoCivilTokenControllerAddress is thrown when `GRAPHQL_CONTRACT_ADDRESSES` envvar does not contain `CivilTokenController`
	ErrNoCivilTokenControllerAddress = errors.New("no CivilTokenController address provided in configuration")
)

// Service provides a set of helpers to interact with the token controller
type Service struct {
	ethHelper       *eth.Helper
	tokenController *contract.CivilTokenControllerContract
}

// NewService builds a new Service instance
func NewService(civilTokenControllerAddress string, ethHelper *eth.Helper) (*Service, error) {
	tokenControllerAddress := common.HexToAddress(civilTokenControllerAddress)
	if tokenControllerAddress == common.HexToAddress("") {
		return nil, ErrNoCivilTokenControllerAddress
	}
	tokenController, err := contract.NewCivilTokenControllerContract(tokenControllerAddress, ethHelper.Blockchain)
	if err != nil {
		return nil, err
	}

	return &Service{
		ethHelper,
		tokenController,
	}, nil
}

// AddToCivilians adds the provided address to the Civilian Whitelist
func (s *Service) AddToCivilians(addr common.Address) (common.Hash, error) {
	isCivilian, err := s.tokenController.CivilianList(&bind.CallOpts{}, addr)
	if err != nil {
		return common.Hash{}, err
	}
	if isCivilian {
		return common.Hash{}, ErrAlreadyOnList
	}
	tx, err := s.tokenController.AddToCivilians(s.ethHelper.Transact(), addr)
	if err != nil {
		return common.Hash{}, err
	}
	return tx.Hash(), nil
}
