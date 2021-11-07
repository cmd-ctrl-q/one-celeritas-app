package main

import (
	"fmt"
	"myapp/data"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func (a *application) routes() *chi.Mux {
	// middleware must come before any routes

	// add routes
	a.App.Routes.Get("/", a.Handlers.Home)
	a.App.Routes.Get("/go-page", a.Handlers.GoPage)
	a.App.Routes.Get("/jet-page", a.Handlers.JetPage)
	a.App.Routes.Get("/sessions", a.Handlers.SessionTest)

	// GET: for retreiving the login page
	a.App.Routes.Get("/users/login", a.Handlers.GetUserLogin)
	// POST: for handling the login form
	a.App.Routes.Post("/users/login", a.Handlers.PostUserLogin)
	a.App.Routes.Get("/users/logout", a.Handlers.Logout)

	a.App.Routes.Get("/form", a.Handlers.Form)
	a.App.Routes.Post("/form", a.Handlers.PostForm)

	a.App.Routes.Get("/create-user", func(rw http.ResponseWriter, r *http.Request) {
		u := data.User{
			FirstName: "apple",
			LastName:  "banana",
			Email:     "apple@banana.com",
			Active:    1,
			Password:  "password",
		}

		id, err := a.Models.Users.Insert(u)
		if err != nil {
			// bad request
			a.App.ErrorLog.Println(err)
			return
		}

		fmt.Fprintf(rw, "%d: %s", id, u.FirstName)
	})

	a.App.Routes.Get("/get-all-users", func(rw http.ResponseWriter, r *http.Request) {
		users, err := a.Models.Users.GetAll()
		if err != nil {
			// bad request
			a.App.ErrorLog.Println(err)
			return
		}
		for _, x := range users {
			fmt.Fprint(rw, x.LastName)
		}
	})

	a.App.Routes.Get("/get-user/{id}", func(rw http.ResponseWriter, r *http.Request) {
		id, _ := strconv.Atoi(chi.URLParam(r, "id"))
		u, err := a.Models.Users.Get(id)
		if err != nil {
			// bad request
			a.App.ErrorLog.Println(err)
			return
		}

		fmt.Fprintf(rw, "%s %s %s", u.FirstName, u.LastName, u.Email)
	})

	a.App.Routes.Get("/update-user/{id}", func(rw http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if id == 0 {
			// bad request
			a.App.ErrorLog.Println("user id not provided: %w", err)
			return
		}
		u, err := a.Models.Users.Get(id)
		if err != nil {
			// bad request
			a.App.ErrorLog.Println(err)
			return
		}
		u.LastName = a.App.RandomString(10)

		// validation
		validator := a.App.Validator(nil)
		u.LastName = "diepenbrock"

		u.Validate(validator)

		if !validator.Valid() {
			fmt.Fprint(rw, "failed validation")
			return
		}

		err = u.Update(*u)
		if err != nil {
			// bad request
			a.App.ErrorLog.Println(err)
			return
		}

		fmt.Fprintf(rw, "update last name to %s", u.LastName)
	})

	a.App.Routes.Get("/test-database", func(rw http.ResponseWriter, r *http.Request) {
		query := "select id, first_name from users where id = 1"
		row := a.App.DB.Pool.QueryRowContext(r.Context(), query)

		var id int
		var name string
		err := row.Scan(&id, &name)
		if err != nil {
			a.App.ErrorLog.Println(err)
			return
		}

		fmt.Fprintf(rw, "%d %s", id, name)
	})

	// static assets
	fileServer := http.FileServer(http.Dir("./public"))
	// add these routes to the celeritas routes
	a.App.Routes.Handle("/public/*", http.StripPrefix("/public", fileServer))

	return a.App.Routes
}
