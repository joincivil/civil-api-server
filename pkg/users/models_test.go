package users_test

import (
	"testing"

	"github.com/joincivil/civil-api-server/pkg/users"
)

func TestGenerateUID(t *testing.T) {
	user := &users.User{}
	err := user.GenerateUID()
	if err != nil {
		t.Errorf("Should not have gotten error generating UID: err: %v", err)
	}

	err = user.GenerateUID()
	if err == nil {
		t.Errorf("Should have gotten error with existing UID: err: %v", err)
	}
}

func TestUserPurchaseTxHashes(t *testing.T) {
	user := &users.User{}

	if user.PurchaseTxHashesStr != "" {
		t.Fatalf("Should have had an empty tx hash list")
	}

	if len(user.PurchaseTxHashes()) > 0 {
		t.Fatalf("Should have had an empty tx hash list")
	}

	testTx := "0x161df03a04629bc6d8e5f1ad14489edf76d508e8c0a6bcb6a43b85cfaa226aa0"
	err := user.AddPurchaseTxHash(testTx)
	if err != nil {
		t.Errorf("Should have not gotten error for add tx hash: err: %v", err)
	}

	hashes := user.PurchaseTxHashes()
	if len(hashes) != 1 {
		t.Fatalf("Should not have had an empty tx hash list")
	}

	if hashes[0] != testTx {
		t.Fatalf("Should have had the correct txhash in list")
	}
}