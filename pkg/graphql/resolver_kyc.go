package graphql

import (
	context "context"
	"fmt"

	"github.com/joincivil/civil-api-server/pkg/auth"
	"github.com/joincivil/civil-api-server/pkg/generated/graphql"
	"github.com/joincivil/civil-api-server/pkg/kyc"
	"github.com/joincivil/civil-api-server/pkg/users"
)

func (r *mutationResolver) KycCreateApplicant(ctx context.Context, applicant graphql.KycCreateApplicantInput) (*string, error) {
	token := auth.ForContext(ctx)
	if token == nil {
		return nil, fmt.Errorf("Access denied")
	}

	newAddress := kyc.Address{}
	if applicant.AptNumber != nil {
		newAddress.FlatNumber = *applicant.AptNumber
	}
	if applicant.BuildingNumber != nil {
		newAddress.BuildingNumber = *applicant.BuildingNumber
	}
	if applicant.Street != nil {
		newAddress.Street = *applicant.Street
	}
	if applicant.City != nil {
		newAddress.Town = *applicant.City
	}
	if applicant.State != nil {
		newAddress.State = *applicant.State
	}
	if applicant.Zipcode != nil {
		newAddress.Postcode = *applicant.Zipcode
	}
	if applicant.CountryOfResidence != nil {
		newAddress.Country = *applicant.CountryOfResidence
	}

	newApplicant := &kyc.Applicant{}
	newApplicant.Addresses = []kyc.Address{newAddress}
	newApplicant.FirstName = applicant.FirstName
	newApplicant.LastName = applicant.LastName

	if applicant.MiddleName != nil {
		newApplicant.MiddleName = *applicant.MiddleName
	}
	if applicant.Email != nil {
		newApplicant.Email = *applicant.Email
	}
	if applicant.DateOfBirth != nil {
		newApplicant.Dob = *applicant.DateOfBirth
	}
	if applicant.CountryOfResidence != nil {
		newApplicant.Country = *applicant.CountryOfResidence
	}

	// Not using UserService here bc we want to check for user before
	// we create an applicant.
	user, err := r.userPersister.User(&users.UserCriteria{Email: token.Sub})
	if err != nil {
		return nil, err
	}

	returnedApplicant, err := r.onfidoAPI.CreateApplicant(newApplicant)
	if err != nil {
		return nil, err
	}

	user.OnfidoApplicantID = returnedApplicant.ID
	updatedFields := []string{"OnfidoApplicantID"}
	err = r.userPersister.UpdateUser(user, updatedFields)
	if err != nil {
		return nil, err
	}

	return &returnedApplicant.ID, nil
}

func (r *mutationResolver) KycGenerateSdkToken(ctx context.Context, applicantID string) (*string, error) {
	token, err := r.onfidoAPI.GenerateSDKToken(applicantID, r.onfidoTokenReferrer)
	if err != nil {
		return nil, err
	}

	return &token, err
}

func (r *mutationResolver) KycCreateCheck(ctx context.Context, applicantID string, facialVariant *string) (*string, error) {
	var rep *kyc.Report
	if facialVariant != nil && *facialVariant == kyc.ReportVariantFacialSimilarityVideo {
		rep = kyc.FacialSimilarityVideoReport
	} else {
		rep = kyc.FacialSimilarityStandardReport
	}
	newCheck := &kyc.Check{
		Type: kyc.CheckTypeExpress,
		Reports: []kyc.Report{
			// *kyc.IdentityKycReport,
			*kyc.DocumentReport,
			*rep,
			// *kyc.WatchlistKycReport,
		},
	}

	returnedCheck, err := r.onfidoAPI.CreateCheck(applicantID, newCheck)
	if err != nil {
		return nil, err
	}

	return &returnedCheck.ID, nil
}