package main

import (
	"flag"
	"os"

	log "github.com/golang/glog"

	"github.com/joho/godotenv"
	"github.com/joincivil/civil-api-server/pkg/graphqlmain"
	"github.com/joincivil/civil-api-server/pkg/utils"
)

func main() {
	config := &utils.GraphQLConfig{}
	flag.Usage = func() {
		config.OutputUsage()
		os.Exit(0)
	}
	flag.Parse()

	env := os.Getenv("GRAPHQL_ENV")
	if "" == env {
		env = "development"
	}

	err := godotenv.Load(".env." + env + ".local")
	if err != nil {
		log.Errorf("Did not load .env.%v.local", env)
	}
	if "test" != env {
		err := godotenv.Load(".env.local")
		if err != nil {
			log.Errorf("Did not load .env.local")
		}
	}
	err = godotenv.Load(".env." + env)
	if err != nil {
		log.Errorf("Did not load .env." + env)
	}
	err = godotenv.Load()
	if err != nil {
		log.Errorf("Did not load .env")
	}

	err = config.PopulateFromEnv()
	if err != nil {
		config.OutputUsage()
		log.Errorf("Invalid graphql config: err: %v\n", err)
		os.Exit(2)
	}

	quit := make(chan bool)

	// Setup the kill hook
	graphqlmain.SetupKillNotify(quit)

	// Starts up the events workers
	err = graphqlmain.RunTokenEventsWorkers(config, quit)
	if err != nil {
		log.Errorf("Error starting token events workers: err: %v\n", err)
	}

	// Starts up the GraphQL/API server
	err = graphqlmain.RunServer(config)
	if err != nil {
		log.Errorf("Error starting graphql server: err: %v\n", err)
		os.Exit(2)
	}
}
