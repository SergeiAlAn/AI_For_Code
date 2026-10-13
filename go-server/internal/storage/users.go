package storage

import (
	"context"
	"database/sql"
	"errors"

	"ai-for-code/go-server/internal/models"
)

var ErrNotFound = errors.New("not found")

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, name string) (models.User, error) {
	res, err := r.db.ExecContext(ctx, "INSERT INTO users (name) VALUES (?)", name)
	if err != nil {
		return models.User{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return models.User{}, err
	}
	return models.User{ID: id, Name: name}, nil
}

// GetByID returns ErrNotFound if there is no user with the given id.
func (r *UserRepository) GetByID(ctx context.Context, id int64) (models.User, error) {
	var u models.User
	err := r.db.QueryRowContext(ctx, "SELECT id, name FROM users WHERE id = ?", id).Scan(&u.ID, &u.Name)
	if errors.Is(err, sql.ErrNoRows) {
		return models.User{}, ErrNotFound
	}
	return u, err
}

func (r *UserRepository) Exists(ctx context.Context, id int64) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE id = ?)", id).Scan(&exists)
	return exists, err
}
