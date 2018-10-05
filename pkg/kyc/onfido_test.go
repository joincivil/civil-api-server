package kyc_test

import (
	"testing"

	"github.com/joincivil/civil-api-server/pkg/kyc"
)

func TestPostJSONFromStruct(t *testing.T) {
	address := kyc.Address{
		Street:   "68 3rd Avenue",
		Town:     "Brooklyn",
		State:    "NY",
		Postcode: "10000",
		Country:  "USA",
	}
	addresses := []kyc.Address{address}
	applicant := &kyc.Applicant{
		Title:     "Mr.",
		FirstName: "Firsty",
		LastName:  "Lastly",
		Addresses: addresses,
	}
	jsonData, err := applicant.EncodeToJSON()
	if err != nil {
		t.Errorf("Should have not gotten an error encoding to JSON string: err: %v", err)
	}
	t.Logf("jsonData = %v\n", jsonData)
}
