package model

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	mgo "gopkg.in/mgo.v2"
	"github.com/lworkltd/kits/service/profile"
)


var MongoDBName = ""

func InitMongoDB(config *profile.Mongo) {
	MongoDBName = config.Db
	if len(MongoDBName) <= 0 {
		logrus.Fatal("MongoDBName is empty")
	}
	// mongo初始化
	initMgo(config.Url, MongoDBName)
}

// ExecDB method
func ExecDB(collName string, f func(*mgo.Collection) error) error {
	sess := GetSession()
	defer sess.Close()

	coll := sess.DB(MongoDBName).C(collName)
	return f(coll)
}

// EnsureIndex method; create index;unique means unique index key
func EnsureIndex(collName string, unique bool, keys ...string) error {
	return ExecDB(collName, func(coll *mgo.Collection) error {
		return coll.EnsureIndex(mgo.Index{
			Key:    keys,
			Unique: unique,
			Sparse: true,
		})
	})
}

func MongoSave(collName string, id string, entity interface{}) {
	logrus.Debugf("MongoSave, collName:%s, id:%d", collName, id)
	ExecDB(collName, func(coll *mgo.Collection) error {
		_, err := coll.UpsertId(id, entity)
		if nil != err {
			logrus.WithFields(logrus.Fields{
				"error":    err,
				"collName": collName,
			}).Errorf("mongo save error entity:%v", entity)
		}
		return err
	})
}

func MongoInsert(collName string, event interface{}) {
	logrus.Debugf("MongoInsert, collName:%s", collName)
	ExecDB(collName, func(coll *mgo.Collection) error {
		err := coll.Insert(event)
		if nil != err {
			logrus.WithFields(logrus.Fields{
				"error":    err,
				"collName": collName,
			}).Errorf("mongo insert error event:%v", event)
		}
		return err
	})
}

func MongoFindOne(collName string, query interface{}, result interface{}) error {
	logrus.Debugf("MongoFindOne, collName:%s", collName)
	return ExecDB(collName, func(coll *mgo.Collection) error {
		err := coll.Find(query).One(result)
		if nil != err {
			logrus.WithFields(logrus.Fields{
				"error":    err,
				"collName": collName,
			}).Errorf("mongo find one, error query:%v", query)
		}
		return err
	})
}

func MongoFindAll(collName string, query interface{}, result interface{}) error {
	logrus.Debugf("MongoFindAll, collName:%s", collName)
	return ExecDB(collName, func(coll *mgo.Collection) error {
		err := coll.Find(query).All(result)
		if nil != err {
			logrus.WithFields(logrus.Fields{
				"error":    err,
				"collName": collName,
			}).Error("mongo find all, error query:%v", query)
		}
		return err
	})
}

func MongoDelete(collName string, id int64) {
	logrus.Debugf("MongoDelete, collName:%s, id:%d", collName, id)
	ExecDB(collName, func(coll *mgo.Collection) error {
		err := coll.RemoveId(id)
		if nil != err {
			logrus.WithFields(logrus.Fields{
				"error":    err,
				"collName": collName,
			}).Errorf("mongo insert error id:%v", id)
		}
		return err
	})
}

func MongoExists(collName string, query interface{}) bool {
	err := ExecDB(collName, func(coll *mgo.Collection) error {
		count, err := coll.Find(query).Count()
		if nil != err {
			logrus.WithFields(logrus.Fields{
				"error":    err,
				"collName": collName,
			}).Errorf("MongoExists, error query:%v", query)
			return err
		}

		if count > 0 {
			return fmt.Errorf("Exists")
		}

		return nil
	})

	if err != nil {
		return true
	}

	return false
}
