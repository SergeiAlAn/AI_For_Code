// Package handlers implements the HTTP endpoints.
package handlers

import (
	"context"
	"net/http"
	"strconv"

	"ai-for-code/go-server/internal/models"
)

// UserStore is implemented by storage.UserRepository. Lookups of a missing
// user must return storage.ErrNotFound.
type UserStore interface {
	Create(ctx context.Context, name string) (models.User, error)
	GetByID(ctx context.Context, id int64) (models.User, error)
	Exists(ctx context.Context, id int64) (bool, error)
}

// ActiveTracker is implemented by activity.Tracker.
type ActiveTracker interface {
	Mark(id int64) []int64
	List() []int64
}

type Handler struct {
	users  UserStore
	active ActiveTracker
}

func New(users UserStore, active ActiveTracker) *Handler {
	return &Handler{users: users, active: active}
}

func pathID(r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	return id, err == nil
}
