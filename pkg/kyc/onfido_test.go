// build +integration

package kyc_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/joincivil/civil-api-server/pkg/kyc"
)

const (
	sandboxAPIKey = "test_2wKCChuc-QVWXdtM1bht4Lj3QmpLoloW"
)

func TestSandboxCreateApplicantGenTokenCheck(t *testing.T) {
	onfido := kyc.NewOnfidoAPI(
		kyc.ProdAPIURL,
		sandboxAPIKey,
	)
	address1 := kyc.Address{
		FlatNumber:     "4F",
		BuildingNumber: "100",
		Street:         "42nd Street",
		Town:           "New York",
		State:          "NY",
		Postcode:       "10011",
		Country:        "USA",
		StartDate:      "2007-01-01",
		EndDate:        "2010-01-01",
	}
	address2 := kyc.Address{
		FlatNumber:     "201",
		BuildingNumber: "10",
		Street:         "13th Street",
		Town:           "New York",
		State:          "NY",
		Postcode:       "10001",
		Country:        "USA",
		StartDate:      "2010-01-02",
	}
	newApplicant := &kyc.Applicant{
		Title:     "Mr",
		FirstName: "John",
		LastName:  "Smith",
		Email:     "peter@civil.co",
		Gender:    kyc.ApplicantGenderMale,
		Dob:       "1978-01-01",
		Telephone: "2125551212",
		Country:   "USA",
		Addresses: []kyc.Address{address1, address2},
	}
	returnedApplicant, err := onfido.CreateApplicant(newApplicant)
	if err != nil {
		t.Fatalf("Should have not failed when creating new applicant: err: %v", err)
	}
	if returnedApplicant.FirstName != newApplicant.FirstName {
		t.Errorf("Should have returned a different first name: err: %v", err)
	}
	if returnedApplicant.LastName != newApplicant.LastName {
		t.Errorf("Should have returned a different last name: err: %v", err)
	}
	if returnedApplicant.Gender != newApplicant.Gender {
		t.Errorf("Should have returned a different first name: err: %v", err)
	}
	bys, err := json.Marshal(returnedApplicant)
	if err != nil {
		t.Fatalf("Should have unmarshalled into valid JSON: err: %v", err)
	}
	t.Logf("returned applicant: %v", string(bys))

	token, err := onfido.GenerateSDKToken(returnedApplicant.ID, "https://*.civil.co/*")
	if err != nil {
		t.Fatalf("Should have returned a valid token: err: %v", err)
	}
	t.Logf("returned token: %v", token)

	if strings.Count(token, ".") != 2 {
		t.Errorf("Does not look like valid token: err: %v", err)
	}

	// XXX(PN): Bc we arent uploading the document/facial stuff, this will fail.
	// newCheck := &kyc.Check{
	// 	Type: kyc.CheckTypeExpress,
	// 	Reports: []kyc.Report{
	// 		// *kyc.IdentityKycReport,
	// 		*kyc.DocumentReport,
	// 		*kyc.FacialSimilarityStandardReport,
	// 		// *kyc.WatchlistKycReport,
	// 	},
	// }

	// returnedCheck, err := onfido.CreateCheck(returnedApplicant.ID, newCheck)
	// if err != nil {
	// 	t.Errorf("Should not have had error returning new check: err: %v", err)
	// }

	// bys, err = json.Marshal(returnedCheck)
	// if err != nil {
	// 	t.Fatalf("Should have unmarshalled into valid JSON: err: %v", err)
	// }
	// t.Logf("returned check: %v", string(bys))

}
