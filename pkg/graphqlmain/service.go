package graphqlmain

import (
	shell "github.com/ipfs/go-ipfs-api"
	"github.com/joincivil/civil-api-server/pkg/discourse"

	"github.com/joincivil/civil-api-server/pkg/storefront"
	"github.com/joincivil/civil-api-server/pkg/users"
	"github.com/joincivil/civil-api-server/pkg/utils"

	"github.com/joincivil/go-common/pkg/eth"

	cemail "github.com/joincivil/go-common/pkg/email"
)

func initIPFS() *shell.Shell {
	return shell.NewShell("https://ipfs.infura.io:5001")
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

func initDiscourseService(config *utils.GraphQLConfig, listingMapPersister discourse.ListingMapPersister) (
	*discourse.Service, error) {
	return discourse.NewService(listingMapPersister), nil
}
