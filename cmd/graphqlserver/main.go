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

	err = graphqlmain.RunServer(config)
	if err != nil {
		log.Errorf("Error starting graphql server: err: %v\n", err)
		os.Exit(2)
	}
}
