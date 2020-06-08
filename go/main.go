package main

import (
	"context"
	"encoding/gob"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/csrf"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/spotify"
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
	env := Env{}
	userService, err := NewUserService("http://dynamodb:8000")
	if err != nil {
		fmt.Println("error", err)
		return
	}
	userService.CreateTable()
	env.userService = userService
	host := os.Getenv("HOST")
	env.oauthConfig = &oauth2.Config{
		ClientID:     os.Getenv("SPOTIFY_KEY"),
		ClientSecret: os.Getenv("SPOTIFY_SECRET"),
		RedirectURL:  fmt.Sprintf("%s/complete/spotify/", host),
		Scopes: []string{
			"user-read-playback-state",
			"user-modify-playback-state",
			"user-read-currently-playing",
			"playlist-read-collaborative",
		},
		Endpoint: spotify.Endpoint,
	}
	gob.Register(oauth2.Token{})
	r.HandleFunc("/", requiresAuth(env.indexHandler)).Methods("GET")
	r.HandleFunc("/update/", requiresAuth(env.updatePlayingHandler)).Methods("GET")
	r.HandleFunc("/logout/", env.logoutHandler).Methods("GET")
	r.HandleFunc("/login/", env.loginHandler).Methods("GET")
	r.HandleFunc("/login/spotify/", env.loginInitHandler).Methods("GET")
	r.HandleFunc("/complete/spotify/", env.loginCompleteHandler).Methods("GET")
	r.HandleFunc("/follow/", env.followHandler).Methods("POST")
	http.Handle("/", r)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	fmt.Printf("Listening on port %s\n", port)
	loggedRouter := handlers.LoggingHandler(os.Stdout, r)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), loggedRouter))
}
