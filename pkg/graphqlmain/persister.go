package graphqlmain

import (
	"fmt"

	"github.com/jinzhu/gorm"
	"github.com/joincivil/civil-api-server/pkg/discourse"

	"github.com/joincivil/civil-api-server/pkg/jsonstore"
	"github.com/joincivil/civil-api-server/pkg/posts"
	"github.com/joincivil/civil-api-server/pkg/users"
	"github.com/joincivil/civil-api-server/pkg/utils"
)

// RunPersisterMigrations creates tables, indices, and migrations for persisters
func RunPersisterMigrations(userPersister *users.PostgresPersister) error {
	err := userPersister.CreateTables()
	if err != nil {
		return fmt.Errorf("Error creating tables: err: %v", err)
	}
	err = userPersister.CreateIndices()
	if err != nil {
		return fmt.Errorf("Error creating indices: err: %v", err)
	}
	err = userPersister.RunMigrations()
	if err != nil {
		return fmt.Errorf("Error running migrations: err: %v", err)
	}

	return nil
}

// RunPostPersisterMigrations creates views for persister
func RunPostPersisterMigrations(postPersister posts.PostPersister) error {
	err := (postPersister).CreateViews()
	if err != nil {
		return fmt.Errorf("Error creating views: err: %v", err)
	}
	return nil
}

func initJsonbPersister(config *utils.GraphQLConfig) (jsonstore.JsonbPersister, error) {
	jsonbPersister, err := jsonstore.NewPostgresPersister(
		config.PostgresAddress(),
		config.PostgresPort(),
		config.PostgresUser(),
		config.PostgresPw(),
		config.PostgresDbname(),
	)
	if err != nil {
		return nil, err
	}
	err = jsonbPersister.CreateTables()
	if err != nil {
		return nil, fmt.Errorf("Error creating tables: err: %v", err)
	}
	err = jsonbPersister.CreateIndices()
	if err != nil {
		return nil, fmt.Errorf("Error creating indices: err: %v", err)
	}
	err = jsonbPersister.RunMigrations()
	if err != nil {
		return nil, fmt.Errorf("Error running migrations: err: %v", err)
	}
	return jsonbPersister, nil
}

func initDiscourseListingMapPersister(db *gorm.DB) (discourse.ListingMapPersister, error) {
	listingMapPersister, err := discourse.NewPostgresPersister(db)
	if err != nil {
		return nil, err
	}
	return listingMapPersister, nil
}
