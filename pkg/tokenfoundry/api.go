package tokenfoundry

import (
	"encoding/json"
	"net/http"
	"net/url"

	chttp "github.com/joincivil/go-common/pkg/http"
)

// API allows you to interact with the TokenFoundry REST API
type API struct {
	rest *chttp.RestHelper
}

// NewAPI constructs a new instance of an API
func NewAPI(baseURL string, username string, password string) *API {
	authHeader := chttp.BuildBasicAuthHeader(username, password)
	rest := chttp.NewRestHelper(baseURL, authHeader)

	return &API{rest: rest}
}

// KYCStatusResult is the response from /api/temp/projects/civil/kyc/ (without the wrapper)
// example: {"email":"foo@gmail.com","eth_address":null,"found":true,"passed":true,"error":null}
type KYCStatusResult struct {
	Email      string `json:"email"`
	EthAddress string `json:"eth_address"`
	Found      bool   `json:"found"`
	Passed     bool   `json:"passed"`
	Error      string `json:"error"`
}

// GetKYCStatus determines if the an email is registered to buy CVL on TokenFoundry
func (api *API) GetKYCStatus(email string) (bool, error) {

	params := &url.Values{}
	params.Set("emails", email)
	rspBodyData, err := api.rest.SendRequest("api/temp/projects/civil/kyc/", http.MethodGet, params, nil)
	if err != nil {
		return false, err
	}

	wrapper := &struct {
		Count int               `json:"count"`
		Data  []KYCStatusResult `json:"data"`
	}{}
	err = json.Unmarshal(rspBodyData, wrapper)
	if err != nil {
		return false, err
	}

	if len(wrapper.Data) != 1 {
		return false, nil
	}

	user := wrapper.Data[0]

	return user.Passed, nil
}
