package rolectl

import (
	"bw/bw-user/api/handler/util"
	"github.com/Sirupsen/logrus"

	//"bw/bw-user/constant"
	"bw/bw-user/errcode"

	"bw/bw-user/model"
	"github.com/gin-gonic/gin"
	"github.com/lworkltd/kits/service/context"
	"github.com/lworkltd/kits/service/restful/code"
	//"encoding/json"
	//"net/url"
	//"github.com/Sirupsen/logrus"
	"strconv"
	"time"
)

//根据ROLE获取子角色列表
func GetRoleChild(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	roleID, err := strconv.ParseInt(ctx.Query("roleId"), 10, 64)
	if err != nil || roleID == 0 {
		return nil, errcode.CerrParamater
	}

	//逻辑处理
	roleSimples, errChild := model.ProcessGetRoleChild(feignKey, roleID)
	if nil != errChild {
		srvContext.Errorf("Process Failed,%v", errChild)
		return nil, errChild
	}

	return roleSimples, nil
}

//查询单个角色详情 子集
func GetRoleChildTree(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	roleID, err := strconv.ParseInt(ctx.Query("roleId"), 10, 64)
	if err != nil || roleID == 0 {
		return nil, errcode.CerrParamater
	}

	//逻辑处理
	nodes, errChild := model.ProcessGetRoleChildTree(feignKey, roleID)
	if nil != errChild {
		srvContext.Errorf("Process Failed,%v", errChild)
		return nil, errChild
	}

	return nodes, nil
}

//当前用户可以在新增/修改用户时设置的角色信息
func FindCurrentSetRole(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	//逻辑处理
	roleSimples, errChild := model.ProcessCurrentSetRole(feignKey)
	if nil != errChild {
		srvContext.Errorf("Process Failed,%v", errChild)
		return nil, errChild
	}

	return roleSimples, nil
}

//查询截止到今天凌晨0点的层级数
func GetLevelCountYesterday(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	todayBeginTimeStr := time.Now().Format("2006-01-02")
	todayBeginTime, errParse := time.Parse("2006-01-02", todayBeginTimeStr)
	if nil != errParse {
		srvContext.Errorf("Process Failed,%v", errParse)
		return nil, errcode.CerrInternal
	}
	//逻辑处理
	levelList, errChild := model.GetLevelListBeforeOneCreateTime(feignKey.TenantId, todayBeginTime)
	if nil != errChild {
		srvContext.Errorf("Process Failed,%v", errChild)
		return nil, errChild
	}

	return len(levelList), nil
}

//查询截止到今天凌晨0点的角色数
func GetRoleCountYesterday(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	todayBeginTimeStr := time.Now().Format("2006-01-02")
	todayBeginTime, errParse := time.Parse("2006-01-02", todayBeginTimeStr)
	if nil != errParse {
		srvContext.Errorf("Process Failed,%v", errParse)
		return nil, errcode.CerrInternal
	}
	//逻辑处理
	roleList, errChild := model.GetRoleListBeforeOneCreateTime(feignKey.TenantId, todayBeginTime)
	if nil != errChild {
		srvContext.Errorf("Process Failed,%v", errChild)
		return nil, errChild
	}

	return len(roleList), nil
}

//当前用户角色具有的权限集合
func GetCurrentRight(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	//逻辑处理
	rightStrList, errRight := model.GetCurrentUserRight(feignKey)
	if nil != errRight {
		srvContext.Errorf("Process Failed,%v", errRight)
		return nil, errRight
	}

	return rightStrList, nil
}

func GetRoleDetails(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	//逻辑处理
	roleDetails, errRole := model.GetRoleDetailWithRightId(feignKey)
	if nil != errRole {
		srvContext.Errorf("Process Failed,%v", errRole)
		return nil, errRole
	}

	return roleDetails, nil
}

func GetRoleSimple(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	//逻辑处理
	roleDetails, errRole := model.GetAllRoleLis(feignKey.TenantId)
	if nil != errRole {
		srvContext.Errorf("Process Failed,%v", errRole)
		return nil, errRole
	}

	var simpleList []model.RoleSimple = make([]model.RoleSimple, 0)
	for index, _ := range roleDetails {
		var simple model.RoleSimple
		simple.Id = roleDetails[index].Id
		simple.Name = roleDetails[index].Name
		simple.EntityNo = roleDetails[index].EntityNo
		simple.RoleTypeId = roleDetails[index].RoleTypeId
		simpleList = append(simpleList, simple)
	}

	return simpleList, nil
}

//当前用户角色具有的权限集合
func GetRightIDsByRoleIDs(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	//解析请求包体
	var idSearchReq model.IdSearchDTO
	if err := ctx.BindJSON(&idSearchReq); err != nil {
		srvContext.Errorf("BindJson Failed,%v", err)
		return nil, errcode.CerrApiApiBadJsonPayload(err)
	}

	//逻辑处理
	rightList, errRight := model.ProcessGetRightIDsByRoleIDs(feignKey, &idSearchReq)
	if nil != errRight {
		srvContext.Errorf("Process Failed,%v", errRight)
		return nil, errRight
	}

	return rightList, nil
}

//根据一个RoleId获取其权限ID列表
func GetRightIDsByRoleID(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	roleIdStr := ctx.Query("roleId")
	roleId, err := strconv.ParseInt(roleIdStr, 10, 64)
	if nil != err || roleId == 0 {
		srvContext.Errorf("parameter roleId abnormal")
		return nil, errcode.CerrParamater
	}

	//逻辑处理
	var idSearchReq model.IdSearchDTO
	idSearchReq.Ids = []string{roleIdStr}
	rightList, errRight := model.ProcessGetRightIDsByRoleIDs(feignKey, &idSearchReq)
	if nil != errRight {
		srvContext.Errorf("Process Failed,%v", errRight)
		return nil, errRight
	}

	return rightList, nil
}

//获取用户权限列表
func GetRightKeyByRoleID(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	roleIdStr := ctx.Query("roleId")
	roleId, err := strconv.ParseInt(roleIdStr, 10, 64)
	if nil != err || roleId == 0 {
		srvContext.Errorf("parameter roleId abnormal")
		return nil, errcode.CerrParamater
	}

	//逻辑处理
	var idSearchReq model.IdSearchDTO
	idSearchReq.Ids = []string{roleIdStr}
	rightList, errRight := model.GetRoleRight(feignKey, roleId)
	if nil != errRight {
		srvContext.Errorf("Process Failed,%v", errRight)
		return nil, errRight
	}

	return rightList, nil
}

func UpsertRole(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	parentId, errParse := strconv.ParseInt(ctx.Query("parentId"), 10, 64)
	if nil != errParse {
		srvContext.Errorf("Parse parentId failed,%v, parentID:%v", errParse, parentId)
		return nil, errcode.CerrParamater
	}

	//解析请求包体
	var upsetReq model.RoleDTO
	if err := ctx.BindJSON(&upsetReq); err != nil {
		srvContext.Errorf("BindJson Failed,%v", err)
		return nil, errcode.CerrApiApiBadJsonPayload(err)
	}

	//检查parent
	if parentId != 0 {
		parent, errGetRole := model.GetRoleInfoByID(feignKey.TenantId, parentId)
		if nil != errGetRole {
			srvContext.Errorf("Get parent info faile,%v", errGetRole)
			return nil, errGetRole
		} else if nil == parent {
			srvContext.Errorf("Parent not exist, parentId:%v", parentId)
			return nil, errcode.CerrRoleIDNotExist
		}
	}

	//逻辑处理
	var errProcess code.Error
	if 0 == upsetReq.Id {
		errProcess = model.ProcessAddRole(feignKey, &upsetReq, parentId)
	} else {
		errProcess = model.ProcessUpdateRole(feignKey, &upsetReq, parentId)
	}

	return nil, errProcess
}

//删除角色
func RemoveRole(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	roleId, errParse := strconv.ParseInt(ctx.Query("roleId"), 10, 64)
	if nil != errParse || roleId == 0 {
		srvContext.Errorf("Parse roleId failed,%v", errParse)
		return nil, errcode.CerrParamater
	}

	//无请求包体
	//逻辑处理
	errProcess := model.ProcessRemoveRole(feignKey, roleId)
	return nil, errProcess
}

//查询具体任务处理权限角色
func GetSimpleRoleByRight(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	rightEntityNo := ctx.Query("rightEntityNo")
	if "" == rightEntityNo {
		srvContext.Errorf("query parameter rightEntityNo is empty")
		return nil, errcode.CerrParamater
	}

	//无解析请求包体
	//逻辑处理

	roleList, errProcess := model.ProcessGetSimpleRoleByRight(feignKey, rightEntityNo)
	if nil != errProcess {
		srvContext.Errorf("ProcessGetSimpleRoleByRight failed, error:%v", errProcess)
		return nil, errProcess
	}

	return roleList, nil
}

// 获取角色类型
func GetRoleTypeList(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	dtos := make([]model.RoleTypeDTO, 0)
	roleTypes, errGet := model.GetAllRoleType()
	if nil != errGet {
		return nil, errGet
	}
	for _, item := range roleTypes {
		dtos = append(dtos, model.RoleTypeDTO{Id: item.Id, TypeName: item.TypeName[feignKey.Language]})
	}
	return dtos, nil
}

// 获取角色类型对应的权限ID列表
func GetRightByRoleTypeId(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	roleTypeId := ctx.Param("roleTypeId")
	roleType, typeErr := model.GetRoleTypeById(roleTypeId)
	if nil != typeErr {
		return nil, typeErr
	}
	if len(roleType.EntityNo) == 0 {
		return nil, nil
	}
	rights, rightErr := model.GetRightByEntityNos(roleType.EntityNo, feignKey.TenantId)
	if nil != rightErr {
		return nil, rightErr
	}
	rightIds := make([]int64, 0)
	for _, item := range rights {
		rightIds = append(rightIds, item.Id)
	}
	return rightIds, nil
}

// 刷新角色，补全角色类型
func FreshRole4RoleType(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	return nil, model.FreshRoleType(feignKey.TenantId)
}

//自定义获取角色信息
func SearchRoleSimpleInfo(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	//1.解析包体
	var searchReq model.SearchDTO
	if err := ctx.BindJSON(&searchReq); err != nil {
		srvContext.Errorf("BindJson Failed,%v", err)
		return nil, errcode.CerrApiApiBadJsonPayload(err)
	}

	//2.查询角色类型
	allRoleType, err := model.GetAllRoleType()
	if err != nil {
		logrus.WithFields(logrus.Fields{"error": err}).Error("GetAllRoleType failed")
		return nil, err
	}
	roleTypeMap := make(map[string]model.RoleType)
	for _, item := range allRoleType {
		roleTypeMap[item.Id] = item
	}

	//3.查询角色记录
	roleRecords, err := model.SearchRoleByCondition(feignKey, &searchReq)
	if err != nil {
		logrus.WithFields(logrus.Fields{"error": err}).Error("SearchRoleByCondition failed")
		return nil, err
	}

	//4.转换输出结果
	response := make([]*model.RoleSimple, 0)
	for _, role := range roleRecords {
		roleSimple := new(model.RoleSimple)
		roleSimple.Id = role.Id
		roleSimple.Name = role.Name
		roleSimple.EntityNo = role.EntityNo
		roleSimple.RoleTypeId = role.RoleTypeId

		if nil != &role.RoleTypeId && "" != role.RoleTypeId {
			roleType := roleTypeMap[role.RoleTypeId]
			if nil != &roleType {
				roleSimple.RoleTypeName = roleType.TypeName[feignKey.Language]
			}
		}
		response = append(response, roleSimple)
	}
	return response, nil
}
