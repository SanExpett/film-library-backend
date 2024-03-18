package usecases

import (
	"context"
	"fmt"
	actorrepo "github.com/SanExpett/film-library-backend/internal/actor/repository"
	"github.com/SanExpett/film-library-backend/pkg/models"
	myerrors "github.com/SanExpett/film-library-backend/pkg/my_errors"
	"github.com/SanExpett/film-library-backend/pkg/my_logger"
	"github.com/SanExpett/film-library-backend/pkg/utils"
	"go.uber.org/zap"
	"io"
)

var _ IActorStorage = (*actorrepo.ActorStorage)(nil)

type IActorStorage interface {
	AddActor(ctx context.Context, preActor *models.ActorWithoutID, userID uint64) (uint64, error)
	GetActor(ctx context.Context, ActorID uint64) (*models.Actor, error)
	UpdateActor(ctx context.Context, actorID uint64, userID uint64, updateFields map[string]interface{}) error
	DeleteActor(ctx context.Context, actorID uint64, userID uint64) error
	GetListOfActorsInFilm(ctx context.Context, filmID uint64) ([]*models.Actor, error)
}

type ActorService struct {
	storage IActorStorage
	logger  *zap.SugaredLogger
}

func NewActorService(actorStorage IActorStorage) (*ActorService, error) {
	logger, err := my_logger.Get()
	if err != nil {
		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	return &ActorService{storage: actorStorage, logger: logger}, nil
}

func (a *ActorService) AddActor(ctx context.Context, r io.Reader, userID uint64) (uint64, error) {
	preActor, err := ValidatePreActor(r)
	if err != nil {
		return 0, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	ActorID, err := a.storage.AddActor(ctx, preActor, userID)
	if err != nil {
		return 0, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	return ActorID, nil
}

func (a *ActorService) GetActor(ctx context.Context, actorID uint64) (*models.Actor, error) {
	actor, err := a.storage.GetActor(ctx, actorID)
	if err != nil {
		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	actor.Sanitize()

	return actor, nil
}

func (p *ActorService) DeleteActor(ctx context.Context, actorID uint64, userID uint64) error {
	err := p.storage.DeleteActor(ctx, actorID, userID)
	if err != nil {
		return fmt.Errorf(myerrors.ErrTemplate, err)
	}

	return nil
}

func (a *ActorService) GetListOfActorsInFilm(ctx context.Context, filmID uint64) ([]*models.Actor, error) {
	actors, err := a.storage.GetListOfActorsInFilm(ctx, filmID)
	if err != nil {
		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	for _, actor := range actors {
		actor.Sanitize()
	}

	return actors, nil
}

func (a *ActorService) UpdateActor(ctx context.Context,
	r io.Reader, isPartialUpdate bool, actorID uint64, userID uint64,
) error {
	var preActor *models.ActorWithoutID

	var err error

	if isPartialUpdate {
		preActor, err = ValidatePartOfPreActor(r)
		if err != nil {
			return fmt.Errorf(myerrors.ErrTemplate, err)
		}
	} else {
		preActor, err = ValidatePreActor(r)
		if err != nil {
			return fmt.Errorf(myerrors.ErrTemplate, err)
		}
	}

	updateFieldsMap := utils.StructToMap(preActor)

	err = a.storage.UpdateActor(ctx, actorID, userID, updateFieldsMap)
	if err != nil {
		return fmt.Errorf(myerrors.ErrTemplate, err)
	}

	return nil
}
