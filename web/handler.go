package web

import (
	"github.com/aleury/goreddit"
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/csrf"
)

type Handler struct {
	*chi.Mux
}

func NewHandler(store goreddit.Store, sessions *scs.SessionManager, csrfKey []byte) *Handler {
	h := &Handler{Mux: chi.NewMux()}

	pages := PageHandler{store: store, sessions: sessions}
	threads := ThreadHandler{store: store, sessions: sessions}
	posts := PostHandler{store: store, sessions: sessions}
	comments := CommentHandler{store: store, sessions: sessions}

	h.Use(middleware.Logger)

	// TODO: This is for development purposes only.
	// Set Secure flag to true when deploying to prod.
	h.Use(csrf.Protect(csrfKey, csrf.Secure(false)))

	h.Use(sessions.LoadAndSave)

	h.Get("/", pages.Home())
	h.Route("/threads", func(r chi.Router) {
		r.Get("/", threads.List())
		r.Get("/new", threads.New())
		r.Post("/", threads.Create())
		r.Get("/{id}", threads.Show())
		r.Post("/{id}/delete", threads.Delete())

		r.Get("/{threadId}/posts/new", posts.New())
		r.Post("/{threadId}/posts", posts.Create())
		r.Get("/{threadId}/posts/{postId}", posts.Show())
		r.Get("/{threadId}/posts/{postId}/vote", posts.Vote())

		r.Post("/{threadId}/posts/{postId}/comments", comments.Create())
	})
	h.Get("/comments/{id}/vote", comments.Vote())

	return h
}
