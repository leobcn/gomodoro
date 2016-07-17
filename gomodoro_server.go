package main

//go:generate go-bindata -pkg gomodoro -o gomodoro/bindata.go credentials.json

import (
	"encoding/json"
	"errors"
	"github.com/microo8/gomodoro/gomodoro"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"log"
	"net/http"
)

var (
	conf  *oauth2.Config
	users map[string]*User
)

type User struct {
	ID          string
	DisplayNane string
	GivenName   string
	FamilyName  string
	Email       string
	Token       *oauth2.Token
}

func (self *User) Update() error {
	client := conf.Client(oauth2.NoContext, self.Token)
	profileBytes, err := client.Get("https://people.googleapis.com/v1/people/me")
	if err != nil {
		return err
	}

	profile := make(map[string]interface{})
	err = json.NewDecoder(profileBytes.Body).Decode(&profile)
	if err != nil {
		return err
	}
	namesArray := profile["names"].([]interface{})
	names := namesArray[0].(map[string]interface{})
	metadata := names["metadata"].(map[string]interface{})
	source := metadata["source"].(map[string]interface{})
	self.ID = source["id"].(string)
	self.DisplayNane = names["displayName"].(string)
	self.GivenName = names["givenName"].(string)
	self.FamilyName = names["familyName"].(string)
	emailAddressesArray := profile["emailAddresses"].([]interface{})
	emailAddresses := emailAddressesArray[0].(map[string]interface{})
	self.Email = emailAddresses["value"].(string)
	return nil
}

func userSession(req *http.Request) (*User, error) {
	user_id, err := req.Cookie("user_id")
	if err != nil {
		return nil, errors.New("no user")
	}
	user, ok := users[user_id.Value]
	if !ok {
		return nil, errors.New("user_id not found")
	}
	user.Update()
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

	//get and Unmarshal profile info
	user := &User{Token: tok}
	err = user.Update()
	if err != nil {
		log.Fatal(err)
	}
	users[user.ID] = user
	cookie := &http.Cookie{Name: "user_id", Value: user.ID}
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
	users = make(map[string]*User)

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
