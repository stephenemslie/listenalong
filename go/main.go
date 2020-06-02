package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

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

func indexHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := baseTemplate.Clone()
	t.ParseFiles("templates/index.html")
	data := struct{}{}
	err := t.Execute(w, data)
	if err != nil {
		fmt.Println(err)
	}
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
}

func main() {
	r := mux.NewRouter()
	r.Use(sessionMiddleware)
	r.HandleFunc("/", requiresAuth(indexHandler)).Methods("GET")
	http.Handle("/", r)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	fmt.Printf("Listening on port %s\n", port)
	loggedRouter := handlers.LoggingHandler(os.Stdout, r)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), loggedRouter))
}
