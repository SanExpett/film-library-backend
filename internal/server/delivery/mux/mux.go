package mux

import (
	"context"
	"github.com/SanExpett/film-library-backend/pkg/middleware"
	"net/http"

	actordelivery "github.com/SanExpett/film-library-backend/internal/actor/delivery"
	filmdelivery "github.com/SanExpett/film-library-backend/internal/film/delivery"
	userdelivery "github.com/SanExpett/film-library-backend/internal/user/delivery"

	"go.uber.org/zap"
)

type ConfigMux struct {
	addrOrigin string
	schema     string
	portServer string
}

func NewConfigMux(addrOrigin string, schema string, portServer string) *ConfigMux {
	return &ConfigMux{
		addrOrigin: addrOrigin,
		schema:     schema,
		portServer: portServer,
	}
}

func NewMux(ctx context.Context, configMux *ConfigMux, userService userdelivery.IUserService,
	actorService actordelivery.IActorService, filmService filmdelivery.IFilmService, logger *zap.SugaredLogger,
) (http.Handler, error) {
	router := http.NewServeMux()

	userHandler, err := userdelivery.NewUserHandler(userService)
	if err != nil {
		return nil, err
	}

	actorHandler, err := actordelivery.NewActorHandler(actorService)
	if err != nil {
		return nil, err
	}

	filmHandler, err := filmdelivery.NewFilmHandler(filmService)
	if err != nil {
		return nil, err
	}

	router.Handle("/api/v1/signup", middleware.Context(ctx,
		middleware.SetupCORS(userHandler.SignUpHandler, configMux.addrOrigin, configMux.schema)))
	router.Handle("/api/v1/signin", middleware.Context(ctx,
		middleware.SetupCORS(userHandler.SignInHandler, configMux.addrOrigin, configMux.schema)))
	router.Handle("/api/v1/logout", middleware.Context(ctx, http.HandlerFunc(userHandler.LogOutHandler)))

	router.Handle("/api/v1/actor/add", middleware.Context(ctx,
		middleware.SetupCORS(actorHandler.AddActorHandler, configMux.addrOrigin, configMux.schema)))
	router.Handle("/api/v1/actor/get", middleware.Context(ctx,
		middleware.SetupCORS(actorHandler.GetActorHandler, configMux.addrOrigin, configMux.schema)))
	router.Handle("/api/v1/actor/update", middleware.Context(ctx,
		middleware.SetupCORS(actorHandler.UpdateActorHandler, configMux.addrOrigin, configMux.schema)))
	router.Handle("/api/v1/actor/delete", middleware.Context(ctx,
		middleware.SetupCORS(actorHandler.DeleteActorHandler, configMux.addrOrigin, configMux.schema)))
	router.Handle("/api/v1/actor/get_list_of_actors_in_film", middleware.Context(ctx,
		middleware.SetupCORS(actorHandler.GetActorsListInFilmHandler, configMux.addrOrigin, configMux.schema)))

	router.Handle("/api/v1/film/add", middleware.Context(ctx,
		middleware.SetupCORS(filmHandler.AddFilmHandler, configMux.addrOrigin, configMux.schema)))
	router.Handle("/api/v1/film/get", middleware.Context(ctx,
		middleware.SetupCORS(filmHandler.GetFilmHandler, configMux.addrOrigin, configMux.schema)))
	router.Handle("/api/v1/film/update", middleware.Context(ctx,
		middleware.SetupCORS(filmHandler.UpdateFilmHandler, configMux.addrOrigin, configMux.schema)))
	router.Handle("/api/v1/film/delete", middleware.Context(ctx,
		middleware.SetupCORS(filmHandler.DeleteFilmHandler, configMux.addrOrigin, configMux.schema)))
	router.Handle("/api/v1/film/get_list_of_films_with_actor", middleware.Context(ctx,
		middleware.SetupCORS(filmHandler.GetFilmsListWithActorHandler, configMux.addrOrigin, configMux.schema)))
	router.Handle("/api/v1/film/get_list_of_films", middleware.Context(ctx,
		middleware.SetupCORS(filmHandler.GetFilmsListHandler, configMux.addrOrigin, configMux.schema)))
	router.Handle("/api/v1/film/search_by_title", middleware.Context(ctx,
		middleware.SetupCORS(filmHandler.SearchFilmByTitleHandler, configMux.addrOrigin, configMux.schema)))
	router.Handle("/api/v1/film/search_by_actors_name", middleware.Context(ctx,
		middleware.SetupCORS(filmHandler.SearchFilmByActorsNameHandler, configMux.addrOrigin, configMux.schema)))

	mux := http.NewServeMux()
	mux.Handle("/", middleware.Panic(router, logger))

	return mux, nil
}
