package users_test

import (
	"errors"
	"testing"

	"github.com/joincivil/civil-api-server/pkg/users"
)

func buildService() *users.UserService {
	initUsers := map[string]*users.User{
		"foo@bar.com": {Email: "foo@bar.com"},
	}
	persister := &InMemoryUserPersister{users: initUsers}

	return users.NewUserService(persister)
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

type InMemoryUserPersister struct {
	users map[string]*users.User
}

func (r *InMemoryUserPersister) User(criteria *users.UserCriteria) (*users.User, error) {
	u := r.users[criteria.Email]

	if u == nil {
		return nil, errors.New("No results from persister")
	}

	return u, nil

}

func (r *InMemoryUserPersister) SaveUser(user *users.User) error {
	r.users[user.UID] = user

	return nil
}

func (r *InMemoryUserPersister) UpdateUser(user *users.User, updatedFields []string) error {
	r.users[user.UID] = user

	return nil
}
