package models

import (
	"github.com/microcosm-cc/bluemonday"
	"strings"
	"time"
)

type Film struct {
	ID          uint64    `json:"id"           valid:"required"`
	AuthorID    uint64    `json:"autor_id"     valid:"required"`
	Title       string    `json:"title"        valid:"required, length(1|150)~Title length must be from 1 to 150"`
	Description string    `json:"description"  valid:"required, length(1|1000)~Description length must be from 1 to 1000"` //nolint
	ReleaseDate time.Time `json:"release_date" valid:"optional"`
	Rating      uint8     `json:"rating"       valid:"required, range(0|10)"`
	CreatedAt   time.Time `json:"created_at"   valid:"required"`
}

type FilmWithoutID struct {
	Title       string    `json:"title"        valid:"required, length(1|150)~Title length must be from 1 to 150"`
	Description string    `json:"description"  valid:"required, length(1|1000)~Description length must be from 1 to 1000"` //nolint
	ReleaseDate time.Time `json:"release_date" valid:"optional"`
	Rating      uint8     `json:"rating"       valid:"required, range(0|10)"`
	CreatedAt   time.Time `json:"created_at"   valid:"required"`
}

func (f *Film) Trim() {
	f.Title = strings.TrimSpace(f.Title)
	f.Description = strings.TrimSpace(f.Description)
}

func (f *FilmWithoutID) Trim() {
	f.Title = strings.TrimSpace(f.Title)
	f.Description = strings.TrimSpace(f.Description)
}

func (f *Film) Sanitize() {
	sanitizer := bluemonday.UGCPolicy()

	f.Title = sanitizer.Sanitize(f.Title)
	f.Description = sanitizer.Sanitize(f.Description)
}
