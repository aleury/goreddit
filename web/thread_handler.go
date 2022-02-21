package web

import (
	"html/template"
	"net/http"

	"github.com/aleury/goreddit"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ThreadHandler struct {
	store goreddit.Store
}

func (h *ThreadHandler) List() http.HandlerFunc {
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

func (h *ThreadHandler) New() http.HandlerFunc {
	tmpl := template.Must(template.ParseFiles(
		"templates/layout.html",
		"templates/thread_create.html",
	))
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl.Execute(w, nil)
	}
}

func (h *ThreadHandler) Create() http.HandlerFunc {
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

func (h *ThreadHandler) Show() http.HandlerFunc {
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

func (h *ThreadHandler) Delete() http.HandlerFunc {
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
