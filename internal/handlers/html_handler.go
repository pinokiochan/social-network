package handlers

import (
	"net/http"
	"path/filepath"
)

func ServeHTML(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := filepath.Join("web", "templates", "index.html")
	http.ServeFile(w, r, path)
}
