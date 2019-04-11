package graphqlmain

import (
	"errors"

	log "github.com/golang/glog"

	"github.com/joincivil/civil-api-server/pkg/auth"
	"github.com/joincivil/civil-api-server/pkg/invoicing"
	"github.com/joincivil/civil-api-server/pkg/jsonstore"
	"github.com/joincivil/civil-api-server/pkg/kyc"
	"github.com/joincivil/civil-api-server/pkg/nrsignup"
	"github.com/joincivil/civil-api-server/pkg/storefront"
	"github.com/joincivil/civil-api-server/pkg/tokencontroller"
	"github.com/joincivil/civil-api-server/pkg/tokenfoundry"
	"github.com/joincivil/civil-api-server/pkg/users"
	"github.com/joincivil/civil-api-server/pkg/utils"

	"github.com/joincivil/go-common/pkg/eth"

	cemail "github.com/joincivil/go-common/pkg/email"
)

type dependencies struct {
	emailer                *cemail.Emailer
	mailchimp              *cemail.MailchimpAPI
	jwtGenerator           *auth.JwtTokenGenerator
	invoicePersister       *invoicing.PostgresPersister
	checkbookIO            *invoicing.CheckbookIO
	userService            *users.UserService
	authService            *auth.Service
	jsonbService           *jsonstore.Service
	nrsignupService        *nrsignup.Service
	tokenFoundry           *tokenfoundry.API
	onfido                 *kyc.OnfidoAPI
	ethHelper              *eth.Helper
	storefrontService      *storefront.Service
	tokenControllerService *tokencontroller.Service
}

func initDependencies(config *utils.GraphQLConfig) (*dependencies, error) {
	var err error

	ethHelper, err := initETHHelper(config)
	if err != nil {
		log.Fatalf("Error w init ETH Helper: err: %v", err)
		return nil, err
	}

	tokenControllerService, err := initTokenControllerService(config, ethHelper)
	if err != nil {
		log.Fatalf("Error w init tokenControllerService Service: err: %v", err)
		return nil, err
	}

	var checkbookIO *invoicing.CheckbookIO
	if config.EnableInvoicing {
		checkbookIO, err = invoiceCheckbookIO(config)
		if err != nil {
			log.Fatalf("Error setting up invoicing client: err: %v", err)
		}
	}

	invoicePersister, err := initInvoicePersister(config)
	if err != nil {
		log.Fatalf("Error setting up invoicing persister: err: %v", err)
		return nil, err
	}

	jwtGenerator := auth.NewJwtTokenGenerator([]byte(config.JwtSecret))
	tokenFoundry := initTokenFoundryAPI(config)
	onfido := initOnfidoAPI(config)

	var emailer *cemail.Emailer
	if config.SendgridKey != "" {
		emailer = cemail.NewEmailer(config.SendgridKey)
	}

	var mailchimpAPI *cemail.MailchimpAPI
	if config.MailchimpKey != "" {
		mailchimpAPI = cemail.NewMailchimpAPI(config.MailchimpKey)
	}

	userService, err := initUserService(config, nil, tokenControllerService)
	if err != nil {
		log.Fatalf("Error w init userService: err: %v", err)
		return nil, err
	}

	jsonbService, err := initJsonbService(config, nil)
	if err != nil {
		log.Fatalf("Error w init jsonbService: err: %v", err)
		return nil, err
	}
	nrsignupService, err := initNrsignupService(
		config,
		ethHelper.Blockchain,
		emailer,
		userService,
		jsonbService,
		jwtGenerator,
	)
	if err != nil {
		log.Fatalf("Error w init newsroom signup service: err: %v", err)
		return nil, err
	}

	authService, err := initAuthService(
		config,
		emailer,
		userService,
		jwtGenerator,
	)
	if err != nil {
		log.Fatalf("Error w init auth service: err: %v", err)
		return nil, err
	}

	storefrontService, err := initStorefrontService(
		config,
		ethHelper,
		userService,
		mailchimpAPI,
	)
	if err != nil {
		log.Fatalf("Error w init Storefront Service: err: %v", err)
		return nil, err
	}

	return &dependencies{
		emailer:                emailer,
		mailchimp:              mailchimpAPI,
		jwtGenerator:           jwtGenerator,
		invoicePersister:       invoicePersister,
		checkbookIO:            checkbookIO,
		userService:            userService,
		authService:            authService,
		jsonbService:           jsonbService,
		nrsignupService:        nrsignupService,
		tokenFoundry:           tokenFoundry,
		onfido:                 onfido,
		ethHelper:              ethHelper,
		storefrontService:      storefrontService,
		tokenControllerService: tokenControllerService,
	}, nil

}

func invoiceCheckbookIO(config *utils.GraphQLConfig) (*invoicing.CheckbookIO, error) {
	key := config.CheckbookKey
	secret := config.CheckbookSecret
	test := config.CheckbookTest

	if key == "" || secret == "" {
		return nil, errors.New("Checkbook key and secret required")
	}

	checkbookBaseURL := invoicing.ProdCheckbookIOBaseURL
	if test {
		checkbookBaseURL = invoicing.SandboxCheckbookIOBaseURL
	}

	checkbookIOClient := invoicing.NewCheckbookIO(
		checkbookBaseURL,
		key,
		secret,
		test,
	)
	return checkbookIOClient, nil
}

func initTokenFoundryAPI(config *utils.GraphQLConfig) *tokenfoundry.API {
	return tokenfoundry.NewAPI(
		"https://tokenfoundry.com",
		config.TokenFoundryUser,
		config.TokenFoundryPassword,
	)
}

func initOnfidoAPI(config *utils.GraphQLConfig) *kyc.OnfidoAPI {
	return kyc.NewOnfidoAPI(
		kyc.ProdAPIURL,
		config.OnfidoKey,
	)
}

func initETHHelper(config *utils.GraphQLConfig) (*eth.Helper, error) {
	if config.EthAPIURL != "" {
		// todo(dankins): we don't actually need any private keys yet, but we will for CIVIL-5
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
