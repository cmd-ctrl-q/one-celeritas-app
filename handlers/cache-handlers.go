package handlers

import "net/http"

func (h *Handlers) ShowCachePage(w http.ResponseWriter, r *http.Request) {
	err := h.render(w, r, "cache", nil, nil)
	if err != nil {
		h.App.ErrorLog.Println("error rendering:", err)
	}
}

func (h *Handlers) SaveInCache(w http.ResponseWriter, r *http.Request) {
	var userInput struct {
		Name  string `json:"name"`
		Value string `json:"value"`
		CSRF  string `json:"csrf_token"`
	}

	// read json from client
	err := h.App.ReadJSON(w, r, &userInput)
	if err != nil {
		h.App.Error404(w, r)
		return
	}

	// set value in cache
	err = h.App.Cache.Set(userInput.Name, userInput.Value)
	if err != nil {
		h.App.Error404(w, r)
		return
	}

	var resp struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}

	resp.Error = false
	resp.Message = "Saved in cache"

	_ = h.App.WriteJSON(w, http.StatusCreated, resp)
}

func (h *Handlers) GetFromCache(w http.ResponseWriter, r *http.Request) {

}

func (h *Handlers) DeleteFromCache(w http.ResponseWriter, r *http.Request) {

}

func (h *Handlers) EmptyCache(w http.ResponseWriter, r *http.Request) {

}
