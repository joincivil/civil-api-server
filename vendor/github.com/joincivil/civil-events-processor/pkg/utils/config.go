// Package utils contains various common utils separate by utility types
package utils

import (
	"errors"
	"fmt"

	"github.com/kelseyhightower/envconfig"
	"github.com/robfig/cron"

	cconfig "github.com/joincivil/go-common/pkg/config"
	cstrings "github.com/joincivil/go-common/pkg/strings"
)

const (
	envVarPrefixProcessor = "processor"
)

// NOTE(PN): After envconfig populates ProcessorConfig with the environment vars,
// there is nothing preventing the ProcessorConfig fields from being mutated.

// ProcessorConfig is the master config for the processor derived from environment
// variables.
type ProcessorConfig struct {
	CronConfig string `envconfig:"cron_config" required:"true" desc:"Cron config string * * * * *"`
	EthAPIURL  string `envconfig:"eth_api_url" required:"true" desc:"Ethereum API address"`

	PersisterType            cconfig.PersisterType `ignored:"true"`
	PersisterTypeName        string                `split_words:"true" required:"true" desc:"Sets the persister type to use"`
	PersisterPostgresAddress string                `split_words:"true" desc:"If persister type is Postgresql, sets the address"`
	PersisterPostgresPort    int                   `split_words:"true" desc:"If persister type is Postgresql, sets the port"`
	PersisterPostgresDbname  string                `split_words:"true" desc:"If persister type is Postgresql, sets the database name"`
	PersisterPostgresUser    string                `split_words:"true" desc:"If persister type is Postgresql, sets the database user"`
	PersisterPostgresPw      string                `split_words:"true" desc:"If persister type is Postgresql, sets the database password"`
}

// PersistType returns the persister type, implements PersisterConfig
func (c *ProcessorConfig) PersistType() cconfig.PersisterType {
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
	cconfig.OutputUsage(c, envVarPrefixProcessor, envVarPrefixProcessor)
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
	if c.EthAPIURL == "" || !cstrings.IsValidEthAPIURL(c.EthAPIURL) {
		return fmt.Errorf("Invalid eth API URL: '%v'", c.EthAPIURL)
	}
	return nil
}

func (c *ProcessorConfig) validatePersister() error {
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

func (c *ProcessorConfig) populatePersisterType() error {
	var err error
	c.PersisterType, err = cconfig.PersisterTypeFromName(c.PersisterTypeName)
	return err
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
