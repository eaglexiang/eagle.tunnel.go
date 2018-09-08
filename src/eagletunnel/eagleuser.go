package eagletunnel

import (
	"errors"
	"strings"
)

type EagleUser struct {
	Id       string
	Password string
}

func ParseEagleUser(userStr string) (*EagleUser, error) {
	var user EagleUser
	var err error
	items := strings.Split(userStr, ":")
	if len(items) >= 2 {
		user = EagleUser{Id: items[0], Password: items[1]}
	} else {
		err = errors.New("invalid user")
	}
	return &user, err
}

func (user *EagleUser) toString() string {
	return user.Id + ":" + user.Password
}

func (user *EagleUser) Check(user2Check *EagleUser) bool {
	return user.Password == user2Check.Password
}
