package nrsignup

import (
	"encoding/json"

	"github.com/joincivil/go-common/pkg/newsroom"
)

// SignupUserJSONData represents the data being stored by the client into
// the JSON store.  To be unmarshaled/marshalled to/from a JSON string.
type SignupUserJSONData struct {
	OnboardedTs        int               `json:"onboardedTimestamp,omitempty"`
	Charter            *newsroom.Charter `json:"charter,omitempty"`
	CharterLastUpdated int               `json:"charterLastUpdated,omitempty"`
	GrantRequested     *bool             `json:"grantRequested,omitempty"`
	GrantApproved      *bool             `json:"grantApproved,omitempty"`
	NewsroomDeployTx   string            `json:"newsroomDeployTx,omitempty"`
	NewsroomAddress    string            `json:"newsroomAddress,omitempty"`
	NewsroomName       string            `json:"newsroomName,omitempty"`
	TcrApplyTx         string            `json:"tcrApplyTx,omitempty"`
}

// AsJSONStr is a convenience method to return this struct as a JSON string
func (s *SignupUserJSONData) AsJSONStr() (string, error) {
	bys, err := json.Marshal(s)
	if err != nil {
		return "", err
	}
	return string(bys), nil
}
