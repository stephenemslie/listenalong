package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

var (
	baseTemplate *template.Template
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

func init() {
	baseTemplate = template.New("base.html")
	var err error
	baseTemplate, err = baseTemplate.ParseGlob("templates/layout/*.html")
	if err != nil {
		fmt.Println(err)
	}
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", indexHandler).Methods("GET")
	http.Handle("/", r)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	fmt.Printf("Listening on port %s\n", port)
	loggedRouter := handlers.LoggingHandler(os.Stdout, r)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), loggedRouter))
}
