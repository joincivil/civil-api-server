package nrsignup

import (
	"encoding/json"
)

// Newsroom represents data about a newsroom, including the charter
type Newsroom struct {
	Name    string   `json:"name"`
	Charter *Charter `json:"charter"`
}

// Charter represents charter data for a newsroom, mirrors structure from the
// FE client store as defined here
// https://github.com/joincivil/Civil/blob/master/packages/core/src/types.ts#L73-L87
type Charter struct {
	LogoURL     string                          `json:"logoUrl"`
	NewsroomURL string                          `json:"newsroomUrl"`
	Tagline     string                          `json:"tagline"`
	Roster      []*CharterRosterMember          `json:"roster"`
	Signatures  []*CharterConstitutionSignature `json:"signatures"`
	Mission     *CharterMission                 `json:"mission"`
	SocialURLs  []*CharterSocialURL             `json:"socialUrls"`
}

// CharterMission represents mission statements for a charter
type CharterMission struct {
	Purpose       string `json:"purpose"`
	Structure     string `json:"structure"`
	Revenue       string `json:"revenue"`
	Encumbrances  string `json:"encumbrances"`
	Miscellaneous string `json:"miscellaneous"`
}

// AsMap converts the CharterMission to a map
func (c *CharterMission) AsMap() map[string]interface{} {
	mission := map[string]interface{}{}
	mission["purpose"] = c.Purpose
	mission["structure"] = c.Structure
	mission["revenue"] = c.Revenue
	mission["encumbrances"] = c.Encumbrances
	mission["misc"] = c.Miscellaneous
	return mission
}

// CharterRosterMember represents a member of a newsroom roster
type CharterRosterMember struct {
	Name       string              `json:"name"`
	Role       string              `json:"role"`
	Bio        string              `json:"bio"`
	EthAddress string              `json:"ethAddress"`
	SocialURLs []*CharterSocialURL `json:"socialUrls"`
	AvatarURL  string              `json:"avatarUrl"`
	Signature  string              `json:"signature"`
}

// AsMap converts the CharterRosterMember to a map
func (c *CharterRosterMember) AsMap() map[string]interface{} {
	member := map[string]interface{}{}
	member["name"] = c.Name
	member["role"] = c.Role
	member["bio"] = c.Bio
	member["ethAddress"] = c.EthAddress
	member["avatarUrl"] = c.AvatarURL
	member["signature"] = c.Signature

	socials := []map[string]interface{}{}
	for _, social := range c.SocialURLs {
		socials = append(socials, social.AsMap())
	}
	member["socialUrls"] = socials

	return member
}

// CharterConstitutionSignature represents the signing of the constitution for a
// newsroom
type CharterConstitutionSignature struct {
	Signer    string `json:"signer"`
	Signature string `json:"signature"`
	Message   string `json:"message"`
}

// AsMap converts the CharterConstitutionSignature to a map
func (c *CharterConstitutionSignature) AsMap() map[string]interface{} {
	member := map[string]interface{}{}
	member["signer"] = c.Signer
	member["signature"] = c.Signature
	member["message"] = c.Message
	return member
}

// CharterSocialURL represents a social URL in the charter
type CharterSocialURL struct {
	Service string `json:"service"`
	URL     string `json:"url"`
}

// AsMap converts the CharterSocialURL to a map
func (c *CharterSocialURL) AsMap() map[string]interface{} {
	social := map[string]interface{}{}
	social["service"] = c.Service
	social["url"] = c.URL
	return social
}

// SignupUserJSONData represents the data being stored by the client into
// the JSON store.  To be unmarshaled/marshalled to/from a JSON string.
// TODO(PN): Ensure this is in sync with the client team otherwise we f-ed.
type SignupUserJSONData struct {
	WalletAddress      string   `json:"walletAddress"`
	Email              string   `json:"email"`
	OnboardedTs        int      `json:"onboardedTimestamp"`
	Charter            *Charter `json:"charter"`
	CharterLastUpdated int      `json:"charterLastUpdated"`
	GrantRequested     bool     `json:"grantRequested"`
	GrantApproved      bool     `json:"grantApproved"`
	NewsroomDeployTx   string   `json:"newsroomDeployTx"`
	NewsroomAddress    string   `json:"newsroomAddress"`
	TcrApplyTx         string   `json:"tcrApplyTx"`
}

// AsJSONStr is a convenience method to return this struct as a JSON string
func (s *SignupUserJSONData) AsJSONStr() (string, error) {
	bys, err := json.Marshal(s)
	if err != nil {
		return "", err
	}
	return string(bys), nil
}
