package statistic

import (
	"bw/bw-user/api/handler/util"
	"bw/bw-user/errcode"
	"bw/bw-user/model"
	"github.com/gin-gonic/gin"
	"github.com/lworkltd/kits/service/context"
	"github.com/lworkltd/kits/service/restful/code"
	"strconv"
)

//根据id列表，查询用户信息
func RemainUserCount(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	timeStr := ctx.Query("time")
	if  "" == timeStr {
		return nil, errcode.CerrParamater
	}

	timeMsc, err := strconv.ParseInt(timeStr, 10, 64)
	if  err != nil {
		return nil, errcode.CerrParamater
	}

	//逻辑处理
	count, errProcess := model.RemainUserCount(feignKey, timeMsc)
	if nil != errProcess {
		srvContext.Errorf("Process Failed,%v", errProcess)
		return nil, errProcess
	}

	return count, nil
}