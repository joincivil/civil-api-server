// Package utils contains various common utils separate by utility types
package utils

import (
	"errors"

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
	Port            int  `required:"true" desc:"Sets the GraphQL service port"`
	Debug           bool `default:"false" desc:"If true, enables the GraphQL playground"`
	EnableGraphQL   bool `envconfig:"enable_graphql" split_words:"true" default:"true" desc:"If true, enables the GraphQL endpoint"`
	EnableInvoicing bool `split_words:"true" default:"false" desc:"If true, enables the invoicing endpoint"`
	EnableKYC       bool `split_words:"true" default:"false" desc:"If true, enables the KYC endpoint"`

	JwtSecret string `split_words:"true" desc:"Secret used to encode JWT tokens"`

	SendgridKey string `split_words:"true" desc:"The SendGrid API key"`

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
