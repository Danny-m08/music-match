package types

import "fmt"

type User struct {
	Name      string     `json:"name"`
	Username  string     `json:"username"`
	Password  string     `json:"password"`
	Email     string     `json:"email"`
	Following []*User    `json:"following,omitempty"`
	Followers []*User    `json:"followers,omitempty"`
	Listings  []*Listing `json:"listings,omitempty"`
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
	return fmt.Sprintf("{name: '%s', username: '%s', email: '%s', password: '%s'}", U.Name, U.Username, U.Email, U.Password)
}
