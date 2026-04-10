package model

import (

"bw/bw-user/constant"
"encoding/json"
"github.com/Sirupsen/logrus"
"github.com/lworkltd/kits/service/invoke"
"github.com/lworkltd/kits/service/restful/code"
invokeutils "github.com/lworkltd/kits/utils/invoke"
"strconv"

)


//用户佣金规则详细-添加修改
func ClientAddOrUpdateCommission(feignKey *XFeignKey, addReq *UserCommissionUpdateRequestDTO ) (bool, code.Error){
	feignkeyBypes,_ := json.Marshal(feignKey)
	rsp, err := invoke.Name(constant.ServiceNameBwReport).
	Post(constant.PathBwReportAddOrUpdate).
	Header("X-Feign-Key", string(feignkeyBypes)).
	Json(addReq).
	Response()

	var response bool
	cerr := invokeutils.ExtractHttpResponse(constant.ServiceNameBwReport, err, rsp, &response)
	if cerr != nil {
		logrus.WithFields(logrus.Fields{
			"userID": addReq.UserId,
			"data": response,
		}).Error("Request add or update commission failed")
		return false, cerr
	}

	logrus.WithFields(logrus.Fields{
		"userID": addReq.UserId,
		"data": response,
	}).Debug("Request add or update commission success")
	return response, nil
}


//用户佣金规则详细-删除
func ClienDeleteUserCommission(feignKey *XFeignKey, userIdNum int64) (bool, code.Error){
	feignkeyBypes,_ := json.Marshal(feignKey)
	rsp, err := invoke.Name(constant.ServiceNameBwReport).
		Post(constant.PathBwReportUserDelete).
		Header("X-Feign-Key", string(feignkeyBypes)).
		Route("userId", strconv.FormatInt(userIdNum, 10)).
		Response()

	var response bool
	cerr := invokeutils.ExtractHttpResponse(constant.ServiceNameBwReport, err, rsp, &response)
	if cerr != nil {
		logrus.WithFields(logrus.Fields{
			"userID": userIdNum,
			"data": response,
		}).Error("Request delete commission failed")
		return false, cerr
	}

	logrus.WithFields(logrus.Fields{
		"userID":userIdNum,
		"data": response,
	}).Debug("Request add or update commission success")
	return response, nil
}

//层级是否不在使用用
func clienIsLevelNotUsed(feignKey *XFeignKey, levelId int64) (bool, code.Error){
	feignkeyBypes,_ := json.Marshal(feignKey)
	rsp, err := invoke.Name(constant.ServiceNameBwReport).
		Post(constant.PathIsLevelNotUsed).
		Header("X-Feign-Key", string(feignkeyBypes)).
		Route("levelId", strconv.FormatInt(levelId, 10)).
		Response()

	var response bool
	cerr := invokeutils.ExtractHttpResponse(constant.ServiceNameBwReport, err, rsp, &response)
	if cerr != nil {
		logrus.WithFields(logrus.Fields{"levelId": levelId,"data": response,}).Error("Request is level used failed")
		return false, cerr
	}
	return response, nil
}

