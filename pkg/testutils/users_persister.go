package testutils

import (
	"strings"

	"github.com/joincivil/civil-api-server/pkg/users"

	cpersist "github.com/joincivil/go-common/pkg/persistence"
)

// InMemoryUserPersister is an implementation of users.UserPersister for testing
type InMemoryUserPersister struct {
	Users map[string]*users.User
}

// User persists users
func (r *InMemoryUserPersister) User(criteria *users.UserCriteria) (*users.User, error) {
	var u *users.User
	var target string

	if criteria.Email != "" {
		target = criteria.Email
	} else if criteria.EthAddress != "" {
		target = criteria.EthAddress
	} else if criteria.UID != "" {
		target = criteria.UID
	}

	for _, user := range r.Users {
		if target == user.Email || strings.ToLower(target) == strings.ToLower(user.EthAddress) || target == user.UID {
			u = user
			break
		}
	}

	if u == nil {
		return nil, cpersist.ErrPersisterNoResults
	}

	return u, nil

}

// SaveUser saves user instances
func (r *InMemoryUserPersister) SaveUser(user *users.User) error {
	if user.UID == "" {
		user.GenerateUID() // nolint: errcheck
	}
	r.Users[user.UID] = user

	return nil
}

// UpdateUser updates user instances
func (r *InMemoryUserPersister) UpdateUser(user *users.User, updatedFields []string) error {
	r.Users[user.UID] = user

	return nil
}
