package servererrors

import (
	"log"
	"net/http"
)

// InternalServerError sends internal server error to the server and logs the actual error to the stdout
func InternalServerError(w http.ResponseWriter, err string) {
	s := http.StatusInternalServerError
	http.Error(w, "Internal server error", s)
	log.Printf("err %d: %v\n", s, err)
}
