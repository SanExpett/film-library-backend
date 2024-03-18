package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/SanExpett/film-library-backend/internal/server/repository"
	"github.com/SanExpett/film-library-backend/pkg/models"
	myerrors "github.com/SanExpett/film-library-backend/pkg/my_errors"
	"github.com/SanExpett/film-library-backend/pkg/my_logger"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"strings"
)

var (
	ErrFilmNotFound       = myerrors.NewError("Этот актер не найден")
	ErrNotAuthorUpdate    = myerrors.NewError("Только автор может обновлять данные фильма")
	ErrNoUpdateFields     = myerrors.NewError("Вы пытаетесь обновить пустое количество полей фильма")
	ErrNoAdminAddFilm     = myerrors.NewError("Только администратор может добавлять информацию о фильмах")
	ErrNoAffectedFilmRows = myerrors.NewError("Не получилось обновить данные фильма")

	NameSeqFilm = pgx.Identifier{"public", "film_id_seq"} //nolint:gochecknoglobals
)

const (
	byTime  = 1
	byTitle = 2
)

type FilmStorage struct {
	pool   *pgxpool.Pool
	logger *zap.SugaredLogger
}

func NewFilmStorage(pool *pgxpool.Pool) (*FilmStorage, error) {
	logger, err := my_logger.Get()
	if err != nil {
		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	return &FilmStorage{
		pool:   pool,
		logger: logger,
	}, nil
}

func (f *FilmStorage) createFilm(ctx context.Context, tx pgx.Tx, preFilm *models.FilmWithoutID, userID uint64) error {
	var SQLCreateFilm string

	var err error

	SQLCreateFilm = `INSERT INTO public."film" (title, description, release_date, rating, author_id) 
						VALUES ($1, $2, $3, $4, $5);`
	_, err = tx.Exec(ctx, SQLCreateFilm,
		preFilm.Title, preFilm.Description, preFilm.ReleaseDate, preFilm.Rating, userID)

	if err != nil {
		f.logger.Errorf("in createFilm: preFilm%+v err=%+v", preFilm, err)

		return fmt.Errorf(myerrors.ErrTemplate, err)
	}

	return nil
}

func (f *FilmStorage) AddFilm(ctx context.Context, preFilm *models.FilmWithoutID, userID uint64) (uint64, error) {
	Film := models.Film{} //nolint:exhaustruct

	err := pgx.BeginFunc(ctx, f.pool, func(tx pgx.Tx) error {
		isAdmin, err := repository.SelectIsAdminByUserID(ctx, tx, userID)
		if err != nil {
			return fmt.Errorf(myerrors.ErrTemplate, err)
		}

		if !isAdmin {
			f.logger.Errorln(ErrNoAdminAddFilm)

			return fmt.Errorf(myerrors.ErrTemplate, ErrNoAdminAddFilm)
		}

		err = f.createFilm(ctx, tx, preFilm, userID)
		if err != nil {
			return fmt.Errorf(myerrors.ErrTemplate, err)
		}

		id, err := repository.GetLastValSeq(ctx, tx, NameSeqFilm)
		if err != nil {
			return fmt.Errorf(myerrors.ErrTemplate, err)
		}

		Film.ID = id

		return nil
	})
	if err != nil {
		return 0, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	return Film.ID, nil
}

func (f *FilmStorage) selectFilmByID(ctx context.Context, tx pgx.Tx, filmID uint64) (*models.Film, error) {
	SQLSelectFilm := `SELECT author_id, title, description, rating, release_date, created_at FROM public."film" WHERE id=$1`
	film := &models.Film{ID: filmID} //nolint:exhaustruct

	FilmRow := tx.QueryRow(ctx, SQLSelectFilm, filmID)
	if err := FilmRow.Scan(&film.AuthorID, &film.Title, &film.Description,
		&film.Rating, &film.ReleaseDate, &film.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf(myerrors.ErrTemplate, ErrFilmNotFound)
		}

		f.logger.Errorf("error with filmId=%d: %+v", filmID, err)

		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	return film, nil
}

func (f *FilmStorage) GetFilm(ctx context.Context, filmID uint64) (*models.Film, error) {
	var film *models.Film

	err := pgx.BeginFunc(ctx, f.pool, func(tx pgx.Tx) error {
		filmInner, err := f.selectFilmByID(ctx, tx, filmID)
		if err != nil {
			return err
		}

		film = filmInner

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	return film, nil
}

func (f *FilmStorage) deleteFilm(ctx context.Context, tx pgx.Tx, filmID uint64, userID uint64) error {
	SQLDeleteFilm := `DELETE FROM public."film" WHERE id=$1 AND author_id=$2`

	result, err := tx.Exec(ctx, SQLDeleteFilm, filmID, userID)
	if err != nil {
		f.logger.Errorln(err)

		return fmt.Errorf(myerrors.ErrTemplate, err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf(myerrors.ErrTemplate, ErrNoAffectedFilmRows)
	}

	return nil
}

func (f *FilmStorage) DeleteFilm(ctx context.Context, filmID uint64, userID uint64) error {
	err := pgx.BeginFunc(ctx, f.pool, func(tx pgx.Tx) error {
		err := f.deleteFilm(ctx, tx, filmID, userID)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		f.logger.Errorln(err)

		return fmt.Errorf(myerrors.ErrTemplate, err)
	}

	return nil
}

func (f *FilmStorage) selectAuthorIDOfFilm(ctx context.Context, tx pgx.Tx, filmID uint64) (uint64, error) {
	var authorID uint64

	SQLIsAuthorByUserIDAndFilmID := `SELECT author_id FROM public."film" WHERE id=$1`

	authorIDRow := tx.QueryRow(ctx, SQLIsAuthorByUserIDAndFilmID, filmID)
	if err := authorIDRow.Scan(&authorID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, nil
		}

		f.logger.Errorln(err)

		return 0, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	return authorID, nil
}

func (f *FilmStorage) updateFilm(ctx context.Context, tx pgx.Tx,
	filmID uint64, updateFields map[string]interface{},
) error {
	if len(updateFields) == 0 {
		return ErrNoUpdateFields
	}

	query := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar).Update(`public."Film"`).
		Where(squirrel.Eq{"id": filmID}).SetMap(updateFields)

	queryString, args, err := query.ToSql()
	if err != nil {
		f.logger.Errorln(err)

		return fmt.Errorf(myerrors.ErrTemplate, err)
	}

	result, err := tx.Exec(ctx, queryString, args...)
	if err != nil {
		f.logger.Errorln(err)

		return fmt.Errorf(myerrors.ErrTemplate, err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf(myerrors.ErrTemplate, ErrNoAffectedFilmRows)
	}

	return nil
}

func (f *FilmStorage) UpdateFilm(ctx context.Context, filmID uint64, userID uint64,
	updateFields map[string]interface{},
) error {
	err := pgx.BeginFunc(ctx, f.pool, func(tx pgx.Tx) error {
		authorID, err := f.selectAuthorIDOfFilm(ctx, tx, filmID)
		if authorID != userID {
			return ErrNotAuthorUpdate
		}

		err = f.updateFilm(ctx, tx, filmID, updateFields)

		return err
	})
	if err != nil {
		f.logger.Errorln(err)

		return fmt.Errorf(myerrors.ErrTemplate, err)
	}

	return nil
}

func (f *FilmStorage) selectFilmsIDsByActorID(ctx context.Context, tx pgx.Tx,
	actorID uint64) ([]uint64, error) {
	SQLSelectFilmsIDsByFilmID :=
		`SELECT film_id
		FROM public."film_actor" 
		WHERE actor_id = $1`

	filmsIDsByActorIDRows, err := tx.Query(ctx, SQLSelectFilmsIDsByFilmID, actorID)
	if err != nil {
		f.logger.Errorln(err)

		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	var curFilmID uint64
	var slFilmIDs []uint64

	_, err = pgx.ForEachRow(filmsIDsByActorIDRows, []any{
		&curFilmID,
	}, func() error {
		slFilmIDs = append(slFilmIDs, curFilmID)

		return nil
	})
	if err != nil {
		f.logger.Errorln(err)

		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	return slFilmIDs, nil
}

func (f *FilmStorage) GetFilmsListWithActorHandler(ctx context.Context, actorID uint64) ([]*models.Film, error) {
	var slFilms []*models.Film

	err := pgx.BeginFunc(ctx, f.pool, func(tx pgx.Tx) error {
		slFilmsIDs, err := f.selectFilmsIDsByActorID(ctx, tx, actorID)
		if err != nil {
			return err
		}

		for _, filmID := range slFilmsIDs {
			Film, err := f.selectFilmByID(ctx, tx, filmID)
			if err != nil {
				return err
			}

			slFilms = append(slFilms, Film)
		}

		return nil
	})
	if err != nil {
		f.logger.Errorln(err)

		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	return slFilms, nil
}

func (f *FilmStorage) selectFilmsInFeedWithOrderLimitOffset(ctx context.Context, tx pgx.Tx,
	limit uint64, offset uint64, orderByClause []string,
) ([]*models.Film, error) {
	query := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar).Select("id," +
		"author_id, title, description, rating, release_date, created_at").From(`public."film"`).
		OrderBy(orderByClause...).Limit(limit).Offset(offset)

	SQLQuery, args, err := query.ToSql()
	if err != nil {
		f.logger.Errorln(err)

		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	rowsFilms, err := tx.Query(ctx, SQLQuery, args...)
	if err != nil {
		f.logger.Errorln(err)

		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	curFilm := new(models.Film)

	var slFilm []*models.Film

	_, err = pgx.ForEachRow(rowsFilms, []any{
		&curFilm.ID, &curFilm.AuthorID,
		&curFilm.Title, &curFilm.Description,
		&curFilm.Rating, &curFilm.ReleaseDate, &curFilm.CreatedAt,
	}, func() error {
		slFilm = append(slFilm, &models.Film{ //nolint:exhaustruct
			ID:          curFilm.ID,
			AuthorID:    curFilm.AuthorID,
			Title:       curFilm.Title,
			Description: curFilm.Description,
			Rating:      curFilm.Rating,
			ReleaseDate: curFilm.ReleaseDate,
			CreatedAt:   curFilm.CreatedAt,
		})

		return nil
	})
	if err != nil {
		f.logger.Errorln(err)

		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	return slFilm, nil
}

func (f *FilmStorage) GetFilmsList(ctx context.Context, limit uint64, offset uint64, sortType uint64,
) ([]*models.Film, error) {
	var slFilms []*models.Film

	var orderByClause []string

	switch sortType {
	case byTime:
		orderByClause = []string{"created_at DESC"}
	case byTitle:
		orderByClause = []string{"title ASC"}
	default:
		orderByClause = []string{"rating DESC"}
	}

	err := pgx.BeginFunc(ctx, f.pool, func(tx pgx.Tx) error {
		slFilmsInner, err := f.selectFilmsInFeedWithOrderLimitOffset(ctx, tx, limit, offset, orderByClause)
		if err != nil {
			return err
		}

		slFilms = slFilmsInner

		return nil
	})
	if err != nil {
		f.logger.Errorln(err)

		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	return slFilms, nil
}

func (f *FilmStorage) searchFilmByTitle(ctx context.Context,
	tx pgx.Tx, searchInput string,
) ([]*models.Film, error) {
	SQLSearchFilm := `SELECT id, author_id, title, description, rating, release_date, created_at
FROM public."film"
WHERE to_tsvector(title) @@ to_tsquery(replace($1 || ':*', ' ', ' | '))
    ORDER BY ts_rank(to_tsvector(title), to_tsquery(replace($1 || ':*', ' ', ' | '))) DESC;`

	var films []*models.Film

	filmsRows, err := tx.Query(ctx, SQLSearchFilm, "%"+strings.ToLower(searchInput)+"%")
	if err != nil {
		f.logger.Errorln(err)

		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	curFilm := new(models.Film)

	_, err = pgx.ForEachRow(filmsRows, []any{
		&curFilm.ID, &curFilm.AuthorID, &curFilm.Title, &curFilm.Description,
		&curFilm.Rating, &curFilm.ReleaseDate, &curFilm.CreatedAt,
	}, func() error {
		films = append(films, &models.Film{
			ID:          curFilm.ID,
			AuthorID:    curFilm.AuthorID,
			Title:       curFilm.Title,
			Description: curFilm.Description,
			Rating:      curFilm.Rating,
			ReleaseDate: curFilm.ReleaseDate,
			CreatedAt:   curFilm.CreatedAt,
		})

		return nil
	})
	if err != nil {
		f.logger.Errorln(err)

		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	return films, nil
}

func (f *FilmStorage) SearchFilmByTitle(ctx context.Context, searchInput string) ([]*models.Film, error) {
	var films []*models.Film

	err := pgx.BeginFunc(ctx, f.pool, func(tx pgx.Tx) error {
		filmsInner, err := f.searchFilmByTitle(ctx, tx, searchInput)
		if err != nil {
			return err
		}

		films = filmsInner

		return nil
	})
	if err != nil {
		f.logger.Errorln(err)

		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	return films, nil
}

func (f *FilmStorage) searchFilmByActorsName(ctx context.Context,
	tx pgx.Tx, searchInput string,
) ([]*models.Film, error) {
	SQLSearchFilm := `SELECT f.id, f.author_id, f.title, f.description, f.rating, f.release_date, f.created_at
FROM public."film" f
JOIN public."film_actor" fa ON f.id = fa.film_id
JOIN public."actor" a ON fa.actor_id = a.id
WHERE to_tsvector(a.name) @@ to_tsquery(replace($1 || ':*', ' ', ' | '))
ORDER BY ts_rank(to_tsvector(a.name), to_tsquery(replace($1 || ':*', ' ', ' | '))) DESC;`

	var films []*models.Film

	filmsRows, err := tx.Query(ctx, SQLSearchFilm, "%"+strings.ToLower(searchInput)+"%")
	if err != nil {
		f.logger.Errorln(err)

		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	curFilm := new(models.Film)

	_, err = pgx.ForEachRow(filmsRows, []any{
		&curFilm.ID, &curFilm.AuthorID, &curFilm.Title, &curFilm.Description,
		&curFilm.Rating, &curFilm.ReleaseDate, &curFilm.CreatedAt,
	}, func() error {
		films = append(films, &models.Film{
			ID:          curFilm.ID,
			AuthorID:    curFilm.AuthorID,
			Title:       curFilm.Title,
			Description: curFilm.Description,
			Rating:      curFilm.Rating,
			ReleaseDate: curFilm.ReleaseDate,
			CreatedAt:   curFilm.CreatedAt,
		})

		return nil
	})
	if err != nil {
		f.logger.Errorln(err)

		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	return films, nil
}

func (f *FilmStorage) SearchFilmByActorsName(ctx context.Context, searchInput string) ([]*models.Film, error) {
	var films []*models.Film

	err := pgx.BeginFunc(ctx, f.pool, func(tx pgx.Tx) error {
		filmsInner, err := f.searchFilmByActorsName(ctx, tx, searchInput)
		if err != nil {
			return err
		}

		films = filmsInner

		return nil
	})
	if err != nil {
		f.logger.Errorln(err)

		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	return films, nil
}
