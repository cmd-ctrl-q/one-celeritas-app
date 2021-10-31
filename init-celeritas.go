package main

import (
	"log"
	"myapp/data"
	"myapp/handlers"
	"os"

	"github.com/cmd-ctrl-q/celeritas"
)

func initApplication() *application {
	path, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	// initialize celeritas
	cel := &celeritas.Celeritas{}
	err = cel.New(path)
	if err != nil {
		log.Fatal(err)
	}

	cel.AppName = "myapp"

	myHandlers := &handlers.Handlers{
		App: cel,
	}

	// build app variable
	app := &application{
		App:      cel,
		Handlers: myHandlers,
	}

	// for global access
	app.App.Routes = app.routes()
	app.Models = data.New(app.App.DB.Pool)

	// gives handlers package access to models
	myHandlers.Models = app.Models

	return app
}
