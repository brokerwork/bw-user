package model

import (
	"github.com/lworkltd/kits/service/restful/code"
	"bw/bw-user/errcode"
	"gopkg.in/mgo.v2/bson"
	"github.com/Sirupsen/logrus"
	"gopkg.in/mgo.v2"
	"time"
)

const (
	Collection_DashboardConfig = "t_dashboard_config"
)


func ProcessGetConfigByUserId(feignKey *XFeignKey, userId string) (*DashboardConfig, code.Error) {
	if nil == feignKey {
		return nil, errcode.CerrParamater
	}

	config := &DashboardConfig{}
	err := MongoFindOne(Collection_DashboardConfig, bson.M{"tenantId": feignKey.TenantId, "userId": userId}, config)
	if err != nil {
		if err != mgo.ErrNotFound {
			logrus.WithFields(logrus.Fields{
				"error":    err,
				"collName": Collection_DashboardConfig,
			}).Error("GetConfigByUserId, find error, userId:%s", userId)
			return nil, nil
		} else {
			config = &DashboardConfig{
				TenantId: feignKey.TenantId,
				UserId: userId,
				Type: []string{"NEW_CUSTOMER_PAGE",
					"DEPOSIT_MONEY_TREND",
					"DEPOSIT_DISTRIBUTE_TREND",
					"DEPOSIT_CUSTOMER_TREND",
					"NEW_DEAL_PAGE",
					"USER_RANK_NEW_CUSTOMER_PAGE",
					"SOURCE_TUNNEL",
					"NEW_CUSTOMER_PAGE_DISTRIBUTE",
					"NEW_CUSTOMER_TENANT_ACTIVE_LOGIN_PAGE",
					"NEW_CUSTOMER_TENANT_ACTIVE_DEAL_PAGE",
					"TRADE_VARIRTY_DISTRIBUTE",
					"NEW_CUSTOMER_TENANT_DORMANT_PAGE"},
				Time: int64(time.Now().Second() * 1000),
			}
			err := ProcessSaveConfig(feignKey, config)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"error":    err,
					"collName": Collection_DashboardConfig,
				}).Error("ProcessSaveConfig, error, userId:%s", userId)
				return nil, nil
			}
		}

	}
	return config, nil
}


func ProcessSaveConfig(feignKey *XFeignKey, config *DashboardConfig) code.Error {
	if nil == feignKey {
		return errcode.CerrParamater
	}

	config.TenantId = feignKey.TenantId
	config.Time = int64(time.Now().Second() * 1000)
	MongoSave(Collection_DashboardConfig, feignKey.TenantId + "_" + config.UserId, config)
	return nil
}


func ProcessDeleteConfig(feignKey *XFeignKey, config *DashboardConfig) code.Error {
	if nil == feignKey {
		return errcode.CerrParamater
	}

	config.Time = int64(time.Now().Second() * 1000);
	ExecDB(Collection_DashboardConfig, func(coll *mgo.Collection) error {
		err := coll.Update(bson.M{"tenantId": feignKey.TenantId, "userId": config.UserId}, bson.M{"$pull": bson.M{"type": bson.M{"$in": config.Type}}})
		if nil != err {
			logrus.WithFields(logrus.Fields{
				"error":    err,
				"collName": Collection_DashboardConfig,
			}).Errorf("ProcessDeleteConfig mongo save error entity:%v", config)
		}
		return err
	})
	return nil
}