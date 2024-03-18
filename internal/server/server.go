package server

import (
	"context"
	actorrepo "github.com/SanExpett/film-library-backend/internal/actor/repository"
	actorusecases "github.com/SanExpett/film-library-backend/internal/actor/usecases"
	filmrepo "github.com/SanExpett/film-library-backend/internal/film/repository"
	filmusecases "github.com/SanExpett/film-library-backend/internal/film/usecases"
	"github.com/SanExpett/film-library-backend/internal/server/delivery/mux"
	"github.com/SanExpett/film-library-backend/internal/server/repository"
	userrepo "github.com/SanExpett/film-library-backend/internal/user/repository"
	userusecases "github.com/SanExpett/film-library-backend/internal/user/usecases"
	"github.com/SanExpett/film-library-backend/pkg/config"
	"github.com/SanExpett/film-library-backend/pkg/my_logger"
	"net/http"
	"strings"
	"time"
)

const (
	basicTimeout = 10 * time.Second
)

type Server struct {
	httpServer *http.Server
}

func (s *Server) Run(config *config.Config) error {
	baseCtx := context.Background()

	pool, err := repository.NewPgxPool(baseCtx, config.URLDataBase)
	if err != nil {
		return err //nolint:wrapcheck
	}

	logger, err := my_logger.New(strings.Split(config.OutputLogPath, " "),
		strings.Split(config.ErrorOutputLogPath, " "))
	if err != nil {
		return err //nolint:wrapcheck
	}

	defer logger.Sync()

	userStorage, err := userrepo.NewUserStorage(pool)
	if err != nil {
		return err
	}

	userService, err := userusecases.NewUserService(userStorage)
	if err != nil {
		return err
	}

	actorStorage, err := actorrepo.NewActorStorage(pool)
	if err != nil {
		return err
	}

	actorService, err := actorusecases.NewActorService(actorStorage)
	if err != nil {
		return err
	}

	filmStorage, err := filmrepo.NewFilmStorage(pool)
	if err != nil {
		return err
	}

	filmService, err := filmusecases.NewFilmService(filmStorage)
	if err != nil {
		return err
	}

	handler, err := mux.NewMux(baseCtx, mux.NewConfigMux(config.AllowOrigin,
		config.Schema, config.PortServer), userService, actorService, filmService, logger)
	if err != nil {
		return err
	}

	s.httpServer = &http.Server{ //nolint:exhaustruct
		Addr:           ":" + config.PortServer,
		Handler:        handler,
		MaxHeaderBytes: http.DefaultMaxHeaderBytes,
		ReadTimeout:    basicTimeout,
		WriteTimeout:   basicTimeout,
	}

	logger.Infof("Start server:%s", config.PortServer)

	return s.httpServer.ListenAndServe() //nolint:wrapcheck
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx) //nolint:wrapcheck
}
