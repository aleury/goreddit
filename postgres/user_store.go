package postgres

import (
	"fmt"

	"github.com/aleury/goreddit"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type UserStore struct {
	*sqlx.DB
}

func (s *UserStore) User(id uuid.UUID) (goreddit.User, error) {
	var t goreddit.User

	err := s.Get(&t, `SELECT * FROM users WHERE id = $1`, id)
	if err != nil {
		return goreddit.User{}, fmt.Errorf("error getting user: %w", err)
	}

	return t, nil
}

func (s *UserStore) UserByUsername(username string) (goreddit.User, error) {
	var u goreddit.User

	err := s.Get(&u, `SELECT * FROM users WHERE username = $1`, username)
	if err != nil {
		return goreddit.User{}, fmt.Errorf("error getting user: %w", err)
	}

	return u, nil
}

func (s *UserStore) CreateUser(u *goreddit.User) error {
	query := `INSERT INTO users VALUES ($1, $2, $3) RETURNING *`

	err := s.Get(u, query, u.ID, u.Username, u.Password)
	if err != nil {
		return fmt.Errorf("error creating user: %w", err)
	}

	return nil
}

func (s *UserStore) UpdateUser(u *goreddit.User) error {
	query := `UPDATE users SET username = $1, password = $2 WHERE id = $3 RETURNING *`

	err := s.Get(u, query, u.Username, u.Password, u.ID)
	if err != nil {
		return fmt.Errorf("error updating user: %w", err)
	}

	return nil
}

func (s *UserStore) DeleteUser(id uuid.UUID) error {
	_, err := s.Exec(`DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("error deleting user: %w", err)
	}
	return nil
}
