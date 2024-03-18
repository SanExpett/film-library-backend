package delivery

import (
	"context"
	"fmt"
	"github.com/SanExpett/film-library-backend/internal/actor/usecases"
	"github.com/SanExpett/film-library-backend/internal/server/delivery"
	"github.com/SanExpett/film-library-backend/pkg/models"
	myerrors "github.com/SanExpett/film-library-backend/pkg/my_errors"
	"github.com/SanExpett/film-library-backend/pkg/my_logger"
	"github.com/SanExpett/film-library-backend/pkg/utils"
	"go.uber.org/zap"
	"io"
	"net/http"
)

var _ IActorService = (*usecases.ActorService)(nil)

type IActorService interface {
	AddActor(ctx context.Context, r io.Reader, userID uint64) (uint64, error)
	GetActor(ctx context.Context, actorID uint64) (*models.Actor, error)
	UpdateActor(ctx context.Context, r io.Reader, isPartialUpdate bool, actorID uint64, userID uint64) error
	GetListOfActorsInFilm(ctx context.Context, filmID uint64) ([]*models.Actor, error)
	DeleteActor(ctx context.Context, actorID uint64, userID uint64) error
}

type ActorHandler struct {
	service IActorService
	logger  *zap.SugaredLogger
}

func NewActorHandler(actorService IActorService) (*ActorHandler, error) {
	logger, err := my_logger.Get()
	if err != nil {
		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	return &ActorHandler{
		service: actorService,
		logger:  logger,
	}, nil
}

// AddActorHandler godoc
//
//	@Summary    add Actor
//	@Description  add Actor by data
//	@Description Error.status can be:
//	@Description StatusErrBadRequest      = 400
//	@Description  StatusErrInternalServer  = 500
//	@Tags Actor
//
//	@Accept      json
//	@Produce    json
//	@Param      Actor  body models.ActorWithoutID true  "Actor data for adding"
//	@Success    200  {object} delivery.ResponseID
//	@Failure    405  {string} string
//	@Failure    500  {string} string
//	@Failure    222  {object} delivery.ErrorResponse "Error"
//	@Router      /actor/add [post]
func (a *ActorHandler) AddActorHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `Method not allowed`, http.StatusMethodNotAllowed)

		return
	}

	ctx := r.Context()

	userID, err := delivery.GetUserIDFromCookie(r)
	if err != nil {
		delivery.HandleErr(w, a.logger, err)
		return
	}

	actorID, err := a.service.AddActor(ctx, r.Body, userID)
	if err != nil {
		delivery.HandleErr(w, a.logger, err)

		return
	}

	delivery.SendOkResponse(w, a.logger, delivery.NewResponseID(actorID))
	a.logger.Infof("in AddActorHandler: added Actor id= %+v", actorID)
}

// GetActorHandler godoc
//
//	@Summary    get Actor
//	@Description  get Actor by id
//	@Tags Actor
//	@Accept      json
//	@Produce    json
//	@Param      id  query uint64 true  "Actor id"
//	@Success    200  {object} ActorResponse
//	@Failure    405  {string} string
//	@Failure    500  {string} string
//	@Failure    222  {object} delivery.ErrorResponse "Error"
//	@Router      /actor/get [get]
func (a *ActorHandler) GetActorHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `Method not allowed`, http.StatusMethodNotAllowed)

		return
	}

	ctx := r.Context()

	actorID, err := utils.ParseUint64FromRequest(r, "id")
	if err != nil {
		delivery.HandleErr(w, a.logger, err)

		return
	}

	actor, err := a.service.GetActor(ctx, actorID)
	if err != nil {
		delivery.HandleErr(w, a.logger, err)

		return
	}

	delivery.SendOkResponse(w, a.logger, NewActorResponse(delivery.StatusResponseSuccessful, actor))
	a.logger.Infof("in GetActorHandler: get Actor: %+v", actor)
}

// DeleteActorHandler godoc
//
//	@Summary     delete Actor
//	@Description  delete Actor for author using user id from cookies\jwt.
//	@Description  This totally removed Actor. Recovery will be impossible
//	@Tags Actor
//	@Accept      json
//	@Produce    json
//	@Param      id  query uint64 true  "Actor id"
//	@Success    200  {object} delivery.Response
//	@Failure    405  {string} string
//	@Failure    500  {string} string
//	@Failure    222  {object} delivery.ErrorResponse "Error"
//	@Router      /actor/delete [delete]
func (a *ActorHandler) DeleteActorHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, `Method not allowed`, http.StatusMethodNotAllowed)

		return
	}

	ctx := r.Context()

	userID, err := delivery.GetUserIDFromCookie(r)
	if err != nil {
		delivery.HandleErr(w, a.logger, err)

		return
	}

	actorID, err := utils.ParseUint64FromRequest(r, "id")
	if err != nil {
		delivery.HandleErr(w, a.logger, err)

		return
	}

	err = a.service.DeleteActor(ctx, actorID, userID)
	if err != nil {
		delivery.HandleErr(w, a.logger, err)

		return
	}

	delivery.SendOkResponse(w, a.logger,
		delivery.NewResponse(delivery.StatusResponseSuccessful, ResponseSuccessfulDeleteActor))
	a.logger.Infof("in DeleteActorHandler: delete Actor id=%d", actorID)
}

// GetActorsListInFilmHandler godoc
//
//		@Summary    get actors list starred in film
//		@Description  get actors by film id
//		@Tags Actor
//		@Accept      json
//		@Produce    json
//	    @Param      film_id  query uint64 true  "film id"
//		@Success    200  {object} ActorListResponse
//		@Failure    405  {string} string
//		@Failure    500  {string} string
//		@Failure    222  {object} delivery.ErrorResponse "Error"
//		@Router      /actor/get_list_of_actors_in_film [get]
func (a *ActorHandler) GetActorsListInFilmHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `Method not allowed`, http.StatusMethodNotAllowed)

		return
	}

	ctx := r.Context()

	filmID, err := utils.ParseUint64FromRequest(r, "film_id")
	if err != nil {
		delivery.HandleErr(w, a.logger, err)

		return
	}

	Actors, err := a.service.GetListOfActorsInFilm(ctx, filmID)
	if err != nil {
		delivery.HandleErr(w, a.logger, err)

		return
	}

	delivery.SendOkResponse(w, a.logger, NewActorListResponse(delivery.StatusResponseSuccessful, Actors))
	a.logger.Infof("in GetActorsListInFilmHandler: get Actor list: %+v", Actors)
}

// UpdateActorHandler godoc
//
//	@Summary    update Actor
//	@Description  update Actor by id
//	@Tags Actor
//	@Accept      json
//	@Produce    json
//	@Param      id  query uint64 true  "Actor id"
//	@Param      preActor  body models.ActorWithoutID false  "полностью опционален"
//	@Success    200  {object} delivery.ResponseID
//	@Failure    405  {string} string
//	@Failure    500  {string} string
//	@Failure    222  {object} delivery.ErrorResponse "Error"
//	@Router      /actor/update [patch]
//	@Router      /actor/update [put]
func (a *ActorHandler) UpdateActorHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch && r.Method != http.MethodPut {
		http.Error(w, `Method not allowed`, http.StatusMethodNotAllowed)

		return
	}

	actorID, err := utils.ParseUint64FromRequest(r, "id")
	if err != nil {
		delivery.HandleErr(w, a.logger, err)

		return
	}

	ctx := r.Context()

	userID, err := delivery.GetUserIDFromCookie(r)
	if err != nil {
		delivery.HandleErr(w, a.logger, err)

		return
	}

	if r.Method == http.MethodPatch {
		err = a.service.UpdateActor(ctx, r.Body, true, actorID, userID)
	} else {
		err = a.service.UpdateActor(ctx, r.Body, false, actorID, userID)
	}

	if err != nil {
		delivery.HandleErr(w, a.logger, err)

		return
	}

	delivery.SendOkResponse(w, a.logger, delivery.NewResponseID(actorID))
	a.logger.Infof("in UpdateActorHandler: updated Actor with id = %+v", actorID)
}
