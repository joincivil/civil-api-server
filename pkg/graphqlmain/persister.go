package graphqlmain

import (
	"fmt"

	"github.com/jinzhu/gorm"
	"github.com/joincivil/civil-api-server/pkg/discourse"

	"github.com/joincivil/civil-api-server/pkg/jsonstore"
	"github.com/joincivil/civil-api-server/pkg/posts"
	"github.com/joincivil/civil-api-server/pkg/users"
)

// RunJsonbPersisterMigrations creates tables for jsonb
func RunJsonbPersisterMigrations(postgresPersister *jsonstore.PostgresPersister) error {
	err := postgresPersister.CreateTables()
	if err != nil {
		return fmt.Errorf("Error creating tables: err: %v", err)
	}
	return nil
}

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

func initDiscourseListingMapPersister(db *gorm.DB) (discourse.ListingMapPersister, error) {
	listingMapPersister, err := discourse.NewPostgresPersister(db)
	if err != nil {
		return nil, err
	}
	return listingMapPersister, nil
}
