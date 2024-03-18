package repository

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/SanExpett/film-library-backend/internal/server/repository"
	"github.com/SanExpett/film-library-backend/pkg/models"
	myerrors "github.com/SanExpett/film-library-backend/pkg/my_errors"
	"github.com/SanExpett/film-library-backend/pkg/my_logger"
	"github.com/SanExpett/film-library-backend/pkg/utils"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

var (
	ErrEmailBusy          = myerrors.NewError("Такой email уже занят")
	ErrEmailNotExist      = myerrors.NewError("Такой email не существует")
	ErrPhoneBusy          = myerrors.NewError("Такой телефон уже занят")
	ErrWrongPassword      = myerrors.NewError("Некорректный пароль")
	ErrNoUpdateFields     = myerrors.NewError("Вы пытаетесь обновить пустое количество полей")
	ErrNoAffectedUserRows = myerrors.NewError("Не получилось обновить данные пользователя")

	NameSeqUser = pgx.Identifier{"public", "user_id_seq"} //nolint:gochecknoglobals
)

type UserStorage struct {
	pool   *pgxpool.Pool
	logger *zap.SugaredLogger
}

func NewUserStorage(pool *pgxpool.Pool) (*UserStorage, error) {
	logger, err := my_logger.Get()
	if err != nil {
		return nil, err
	}

	return &UserStorage{
		pool:   pool,
		logger: logger,
	}, nil
}

func (u *UserStorage) createUser(ctx context.Context, tx pgx.Tx, preUser *models.UserWithoutID) error {
	var SQLCreateUser string

	var err error

	SQLCreateUser = `INSERT INTO public."user" (email, password) VALUES ($1, $2);`
	_, err = tx.Exec(ctx, SQLCreateUser,
		preUser.Email, preUser.Password)

	if err != nil {
		u.logger.Errorf("in createUser: preUser=%+v err=%+v", preUser, err)

		return fmt.Errorf(myerrors.ErrTemplate, err)
	}

	return nil
}

func (u *UserStorage) AddUser(ctx context.Context, preUser *models.UserWithoutID) (*models.User, error) {
	user := models.User{} //nolint:exhaustruct

	err := pgx.BeginFunc(ctx, u.pool, func(tx pgx.Tx) error {
		emailBusy, err := u.isEmailBusy(ctx, tx, preUser.Email)
		if err != nil {
			return fmt.Errorf(myerrors.ErrTemplate, err)
		}

		if emailBusy {
			return ErrEmailBusy
		}

		err = u.createUser(ctx, tx, preUser)
		if err != nil {
			return fmt.Errorf(myerrors.ErrTemplate, err)
		}

		id, err := repository.GetLastValSeq(ctx, tx, NameSeqUser)
		if err != nil {
			return fmt.Errorf(myerrors.ErrTemplate, err)
		}

		user.ID = id

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	user.Email = preUser.Email
	user.Password = preUser.Password

	return &user, nil
}

func (u *UserStorage) isEmailBusy(ctx context.Context, tx pgx.Tx, email string) (bool, error) {
	SQLIsEmailBusy := `SELECT id FROM public."user" WHERE email=$1;`
	userRow := tx.QueryRow(ctx, SQLIsEmailBusy, email)

	var user string

	if err := userRow.Scan(&user); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}

		u.logger.Errorln(err)

		return false, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	return true, nil
}

func (u *UserStorage) getUserByEmail(ctx context.Context, tx pgx.Tx, email string) (*models.User, error) {
	SQLGetUserByEmail := `SELECT id, email, password FROM public."user" WHERE email=$1;`
	userLine := tx.QueryRow(ctx, SQLGetUserByEmail, email)

	user := models.User{ //nolint:exhaustruct
		Email: email,
	}

	if err := userLine.Scan(&user.ID, &user.Email, &user.Password); err != nil {
		u.logger.Errorln(err)

		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	return &user, nil
}

func (u *UserStorage) GetUser(ctx context.Context, email string, password string) (*models.UserWithoutPassword, error) {
	user := &models.User{}                           //nolint:exhaustruct
	userWithoutPass := &models.UserWithoutPassword{} //nolint:exhaustruct

	err := pgx.BeginFunc(ctx, u.pool, func(tx pgx.Tx) error {
		emailBusy, err := u.isEmailBusy(ctx, tx, email)
		if err != nil {
			return fmt.Errorf(myerrors.ErrTemplate, err)
		}

		if !emailBusy {
			return ErrEmailNotExist
		}

		user, err = u.getUserByEmail(ctx, tx, email)
		if err != nil {
			return fmt.Errorf(myerrors.ErrTemplate, err)
		}

		hashPass, err := hex.DecodeString(user.Password)
		if err != nil {
			return fmt.Errorf(myerrors.ErrTemplate, err)
		}

		if !utils.ComparePassAndHash(hashPass, password) {
			return ErrWrongPassword
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	userWithoutPass.ID = user.ID
	userWithoutPass.Email = user.Email

	return userWithoutPass, nil
}
