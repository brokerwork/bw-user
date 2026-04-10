package model

import (
	"bw/bw-user/conf"

	"github.com/Sirupsen/logrus"
	"github.com/lworkltd/kits/helper/mq"
	"github.com/streadway/amqp"

	"bw/bw-user/constant"
	"encoding/json"
	"strings"
)

func InitMqListener(config *conf.Application) error {
	if err := initAccountListenrMq(config); err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("Failed to boot mq listener")
		return err
	}

	if err := initUserSenderMq(config); err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("Failed to boot mq listener")
		return err
	}

	return nil
}

var account_mq_session *mq.Session

func initAccountListenrMq(config *conf.Application) error {
	var exchanger = "ACCOUNT_REPORT"
	var routing_key = "ACCOUNT.*.*.*"
	var queue = "BW_USER_RCV_ACCOUNT_QUEUE_NEW"

	account_mq_session, err := mq.Dial(config.MqUrl)
	if err != nil {
		logrus.Error("connectMq mq.Dial failed")
		return err
	}

	if err := account_mq_session.HandleExchange(
		queue,
		exchanger,
		"topic",
		accountMsgHandler,
		mq.MakeupSettings(
			mq.NewExchangeSettings().Durable(true),
			mq.NewQueueSettings().Durable(true).Exclusive(false).AutoDelete(false),
			mq.NewConsumeSettings().AutoAck(true).Exclusive(false).NoLocal(false),
		), routing_key); err != nil {
		panic(err)
	}

	return nil
}

func accountMsgHandler(delivery *amqp.Delivery) {
	logrus.Infof("Rcv Msg: %s %s", delivery.RoutingKey, delivery.Body)

	var event BwMsgEvent;
	if err := json.Unmarshal([]byte(delivery.Body), &event); nil != err {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Warn("Failed to handler mq msg")
		return
	}

	infoList := strings.Split(delivery.RoutingKey, ".")

	var feignKey XFeignKey = XFeignKey{
		TenantId:  infoList[1],
		ProductId: "BW",
	}

	if event.Type == constant.BwEvent_DELETE {
		//断开关系
		breakUserAccountRelation(&feignKey, infoList[2], infoList[3], event.Id)
	}
}

var user_mq_session *mq.Session

func initUserSenderMq(config *conf.Application) error {
	session, err := mq.Dial(config.MqUrl)
	if err != nil {
		logrus.Error("connectMq mq.Dial failed")
		return err
	}
	user_mq_session = session

	return nil
}

func SendUserMsg(tenantId string, userId string, eventType string, userDetail *UserDetail) error {
	var event BwMsgEvent = BwMsgEvent{
		TenantId: tenantId,
		Id:       userId,
		Type:     eventType,
		ObjType:  "USER",
		Detail:   userDetail,
	}

	var exchanger = "USER_REPORT"

	msg, err := json.Marshal(event)
	if nil != err {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Warn("Failed to Marshal mq msg")
		return err
	}

	var routingKey = "USER." + tenantId
	user_mq_session.Publish(msg, exchanger, routingKey)
	logrus.Infof("Send Msg: %s %s", routingKey, msg)
	return nil
}
