package middleware

import (
	"fmt"
	"myapp/data"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func (m *Middleware) CheckRemember(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		// check session for user id
		if !m.App.Session.Exists(r.Context(), "userID") {
			// user not logged in, check if cookie exists
			cookie, err := r.Cookie(fmt.Sprintf("_%s_remember", m.App.AppName))
			if err != nil {
				// no cookie, go to next middleware
				next.ServeHTTP(rw, r)
			} else {
				// found cookie
				key := cookie.Value
				var u data.User
				if len(key) > 0 {
					// cookie contains data, validate it
					split := strings.Split(key, "|")
					uid, hash := split[0], split[1]
					id, _ := strconv.Atoi(uid)
					// check if hash is valid
					validHash := u.CheckForRememberToken(id, hash)
					if !validHash {
						// invalid hash, expire it
						m.deleteRememberCookie(rw, r)
						m.App.Session.Put(r.Context(), "error", "You've been logged out from another device")
						next.ServeHTTP(rw, r)
					} else {
						// valid hash, log user in
						user, _ := u.Get(id)
						m.App.Session.Put(r.Context(), "userID", user.ID)
						m.App.Session.Put(r.Context(), "remember_token", hash)
						next.ServeHTTP(rw, r)
					}
				} else {
					// key length = 0, probably leftover cookie (user has not closed browser)
					m.deleteRememberCookie(rw, r)
					next.ServeHTTP(rw, r)
				}
			}
		} else {
			// user logged in
			next.ServeHTTP(rw, r)
		}
	})
}

func (m *Middleware) deleteRememberCookie(w http.ResponseWriter, r *http.Request) {
	_ = m.App.Session.RenewToken(r.Context())
	newCookie := http.Cookie{
		Name:     fmt.Sprintf("_%s_remember", m.App.AppName),
		Value:    "",
		Path:     "/",
		Expires:  time.Now().Add(-100 * time.Hour),
		HttpOnly: true,
		Domain:   m.App.Session.Cookie.Domain,
		MaxAge:   -1,
		Secure:   m.App.Session.Cookie.Secure,
		SameSite: http.SameSiteStrictMode,
	}

	// set cookie
	http.SetCookie(w, &newCookie)

	// log out user
	m.App.Session.Remove(r.Context(), "userID")
	m.App.Session.Destroy(r.Context())
	_ = m.App.Session.RenewToken(r.Context())
}
