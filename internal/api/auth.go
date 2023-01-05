package api

import (
	"context"
	"net/http"
)

type ContextKey int

const (
	ContextKeyAccount ContextKey = iota + 1
)

func (a *API) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("api_key")
		if key == "" {
			a.writeError(w, 401, "Missing ?api_key query param")
			return
		}
		acc, err := a.db.GetAccountByAPIKey(key)
		if err != nil {
			a.writeError(w, 401, "Invalid API Key", err.Error())
			return
		}
		next.ServeHTTP(w, r.WithContext(
			context.WithValue(r.Context(), ContextKeyAccount, acc)),
		)
	})
}

func (a *API) optionalAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("api_key")
		if key == "" {
			next.ServeHTTP(w, r)
			return
		}
		acc, err := a.db.GetAccountByAPIKey(key)
		if err != nil {
			a.writeError(w, 401, "Invalid API Key", err.Error())
			return
		}
		next.ServeHTTP(w, r.WithContext(
			context.WithValue(r.Context(), ContextKeyAccount, acc)),
		)
	})
}
