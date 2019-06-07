package graphqlmain

import (
	"fmt"

	"github.com/ethereum/go-ethereum"
	"github.com/jinzhu/gorm"

	"github.com/joincivil/civil-api-server/pkg/auth"
	"github.com/joincivil/civil-api-server/pkg/jsonstore"
	"github.com/joincivil/civil-api-server/pkg/nrsignup"
	"github.com/joincivil/civil-api-server/pkg/payments"
	"github.com/joincivil/civil-api-server/pkg/storefront"
	"github.com/joincivil/civil-api-server/pkg/tokencontroller"
	"github.com/joincivil/civil-api-server/pkg/users"
	"github.com/joincivil/civil-api-server/pkg/utils"

	"github.com/joincivil/go-common/pkg/eth"

	cemail "github.com/joincivil/go-common/pkg/email"
)

func initUserService(config *utils.GraphQLConfig, userPersister *users.PostgresPersister,
	tokenControllerService *tokencontroller.Service) (
	*users.UserService, error) {
	userService := users.NewUserService(userPersister, tokenControllerService)
	if userService == nil {
		return nil, fmt.Errorf("User service was not initialized")
	}
	return userService, nil

}

func initNrsignupService(config *utils.GraphQLConfig, ethHelper *eth.Helper,
	emailer *cemail.Emailer, userService *users.UserService, jsonbService *jsonstore.Service,
	jwtGenerator *auth.JwtTokenGenerator) (*nrsignup.Service, error) {
	nrsignupService, err := nrsignup.NewNewsroomSignupService(
		ethHelper,
		emailer,
		userService,
		jsonbService,
		jwtGenerator,
		config.ApproveGrantProtoHost,
		config.ContractAddresses["civilparameterizer"],
		config.RegistryAlertsID,
	)
	if err != nil {
		return nil, err
	}
	return nrsignupService, nil
}

func initJsonbService(config *utils.GraphQLConfig, jsonbPersister jsonstore.JsonbPersister) (
	*jsonstore.Service, error) {
	if jsonbPersister == nil {
		var perr error
		jsonbPersister, perr = initJsonbPersister(config)
		if perr != nil {
			return nil, perr
		}
	}
	jsonbService := jsonstore.NewJsonbService(jsonbPersister)
	return jsonbService, nil
}

func initStorefrontService(config *utils.GraphQLConfig, ethHelper *eth.Helper,
	userService *users.UserService, mailchimp cemail.ListMemberManager) (*storefront.Service, error) {
	emailLists := storefront.NewMailchimpServiceEmailLists(mailchimp)

	return storefront.NewService(
		config.ContractAddresses["CVLToken"],
		config.TokenSaleAddresses,
		ethHelper,
		userService,
		emailLists,
	)
}

func initTokenControllerService(config *utils.GraphQLConfig, ethHelper *eth.Helper) (
	*tokencontroller.Service, error) {
	return tokencontroller.NewService(config.ContractAddresses["CivilTokenController"], ethHelper)
}

func initPaymentService(config *utils.GraphQLConfig, db *gorm.DB, ethHelper *eth.Helper) *payments.Service {
	stripe := payments.NewStripeService(config.StripeAPIKey)
	ethereum := payments.NewEthereumService(ethHelper.Blockchain.(ethereum.TransactionReader))
	return payments.NewService(db, stripe, ethereum)
}
