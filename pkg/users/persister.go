package users

// UserPersister is an interface for CRUD of Users
type UserPersister interface {
	Users(criteria *UserCriteria) ([]*User, error)
	User(criteria *UserCriteria) (*User, error)
	SaveUser(user *User) (*User, error)
	UpdateUser(user *User, updatedFields []string) error
}

// UserCriteria is used to query for a particular user(s)
type UserCriteria struct {
	UID          string `db:"uid"`
	Email        string `db:"email"`
	EthAddress   string `db:"eth_address"`
	AppReferral  string `db:"app_refer"`
	NewsroomAddr string `db:"nr_addr"`
}
