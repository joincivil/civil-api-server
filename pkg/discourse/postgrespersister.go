package discourse

import (
	"github.com/jinzhu/gorm"

	ceth "github.com/joincivil/go-common/pkg/eth"
	cpersist "github.com/joincivil/go-common/pkg/persistence"
)

// NewPostgresPersister creates a new postgres persister instance
func NewPostgresPersister(db *gorm.DB) (*PostgresPersister, error) {
	pgPersister := &PostgresPersister{
		db: db,
	}
	return pgPersister, nil
}

// PostgresPersister implements the persister for Postgresql
type PostgresPersister struct {
	db *gorm.DB
}

// RetrieveListingMap retrieves a populated  ListingMap struct
func (p *PostgresPersister) RetrieveListingMap(listingAddress string) (
	*ListingMap, error) {
	addr := ceth.NormalizeEthAddress(listingAddress)

	ldm := &ListingMap{}

	err := p.db.Where(&ListingMap{ListingAddress: addr}).First(&ldm).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, cpersist.ErrPersisterNoResults
		}
		return nil, err
	}

	return ldm, nil
}

// SaveListingMap saves a populated ListingMap
func (p *PostgresPersister) SaveListingMap(ldm *ListingMap) error {
	ldm.ListingAddress = ceth.NormalizeEthAddress(ldm.ListingAddress)

	updated := &ListingMap{}
	// XXX(PN): This appears to duplicate parts of the where clause
	err := p.db.Where(&ListingMap{ListingAddress: ldm.ListingAddress}).
		Assign(&ListingMap{TopicID: ldm.TopicID}).
		FirstOrCreate(updated).Error
	if err != nil {
		return err
	}

	return nil
}
