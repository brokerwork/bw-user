package model

import (
	"github.com/lworkltd/kits/service/restful/code"
	"bw/bw-user/constant"
	"github.com/lworkltd/kits/service/invoke"
	invokeutils "github.com/lworkltd/kits/utils/invoke"
	"github.com/Sirupsen/logrus"
	"encoding/json"
)




type PubUserInfoDTO struct{
	UserId           string            `json:"userId"`
	Mail             string            `json:"mail"`
	Phone            string            `json:"phone"`
	Password         string            `json:"password"`
}


func ClientGetPubAuthUserNumber(feignKey *XFeignKey) (int, code.Error){
	feignkeyBypes,_ := json.Marshal(feignKey)
	rsp, err := invoke.Name(constant.ServiceNamePubAuth).
		Get(constant.PathPubUserNumber).
		Header("X-Feign-Key", string(feignkeyBypes)).
		Query("tenantId", feignKey.TenantId).
		Query("productId", feignKey.ProductId).
		Response()

	var response int
	cerr := invokeutils.ExtractHttpResponse(constant.ServiceNamePubAuth, err, rsp, &response)
	if cerr != nil {
		logrus.WithFields(logrus.Fields{
			"tenantId": feignKey.TenantId,
			"productId": feignKey.ProductId,
			"error": cerr,
			"data": response,
		}).Error("Request query pub-auth user number failed")
		//return "", errcode.CerrRequestPubAuth
		return 0, cerr
	}
	return response, nil
}

func ClientAddPubAuthUser(feignKey *XFeignKey, addPubUserReq *PubUserInfoDTO ) (string, code.Error){
	feignkeyBypes,_ := json.Marshal(feignKey)
	rsp, err := invoke.Name(constant.ServiceNamePubAuth).
	Post(constant.PathPubAuthAddUser).
	Header("X-Feign-Key", string(feignkeyBypes)).
	Json(addPubUserReq).
	Response()

	var response string
	cerr := invokeutils.ExtractHttpResponse(constant.ServiceNamePubAuth, err, rsp, &response)
	if cerr != nil || "" == response{
		logrus.WithFields(logrus.Fields{
			"mail": addPubUserReq.Mail,
			"error": cerr,
			"data": response,
		}).Error("Request add user to pub-auth failed")

		//返回失败的时候去pub-auth删除下，保证没有垃圾数据
		ClientDeletePubAuthUser(feignKey, addPubUserReq)
		//return "", errcode.CerrRequestPubAuth
		return "", cerr
	}

	logrus.WithFields(logrus.Fields{
		"mail":addPubUserReq.Mail,
		"data": response,
	}).Debug("Add user to pub-auth success")
	return response, nil
}

//管理员删除用户（软删除，移至回收站）
func ClientDeletePubAuthUser(feignKey *XFeignKey, addPubUserReq *PubUserInfoDTO ) (string, code.Error){
	feignkeyBypes,_ := json.Marshal(feignKey)
	rsp, err := invoke.Name(constant.ServiceNamePubAuth).
		Post(constant.PathPubTrashUser).
		Header("X-Feign-Key", string(feignkeyBypes)).
		Json(addPubUserReq).
		Response()

	cerr := invokeutils.ExtractHttpResponse(constant.ServiceNamePubAuth, err, rsp, nil)
	if cerr != nil {
		logrus.WithFields(logrus.Fields{
			"userId": addPubUserReq.UserId,
			"error": cerr,
		}).Error("Request trash(delete) user to pub-auth failed")
		//return "", errcode.CerrRequestPubAuth
		return "", cerr
	}

	logrus.WithFields(logrus.Fields{
		"userID":addPubUserReq.UserId,
	}).Debug("Trash(delete) user to pub-auth success")
	return "", nil
}


//管理员修改用户
func clientAdminSetUser(feignKey *XFeignKey, setPubUserReq *PubUserInfoDTO ) (code.Error){
	feignkeyBypes,_ := json.Marshal(feignKey)
	rsp, err := invoke.Name(constant.ServiceNamePubAuth).
		Post(constant.PathPubUserSet).
		Header("X-Feign-Key", string(feignkeyBypes)).
		Json(setPubUserReq).
		Response()

	cerr := invokeutils.ExtractHttpResponse(constant.ServiceNamePubAuth, err, rsp, nil)
	if cerr != nil {
		logrus.WithFields(logrus.Fields{"userId": setPubUserReq.UserId,"error": cerr,}).Error("Request set user to pub-auth failed")
		//return errcode.CerrRequestPubAuth
		return cerr
	}

	logrus.WithFields(logrus.Fields{"userID":setPubUserReq.UserId,}).Debug("Set user to pub-auth success")
	return nil
}

