package servererrors

import (
	"log"
	"net/http"
)

func InternalServerError(w http.ResponseWriter, err string) {
	s := http.StatusInternalServerError
	http.Error(w, "Internal server error", s)
	log.Printf("err %d: %v\n", s, err)
}
