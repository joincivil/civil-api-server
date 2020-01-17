// Package utils contains various common utils separate by utility types
package utils

import (
	"errors"
	"fmt"
	"strings"

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
	GqlPort int  `required:"true" desc:"Sets the GraphQL service port"`
	Debug   bool `default:"false" desc:"If true, enables the GraphQL playground"`

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

	PersisterType             cconfig.PersisterType `ignored:"true"`
	PersisterTypeName         string                `split_words:"true" required:"true" desc:"Sets the persister type to use"`
	PersisterPostgresAddress  string                `split_words:"true" desc:"If persister type is Postgresql, sets the address"`
	PersisterPostgresPort     int                   `split_words:"true" desc:"If persister type is Postgresql, sets the port"`
	PersisterPostgresDbname   string                `split_words:"true" desc:"If persister type is Postgresql, sets the database name"`
	PersisterPostgresUser     string                `split_words:"true" desc:"If persister type is Postgresql, sets the database user"`
	PersisterPostgresPw       string                `split_words:"true" desc:"If persister type is Postgresql, sets the database password"`
	PersisterPostgresMaxConns *int                  `split_words:"true" desc:"If persister type is Postgresql, sets the max conns in pool"`
	PersisterPostgresMaxIdle  *int                  `split_words:"true" desc:"If persister type is Postgresql, sets the max idle conns in pool"`
	PersisterPostgresConnLife *int                  `split_words:"true" desc:"If persister type is Postgresql, sets the max conn lifetime in secs"`

	VersionNumber string `split_words:"true" desc:"Sets the version to use for crawler related Postgres tables"`

	StripeAPIKey          string   `envconfig:"stripe_api_key" split_words:"true" desc:"API key for stripe"`
	StripeApplePayDomains []string `split_words:"true" desc:"Domains to enable Apple Pay on" default:"" `

	TokenFoundryUser     string `split_words:"true" desc:"TokenFoundry User"`
	TokenFoundryPassword string `split_words:"true" desc:"TokenFoundry Password"`

	EthAPIURL string `envconfig:"eth_api_url" desc:"Ethereum API address"`

	// ContractAddresses map a contract type to a string of contract addresses.  If there are more than 1
	// contract to be tracked for a particular type, delimit the addresses with '|'.
	ContractAddresses map[string]string `split_words:"true" desc:"<contract name>:<contract addr>. Delimit contract address with '|' for multiple addresses"`

	TokenSaleAddresses []common.Address `split_words:"true" desc:"Addresses that contain tokens to be sold as part of the Token Sale"`

	EthereumDefaultPrivateKey string `split_words:"true" desc:"Private key to use when sending Ethereum transactions"`

	RefreshTokenBlacklist []string `split_words:"true" desc:"List of refresh tokens to blacklist"`

	FastPassRescueMultisig common.Address `split_words:"true" desc:"Address to add to FastPassed newsroom multisigs"`
	TcrApplicationTokens   int64          `split_words:"true" desc:"Number of tokens needed to apply to registry"`

	// Runs the pubsub worker
	// Should eventually move this to it's own repo and codebase
	PubSubProjectID           string `split_words:"true" desc:"Sets GPubSub project ID. If not set, will not pub or sub."`
	PubSubTokenTopicName      string `split_words:"true" desc:"Sets GPubSub topic name for cvltoken events."`
	PubSubTokenSubName        string `split_words:"true" desc:"Sets GPubSub subscription name for cvltoken events."`
	PubSubMultiSigTopicName   string `split_words:"true" desc:"Sets GPubSub topic name for multi sig events."`
	PubSubMultiSigSubName     string `split_words:"true" desc:"Sets GPubSub subscription name for multi sig events."`
	PubSubGovernanceTopicName string `split_words:"true" desc:"Sets GPubSub topic name for governance events."`
	PubSubGovernanceSubName   string `split_words:"true" desc:"Sets GPubSub subscription name for governance events."`

	StackDriverProjectID string `split_words:"true" desc:"Sets the Stackdriver project ID"`
	SentryDsn            string `split_words:"true" desc:"Sets the Sentry DSN"`
	SentryEnv            string `split_words:"true" desc:"Sets the Sentry environment"`
}

// PersistType returns the persister type, implements PersisterConfig
func (c *GraphQLConfig) PersistType() cconfig.PersisterType {
	return c.PersisterType
}

// Address returns the persister address, implements PersisterConfig
func (c *GraphQLConfig) Address() string {
	return c.PersisterPostgresAddress
}

// Port returns the persister port, implements PersisterConfig
func (c *GraphQLConfig) Port() int {
	return c.PersisterPostgresPort
}

// Dbname returns the persister db name, implements PersisterConfig
func (c *GraphQLConfig) Dbname() string {
	return c.PersisterPostgresDbname
}

// User returns the persister user, implements PersisterConfig
func (c *GraphQLConfig) User() string {
	return c.PersisterPostgresUser
}

// Password returns the persister password, implements PersisterConfig
func (c *GraphQLConfig) Password() string {
	return c.PersisterPostgresPw
}

// PoolMaxConns returns the max conns for a pool, if configured, implements PersisterConfig
func (c *GraphQLConfig) PoolMaxConns() *int {
	return c.PersisterPostgresMaxConns
}

// PoolMaxIdleConns returns the max idleconns for a pool, if configured, implements PersisterConfig
func (c *GraphQLConfig) PoolMaxIdleConns() *int {
	return c.PersisterPostgresMaxIdle
}

// PoolConnLifetimeSecs returns the conn lifetime for a pool, if configured, implements PersisterConfig
func (c *GraphQLConfig) PoolConnLifetimeSecs() *int {
	return c.PersisterPostgresConnLife
}

// OutputUsage prints the usage string to os.Stdout
func (c *GraphQLConfig) OutputUsage() {
	cconfig.OutputUsage(c, "graphql", envVarPrefixGraphQL)
}

// PopulateFromEnv processes the environment vars, populates GraphQLConfig
// with the respective values, and validates the values.
func (c *GraphQLConfig) PopulateFromEnv() error {
	envEnvVar := fmt.Sprintf("%v_ENV", strings.ToUpper(envVarPrefixGraphQL))
	err := cconfig.PopulateFromDotEnv(envEnvVar)
	if err != nil {
		return err
	}

	err = envconfig.Process(envVarPrefixGraphQL, c)
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
