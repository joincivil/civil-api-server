// Package utils contains various common utils separate by utility types
package utils

import (
	"errors"
	"fmt"
	"os"
	"text/tabwriter"

	putils "github.com/joincivil/civil-events-processor/pkg/utils"

	"github.com/kelseyhightower/envconfig"
)

const (
	envVarPrefixGraphQL = "graphql"

	usageListFormat = `The %v is configured via environment vars only. The following environment variables can be used:
{{range .}}
{{usage_key .}}
  description: {{usage_description .}}
  type:        {{usage_type .}}
  default:     {{usage_default .}}
  required:    {{usage_required .}}
{{end}}
`
)

// PersisterConfig defines the interfaces for persister-related configuration
type PersisterConfig interface {
	PersistType() putils.PersisterType
	PostgresAddress() string
	PostgresPort() int
	PostgresDbname() string
	PostgresUser() string
	PostgresPw() string
}

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

	PersisterType            putils.PersisterType `ignored:"true"`
	PersisterTypeName        string               `split_words:"true" required:"true" desc:"Sets the persister type to use"`
	PersisterPostgresAddress string               `split_words:"true" desc:"If persister type is Postgresql, sets the address"`
	PersisterPostgresPort    int                  `split_words:"true" desc:"If persister type is Postgresql, sets the port"`
	PersisterPostgresDbname  string               `split_words:"true" desc:"If persister type is Postgresql, sets the database name"`
	PersisterPostgresUser    string               `split_words:"true" desc:"If persister type is Postgresql, sets the database user"`
	PersisterPostgresPw      string               `split_words:"true" desc:"If persister type is Postgresql, sets the database password"`
}

// PersistType returns the persister type, implements PersisterConfig
func (c *GraphQLConfig) PersistType() putils.PersisterType {
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
	tabs := tabwriter.NewWriter(os.Stdout, 1, 0, 4, ' ', 0)
	usageFormat := fmt.Sprintf(usageListFormat, "graphql api")
	_ = envconfig.Usagef(envVarPrefixGraphQL, c, tabs, usageFormat) // nolint: gosec
	_ = tabs.Flush()                                                // nolint: gosec
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
	c.PersisterType, err = persisterTypeFromName(c.PersisterTypeName)
	return err
}

func (c *GraphQLConfig) validatePersister() error {
	var err error
	if c.PersisterType == putils.PersisterTypePostgresql {
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

func persisterTypeFromName(typeStr string) (putils.PersisterType, error) {
	pType, ok := putils.PersisterNameToType[typeStr]
	if !ok {
		validNames := make([]string, len(putils.PersisterNameToType))
		index := 0
		for name := range putils.PersisterNameToType {
			validNames[index] = name
			index++
		}
		return putils.PersisterTypeInvalid,
			fmt.Errorf("Invalid persister value: %v; valid types %v", typeStr, validNames)
	}
	return pType, nil
}
