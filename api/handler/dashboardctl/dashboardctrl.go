package dashboardctl

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
	"net/url"
)


func GetUserConfig(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	userId, errKey := url.QueryUnescape(ctx.Param("keyUserId"))
	if nil != errKey || "" == userId {
		srvContext.Errorf("GetUserConfig parameter error, userId:%v", userId)
		return  nil, errcode.CerrParamater
	}

	//逻辑处理
	config, err := model.ProcessGetConfigByUserId(feignKey, userId)
	if nil != err {
		srvContext.Errorf("Process Failed,%v", err)
		return nil, err
	}

	return config, nil
}

func SaveUserConfig(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	var config model.DashboardConfig
	if err := ctx.BindJSON(&config); err != nil {
		srvContext.Errorf("BindJson Failed,%v", err)
		return nil, errcode.CerrApiApiBadJsonPayload(err)
	}
	//逻辑处理
	err := model.ProcessSaveConfig(feignKey, &config)
	if nil != err {
		srvContext.Errorf("Process Failed,%v", err)
		return nil, err
	}
	return config, nil
}

func DeleteUserConfig(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	var config model.DashboardConfig
	if err := ctx.BindJSON(&config); err != nil {
		srvContext.Errorf("BindJson Failed,%v", err)
		return nil, errcode.CerrApiApiBadJsonPayload(err)
	}
	//逻辑处理
	return nil, model.ProcessDeleteConfig(feignKey, &config)
}