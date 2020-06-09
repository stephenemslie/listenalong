package api

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

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/spotify"
)

var (
	baseTemplate *template.Template
	sessionStore *sessions.CookieStore
	sessionKey   key
)

type key int

type Env struct {
	userService *UserService
	oauthConfig *oauth2.Config
}

func (env *Env) indexHandler(w http.ResponseWriter, r *http.Request) {
	session := r.Context().Value(sessionKey).(*sessions.Session)
	userID := session.Values["user_id"].(string)
	t, _ := baseTemplate.Clone()
	t.ParseFiles("templates/index.html")
	user := User{}
	env.userService.GetUser(userID, &user)
	data := struct {
		User User
	}{user}
	err := t.Execute(w, data)
	if err != nil {
		fmt.Println(err)
	}
}

func (env *Env) updatePlayingHandler(w http.ResponseWriter, r *http.Request) {
	session := r.Context().Value(sessionKey).(*sessions.Session)
	tok := session.Values["spotify_token"].(oauth2.Token)
	client := env.oauthConfig.Client(oauth2.NoContext, &tok)
	user := User{}
	env.userService.GetUser(session.Values["user_id"].(string), &user)
	_, err := user.UpdatePlaying(client)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "", 400)
		return
	}
	env.userService.PutUser(&user)
	http.Redirect(w, r, "/", http.StatusFound)
}

func (env *Env) loginHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := baseTemplate.Clone()
	t.ParseFiles("templates/login.html")
	err := t.Execute(w, struct{}{})
	if err != nil {
		fmt.Println(err)
	}
}

func (env *Env) logoutHandler(w http.ResponseWriter, r *http.Request) {
	session := r.Context().Value(sessionKey).(*sessions.Session)
	session.Values["user_id"] = ""
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusFound)
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
	url := env.oauthConfig.AuthCodeURL(code)
	http.Redirect(w, r, url, http.StatusFound)
}

func (env *Env) loginCompleteHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session := ctx.Value(sessionKey).(*sessions.Session)
	if session.Values["oauth_state"] != r.FormValue("state") {
		http.Error(w, "State doesn't match", 400)
		return
	}
	code := r.FormValue("code")
	tok, err := env.oauthConfig.Exchange(oauth2.NoContext, code)
	if err != nil {
		http.Error(w, "OAuth2 Error", 400)
		log.Fatal("error:", err)
	}
	client := env.oauthConfig.Client(oauth2.NoContext, tok)
	user, err := NewUserFromSpotify(client)
	err = env.userService.GetOrCreateUser(&user)
	if err != nil {
		log.Fatal(err)
	}
	session.Values["user_id"] = user.ID
	session.Values["spotify_token"] = tok
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusFound)
}

func (env *Env) followHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Printf("ParseForm error %v", err)
	}
	session := r.Context().Value(sessionKey).(*sessions.Session)
	tok := session.Values["spotify_token"].(oauth2.Token)
	userID := session.Values["user_id"].(string)
	user := User{}
	env.userService.GetUser(userID, &user)
	spotifyUsername := r.FormValue("user_id")
	spotifyUser := User{}
	env.userService.GetUser(spotifyUsername, &spotifyUser)
	client := env.oauthConfig.Client(oauth2.NoContext, &tok)
	user.Follow(&spotifyUser, client)
	env.userService.userTable.
		Update("id", userID).
		Set("following_id", spotifyUsername).
		Run()
	http.Redirect(w, r, "/", http.StatusFound)
}

func requiresAuth(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session := r.Context().Value(sessionKey).(*sessions.Session)
		if session == nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		if user_id, ok := session.Values["user_id"].(string); !ok || len(user_id) == 0 {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		fn(w, r)
	}
}

func NewRouter() *mux.Router {
	r := mux.NewRouter()
	r = r.StrictSlash(true)
	// secret := os.Getenv("SECRET_KEY")
	// csrfMiddleware := csrf.Protect([]byte(secret[:32]))
	r.Use(sessionMiddleware)
	// r.Use(csrfMiddleware)
	env := Env{}
	userService, err := NewUserService("http://dynamodb:8000")
	if err != nil {
		log.Fatal(err)
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
	return r
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
