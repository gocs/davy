package main

import (
	"flag"
	"log"

	"net/http"

	"github.com/gocs/davy/router"
)

var (
	session = flag.String("session-key", "soopa-shiikurrets", "sets the session cookie store key")
)

func main() {
	flag.Parse()

	r, err := router.NewRouter(*session)
	if err != nil {
		log.Fatal(err)
	}
	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
