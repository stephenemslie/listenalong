package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/csrf"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

type key int

var (
	baseTemplate *template.Template
	sessionStore *sessions.CookieStore
	sessionKey   key
)

func sessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := sessionStore.Get(r, "listenalong")
		if session.IsNew {
			session.Save(r, w)
		}
		r = r.WithContext(context.WithValue(r.Context(), sessionKey, session))
		next.ServeHTTP(w, r)
	})
}

func init() {
	secret := os.Getenv("SECRET_KEY")
	sessionStore = sessions.NewCookieStore([]byte(secret))
	baseTemplate = template.New("base.html")
	var err error
	baseTemplate, err = baseTemplate.ParseGlob("templates/layout/*.html")
	if err != nil {
		fmt.Println(err)
	}
}

func main() {
	r := mux.NewRouter()
	r = r.StrictSlash(true)
	secret := os.Getenv("SECRET_KEY")
	csrfMiddleware := csrf.Protect([]byte(secret[:32]))
	r.Use(sessionMiddleware)
	r.Use(csrfMiddleware)
	userService, err := NewUserService("http://dynamodb:8000")
	if err != nil {
		fmt.Println("error", err)
		return
	}
	userService.CreateTable()
	env := Env{
		userService: userService,
	}
	r.HandleFunc("/", requiresAuth(env.indexHandler)).Methods("GET")
	r.HandleFunc("/logout/", env.logoutHandler).Methods("GET")
	r.HandleFunc("/login/", env.loginHandler).Methods("GET")
	r.HandleFunc("/login/spotify/", env.loginInitHandler).Methods("GET")
	r.HandleFunc("/complete/spotify/", env.loginCompleteHandler).Methods("GET")
	http.Handle("/", r)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	fmt.Printf("Listening on port %s\n", port)
	loggedRouter := handlers.LoggingHandler(os.Stdout, r)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), loggedRouter))
}
