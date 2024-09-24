package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/andreadebortoli2/GO-Live-Chat/internal/config"
	"github.com/andreadebortoli2/GO-Live-Chat/internal/database"
	"github.com/andreadebortoli2/GO-Live-Chat/internal/helpers"
	"github.com/andreadebortoli2/GO-Live-Chat/internal/models"
	"github.com/andreadebortoli2/GO-Live-Chat/internal/render"
)

var Repo *Repository

type Repository struct {
	App *config.AppConfig
}

func NewRepo(a *config.AppConfig) *Repository {
	return &Repository{
		App: a,
	}
}
func NewHandlers(r *Repository) {
	Repo = r
}

// ShowHomePage show home page
func (m *Repository) ShowHomePage(w http.ResponseWriter, r *http.Request) {
	render.RenderPage(w, r, "home", &render.TemplateData{})
}

// ShowLoginPage show login page
func (m *Repository) ShowLoginPage(w http.ResponseWriter, r *http.Request) {
	render.RenderPage(w, r, "login", &render.TemplateData{})
}

// PostLogin logic to login the user
func (m *Repository) PostLogin(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")

	fields := map[string]string{
		"email":    email,
		"password": password,
	}

	err = helpers.LoginValidation(fields)
	if err != nil {
		log.Println(err)
		strErr := err.Error()
		render.RenderPage(w, r, "login", &render.TemplateData{
			StringMap: fields,
			Error:     strErr,
		})
		return
	}

	user, err := database.Login(email, password)
	if err != nil {
		log.Println(err)
		strErr := err.Error()
		render.RenderPage(w, r, "login", &render.TemplateData{
			StringMap: fields,
			Error:     strErr,
		})
		return
	}

	// if a user is returned give the authorization saving auth level in the session
	if user != (models.User{}) {
		_ = m.App.Session.RenewToken(r.Context())

		m.App.Session.Put(r.Context(), "user", user)

		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	}
}

// ShowLogoutPage logic to logout the user
func (m *Repository) ShowLogoutPage(w http.ResponseWriter, r *http.Request) {
	_ = m.App.Session.Destroy(r.Context())
	_ = m.App.Session.RenewToken(r.Context())
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// ShowNewUserPage show new-user page
func (m *Repository) ShowNewUserPage(w http.ResponseWriter, r *http.Request) {
	render.RenderPage(w, r, "new-user", &render.TemplateData{})
}

// PostNewUserPage add new user to DB
func (m *Repository) PostNewUser(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/new-user", http.StatusSeeOther)
		return
	}

	userName := r.Form.Get("user_name")
	email := r.Form.Get("email")
	password := r.Form.Get("password")

	fields := map[string]string{
		"user_name": userName,
		"email":     email,
		"password":  password,
	}

	err = helpers.NewUserValidation(fields)
	if err != nil {
		log.Println(err)
		strErr := err.Error()
		render.RenderPage(w, r, "new-user", &render.TemplateData{
			StringMap: fields,
			Error:     strErr,
		})
		return
	}

	err = database.AddUser(userName, email, password)
	if err != nil {
		log.Println(err)
		strErr := err.Error()
		render.RenderPage(w, r, "new-user", &render.TemplateData{
			StringMap: fields,
			Error:     strErr,
		})
		return
	}

	// after correct registration immediatly login the user and redirect
	user, err := database.Login(email, password)
	if err != nil {
		log.Println(err)
		strErr := err.Error()
		render.RenderPage(w, r, "login", &render.TemplateData{
			StringMap: fields,
			Error:     strErr,
		})
		return
	}
	if user != (models.User{}) {
		_ = m.App.Session.RenewToken(r.Context())

		m.App.Session.Put(r.Context(), "user", user)

		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	}
}

// ShowDashboardPage show dashboard page
func (m *Repository) ShowDashboardPage(w http.ResponseWriter, r *http.Request) {
	render.RenderPage(w, r, "dashboard", &render.TemplateData{})
}

// ShowProfilePage show dashboard page
func (m *Repository) ShowProfilePage(w http.ResponseWriter, r *http.Request) {
	render.RenderPage(w, r, "profile", &render.TemplateData{})
}

// ShowAdminAllUsersPage show the administraation page with all the users
func (m *Repository) ShowAdminAllUsersPage(w http.ResponseWriter, r *http.Request) {
	users, err := database.GetAllUsers()
	if err != nil {
		log.Println(err)
	}
	datausers := make(map[string]interface{})
	for i, u := range users {
		index := strconv.Itoa(i)
		datausers[index] = u
	}
	data := make(map[string]interface{})
	data["users"] = datausers
	render.RenderPage(w, r, "admin-all-users", &render.TemplateData{
		Data: data,
	})
}

func (m *Repository) PostChangeAccessLevel(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/admin/all-users", http.StatusSeeOther)
		return
	}
	accLvl := r.Form.Get("moderator")
	userID := r.Form.Get("user-id")

	err = database.SetModerator(accLvl, userID)
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/admin/all-users", http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/admin/all-users", http.StatusSeeOther)
}

func (m *Repository) PostDeleteUser(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/admin/all-users", http.StatusSeeOther)
		return
	}

	userID := r.Form.Get("user-id")

	database.DeleteUserByID(userID)

	http.Redirect(w, r, "/admin/all-users", http.StatusSeeOther)
}
