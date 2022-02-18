package postgres

import (
	"fmt"

	"github.com/aleury/goreddit"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type ThreadStore struct {
	*sqlx.DB
}

func (s *ThreadStore) Thread(id uuid.UUID) (goreddit.Thread, error) {
	var t goreddit.Thread

	err := s.Get(&t, `SELECT * FROM threads WHERE id = $1`, id)
	if err != nil {
		return goreddit.Thread{}, fmt.Errorf("error getting thread: %w", err)
	}

	return t, nil
}

func (s *ThreadStore) Threads() ([]goreddit.Thread, error) {
	var tt []goreddit.Thread

	err := s.Select(&tt, `SELECT * FROM threads`)
	if err != nil {
		return []goreddit.Thread{}, fmt.Errorf("error getting threads: %w", err)
	}

	return tt, nil
}

func (s *ThreadStore) CreateThread(t *goreddit.Thread) error {
	query := `INSERT INTO threads VALUES ($1, $2, $3) RETURNING *`

	err := s.Get(t, query, t.ID, t.Title, t.Description)
	if err != nil {
		return fmt.Errorf("error creating thread: %w", err)
	}

	return nil
}

func (s *ThreadStore) UpdateThread(t *goreddit.Thread) error {
	query := `UPDATE threads SET title = $1, description = $2 WHERE id = $3 RETURNING *`

	err := s.Get(t, query, t.Title, t.Description, t.ID)
	if err != nil {
		return fmt.Errorf("error updating thread: %w", err)
	}

	return nil
}

func (s *ThreadStore) DeleteThread(id uuid.UUID) error {
	_, err := s.Exec(`DELETE FROM threads WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("error deleting thread: %w", err)
	}
	return nil
}
