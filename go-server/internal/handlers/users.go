package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"ai-for-code/go-server/internal/models"
	"ai-for-code/go-server/internal/respond"
	"ai-for-code/go-server/internal/storage"
	"ai-for-code/go-server/internal/validation"
)

const maxBodyBytes = 1 << 20

var errBodyNotObject = errors.New("request body must be a JSON object")

func (h *Handler) AddUser(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	req, err := decodeCreateUser(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	name, err := validation.UserName(req.Name)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	user, err := h.users.Create(r.Context(), name)
	if err != nil {
		respond.InternalError(w, err)
		return
	}
	respond.JSON(w, http.StatusCreated, user)
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r)
	if !ok {
		respond.Error(w, http.StatusNotFound, "user not found")
		return
	}

	user, err := h.users.GetByID(r.Context(), id)
	if errors.Is(err, storage.ErrNotFound) {
		respond.Error(w, http.StatusNotFound, "user not found")
		return
	}
	if err != nil {
		respond.InternalError(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, user)
}

// decodeCreateUser requires a JSON object body; a missing or null "name"
// decodes to "" and is rejected later by validation.
func decodeCreateUser(r *http.Request) (models.CreateUserRequest, error) {
	var fields map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&fields); err != nil || fields == nil {
		return models.CreateUserRequest{}, errBodyNotObject
	}

	var req models.CreateUserRequest
	if raw, ok := fields["name"]; ok {
		if err := json.Unmarshal(raw, &req.Name); err != nil {
			return models.CreateUserRequest{}, validation.ErrNameEmpty
		}
	}
	return req, nil
}
