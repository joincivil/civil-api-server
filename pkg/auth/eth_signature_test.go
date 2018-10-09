package auth_test

import (
	"testing"
	"time"

	"github.com/joincivil/civil-api-server/pkg/auth"
)

func TestVerifySignature(t *testing.T) {

	const address = "0xddB9e9957452d0E39A5E43Fd1AB4aE818aecC6aD"
	const message = "Login to Civil @ 2018-01-09T20:08:57Z"
	const signature = "0x520c1f6a0f1f968db5aaa39c08055bf2bd33dc9162d0237423549d31e91b6c661aa171e475cca20e1f0347685eaca6a0e443ecf5de3f53fb88dbb006ade5fc001b"

	var result, err = auth.VerifyEthSignature(address, message, signature)

	if err != nil {
		t.Fatalf("error thrown: %s", err)
	}
	if !result {
		t.Errorf("signature was not verified")
	}
}

func TestVerifyChallengeMalformed(t *testing.T) {

	const prefix = "Log in to Civil"
	const challenge = "Invalid prefix @ 2018-01-09T20:08:57Z"

	var err = auth.VerifyEthChallenge(prefix, 100, challenge)

	if err == nil {
		t.Fatalf("challenge was verified when it should not have been")
	} else if err.Error() != "challenge does not start with `Log in to Civil`" {
		t.Fatalf("did not expect this error message: " + err.Error())
	}
}

func TestVerifyChallengeExpired(t *testing.T) {

	const prefix = "Log in to Civil"
	const challenge = "Log in to Civil @ 2018-01-09T20:08:57Z"

	var err = auth.VerifyEthChallenge(prefix, 100, challenge)

	if err == nil {
		t.Fatalf("challenge was verified when it should not have been")
	} else if err.Error() != "expired" {
		t.Fatalf("did not expect this error message: " + err.Error())
	}
}

func LoginChallenge() string {
	return "Log in to Civil @ " + time.Now().Format(time.RFC3339)
}

func TestVerifyChallengeValid(t *testing.T) {

	const prefix = "Log in to Civil"
	challenge := LoginChallenge()

	var err = auth.VerifyEthChallenge(prefix, 100, challenge)

	if err != nil {
		t.Fatalf("error thrown: %s", err)
	}
}
