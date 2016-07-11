package main

import (
	"github.com/microo8/gomodoro/gomodoro"
	"net/http"
)

func main() {
	es := gomodoro.NewEventStore()
	defer es.Close()
	ps := gomodoro.NewPomodoroStore(es.FwdChannel())
	mux := http.NewServeMux()
	mux.Handle("/", es)
	mux.Handle("/pomodoro", ps)
	http.ListenAndServe(":8080", mux)
}
