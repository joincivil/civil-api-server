// +build integration

package nrsignup_test

import (
	"os"
	"strings"
	"testing"

	"github.com/joincivil/civil-api-server/pkg/auth"
	"github.com/joincivil/civil-api-server/pkg/jsonstore"
	"github.com/joincivil/civil-api-server/pkg/nrsignup"
	"github.com/joincivil/civil-api-server/pkg/testutils"
	"github.com/joincivil/civil-api-server/pkg/users"

	"github.com/joincivil/go-common/pkg/email"
	ctime "github.com/joincivil/go-common/pkg/time"
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
	persister := &testutils.InMemoryUserPersister{UsersInMemory: initUsers}

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
	persister := &testutils.InMemoryJSONbPersister{
		Store: initStore,
	}
	return jsonstore.NewJsonbService(persister)
}

func newTestNewsroomSignupService(t *testing.T, sendGridKey string) (
	*nrsignup.Service, *jsonstore.Service, *users.UserService, error) {
	jsonbService := buildJsonbService(t)
	userService := buildUserService()
	signupService, err := nrsignup.NewNewsroomSignupService(
		nil,
		email.NewEmailerWithSandbox(sendGridKey, useSandbox),
		userService,
		jsonbService,
		auth.NewJwtTokenGenerator([]byte(testSecret)),
		"http://localhost:8080",
		"",
		"",
	)
	return signupService, jsonbService, userService, err
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

	signup, _, _, err := newTestNewsroomSignupService(t, sendGridKey)
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

	signup, _, _, err := newTestNewsroomSignupService(t, sendGridKey)
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

	signup, jsonbService, _, err := newTestNewsroomSignupService(t, sendGridKey)
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

	signup, _, _, err := newTestNewsroomSignupService(t, sendGridKey)
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

	signup, jsonbService, _, err := newTestNewsroomSignupService(t, sendGridKey)
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

	signup, jsonbService, _, err := newTestNewsroomSignupService(t, sendGridKey)
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

func TestUpdateCharter(t *testing.T) {
	sendGridKey := getSendGridKeyFromEnvVar()
	if sendGridKey == "" {
		t.Log("No SENDGRID_TEST_KEY set, skipping test")
		return
	}

	signup, _, _, err := newTestNewsroomSignupService(t, sendGridKey)
	if err != nil {
		t.Fatalf("Error init signup service: err: %v", err)
	}

	newName := "This is a new name"

	testNewsroom := buildTestNewsroom()
	updatedCharter := nrsignup.Charter{
		Name:        newName,
		LogoURL:     testNewsroom.Charter.LogoURL,
		NewsroomURL: testNewsroom.Charter.NewsroomURL,
		Tagline:     testNewsroom.Charter.Tagline,
		Roster:      testNewsroom.Charter.Roster,
		Signatures:  testNewsroom.Charter.Signatures,
		Mission:     testNewsroom.Charter.Mission,
		SocialURLs:  testNewsroom.Charter.SocialURLs,
	}

	err = signup.UpdateCharter("1", updatedCharter)
	if err != nil {
		t.Errorf("Error updating charter: %v", err)
	}

	data, err := signup.RetrieveUserJSONData("1")
	if err != nil {
		t.Errorf("Error retrieving json data: %v", err)
	}

	if data.Charter.Name != newName {
		t.Errorf("Names should have matched with new name")
	}
}

func TestUpdateUserSteps(t *testing.T) {
	sendGridKey := getSendGridKeyFromEnvVar()
	if sendGridKey == "" {
		t.Log("No SENDGRID_TEST_KEY set, skipping test")
		return
	}

	signup, _, _, err := newTestNewsroomSignupService(t, sendGridKey)
	if err != nil {
		t.Fatalf("Error init signup service: err: %v", err)
	}

	step := 3
	furthestStep := 9
	lastSeen := ctime.CurrentEpochSecsInInt()

	err = signup.UpdateUserSteps("1", &step, &furthestStep, &lastSeen)
	if err != nil {
		t.Errorf("Should not have gotten error updating user steps: err: %v", err)
	}

	// XXX(PN): This needs to be fixed since our testutils persister doesn't
	// copy objects, it is causing some pointer/same data issues
	step = 13
	furthestStep = 13
	lastSeen = ctime.CurrentEpochSecsInInt()

	err = signup.UpdateUserSteps("1", &step, &furthestStep, &lastSeen)
	if err != nil {
		t.Errorf("Should not have gotten error updating user steps: err: %v", err)
	}
}

func TestSaveNewsroomAddress(t *testing.T) {
	sendGridKey := getSendGridKeyFromEnvVar()
	if sendGridKey == "" {
		t.Log("No SENDGRID_TEST_KEY set, skipping test")
		return
	}

	signup, _, userService, err := newTestNewsroomSignupService(t, sendGridKey)
	if err != nil {
		t.Fatalf("Error init signup service: err: %v", err)
	}

	newAddress := "0x02a8d60444d8aacc1b6c2fcefdca318af1cc5aed"

	err = signup.SaveNewsroomAddress("1", newAddress)
	if err != nil {
		t.Errorf("Should not have returned error saving address: err: %v", err)
	}

	user, err := userService.MaybeGetUser(users.UserCriteria{
		UID: "1",
	})
	if err != nil {
		t.Errorf("Should not have returned retrieving user: err: %v", err)
	}

	if len(user.AssocNewsoomAddr) != 1 {
		t.Errorf("Should have only returned 1 address")
	}

	found := false
	for _, addr := range user.AssocNewsoomAddr {
		if strings.ToLower(addr) == strings.ToLower(newAddress) {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Should have added the listing address")
	}

	// Trying to save the same address
	err = signup.SaveNewsroomAddress("1", newAddress)
	if err != nil {
		t.Errorf("Should not have returned error saving address: err: %v", err)
	}
	// Should remain 1 addr since it's the same
	if len(user.AssocNewsoomAddr) != 1 {
		t.Errorf("Should have only returned 1 address")
	}
}
