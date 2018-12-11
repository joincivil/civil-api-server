# run this from the root directory
make build
GRAPHQL_PORT=8080 \
GRAPHQL_DEBUG=true \
GRAPHQL_ETHEREUM_RPC_ADDRESS=https://rinkeby.infura.io \
GRAPHQL_PERSISTER_TYPE_NAME=postgresql \
GRAPHQL_PERSISTER_POSTGRES_ADDRESS=127.0.0.1 \
GRAPHQL_PERSISTER_POSTGRES_PORT=5432 \
GRAPHQL_PERSISTER_POSTGRES_DBNAME=civil_crawler \
GRAPHQL_PERSISTER_POSTGRES_USER=docker \
GRAPHQL_PERSISTER_POSTGRES_PW=docker \
GRAPHQL_JWT_SECRET=civiliscool \
go run ./cmd/graphqlserver/main.go -logtostderr=true -stderrthreshold=INFO -v=2
