package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/joincivil/civil-api-server/pkg/newsrooms"
	cconfig "github.com/joincivil/go-common/pkg/config"
	"github.com/joincivil/go-common/pkg/eth"
	"github.com/joincivil/go-common/pkg/generated/contract"
	"github.com/joincivil/go-common/pkg/jobs"
	"github.com/joincivil/go-common/pkg/newsroom"
	"github.com/kelseyhightower/envconfig"
	"go.uber.org/fx"
)

// var contractInitializer func() eth.DeployerContractAddresses
var foundationMultisig common.Address

func main() {

	app := fx.New(
		fx.Provide(
			loadConfig,
			initContractAddresses,
			initTXListener,
			initETHHelper,
			initCVLTokenContract,
			newsroom.NewService,
			newsrooms.NewTools,
		),
		fx.Invoke(runCLI),
	)

	if err := app.Start(context.Background()); err != nil {
		fmt.Printf("error: %v", err)
	}

}

func runCLI(tools *newsrooms.Tools, cfg *config) {

	args := os.Args[1:]

	switch args[0] {
	case "create":
		charter := loadCharter(args[1])

		fmt.Printf(
			"Starting newsroom creation:\n"+
				"\tEthereum network: %v -> %v\n"+
				"\tNewsroom name: %v\n"+
				"\tTokens to apply: %v\n",
			cfg.EthEnv,
			cfg.EthAPIURL,
			charter.Name,
			cfg.ApplicationTokens,
		)

		applicationTokens := big.NewInt(cfg.ApplicationTokens)
		applicationTokens = applicationTokens.Mul(applicationTokens, big.NewInt(1e18))
		updates := make(chan string)

		go func() {
			err := tools.CreateAndApply(updates, charter, applicationTokens)
			if err != nil {
				fmt.Printf("error with CreateAndApply: %v\n", err)
			}
		}()

		for update := range updates {
			fmt.Printf("received update: %v\n", update)
		}
	case "handoff":
		updates := make(chan string)
		newsroomAddress := common.HexToAddress(args[1])
		newAddress := common.HexToAddress(args[2])
		foundationRecovery := true

		fmt.Printf(
			"Starting newsroom handoff:\n"+
				"\tEthereum network: %v %v\n"+
				"\tNewsroom Address: %v\n"+
				"\tNew Owner: %v\n"+
				"\tAdding Foundation Multisig for Recovery:%v\n",
			cfg.EthEnv,
			cfg.EthAPIURL,
			newsroomAddress.String(),
			newAddress.String(),
			foundationRecovery,
		)

		newOwners := []common.Address{newAddress}
		if foundationRecovery {
			newOwners = append(newOwners, foundationMultisig)
		}

		go func() {
			err := tools.HandoffNewsroom(updates, newsroomAddress, newOwners)
			if err != nil {
				fmt.Printf("error with HandoffNewsroom: %v\n", err)
			}
		}()

		for update := range updates {
			fmt.Printf("received update: %v\n", update)
		}
	default:
		fmt.Println("unknown command, try either `create` or `handoff`")
	}

	os.Exit(0)
}

func loadCharter(filename string) *newsroom.Charter {
	jsonFile, err := os.Open(filename)
	if err != nil {
		fmt.Printf("eror opening file: %v", err)
		os.Exit(0)
	}
	fmt.Printf("Successfully Opened %v\n", filename)
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)

	input := &newsroom.Charter{}
	err = json.Unmarshal(byteValue, input)

	if err != nil {
		fmt.Printf("Error parsing JSON input: %v\n", err)
		os.Exit(0)
	}

	return input
}

type config struct {
	PrivateKey        string `required:"true" split_words:"true"  desc:"ethereum private key"`
	EthAPIURL         string `required:"true" envconfig:"eth_api_url" desc:"Ethereum API address"`
	EthEnv            string `required:"true" envconfig:"eth_env" desc:"rinkeby or mainnet"`
	ApplicationTokens int64  `required:"true" split_words:"true" desc:"number of tokens required to apply"`
}

func loadConfig() (*config, error) {
	flag.Usage = func() {
		os.Exit(0)
	}
	flag.Parse()
	err := cconfig.PopulateFromDotEnv("newsroom")
	if err != nil {
		return nil, err
	}

	c := &config{}
	err = envconfig.Process("newsroom", c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func initETHHelper(cfg *config) (*eth.Helper, error) {
	accounts := map[string]string{
		"default": cfg.PrivateKey,
	}
	return eth.NewETHClientHelper(cfg.EthAPIURL, accounts)
}

func initCVLTokenContract(addresses eth.DeployerContractAddresses, ethHelper *eth.Helper) *contract.CVLTokenContract {
	tokenContract, err := contract.NewCVLTokenContract(addresses.CVLToken, ethHelper.Blockchain)
	if err != nil {
		panic("could not initialize CVLToken contract")
	}

	return tokenContract
}

func initTXListener(ethHelper *eth.Helper) *eth.TxListener {
	jobsService := jobs.NewInMemoryJobService()
	listener := eth.NewTxListenerWithWaitPeriod(ethHelper.Blockchain.(ethereum.TransactionReader), jobsService, 420*time.Second)

	return listener
}

func initContractAddresses(cfg *config) eth.DeployerContractAddresses {
	// TODO(dankins): I am taking some shortcuts here, these should probably be configuration driven
	// hardcoding here for expediency
	if cfg.EthEnv == "mainnet" {
		// mainnet
		fmt.Println("mainnet")
		return eth.DeployerContractAddresses{
			NewsroomFactory:       common.HexToAddress("0x3e39fa983abcd349d95aed608e798817397cf0d1"),
			MultisigFactory:       common.HexToAddress("0x15f91c1936d854e74d6793efffe9f0b1a81098c5"),
			CivilTokenController:  common.HexToAddress("0x6d3dc15e04dd1d8968556d02cf209a4fb4ab8736"),
			CreateNewsroomInGroup: common.HexToAddress("0x870188243e72210d2b9a276be411fd4f32c8eb40"),
			PLCR:                  common.HexToAddress("0x55656b8a58df94c1e8b5142f8da973301452ea65"),
			TCR:                   common.HexToAddress("0xbd5a95a66dd4e78bcb597198df222c4eddc14da7"),
			CVLToken:              common.HexToAddress("0x01FA555c97D7958Fa6f771f3BbD5CCD508f81e22"),
			Parameterizer:         common.HexToAddress("0x0b8170f7cec8564492ffea951be88b915a4e26d2"),
			Government:            common.HexToAddress("0xc625fc42ab6d07746b953ae98b4ec22622e1b9a9"),
		}
	}

	// rinkeby
	fmt.Println("rinkeby")
	return eth.DeployerContractAddresses{
		NewsroomFactory:       common.HexToAddress("0x0437bb69da983a508ce77b5f16dc86a96163318b"),
		MultisigFactory:       common.HexToAddress("0x8c0938fe37e91a143fe25812c116241c9e49e14c"),
		CivilTokenController:  common.HexToAddress("0x6c343cfded474f9800e7b49287b210b608c2ea9b"),
		CreateNewsroomInGroup: common.HexToAddress("0x49d731e07c37f2e6eb9022a14b4198dbccc4452a"),
		PLCR:                  common.HexToAddress("0x570bf68d286dd4225e2e77384c08dfebc4b01b5c"),
		TCR:                   common.HexToAddress("0xdad6d7ea1e43f8492a78bab8bb0d45a889ed6ac3"),
		CVLToken:              common.HexToAddress("0x3e39fa983abcd349d95aed608e798817397cf0d1"),
		Parameterizer:         common.HexToAddress("0x4f4b97a4faebf2bd835a10c479e61faccef8755d"),
		Government:            common.HexToAddress("0x4dc4168bfbe5b8bb2a4ccad35fe15e2417c022e8"),
	}
}
