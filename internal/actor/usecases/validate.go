package usecases

import (
	"encoding/json"
	"fmt"
	"github.com/SanExpett/film-library-backend/pkg/models"
	myerrors "github.com/SanExpett/film-library-backend/pkg/my_errors"
	"github.com/SanExpett/film-library-backend/pkg/my_logger"
	"github.com/asaskevich/govalidator"
	"io"
)

var (
	ErrDecodePreActor = myerrors.NewError("Некорректный json актер")
)

func validateActorWithoutID(r io.Reader) (*models.ActorWithoutID, error) {
	logger, err := my_logger.Get()
	if err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(r)
	preActor := &models.ActorWithoutID{}
	if err := decoder.Decode(preActor); err != nil {
		logger.Errorln(err)

		return nil, fmt.Errorf(myerrors.ErrTemplate, ErrDecodePreActor)
	}

	preActor.Trim()

	_, err = govalidator.ValidateStruct(preActor)
	if err != nil {
		logger.Errorln(err)

		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	return preActor, nil
}

func ValidatePreActor(r io.Reader) (*models.ActorWithoutID, error) {
	preActor, err := validateActorWithoutID(r)
	if err != nil {
		return nil, myerrors.NewError(err.Error())
	}

	return preActor, nil
}

func ValidatePartOfPreActor(r io.Reader) (*models.ActorWithoutID, error) {
	logger, err := my_logger.Get()
	if err != nil {
		return nil, err
	}

	preActor, err := validateActorWithoutID(r)
	if preActor == nil {
		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	if err != nil {
		validationErrors := govalidator.ErrorsByField(err)

		for field, err := range validationErrors {
			if err != "non zero value required" {
				logger.Errorln(err)

				return nil, myerrors.NewError("%s error: %s", field, err)
			}
		}
	}

	return preActor, nil
}
