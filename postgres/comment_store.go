package postgres

import (
	"fmt"

	"github.com/aleury/goreddit"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type CommentStore struct {
	*sqlx.DB
}

func (s *CommentStore) Comment(id uuid.UUID) (goreddit.Comment, error) {
	var c goreddit.Comment

	err := s.Get(&c, `SELECT * FROM comments WHERE id = $1`, id)
	if err != nil {
		return goreddit.Comment{}, fmt.Errorf("error getting comment: %w", err)
	}

	return c, nil
}

func (s *CommentStore) CommentsbyPost(postID uuid.UUID) ([]goreddit.Comment, error) {
	var cc []goreddit.Comment

	err := s.Select(&cc, `SELECT * FROM comments WHERE post_id = $1 ORDER BY votes DESC`, postID)
	if err != nil {
		return []goreddit.Comment{}, fmt.Errorf("error getting comments: %w", err)
	}

	return cc, nil
}

func (s *CommentStore) CreateComment(c *goreddit.Comment) error {
	query := `INSERT INTO comments VALUES ($1, $2, $3, $4) RETURNING *`

	err := s.Get(c, query, c.ID, c.PostID, c.Content, c.Votes)
	if err != nil {
		return fmt.Errorf("error creating comment: %w", err)
	}

	return nil
}

func (s *CommentStore) UpdateComment(c *goreddit.Comment) error {
	query := `UPDATE comments SET post_id = $1, content = $2, votes = $3 WHERE id = $4 RETURNING *`

	err := s.Get(c, query, c.PostID, c.Content, c.Votes, c.ID)
	if err != nil {
		return fmt.Errorf("error updating comment: %w", err)
	}

	return nil
}

func (s *CommentStore) DeleteComment(id uuid.UUID) error {
	_, err := s.Exec(`DELETE FROM comments WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("error deleting comment: %w", err)
	}
	return nil
}
