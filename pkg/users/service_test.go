package users_test

import (
	"testing"

	"github.com/joincivil/civil-api-server/pkg/testutils"
	"github.com/joincivil/civil-api-server/pkg/users"

	cpersist "github.com/joincivil/go-common/pkg/persistence"
)

func buildService() *users.UserService {
	initUsers := map[string]*users.User{
		"1": {UID: "1", Email: "foo@bar.com"},
	}
	persister := &testutils.InMemoryUserPersister{Users: initUsers}

	return users.NewUserService(persister)
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
		QuizStatus:        "test",
		OnfidoApplicantID: "test",
		OnfidoCheckID:     "test",
		KycStatus:         "test",
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
	svc.CreateUser(crit)

	user, err := svc.SetEthAddress(crit, "foobar")
	if err != nil {
		t.Fatalf("error in SetEthAddress")
	}

	if user.EthAddress != "foobar" {
		t.Fatalf("eth address was not set as expected")
	}
}
