// +build integration

package nrsignup_test

import (
	"os"
	"testing"

	"github.com/joincivil/civil-api-server/pkg/auth"
	"github.com/joincivil/civil-api-server/pkg/jsonstore"
	"github.com/joincivil/civil-api-server/pkg/nrsignup"
	"github.com/joincivil/civil-api-server/pkg/users"

	"github.com/joincivil/go-common/pkg/email"
	"github.com/joincivil/go-common/pkg/eth"
)

const (
	sendGridKeyEnvVar = "SENDGRID_TEST_KEY"
	testEmailAddress  = "foo@bar.com"
	// testEmailAddress = "peter@civil.co"
	testSecret = "testsecret"

	useSandbox = true
	// useSandbox = false
)

func buildTestNewsroom() *nrsignup.Newsroom {
	member1 := &nrsignup.CharterRosterMember{
		Name:       "Michael Bluth",
		Role:       "Chief Banana Man",
		Bio:        "Part of the Bluth family",
		EthAddress: "thisissomehexaddress",
		SocialURLs: &nrsignup.CharterSocialURLs{
			Twitter: "http://twitter.com/mbluth",
		},
		AvatarURL: "",
		Signature: "",
	}

	member2 := &nrsignup.CharterRosterMember{
		Name:       "G.O.B",
		Role:       "Magic Man",
		Bio:        "Part of the Bluth family",
		EthAddress: "thisissomehexaddress2",
		SocialURLs: &nrsignup.CharterSocialURLs{
			Twitter: "http://twitter.com/thenotoriousgob",
		},
		AvatarURL: "",
		Signature: "",
	}

	roster := []*nrsignup.CharterRosterMember{member1, member2}

	social_urls := &nrsignup.CharterSocialURLs{
		Twitter: "http://twitter.com/testnewsroom",
	}
	signatures := []*nrsignup.CharterConstitutionSignature{
		{
			Signer:    "signervalue",
			Signature: "signaturevlaue",
			Message:   "messagevalue",
		},
	}

	testCharter := &nrsignup.Charter{
		LogoURL:     "https://logopond.com/logos/23a5bbd721473954cb21e7404c51c2b8.png",
		NewsroomURL: "http://civil.co",
		Tagline:     "All the news that's fit to print",
		Roster:      roster,
		Mission: &nrsignup.CharterMission{
			Purpose:       "To save the world",
			Structure:     "This is the structure",
			Revenue:       "We are going to make money",
			Encumbrances:  "But we can only talk about dogs",
			Miscellaneous: "",
		},
		SocialURLs: social_urls,
		Signatures: signatures,
	}

	testNewsroom := &nrsignup.Newsroom{
		Name:    "Test Newsroom",
		Charter: testCharter,
	}

	return testNewsroom
}

func buildUserService() *users.UserService {
	initUsers := map[string]*users.User{
		"1": {UID: "1", Email: testEmailAddress},
	}
	persister := &users.InMemoryUserPersister{UsersInMemory: initUsers}

	return users.NewUserService(persister, nil)
}

func buildJsonbService(t *testing.T) *jsonstore.Service {
	testNewsroom := buildTestNewsroom()
	grantRequested := false
	grantApproved := false
	jsonData := &nrsignup.SignupUserJSONData{
		OnboardedTs:        1547827105,
		Charter:            testNewsroom.Charter,
		CharterLastUpdated: 1547827105,
		GrantRequested:     &grantRequested,
		GrantApproved:      &grantApproved,
		NewsroomDeployTx:   "",
		NewsroomAddress:    "",
		TcrApplyTx:         "",
	}
	rawJson, err := jsonData.AsJSONStr()
	if err != nil {
		t.Errorf("Should have returned json data: err: %v", err)
		rawJson = ""
	}
	namespace := jsonstore.DefaultJsonbGraphqlNs
	ID := "nrsignup"
	salt := "1" // Faking the ID

	key, _ := jsonstore.NamespaceIDSaltHashKey(namespace, ID, salt)
	json := &jsonstore.JSONb{
		Key:     key,
		ID:      ID,
		RawJSON: rawJson,
	}
	initStore := map[string]*jsonstore.JSONb{
		key: json,
	}
	persister := &jsonstore.InMemoryJSONbPersister{
		Store: initStore,
	}
	return jsonstore.NewJsonbService(persister)
}

func newTestNewsroomSignupService(t *testing.T, sendGridKey string) (
	*nrsignup.Service, *jsonstore.Service, error) {
	jsonbService := buildJsonbService(t)
	signupService, err := nrsignup.NewNewsroomSignupService(
		&eth.Helper{},
		email.NewEmailerWithSandbox(sendGridKey, useSandbox),
		buildUserService(),
		jsonbService,
		auth.NewJwtTokenGenerator([]byte(testSecret)),
		"http://localhost:8080",
		"",
		"",
	)
	return signupService, jsonbService, err
}

func getSendGridKeyFromEnvVar() string {
	return os.Getenv(sendGridKeyEnvVar)
}

func TestSendWelcomeEmail(t *testing.T) {
	sendGridKey := getSendGridKeyFromEnvVar()
	if sendGridKey == "" {
		t.Log("No SENDGRID_TEST_KEY set, skipping test")
		return
	}

	signup, _, err := newTestNewsroomSignupService(t, sendGridKey)
	if err != nil {
		t.Fatalf("Error init signup service: err: %v", err)
	}

	err = signup.SendWelcomeEmail("1")
	if err != nil {
		t.Fatalf("Should not have failed to send to valid user: %v", err)
	}
}

func TestSendWelcomeEmailNoUser(t *testing.T) {
	sendGridKey := getSendGridKeyFromEnvVar()
	if sendGridKey == "" {
		t.Log("No SENDGRID_TEST_KEY set, skipping test")
		return
	}

	signup, _, err := newTestNewsroomSignupService(t, sendGridKey)
	if err != nil {
		t.Fatalf("Error init signup service: err: %v", err)
	}

	err = signup.SendWelcomeEmail("2")
	if err == nil {
		t.Fatalf("Should have return an error when sending to invalid user")
	}
}
func TestRequestGrant(t *testing.T) {
	sendGridKey := getSendGridKeyFromEnvVar()
	if sendGridKey == "" {
		t.Log("No SENDGRID_TEST_KEY set, skipping test")
		return
	}

	signup, jsonbService, err := newTestNewsroomSignupService(t, sendGridKey)
	if err != nil {
		t.Fatalf("Should have init signup service: err: %v", err)
	}

	err = signup.RequestGrant("1", true)
	if err != nil {
		t.Fatalf("Should not have error requesting grant: err: %v", err)
	}

	jsonbs, err := jsonbService.RetrieveJSONb(
		nrsignup.DefaultJsonbID,
		jsonstore.DefaultJsonbGraphqlNs,
		"1",
	)
	if err != nil {
		t.Fatalf("Should have retrieved jsonb: err: %v", err)
	}

	if len(jsonbs) != 1 {
		t.Fatalf("Should have returned one item")
	}

	jsonb := jsonbs[0]
	grantRequestedFieldFound := false
	for _, field := range jsonb.JSON {
		if field.Key == "grantRequested" {
			if !field.Value.Value.(bool) {
				t.Errorf("Should have set grantRequested to true")
			}
			grantRequestedFieldFound = true
			break
		}
	}

	if !grantRequestedFieldFound {
		t.Errorf("Should have found grantRequested field")
	}
}

func TestRequestGrantNoUser(t *testing.T) {
	sendGridKey := getSendGridKeyFromEnvVar()
	if sendGridKey == "" {
		t.Log("No SENDGRID_TEST_KEY set, skipping test")
		return
	}

	signup, _, err := newTestNewsroomSignupService(t, sendGridKey)
	if err != nil {
		t.Fatalf("Error init signup service: err: %v", err)
	}

	err = signup.RequestGrant("2", true)
	if err == nil {
		t.Fatalf("Should have failed to return with invalid user")
	}
}

func TestApproveGrant(t *testing.T) {
	sendGridKey := getSendGridKeyFromEnvVar()
	if sendGridKey == "" {
		t.Log("No SENDGRID_TEST_KEY set, skipping test")
		return
	}

	signup, jsonbService, err := newTestNewsroomSignupService(t, sendGridKey)
	if err != nil {
		t.Fatalf("Should have init signup service: err: %v", err)
	}

	err = signup.RequestGrant("1", true)
	if err != nil {
		t.Fatalf("Should not have error requesting grant: err: %v", err)
	}

	err = signup.ApproveGrant("1", true)
	if err != nil {
		t.Fatalf("Should not have error requesting grant: err: %v", err)
	}

	jsonbs, err := jsonbService.RetrieveJSONb(
		nrsignup.DefaultJsonbID,
		jsonstore.DefaultJsonbGraphqlNs,
		"1",
	)
	if err != nil {
		t.Fatalf("Should have retrieved jsonb: err: %v", err)
	}

	if len(jsonbs) != 1 {
		t.Fatalf("Should have returned one item")
	}

	jsonb := jsonbs[0]
	grantApprovedFieldFound := false
	for _, field := range jsonb.JSON {
		if field.Key == "grantApproved" {
			if !field.Value.Value.(bool) {
				t.Errorf("Should have set grantApproved to true")
			}
			grantApprovedFieldFound = true
			break
		}
	}

	if !grantApprovedFieldFound {
		t.Errorf("Should have found grantApproved field")
	}
}

func TestRejectGrant(t *testing.T) {
	sendGridKey := getSendGridKeyFromEnvVar()
	if sendGridKey == "" {
		t.Log("No SENDGRID_TEST_KEY set, skipping test")
		return
	}

	signup, jsonbService, err := newTestNewsroomSignupService(t, sendGridKey)
	if err != nil {
		t.Fatalf("Should have init signup service: err: %v", err)
	}

	err = signup.RequestGrant("1", true)
	if err != nil {
		t.Fatalf("Should not have error requesting grant: err: %v", err)
	}

	err = signup.ApproveGrant("1", false)
	if err != nil {
		t.Fatalf("Should not have error requesting grant: err: %v", err)
	}

	jsonbs, err := jsonbService.RetrieveJSONb(
		nrsignup.DefaultJsonbID,
		jsonstore.DefaultJsonbGraphqlNs,
		"1",
	)
	if err != nil {
		t.Fatalf("Should have retrieved jsonb: err: %v", err)
	}

	if len(jsonbs) != 1 {
		t.Fatalf("Should have returned one item")
	}

	jsonb := jsonbs[0]
	grantApprovedFieldFound := false
	for _, field := range jsonb.JSON {
		if field.Key == "grantApproved" {
			if field.Value.Value.(bool) {
				t.Errorf("Should have set grantApproved to false")
			}
			grantApprovedFieldFound = true
			break
		}
	}

	if !grantApprovedFieldFound {
		t.Errorf("Should have found grantApproved field")
	}
}
