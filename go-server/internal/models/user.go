// Package models contains the data types shared between the storage and HTTP layers.
package models

type User struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type CreateUserRequest struct {
	Name string `json:"name"`
}

type ActiveUsersResponse struct {
	Active []int64 `json:"active"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
