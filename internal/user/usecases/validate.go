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
	ErrWrongCredentials = myerrors.NewError("Некорректный логин или пароль")
	ErrDecodeUser       = myerrors.NewError("Некорректный json пользователя")
)

func validateUserWithoutID(r io.Reader) (*models.UserWithoutID, error) {
	logger, err := my_logger.Get()
	if err != nil {
		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	decoder := json.NewDecoder(r)

	userWithoutID := new(models.UserWithoutID)
	if err := decoder.Decode(userWithoutID); err != nil {
		logger.Errorln(err)

		return nil, fmt.Errorf(myerrors.ErrTemplate, ErrDecodeUser)
	}

	userWithoutID.Trim()

	_, err = govalidator.ValidateStruct(userWithoutID)

	return userWithoutID, err //nolint:wrapcheck
}

func ValidateUserWithoutID(r io.Reader) (*models.UserWithoutID, error) {
	logger, err := my_logger.Get()
	if err != nil {
		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	userWithoutID, err := validateUserWithoutID(r)
	if err != nil {
		logger.Errorln(err)

		return nil, myerrors.NewError(err.Error())
	}

	return userWithoutID, nil
}

func ValidateUserCredentials(email string, password string) (*models.UserWithoutID, error) {
	logger, err := my_logger.Get()
	if err != nil {
		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	userWithoutID := new(models.UserWithoutID)

	userWithoutID.Email = email
	userWithoutID.Password = password
	userWithoutID.Trim()
	logger.Infoln(userWithoutID)

	_, err = govalidator.ValidateStruct(userWithoutID)
	if err != nil && (govalidator.ErrorByField(err, "email") != "" ||
		govalidator.ErrorByField(err, "password") != "") {
		logger.Errorln(err)

		return nil, ErrWrongCredentials
	}

	return userWithoutID, nil
}
