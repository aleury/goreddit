package web

import (
	"context"
	"net/http"

	"github.com/aleury/goreddit"
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/gorilla/csrf"
)

type Handler struct {
	*chi.Mux
	store    goreddit.Store
	sessions *scs.SessionManager
}

func NewHandler(store goreddit.Store, sessions *scs.SessionManager, csrfKey []byte) *Handler {
	h := &Handler{
		Mux:      chi.NewMux(),
		store:    store,
		sessions: sessions,
	}

	pages := PageHandler{store: store, sessions: sessions}
	threads := ThreadHandler{store: store, sessions: sessions}
	posts := PostHandler{store: store, sessions: sessions}
	comments := CommentHandler{store: store, sessions: sessions}
	users := UserHandler{store: store, sessions: sessions}

	h.Use(middleware.Logger)

	// TODO: This is for development purposes only.
	// Set Secure flag to true when deploying to prod.
	h.Use(csrf.Protect(csrfKey, csrf.Secure(false)))

	h.Use(sessions.LoadAndSave)
	h.Use(h.withUser)

	h.Get("/", pages.Home())
	h.Get("/register", users.Register())
	h.Post("/register", users.RegisterSubmit())
	h.Get("/login", users.Login())
	h.Post("/login", users.LoginSubmit())
	h.Get("/logout", users.Logout())

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

func (h *Handler) withUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		userId, _ := h.sessions.Get(r.Context(), "user_id").(uuid.UUID)

		user, err := h.store.User(userId)
		if err != nil {
			next.ServeHTTP(rw, r)
			return
		}

		ctx := context.WithValue(r.Context(), ctxKey("user"), user)
		next.ServeHTTP(rw, r.WithContext(ctx))
	})
}
