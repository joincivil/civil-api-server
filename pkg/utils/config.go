// Package utils contains various common utils separate by utility types
package utils

import (
	"errors"

	"github.com/ethereum/go-ethereum/common"
	cconfig "github.com/joincivil/go-common/pkg/config"

	"github.com/kelseyhightower/envconfig"
)

const (
	envVarPrefixGraphQL = "graphql"
)

// NOTE(PN): After envconfig populates GraphQLConfig with the environment vars,
// there is nothing preventing the GraphQLConfig fields from being mutated.

// GraphQLConfig is the master config for the GraphQL API derived from environment
// variables.
type GraphQLConfig struct {
	Port  int  `required:"true" desc:"Sets the GraphQL service port"`
	Debug bool `default:"false" desc:"If true, enables the GraphQL playground"`

	JwtSecret string `split_words:"true" desc:"Secret used to encode JWT tokens"`

	AuthEmailSignupTemplates map[string]string `split_words:"true" required:"false" desc:"<appname>:<template id>,..."`
	AuthEmailLoginTemplates  map[string]string `split_words:"true" required:"false" desc:"<appname>:<template id>,..."`

	ApproveGrantProtoHost string `split_words:"true" desc:"Newsroom signup grant approval landing proto/host" required:"false"`
	SignupLoginProtoHost  string `split_words:"true" desc:"Signup/login proto/host" required:"false"`

	RegistryAlertsID string `split_words:"true" desc:"Sets the registry alerts list ID"`
	SendgridKey      string `split_words:"true" desc:"The SendGrid API key"`
	MailchimpKey     string `split_words:"true" desc:"The Mailchimp API key"`

	OnfidoKey          string `split_words:"true" desc:"The Onfido API key"`
	OnfidoReferrer     string `split_words:"true" desc:"The Onfido token referrer"`
	OnfidoWebhookToken string `split_words:"true" desc:"The Onfido webhook secret token"`

	CheckbookKey    string `split_words:"true" desc:"The checkbook.io api key"`
	CheckbookSecret string `split_words:"true" desc:"The checkbook.io api secret"`
	CheckbookTest   bool   `split_words:"true" default:"false" desc:"If true, enables uses the checkbook sandbox"`

	PersisterType            cconfig.PersisterType `ignored:"true"`
	PersisterTypeName        string                `split_words:"true" required:"true" desc:"Sets the persister type to use"`
	PersisterPostgresAddress string                `split_words:"true" desc:"If persister type is Postgresql, sets the address"`
	PersisterPostgresPort    int                   `split_words:"true" desc:"If persister type is Postgresql, sets the port"`
	PersisterPostgresDbname  string                `split_words:"true" desc:"If persister type is Postgresql, sets the database name"`
	PersisterPostgresUser    string                `split_words:"true" desc:"If persister type is Postgresql, sets the database user"`
	PersisterPostgresPw      string                `split_words:"true" desc:"If persister type is Postgresql, sets the database password"`

	TokenFoundryUser     string `split_words:"true" desc:"TokenFoundry User"`
	TokenFoundryPassword string `split_words:"true" desc:"TokenFoundry Password"`

	EthAPIURL string `envconfig:"eth_api_url" desc:"Ethereum API address"`

	// ContractAddresses map a contract type to a string of contract addresses.  If there are more than 1
	// contract to be tracked for a particular type, delimit the addresses with '|'.
	ContractAddresses map[string]string `split_words:"true" desc:"<contract name>:<contract addr>. Delimit contract address with '|' for multiple addresses"`

	TokenSaleAddresses []common.Address `split_words:"true" desc:"Addresses that contain tokens to be sold as part of the Token Sale"`

	EthereumDefaultPrivateKey string `split_words:"true" desc:"Private key to use when sending Ethereum transactions"`

	RefreshTokenBlacklist []string `split_words:"true" desc:"List of refresh tokens to blacklist"`
}

// PersistType returns the persister type, implements PersisterConfig
func (c *GraphQLConfig) PersistType() cconfig.PersisterType {
	return c.PersisterType
}

// PostgresAddress returns the postgres persister address, implements PersisterConfig
func (c *GraphQLConfig) PostgresAddress() string {
	return c.PersisterPostgresAddress
}

// PostgresPort returns the postgres persister port, implements PersisterConfig
func (c *GraphQLConfig) PostgresPort() int {
	return c.PersisterPostgresPort
}

// PostgresDbname returns the postgres persister db name, implements PersisterConfig
func (c *GraphQLConfig) PostgresDbname() string {
	return c.PersisterPostgresDbname
}

// PostgresUser returns the postgres persister user, implements PersisterConfig
func (c *GraphQLConfig) PostgresUser() string {
	return c.PersisterPostgresUser
}

// PostgresPw returns the postgres persister password, implements PersisterConfig
func (c *GraphQLConfig) PostgresPw() string {
	return c.PersisterPostgresPw
}

// OutputUsage prints the usage string to os.Stdout
func (c *GraphQLConfig) OutputUsage() {
	cconfig.OutputUsage(c, "graphql", envVarPrefixGraphQL)
}

// PopulateFromEnv processes the environment vars, populates GraphQLConfig
// with the respective values, and validates the values.
func (c *GraphQLConfig) PopulateFromEnv() error {
	err := envconfig.Process(envVarPrefixGraphQL, c)
	if err != nil {
		return err
	}

	err = c.populatePersisterType()
	if err != nil {
		return err
	}

	err = c.populateAuthEmailTemplates()
	if err != nil {
		return err
	}

	return c.validatePersister()
}

func (c *GraphQLConfig) populatePersisterType() error {
	var err error
	c.PersisterType, err = cconfig.PersisterTypeFromName(c.PersisterTypeName)
	return err
}

func (c *GraphQLConfig) validatePersister() error {
	var err error
	if c.PersisterType == cconfig.PersisterTypePostgresql {
		err = validatePostgresqlPersisterParams(
			c.PersisterPostgresAddress,
			c.PersisterPostgresPort,
			c.PersisterPostgresDbname,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *GraphQLConfig) populateAuthEmailTemplates() error {
	return nil
}

func validatePostgresqlPersisterParams(address string, port int, dbname string) error {
	if address == "" {
		return errors.New("Postgresql address required")
	}
	if port == 0 {
		return errors.New("Postgresql port required")
	}
	if dbname == "" {
		return errors.New("Postgresql db name required")
	}
	return nil
}
