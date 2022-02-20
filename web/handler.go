package web

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/aleury/goreddit"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
)

type Handler struct {
	*chi.Mux

	store goreddit.Store
}

func NewHandler(store goreddit.Store) *Handler {
	h := &Handler{
		Mux:   chi.NewMux(),
		store: store,
	}

	h.Use(middleware.Logger)

	h.Get("/", h.Home())

	// Threads
	h.Route("/threads", func(r chi.Router) {
		r.Get("/", h.ThreadsList())
		r.Get("/new", h.ThreadsNew())
		r.Post("/", h.ThreadsCreate())
		r.Get("/{id}", h.ThreadsShow())
		r.Post("/{id}/delete", h.ThreadsDelete())

		// Posts
		r.Get("/{threadId}/posts/new", h.PostsNew())
		r.Post("/{threadId}/posts", h.PostsCreate())
		r.Get("/{threadId}/posts/{postId}", h.PostsShow())

		// Comments
		r.Post("/{threadId}/posts/{postId}/comments", h.CommentsCreate())
	})

	// Comments
	h.Get("/comments/{id}/vote", h.CommentsVote())

	return h
}

func (h *Handler) Home() http.HandlerFunc {
	tmpl := template.Must(template.ParseFiles(
		"templates/layout.html",
		"templates/home.html",
	))
	return func(rw http.ResponseWriter, r *http.Request) {
		tmpl.Execute(rw, nil)
	}
}

func (h *Handler) ThreadsList() http.HandlerFunc {
	type data struct {
		Threads []goreddit.Thread
	}

	tmpl := template.Must(template.ParseFiles(
		"templates/layout.html",
		"templates/threads.html",
	))
	return func(rw http.ResponseWriter, r *http.Request) {
		tt, err := h.store.Threads()
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		tmpl.Execute(rw, data{Threads: tt})
	}
}

func (h *Handler) ThreadsNew() http.HandlerFunc {
	tmpl := template.Must(template.ParseFiles(
		"templates/layout.html",
		"templates/thread_create.html",
	))
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl.Execute(w, nil)
	}
}

func (h *Handler) ThreadsCreate() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		title := r.FormValue("title")
		description := r.FormValue("description")

		err := h.store.CreateThread(&goreddit.Thread{
			ID:          uuid.New(),
			Title:       title,
			Description: description,
		})
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(rw, r, "/threads", http.StatusFound)
	}
}

func (h *Handler) ThreadsShow() http.HandlerFunc {
	type data struct {
		Thread goreddit.Thread
		Posts  []goreddit.Post
	}

	tmpl := template.Must(template.ParseFiles(
		"templates/layout.html",
		"templates/thread.html",
	))
	return func(rw http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}

		t, err := h.store.Thread(id)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		pp, err := h.store.PostsByThread(t.ID)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		tmpl.Execute(rw, data{Thread: t, Posts: pp})
	}
}

func (h *Handler) ThreadsDelete() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}

		err = h.store.DeleteThread(id)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(rw, r, "/threads", http.StatusFound)
	}
}

func (h *Handler) PostsNew() http.HandlerFunc {
	type data struct {
		Thread goreddit.Thread
	}

	tmpl := template.Must(template.ParseFiles(
		"templates/layout.html",
		"templates/post_create.html",
	))
	return func(rw http.ResponseWriter, r *http.Request) {
		threadId, err := uuid.Parse(chi.URLParam(r, "threadId"))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}

		t, err := h.store.Thread(threadId)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		tmpl.Execute(rw, data{Thread: t})
	}
}

func (h *Handler) PostsCreate() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		threadId, err := uuid.Parse(chi.URLParam(r, "threadId"))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}

		t, err := h.store.Thread(threadId)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		title := r.FormValue("title")
		content := r.FormValue("content")

		p := &goreddit.Post{
			ID:       uuid.New(),
			ThreadID: t.ID,
			Title:    title,
			Content:  content,
		}
		err = h.store.CreatePost(p)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		redirect_url := fmt.Sprintf("/threads/%s/posts/%s", t.ID.String(), p.ID.String())
		http.Redirect(rw, r, redirect_url, http.StatusFound)
	}
}

func (h *Handler) PostsShow() http.HandlerFunc {
	type data struct {
		Thread   goreddit.Thread
		Post     goreddit.Post
		Comments []goreddit.Comment
	}

	tmpl := template.Must(template.ParseFiles(
		"templates/layout.html",
		"templates/post.html",
	))
	return func(rw http.ResponseWriter, r *http.Request) {
		threadId, err := uuid.Parse(chi.URLParam(r, "threadId"))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}

		postId, err := uuid.Parse(chi.URLParam(r, "postId"))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}

		t, err := h.store.Thread(threadId)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		p, err := h.store.Post(postId)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		cc, err := h.store.CommentsbyPost(p.ID)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		tmpl.Execute(rw, data{Thread: t, Post: p, Comments: cc})
	}
}

func (h *Handler) CommentsCreate() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		threadId, err := uuid.Parse(chi.URLParam(r, "threadId"))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}

		postId, err := uuid.Parse(chi.URLParam(r, "postId"))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}

		_, err = h.store.Thread(threadId)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		p, err := h.store.Post(postId)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		content := r.FormValue("content")

		err = h.store.CreateComment(&goreddit.Comment{
			ID:      uuid.New(),
			PostID:  p.ID,
			Content: content,
		})
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(rw, r, r.Referer(), http.StatusFound)
	}
}

func (h *Handler) CommentsVote() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}

		c, err := h.store.Comment(id)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		switch r.URL.Query().Get("dir") {
		case "up":
			c.Votes++
		case "down":
			c.Votes--
		}

		err = h.store.UpdateComment(&c)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(rw, r, r.Referer(), http.StatusFound)
	}
}
