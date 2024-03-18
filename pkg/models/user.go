package models

import (
	"strings"
	"time"

	"github.com/microcosm-cc/bluemonday"

	"github.com/asaskevich/govalidator"
)

const MinLenPassword = 6

//nolint:gochecknoinits
func init() {
	govalidator.CustomTypeTagMap.Set(
		"password",
		func(i interface{}, o interface{}) bool {
			subject, ok := i.(string)
			if !ok {
				return false
			}
			if len(subject) < MinLenPassword {
				return false
			}

			return true
		},
	)
}

type User struct {
	ID       uint64 `json:"id"       valid:"required"`
	Email    string `json:"email"    valid:"required,email~Not valid email"`
	Password string `json:"password" valid:"required,password~Password must be at least 6 symbols"`
}

type UserWithoutPassword struct {
	ID        uint64    `json:"id"          valid:"required"`
	Email     string    `json:"email"       valid:"required,email~Not valid email"`
	CreatedAt time.Time `json:"created_at"  valid:"required"`
}

func (u *UserWithoutPassword) Trim() {
	u.Email = strings.TrimSpace(u.Email)
}

type UserWithoutID struct {
	Email    string `json:"email"    valid:"required,email~Not valid email"`
	Password string `json:"password" valid:"required,password~Password must be at least 6 symbols"`
}

func (u *UserWithoutID) Trim() {
	u.Email = strings.TrimSpace(u.Email)
}

func (u *UserWithoutPassword) Sanitize() {
	sanitizer := bluemonday.UGCPolicy()

	u.Email = sanitizer.Sanitize(u.Email)
}
