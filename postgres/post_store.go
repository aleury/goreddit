package postgres

import (
	"fmt"

	"github.com/aleury/goreddit"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type PostStore struct {
	*sqlx.DB
}

func (s *PostStore) Post(id uuid.UUID) (goreddit.Post, error) {
	var p goreddit.Post

	err := s.Get(&p, `SELECT * FROM posts WHERE id = $1`, id)
	if err != nil {
		return goreddit.Post{}, fmt.Errorf("error getting post: %w", err)
	}

	return p, nil
}

func (s *PostStore) Posts() ([]goreddit.Post, error) {
	var pp []goreddit.Post
	var query string = `
		SELECT
			posts.*,
			threads.title as thread_title,
			COUNT(comments.*) as comments_count
		FROM posts
		LEFT JOIN threads ON threads.id = posts.thread_id
		LEFT JOIN comments ON comments.post_id = posts.id
		GROUP BY posts.id, threads.title
		ORDER BY posts.votes DESC
	`

	err := s.Select(&pp, query)
	if err != nil {
		return []goreddit.Post{}, fmt.Errorf("error getting posts: %w", err)
	}

	return pp, nil
}

func (s *PostStore) PostsByThread(threadID uuid.UUID) ([]goreddit.Post, error) {
	var pp []goreddit.Post
	var query string = `
		SELECT
			posts.*,
			COUNT(comments.*) as comments_count
		FROM posts
		LEFT JOIN comments ON comments.post_id = posts.id
		WHERE posts.thread_id = $1
		GROUP BY posts.id
		ORDER BY posts.votes DESC
	`

	err := s.Select(&pp, query, threadID)
	if err != nil {
		return []goreddit.Post{}, fmt.Errorf("error getting posts: %w", err)
	}

	return pp, nil
}

func (s *PostStore) CreatePost(p *goreddit.Post) error {
	query := `INSERT INTO posts VALUES ($1, $2, $3, $4, $5) RETURNING *`

	err := s.Get(p, query, p.ID, p.ThreadID, p.Title, p.Content, p.Votes)
	if err != nil {
		return fmt.Errorf("error creating post: %w", err)
	}

	return nil
}

func (s *PostStore) UpdatePost(p *goreddit.Post) error {
	query := `UPDATE posts SET thread_id = $1, title = $2, content = $3, votes = $4 WHERE id = $5 RETURNING *`

	err := s.Get(p, query, p.ThreadID, p.Title, p.Content, p.Votes, p.ID)
	if err != nil {
		return fmt.Errorf("error updating post: %w", err)
	}

	return nil
}

func (s *PostStore) DeletePost(id uuid.UUID) error {
	_, err := s.Exec(`DELETE FROM posts WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("error deleting post: %w", err)
	}
	return nil
}
