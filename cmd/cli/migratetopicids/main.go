package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/jinzhu/gorm"
	"github.com/kelseyhightower/envconfig"

	"github.com/joincivil/civil-api-server/pkg/discourse"

	"github.com/joincivil/civil-events-processor/pkg/helpers"
	"github.com/joincivil/civil-events-processor/pkg/model"
	cconfig "github.com/joincivil/go-common/pkg/config"
)

// MigrationConfig is the config for this script
type MigrationConfig struct {
	PersisterPostgresAddress string `split_words:"true" desc:"Sets the address"`
	PersisterPostgresPort    int    `split_words:"true" desc:"Sets the port"`
	PersisterPostgresDbname  string `split_words:"true" desc:"Sets the database name"`
	PersisterPostgresUser    string `split_words:"true" desc:"Sets the database user"`
	PersisterPostgresPw      string `split_words:"true" desc:"Sets the database password"`

	VersionNumber string `split_words:"true" desc:"Sets the version to use for crawler related Postgres tables"`
}

// PersistType returns the persister type, implements PersisterConfig
func (c *MigrationConfig) PersistType() cconfig.PersisterType {
	return cconfig.PersisterTypePostgresql
}

// PostgresAddress returns the postgres persister address, implements PersisterConfig
func (c *MigrationConfig) PostgresAddress() string {
	return c.PersisterPostgresAddress
}

// PostgresPort returns the postgres persister port, implements PersisterConfig
func (c *MigrationConfig) PostgresPort() int {
	return c.PersisterPostgresPort
}

// PostgresDbname returns the postgres persister db name, implements PersisterConfig
func (c *MigrationConfig) PostgresDbname() string {
	return c.PersisterPostgresDbname
}

// PostgresUser returns the postgres persister user, implements PersisterConfig
func (c *MigrationConfig) PostgresUser() string {
	return c.PersisterPostgresUser
}

// PostgresPw returns the postgres persister password, implements PersisterConfig
func (c *MigrationConfig) PostgresPw() string {
	return c.PersisterPostgresPw
}

// OutputUsage prints the usage string to os.Stdout
func (c *MigrationConfig) OutputUsage() {
	cconfig.OutputUsage(c, "migrate", "migrate")
}

func persisters(config *MigrationConfig) (model.ListingPersister, discourse.ListingMapPersister, error) {
	dbstr := fmt.Sprintf("host=%v port=%v user=%v dbname=%v password=%v sslmode=disable",
		config.PostgresAddress(),
		config.PostgresPort(),
		config.PostgresUser(),
		config.PostgresDbname(),
		config.PostgresPw(),
	)
	fmt.Printf("dbstr = %v\n", dbstr)

	db, err := gorm.Open(
		"postgres",
		fmt.Sprintf("host=%v port=%v user=%v dbname=%v password=%v sslmode=disable",
			config.PostgresAddress(),
			config.PostgresPort(),
			config.PostgresUser(),
			config.PostgresDbname(),
			config.PostgresPw(),
		))
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return nil, nil, err
	}

	amErr := db.AutoMigrate(
		&discourse.ListingMap{},
	).Error
	if amErr != nil {
		fmt.Printf("automigration error: %v\n", amErr)
		return nil, nil, amErr
	}

	listingPersister, err := helpers.ListingPersister(config, config.VersionNumber)
	if err != nil {
		fmt.Printf("error init persister: %v\n", err)
		return nil, nil, err
	}

	discoursePersister, err := discourse.NewPostgresPersister(db)
	if err != nil {
		fmt.Printf("error init persister: %v\n", err)
		return nil, nil, err
	}

	return listingPersister, discoursePersister, nil
}

func runMigrate(listingPersister model.ListingPersister, discourseService *discourse.Service) {
	listings, err := listingPersister.ListingsByCriteria(&model.ListingCriteria{
		Count: 200,
	})
	if err != nil {
		fmt.Printf("error retrieving listings: %v", err)
		return
	}

	for _, listing := range listings {
		if listing.DiscourseTopicID() != 0 {
			err = discourseService.SaveDiscourseTopicID(
				listing.ContractAddress(),
				listing.DiscourseTopicID(),
			)
			if err != nil {
				fmt.Printf("save error: %v\n", err)
			}
		}
	}
}

func main() {
	config := &MigrationConfig{}
	flag.Usage = func() {
		config.OutputUsage()
		os.Exit(0)
	}
	flag.Parse()

	err := cconfig.PopulateFromDotEnv("MIGRATE_ENV")
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}

	err = envconfig.Process("migrate", config)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}

	listingPersister, discoursePersister, err := persisters(config)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	service := discourse.NewService(discoursePersister)

	runMigrate(listingPersister, service)
	fmt.Printf("done.\n")
}
