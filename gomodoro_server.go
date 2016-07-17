package main

//go:generate go-bindata -pkg gomodoro -o gomodoro/bindata.go credentials.json

import (
	"encoding/json"
	"errors"
	"github.com/microo8/gomodoro/gomodoro"
	"github.com/satori/go.uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"log"
	"net/http"
)

var (
	conf     *oauth2.Config
	sessions map[string]*User
)

type User struct {
	DisplayNane string
	GivenName   string
	FamilyName  string
	Email       string
	Token       *oauth2.Token
}

func userSession(req *http.Request) (*User, error) {
	session_id, err := req.Cookie("session_id")
	if err != nil {
		return nil, errors.New("no session")
	}
	user, ok := sessions[session_id.Value]
	if !ok {
		return nil, errors.New("session_id not found")
	}
	return user, nil
}

func index(res http.ResponseWriter, req *http.Request) {
	user, err := userSession(req)
	if err != nil { //not logged
		res.Write([]byte("<a href=\"redirect\">login</a>"))
	} else {
		res.Write([]byte("<p>Hello " + user.DisplayNane + "!</p>"))
	}
}

func redirect(res http.ResponseWriter, req *http.Request) {
	// Redirect user to Google's consent page to ask for permission
	// for the scopes specified above.
	url := conf.AuthCodeURL("state2")
	res.Write([]byte(`<html><head></head><body onload="window.location='` + url + `'"></body></html>`))
}

func auth(res http.ResponseWriter, req *http.Request) {
	tok, err := conf.Exchange(oauth2.NoContext, req.URL.Query().Get("code"))
	if err != nil {
		log.Fatal(err)
	}
	client := conf.Client(oauth2.NoContext, tok)

	//get and Unmarshal profile info
	profileBytes, err := client.Get("https://people.googleapis.com/v1/people/me")
	if err != nil {
		log.Fatal(err)
	}
	user := new(User)
	user.Token = tok
	profile := make(map[string]interface{})
	err = json.NewDecoder(profileBytes.Body).Decode(&profile)
	if err != nil {
		log.Fatal(err)
	}
	namesArray := profile["names"].([]interface{})
	names := namesArray[0].(map[string]interface{})
	user.DisplayNane = names["displayName"].(string)
	user.GivenName = names["givenName"].(string)
	user.FamilyName = names["familyName"].(string)
	emailAddressesArray := profile["emailAddresses"].([]interface{})
	emailAddresses := emailAddressesArray[0].(map[string]interface{})
	user.Email = emailAddresses["value"].(string)

	session_id, err := uuid.NewV4().MarshalText()
	if err != nil {
		log.Fatal(err)
	}
	sessions[string(session_id)] = user
	cookie := &http.Cookie{Name: "session_id", Value: string(session_id)}
	http.SetCookie(res, cookie)
	res.Write([]byte(`<html><head></head><body onload="window.location='/'"></body></html>`))
}

func main() {
	jsonKey, err := gomodoro.Asset("credentials.json")
	if err != nil {
		log.Fatal(err)
	}
	conf, err = google.ConfigFromJSON(jsonKey, "https://www.googleapis.com/auth/userinfo.profile", "https://www.googleapis.com/auth/userinfo.email")
	if err != nil {
		log.Fatal(err)
	}
	sessions = make(map[string]*User)

	es := gomodoro.NewEventStore()
	defer es.Close()
	ps := gomodoro.NewPomodoroStore(es.FwdChannel())
	mux := http.NewServeMux()
	mux.HandleFunc("/", index)
	mux.HandleFunc("/redirect", redirect)
	mux.HandleFunc("/oauth2redirect", auth)
	mux.Handle("/put", es)
	mux.Handle("/pomodoro", ps)
	log.Println("listening")
	http.ListenAndServe(":80", mux)
}
