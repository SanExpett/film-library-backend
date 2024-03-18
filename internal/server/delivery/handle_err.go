package delivery

import (
	"errors"
	myerrors "github.com/SanExpett/film-library-backend/pkg/my_errors"
	"go.uber.org/zap"
	"net/http"
)

func HandleErr(w http.ResponseWriter, logger *zap.SugaredLogger, err error) {
	myErr := &myerrors.Error{}
	if errors.As(err, &myErr) {
		SendErrResponse(w, logger, NewErrResponse(StatusErrBadRequest, err.Error()))

		return
	}

	SendErrResponse(w, logger, NewErrResponse(StatusErrInternalServer, ErrInternalServer))
}
