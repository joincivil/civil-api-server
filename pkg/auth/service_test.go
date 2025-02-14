package auth_test

import (
	"encoding/hex"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/joincivil/civil-api-server/pkg/auth"
	"github.com/joincivil/civil-api-server/pkg/testutils"
	"github.com/joincivil/civil-api-server/pkg/users"
	"github.com/joincivil/civil-api-server/pkg/utils"

	"github.com/joincivil/go-common/pkg/email"
	"github.com/joincivil/go-common/pkg/eth"

	ctime "github.com/joincivil/go-common/pkg/time"
)

const (
	testEmail          = "PETER@civil.co"
	testEmailWeirdCase = "PeTeR@civil.co"

	testSignupLoginProtoHost = "http://localhost:8080"

	sendGridKeyEnvVar = "SENDGRID_TEST_KEY"

	useSandbox = true
)

var (
	defaultSignupTemplateIDs = map[string]string{
		"DEFAULT":    "d-88f731b52a524e6cafc308d0359b84a6",
		"NEWSROOM":   "d-88f731b52a524e6cafc308d0359b84a6",
		"STOREFRONT": "d-88f731b52a524e6cafc308d0359b84a6",
	}
	defaultLoginTemplateIDs = map[string]string{
		"DEFAULT":    "d-88f731b52a524e6cafc308d0359b84a6",
		"NEWSROOM":   "d-88f731b52a524e6cafc308d0359b84a6",
		"STOREFRONT": "d-88f731b52a524e6cafc308d0359b84a6",
	}
)

func getSendGridKeyFromEnvVar() string {
	return os.Getenv(sendGridKeyEnvVar)
}

func TestSignupEmailSendForApplication(t *testing.T) {
	sendGridKey := getSendGridKeyFromEnvVar()
	if sendGridKey == "" {
		t.Log("No SENDGRID_TEST_KEY set, skipping signup email test")
		return
	}

	service, err := buildService(sendGridKey)
	if err != nil {
		t.Fatalf("Should have return a valid auth service: err: %v", err)
	}

	result, token, err := service.SignupEmailSendForApplication(testEmail, auth.ApplicationEnumNewsroom)
	if err != nil {
		t.Errorf("Should have not gotten an error sending signup email: err: %v", err)
	}
	if result != "ok" {
		t.Errorf("Should have gotten an OK response")
	}

	if token == "" {
		t.Errorf("Should have gotten token")
	}

	resp, err := service.SignupEmailConfirm(token)
	if err != nil {
		t.Errorf("Should not have gotten an error when confirming: err: %v", err)
	}

	if resp.RefreshToken == "" {
		t.Errorf("Should have gotten a refresh token")
	}
	if resp.Token == "" {
		t.Errorf("Should have gotten an access token")
	}
	if resp.UID == "" {
		t.Errorf("Should have gotten a UID")
	}

	result, _, err = service.SignupEmailSendForApplication(testEmailWeirdCase, auth.ApplicationEnumNewsroom)
	if err != nil {
		t.Errorf("Should have not gotten an error sending signup email: err: %v", err)
	}
	if result != auth.EmailExistsResponse {
		t.Errorf("Should have gotten the email exist response")
	}
}
func TestSignupEmailSendForApplicationUserExists(t *testing.T) {
	sendGridKey := getSendGridKeyFromEnvVar()
	if sendGridKey == "" {
		t.Log("No SENDGRID_TEST_KEY set, skipping signup email test")
		return
	}

	service, err := buildServiceWithExistingUser(sendGridKey)
	if err != nil {
		t.Fatalf("Should have return a valid auth service: err: %v", err)
	}

	result, token, err := service.SignupEmailSendForApplication(testEmail, auth.ApplicationEnumNewsroom)
	if err != nil {
		t.Errorf("Should have not gotten an error sending signup email: err: %v", err)
	}

	if result != auth.EmailExistsResponse {
		t.Fatalf("Should have gotten an email exists response: result = %v", result)
	}

	if token != "" {
		t.Fatal("Should not have gotten token")
	}
}

func TestSignupEmailSend(t *testing.T) {
	sendGridKey := getSendGridKeyFromEnvVar()
	if sendGridKey == "" {
		t.Log("No SENDGRID_TEST_KEY set, skipping signup email test")
		return
	}

	service, err := buildService(sendGridKey)
	if err != nil {
		t.Fatalf("Should have return a valid auth service: err: %v", err)
	}

	result, token, err := service.SignupEmailSend(testEmail)
	if err != nil {
		t.Errorf("Should have not gotten an error sending signup email: err: %v", err)
	}
	if result != "ok" {
		t.Errorf("Should have gotten an OK response")
	}

	if token == "" {
		t.Errorf("Should have gotten token")
	}

	resp, err := service.SignupEmailConfirm(token)
	if err != nil {
		t.Errorf("Should not have gotten an error when confirming: err: %v", err)
	}

	if resp.RefreshToken == "" {
		t.Errorf("Should have gotten a refresh token")
	}
	if resp.Token == "" {
		t.Errorf("Should have gotten an access token")
	}
	if resp.UID == "" {
		t.Errorf("Should have gotten a UID")
	}
}

func TestLoginEmailSendForApplication(t *testing.T) {
	sendGridKey := getSendGridKeyFromEnvVar()
	if sendGridKey == "" {
		t.Log("No SENDGRID_TEST_KEY set, skipping login email test")
		return
	}

	service, err := buildServiceWithExistingUser(sendGridKey)
	if err != nil {
		t.Fatalf("Should have return a valid auth service: err: %v", err)
	}

	result, token, err := service.LoginEmailSendForApplication(testEmail, auth.ApplicationEnumNewsroom)
	if err != nil {
		t.Errorf("Should have not gotten an error sending login email: err: %v", err)
	}

	if result != "ok" {
		t.Fatalf("Should have gotten an OK response: result = %v", result)
	}

	if token == "" {
		t.Fatal("Should have gotten token")
	}

	resp, err := service.LoginEmailConfirm(token)
	if err != nil {
		t.Errorf("Should not have gotten an error when confirming: err: %v", err)
	}

	if resp.RefreshToken == "" {
		t.Errorf("Should have gotten a refresh token")
	}
	if resp.Token == "" {
		t.Errorf("Should have gotten an access token")
	}
	if resp.UID == "" {
		t.Errorf("Should have gotten a UID")
	}

	result, _, err = service.LoginEmailSendForApplication("nonexistent@email.com", auth.ApplicationEnumNewsroom)
	if err != nil {
		t.Errorf("Should have not gotten an error sending login email: err: %v", err)
	}
	if result != auth.EmailNotFoundResponse {
		t.Errorf("Should have gotten the email not found response")
	}

	// Test for a user that does exist to make sure there is an error and a new
	// user is not created.
	generator := utils.NewJwtTokenGenerator([]byte("secret"))
	noUserToken, err := generator.GenerateToken("nouser@civil.co", 60*3)
	if err != nil {
		t.Errorf("Should not have gotten generating no user token: err: %v", err)
	}
	resp, err = service.LoginEmailConfirm(noUserToken)
	if err == nil {
		t.Errorf("Should have gotten an error when confirming: err: %v", err)
	}
	if resp != nil {
		t.Errorf("Should have gotten a nil response")
	}
}

func TestLoginEmailSendForApplicationNoUser(t *testing.T) {
	sendGridKey := getSendGridKeyFromEnvVar()
	if sendGridKey == "" {
		t.Log("No SENDGRID_TEST_KEY set, skipping login email test")
		return
	}

	service, err := buildService(sendGridKey)
	if err != nil {
		t.Fatalf("Should have return a valid auth service: err: %v", err)
	}

	result, token, err := service.LoginEmailSendForApplication(testEmail, auth.ApplicationEnumNewsroom)
	if err != nil {
		t.Errorf("Should have not gotten an error sending login email: err: %v", err)
	}

	if result != auth.EmailNotFoundResponse {
		t.Fatalf("Should have gotten an email not found response: result = %v", result)
	}

	if token != "" {
		t.Fatal("Should not have gotten token")
	}
}

func TestSignupEmailSendNoEmailer(t *testing.T) {
	sendGridKey := getSendGridKeyFromEnvVar()
	if sendGridKey == "" {
		t.Log("No SENDGRID_TEST_KEY set, skipping signup email test")
		return
	}

	persister := &users.InMemoryUserPersister{
		UsersInMemory: map[string]*users.User{},
	}
	userService := users.NewUserService(persister, &testutils.ControllerUpdaterSpy{}, &testutils.MockChannelHelper{})
	generator := utils.NewJwtTokenGenerator([]byte("secret"))
	service, err := auth.NewAuthService(userService, generator, nil, defaultSignupTemplateIDs,
		defaultLoginTemplateIDs, testSignupLoginProtoHost, []string{}, []string{testSignupLoginProtoHost})
	if err != nil {
		t.Errorf("Should have not failed to create a new auth service: err: %v", err)
	}

	_, _, err = service.SignupEmailSendForApplication(testEmail, auth.ApplicationEnumNewsroom)
	if err == nil {
		t.Errorf("Should have gotten an error sending signup email")
	}
}

func TestSignupEmailSendNoProtoHost(t *testing.T) {
	sendGridKey := getSendGridKeyFromEnvVar()
	if sendGridKey == "" {
		t.Log("No SENDGRID_TEST_KEY set, skipping signup email test")
		return
	}

	persister := &users.InMemoryUserPersister{
		UsersInMemory: map[string]*users.User{},
	}
	userService := users.NewUserService(persister, &testutils.ControllerUpdaterSpy{}, &testutils.MockChannelHelper{})
	emailer := email.NewEmailerWithSandbox(sendGridKey, useSandbox)
	generator := utils.NewJwtTokenGenerator([]byte("secret"))
	service, err := auth.NewAuthService(userService, generator, emailer, defaultSignupTemplateIDs,
		defaultLoginTemplateIDs, "", []string{}, []string{})
	if err != nil {
		t.Errorf("Should have not failed to create a new auth service: err: %v", err)
	}

	_, _, err = service.SignupEmailSendForApplication(testEmail, auth.ApplicationEnumNewsroom)
	if err == nil {
		t.Errorf("Should have gotten an error sending signup email")
	}
}

func TestSignupEth(t *testing.T) {
	svc, err := buildService("")
	if err != nil {
		t.Fatalf("was not expecting an error %v", err)
	}

	pk, err := crypto.HexToECDSA("fad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d5817ac83d38b6a19")
	if err != nil {
		t.Fatalf("was not expecting an error %v", err)
	}

	msg := "Authenticate to http://localhost:8080 @ " + time.Now().Format(time.RFC3339)
	ethMsg := eth.AsEthereumSignature(msg)
	hash := crypto.Keccak256Hash([]byte(ethMsg))

	signature, err := crypto.Sign(hash.Bytes(), pk)
	if err != nil {
		t.Fatalf("was not expecting an error %v", err)
	}

	address := eth.GetEthAddressFromPrivateKey(pk).Hex()

	input := &users.SignatureInput{
		Message:     msg,
		MessageHash: hash.String(),
		Signature:   "0x" + hex.EncodeToString(signature),
		Signer:      address,
	}
	_, err = svc.SignupEth(input)

	if err != nil {
		t.Fatalf("was not expecting an error %v", err)
	}
}

func TestLoginEth(t *testing.T) {
	pk, err := crypto.HexToECDSA("fad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d5817ac83d38b6a19")
	if err != nil {
		t.Fatalf("was not expecting an error %v", err)
	}

	msg := "Authenticate to http://localhost:8080 @ " + time.Now().Format(time.RFC3339)
	ethMsg := eth.AsEthereumSignature(msg)
	hash := crypto.Keccak256Hash([]byte(ethMsg))

	signature, err := crypto.Sign(hash.Bytes(), pk)
	if err != nil {
		t.Fatalf("was not expecting an error %v", err)
	}

	address := eth.GetEthAddressFromPrivateKey(pk).Hex()

	user1 := &users.User{
		Email:       testEmail,
		EthAddress:  address,
		DateCreated: ctime.CurrentEpochSecsInInt64(),
		DateUpdated: ctime.CurrentEpochSecsInInt64(),
	}
	user1.GenerateUID() // nolint: errcheck

	persister := &users.InMemoryUserPersister{
		UsersInMemory: map[string]*users.User{
			user1.UID: user1,
		},
	}
	userService := users.NewUserService(persister, &testutils.ControllerUpdaterSpy{}, &testutils.MockChannelHelper{})
	generator := utils.NewJwtTokenGenerator([]byte("secret"))
	emailer := email.NewEmailerWithSandbox("", useSandbox)
	svc, err := auth.NewAuthService(userService, generator, emailer, defaultSignupTemplateIDs,
		defaultLoginTemplateIDs, testSignupLoginProtoHost, []string{}, []string{testSignupLoginProtoHost})
	if err != nil {
		t.Errorf("Should have not failed to create a new auth service: err: %v", err)
	}

	input := &users.SignatureInput{
		Message:     msg,
		MessageHash: hash.String(),
		Signature:   "0x" + hex.EncodeToString(signature),
		Signer:      address,
	}
	_, err = svc.LoginEth(input)

	if err != nil {
		t.Fatalf("was not expecting an error %v", err)
	}
}

func TestLoginEthNoUser(t *testing.T) {
	pk, err := crypto.HexToECDSA("fad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d5817ac83d38b6a19")
	if err != nil {
		t.Fatalf("was not expecting an error %v", err)
	}

	msg := "Log in to Civil @ " + time.Now().Format(time.RFC3339)
	hash := crypto.Keccak256Hash([]byte(msg))

	signature, err := crypto.Sign(hash.Bytes(), pk)
	if err != nil {
		t.Fatalf("was not expecting an error %v", err)
	}

	address := eth.GetEthAddressFromPrivateKey(pk).Hex()

	svc, err := buildService("")
	if err != nil {
		t.Errorf("Should have not failed to create a new auth service: err: %v", err)
	}

	input := &users.SignatureInput{
		Message:     msg,
		MessageHash: hash.String(),
		Signature:   "0x" + hex.EncodeToString(signature),
		Signer:      address,
	}
	_, err = svc.LoginEth(input)

	if err == nil {
		t.Fatalf("Should have return error with no user")
	}
}

func TestConfigTemplateIDs(t *testing.T) {
	signupTemplateIDs := map[string]string{
		"DEFAULT":    "d-88f731b52a524e6cafc308d0359b84a6",
		"NEWSROOM":   "d-88f731b52a524e6cafc308d0359b84a6",
		"STOREFRONT": "d-88f731b52a524e6cafc308d0359b84a6",
	}
	loginTemplateIDs := map[string]string{
		"DEFAULT":    "d-88f731b52a524e6cafc308d0359b84a6",
		"NEWSROOM":   "d-88f731b52a524e6cafc308d0359b84a6",
		"STOREFRONT": "d-88f731b52a524e6cafc308d0359b84a6",
	}

	persister := &users.InMemoryUserPersister{
		UsersInMemory: map[string]*users.User{},
	}
	userService := users.NewUserService(persister, &testutils.ControllerUpdaterSpy{}, &testutils.MockChannelHelper{})
	generator := utils.NewJwtTokenGenerator([]byte("secret"))
	_, err := auth.NewAuthService(userService, generator, nil, signupTemplateIDs,
		loginTemplateIDs, testSignupLoginProtoHost, []string{}, []string{testSignupLoginProtoHost})
	if err != nil {
		t.Errorf("Should not have failed to create a new auth service: err: %v", err)
	}
}

func TestBadConfigTemplateIDsAppName(t *testing.T) {
	signupTemplateIDs := map[string]string{
		"DEFAULT":    "d-88f731b52a524e6cafc308d0359b84a6",
		"NEWSROOM":   "d-88f731b52a524e6cafc308d0359b84a6",
		"STOREFRONT": "d-88f731b52a524e6cafc308d0359b84a6",
		"BADAPPNAME": "d-88f731b52a524e6cafc308d0359b84a6",
	}
	loginTemplateIDs := map[string]string{
		"DEFAULT":    "d-88f731b52a524e6cafc308d0359b84a6",
		"NEWSROOM":   "d-88f731b52a524e6cafc308d0359b84a6",
		"STOREFRONT": "d-88f731b52a524e6cafc308d0359b84a6",
	}

	persister := &users.InMemoryUserPersister{
		UsersInMemory: map[string]*users.User{},
	}
	userService := users.NewUserService(persister, &testutils.ControllerUpdaterSpy{}, &testutils.MockChannelHelper{})
	generator := utils.NewJwtTokenGenerator([]byte("secret"))
	_, err := auth.NewAuthService(userService, generator, nil, signupTemplateIDs,
		loginTemplateIDs, testSignupLoginProtoHost, []string{}, []string{testSignupLoginProtoHost})
	if err == nil {
		t.Errorf("Should have failed to create a new auth service: err: %v", err)
	}
}

func TestBadConfigTemplateIDsAppNameCase(t *testing.T) {
	signupTemplateIDs := map[string]string{
		"DEFAULT":    "d-88f731b52a524e6cafc308d0359b84a6",
		"NEWSROOM":   "d-88f731b52a524e6cafc308d0359b84a6",
		"storefront": "d-88f731b52a524e6cafc308d0359b84a6",
	}
	loginTemplateIDs := map[string]string{
		"DEFAULT":    "d-88f731b52a524e6cafc308d0359b84a6",
		"NEWSROOM":   "d-88f731b52a524e6cafc308d0359b84a6",
		"STOREFRONT": "d-88f731b52a524e6cafc308d0359b84a6",
	}

	persister := &users.InMemoryUserPersister{
		UsersInMemory: map[string]*users.User{},
	}
	userService := users.NewUserService(persister, &testutils.ControllerUpdaterSpy{}, &testutils.MockChannelHelper{})
	generator := utils.NewJwtTokenGenerator([]byte("secret"))
	_, err := auth.NewAuthService(userService, generator, nil, signupTemplateIDs,
		loginTemplateIDs, testSignupLoginProtoHost, []string{}, []string{testSignupLoginProtoHost})
	if err == nil {
		t.Errorf("Should have failed to create a new auth service: err: %v", err)
	}
}

func TestBadConfigTemplateIDs(t *testing.T) {
	signupTemplateIDs := map[string]string{
		"DEFAULT":    "d-88f731b52a524e6cafc308d0359b84a6",
		"NEWSROOM":   "d-88f731b52a524e6cafc308d0359b84a6",
		"STOREFRONT": "",
	}
	loginTemplateIDs := map[string]string{
		"DEFAULT":    "d-88f731b52a524e6cafc308d0359b84a6",
		"NEWSROOM":   "d-88f731b52a524e6cafc308d0359b84a6",
		"STOREFRONT": "d-88f731b52a524e6cafc308d0359b84a6",
	}

	persister := &users.InMemoryUserPersister{
		UsersInMemory: map[string]*users.User{},
	}
	userService := users.NewUserService(persister, &testutils.ControllerUpdaterSpy{}, &testutils.MockChannelHelper{})
	generator := utils.NewJwtTokenGenerator([]byte("secret"))
	_, err := auth.NewAuthService(userService, generator, nil, signupTemplateIDs,
		loginTemplateIDs, testSignupLoginProtoHost, []string{}, []string{testSignupLoginProtoHost})
	if err == nil {
		t.Errorf("Should have failed to create a new auth service: err: %v", err)
	}
}

func TestConfigTemplateIDsNotAll(t *testing.T) {
	signupTemplateIDs := map[string]string{
		"DEFAULT":  "d-88f731b52a524e6cafc308d0359b84a6",
		"NEWSROOM": "d-88f731b52a524e6cafc308d0359b84a6",
	}
	loginTemplateIDs := map[string]string{
		"DEFAULT":    "d-88f731b52a524e6cafc308d0359b84a6",
		"STOREFRONT": "d-88f731b52a524e6cafc308d0359b84a6",
	}

	persister := &users.InMemoryUserPersister{
		UsersInMemory: map[string]*users.User{},
	}
	userService := users.NewUserService(persister, &testutils.ControllerUpdaterSpy{}, &testutils.MockChannelHelper{})
	generator := utils.NewJwtTokenGenerator([]byte("secret"))
	_, err := auth.NewAuthService(userService, generator, nil,
		signupTemplateIDs, loginTemplateIDs, testSignupLoginProtoHost, []string{}, []string{testSignupLoginProtoHost})
	if err != nil {
		t.Errorf("Should not have failed to create a new auth service: err: %v", err)
	}
}

func TestSignupTemplateIDFromApplication(t *testing.T) {
	signupTemplateIDs := map[string]string{
		"DEFAULT": "d-88f731b52a524e6cafc308d0359b84a6",
	}
	loginTemplateIDs := map[string]string{
		"DEFAULT": "d-88f731b52a524e6cafc308d0359b84a6",
	}

	persister := &users.InMemoryUserPersister{
		UsersInMemory: map[string]*users.User{},
	}
	userService := users.NewUserService(persister, &testutils.ControllerUpdaterSpy{}, &testutils.MockChannelHelper{})
	generator := utils.NewJwtTokenGenerator([]byte("secret"))
	service, err := auth.NewAuthService(userService, generator, nil,
		signupTemplateIDs, loginTemplateIDs, testSignupLoginProtoHost, []string{}, []string{testSignupLoginProtoHost})
	if err != nil {
		t.Errorf("Should not have failed to create a new auth service: err: %v", err)
	}

	templateID, err := service.SignupTemplateIDForApplication(auth.ApplicationEnumDefault)
	if err != nil {
		t.Errorf("Should not have failed to retrieve a templateID: err: %v", err)
	}
	if templateID == "" {
		t.Errorf("Should not returned an empty templateID")
	}

	templateID, err = service.SignupTemplateIDForApplication(auth.ApplicationEnumNewsroom)
	if err == nil {
		t.Errorf("Should have failed to retrieve a templateID: err: %v", err)
	}
	if templateID != "" {
		t.Errorf("Should have returned an empty templateID")
	}

	templateID, err = service.LoginTemplateIDForApplication(auth.ApplicationEnumDefault)
	if err != nil {
		t.Errorf("Should not have failed to retrieve a templateID: err: %v", err)
	}
	if templateID == "" {
		t.Errorf("Should not returned an empty templateID")
	}

	templateID, err = service.LoginTemplateIDForApplication(auth.ApplicationEnumNewsroom)
	if err == nil {
		t.Errorf("Should have failed to retrieve a templateID: err: %v", err)
	}
	if templateID != "" {
		t.Errorf("Should have returned an empty templateID")
	}
}

func buildService(sendGridKey string) (*auth.Service, error) {
	persister := &users.InMemoryUserPersister{
		UsersInMemory: map[string]*users.User{},
	}
	userService := users.NewUserService(persister, &testutils.ControllerUpdaterSpy{}, &testutils.MockChannelHelper{})
	emailer := email.NewEmailerWithSandbox(sendGridKey, useSandbox)
	generator := utils.NewJwtTokenGenerator([]byte("secret"))
	return auth.NewAuthService(userService, generator, emailer, defaultSignupTemplateIDs,
		defaultLoginTemplateIDs, testSignupLoginProtoHost, []string{}, []string{testSignupLoginProtoHost})
}

func buildServiceWithExistingUser(sendGridKey string) (*auth.Service, error) {
	user1 := &users.User{
		Email:       strings.ToLower(testEmail),
		EthAddress:  "0x5385A3a9a1468b7D900A93E6f21E903E30928765",
		DateCreated: ctime.CurrentEpochSecsInInt64(),
		DateUpdated: ctime.CurrentEpochSecsInInt64(),
	}
	user1.GenerateUID() // nolint: errcheck

	persister := &users.InMemoryUserPersister{
		UsersInMemory: map[string]*users.User{
			user1.UID: user1,
		},
	}
	userService := users.NewUserService(persister, &testutils.ControllerUpdaterSpy{}, &testutils.MockChannelHelper{})
	generator := utils.NewJwtTokenGenerator([]byte("secret"))
	emailer := email.NewEmailerWithSandbox(sendGridKey, useSandbox)
	return auth.NewAuthService(userService, generator, emailer, defaultSignupTemplateIDs,
		defaultLoginTemplateIDs, testSignupLoginProtoHost, []string{}, []string{testSignupLoginProtoHost})

}

func TestCheckAuthDomain(t *testing.T) {
	var domains = []string{"http://localhost:3000", "https://staging.civil.app"}
	var domaintests = []struct {
		input   string
		domains []string
		result  string
	}{
		{"Authenticate to http://localhost:3000", domains, "http://localhost:3000"},
		{"Authenticate to https://staging.civil.app", domains, "https://staging.civil.app"},
		{"Authenticate to https://foo.bar", domains, ""},
	}

	for _, tt := range domaintests {
		domain, err := auth.CheckAuthDomain(tt.input, tt.domains)
		if tt.result == "" && err != nil {
			if domain != "" {
				t.Errorf("expecting an error but got a result: %v", domain)
			}
			if err.Error() != "cannot authenticate to domain" {
				t.Errorf("expecting auth failure: %v", domain)
			}
		} else if err != nil {
			t.Errorf("error: %v", err)
		} else if domain != tt.result {
			t.Errorf("expecting: %v but got: %v", tt.result, domain)
		}

	}
}
