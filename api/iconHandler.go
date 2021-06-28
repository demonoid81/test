package api

import (
	"github.com/gorilla/mux"
	"net/http"
)

func IconHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		icon, err := vars["icon"]
		if !err {
			http.Error(w, "Not found icon", http.StatusNotFound)
			return
		}
		path := "./static/" + icon + ".png"
		http.ServeFile(w, r, path)
	}
}


