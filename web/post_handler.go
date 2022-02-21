package web

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/aleury/goreddit"
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gorilla/csrf"
)

type PostHandler struct {
	store    goreddit.Store
	sessions *scs.SessionManager
}

func (h *PostHandler) New() http.HandlerFunc {
	type data struct {
		SessionData

		CSRF   template.HTML
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

		tmpl.Execute(rw, data{
			Thread:      t,
			CSRF:        csrf.TemplateField(r),
			SessionData: GetSessionData(h.sessions, r.Context()),
		})
	}
}

func (h *PostHandler) Create() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		form := CreatePostForm{
			Title:   r.FormValue("title"),
			Content: r.FormValue("content"),
		}
		if !form.Validate() {
			h.sessions.Put(r.Context(), "form", form)
			http.Redirect(rw, r, r.Referer(), http.StatusFound)
			return
		}

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

		p := &goreddit.Post{
			ID:       uuid.New(),
			ThreadID: t.ID,
			Title:    form.Title,
			Content:  form.Content,
		}
		err = h.store.CreatePost(p)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		h.sessions.Put(r.Context(), "flash", "Your new post has been created.")

		redirect_url := fmt.Sprintf("/threads/%s/posts/%s", t.ID.String(), p.ID.String())
		http.Redirect(rw, r, redirect_url, http.StatusFound)
	}
}

func (h *PostHandler) Show() http.HandlerFunc {
	type data struct {
		SessionData

		CSRF     template.HTML
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

		tmpl.Execute(rw, data{
			Thread:      t,
			Post:        p,
			Comments:    cc,
			CSRF:        csrf.TemplateField(r),
			SessionData: GetSessionData(h.sessions, r.Context()),
		})
	}
}

func (h *PostHandler) Vote() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
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

		switch r.URL.Query().Get("dir") {
		case "up":
			p.Votes++
		case "down":
			p.Votes--
		}

		err = h.store.UpdatePost(&p)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(rw, r, r.Referer(), http.StatusFound)
	}
}
