package web

import (
	"net/http"

	"github.com/aleury/goreddit"
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type CommentHandler struct {
	store    goreddit.Store
	sessions *scs.SessionManager
}

func (h *CommentHandler) Create() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		form := CreateCommentForm{
			Content: r.FormValue("content"),
		}
		if !form.Validate() {
			h.sessions.Put(r.Context(), "form", form)
			http.Redirect(rw, r, r.Referer(), http.StatusFound)
			return
		}

		postId, err := uuid.Parse(chi.URLParam(r, "postId"))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}

		p, err := h.store.Post(postId)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		err = h.store.CreateComment(&goreddit.Comment{
			ID:      uuid.New(),
			PostID:  p.ID,
			Content: form.Content,
		})
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		h.sessions.Put(r.Context(), "flash", "Your comment has been submitted.")

		http.Redirect(rw, r, r.Referer(), http.StatusFound)
	}
}

func (h *CommentHandler) Vote() http.HandlerFunc {
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
