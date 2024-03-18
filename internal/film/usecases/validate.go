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
	ErrDecodePreFilm = myerrors.NewError("Некорректный json фильма")
)

func validateFilmWithoutID(r io.Reader) (*models.FilmWithoutID, error) {
	logger, err := my_logger.Get()
	if err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(r)
	preFilm := &models.FilmWithoutID{}
	if err := decoder.Decode(preFilm); err != nil {
		logger.Errorln(err)

		return nil, fmt.Errorf(myerrors.ErrTemplate, ErrDecodePreFilm)
	}

	preFilm.Trim()

	_, err = govalidator.ValidateStruct(preFilm)
	if err != nil {
		logger.Errorln(err)

		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	return preFilm, nil
}

func ValidatePreFilm(r io.Reader) (*models.FilmWithoutID, error) {
	preFilm, err := validateFilmWithoutID(r)
	if err != nil {
		return nil, myerrors.NewError(err.Error())
	}

	return preFilm, nil
}

func ValidatePartOfPreFilm(r io.Reader) (*models.FilmWithoutID, error) {
	logger, err := my_logger.Get()
	if err != nil {
		return nil, err
	}

	preFilm, err := validateFilmWithoutID(r)
	if preFilm == nil {
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

	return preFilm, nil
}
