package models

import (
	"github.com/microcosm-cc/bluemonday"
	"strings"
	"time"
)

type Actor struct {
	ID        uint64    `json:"id"          valid:"required"`
	AuthorID  uint64    `json:"autor_id"    valid:"required"`
	Name      string    `json:"name"       valid:"required"`
	Birthday  time.Time `json:"birthday"    valid:"optional"`
	Gender    string    `json:"gender"      valid:"optional,in(male|female|other)"`
	CreatedAt time.Time `json:"created_at"  valid:"required"`
}

type ActorWithoutID struct {
	Name     string    `json:"name"        valid:"required"`
	Birthday time.Time `json:"birthday"    valid:"optional"`
	Gender   string    `json:"gender"      valid:"optional,in(male|female|other)"`
}

func (a *Actor) Trim() {
	a.Name = strings.TrimSpace(a.Name)
	a.Gender = strings.TrimSpace(a.Gender)
}

func (a *ActorWithoutID) Trim() {
	a.Name = strings.TrimSpace(a.Name)
	a.Gender = strings.TrimSpace(a.Gender)
}

func (a *Actor) Sanitize() {
	sanitizer := bluemonday.UGCPolicy()

	a.Name = sanitizer.Sanitize(a.Name)
	a.Gender = sanitizer.Sanitize(a.Gender)
}
