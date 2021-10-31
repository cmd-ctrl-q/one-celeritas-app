package handlers

import "net/http"

// UserLogin displays the login page
func (h *Handlers) GetUserLogin(w http.ResponseWriter, r *http.Request) {
	err := h.App.Render.Page(w, r, "login", nil, nil)
	if err != nil {
		h.App.ErrorLog.Println(err)
		return
	}
}

func (h *Handlers) PostUserLogin(w http.ResponseWriter, r *http.Request) {
	// get info from the request
	err := r.ParseForm()
	if err != nil {
		// bad request
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")

	user, err := h.Models.Users.GetByEmail(email)
	if err != nil {
		// internal server error
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	matches, err := user.PasswordMatches(password)
	if err != nil {
		// internal server error
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error validating password"))
		return
	}

	if !matches {
		// unauthorized
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("invalid password"))
		return
	}

	// login user
	h.App.Session.Put(r.Context(), "userID", user.ID)

	// redirect user
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handlers) Logout(w http.ResponseWriter, r *http.Request) {
	h.App.Session.RenewToken(r.Context())
	h.App.Session.Remove(r.Context(), "userID")

	http.Redirect(w, r, "/users/login", http.StatusSeeOther)
}
