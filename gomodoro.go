package main

import (
	"github.com/microo8/meh/qomodoro"
	"net/http"
)

func main() {
	es := qomodoro.NewEventStore()
	defer es.Close()
	mux := http.NewServeMux()
	mux.Handle("/pomodoro", es)
	http.ListenAndServe(":8080", mux)
}
