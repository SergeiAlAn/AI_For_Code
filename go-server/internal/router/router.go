// Package router maps URL paths to handlers.
package router

import (
	"net/http"

	"ai-for-code/go-server/internal/docs"
	"ai-for-code/go-server/internal/handlers"
	"ai-for-code/go-server/internal/middleware"
)

func New(h *handlers.Handler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /adduser", h.AddUser)
	mux.HandleFunc("GET /user/{id}", h.GetUser)
	mux.HandleFunc("POST /user/{id}/active", h.SetActive)
	mux.HandleFunc("GET /active", h.GetActive)

	mux.HandleFunc("GET /{$}", docs.RedirectToPage)
	mux.HandleFunc("GET /docs", docs.Page)
	mux.HandleFunc("GET /openapi.yaml", docs.Spec)
	return middleware.Logging(mux)
}
