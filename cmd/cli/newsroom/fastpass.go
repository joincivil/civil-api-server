package main

import (
	"fmt"
	shell "github.com/ipfs/go-ipfs-api"
	"github.com/joincivil/civil-api-server/pkg/graphqlmain"
	"github.com/joincivil/civil-api-server/pkg/newsrooms"
	"github.com/joincivil/civil-api-server/pkg/runtime"
	"github.com/joincivil/civil-api-server/pkg/tokencontroller"
	"github.com/joincivil/civil-api-server/pkg/utils"
	"github.com/joincivil/go-common/pkg/email"
	"go.uber.org/fx"
	"os"
)

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		fmt.Println("must pass newsroomOwnerUID as command line argument")
		os.Exit(1)
	}

	app := fx.New(
		runtime.Module,
		fx.Provide(
			graphqlmain.NewGorm,
			graphqlmain.BuildConfig,
			tokencontroller.NewService,
			func(config *utils.GraphQLConfig) *utils.JwtTokenGenerator {
				return utils.NewJwtTokenGenerator([]byte(config.JwtSecret))
			},
			func() *shell.Shell {
				return shell.NewShell("https://ipfs.infura.io:5001")
			},
			func(config *utils.GraphQLConfig) *email.Emailer {
				return email.NewEmailer(config.SendgridKey)
			},
		),
		fx.Invoke(func(tools *newsrooms.Tools, config *utils.GraphQLConfig) {
			newsroomOwnerID := args[0]
			fmt.Println("starting Fastpass for user: " + newsroomOwnerID)

			fmt.Printf("DB Host: %v\n", config.PersisterPostgresAddress)
			fmt.Printf("Ethereum: %v\n", config.EthAPIURL)
			fmt.Printf("Rescue Address: %v\n", config.FastPassRescueMultisig.String())
			fmt.Printf("Application Tokens: %v\n", config.TcrApplicationTokens)

			updates := make(chan string)

			go func() {
				err := tools.FastPassNewsroom(updates, newsroomOwnerID)
				if err != nil {
					fmt.Printf("error with fastpass: %v", err)
					os.Exit(1)
				}
				close(updates)
			}()

			for update := range updates {
				fmt.Printf("update: %v\n", update)
			}

			os.Exit(0)
		}),
	)

	app.Run()

}
