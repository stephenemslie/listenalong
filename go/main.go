package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

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
	oauthConfig  *oauth2.Config
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := baseTemplate.Clone()
	t.ParseFiles("templates/index.html")
	err := t.Execute(w, struct{}{})
	if err != nil {
		fmt.Println(err)
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := baseTemplate.Clone()
	t.ParseFiles("templates/login.html")
	err := t.Execute(w, struct{}{})
	if err != nil {
		fmt.Println(err)
	}
}

func loginInitHandler(w http.ResponseWriter, r *http.Request) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	code := base64.StdEncoding.EncodeToString(b)
	session := r.Context().Value(sessionKey).(*sessions.Session)
	session.Values["oauth_state"] = code
	session.Save(r, w)
	url := oauthConfig.AuthCodeURL(code)
	http.Redirect(w, r, url, http.StatusFound)
}

func loginCompleteHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session := ctx.Value(sessionKey).(*sessions.Session)
	if session.Values["oauth_state"] != r.FormValue("state") {
		fmt.Println("error: State doesn't match")
		return
	}
	code := r.FormValue("code")
	tok, err := oauthConfig.Exchange(oauth2.NoContext, code)
	if err != nil {
		fmt.Println("error:", err)
	}
	session.Values["spotify_token"] = tok
	session.Values["authenticated"] = true
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusFound)
}

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

func requiresAuth(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session := r.Context().Value(sessionKey).(*sessions.Session)
		if session == nil {
			http.Redirect(w, r, "/login", http.StatusFound)
		}
		if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
			http.Redirect(w, r, "/login", http.StatusFound)
		}
	}
}

func init() {
	secret := os.Getenv("SECRET")
	sessionStore = sessions.NewCookieStore([]byte(secret))
	baseTemplate = template.New("base.html")
	var err error
	baseTemplate, err = baseTemplate.ParseGlob("templates/layout/*.html")
	if err != nil {
		fmt.Println(err)
	}
	host := os.Getenv("HOST")
	oauthConfig = &oauth2.Config{
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
}

func main() {
	r := mux.NewRouter()
	r.Use(sessionMiddleware)
	r.HandleFunc("/", requiresAuth(indexHandler)).Methods("GET")
	r.HandleFunc("/login", loginHandler).Methods("GET")
	r.HandleFunc("/login/spotify/", loginInitHandler).Methods("GET")
	r.HandleFunc("/complete/spotify/", loginCompleteHandler).Methods("GET")
	http.Handle("/", r)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	fmt.Printf("Listening on port %s\n", port)
	loggedRouter := handlers.LoggingHandler(os.Stdout, r)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), loggedRouter))
}
