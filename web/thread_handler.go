package web

import (
	"html/template"
	"net/http"

	"github.com/aleury/goreddit"
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gorilla/csrf"
)

type ThreadHandler struct {
	store    goreddit.Store
	sessions *scs.SessionManager
}

func (h *ThreadHandler) List() http.HandlerFunc {
	type data struct {
		SessionData

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

		tmpl.Execute(rw, data{
			Threads:     tt,
			SessionData: GetSessionData(h.sessions, r.Context()),
		})
	}
}

func (h *ThreadHandler) New() http.HandlerFunc {
	type data struct {
		SessionData

		CSRF template.HTML
	}

	tmpl := template.Must(template.ParseFiles(
		"templates/layout.html",
		"templates/thread_create.html",
	))
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl.Execute(w, data{
			CSRF:        csrf.TemplateField(r),
			SessionData: GetSessionData(h.sessions, r.Context()),
		})
	}
}

func (h *ThreadHandler) Create() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		form := CreateThreadForm{
			Title:       r.FormValue("title"),
			Description: r.FormValue("description"),
		}
		if !form.Validate() {
			h.sessions.Put(r.Context(), "form", form)
			http.Redirect(rw, r, r.Referer(), http.StatusFound)
			return
		}

		err := h.store.CreateThread(&goreddit.Thread{
			ID:          uuid.New(),
			Title:       form.Title,
			Description: form.Description,
		})
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		h.sessions.Put(r.Context(), "flash", "You new thread has been created.")

		http.Redirect(rw, r, "/threads", http.StatusFound)
	}
}

func (h *ThreadHandler) Show() http.HandlerFunc {
	type data struct {
		SessionData

		CSRF   template.HTML
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

		tmpl.Execute(rw, data{
			Thread:      t,
			Posts:       pp,
			CSRF:        csrf.TemplateField(r),
			SessionData: GetSessionData(h.sessions, r.Context()),
		})
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

		h.sessions.Put(r.Context(), "flash", "The thread has been deleted.")

		http.Redirect(rw, r, "/threads", http.StatusFound)
	}
}
