package main

import (
	"flag"

	"github.com/gocs/davy/router"
	"net/http"
)

var (
	session = flag.String("session-key", "soopa-shiikurrets", "sets the session cookie store key")
)

func main() {
	flag.Parse()

	r := router.NewRouter(*session)
	http.Handle("/", r)
	http.ListenAndServe(":8080", nil)
}
