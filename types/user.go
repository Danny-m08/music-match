package types

import "fmt"

type User struct {
	Name      string
	Username  string
	Password  string
	Email     string
	Following []*User
	Followers []*User
	Listings  []*Listing
}

func NewUser() *User {
	return &User{
		Username:  "",
		Password:  "",
		Email:     "",
		Following: nil,
		Followers: nil,
		Listings:  nil,
	}
}

func (U *User) String() string {
	return fmt.Sprintf("{username: '%s', email: '%s', password: '%s'}", U.Username, U.Email, U.Password)
}
