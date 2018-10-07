//go:generate gorunpkg github.com/99designs/gqlgen

package kyc

import (
	context "context"
	log "github.com/golang/glog"

	kyc "github.com/joincivil/civil-api-server/pkg/generated/kyc"
)

// NewResolver is a convenience function to init a Resolver struct
func NewResolver(onfidoAPI *OnfidoAPI, onfidoTokenReferrer string) *Resolver {
	if onfidoTokenReferrer == "" {
		onfidoTokenReferrer = DefaultTokenReferrer
	}
	return &Resolver{
		onfidoAPI:           onfidoAPI,
		onfidoTokenReferrer: onfidoTokenReferrer,
	}
}

// Resolver is the main resolver for the KYC GraphQL endpoint
type Resolver struct {
	onfidoAPI           *OnfidoAPI
	onfidoTokenReferrer string
}

// Mutation is the resolver for the Mutation type
func (r *Resolver) Mutation() kyc.MutationResolver {
	return &mutationResolver{r}
}

// Query is the resolver for the Query type
func (r *Resolver) Query() kyc.QueryResolver {
	return &queryResolver{r}
}

type mutationResolver struct{ *Resolver }

func (r *mutationResolver) KycCreateApplicant(ctx context.Context, applicant kyc.KycCreateApplicantInput) (*string, error) {
	newAddress := Address{}
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

	newApplicant := &Applicant{}
	newApplicant.Addresses = []Address{newAddress}
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

	returnedApplicant, err := r.onfidoAPI.CreateApplicant(newApplicant)
	if err != nil {
		log.Errorf("Error creating applicant on onfido: err: %v", err)
		return nil, err
	}

	return &returnedApplicant.ID, nil
}
func (r *mutationResolver) KycGenerateSdkToken(ctx context.Context, applicantID string) (*string, error) {
	token, err := r.onfidoAPI.GenerateSDKToken(applicantID, r.onfidoTokenReferrer)
	if err != nil {
		log.Errorf("Error creating generating new JWT on onfido: err: %v", err)
		return nil, err
	}

	return &token, err
}
func (r *mutationResolver) KycCreateCheck(ctx context.Context, applicantID string, facialVariant *string) (*string, error) {
	var rep *Report
	if facialVariant != nil && *facialVariant == ReportVariantFacialSimilarityVideo {
		rep = FacialSimilarityVideoReport
	} else {
		rep = FacialSimilarityStandardReport
	}
	newCheck := &Check{
		Type: CheckTypeExpress,
		Reports: []Report{
			// *kyc.IdentityKycReport,
			*DocumentReport,
			*rep,
			// *kyc.WatchlistKycReport,
		},
	}

	returnedCheck, err := r.onfidoAPI.CreateCheck(applicantID, newCheck)
	if err != nil {
		log.Errorf("Error creating check on onfido: err: %v", err)
		return nil, err
	}

	return &returnedCheck.ID, nil
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) KycApplicants(ctx context.Context) ([]*string, error) {
	notImplmented := "Not implemented"
	return []*string{&notImplmented}, nil
}
