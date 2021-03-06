package handlers

import (
	"net/http"

	"github.com/justinas/nosurf"
)

func (h *Handlers) ShowCachePage(w http.ResponseWriter, r *http.Request) {
	err := h.render(w, r, "cache", nil, nil)
	if err != nil {
		h.App.ErrorLog.Println("error rendering:", err)
	}
}

func (h *Handlers) SaveInCache(w http.ResponseWriter, r *http.Request) {

	var ok = true
	var msg string

	var userInput struct {
		Name  string `json:"name"`
		Value string `json:"value"`
		CSRF  string `json:"csrf_token"`
	}

	// read json from client
	err := h.App.ReadJSON(w, r, &userInput)
	if err != nil {
		ok = false
		msg = "Error reading json"
		h.App.Error404(w, r)
	}

	// verify csrf token
	if !nosurf.VerifyToken(nosurf.Token(r), userInput.CSRF) {
		h.App.Error500(w, r)
		return
	}

	// set value in cache
	err = h.App.Cache.Set(userInput.Name, userInput.Value)
	if err != nil {
		ok = false
		msg = "Error setting values in cache"
		h.App.Error500(w, r)
	}

	var resp struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}

	if !ok {
		resp.Error = true
		resp.Message = msg
	} else {
		resp.Error = false
		resp.Message = "Saved in cache"
	}

	_ = h.App.WriteJSON(w, http.StatusCreated, resp)
}

func (h *Handlers) GetFromCache(w http.ResponseWriter, r *http.Request) {
	var userInput struct {
		Name string `json:"name"`
		CSRF string `json:"csrf_token"`
	}

	var msg string
	var inCache = true

	err := h.App.ReadJSON(w, r, &userInput)
	if err != nil {
		h.App.Error404(w, r)
		return
	}

	// verify csrf token
	if !nosurf.VerifyToken(nosurf.Token(r), userInput.CSRF) {
		h.App.Error500(w, r)
		return
	}

	// save user input
	fromCache, err := h.App.Cache.Get(userInput.Name)
	if err != nil {
		msg = "Not found in cache"
		inCache = false
	}

	var resp struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
		Value   string `json:"value"`
	}

	if inCache {
		resp.Error = false
		resp.Message = "Found in cache"
		resp.Value = fromCache.(string)
	} else {
		resp.Error = true
		resp.Message = msg
	}

	// write json back to user
	_ = h.App.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handlers) DeleteFromCache(w http.ResponseWriter, r *http.Request) {
	var userInput struct {
		Name string `json:"name"`
		CSRF string `json:"csrf_token"`
	}

	err := h.App.ReadJSON(w, r, &userInput)
	if err != nil {
		h.App.Error404(w, r)
		return
	}

	// verify csrf token
	if !nosurf.VerifyToken(nosurf.Token(r), userInput.CSRF) {
		h.App.Error500(w, r)
		return
	}

	err = h.App.Cache.Forget(userInput.Name)
	if err != nil {
		h.App.Error500(w, r)
		return
	}

	var resp struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}

	resp.Error = false
	resp.Message = "Deleted from cache (if it existed)"

	_ = h.App.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handlers) EmptyCache(w http.ResponseWriter, r *http.Request) {
	var userInput struct {
		CSRF string `json:"csrf_token"`
	}

	err := h.App.ReadJSON(w, r, &userInput)
	if err != nil {
		h.App.Error404(w, r)
		return
	}

	// verify csrf token
	if !nosurf.VerifyToken(nosurf.Token(r), userInput.CSRF) {
		h.App.Error500(w, r)
		return
	}

	err = h.App.Cache.Empty()
	if err != nil {
		h.App.Error500(w, r)
		return
	}

	var resp struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}

	resp.Error = false
	resp.Message = "Cache dumped"

	_ = h.App.WriteJSON(w, http.StatusOK, resp)
}
