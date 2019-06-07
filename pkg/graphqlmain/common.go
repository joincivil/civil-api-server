package graphqlmain

import (
	"os"
	"os/signal"
	"syscall"

	log "github.com/golang/glog"

	"github.com/joincivil/civil-api-server/pkg/auth"
	"github.com/joincivil/civil-api-server/pkg/jsonstore"
	"github.com/joincivil/civil-api-server/pkg/nrsignup"
	"github.com/joincivil/civil-api-server/pkg/payments"
	"github.com/joincivil/civil-api-server/pkg/posts"
	"github.com/joincivil/civil-api-server/pkg/storefront"
	"github.com/joincivil/civil-api-server/pkg/tokencontroller"
	"github.com/joincivil/civil-api-server/pkg/users"
	"github.com/joincivil/civil-api-server/pkg/utils"

	"github.com/joincivil/go-common/pkg/eth"

	cemail "github.com/joincivil/go-common/pkg/email"
)

// SetupKillNotify sets up the kill signal hook
func SetupKillNotify(quitChan chan<- bool) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		close(quitChan)
		os.Exit(1)
	}()
}

type dependencies struct {
	emailer                *cemail.Emailer
	mailchimp              *cemail.MailchimpAPI
	jwtGenerator           *auth.JwtTokenGenerator
	userService            *users.UserService
	authService            *auth.Service
	jsonbService           *jsonstore.Service
	nrsignupService        *nrsignup.Service
	paymentService         *payments.Service
	postService            *posts.Service
	ethHelper              *eth.Helper
	storefrontService      *storefront.Service
	tokenControllerService *tokencontroller.Service
}

func initDependencies(config *utils.GraphQLConfig) (*dependencies, error) {
	var err error

	db, err := initGorm(config)
	if err != nil {
		log.Fatalf("Error initializing database: err: %v", err)
		return nil, err
	}

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

	jwtGenerator := auth.NewJwtTokenGenerator([]byte(config.JwtSecret))

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

	postService := initPostService(config, db)

	paymentService := initPaymentService(config, db, ethHelper)

	return &dependencies{
		emailer:                emailer,
		mailchimp:              mailchimpAPI,
		jwtGenerator:           jwtGenerator,
		userService:            userService,
		authService:            authService,
		jsonbService:           jsonbService,
		nrsignupService:        nrsignupService,
		ethHelper:              ethHelper,
		postService:            postService,
		storefrontService:      storefrontService,
		paymentService:         paymentService,
		tokenControllerService: tokenControllerService,
	}, nil

}

func initETHHelper(config *utils.GraphQLConfig) (*eth.Helper, error) {
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
