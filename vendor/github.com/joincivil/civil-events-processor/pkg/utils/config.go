// Package utils contains various common utils separate by utility types
package utils

import (
	"errors"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/kelseyhightower/envconfig"
	"github.com/robfig/cron"

	crawlerutils "github.com/joincivil/civil-events-crawler/pkg/utils"
)

// PersisterType is the type of persister to use.
type PersisterType int

const (
	// PersisterTypeInvalid is an invalid persister value
	PersisterTypeInvalid PersisterType = iota

	// PersisterTypeNone is a persister that does nothing but return default values
	PersisterTypeNone

	// PersisterTypePostgresql is a persister that uses PostgreSQL as the backend
	PersisterTypePostgresql
)

var (
	// PersisterNameToType maps valid persister names to the types above
	PersisterNameToType = map[string]PersisterType{
		"none":       PersisterTypeNone,
		"postgresql": PersisterTypePostgresql,
	}
)

const (
	envVarPrefixProcessor = "processor"
	envVarPrefixGraphQL   = "graphql"

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
	PersistType() PersisterType
	PostgresAddress() string
	PostgresPort() int
	PostgresDbname() string
	PostgresUser() string
	PostgresPw() string
}

// NOTE(PN): After envconfig populates ProcessorConfig with the environment vars,
// there is nothing preventing the ProcessorConfig fields from being mutated.

// ProcessorConfig is the master config for the processor derived from environment
// variables.
type ProcessorConfig struct {
	CronConfig string `envconfig:"cron_config" required:"true" desc:"Cron config string * * * * *"`
	EthAPIURL  string `envconfig:"eth_api_url" required:"true" desc:"Ethereum API address"`

	PersisterType            PersisterType `ignored:"true"`
	PersisterTypeName        string        `split_words:"true" required:"true" desc:"Sets the persister type to use"`
	PersisterPostgresAddress string        `split_words:"true" desc:"If persister type is Postgresql, sets the address"`
	PersisterPostgresPort    int           `split_words:"true" desc:"If persister type is Postgresql, sets the port"`
	PersisterPostgresDbname  string        `split_words:"true" desc:"If persister type is Postgresql, sets the database name"`
	PersisterPostgresUser    string        `split_words:"true" desc:"If persister type is Postgresql, sets the database user"`
	PersisterPostgresPw      string        `split_words:"true" desc:"If persister type is Postgresql, sets the database password"`
}

// PersistType returns the persister type, implements PersisterConfig
func (c *ProcessorConfig) PersistType() PersisterType {
	return c.PersisterType
}

// PostgresAddress returns the postgres persister address, implements PersisterConfig
func (c *ProcessorConfig) PostgresAddress() string {
	return c.PersisterPostgresAddress
}

// PostgresPort returns the postgres persister port, implements PersisterConfig
func (c *ProcessorConfig) PostgresPort() int {
	return c.PersisterPostgresPort
}

// PostgresDbname returns the postgres persister db name, implements PersisterConfig
func (c *ProcessorConfig) PostgresDbname() string {
	return c.PersisterPostgresDbname
}

// PostgresUser returns the postgres persister user, implements PersisterConfig
func (c *ProcessorConfig) PostgresUser() string {
	return c.PersisterPostgresUser
}

// PostgresPw returns the postgres persister password, implements PersisterConfig
func (c *ProcessorConfig) PostgresPw() string {
	return c.PersisterPostgresPw
}

// OutputUsage prints the usage string to os.Stdout
func (c *ProcessorConfig) OutputUsage() {
	tabs := tabwriter.NewWriter(os.Stdout, 1, 0, 4, ' ', 0)
	usageFormat := fmt.Sprintf(usageListFormat, "processor")
	_ = envconfig.Usagef(envVarPrefixProcessor, c, tabs, usageFormat) // nolint: gosec
	_ = tabs.Flush()                                                  // nolint: gosec
}

// PopulateFromEnv processes the environment vars, populates ProcessorConfig
// with the respective values, and validates the values.
func (c *ProcessorConfig) PopulateFromEnv() error {
	err := envconfig.Process(envVarPrefixProcessor, c)
	if err != nil {
		return err
	}

	err = c.validateCronConfig()
	if err != nil {
		return err
	}

	err = c.validateAPIURL()
	if err != nil {
		return err
	}

	err = c.populatePersisterType()
	if err != nil {
		return err
	}

	return c.validatePersister()
}

func (c *ProcessorConfig) validateCronConfig() error {
	parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	_, err := parser.Parse(c.CronConfig)
	if err != nil {
		return fmt.Errorf("Invalid cron config: '%v'", c.CronConfig)
	}
	return nil
}

func (c *ProcessorConfig) validateAPIURL() error {
	if c.EthAPIURL == "" || !crawlerutils.IsValidEthAPIURL(c.EthAPIURL) {
		return fmt.Errorf("Invalid eth API URL: '%v'", c.EthAPIURL)
	}
	return nil
}

func (c *ProcessorConfig) validatePersister() error {
	var err error
	if c.PersisterType == PersisterTypePostgresql {
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

func (c *ProcessorConfig) populatePersisterType() error {
	var err error
	c.PersisterType, err = persisterTypeFromName(c.PersisterTypeName)
	return err
}

// NOTE(PN): After envconfig populates GraphQLConfig with the environment vars,
// there is nothing preventing the GraphQLConfig fields from being mutated.

// GraphQLConfig is the master config for the GraphQL API derived from environment
// variables.
type GraphQLConfig struct {
	Port  int  `required:"true" desc:"Sets the GraphQL service port"`
	Debug bool `default:"false" desc:"If true, enables the GraphQL playground"`

	PersisterType            PersisterType `ignored:"true"`
	PersisterTypeName        string        `split_words:"true" required:"true" desc:"Sets the persister type to use"`
	PersisterPostgresAddress string        `split_words:"true" desc:"If persister type is Postgresql, sets the address"`
	PersisterPostgresPort    int           `split_words:"true" desc:"If persister type is Postgresql, sets the port"`
	PersisterPostgresDbname  string        `split_words:"true" desc:"If persister type is Postgresql, sets the database name"`
	PersisterPostgresUser    string        `split_words:"true" desc:"If persister type is Postgresql, sets the database user"`
	PersisterPostgresPw      string        `split_words:"true" desc:"If persister type is Postgresql, sets the database password"`
}

// PersistType returns the persister type, implements PersisterConfig
func (c *GraphQLConfig) PersistType() PersisterType {
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
	if c.PersisterType == PersisterTypePostgresql {
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

func persisterTypeFromName(typeStr string) (PersisterType, error) {
	pType, ok := PersisterNameToType[typeStr]
	if !ok {
		validNames := make([]string, len(PersisterNameToType))
		index := 0
		for name := range PersisterNameToType {
			validNames[index] = name
			index++
		}
		return PersisterTypeInvalid,
			fmt.Errorf("Invalid persister value: %v; valid types %v", typeStr, validNames)
	}
	return pType, nil
}
