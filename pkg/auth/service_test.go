package auth_test

import (
	"encoding/hex"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/joincivil/civil-api-server/pkg/auth"
	"github.com/joincivil/civil-api-server/pkg/testutils"
	"github.com/joincivil/civil-api-server/pkg/users"
	"github.com/joincivil/go-common/pkg/eth"
)

func TestSignupEth(t *testing.T) {

	svc, err := buildService()
	if err != nil {
		t.Fatalf("was not expecting an error %v", err)
	}

	pk, err := crypto.HexToECDSA("fad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d5817ac83d38b6a19")
	if err != nil {
		t.Fatalf("was not expecting an error %v", err)
	}

	msg := "Sign up with Civil @ " + time.Now().Format(time.RFC3339)
	hash := crypto.Keccak256Hash([]byte(msg))

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

	persister := &testutils.InMemoryUserPersister{
		Users: map[string]*users.User{},
	}
	userService := users.NewUserService(persister)
	generator := auth.NewJwtTokenGenerator([]byte("secret"))
	_, err := auth.NewAuthService(userService, generator, nil, signupTemplateIDs, loginTemplateIDs)
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

	persister := &testutils.InMemoryUserPersister{
		Users: map[string]*users.User{},
	}
	userService := users.NewUserService(persister)
	generator := auth.NewJwtTokenGenerator([]byte("secret"))
	_, err := auth.NewAuthService(userService, generator, nil, signupTemplateIDs, loginTemplateIDs)
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

	persister := &testutils.InMemoryUserPersister{
		Users: map[string]*users.User{},
	}
	userService := users.NewUserService(persister)
	generator := auth.NewJwtTokenGenerator([]byte("secret"))
	_, err := auth.NewAuthService(userService, generator, nil, signupTemplateIDs, loginTemplateIDs)
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

	persister := &testutils.InMemoryUserPersister{
		Users: map[string]*users.User{},
	}
	userService := users.NewUserService(persister)
	generator := auth.NewJwtTokenGenerator([]byte("secret"))
	_, err := auth.NewAuthService(userService, generator, nil, signupTemplateIDs, loginTemplateIDs)
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

	persister := &testutils.InMemoryUserPersister{
		Users: map[string]*users.User{},
	}
	userService := users.NewUserService(persister)
	generator := auth.NewJwtTokenGenerator([]byte("secret"))
	_, err := auth.NewAuthService(userService, generator, nil, signupTemplateIDs, loginTemplateIDs)
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

	persister := &testutils.InMemoryUserPersister{
		Users: map[string]*users.User{},
	}
	userService := users.NewUserService(persister)
	generator := auth.NewJwtTokenGenerator([]byte("secret"))
	service, err := auth.NewAuthService(userService, generator, nil, signupTemplateIDs, loginTemplateIDs)
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

func buildService() (*auth.Service, error) {
	persister := &testutils.InMemoryUserPersister{
		Users: map[string]*users.User{},
	}
	userService := users.NewUserService(persister)
	generator := auth.NewJwtTokenGenerator([]byte("secret"))
	return auth.NewAuthService(userService, generator, nil, nil, nil)
}
