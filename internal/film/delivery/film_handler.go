package delivery

import (
	"context"
	"fmt"
	"github.com/SanExpett/film-library-backend/internal/film/usecases"
	"github.com/SanExpett/film-library-backend/internal/server/delivery"
	"github.com/SanExpett/film-library-backend/pkg/models"
	myerrors "github.com/SanExpett/film-library-backend/pkg/my_errors"
	"github.com/SanExpett/film-library-backend/pkg/my_logger"
	"github.com/SanExpett/film-library-backend/pkg/utils"
	"go.uber.org/zap"
	"io"
	"net/http"
)

var _ IFilmService = (*usecases.FilmService)(nil)

type IFilmService interface {
	AddFilm(ctx context.Context, r io.Reader, userID uint64) (uint64, error)
	GetFilm(ctx context.Context, filmID uint64) (*models.Film, error)
	UpdateFilm(ctx context.Context, r io.Reader, isPartialUpdate bool, filmID uint64, userID uint64) error
	GetFilmsListWithActorHandler(ctx context.Context, actorID uint64) ([]*models.Film, error)
	DeleteFilm(ctx context.Context, filmID uint64, userID uint64) error
	GetFilmsList(ctx context.Context, limit uint64, offset uint64, sortType uint64) ([]*models.Film, error)
	SearchFilmByTitle(ctx context.Context, searchedTitle string) ([]*models.Film, error)
	SearchFilmByActorsName(ctx context.Context, searchedTitle string) ([]*models.Film, error)
}

type FilmHandler struct {
	service IFilmService
	logger  *zap.SugaredLogger
}

func NewFilmHandler(filmService IFilmService) (*FilmHandler, error) {
	logger, err := my_logger.Get()
	if err != nil {
		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	return &FilmHandler{
		service: filmService,
		logger:  logger,
	}, nil
}

// AddFilmHandler godoc
//
//	@Summary    add Film
//	@Description  add Film by data
//	@Description Error.status can be:
//	@Description StatusErrBadRequest      = 400
//	@Description  StatusErrInternalServer  = 500
//	@Tags Film
//
//	@Accept      json
//	@Produce    json
//	@Param      Film  body models.FilmWithoutID true  "Film data for adding"
//	@Success    200  {object} delivery.ResponseID
//	@Failure    405  {string} string
//	@Failure    500  {string} string
//	@Failure    222  {object} delivery.ErrorResponse "Error"
//	@Router      /film/add [post]
func (f *FilmHandler) AddFilmHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `Method not allowed`, http.StatusMethodNotAllowed)

		return
	}

	ctx := r.Context()

	userID, err := delivery.GetUserIDFromCookie(r)
	if err != nil {
		delivery.HandleErr(w, f.logger, err)

		return
	}

	filmID, err := f.service.AddFilm(ctx, r.Body, userID)
	if err != nil {
		delivery.HandleErr(w, f.logger, err)

		return
	}

	delivery.SendOkResponse(w, f.logger, delivery.NewResponseID(filmID))
	f.logger.Infof("in AddFilmHandler: added Film id= %+v", filmID)
}

// GetFilmHandler godoc
//
//	@Summary    get Film
//	@Description  get Film by id
//	@Tags Film
//	@Accept      json
//	@Produce    json
//	@Param      id  query uint64 true  "Film id"
//	@Success    200  {object} FilmResponse
//	@Failure    405  {string} string
//	@Failure    500  {string} string
//	@Failure    222  {object} delivery.ErrorResponse "Error"
//	@Router      /film/get [get]
func (f *FilmHandler) GetFilmHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `Method not allowed`, http.StatusMethodNotAllowed)

		return
	}

	ctx := r.Context()

	filmID, err := utils.ParseUint64FromRequest(r, "id")
	if err != nil {
		delivery.HandleErr(w, f.logger, err)

		return
	}

	film, err := f.service.GetFilm(ctx, filmID)
	if err != nil {
		delivery.HandleErr(w, f.logger, err)

		return
	}

	delivery.SendOkResponse(w, f.logger, NewFilmResponse(delivery.StatusResponseSuccessful, film))
	f.logger.Infof("in GetFilmHandler: get Film: %+v", film)
}

// DeleteFilmHandler godoc
//
//	@Summary     delete Film
//	@Description  delete Film for author using user id from cookies\jwt.
//	@Description  This totally removed Film. Recovery will be impossible
//	@Tags Film
//	@Accept      json
//	@Produce    json
//	@Param      id  query uint64 true  "Film id"
//	@Success    200  {object} delivery.Response
//	@Failure    405  {string} string
//	@Failure    500  {string} string
//	@Failure    222  {object} delivery.ErrorResponse "Error"
//	@Router      /film/delete [delete]
func (f *FilmHandler) DeleteFilmHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, `Method not allowed`, http.StatusMethodNotAllowed)

		return
	}

	ctx := r.Context()

	userID, err := delivery.GetUserIDFromCookie(r)
	if err != nil {
		delivery.HandleErr(w, f.logger, err)

		return
	}

	filmID, err := utils.ParseUint64FromRequest(r, "id")
	if err != nil {
		delivery.HandleErr(w, f.logger, err)

		return
	}

	err = f.service.DeleteFilm(ctx, filmID, userID)
	if err != nil {
		delivery.HandleErr(w, f.logger, err)

		return
	}

	delivery.SendOkResponse(w, f.logger,
		delivery.NewResponse(delivery.StatusResponseSuccessful, ResponseSuccessfulDeleteFilm))
	f.logger.Infof("in DeleteFilmHandler: delete Film id=%d", filmID)
}

// UpdateFilmHandler godoc
//
//	@Summary    update Film
//	@Description  update Film by id
//	@Tags Film
//	@Accept      json
//	@Produce    json
//	@Param      id query uint64 true  "Film id"
//	@Param      preFilm  body models.PreFilm false  "полностью опционален"
//	@Success    200  {object} delivery.ResponseID
//	@Failure    405  {string} string
//	@Failure    500  {string} string
//	@Failure    222  {object} delivery.ErrorResponse "Error"
//	@Router      /film/update [patch]
//	@Router      /film/update [put]
func (f *FilmHandler) UpdateFilmHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch && r.Method != http.MethodPut {
		http.Error(w, `Method not allowed`, http.StatusMethodNotAllowed)

		return
	}

	filmID, err := utils.ParseUint64FromRequest(r, "id")
	if err != nil {
		delivery.HandleErr(w, f.logger, err)

		return
	}

	ctx := r.Context()

	userID, err := delivery.GetUserIDFromCookie(r)
	if err != nil {
		delivery.HandleErr(w, f.logger, err)

		return
	}

	if r.Method == http.MethodPatch {
		err = f.service.UpdateFilm(ctx, r.Body, true, filmID, userID)
	} else {
		err = f.service.UpdateFilm(ctx, r.Body, false, filmID, userID)
	}

	if err != nil {
		delivery.HandleErr(w, f.logger, err)

		return
	}

	delivery.SendOkResponse(w, f.logger, delivery.NewResponseID(filmID))
	f.logger.Infof("in UpdateFilmHandler: updated Film with id = %+v", filmID)
}

// GetFilmsListWithActorHandler godoc
//
//		@Summary    get Films list starred in film
//		@Description  get Films by film id
//		@Tags Film
//		@Accept      json
//		@Produce    json
//	    @Param      film_id  query uint64 true  "film id"
//		@Success    200  {object} FilmListResponse
//		@Failure    405  {string} string
//		@Failure    500  {string} string
//		@Failure    222  {object} delivery.ErrorResponse "Error"
//		@Router      /film/get_list_of_films_with_actor [get]
func (p *FilmHandler) GetFilmsListWithActorHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `Method not allowed`, http.StatusMethodNotAllowed)

		return
	}

	ctx := r.Context()

	actorID, err := utils.ParseUint64FromRequest(r, "actor_id")
	if err != nil {
		delivery.HandleErr(w, p.logger, err)

		return
	}

	films, err := p.service.GetFilmsListWithActorHandler(ctx, actorID)
	if err != nil {
		delivery.HandleErr(w, p.logger, err)

		return
	}

	delivery.SendOkResponse(w, p.logger, NewFilmListResponse(delivery.StatusResponseSuccessful, films))
	p.logger.Infof("in GetFilmsListInFilmHandler: get Film list: %+v", films)
}

// GetFilmsListHandler godoc
//
//	@Summary    get Films list
//	@Description  get Films by count and last_id return old Films
//	@Tags Film
//	@Accept      json
//	@Produce    json
//	@Param      limit  query uint64 true  "limit Films"
//	@Param      offset  query uint64 true  "offset of Films"
//	@Param      sort_type query uint64 true  "type of sort(nil - by rating, 1 - by time, 2 - by title)"
//	@Success    200  {object} FilmListResponse
//	@Failure    405  {string} string
//	@Failure    500  {string} string
//	@Failure    222  {object} responses.ErrorResponse "Error"
//	@Router      /film/get_list_of_films [get]
func (f *FilmHandler) GetFilmsListHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `Method not allowed`, http.StatusMethodNotAllowed)

		return
	}

	ctx := r.Context()

	limit, err := utils.ParseUint64FromRequest(r, "limit")
	if err != nil {
		limit = 10
	}

	offset, err := utils.ParseUint64FromRequest(r, "offset")
	if err != nil {
		offset = 0
	}

	sortType, err := utils.ParseUint64FromRequest(r, "sort_type")
	if err != nil {
		sortType = 0
	}

	films, err := f.service.GetFilmsList(ctx, limit, offset, sortType)
	if err != nil {
		delivery.HandleErr(w, f.logger, err)

		return
	}

	delivery.SendOkResponse(w, f.logger, NewFilmListResponse(delivery.StatusResponseSuccessful, films))
	f.logger.Infof("in GetFilmListHandler: get film list: %+v", films)
}

// SearchFilmByTitleHandler godoc
//
//	@Summary    search Film
//	@Description  search top 5 common named films
//	@Tags Film
//	@Produce    json
//	@Param      searched  query string true  "searched string"
//	@Success    200  {object} FilmListResponse
//	@Failure    405  {string} string
//	@Failure    500  {string} string
//	@Failure    222  {object} responses.ErrorResponse "Error"
//	@Router      /film/search_by_title [get]
func (f *FilmHandler) SearchFilmByTitleHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `Method not allowed`, http.StatusMethodNotAllowed)

		return
	}

	ctx := r.Context()

	searchInput := utils.ParseStringFromRequest(r, "searched")

	films, err := f.service.SearchFilmByTitle(ctx, searchInput)
	if err != nil {
		delivery.HandleErr(w, f.logger, err)

		return
	}

	delivery.SendOkResponse(w, f.logger, NewFilmListResponse(delivery.StatusResponseSuccessful, films))
	f.logger.Infof("in SearchFilmByTitleHandler: get film list: %+v", films)
}

// SearchFilmByActorsNameHandler godoc
//
//	@Summary    search film by actors name
//	@Description  search top 5 common named films
//	@Tags Film
//	@Produce    json
//	@Param      searched  query string true  "searched string"
//	@Success    200  {object} FilmListResponse
//	@Failure    405  {string} string
//	@Failure    500  {string} string
//	@Failure    222  {object} responses.ErrorResponse "Error"
//	@Router      /film/search_by_actors_name [get]
func (f *FilmHandler) SearchFilmByActorsNameHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `Method not allowed`, http.StatusMethodNotAllowed)

		return
	}

	ctx := r.Context()

	searchInput := utils.ParseStringFromRequest(r, "searched")

	films, err := f.service.SearchFilmByActorsName(ctx, searchInput)
	if err != nil {
		delivery.HandleErr(w, f.logger, err)

		return
	}

	delivery.SendOkResponse(w, f.logger, NewFilmListResponse(delivery.StatusResponseSuccessful, films))
	f.logger.Infof("in SearchFilmByActorsNameHandler: get film list: %+v", films)
}
