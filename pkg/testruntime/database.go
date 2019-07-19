package testruntime

import (
	"github.com/jinzhu/gorm"
	"github.com/joincivil/civil-api-server/pkg/channels"
	"github.com/joincivil/civil-api-server/pkg/payments"
	"github.com/joincivil/civil-api-server/pkg/posts"
	"github.com/pkg/errors"

	// load postgres specific dialect
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

// RunMigrations runs table migrations
func RunMigrations(db *gorm.DB) error {

	models := []interface{}{
		&channels.Channel{},
		&channels.ChannelMember{},
		&posts.PostModel{},
		&payments.PaymentModel{},
	}

	for _, model := range models {
		if err := db.AutoMigrate(model).Error; err != nil {
			return errors.Wrap(err, "error in migration")
		}
	}

	return nil
}
