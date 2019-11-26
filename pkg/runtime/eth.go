package runtime

import (
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	log "github.com/golang/glog"
	"github.com/joincivil/civil-api-server/pkg/utils"
	"github.com/joincivil/go-common/pkg/eth"
	"github.com/joincivil/go-common/pkg/generated/contract"
	"github.com/joincivil/go-common/pkg/jobs"
	"go.uber.org/fx"
	"time"
)

// EthRuntime provides the services in the eth package
var EthRuntime = fx.Options(
	fx.Provide(
		NewETHHelper,
		NewTXListener,
		NewCVLTokenContract,
		NewDeployerContractAddresses,
	),
)

// NewETHHelper builds a new eth.Helper instance
func NewETHHelper(config *utils.GraphQLConfig) (*eth.Helper, error) {
	if config.EthAPIURL != "" {
		accounts := map[string]string{}
		if config.EthereumDefaultPrivateKey != "" {
			log.Infof("Initialized default Ethereum account\n")
			accounts["default"] = config.EthereumDefaultPrivateKey
		}
		ethHelper, err := eth.NewETHClientHelper(config.EthAPIURL, accounts)
		if err != nil {
			return nil, err
		}
		log.Infof("Connected to Ethereum using %v\n", config.EthAPIURL)
		return ethHelper, nil
	}

	ethHelper, err := eth.NewSimulatedBackendHelper()
	if err != nil {
		return nil, err
	}
	log.Infof("Connected to Ethereum using Simulated Backend\n")
	return ethHelper, nil
}

// NewTXListener constructs a new eth.TxListener
func NewTXListener(ethHelper *eth.Helper) *eth.TxListener {
	jobsService := jobs.NewInMemoryJobService()
	listener := eth.NewTxListenerWithWaitPeriod(ethHelper.Blockchain.(ethereum.TransactionReader), jobsService, 420*time.Second)

	return listener
}

// NewCVLTokenContract constructs a new CVLTokenContract
func NewCVLTokenContract(addresses eth.DeployerContractAddresses, ethHelper *eth.Helper) *contract.CVLTokenContract {
	tokenContract, err := contract.NewCVLTokenContract(addresses.CVLToken, ethHelper.Blockchain)
	if err != nil {
		panic("could not initialize CVLToken contract")
	}

	return tokenContract
}

// NewDeployerContractAddresses builds a new DeployerContractAddresses instance
func NewDeployerContractAddresses(config *utils.GraphQLConfig) eth.DeployerContractAddresses {

	return eth.DeployerContractAddresses{
		NewsroomFactory:       extractContractAddress(config, "NewsroomFactory"),
		MultisigFactory:       extractContractAddress(config, "MultisigFactory"),
		CivilTokenController:  extractContractAddress(config, "CivilTokenController"),
		CreateNewsroomInGroup: extractContractAddress(config, "CreateNewsroomInGroup"),
		PLCR:                  extractContractAddress(config, "PLCR"),
		TCR:                   extractContractAddress(config, "CivilTCR"),
		CVLToken:              extractContractAddress(config, "CVLToken"),
		Parameterizer:         extractContractAddress(config, "Parameterizer"),
		Government:            extractContractAddress(config, "Government"),
	}
}
func extractContractAddress(config *utils.GraphQLConfig, contractName string) common.Address {
	return common.HexToAddress(config.ContractAddresses[contractName])
}
