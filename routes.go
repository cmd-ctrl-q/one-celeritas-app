package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (a *application) routes() *chi.Mux {
	// middleware must come before any routes

	// add routes
	a.App.Routes.Get("/", a.Handlers.Home)

	// static assets
	fileServer := http.FileServer(http.Dir("./public"))
	// add these routes to the celeritas routes
	a.App.Routes.Handle("/public/*", http.StripPrefix("/public", fileServer))

	return a.App.Routes
}
