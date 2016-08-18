package filehost

import (
	"net/http"

	"github.com/gorilla/mux"
)

func serveFile(w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	fileName := v["id"]
	http.ServeFile(w, r, "./files/"+fileName)
}
