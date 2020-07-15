package middleware

import (
	"net/http"
	"strings"

	"sibsiu.ru/context"
	"sibsiu.ru/models"
)

// User middleware will lookup the current user via their
// remember_token cookie using the UserService. If the user
// is found, they will be set on the request context.
// Regardless, the next handler is always called.
// Ищет юзера по remember token-у и прикрепляет его к контексту
type User struct {
	models.UserService
}

func (mw *User) Apply(next http.Handler) http.HandlerFunc {
	return mw.ApplyFn(next.ServeHTTP)
}
func (mw *User) ApplyFn(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		// If the user is requesting a static asset or image
		// we will not need to lookup the current user so we skip
		// doing that.
		if strings.HasPrefix(path, "/assets/") ||
			strings.HasPrefix(path, "/images/") {
			next(w, r)
			return
		}

		cookie, err := r.Cookie("remember_token")
		if err != nil {
			next(w, r)
			return
		}

		user, err := mw.UserService.ByRemember(cookie.Value)
		if err != nil {
			next(w, r)
			return
		}

		ctx := r.Context()
		ctx = context.WithUser(ctx, user)
		r = r.WithContext(ctx)
		next(w, r)
	})
}

// RequireUser will redirect a user to the /login page
// if they are not logged in. This middleware assumes
// that User middleware has already been run, otherwise
// it will always redirect users.
type RequireUser struct{}

func (mw *RequireUser) Apply(next http.Handler) http.HandlerFunc {
	return mw.ApplyFn(next.ServeHTTP)
}

func (mw *RequireUser) ApplyFn(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := context.User(r.Context())
		if user == nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		next(w, r)
	})
}

// RequireClasses реализует механизм авторизации.
// This middleware assumes that RequireUser middleware has already been run.
type RequireClasses struct{}

func (mw *RequireClasses) Apply(next http.Handler) http.HandlerFunc {
	return mw.ApplyFn(next.ServeHTTP)
}

func (mw *RequireClasses) ApplyFn(next http.HandlerFunc, classes ...string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := context.User(r.Context())

		for _, class := range classes {
			if class == user.Class {
				next(w, r)
				return
			}
		}

		http.Error(w, "У вас нет права на выполнение данного действия.", http.StatusUnauthorized)
	})
}
