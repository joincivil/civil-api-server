package storefront_test

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/joincivil/civil-api-server/pkg/storefront"

	"github.com/joincivil/civil-api-server/pkg/testutils"
	"github.com/joincivil/civil-api-server/pkg/users"
)

type TestServiceEmailLists struct {
	Test *testing.T
}

func (t *TestServiceEmailLists) PurchaseCompleteAddToMembersList(user *users.User) {
	if user == nil {
		t.Test.Errorf("Should have gotten a populated user struct")
	}
}

func (t *TestServiceEmailLists) PurchaseCancelRemoveFromAbandonedList(user *users.User) {
	if user == nil {
		t.Test.Errorf("Should have gotten a populated user struct")
	}
}

func buildUserService(emailAddress string) *users.UserService {
	initUsers := map[string]*users.User{
		"1234": {UID: "1234", Email: emailAddress},
	}
	persister := &testutils.InMemoryUserPersister{Users: initUsers}

	return users.NewUserService(persister, &testutils.ControllerUpdaterSpy{})
}

func TestPurchaseTransactionComplete(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	testEmail := fmt.Sprintf("testuser%d@civil.co", rand.Intn(500))

	userService := buildUserService(testEmail)
	emailLists := &TestServiceEmailLists{Test: t}

	service, err := storefront.NewService("", nil, nil, userService, emailLists)
	if err != nil {
		t.Errorf("Should not have gotten error init new service: err: %v", err)
	}

	testTxHash := "0x6f73ffd316cfe4b7e9db1ca50bd9ea316cffc60c87d240eee9d9693e26d0381b"

	err = service.PurchaseTransactionComplete("1234", testTxHash)
	if err != nil {
		t.Errorf("Should not have gotten error calling on complete: err: %v", err)
	}

	criteria := users.UserCriteria{UID: "1234"}
	user, err := userService.GetUser(criteria)
	if err != nil {
		t.Fatalf("Should not have gotten error getting user: err: %v", err)
	}

	if user.PurchaseTxHashesStr == "" {
		t.Fatal("Should have set the tx hashes string")
	}

	if len(user.PurchaseTxHashes()) <= 0 {
		t.Fatal("Should have set the tx hashes string")
	}

	if user.PurchaseTxHashes()[0] != testTxHash {
		t.Fatal("Should have set the tx hashes string to the one given")
	}
}

func TestPurchaseTransactionCancel(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	testEmail := fmt.Sprintf("testuser%d@civil.co", rand.Intn(500))

	userService := buildUserService(testEmail)
	emailLists := &TestServiceEmailLists{Test: t}

	service, err := storefront.NewService("", nil, nil, userService, emailLists)
	if err != nil {
		t.Errorf("Should not have gotten error init new service: err: %v", err)
	}

	err = service.PurchaseTransactionCancel("1234")
	if err != nil {
		t.Errorf("Should not have gotten error calling on cancel: err: %v", err)
	}
}

func TestPurchaseTransactionCompleteDuplicateHash(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	testEmail := fmt.Sprintf("testuser%d@civil.co", rand.Intn(500))

	userService := buildUserService(testEmail)
	emailLists := &TestServiceEmailLists{Test: t}

	service, err := storefront.NewService("", nil, nil, userService, emailLists)
	if err != nil {
		t.Errorf("Should not have gotten error init new service: err: %v", err)
	}

	testTxHash := "0x6f73ffd316cfe4b7e9db1ca50bd9ea316cffc60c87d240eee9d9693e26d0381b"

	err = service.PurchaseTransactionComplete("1234", testTxHash)
	if err != nil {
		t.Errorf("Should not have gotten error calling on complete: err: %v", err)
	}

	err = service.PurchaseTransactionComplete("1234", testTxHash)
	if err == nil {
		t.Errorf("Should have gotten error calling on duplicate complete: err: %v", err)
	}
}
