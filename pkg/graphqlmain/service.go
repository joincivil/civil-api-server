package graphqlmain

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/jinzhu/gorm"

	"github.com/joincivil/civil-api-server/pkg/auth"
	"github.com/joincivil/civil-api-server/pkg/jsonstore"
	"github.com/joincivil/civil-api-server/pkg/nrsignup"
	"github.com/joincivil/civil-api-server/pkg/posts"
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
	if userPersister == nil {
		var perr error
		userPersister, perr = initUserPersister(config)
		if perr != nil {
			return nil, perr
		}
	}
	userService := users.NewUserService(userPersister, tokenControllerService)
	if userService == nil {
		return nil, fmt.Errorf("User service was not initialized")
	}
	return userService, nil

}

func initNrsignupService(config *utils.GraphQLConfig, client bind.ContractBackend,
	emailer *cemail.Emailer, userService *users.UserService, jsonbService *jsonstore.Service,
	jwtGenerator *auth.JwtTokenGenerator) (*nrsignup.Service, error) {
	nrsignupService, err := nrsignup.NewNewsroomSignupService(
		client,
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
	userService *users.UserService, mailchimp *cemail.MailchimpAPI) (*storefront.Service, error) {
	emailLists := storefront.NewMailchimpServiceEmailLists(mailchimp)

	return storefront.NewService(
		config.ContractAddresses["CVLToken"],
		config.TokenSaleAddresses,
		ethHelper,
		userService,
		emailLists,
	)
}

func initAuthService(config *utils.GraphQLConfig, emailer *cemail.Emailer,
	userService *users.UserService, jwtGenerator *auth.JwtTokenGenerator) (*auth.Service, error) {
	return auth.NewAuthService(
		userService,
		jwtGenerator,
		emailer,
		config.AuthEmailSignupTemplates,
		config.AuthEmailLoginTemplates,
		config.SignupLoginProtoHost,
		config.RefreshTokenBlacklist,
	)
}

func initTokenControllerService(config *utils.GraphQLConfig, ethHelper *eth.Helper) (
	*tokencontroller.Service, error) {
	return tokencontroller.NewService(config.ContractAddresses["CivilTokenController"], ethHelper)
}

func initPostService(config *utils.GraphQLConfig, db *gorm.DB) *posts.Service {
	persister := posts.NewDBPostPersister(db)
	return posts.NewService(persister)
}
