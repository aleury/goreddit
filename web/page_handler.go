package web

import (
	"html/template"
	"net/http"

	"github.com/aleury/goreddit"
)

type PageHandler struct {
	store goreddit.Store
}

func (h *PageHandler) Home() http.HandlerFunc {
	type data struct {
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
		tmpl.Execute(rw, data{Posts: pp})
	}
}
