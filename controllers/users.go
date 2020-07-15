package controllers

import (
	"fmt"
	"net/http"
	"time"

	"sibsiu.ru/context"
	"sibsiu.ru/models"
	"sibsiu.ru/rand"
	"sibsiu.ru/views"
)

type SignupForm struct {
	Name     string `schema:"name"`
	Email    string `schema:"email"`
	Password string `schema:"password"`
}

type LoginForm struct {
	Login      string `schema:"login"`
	Password   string `schema:"password"`
	RememberMe string `schema:"remember"`
}

type Users struct {
	NewView   *views.View
	LoginView *views.View
	us        models.UserService
}

func NewUsers(us models.UserService) *Users {
	return &Users{
		NewView:   views.NewView("bootstrap", "users/new"),
		LoginView: views.NewView("bootstrap", "users/login"),
		us:        us,
	}
}

func (u *Users) New(w http.ResponseWriter, r *http.Request) {
	u.NewView.Render(w, r, nil)
}

// Logout is used to delete a user's session cookie
// and invalidate their current remember token, which will
// sign the current user out.
//
// POST /logout
func (u *Users) Logout(w http.ResponseWriter, r *http.Request) {
	// First expire the user's cookie
	cookie := http.Cookie{
		Name:     "remember_token",
		Value:    "",
		Expires:  time.Now(),
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)
	// Then we update the user with a new remember token
	user := context.User(r.Context())
	// We are ignoring errors for now because they are
	// unlikely, and even if they do occur we can't recover
	// now that the user doesn't have a valid cookie
	token, _ := rand.RememberToken()
	user.Remember = token
	u.us.Update(user)
	// Finally send the user to the home page
	http.Redirect(w, r, "/", http.StatusFound)
}

// todo old implementation, it doesn't work properly
// Create is used to process the signup form when a user
// tries to create a new user account.
//
// POST /signup
func (u *Users) Create(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	var form SignupForm
	if err := parseForm(r, &form); err != nil {
		vd.SetAlert(err)
		u.NewView.Render(w, r, vd)
		return
	}

	//test
	user := models.User{
		Login:      "levankov_nv",
		FirstName:  "Никита",
		MiddleName: "Владимирович",
		LastName:   "Леванков",
		Email:      form.Email,
		Password:   form.Password,
		Class:      "Студент",
	}
	if err := u.us.Create(&user); err != nil {
		vd.SetAlert(err)
		u.NewView.Render(w, r, vd)
		return
	}

	err := u.signIn(w, &user, false)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	alert := views.Alert{
		Level:   views.AlertLvlSuccess,
		Message: "created",
	}
	views.RedirectAlert(w, r, "/", http.StatusFound, alert)
}

// Login is used to process the login form when a user
// tries to log in as an existing user (via login & pw).
//
// POST /login
func (u *Users) Login(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	var form LoginForm
	if err := parseForm(r, &form); err != nil {
		vd.SetAlert(err)
		u.LoginView.Render(w, r, vd)
		return
	}

	user, err := u.us.Authenticate(form.Login, form.Password)
	if err != nil {
		switch err {
		case models.ErrNotFound:
			vd.AlertError("No user exists with that login")
		default:
			vd.SetAlert(err)
		}
		u.LoginView.Render(w, r, vd)
		return
	}

	if form.RememberMe != "" {
		err = u.signIn(w, user, true)
	} else {
		err = u.signIn(w, user, false)
	}
	if err != nil {
		vd.SetAlert(err)
		u.LoginView.Render(w, r, vd)
		return
	}

	alert := views.Alert{
		Level:   views.AlertLvlSuccess,
		Message: "logged in",
	}

	if user.Class == "Студент" || user.Class == "Староста" {
		views.RedirectAlert(w, r, "/scores", http.StatusFound, views.Alert{})
		return
	}

	views.RedirectAlert(w, r, "/", http.StatusFound, alert)
}

// signIn is used to sign the given user in via cookies
func (u *Users) signIn(w http.ResponseWriter, user *models.User, rememberUser bool) error {
	if user.Remember == "" {
		token, err := rand.RememberToken()
		if err != nil {
			return err
		}
		user.Remember = token
		err = u.us.Update(user)
		if err != nil {
			return err
		}
	}

	var cookie http.Cookie

	if rememberUser {
		cookie = http.Cookie{
			Name:     "remember_token",
			Value:    user.Remember,
			Expires:  time.Now().Add(time.Hour * 3),
			HttpOnly: true,
		}
	} else {
		cookie = http.Cookie{
			Name:     "remember_token",
			Value:    user.Remember,
			HttpOnly: true,
		}
	}

	http.SetCookie(w, &cookie)
	return nil
}

func (u *Users) CookieTest(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("remember_token")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	user, err := u.us.ByRemember(cookie.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintln(w, user)

}
