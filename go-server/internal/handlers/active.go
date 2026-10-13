package handlers

import (
	"net/http"

	"ai-for-code/go-server/internal/models"
	"ai-for-code/go-server/internal/respond"
)

func (h *Handler) SetActive(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r)
	if !ok {
		respond.Error(w, http.StatusNotFound, "user not found")
		return
	}

	exists, err := h.users.Exists(r.Context(), id)
	if err != nil {
		respond.InternalError(w, err)
		return
	}
	if !exists {
		respond.Error(w, http.StatusNotFound, "user not found")
		return
	}
	respond.JSON(w, http.StatusOK, models.ActiveUsersResponse{Active: h.active.Mark(id)})
}

func (h *Handler) GetActive(w http.ResponseWriter, r *http.Request) {
	respond.JSON(w, http.StatusOK, models.ActiveUsersResponse{Active: h.active.List()})
}
