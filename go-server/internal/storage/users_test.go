package storage

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
)

func newRepo(t *testing.T) *UserRepository {
	t.Helper()
	db, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return NewUserRepository(db)
}

func TestCreateAndGet(t *testing.T) {
	repo := newRepo(t)
	ctx := context.Background()
	name := "x'); DROP TABLE users;--"

	created, err := repo.Create(ctx, name)
	if err != nil {
		t.Fatal(err)
	}
	got, err := repo.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got != created || got.Name != name {
		t.Fatalf("got %+v, want %+v", got, created)
	}

	second, err := repo.Create(ctx, "second")
	if err != nil {
		t.Fatal(err)
	}
	if second.ID == created.ID {
		t.Fatal("ids must be unique")
	}
}

func TestGetByIDNotFound(t *testing.T) {
	_, err := newRepo(t).GetByID(context.Background(), 42)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("got %v, want ErrNotFound", err)
	}
}

func TestExists(t *testing.T) {
	repo := newRepo(t)
	ctx := context.Background()
	u, err := repo.Create(ctx, "alice")
	if err != nil {
		t.Fatal(err)
	}

	for id, want := range map[int64]bool{u.ID: true, u.ID + 1: false} {
		got, err := repo.Exists(ctx, id)
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Errorf("Exists(%d) = %v, want %v", id, got, want)
		}
	}
}
