package web

import (
	"html/template"
	"net/http"

	"github.com/aleury/goreddit"
	"github.com/alexedwards/scs/v2"
)

type PageHandler struct {
	store    goreddit.Store
	sessions *scs.SessionManager
}

func (h *PageHandler) Home() http.HandlerFunc {
	type data struct {
		SessionData

		Posts []goreddit.Post
	}

	tmpl := template.Must(template.ParseFiles(
		"templates/layout.html",
		"templates/home.html",
	))
	return func(rw http.ResponseWriter, r *http.Request) {
		pp, err := h.store.Posts()
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		tmpl.Execute(rw, data{
			Posts:       pp,
			SessionData: GetSessionData(h.sessions, r.Context()),
		})
	}
}
