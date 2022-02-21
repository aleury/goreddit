package web

import (
	"html/template"
	"net/http"

	"github.com/aleury/goreddit"
	"github.com/alexedwards/scs/v2"
	"github.com/google/uuid"
	"github.com/gorilla/csrf"
	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct {
	store    goreddit.Store
	sessions *scs.SessionManager
}

func (h *UserHandler) Register() http.HandlerFunc {
	type data struct {
		SessionData
		CSRF template.HTML
	}

	tmpl := template.Must(template.ParseFiles(
		"templates/layout.html",
		"templates/user_register.html",
	))
	return func(rw http.ResponseWriter, r *http.Request) {
		tmpl.Execute(rw, data{
			CSRF:        csrf.TemplateField(r),
			SessionData: GetSessionData(h.sessions, r.Context()),
		})
	}
}

func (h *UserHandler) RegisterSubmit() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		form := RegisterForm{
			Username:      r.FormValue("username"),
			Password:      r.FormValue("password"),
			UsernameTaken: false,
		}
		if _, err := h.store.UserByUsername(form.Username); err == nil {
			form.UsernameTaken = true
		}
		if !form.Validate() {
			h.sessions.Put(r.Context(), "form", form)
			http.Redirect(rw, r, r.Referer(), http.StatusFound)
			return
		}

		password, err := bcrypt.GenerateFromPassword([]byte(form.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		err = h.store.CreateUser(&goreddit.User{
			ID:       uuid.New(),
			Username: form.Username,
			Password: string(password),
		})
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		h.sessions.Put(r.Context(), "flash", "Your registration was successful. Please log in.")
		http.Redirect(rw, r, "/", http.StatusFound)
	}
}

func (h *UserHandler) Login() http.HandlerFunc {
	type data struct {
		SessionData
		CSRF template.HTML
	}

	tmpl := template.Must(template.ParseFiles(
		"templates/layout.html",
		"templates/user_login.html",
	))
	return func(rw http.ResponseWriter, r *http.Request) {
		tmpl.Execute(rw, data{
			CSRF:        csrf.TemplateField(r),
			SessionData: GetSessionData(h.sessions, r.Context()),
		})
	}
}

func (h *UserHandler) LoginSubmit() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		form := LoginForm{
			Username:             r.FormValue("username"),
			Password:             r.FormValue("password"),
			IncorrectCredentials: false,
		}
		user, err := h.store.UserByUsername(form.Username)
		if err != nil {
			form.IncorrectCredentials = true
		} else {
			compareErr := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(form.Password))
			form.IncorrectCredentials = compareErr != nil
		}
		if !form.Validate() {
			h.sessions.Put(r.Context(), "form", form)
			http.Redirect(rw, r, r.Referer(), http.StatusFound)
			return
		}

		h.sessions.Put(r.Context(), "user_id", user.ID)
		h.sessions.Put(r.Context(), "flash", "You have been logged in successfully. Welcome back!")
		http.Redirect(rw, r, "/", http.StatusFound)
	}
}

func (h *UserHandler) Logout() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		h.sessions.Remove(r.Context(), "user_id")
		h.sessions.Put(r.Context(), "flash", "You have been logged out successfully.")
		http.Redirect(rw, r, "/", http.StatusFound)
	}
}
