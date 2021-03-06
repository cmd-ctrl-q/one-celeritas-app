package main

import (
	"fmt"
	"myapp/data"
	"net/http"
	"strconv"

	"github.com/cmd-ctrl-q/celeritas/mailer"
	"github.com/go-chi/chi/v5"
)

func (a *application) routes() *chi.Mux {
	// middleware must come before any routes
	a.use(a.Middleware.CheckRemember)

	// add routes
	a.get("/", a.Handlers.Home)
	a.App.Routes.Get("/go-page", a.Handlers.GoPage)
	a.App.Routes.Get("/jet-page", a.Handlers.JetPage)
	a.App.Routes.Get("/sessions", a.Handlers.SessionTest)

	// GET: for retreiving the login page
	a.App.Routes.Get("/users/login", a.Handlers.GetUserLogin)
	// POST: for handling the login form
	a.post("/users/login", a.Handlers.PostUserLogin)
	a.App.Routes.Get("/users/logout", a.Handlers.Logout)
	a.get("/users/forgot-password", a.Handlers.Forgot)
	a.post("/users/forgot-password", a.Handlers.PostForgot)
	a.get("/users/reset-password", a.Handlers.ResetPasswordForm)
	a.post("/users/reset-password", a.Handlers.PostResetPassword)

	a.App.Routes.Get("/form", a.Handlers.Form)
	a.App.Routes.Post("/form", a.Handlers.PostForm)

	a.get("/json", a.Handlers.JSON)
	a.get("/xml", a.Handlers.XML)
	a.get("/download-file", a.Handlers.DownloadFile)

	a.get("/crypto", a.Handlers.TestCrypto)

	a.get("/cache-test", a.Handlers.ShowCachePage)
	// initiated by calling fetch in javascript
	a.post("/api/save-in-cache", a.Handlers.SaveInCache)
	a.post("/api/get-from-cache", a.Handlers.GetFromCache)
	a.post("/api/delete-from-cache", a.Handlers.DeleteFromCache)
	a.post("/api/empty-cache", a.Handlers.EmptyCache)

	a.get("/test-mail", func(rw http.ResponseWriter, r *http.Request) {
		msg := mailer.Message{
			From:        "sandboxfe0b760e7c85434fb0388c111ec3776a.mailgun.org",
			To:          "ted.diepenbrock@selu.edu",
			Subject:     "Test Subject - sent using an api",
			Template:    "test",
			Attachments: nil,
			Data:        nil,
		}

		// send via channel
		a.App.Mail.Jobs <- msg
		res := <-a.App.Mail.Results
		if res.Error != nil {
			a.App.ErrorLog.Println(res.Error)
		}

		// send via function
		// err := a.App.Mail.SendSMTPMessage(msg)
		// if err != nil {
		// 	a.App.ErrorLog.Println(err)
		// 	return
		// }

		fmt.Fprint(rw, "Sent mail!")
	})

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
