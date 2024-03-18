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
)

var (
	ErrActorNotFound       = myerrors.NewError("Этот актер не найден")
	ErrNotAuthorUpdate     = myerrors.NewError("Только автор может обновлять данные актера")
	ErrNoUpdateFields      = myerrors.NewError("Вы пытаетесь обновить пустое количество полей актера")
	ErrNoAdminAddActor     = myerrors.NewError("Только администратор может добавлять информацию об актерах")
	ErrNoAffectedActorRows = myerrors.NewError("Не получилось обновить данные актера")

	NameSeqActor = pgx.Identifier{"public", "actor_id_seq"} //nolint:gochecknoglobals
)

type ActorStorage struct {
	pool   *pgxpool.Pool
	logger *zap.SugaredLogger
}

func NewActorStorage(pool *pgxpool.Pool) (*ActorStorage, error) {
	logger, err := my_logger.Get()
	if err != nil {
		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	return &ActorStorage{
		pool:   pool,
		logger: logger,
	}, nil
}

func (a *ActorStorage) createActor(ctx context.Context, tx pgx.Tx, preActor *models.ActorWithoutID,
	userID uint64) error {
	var SQLCreateActor string

	var err error

	SQLCreateActor = `INSERT INTO public."actor" (name, birthday, gender, author_id) VALUES ($1, $2, $3, $4);`
	_, err = tx.Exec(ctx, SQLCreateActor,
		preActor.Name, preActor.Birthday, preActor.Gender, userID)

	if err != nil {
		a.logger.Errorf("in createActor: preAcotor%+v err=%+v", preActor, err)

		return fmt.Errorf(myerrors.ErrTemplate, err)
	}

	return nil
}

func (a *ActorStorage) AddActor(ctx context.Context, preActor *models.ActorWithoutID, userID uint64) (uint64, error) {
	actor := models.Actor{} //nolint:exhaustruct

	err := pgx.BeginFunc(ctx, a.pool, func(tx pgx.Tx) error {
		isAdmin, err := repository.SelectIsAdminByUserID(ctx, tx, userID)
		if err != nil {
			return fmt.Errorf(myerrors.ErrTemplate, err)
		}

		if !isAdmin {
			a.logger.Errorln(ErrNoAdminAddActor)

			return fmt.Errorf(myerrors.ErrTemplate, ErrNoAdminAddActor)
		}

		err = a.createActor(ctx, tx, preActor, userID)
		if err != nil {
			return fmt.Errorf(myerrors.ErrTemplate, err)
		}

		id, err := repository.GetLastValSeq(ctx, tx, NameSeqActor)
		if err != nil {
			return fmt.Errorf(myerrors.ErrTemplate, err)
		}

		actor.ID = id

		return nil
	})
	if err != nil {
		return 0, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	return actor.ID, nil
}

func (a *ActorStorage) selectActorByID(ctx context.Context,
	tx pgx.Tx, actorID uint64,
) (*models.Actor, error) {
	SQLSelectActor := `SELECT author_id, name, birthday, gender, created_at FROM public."actor" WHERE id=$1`
	actor := &models.Actor{ID: actorID} //nolint:exhaustruct

	actorRow := tx.QueryRow(ctx, SQLSelectActor, actorID)
	if err := actorRow.Scan(&actor.AuthorID, &actor.Name, &actor.Birthday,
		&actor.Gender, &actor.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf(myerrors.ErrTemplate, ErrActorNotFound)
		}

		a.logger.Errorf("error with actorId=%d: %+v", actorID, err)

		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	return actor, nil
}

func (a *ActorStorage) GetActor(ctx context.Context, actorID uint64) (*models.Actor, error) {
	var actor *models.Actor

	err := pgx.BeginFunc(ctx, a.pool, func(tx pgx.Tx) error {
		actorInner, err := a.selectActorByID(ctx, tx, actorID)
		if err != nil {
			return err
		}

		actor = actorInner

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	return actor, nil
}

func (a *ActorStorage) deleteActor(ctx context.Context, tx pgx.Tx, actorID uint64, userID uint64) error {
	SQLDeleteActor := `DELETE FROM public."actor" WHERE id=$1 AND actor.author_id=$2`

	result, err := tx.Exec(ctx, SQLDeleteActor, actorID, userID)
	if err != nil {
		a.logger.Errorln(err)

		return fmt.Errorf(myerrors.ErrTemplate, err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf(myerrors.ErrTemplate, ErrNoAffectedActorRows)
	}

	return nil
}

func (a *ActorStorage) DeleteActor(ctx context.Context, actorID uint64, userID uint64) error {
	err := pgx.BeginFunc(ctx, a.pool, func(tx pgx.Tx) error {
		err := a.deleteActor(ctx, tx, actorID, userID)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		a.logger.Errorln(err)

		return fmt.Errorf(myerrors.ErrTemplate, err)
	}

	return nil
}

func (a *ActorStorage) selectActorsIDsByFilmID(ctx context.Context, tx pgx.Tx,
	filmID uint64) ([]uint64, error) {
	SQLSelectActorsIDsByFilmID :=
		`SELECT actor_id
		FROM public."film_actor" 
		WHERE film_id = $1`

	actorsIDsByFilmIDRows, err := tx.Query(ctx, SQLSelectActorsIDsByFilmID, filmID)
	if err != nil {
		a.logger.Errorln(err)

		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	var curActorID uint64
	var slActorIDs []uint64

	_, err = pgx.ForEachRow(actorsIDsByFilmIDRows, []any{
		&curActorID,
	}, func() error {
		slActorIDs = append(slActorIDs, curActorID)

		return nil
	})

	if err != nil {
		a.logger.Errorln(err)

		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	return slActorIDs, nil
}

func (a *ActorStorage) GetListOfActorsInFilm(ctx context.Context, filmID uint64) ([]*models.Actor, error) {
	var slActors []*models.Actor

	err := pgx.BeginFunc(ctx, a.pool, func(tx pgx.Tx) error {
		slActorsIDs, err := a.selectActorsIDsByFilmID(ctx, tx, filmID)
		if err != nil {
			return err
		}

		for _, actorID := range slActorsIDs {
			actor, err := a.selectActorByID(ctx, tx, actorID)
			if err != nil {
				return err
			}

			slActors = append(slActors, actor)
		}

		return nil
	})
	if err != nil {
		a.logger.Errorln(err)

		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	return slActors, nil
}

func (a *ActorStorage) selectAuthorIDOfActor(ctx context.Context, tx pgx.Tx, actorID uint64) (uint64, error) {
	var authorID uint64

	SQLIsAuthorByUserIDAndActorID := `SELECT author_id FROM public."actor" WHERE id=$1`

	authorIDRow := tx.QueryRow(ctx, SQLIsAuthorByUserIDAndActorID, actorID)
	if err := authorIDRow.Scan(&authorID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, nil
		}

		a.logger.Errorln(err)

		return 0, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	return authorID, nil
}

func (a *ActorStorage) updateActor(ctx context.Context, tx pgx.Tx,
	actorID uint64, updateFields map[string]interface{},
) error {
	if len(updateFields) == 0 {
		return ErrNoUpdateFields
	}

	query := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar).Update(`public."actor"`).
		Where(squirrel.Eq{"id": actorID}).SetMap(updateFields)

	queryString, args, err := query.ToSql()
	if err != nil {
		a.logger.Errorln(err)

		return fmt.Errorf(myerrors.ErrTemplate, err)
	}

	result, err := tx.Exec(ctx, queryString, args...)
	if err != nil {
		a.logger.Errorln(err)

		return fmt.Errorf(myerrors.ErrTemplate, err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf(myerrors.ErrTemplate, ErrNoAffectedActorRows)
	}

	return nil
}

func (a *ActorStorage) UpdateActor(ctx context.Context, actorID uint64, userID uint64,
	updateFields map[string]interface{},
) error {
	err := pgx.BeginFunc(ctx, a.pool, func(tx pgx.Tx) error {
		authorID, err := a.selectAuthorIDOfActor(ctx, tx, actorID)
		if authorID != userID {
			return ErrNotAuthorUpdate
		}

		err = a.updateActor(ctx, tx, actorID, updateFields)

		return err
	})
	if err != nil {
		a.logger.Errorln(err)

		return fmt.Errorf(myerrors.ErrTemplate, err)
	}

	return nil
}
