// convenience.go contains aliases for http methods
package main

import "net/http"

func (a *application) get(s string, h http.HandlerFunc) {
	a.App.Routes.Get(s, h)
}

func (a *application) post(s string, h http.HandlerFunc) {
	a.App.Routes.Get(s, h)
}

func (a *application) use(m ...func(http.Handler) http.Handler) {
	a.App.Routes.Use(m...)
}