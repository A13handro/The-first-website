package main

import (
	"html/template"
	"net/http"

	"github.com/gorilla/mux"

	_ "github.com/lib/pq"
)

func index(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/index.html")
	if err != nil {
		panic(err)
	}

	t.Execute(w, "index.html")
}
func handlfunc() {
	rtr := mux.NewRouter()

	rtr.HandleFunc("/", index).Methods("GET")

	http.Handle("/", rtr)

	http.ListenAndServe(":8080", nil)

}

func main() {
	handlfunc()
}
