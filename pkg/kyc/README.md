# KYC 

Provides endpoints to handle the KYC process through Onfido (http://onfido.com)

## GraphQL

### Mutations

`kycCreateApplicant`

Mutation to handle creating an Onfido applicant.  Needs all relevant personal data for verification. 

`kycGenerateSdkToken`

Mutation to retrieve a new JWT token for use by the SDK.  Needs the applicant ID.

`kycCreateCheck`

Mutation to handle starting an Onfido check.  Need the applicant ID and the facial variant.

## REST

`POST /v1/kyc/cb`

Callback for Onfido when a check state has been updated (i.e. a check is completed, cancelled, etc.)

## Testing the KYC SDK

```
cd pkg/kyc

# python 2.7
python -m SimpleHTTPServer
```

Need to manually create an applicant via `kycCreateApplicant` or in the Onfido console. Retain the applicant id. Go to [http://localhost:8000/test_kyc.html?applicantID=\<applicantID\>](http://localhost:8000/test_kyc.html?applicantID=\<applicantID\>) in browser.
