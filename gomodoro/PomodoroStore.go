package gomodoro

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type Pomodoro struct {
	user  string
	begin time.Time
	end   time.Time
	flag  string
}

type PomodoroRequest struct {
	query string
	res   chan []*Pomodoro
}

func NewPomodoroRequest(query string) *PomodoroRequest {
	return &PomodoroRequest{
		query: query,
		res:   make(chan []*Pomodoro, 1),
	}
}

type PomodoroStore struct {
	users      map[string][]*Pomodoro
	flags      map[string][]*Pomodoro
	chPomodoro <-chan []byte
	chRequest  <-chan *PomodoroRequest
}

func NewPomodoroStore(pomodoroChan <-chan []byte) (ps *PomodoroStore) {
	ps = &PomodoroStore{
		users:      make(map[string][]*Pomodoro),
		flags:      make(map[string][]*Pomodoro),
		chPomodoro: pomodoroChan,
		chRequest:  make(chan *PomodoroRequest),
	}
	go ps.run()
	return ps
}

//receives json encoded pomodoros from a channel and adds it to the maps
func (self *PomodoroStore) run() {
	for pomodoroJson := range self.chPomodoro {
		pomodoro := new(Pomodoro)
		err := json.Unmarshal(pomodoroJson, pomodoro)
		if err != nil {
			log.Fatal(err)
			continue
		}
		if _, ok := self.users[pomodoro.user]; !ok {
			self.users[pomodoro.user] = make([]*Pomodoro, 1024)
		}
		self.users[pomodoro.user] = append(self.users[pomodoro.user], pomodoro)
		if _, ok := self.flags[pomodoro.flag]; !ok {
			self.users[pomodoro.flag] = make([]*Pomodoro, 1024)
		}
		self.flags[pomodoro.flag] = append(self.flags[pomodoro.flag], pomodoro) //TODO
	}
}

func (self *PomodoroStore) ServeHTTP(res http.ResponseWriter, req *http.Request) {

}
