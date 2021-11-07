package handlers

import (
	"context"
	"net/http"

	"github.com/cmd-ctrl-q/celeritas"
)

// render is an alias to render a template
func (h *Handlers) render(w http.ResponseWriter, r *http.Request, tmpl string, variables, data interface{}) error {
	return h.App.Render.Page(w, r, tmpl, variables, data)
}

// put is an alias to add key-value to a session
func (h *Handlers) put(ctx context.Context, key string, val interface{}) {
	h.App.Session.Put(ctx, key, val)
}

// sessionHas is an alias to check if a key exists in a session
func (h *Handlers) sessionHas(ctx context.Context, key string) bool {
	return h.App.Session.Exists(ctx, key)
}

// sessionGet is an alias to retrieve a value from a session
func (h *Handlers) sessionGet(ctx context.Context, key string) interface{} {
	return h.App.Session.Get(ctx, key)
}

// sessionRemove is an alias to remove key-value from session
func (h *Handlers) sessionRemove(ctx context.Context, key string) {
	h.App.Session.Remove(ctx, key)
}

// sessionRenew is an alias to renew a session token
func (h *Handlers) sessionRenew(ctx context.Context) error {
	return h.App.Session.RenewToken(ctx)
}

// sessionDestroy is an alias to destroy a session token
func (h *Handlers) sessionDestroy(ctx context.Context) error {
	return h.App.Session.Destroy(ctx)
}

// randomString is an alias to generate a random string
func (h *Handlers) randomString(n int) string {
	return h.App.RandomString(n)
}

func (h *Handlers) encrypt(text string) (string, error) {
	enc := celeritas.Encryption{Key: []byte(h.App.EncryptionKey)}

	encrypted, err := enc.Encrypt(text)
	if err != nil {
		return "", err
	}

	return encrypted, nil
}

func (h *Handlers) decrypt(crypto string) (string, error) {
	enc := celeritas.Encryption{Key: []byte(h.App.EncryptionKey)}

	encrypted, err := enc.Decrypt(crypto)
	if err != nil {
		return "", err
	}

	return encrypted, nil
}
