package usecases

import (
	"context"
	"fmt"
	userrepo "github.com/SanExpett/film-library-backend/internal/user/repository"
	"github.com/SanExpett/film-library-backend/pkg/models"
	myerrors "github.com/SanExpett/film-library-backend/pkg/my_errors"
	"github.com/SanExpett/film-library-backend/pkg/my_logger"
	"github.com/SanExpett/film-library-backend/pkg/utils"
	"go.uber.org/zap"
	"io"
)

var _ IUserStorage = (*userrepo.UserStorage)(nil)

type IUserStorage interface {
	AddUser(ctx context.Context, preUser *models.UserWithoutID) (*models.User, error)
	GetUser(ctx context.Context, email string, password string) (*models.UserWithoutPassword, error)
}

type UserService struct {
	storage IUserStorage
	logger  *zap.SugaredLogger
}

func NewUserService(userStorage IUserStorage) (*UserService, error) {
	logger, err := my_logger.Get()
	if err != nil {
		return nil, err
	}

	return &UserService{storage: userStorage, logger: logger}, nil
}

func (u *UserService) AddUser(ctx context.Context, r io.Reader) (*models.User, error) {
	userWithoutID, err := ValidateUserWithoutID(r)
	if err != nil {
		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	userWithoutID.Password, err = utils.HashPass(userWithoutID.Password)
	if err != nil {
		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	user, err := u.storage.AddUser(ctx, userWithoutID)
	if err != nil {
		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	return user, nil
}

func (u *UserService) GetUser(ctx context.Context, email string, password string) (*models.UserWithoutPassword, error) {
	userWithoutID, err := ValidateUserCredentials(email, password)
	if err != nil {
		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	user, err := u.storage.GetUser(ctx, userWithoutID.Email, userWithoutID.Password)
	if err != nil {
		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	user.Sanitize()

	return user, nil
}
