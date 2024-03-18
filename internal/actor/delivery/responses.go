package delivery

import "github.com/SanExpett/film-library-backend/pkg/models"

const (
	ResponseSuccessfulDeleteActor = "Актер успешно удален"
)

type ActorResponse struct {
	Status int           `json:"status"`
	Body   *models.Actor `json:"body"`
}

func NewActorResponse(status int, body *models.Actor) *ActorResponse {
	return &ActorResponse{
		Status: status,
		Body:   body,
	}
}

type ActorListResponse struct {
	Status int             `json:"status"`
	Body   []*models.Actor `json:"body"`
}

func NewActorListResponse(status int, body []*models.Actor) *ActorListResponse {
	return &ActorListResponse{
		Status: status,
		Body:   body,
	}
}
