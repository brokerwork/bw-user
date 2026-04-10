package model

import (
	"bw/bw-user/errcode"
	"github.com/Sirupsen/logrus"
	"github.com/lworkltd/kits/service/restful/code"
)

//根据t_user_detail表中的keyType列获取值在values中的记录记录, values为切片
func RemainUserCount(feignKey *XFeignKey, timeMsc int64) (int64, code.Error) {
	dbConn, errConn := GetDBConnByTenantID(feignKey.TenantId)
	if errConn != nil {
		return 0, errConn
	}

	var total int64 = 0
	if errCount := dbConn.Model(&UserDetailTab{}).Where("UNIX_TIMESTAMP(create_date) < ?", timeMsc).Count(&total).Error; nil != errCount {
		logrus.WithFields(logrus.Fields{"error": errCount}).Error("user statistic count failed")
		return 0, errcode.CerrExecuteSQL
	}
	return total, nil
}