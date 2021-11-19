package handlers

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"myapp/data"
	"net/http"
	"time"
)

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

	// did user check remember me?
	if r.Form.Get("remember") == "remember" {
		// create token to login with cookie
		randomString := h.randomString(12)
		hasher := sha256.New()
		_, err := hasher.Write([]byte(randomString))
		if err != nil {
			h.App.ErrorStatus(w, http.StatusBadRequest)
			return
		}

		// insert token into db
		sha := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
		rm := data.RememberToken{}
		err = rm.InsertToken(user.ID, sha)
		if err != nil {
			h.App.ErrorStatus(w, http.StatusBadRequest)
			return
		}

		// set cookie
		expire := time.Now().Add(365 * 24 * 60 * 60 * time.Second)
		cookie := http.Cookie{
			Name:     fmt.Sprintf("_%s_remember", h.App.AppName),
			Value:    fmt.Sprintf("%d|%s", user.ID, sha),
			Path:     "/",
			Expires:  expire,
			HttpOnly: true,
			Domain:   h.App.Session.Cookie.Domain,
			MaxAge:   315360000,
			Secure:   h.App.Session.Cookie.Secure,
			SameSite: http.SameSiteStrictMode,
		}
		http.SetCookie(w, &cookie)
		// save hash in session
		h.App.Session.Put(r.Context(), "remember_token", sha)
	}

	// login user
	h.App.Session.Put(r.Context(), "userID", user.ID)

	// redirect user
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handlers) Logout(w http.ResponseWriter, r *http.Request) {
	// delete remember token if exists
	if h.App.Session.Exists(r.Context(), "remember_token") {
		rt := data.RememberToken{}
		_ = rt.Delete(h.App.Session.GetString(r.Context(), "remember_token"))
	}

	// delete cookie
	newCookie := http.Cookie{
		Name:     fmt.Sprintf("_%s_remember", h.App.AppName),
		Value:    "",
		Path:     "/",
		Expires:  time.Now().Add(-100 * time.Hour),
		HttpOnly: true,
		Domain:   h.App.Session.Cookie.Domain,
		MaxAge:   -1,
		Secure:   h.App.Session.Cookie.Secure,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, &newCookie)

	// renew token with the expired values
	h.App.Session.RenewToken(r.Context())
	h.App.Session.Remove(r.Context(), "userID")
	h.App.Session.Remove(r.Context(), "remember_token")
	// destroy session
	h.App.Session.Destroy(r.Context())
	// renew again (just in case)
	h.App.Session.RenewToken(r.Context())

	http.Redirect(w, r, "/users/login", http.StatusSeeOther)
}

func (h *Handlers) Forgot(w http.ResponseWriter, r *http.Request) {
	err := h.render(w, r, "forgot", nil, nil)
	if err != nil {
		h.App.ErrorLog.Println("Error rendering:", err)
		h.App.Error500(w, r)
	}
}

func (h *Handlers) PostForgot(w http.ResponseWriter, r *http.Request) {
	// check email address in form.

	// look up user with the email address

	// if user exists, then send instructions to email associated with that user (including a link which is unique to each user)
	// ie generate token, save it to db, have it tied to this particular email address, when user submits token (which is in the link)
	// compare/verify it to the one in the db before changing the users password
}
