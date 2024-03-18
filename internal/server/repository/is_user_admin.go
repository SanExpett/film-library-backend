package repository

import (
	"context"
	"errors"
	"fmt"
	myerrors "github.com/SanExpett/film-library-backend/pkg/my_errors"
	"github.com/SanExpett/film-library-backend/pkg/my_logger"
	"github.com/jackc/pgx/v5"
)

func SelectIsAdminByUserID(ctx context.Context, tx pgx.Tx, userID uint64) (bool, error) {
	logger, err := my_logger.Get()
	if err != nil {
		return false, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	var isAdmin bool

	SQLIsAdminByUserID := `SELECT is_admin FROM public."user" WHERE id=$1`

	isAdminRow := tx.QueryRow(ctx, SQLIsAdminByUserID, userID)
	if err := isAdminRow.Scan(&isAdmin); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}

		logger.Errorln(err)

		return false, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	return isAdmin, nil
}
