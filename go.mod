module github.com/joincivil/civil-api-server

go 1.12

require (
	github.com/99designs/gqlgen v0.9.3
	github.com/DATA-DOG/go-sqlmock v1.3.3 // indirect
	github.com/Jeffail/gabs v1.4.0 // indirect
	github.com/Jeffail/tunny v0.0.0-20181108205650-4921fff29480
	github.com/PuerkitoBio/goquery v1.5.0 // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/didip/tollbooth v4.0.0+incompatible
	github.com/didip/tollbooth_chi v0.0.0-20170928041846-6ab5f3083f3d
	github.com/dyatlov/go-htmlinfo v0.0.0-20180517114536-d9417c75de65
	github.com/dyatlov/go-oembed v0.0.0-20180429203341-4bc5ab7a42e9 // indirect
	github.com/dyatlov/go-opengraph v0.0.0-20180429202543-816b6608b3c8 // indirect
	github.com/dyatlov/go-readability v0.0.0-20150926130635-e7b2080f87f8 // indirect
	github.com/ethereum/go-ethereum v0.0.0-20190528221609-008d250e3c57
	github.com/go-chi/chi v4.0.2+incompatible
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/iancoleman/strcase v0.0.0-20180726023541-3605ed457bf7
	github.com/ipfs/go-ipfs-api v0.0.1
	github.com/jinzhu/gorm v1.9.8
	github.com/jmoiron/sqlx v0.0.0-20180614180643-0dae4fefe7c0
	github.com/joincivil/civil-events-processor v0.0.0-20191118141640-ebf7214a173a
	github.com/joincivil/go-common v0.0.0-20191121182548-7994d712fce3
	github.com/kelseyhightower/envconfig v1.3.0
	github.com/lib/pq v1.1.0
	github.com/nfnt/resize v0.0.0-20180221191011-83c6a9932646
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pilagod/gorm-cursor-paginator v0.1.0
	github.com/pkg/errors v0.8.1
	github.com/rs/cors v1.6.0
	github.com/satori/go.uuid v0.0.0-20180103174451-36e9d2ebbde5
	github.com/stripe/stripe-go v61.0.1+incompatible
	github.com/vektah/gqlparser v1.1.2
	github.com/vincent-petithory/dataurl v0.0.0-20160330182126-9a301d65acbb
	go.uber.org/atomic v1.4.0 // indirect
	go.uber.org/dig v1.7.0 // indirect
	go.uber.org/fx v1.9.0
	go.uber.org/goleak v0.10.0 // indirect
	go.uber.org/multierr v1.1.0 // indirect
	golang.org/x/crypto v0.0.0-20191002192127-34f69633bfdc // indirect
	golang.org/x/net v0.0.0-20191009170851-d66e71096ffb // indirect
	golang.org/x/sys v0.0.0-20191009170203-06d7bd2c5f4f // indirect
)

replace git.apache.org/thrift.git v0.12.0 => github.com/apache/thrift v0.12.0
