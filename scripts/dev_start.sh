# run this from the root directory
make build
go run ./cmd/graphqlserver/main.go -logtostderr=true -stderrthreshold=INFO -v=2
