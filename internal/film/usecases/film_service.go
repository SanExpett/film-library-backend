package usecases

import (
	"context"
	"fmt"
	filmrepo "github.com/SanExpett/film-library-backend/internal/film/repository"
	"github.com/SanExpett/film-library-backend/pkg/models"
	myerrors "github.com/SanExpett/film-library-backend/pkg/my_errors"
	"github.com/SanExpett/film-library-backend/pkg/my_logger"
	"github.com/SanExpett/film-library-backend/pkg/utils"
	"go.uber.org/zap"
	"io"
)

var _ IFilmStorage = (*filmrepo.FilmStorage)(nil)

type IFilmStorage interface {
	AddFilm(ctx context.Context, preFilm *models.FilmWithoutID, userID uint64) (uint64, error)
	GetFilm(ctx context.Context, filmID uint64) (*models.Film, error)
	UpdateFilm(ctx context.Context, filmID uint64, userID uint64, updateFields map[string]interface{}) error
	DeleteFilm(ctx context.Context, filmID uint64, userID uint64) error
	GetFilmsListWithActorHandler(ctx context.Context, actorID uint64) ([]*models.Film, error)
	GetFilmsList(ctx context.Context, limit uint64, offset uint64, sortType uint64) ([]*models.Film, error)
	SearchFilmByTitle(ctx context.Context, searchedTitle string) ([]*models.Film, error)
	SearchFilmByActorsName(ctx context.Context, searchedTitle string) ([]*models.Film, error)
}

type FilmService struct {
	storage IFilmStorage
	logger  *zap.SugaredLogger
}

func NewFilmService(FilmStorage IFilmStorage) (*FilmService, error) {
	logger, err := my_logger.Get()
	if err != nil {
		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	return &FilmService{storage: FilmStorage, logger: logger}, nil
}

func (a *FilmService) AddFilm(ctx context.Context, r io.Reader, userID uint64) (uint64, error) {
	preFilm, err := ValidatePreFilm(r)
	if err != nil {
		return 0, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	filmID, err := a.storage.AddFilm(ctx, preFilm, userID)
	if err != nil {
		return 0, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	return filmID, nil
}

func (a *FilmService) GetFilm(ctx context.Context, filmID uint64) (*models.Film, error) {
	film, err := a.storage.GetFilm(ctx, filmID)
	if err != nil {
		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	film.Sanitize()

	return film, nil
}

func (f *FilmService) DeleteFilm(ctx context.Context, filmID uint64, userID uint64) error {
	err := f.storage.DeleteFilm(ctx, filmID, userID)
	if err != nil {
		return fmt.Errorf(myerrors.ErrTemplate, err)
	}

	return nil
}

func (a *FilmService) UpdateFilm(ctx context.Context,
	r io.Reader, isPartialUpdate bool, filmID uint64, userID uint64,
) error {
	var preFilm *models.FilmWithoutID

	var err error

	if isPartialUpdate {
		preFilm, err = ValidatePartOfPreFilm(r)
		if err != nil {
			return fmt.Errorf(myerrors.ErrTemplate, err)
		}
	} else {
		preFilm, err = ValidatePreFilm(r)
		if err != nil {
			return fmt.Errorf(myerrors.ErrTemplate, err)
		}
	}

	updateFieldsMap := utils.StructToMap(preFilm)

	err = a.storage.UpdateFilm(ctx, filmID, userID, updateFieldsMap)
	if err != nil {
		return fmt.Errorf(myerrors.ErrTemplate, err)
	}

	return nil
}

func (f *FilmService) GetFilmsListWithActorHandler(ctx context.Context, filmID uint64) ([]*models.Film, error) {
	films, err := f.storage.GetFilmsListWithActorHandler(ctx, filmID)
	if err != nil {
		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	for _, film := range films {
		film.Sanitize()
	}

	return films, nil
}

func (f *FilmService) GetFilmsList(ctx context.Context, limit uint64, offset uint64, sortType uint64,
) ([]*models.Film, error) {
	films, err := f.storage.GetFilmsList(ctx, limit, offset, sortType)
	if err != nil {
		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	for _, film := range films {
		film.Sanitize()
	}

	return films, nil
}

func (f *FilmService) SearchFilmByTitle(ctx context.Context, searchedInput string) ([]*models.Film, error) {
	films, err := f.storage.SearchFilmByTitle(ctx, searchedInput)
	if err != nil {
		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	for _, film := range films {
		film.Sanitize()
	}

	return films, nil
}

func (f *FilmService) SearchFilmByActorsName(ctx context.Context, searchedInput string) ([]*models.Film, error) {
	films, err := f.storage.SearchFilmByActorsName(ctx, searchedInput)
	if err != nil {
		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	for _, film := range films {
		film.Sanitize()
	}

	return films, nil
}
