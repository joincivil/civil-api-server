package graphqlmain

import (
	"fmt"

	"github.com/joincivil/civil-api-server/pkg/invoicing"
	"github.com/joincivil/civil-api-server/pkg/jsonstore"
	"github.com/joincivil/civil-api-server/pkg/users"
	"github.com/joincivil/civil-api-server/pkg/utils"
)

func initInvoicePersister(config *utils.GraphQLConfig) (*invoicing.PostgresPersister, error) {
	persister, err := invoicing.NewPostgresPersister(
		config.PostgresAddress(),
		config.PostgresPort(),
		config.PostgresUser(),
		config.PostgresPw(),
		config.PostgresDbname(),
	)
	if err != nil {
		return nil, err
	}
	err = persister.CreateTables()
	if err != nil {
		return nil, fmt.Errorf("Error creating tables: err: %v", err)
	}
	err = persister.CreateIndices()
	if err != nil {
		return nil, fmt.Errorf("Error creating indices: err: %v", err)
	}
	return persister, nil
}

func initUserPersister(config *utils.GraphQLConfig) (*users.PostgresPersister, error) {
	persister, err := users.NewPostgresPersister(
		config.PostgresAddress(),
		config.PostgresPort(),
		config.PostgresUser(),
		config.PostgresPw(),
		config.PostgresDbname(),
	)
	if err != nil {
		return nil, err
	}
	err = persister.CreateTables()
	if err != nil {
		return nil, fmt.Errorf("Error creating tables: err: %v", err)
	}
	err = persister.CreateIndices()
	if err != nil {
		return nil, fmt.Errorf("Error creating indices: err: %v", err)
	}
	err = persister.RunMigrations()
	if err != nil {
		return nil, fmt.Errorf("Error running migrations: err: %v", err)
	}
	return persister, nil
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
