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

	svc := buildService()

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

func buildService() *auth.Service {
	persister := &testutils.InMemoryUserPersister{
		Users: map[string]*users.User{},
	}
	userService := users.NewUserService(persister)
	generator := auth.NewJwtTokenGenerator([]byte("secret"))
	return auth.NewAuthService(userService, generator, nil)
}
