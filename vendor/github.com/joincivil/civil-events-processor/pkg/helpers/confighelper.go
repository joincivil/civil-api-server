// Package helpers contains various common helper functions.
// Normally they are shared functions used by the cmds.
package helpers

import (
	// log "github.com/golang/glog"

	crawlermodel "github.com/joincivil/civil-events-crawler/pkg/model"
	crawlerpersist "github.com/joincivil/civil-events-crawler/pkg/persistence"

	"github.com/joincivil/civil-events-processor/pkg/model"
	"github.com/joincivil/civil-events-processor/pkg/persistence"
	"github.com/joincivil/civil-events-processor/pkg/scraper"
	"github.com/joincivil/civil-events-processor/pkg/utils"

	cconfig "github.com/joincivil/go-common/pkg/config"
)

// CronPersister is a helper function to return the correct cron persister based on
// the given configuration
func CronPersister(config cconfig.PersisterConfig) (model.CronPersister, error) {
	p, err := persister(config)
	if err != nil {
		return nil, err
	}
	return p.(model.CronPersister), nil
}

// EventPersister is a helper function to return the correct event persister based on
// the given configuration
func EventPersister(config cconfig.PersisterConfig) (crawlermodel.EventDataPersister, error) {
	if config.PersistType() == cconfig.PersisterTypePostgresql {
		persister, err := crawlerpersist.NewPostgresPersister(
			config.PostgresAddress(),
			config.PostgresPort(),
			config.PostgresUser(),
			config.PostgresPw(),
			config.PostgresDbname(),
		)
		if err != nil {
			return nil, err
		}
		return persister, nil
	}
	nullPersister := &crawlerpersist.NullPersister{}
	return nullPersister, nil
}

// ListingPersister is a helper function to return the correct listing persister based on
// the given configuration
func ListingPersister(config cconfig.PersisterConfig) (model.ListingPersister, error) {
	p, err := persister(config)
	if err != nil {
		return nil, err
	}
	return p.(model.ListingPersister), nil
}

// ContentRevisionPersister is a helper function to return the correct revision persister based on
// the given configuration
func ContentRevisionPersister(config cconfig.PersisterConfig) (model.ContentRevisionPersister, error) {
	p, err := persister(config)
	if err != nil {
		return nil, err
	}
	return p.(model.ContentRevisionPersister), nil
}

// GovernanceEventPersister is a helper function to return the correct gov event persister based on
// the given configuration
func GovernanceEventPersister(config cconfig.PersisterConfig) (model.GovernanceEventPersister, error) {
	p, err := persister(config)
	if err != nil {
		return nil, err
	}
	return p.(model.GovernanceEventPersister), nil
}

// ChallengePersister is a helper function to return the correct challenge persister based on
// the given configuration
func ChallengePersister(config cconfig.PersisterConfig) (model.ChallengePersister, error) {
	p, err := persister(config)
	if err != nil {
		return nil, err
	}
	return p.(model.ChallengePersister), nil
}

// AppealPersister is a helper function to return the correct appeals persister based on
// the given configuration
func AppealPersister(config cconfig.PersisterConfig) (model.AppealPersister, error) {
	p, err := persister(config)
	if err != nil {
		return nil, err
	}
	return p.(model.AppealPersister), nil
}

// PollPersister is a helper function to return the correct poll persister based on
// the given configuration
func PollPersister(config cconfig.PersisterConfig) (model.PollPersister, error) {
	p, err := persister(config)
	if err != nil {
		return nil, err
	}
	return p.(model.PollPersister), nil
}

func persister(config cconfig.PersisterConfig) (interface{}, error) {
	if config.PersistType() == cconfig.PersisterTypePostgresql {
		return postgresPersister(config)
	}
	// Default to the NullPersister
	return &persistence.NullPersister{}, nil
}

func postgresPersister(config cconfig.PersisterConfig) (*persistence.PostgresPersister, error) {
	persister, err := persistence.NewPostgresPersister(
		config.PostgresAddress(),
		config.PostgresPort(),
		config.PostgresUser(),
		config.PostgresPw(),
		config.PostgresDbname(),
	)
	if err != nil {
		// log.Errorf("Error connecting to Postgresql, stopping...; err: %v", err)
		return nil, err
	}
	// Attempts to create all the necessary tables here
	err = persister.CreateTables()
	if err != nil {
		// log.Errorf("Error creating tables, stopping...; err: %v", err)
		return nil, err
	}
	// Attempts to create all the necessary table indices here
	err = persister.CreateIndices()
	if err != nil {
		// log.Errorf("Error creating table indices, stopping...; err: %v", err)
		return nil, err
	}
	return persister, nil
}

// CivilMetadataScraper is a helper function to return a CivilMetadataScraper based on
// the given configuration
func CivilMetadataScraper(config *utils.ProcessorConfig) model.CivilMetadataScraper {
	return &scraper.CivilMetadataScraper{}
}

// ContentScraper is a helper function to return a ContentScraper based on
// the given configuration
func ContentScraper(config *utils.ProcessorConfig) model.ContentScraper {
	return &scraper.NullScraper{}
}

// MetadataScraper is a helper function to return a MetadataScraper based on
// the given configuration
func MetadataScraper(config *utils.ProcessorConfig) model.MetadataScraper {
	return &scraper.NullScraper{}
}
