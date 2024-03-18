package main

import (
	"fmt"
	"github.com/SanExpett/film-library-backend/internal/server"
	"github.com/SanExpett/film-library-backend/pkg/config"
)

//	@title      FILM-LIBRARY project API
//	@version    1.0
//	@description  This is a server of FILM-LIBRARY server.
//
// @Schemes http
// @BasePath  /api/v1
func main() {
	configServer := config.New()

	srv := new(server.Server)
	if err := srv.Run(configServer); err != nil {
		fmt.Printf("Error in server: %s", err.Error())
	}
}
