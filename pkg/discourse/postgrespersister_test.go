package discourse

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/joincivil/civil-api-server/pkg/testutils"

	cpersist "github.com/joincivil/go-common/pkg/persistence"
)

func setupDBConnection() (*PostgresPersister, error) {
	db, err := testutils.GetTestDBConnection()
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(&ListingMap{}).Error
	if err != nil {
		return nil, err
	}
	return NewPostgresPersister(db)
}

func setupTestTable() (*PostgresPersister, error) {
	persister, err := setupDBConnection()
	if err != nil {
		return persister, fmt.Errorf("Error connecting to DB: %v", err)
	}
	return persister, nil
}

func deleteTestTable(persister *PostgresPersister) error {
	return persister.db.DropTable(&ListingMap{}).Error
}

func TestSaveRetrieveListingMap(t *testing.T) {
	persister, err := setupTestTable()
	if err != nil {
		t.Errorf("Error connecting to DB: %v", err)
	}
	defer deleteTestTable(persister) // nolint: errcheck

	addr := "0x49fd8f1d3e6f88a4d08cd4a6e445f848e9475caf"
	cappedAddr := strings.ToUpper(addr)
	normalizedAddr := common.HexToAddress(addr).Hex()
	topicID := int64(1010101)

	// Try to retrieve from an empty table
	_, err = persister.RetrieveListingMap(addr)
	if err == nil {
		t.Errorf("Should have failed since nothing in table")
	}

	if err != cpersist.ErrPersisterNoResults {
		t.Errorf("Should have failed with no ldm found, not normal error: err: %v", err)
	}

	// Save a mapping
	ldm := &ListingMap{}
	ldm.ListingAddress = addr
	ldm.TopicID = topicID

	err = persister.SaveListingMap(ldm)
	if err != nil {
		t.Errorf("Should not have failed saving a mapping to the table: err: %v", err)
	}

	// Try retrieving the mapping
	ldm, err = persister.RetrieveListingMap(addr)
	if err != nil {
		t.Errorf("Should not have failed: err: %v", err)
	}

	if err == cpersist.ErrPersisterNoResults {
		t.Errorf("Should have not failed")
	}

	if ldm.ListingAddress != normalizedAddr {
		t.Errorf("Addresses do not match")
	}
	if ldm.TopicID != topicID {
		t.Errorf("Topic IDs don't match")
	}

	// Try retrieving the mapping using the capped addr
	ldm, err = persister.RetrieveListingMap(cappedAddr)
	if err != nil {
		t.Errorf("Should not have failed: err: %v", err)
	}

	if err == cpersist.ErrPersisterNoResults {
		t.Errorf("Should have not failed")
	}

	if ldm.ListingAddress != normalizedAddr {
		t.Errorf("Addresses do not match")
	}
	if ldm.TopicID != topicID {
		t.Errorf("Topic IDs don't match")
	}

	// Save the same address should not fail
	err = persister.SaveListingMap(ldm)
	if err != nil {
		t.Errorf("Should not have failed saving a mapping to the table: err: %v", err)
	}

	// Save the same address with a new topic ID
	ldm.TopicID = int64(100)
	err = persister.SaveListingMap(ldm)
	if err != nil {
		t.Errorf("Should not have failed saving a mapping to the table: err: %v", err)
	}

	ldm, err = persister.RetrieveListingMap(cappedAddr)
	if err != nil {
		t.Errorf("Should not have failed: err: %v", err)
	}

	if err == cpersist.ErrPersisterNoResults {
		t.Errorf("Should have not failed")
	}

	if ldm.ListingAddress != normalizedAddr {
		t.Errorf("Addresses do not match")
	}
	if ldm.TopicID != int64(100) {
		t.Errorf("Topic IDs don't match")
	}

	// Test to see the updated ts is working
	time.Sleep(1 * time.Second)
	err = persister.SaveListingMap(ldm)
	if err != nil {
		t.Errorf("Should not have failed saving a mapping to the table: err: %v", err)
	}

	ldm, err = persister.RetrieveListingMap(cappedAddr)
	if err != nil {
		t.Errorf("Should not have failed: err: %v", err)
	}
	if err == cpersist.ErrPersisterNoResults {
		t.Errorf("Should have not failed")
	}
	if ldm.CreatedAt == ldm.UpdatedAt {
		t.Errorf("Should have updated the updatedTs")
	}
}
