package levelctl

import (
	//"bw/bw-user/constant"
	"bw/bw-user/errcode"
	"bw/bw-user/api/handler/util"

	"github.com/lworkltd/kits/service/context"
	"github.com/gin-gonic/gin"
	"github.com/lworkltd/kits/service/restful/code"
	"bw/bw-user/model"
	//"encoding/json"
	//"net/url"
	//"github.com/Sirupsen/logrus"
	//"strconv"
	"strconv"
)


//获取所有level列表信息
func GetLevelList(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	//逻辑处理
	levelDetails, errGetLevel := model.ProcessGetLevelList(feignKey)
	if nil != errGetLevel {
		srvContext.Errorf("Process Failed,%v", errGetLevel)
		return nil, errGetLevel
	}

	return levelDetails, nil
}


//通过用户信息获取level列表信息
func GetLevelListByAuthority(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	//逻辑处理
	levelDetails, errProcess := model.ProcessGetLevelListByAuthority(feignKey)
	if nil != errProcess {
		srvContext.Errorf("Process Failed,%v", errProcess)
		return nil, errProcess
	}

	return levelDetails, nil
}


func GetEarningReportByAuthority(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	//逻辑处理
	levelDetails, errProcess := model.ProcessGetEarningReportByAuthority(feignKey)
	if nil != errProcess {
		srvContext.Errorf("Process Failed,%v", errProcess)
		return nil, errProcess
	}

	return levelDetails, nil
}


func Addlevel(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	//解析请求包体
	var addLevel model.LevelDetail
	if err := ctx.BindJSON(&addLevel); err != nil {
		srvContext.Errorf("BindJson Failed,%v", err)
		return nil, errcode.CerrApiApiBadJsonPayload(err)
	}

	//逻辑处理
	errProcess := model.ProcessAddlevel(feignKey, &addLevel)
	if nil != errProcess {
		srvContext.Errorf("Process Failed,%v", errProcess)
		return nil, errProcess
	}

	return "BW_USER_LEVEL_ADD_NOTE", nil
}


func Deletelevel(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	//解析请求包体
	levelId, errParse := strconv.ParseInt(ctx.Query("levelId"), 10, 64)
	if errParse != nil || levelId <= 0 {
		srvContext.Errorf("Delete level parameter error,%v", errParse)
		return nil, errcode.CerrParamater
	}

	//逻辑处理
	errProcess := model.ProcessDeletelevel(feignKey, levelId)
	if nil != errProcess {
		srvContext.Errorf("Process Failed,%v", errProcess)
		return nil, errProcess
	}

	return nil, nil
}


func UpdateLevel(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	//解析请求包体
	var addLevel model.LevelDetail
	if err := ctx.BindJSON(&addLevel); err != nil {
		srvContext.Errorf("BindJson Failed,%v", err)
		return nil, errcode.CerrApiApiBadJsonPayload(err)
	}

	//逻辑处理
	errProcess := model.ProcessUpdatelevel(feignKey, &addLevel)
	if nil != errProcess {
		srvContext.Errorf("Process Failed,%v", errProcess)
		return nil, errProcess
	}

	return nil, nil
}
