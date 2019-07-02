package users

import (
	"strings"

	cpersist "github.com/joincivil/go-common/pkg/persistence"
)

// InMemoryUserPersister is an implementation of users.UserPersister for testing
type InMemoryUserPersister struct {
	UsersInMemory map[string]*User
}

// Users returns a list of users
func (r *InMemoryUserPersister) Users(criteria *UserCriteria) ([]*User, error) {
	var u []*User
	var target string

	if criteria.Email != "" {
		target = criteria.Email
	} else if criteria.EthAddress != "" {
		target = criteria.EthAddress
	} else if criteria.UID != "" {
		target = criteria.UID
	}

	u = []*User{}
	for _, user := range r.UsersInMemory {
		if target == user.Email || strings.ToLower(target) == strings.ToLower(user.EthAddress) || target == user.UID {
			u = append(u, user)
		}
	}

	if u == nil {
		return nil, cpersist.ErrPersisterNoResults
	}

	return u, nil

}

// User persists users
func (r *InMemoryUserPersister) User(criteria *UserCriteria) (*User, error) {
	var u *User
	var target string

	if criteria.Email != "" {
		target = criteria.Email
	} else if criteria.EthAddress != "" {
		target = criteria.EthAddress
	} else if criteria.UID != "" {
		target = criteria.UID
	}

	for _, user := range r.UsersInMemory {
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
func (r *InMemoryUserPersister) SaveUser(user *User) error {
	if user.UID == "" {
		user.GenerateUID() // nolint: errcheck
	}
	r.UsersInMemory[user.UID] = user

	return nil
}

// UpdateUser updates user instances
func (r *InMemoryUserPersister) UpdateUser(user *User, updatedFields []string) error {
	r.UsersInMemory[user.UID] = user

	return nil
}
