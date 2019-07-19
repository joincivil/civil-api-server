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

// RetrieveListingMaps retrieves a list of populated ListingMap structs from
// a slice of listing addresses. Returns the list in the same order as the
// given listing addresses. If mapping does not exist, will return as nil in th e
// list.
func (p *PostgresPersister) RetrieveListingMaps(listingAddresses []string) (
	[]*ListingMap, error) {
	for ind, addr := range listingAddresses {
		listingAddresses[ind] = ceth.NormalizeEthAddress(addr)
	}

	// Get ldms from store
	var storedLdms []*ListingMap
	err := p.db.Where(listingAddresses).Find(&storedLdms).Error
	if err != nil {
		return nil, err
	}

	// Build a lookup map
	storedLdmMap := make(map[string]*ListingMap, len(storedLdms))
	for _, ldm := range storedLdms {
		addr := ldm.ListingAddressAsAddr().Hex()
		storedLdmMap[addr] = ldm
	}

	// Ensure order and set missing ldms as nil
	ldms := make([]*ListingMap, len(listingAddresses))
	for ind, addr := range listingAddresses {
		ldm, ok := storedLdmMap[addr]
		if ok {
			ldms[ind] = ldm
		} else {
			ldms[ind] = nil
		}
	}

	return ldms, nil
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
