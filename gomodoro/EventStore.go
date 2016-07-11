package gomodoro

import (
	"encoding/binary"
	"github.com/syndtr/goleveldb/leveldb"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/user"
	"path"
)

const appName string = "gomodoro"

type EventStore struct {
	db          *leveldb.DB
	eventCount  uint64
	fwdChannels []chan[]byte
}

func NewEventStore() *EventStore {
	es := new(EventStore)

	//open database
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	if _, err = os.Stat(path.Join(usr.HomeDir, "."+appName)); os.IsNotExist(err) {
		os.Mkdir(path.Join(usr.HomeDir, appName), 0755) //TODO .cache
	}
	es.db, err = leveldb.OpenFile(path.Join(usr.HomeDir, appName, "events"), nil)
	if err != nil {
		log.Fatal(err)
	}

	//count events
	es.eventCount = 0
	iter := es.db.NewIterator(nil, nil)
	defer iter.Release()
	for iter.Next() {
		es.eventCount++
	}
	err = iter.Error()
	if err != nil {
		log.Fatal(err)
	}

	es.fwdChannels = make([]chan[]byte, 10)
	return es
}

func (self *EventStore) Close() {
	self.db.Close()
}

func (self *EventStore) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	self.eventCount++
	var key []byte
	binary.LittleEndian.PutUint64(key, self.eventCount)
	value, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(key, value)
	self.db.Put(key, value, nil)
	res.Write(key)
	for _, fwdCh := range self.fwdChannels {
		fwdCh <- value
	}
}

func (self *EventStore) FwdChannel() chan []byte {
	fwdCh := make(chan []byte)
	self.fwdChannels = append(self.fwdChannels, fwdCh)
	return fwdCh
}
