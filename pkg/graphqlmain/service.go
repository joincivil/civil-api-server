package graphqlmain

import (
	shell "github.com/ipfs/go-ipfs-api"
	"github.com/joincivil/civil-api-server/pkg/discourse"

	"github.com/ethereum/go-ethereum/common"

	"github.com/joincivil/civil-api-server/pkg/auth"
	"github.com/joincivil/civil-api-server/pkg/jsonstore"
	"github.com/joincivil/civil-api-server/pkg/nrsignup"
	"github.com/joincivil/civil-api-server/pkg/storefront"
	"github.com/joincivil/civil-api-server/pkg/users"
	"github.com/joincivil/civil-api-server/pkg/utils"

	"github.com/joincivil/go-common/pkg/eth"

	cemail "github.com/joincivil/go-common/pkg/email"
)

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

func initIPFS() *shell.Shell {
	return shell.NewShell("https://ipfs.infura.io:5001")
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

// NewDeployerContractAddresses builds a new DeployerContractAddresses instance
func NewDeployerContractAddresses(config *utils.GraphQLConfig) eth.DeployerContractAddresses {
	return eth.DeployerContractAddresses{
		NewsroomFactory:       extractContractAddress(config, "NewsroomFactory"),
		MultisigFactory:       extractContractAddress(config, "MultisigFactory"),
		CivilTokenController:  extractContractAddress(config, "CivilTokenController"),
		CreateNewsroomInGroup: extractContractAddress(config, "CreateNewsroomInGroup"),
		PLCR:                  extractContractAddress(config, "PLCR"),
		TCR:                   extractContractAddress(config, "PLCR"),
		CVLToken:              extractContractAddress(config, "CVLToken"),
		Parameterizer:         extractContractAddress(config, "Parameterizer"),
		Government:            extractContractAddress(config, "Government"),
	}
}

func extractContractAddress(config *utils.GraphQLConfig, contractName string) common.Address {
	return common.HexToAddress(config.ContractAddresses[contractName])
}

func initDiscourseService(config *utils.GraphQLConfig, listingMapPersister discourse.ListingMapPersister) (
	*discourse.Service, error) {
	return discourse.NewService(listingMapPersister), nil
}
