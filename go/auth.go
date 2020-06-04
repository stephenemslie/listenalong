package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/spotify"
)

var (
	oauthConfig *oauth2.Config
)

type Env struct {
	userService *UserService
}

func (env *Env) loginHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := baseTemplate.Clone()
	t.ParseFiles("templates/login.html")
	err := t.Execute(w, struct{}{})
	if err != nil {
		fmt.Println(err)
	}
}

func (env *Env) loginInitHandler(w http.ResponseWriter, r *http.Request) {
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

func (env *Env) loginCompleteHandler(w http.ResponseWriter, r *http.Request) {
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

func requiresAuth(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session := r.Context().Value(sessionKey).(*sessions.Session)
		if session == nil {
			http.Redirect(w, r, "/login", http.StatusFound)
		}
		if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
			http.Redirect(w, r, "/login", http.StatusFound)
		}
		fn(w, r)
	}
}

func init() {
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
