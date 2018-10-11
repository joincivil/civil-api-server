package users

// UserPersister is an interface for CRUD of Users
type UserPersister interface {
	User(criteria *UserCriteria) (*User, error)
	SaveUser(user *User) error
	UpdateUser(user *User, updatedFields []string) error
}

// UserCriteria is used to query for a particular user
type UserCriteria struct {
	UID        string
	Email      string
	EthAddress string
}
