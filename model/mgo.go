package model

import (
	log "github.com/Sirupsen/logrus"
	"gopkg.in/mgo.v2"
	"time"
)

var (
	session      *mgo.Session
	dbNamePrefix = ""
	dbUrl        = ""
	// cursor time out seconds
	timeoutInterval = time.Duration(10 * 60)
	// max connections to mongodb
	maxConnection = 512
	isInited      = false
)

var (
	ErrNotFound = mgo.ErrNotFound
)

func initMgo(url string, dbname string) {
	if isInited == true {
		return
	}

	if len(url) != 0 {
		dbUrl = url
	} else {
		log.Fatalf("mongo net url is empty")
	}

	if len(dbname) != 0 {
		dbNamePrefix = dbname
	} else {
		log.Fatalf("mongo database name is empty")
	}

	initConnection()

	isInited = true

	log.Infof("init mongo(%s/%s)", dbUrl, dbNamePrefix)
}

func GetSession() *mgo.Session {
	if session == nil {
		initConnection()
	}

	return session.Copy()
}

func initConnection() {
	var err error
	session, err = mgo.Dial(dbUrl)
	if err != nil {
		log.Fatal("mongo connect to " + dbUrl + " faild:" + err.Error())
	}
	session.SetMode(mgo.Monotonic, false)
	session.SetSyncTimeout(timeoutInterval * time.Second)
	session.SetSocketTimeout(time.Second * 600)
	session.SetPoolLimit(maxConnection)
}

func SetStrongMode(s *mgo.Session) {
	s.SetMode(mgo.Strong, false)
}

func SetEventualMode(s *mgo.Session) {
	s.SetMode(mgo.Eventual, false)
}
