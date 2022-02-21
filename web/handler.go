package web

import (
	"github.com/aleury/goreddit"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Handler struct {
	*chi.Mux
}

func NewHandler(store goreddit.Store) *Handler {
	h := &Handler{
		Mux: chi.NewMux(),
	}

	pages := PageHandler{store: store}
	threads := ThreadHandler{store: store}
	posts := PostHandler{store: store}
	comments := CommentHandler{store: store}

	h.Use(middleware.Logger)

	h.Get("/", pages.Home())

	// Threads
	h.Route("/threads", func(r chi.Router) {
		r.Get("/", threads.List())
		r.Get("/new", threads.New())
		r.Post("/", threads.Create())
		r.Get("/{id}", threads.Show())
		r.Post("/{id}/delete", threads.Delete())

		// Posts
		r.Get("/{threadId}/posts/new", posts.New())
		r.Post("/{threadId}/posts", posts.Create())
		r.Get("/{threadId}/posts/{postId}", posts.Show())
		r.Get("/{threadId}/posts/{postId}/vote", posts.Vote())

		// Comments
		r.Post("/{threadId}/posts/{postId}/comments", comments.Create())
	})

	// Comments
	h.Get("/comments/{id}/vote", comments.Vote())

	return h
}
