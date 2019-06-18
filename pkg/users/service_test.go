package users_test

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/joincivil/civil-api-server/pkg/testutils"
	"github.com/joincivil/civil-api-server/pkg/users"

	cpersist "github.com/joincivil/go-common/pkg/persistence"
)

func buildService() *users.UserService {
	initUsers := map[string]*users.User{
		"1": {UID: "1", Email: "foo@bar.com"},
	}
	persister := &testutils.InMemoryUserPersister{UsersInMemory: initUsers}

	return users.NewUserService(persister, &testutils.ControllerUpdaterSpy{})
}

func TestGetUser(t *testing.T) {
	svc := buildService()
	criteria := users.UserCriteria{Email: "foo@bar.com"}
	user, err := svc.GetUser(criteria)

	if err != nil {
		t.Fatal(err)
	}
	if user.Email != "foo@bar.com" && user.UID != "1" {
		t.Fatalf("unexpected results")
	}

	criteria = users.UserCriteria{Email: "nonexistent@bar.com"}
	_, err = svc.GetUser(criteria)
	if err != cpersist.ErrPersisterNoResults {
		t.Fatalf("should have an error")
	}
}

func TestMaybeGetUser(t *testing.T) {
	svc := buildService()
	criteria := users.UserCriteria{Email: "newuser@bar.com"}
	user, err := svc.MaybeGetUser(criteria)

	// make sure that this returns nil, nil if the user doesn't exist
	if err != nil {
		t.Fatal(err)
	}
	if user != nil {
		t.Error("this should not return a user")
	}

	criteria = users.UserCriteria{Email: "foo@bar.com"}
	user, err = svc.MaybeGetUser(criteria)

	// make sure this returns the user as expected
	if err != nil {
		t.Fatal(err)
	}
	if user == nil {
		t.Error("expecting to find a user")
	}
	if user != nil && user.Email != "foo@bar.com" {
		t.Error("email address should be foo@bar.com")
	}

}

func TestCreateUser(t *testing.T) {
	svc := buildService()
	crit := users.UserCriteria{Email: "newuser@bar.com"}
	_, err := svc.CreateUser(crit)

	if err != nil {
		t.Fatal(err)
	}

	_, err = svc.CreateUser(crit)

	if err != users.ErrUserExists {
		t.Fatalf("should have resulted in ErrUserExists")
	}

}

func TestUpdateUser(t *testing.T) {
	svc := buildService()

	update := &users.UserUpdateInput{
		QuizStatus: "test",
	}
	user, err := svc.UpdateUser("1", update)

	if err != nil {
		t.Fatal(err)
	}
	if user.QuizStatus != "test" {
		t.Fatalf("QuizStatus is not what it should be")
	}

}

func TestSetEthAddress(t *testing.T) {
	svc := buildService()
	crit := users.UserCriteria{Email: "foo@bar.com"}
	svc.CreateUser(crit) // nolint: errcheck

	user, err := svc.SetEthAddress(crit, "foobar")
	if err != nil {
		t.Fatalf("error in SetEthAddress")
	}

	if user.EthAddress != "foobar" {
		t.Fatalf("eth address was not set as expected")
	}
}

func TestQuizComplete(t *testing.T) {
	initUsers := map[string]*users.User{
		"1": {UID: "1", Email: "foo@bar.com", EthAddress: "test"},
		"2": {UID: "2", Email: "alice@bar.com"},
		"3": {UID: "3", Email: "bob@bar.com", EthAddress: common.HexToAddress("0x001").String()},
	}
	persister := &testutils.InMemoryUserPersister{UsersInMemory: initUsers}
	updater := &testutils.ControllerUpdaterSpy{}
	svc := users.NewUserService(persister, updater)

	// non-complete quizs should not call the controller updater
	_, err := svc.UpdateUser("1", &users.UserUpdateInput{
		QuizStatus: "test",
	})
	if err != nil {
		t.Fatalf("not expecting an error: %v", err)
	}
	if updater.Calls != 0 {
		t.Fatalf("expecting token controller spy to have calls = 0 but it was %v", updater.Calls)
	}

	// when QuizStatus is "complete" we need to add the user to the civilian whitelist
	user, err := svc.UpdateUser("1", &users.UserUpdateInput{
		QuizStatus: "complete",
	})
	if err != nil {
		t.Fatalf("not expecting an error: %v", err)
	}

	if updater.Calls != 1 {
		t.Fatalf("expecting token controller spy to have calls = 1 but it was %v", updater.Calls)
	}

	if user.CivilianWhitelistTxID != common.HexToHash("0xf00").String() {
		t.Fatalf("expecting user.CivilianWhitelistTxID to be common.Hash(0xf00) but it is %v", user.CivilianWhitelistTxID)
	}

	// return an error if QuizStatus is "complete" but hasn't set EthAddress
	_, err = svc.UpdateUser("2", &users.UserUpdateInput{
		QuizStatus: "complete",
	})

	if err != users.ErrInvalidState {
		t.Fatal("completing quiz with no ETH address should fail with `users.ErrInvalidState`")
	}

	// when QuizStatus is "complete" and user is already on whitelist
	user, err = svc.UpdateUser("3", &users.UserUpdateInput{
		QuizStatus: "complete",
	})
	if err != nil {
		t.Fatalf("not expecting an error: %v", err)
	}

	if updater.Calls != 2 {
		t.Fatalf("expecting token controller spy to have calls = 2 but it was %v", updater.Calls)
	}

	if user.CivilianWhitelistTxID != common.HexToHash("0xf00").String() {
		t.Fatalf("expecting user.CivilianWhitelistTxID to be common.Hash(0xf00) but it is %v", user.CivilianWhitelistTxID)
	}

}
