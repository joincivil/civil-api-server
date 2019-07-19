package payments

import (
	"time"

	log "github.com/golang/glog"
)

// PaymentUpdaterCron updates ethereum payments on a regular interval
func PaymentUpdaterCron(service *Service) {

	ticker := time.NewTicker(30 * time.Second)
	go func() {
		for range ticker.C {
			err := service.UpdateEtherPayments()
			if err != nil {
				log.Errorf("error updating payments: %v", err)
			}
		}
	}()
}
