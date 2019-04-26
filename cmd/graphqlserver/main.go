package main

import (
	"flag"
	"os"

	log "github.com/golang/glog"

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

	err := config.PopulateFromEnv()
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
		os.Exit(2)
	}

	// Starts up the GraphQL/API server
	err = graphqlmain.RunServer(config)
	if err != nil {
		log.Errorf("Error starting graphql server: err: %v\n", err)
		os.Exit(2)
	}
}
