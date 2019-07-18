package testruntime

import (
	log "github.com/golang/glog"
	"github.com/jinzhu/gorm"
	"github.com/joincivil/civil-api-server/pkg/channels"
	"github.com/joincivil/civil-api-server/pkg/payments"
	"github.com/joincivil/civil-api-server/pkg/posts"

	// load postgres specific dialect
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

// CleanDatabase deletes data
func CleanDatabase(db *gorm.DB) error {
	models := []interface{}{
		&channels.Channel{},
		&channels.ChannelMember{},
		&posts.PostModel{},
		&payments.PaymentModel{},
	}
	if err := db.AutoMigrate(models...).Error; err != nil {
		log.Errorf("error in migration: %v", err)
		return err
	}

	if err := db.Unscoped().Delete(&channels.Channel{}).Delete(&channels.ChannelMember{}).Delete(&posts.PostModel{}).Delete(&payments.PaymentModel{}).Error; err != nil {
		log.Errorf("error deleting data: %v", err)
		return err
	}

	return nil
}