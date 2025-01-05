package handlers

import (
	"html/template"
	"net/http"
)


func ServeHTML(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("./web/templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	t.Execute(w, nil)
}
