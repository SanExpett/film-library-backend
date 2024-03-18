package delivery

import "github.com/SanExpett/film-library-backend/pkg/models"

const (
	ResponseSuccessfulDeleteFilm = "Фильм успешно удален"
)

type FilmResponse struct {
	Status int          `json:"status"`
	Body   *models.Film `json:"body"`
}

func NewFilmResponse(status int, body *models.Film) *FilmResponse {
	return &FilmResponse{
		Status: status,
		Body:   body,
	}
}

type FilmListResponse struct {
	Status int            `json:"status"`
	Body   []*models.Film `json:"body"`
}

func NewFilmListResponse(status int, body []*models.Film) *FilmListResponse {
	return &FilmListResponse{
		Status: status,
		Body:   body,
	}
}
