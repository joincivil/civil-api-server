module github.com/joincivil/civil-api-server

go 1.12

require (
	github.com/99designs/gqlgen v0.9.3
	github.com/DATA-DOG/go-sqlmock v1.3.3 // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/didip/tollbooth v4.0.0+incompatible
	github.com/didip/tollbooth_chi v0.0.0-20170928041846-6ab5f3083f3d
	github.com/ethereum/go-ethereum v0.0.0-20190528221609-008d250e3c57
	github.com/go-chi/chi v3.3.2+incompatible
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/goware/urlx v0.3.1
	github.com/iancoleman/strcase v0.0.0-20180726023541-3605ed457bf7
	github.com/ipfs/go-ipfs-api v0.0.1
	github.com/jinzhu/gorm v1.9.8
	github.com/jmoiron/sqlx v0.0.0-20180614180643-0dae4fefe7c0
	github.com/joincivil/civil-events-processor v0.0.0-20190919192235-874a802ea511
	github.com/joincivil/go-common v0.0.0-20190820182313-639fb94bf980
	github.com/kelseyhightower/envconfig v1.3.0
	github.com/kr/pty v1.1.8 // indirect
	github.com/lib/pq v1.1.0
	github.com/myitcv/gobin v0.0.13 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pilagod/gorm-cursor-paginator v0.1.0
	github.com/pkg/errors v0.8.1
	github.com/rs/cors v1.6.0
	github.com/satori/go.uuid v0.0.0-20180103174451-36e9d2ebbde5
	github.com/stripe/stripe-go v61.0.1+incompatible
	github.com/vektah/gorunpkg v0.0.0-20190126024156-2aeb42363e48 // indirect
	github.com/vektah/gqlparser v1.1.2
	go.uber.org/atomic v1.4.0 // indirect
	go.uber.org/dig v1.7.0 // indirect
	go.uber.org/fx v1.9.0
	go.uber.org/goleak v0.10.0 // indirect
	go.uber.org/multierr v1.1.0 // indirect
	golang.org/x/crypto v0.0.0-20190923035154-9ee001bba392 // indirect
	golang.org/x/net v0.0.0-20190923162816-aa69164e4478 // indirect
	golang.org/x/tools v0.0.0-20190923165424-71c3ad9cb704 // indirect
)

replace git.apache.org/thrift.git v0.12.0 => github.com/apache/thrift v0.12.0
