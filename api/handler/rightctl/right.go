package rightctl

import (
	//"bw/bw-user/constant"
	//"bw/bw-user/errcode"
	"bw/bw-user/api/handler/util"
	"bw/bw-user/errcode"
	"bw/bw-user/model"
	"github.com/gin-gonic/gin"
	"github.com/lworkltd/kits/service/context"
	"github.com/lworkltd/kits/service/restful/code"
	//"encoding/json"
	//"net/url"
	//"github.com/Sirupsen/logrus"
	//"strconv"
)

//获取无上级的权限，其结果集包括child
func GetListTopRights(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	//逻辑处理
	noParentRights, errGetTopRight := model.ProcessListTopRights(feignKey)
	if nil != errGetTopRight {
		srvContext.Errorf("Process Failed,%v", errGetTopRight)
		return nil, errGetTopRight
	}

	return noParentRights, nil
}

//初始化权限树
func InitRight(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	//逻辑处理
	errGetTopRight := model.ProcessInitRight(feignKey)
	if nil != errGetTopRight {
		srvContext.Errorf("Process Failed,%v", errGetTopRight)
		return nil, errGetTopRight
	}

	return nil, nil
}

func CheckIdsRight(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	//1.解析包体
	var keyIdRightDTO model.KeyIdRightDTO
	if err := ctx.BindJSON(&keyIdRightDTO); err != nil {
		srvContext.Errorf("BindJson Failed,%v", err)
		return nil, errcode.CerrApiApiBadJsonPayload(err)
	}
	return model.CheckIdsRight(feignKey, keyIdRightDTO, false)
}

func AddIdsRight(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	//1.解析包体
	var keyIdRightDTO model.KeyIdRightDTO
	if err := ctx.BindJSON(&keyIdRightDTO); err != nil {
		srvContext.Errorf("BindJson Failed,%v", err)
		return nil, errcode.CerrApiApiBadJsonPayload(err)
	}
	return nil, model.AddIdsRight(feignKey, keyIdRightDTO)
}

// 根据角色是否有某个权限设置所有租户各个角色指定权限勾选状态
func AddRoleRightDependOtherRight(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	depend := ctx.Param("depend")
	right := ctx.Param("right")
	if "" == depend || "" == right {
		return nil, errcode.CerrParamater
	}
	return nil, model.AddRightDependOtherRightForRole(feignKey, depend, right)
}
