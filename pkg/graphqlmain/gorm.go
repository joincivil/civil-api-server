package graphqlmain

import (
	"fmt"
	"time"

	log "github.com/golang/glog"

	"github.com/joincivil/civil-api-server/pkg/discourse"

	"github.com/jinzhu/gorm"
	"github.com/joincivil/civil-api-server/pkg/channels"
	"github.com/joincivil/civil-api-server/pkg/payments"
	"github.com/joincivil/civil-api-server/pkg/posts"
	"github.com/joincivil/civil-api-server/pkg/utils"
)

const (
	maxOpenConns    = 10
	maxIdleConns    = 5
	connMaxLifetime = time.Second * 1800 // 30 mins
)

// NewGorm initializes a new gorm instance and runs migrations
func NewGorm(config *utils.GraphQLConfig) (*gorm.DB, error) {

	db, err := gorm.Open("postgres", fmt.Sprintf("host=%v port=%v user=%v dbname=%v password=%v sslmode=disable",
		config.PostgresAddress(),
		config.PostgresPort(),
		config.PostgresUser(),
		config.PostgresDbname(),
		config.PostgresPw(),
	))

	db.DB().SetMaxIdleConns(maxIdleConns)
	db.DB().SetMaxOpenConns(maxOpenConns)
	db.DB().SetConnMaxLifetime(connMaxLifetime)
	db.LogMode(config.Debug)

	amErr := db.AutoMigrate(
		&posts.PostModel{},
		&payments.PaymentModel{},
		&channels.Channel{},
		&channels.ChannelMember{},
	).Error
	if amErr != nil {
		log.Errorf("automigration error: %v", amErr)
	}

	amErr = db.AutoMigrate(
		&discourse.ListingMap{},
	).Error
	if amErr != nil {
		log.Errorf("automigration error: %v", amErr)
	}

	return db, err
}
