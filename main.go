package main

import (
	"bw/bw-user/api/server"
	"bw/bw-user/conf"
	"bw/bw-user/model"
	"github.com/Sirupsen/logrus"
	logutil "github.com/lworkltd/kits/utils/log"
	"log"
	"os"
)

func panicErr(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	configFile := "app.toml"
	if len(os.Args) > 1 {
		configFile = os.Args[1]
	}

	if err := conf.Parse(configFile); err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("Failed to boot service")
	}
	conf.Dump()
	logrus.AddHook(logutil.NewFileLineHook(log.Lshortfile))

	//初始化MQ
	if err := model.InitMqListener(conf.GetApplication()); err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("Failed to boot service")
	}

	//初始化MongoDB
	model.InitMongoDB(conf.GetMongo())

	if err := server.Setup(conf.GetService()); err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("Failed to boot service")
	}
}
