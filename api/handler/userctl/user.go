package userctl

import (
	"bw/bw-user/api/handler/util"
	"bw/bw-user/constant"
	"bw/bw-user/errcode"
	"log"

	"bw/bw-user/model"

	"github.com/gin-gonic/gin"
	"github.com/lworkltd/kits/service/context"
	"github.com/lworkltd/kits/service/restful/code"

	//"encoding/json"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
)

//私有化检查
func Check(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	domain := ctx.Query("domain")
	if "" == domain {
		srvContext.Errorf("parameter domain is empty")
		return nil, errcode.CerrParamater
	}
	return model.CheckTenant(feignKey, domain)
}

//根据id列表，查询用户信息
func ListUserByIds(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	//解析请求包体
	var reqUserIDs []int64 = make([]int64, 0)
	if err := ctx.BindJSON(&reqUserIDs); err != nil {
		srvContext.Errorf("BindJson Failed,%v", err)
		return nil, errcode.CerrApiApiBadJsonPayload(err)
	}
	if len(reqUserIDs) == 0 {
		srvContext.Errorf("Request Ids is empty")
		return nil, errcode.CerrParamater
	}

	//逻辑处理
	userDetails, errProcess := model.ProcessListUserInfosByKey(feignKey, reqUserIDs, constant.QueryKeyId)
	if nil != errProcess {
		srvContext.Errorf("Process Failed,%v", errProcess)
		return nil, errProcess
	}

	return userDetails, nil
}

//获取当前登录用户的信息
func ObtainCurrentUserInfo(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	//无请求包体，逻辑处理
	userDetail, errProcess := model.GetUserDetailByIdOrPubid(feignKey, feignKey.UserId, 0)
	roleTypeId, roleTypeName := model.GetRoleTypeInfo4User(feignKey.TenantId, userDetail.RoleId)
	userDetail.RoleTypeId = roleTypeId
	userDetail.RoleTypeName = roleTypeName
	if nil != errProcess {
		srvContext.Errorf("Process Failed,%v", errProcess)
		return nil, errProcess
	}

	return userDetail, nil
}

//根据pubUserId获取用户的信息
func GetUserInfoByPubid(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {

	pubUserId := ctx.Query("pubUserId")
	logrus.Infof("------>  GetUserInfoByPubid pubUserId:%v", pubUserId)
	if "" == pubUserId {
		srvContext.Errorf("parameter pubUserId is empty")
		return nil, errcode.CerrParamater
	}
	//无请求包体，逻辑处理
	userDetail, errProcess := model.GetUserDetailByIdOrPubid(feignKey, pubUserId, 0)
	if nil != errProcess {
		srvContext.Errorf("Process Failed,%v", errProcess)
		return nil, errProcess
	}
	logrus.Infof("------>  GetUserInfoByPubid userDetail:%v", userDetail)
	return userDetail, nil
}

//增加用户
func AddUser(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	domain, errURLDecode := url.QueryUnescape(ctx.Query("domain"))
	if nil != errURLDecode {
		logrus.WithFields(logrus.Fields{
			"error": errURLDecode,
		}).Warn("QueryUnescape domain error")
	}

	//解析请求包体
	var userInfo model.BWUserDTO
	logrus.Infof("------>  userInfo:%v", userInfo)
	if err := ctx.BindJSON(&userInfo); err != nil {
		srvContext.Errorf("BindJson Failed,%v", err)
		return nil, errcode.CerrApiApiBadJsonPayload(err)
	}

	//逻辑处理
	userDetail, errProcess := model.ProcessAddUser(&userInfo, feignKey, domain)
	if nil != errProcess {
		srvContext.Errorf("Process Failed,%v", errProcess)
		return nil, errProcess
	}

	model.SendUserMsg(feignKey.TenantId, strconv.FormatInt(userInfo.Id, 10), constant.BwEvent_ADD, userDetail)
	return userDetail, nil
}

//根据id列表，删除用户
func DeleteUser(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	//解析请求包体
	var reqUserIDs []int64 = make([]int64, 0)
	if err := ctx.BindJSON(&reqUserIDs); err != nil {
		srvContext.Errorf("BindJson Failed,%v", err)
		return nil, errcode.CerrApiApiBadJsonPayload(err)
	}
	if len(reqUserIDs) == 0 {
		return nil, errcode.CerrParamater
	}

	//查询是否有关联用户和账户
	var queryDTO model.QueryByIdDTO
	userIds := make([]string, 0)
	for _, userId := range reqUserIDs {
		userIds = append(userIds, strconv.FormatInt(userId, 10))
	}
	queryDTO.OwnIds = userIds
	queryDTO.Columns = []string{"oweId"}
	resultList, errCustomer := model.ClientCustomerPropertirsById(&queryDTO, feignKey)
	if nil != errCustomer {
		logrus.WithFields(logrus.Fields{"error": errCustomer}).Error("ClientCustomerPropertirsById failed")
		return nil, errCustomer
	}
	if len(resultList) > 0 {
		logrus.WithFields(logrus.Fields{"Ids": reqUserIDs}).Error("Reference Customer Not Allow Delete")
		return nil, errcode.CerrReferenceCustomerNotAllowDelete
	}

	queryDTO.OwnIds = []string{}
	queryDTO.BindUserIds = userIds
	queryDTO.Columns = []string{"bindUserId"}
	resultList, errCustomer = model.ClientCustomerPropertirsById(&queryDTO, feignKey)
	if nil != errCustomer {
		logrus.WithFields(logrus.Fields{"error": errCustomer}).Error("ClientCustomerPropertirsById failed")
		return nil, errCustomer
	}
	if len(resultList) > 0 {
		logrus.WithFields(logrus.Fields{"Ids": reqUserIDs}).Error("Bind Customer Not Allow Delete")
		return nil, errcode.CerrReferenceCustomerNotAllowDelete
	}

	check, errCheck := model.ClientCheckUserAsOwner(userIds, feignKey)
	if nil != errCheck {
		logrus.WithFields(logrus.Fields{"error": errCheck}).Error("ClientCheckUserAsOwner failed")
		return nil, errCheck
	}
	if true == check {
		logrus.WithFields(logrus.Fields{"Ids": reqUserIDs}).Error("Reference Account Not Allow Delete")
		return nil, errcode.CerrReferenceAccountNotAllowDelete
	}

	for _, userIdNum := range reqUserIDs {
		errProcess := model.ProcessDeleteUserById(feignKey, userIdNum)
		if nil != errProcess {
			logrus.Errorf("Delete one user failed, userID:%v, error:%v, operation userID:%v", userIdNum, errProcess, feignKey.UserId)
			return nil, errProcess
		}
		logrus.Infof("Delete one user success, deleted userID:%v, operation userID:%v", userIdNum, feignKey.UserId)
		model.SendUserMsg(feignKey.TenantId, strconv.FormatInt(userIdNum, 10), constant.BwEvent_DELETE, nil)
	}
	return nil, nil
}

//判断用户是否已经存在
func UserExists(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {

	key := ctx.Query("key")
	value := ctx.Query("value")
	logrus.Infof("feignKey:%v  key:%v  value:%v", feignKey, key, value)
	result, errExist := model.ExistInUserDetailTab(feignKey.TenantId, key, value)
	if nil != errExist {
		logrus.Errorf("ExistInUserDetailTab failed, key:%v, value:%v, error:%v", key, value, errExist)
		return nil, errExist
	}

	return result, nil
}

//更新用户激活标记，/v1/user/updateActiveById?userId=230&isActive=false
func UpdateActiveById(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	userId, errParseUserId := strconv.ParseInt(ctx.Query("userId"), 10, 64)
	isActive, errParseIsActive := strconv.ParseBool(ctx.Query("isActive"))
	if nil != errParseUserId || nil != errParseIsActive {
		logrus.Errorf("Request parameter error")
		return nil, errcode.CerrParamater
	}

	var intActive int = 1
	if false == isActive {
		intActive = 0
	}

	errProcess := model.ProcessUserUpdateOneFieldById(userId, "active", intActive, feignKey)
	if nil != errProcess {
		logrus.Errorf("ProcessUpdateActiveById failed, userID:%v, isActive:%v", userId, isActive)
		return false, errProcess
	}
	return nil, errProcess
}

//根据id列表，查询用户信息
func AddAdminUser(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {

	email, errURLemail := url.QueryUnescape(ctx.Query("email"))
	name, errURLname := url.QueryUnescape(ctx.Query("name"))
	language := ctx.Query("language")
	password := ctx.Query("password")
	if nil != errURLemail || nil != errURLname || "" == email || "" == name || "" == password {
		logrus.WithFields(logrus.Fields{}).Error("Parameter error")
		return nil, errcode.CerrParamater
	}

	//无请求包体

	//逻辑处理
	userId, errProcess := model.ProcessAddAdminUser(feignKey, name, email, password, language)
	if nil != errProcess {
		srvContext.Errorf("Process Failed,%v", errProcess)
		return nil, errProcess
	}
	//初始化角色列表
	model.ProcessInitRoleList(feignKey, language)

	return userId, nil
}

//查询当前用户直属下级用户
func GetBelongUserId(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	//无请求包体，无参数
	curentUserInfo, errGetUser := model.GetUserDetailByIdOrPubid(feignKey, feignKey.UserId, 0)
	if errGetUser != nil || nil == curentUserInfo {
		srvContext.Errorf("Get Curent User Info Failed, error:%v, pubUserId:%v", errGetUser, feignKey.UserId)
		return nil, errGetUser
	} else if nil == curentUserInfo {
		srvContext.Errorf("Not Found User Info, pubUserId:%v", feignKey.UserId)
		return nil, errcode.CerrAccountNotFound
	}

	//逻辑处理
	userRecords, errProcess := model.GetBelongUserRecord(feignKey, []int64{curentUserInfo.Id})
	if nil != errProcess {
		srvContext.Errorf("Process Failed,%v", errProcess)
		return nil, errProcess
	}

	var childIds []int64 = make([]int64, 0)
	for index, _ := range userRecords {
		childIds = append(childIds, userRecords[index].Id)
	}

	return childIds, nil
}

//查询指定用户直属下级用户
func GetUserBelongUser(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	userId, errParseUserId := strconv.ParseInt(ctx.Query("userId"), 10, 64)
	if nil != errParseUserId || userId < 0 {
		return nil, errcode.CerrParamater
	}

	//逻辑处理
	userRecords, errProcess := model.GetBelongUserRecord(feignKey, []int64{userId})
	if nil != errProcess {
		srvContext.Errorf("Process Failed,%v", errProcess)
		return nil, errProcess
	}

	//userRecords 转换为userDetails
	userDetails, errCov := model.UserRecords2UserDetails(userRecords)
	if nil != errCov {
		logrus.WithFields(logrus.Fields{"userId": userId}).Warn("Convert user info from []UserDetailTab to []UserDetail failed")
		return userDetails, errcode.CerrInternal
	}
	return userDetails, nil
}

//根据key和value列表，返回用户详情信息
func GetUserInfosByKey(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	keyValue, errKey := url.QueryUnescape(ctx.Param("keyValue"))
	keyType := ctx.Param("keyType")
	if nil != errKey || "" == keyValue || "" == keyType {
		srvContext.Errorf("GetUserinfosByKey parameter error, keyValue:%v, keyType:%v", ctx.Param("keyValue"), keyType)
		return nil, errcode.CerrParamater
	}

	//keyValue格式：value1,value2,value3
	var intValueList []int64
	var strValueList []string
	var errParseValue code.Error
	var userDetails []model.UserDetail
	var errProcess code.Error

	if constant.QueryKeyId == keyType || constant.QueryKeyRoleId == keyType || constant.QueryKeyLevelId == keyType {
		intValueList, errParseValue = model.SplitToIntList(keyValue, ",") //解析keyValue参数
		if nil != errParseValue || len(intValueList) == 0 {
			srvContext.Errorf("GetUserinfosByKey parameter error, keyValue:%v, keyType:%v", ctx.Param("keyValue"), keyType)
			return nil, errcode.CerrParamater
		}
		//获取记录详情
		userDetails, errProcess = model.ProcessListUserInfosByKey(feignKey, intValueList, keyType)
	} else if constant.QueryKeyEmail == keyType ||
		constant.QueryKeyEntityNo == keyType ||
		constant.QueryKeyName == keyType ||
		constant.QueryKeyIdNum == keyType ||
		constant.QueryKeyLogin == keyType ||
		constant.QueryKeypubUserIds == keyType {
		strValueList, errParseValue = model.SplitToStringList(keyValue, ",") //解析keyValue参数
		if nil != errParseValue || len(strValueList) == 0 {
			srvContext.Errorf("GetUserinfosByKey parameter error, keyValue:%v, keyType:%v", ctx.Param("keyValue"), keyType)
			return nil, errcode.CerrParamater
		}
		//获取记录详情
		userDetails, errProcess = model.ProcessListUserInfosByKey(feignKey, strValueList, keyType)
	} else if constant.QueryKeyAllRecords == keyType {
		//获取记录详情
		logrus.Info("获取记录详情")
		userDetails, errProcess = model.ProcessListUserInfosByKey(feignKey, keyValue, keyType)
	} else {
		srvContext.Errorf("GetUserinfosByKey keytype not support, keyType:%v", keyType)
		return nil, errcode.CerrQueryKeyNotSupport
	}

	if nil != errProcess {
		srvContext.Errorf("ProcessListUserInfosByKey Failed,%v", errProcess)
		return nil, errProcess
	}

	return userDetails, nil
}

//根据key和value列表，返回用户详情信息--Post
func FindUserInfosByKey(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	var userKeyDTO model.UserKeyDTO
	if err := ctx.BindJSON(&userKeyDTO); err != nil {
		srvContext.Errorf("BindJson Failed,%v", err)
		return nil, errcode.CerrApiApiBadJsonPayload(err)
	}
	if 0 == len(userKeyDTO.KeyValue) || "" == userKeyDTO.KeyType {
		srvContext.Errorf("GetUserinfosByKey parameter error, keyValue:%v, keyType:%v", userKeyDTO.KeyValue, userKeyDTO.KeyType)
		return nil, errcode.CerrParamater
	}
	userDetails, errProcess := model.ProcessListUserInfosByKey(feignKey, userKeyDTO.KeyValue, userKeyDTO.KeyType)
	if nil != errProcess {
		srvContext.Errorf("ProcessListUserInfosByKey Failed,%v", errProcess)
		return nil, errProcess
	}
	return userDetails, nil
}

//查询无parent的用户id
func GetNoParentuserIds(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	//无请求包体和请求参数
	userRecords, errProcess := model.ProcessListNoParentUsers(feignKey)
	if nil != errProcess {
		srvContext.Errorf("ProcessListNoParentUsers Failed,%v", errProcess)
		return nil, errProcess
	}

	var noParentIds []int64 = make([]int64, 0)
	for index, _ := range userRecords {
		noParentIds = append(noParentIds, userRecords[index].Id)
	}
	return noParentIds, nil
}

//查询所有用户的简单信息（含Iid，name，entityNo）
func ListSimpleUser(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	//无请求包体和请求参数
	simplesRecords, errProcess := model.ProcessListSimpleUser(feignKey)
	if nil != errProcess {
		srvContext.Errorf("ProcessListNoParentUsers Failed,%v", errProcess)
		return nil, errProcess
	}

	return simplesRecords, nil
}

//查询有返佣账户的用户的信息（ID,name）
func ListSimpleUserHasAccountUser(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	//无请求包体和请求参数
	simplesRecords, errProcess := model.ProcessListSimpleUserHasAccountUser(feignKey)
	if nil != errProcess {
		srvContext.Errorf("ProcessListNoParentUsers Failed,%v", errProcess)
		return nil, errProcess
	}

	return simplesRecords, nil
}

//查询全量用户详细信息
func ListAllUserDetail(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	//无请求包体和请求参数
	allUserRecords, errProcess := model.GetAllUserRecords(feignKey)
	if nil != errProcess {
		srvContext.Errorf("GetAllUserRecords Failed,%v", errProcess)
		return nil, errProcess
	}

	return model.UserRecords2UserDetails(allUserRecords)
}

//查询所有用户的简单信息（含id,name,parent,levelId,levelName,roleId,roleName,entityNo,pubUserId,login,vendorServerId）
func ListUserAndLevel(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	//无请求包体和请求参数
	simplesRecords, errProcess := model.ProcessListUserAndLevel(feignKey)
	if nil != errProcess {
		srvContext.Errorf("ProcessListNoParentUsers Failed,%v", errProcess)
		return nil, errProcess
	}

	return simplesRecords, nil
}

//查询一个用户的简单信息（含id,name,parent,levelId,levelName,roleId,roleName,entityNo,pubUserId,login,vendorServerId）
func GetOneSimpleUser(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	keyValue := ctx.Query("key")
	typestr := ctx.Query("type")
	var userRecords []model.UserDetailTab
	var errProces code.Error = nil
	if constant.QueryKeyId == typestr || constant.QueryKeyRoleId == typestr || constant.QueryKeyLevelId == typestr {
		intKeyValue, errParse := strconv.ParseInt(keyValue, 10, 64)
		if nil != errParse {
			srvContext.Errorf("GetOneSimpleUser parse key to int Failed, error:%v, type:%v, key:%v", errParse, typestr, keyValue)
			return nil, errcode.CerrParamater
		}
		userRecords, errProces = model.ProcessListUserRecordsByKey(feignKey, []int64{intKeyValue}, typestr)
	} else if constant.QueryKeyName == typestr || constant.QueryKeyLogin == typestr || constant.QueryKeyPhone == typestr || constant.QueryKeyEntityNo == typestr || constant.QueryKeyEmail == typestr {
		strKeyValue, errUnescape := url.QueryUnescape(keyValue)
		if "" == keyValue || nil != errUnescape {
			srvContext.Errorf("GetOneSimpleUser, parameter key abnormal, type:%v, key:%v", typestr, keyValue)
			return nil, errcode.CerrParamater
		}
		userRecords, errProces = model.ProcessListUserRecordsByKey(feignKey, []string{strKeyValue}, typestr)
	} else {
		srvContext.Errorf("GetOneSimpleUser, parameter type abnormal, type:%v, key:%v", typestr, keyValue)
		return nil, errcode.CerrParamater
	}

	if nil != errProces {
		srvContext.Errorf("ProcessListUserRecordsByKey failed, parameter, type:%v, key:%v", typestr, keyValue)
		return nil, errProces
	} else if len(userRecords) == 0 {
		return nil, errcode.CerrAccountNotFound
	}

	simplesRecords, errConv := model.UserRecord2SimpleUser(&userRecords[0]) //只返回第一条记录
	return simplesRecords, errConv
}

//查询截止到今天凌晨0点的用户数
func GetUserCountYesterday(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	todayBeginTimeStr := time.Now().Format("2006-01-02")
	todayBeginTime, errParse := time.Parse("2006-01-02", todayBeginTimeStr)
	if nil != errParse {
		srvContext.Errorf("Process Failed,%v", errParse)
		return nil, errcode.CerrInternal
	}
	//逻辑处理
	counter, errChild := model.GetUserCountBeforeOneCreateTime(feignKey.TenantId, todayBeginTime)
	if nil != errChild {
		srvContext.Errorf("Process Failed,%v", errChild)
		return nil, errChild
	}

	return counter, nil
}

//查询登录租户的用户数
func GetTenantUserCount(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	//逻辑处理
	log.Printf("查询登录租户的用户数")
	counter, errChild := model.GetUserCount(feignKey.TenantId)
	if nil != errChild {
		srvContext.Errorf("Process Failed,%v", errChild)
		return nil, errChild
	}

	return counter, nil
}

//用户树查询下级信息接口, 根据不同权限和模块显示不同返回
func GetUserTreeChild(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	strUserId := ctx.Query("userId")
	module := ctx.Query("module")
	var intUserId int64 = 0
	var errParseUserId error = nil

	if "" != strUserId {
		intUserId, errParseUserId = strconv.ParseInt(ctx.Query("userId"), 10, 64)
	}
	if "" == module {
		module = constant.ModuleUser //默认为User
	}
	if nil != errParseUserId || intUserId < 0 || (constant.ModuleUser != module && constant.ModuleAccount != module && constant.ModuleAccountReport != module &&
		constant.ModuleCommissionReport != module && constant.ModuleCustomer != module && constant.ModuleEarningReport != module) {
		srvContext.Errorf("Parameter invalid, userId:%v, module:%v", strUserId, module)
		return nil, errcode.CerrParamater
	}

	//逻辑处理
	treeNodes, errChild := model.ProcessUserTreeChildByModuleRight(feignKey, intUserId, module)
	if nil != errChild {
		srvContext.Errorf("ProcessUserTreeChildByModuleRight Failed,%v", errChild)
		return nil, errChild
	}
	//去除id为1的用户信息
	var result = make([]*model.LazyTreeNodeDTO, 0)
	for index, _ := range treeNodes {
		if "1" == treeNodes[index].Value {
			continue
		}
		result = append(result, treeNodes[index])
	}
	sort.Sort(model.LazyTreeNodeDTOSlice(result))
	return result, nil
}

//用户树查询下级完整用户数信息接口
func GetUserTree(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	strUserId := ctx.Query("userId")
	var intUserId int64 = 0
	var errParseUserId error = nil

	if "" != strUserId {
		intUserId, errParseUserId = strconv.ParseInt(ctx.Query("userId"), 10, 64)
	}

	if nil != errParseUserId || intUserId < 0 {
		srvContext.Errorf("Process Failed, userId:%v", strUserId)
		return nil, errcode.CerrParamater
	}

	//逻辑处理
	return model.ProcessUserTree(feignKey, intUserId)
}

//用户树搜索，根据指定用户ID搜索，返回完整的树型结构
func SearchUserTree(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	strUserId := ctx.Query("userId")
	module := ctx.Query("module")
	var intUserId int64 = 0
	var errParseUserId error = nil

	if "" != strUserId {
		intUserId, errParseUserId = strconv.ParseInt(ctx.Query("userId"), 10, 64)
	}
	if "" == module {
		module = constant.ModuleUser //默认为User
	}
	if nil != errParseUserId || intUserId < 0 || (constant.ModuleUser != module && constant.ModuleAccount != module && constant.ModuleAccountReport != module &&
		constant.ModuleCommissionReport != module && constant.ModuleCustomer != module && constant.ModuleEarningReport != module) {
		srvContext.Errorf("Parameter invalid, userId:%v, module:%v", strUserId, module)
		return nil, errcode.CerrParamater
	}

	treeNodes, err := model.BuildTargetUserTreeByRight(feignKey, intUserId, module)

	//去除id为1的用户信息
	var result = make([]*model.LazyTreeNodeDTO, 0)
	for index, _ := range treeNodes {
		if "1" == treeNodes[index].Value {
			continue
		}
		result = append(result, treeNodes[index])
	}
	sort.Sort(model.LazyTreeNodeDTOSlice(result))

	//逻辑处理
	return result, err
}

//根据条件查询用户的简单信息（含id,name,parent,levelId,levelName,roleId,roleName,entityNo,pubUserId,login,vendorServerId）
//目前只支持QueryType、QueryContent、PageNo、Size四个字段搜索
func GetSimpleUserByPage(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	//解析包体
	var searchReq model.UserSearchDTO
	if err := ctx.BindJSON(&searchReq); err != nil {
		srvContext.Errorf("BindJson Failed,%v", err)
		return nil, errcode.CerrApiApiBadJsonPayload(err)
	}

	//修正参数
	if searchReq.PageNo <= 0 { //页码序号从1开始
		searchReq.PageNo = 1
	}
	if searchReq.Size <= 0 { //每一页的数量
		searchReq.Size = 10
	}

	//逻辑处理
	userPageInfo, errChild := model.ProcessSimpleUserByPage(feignKey, &searchReq)
	if nil != errChild {
		srvContext.Errorf("Process Failed,%v", errChild)
		return nil, errChild
	}

	return userPageInfo, nil
}

//模糊搜索查询用户指定字段值
func FindUserByField(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	//解析包体
	var searchReq model.UserFieldSearchDTO
	if err := ctx.BindJSON(&searchReq); err != nil {
		srvContext.Errorf("BindJson Failed,%v", err)
		return nil, errcode.CerrApiApiBadJsonPayload(err)
	}
	module := ctx.Query("module")
	if "" == module {
		module = constant.ModuleUser
	}

	includeAdminStr := ctx.Query("includeAdmin")
	includeAdmin := false
	if "" == includeAdminStr {
		includeAdmin = false
	} else {
		includeAdminTemp, err := strconv.ParseBool(includeAdminStr)
		if err == nil {
			includeAdmin = includeAdminTemp
		}
	}

	//逻辑处理
	simpleList, errProcess := model.ProcessFindUserByField(feignKey, &searchReq, module)
	if nil != errProcess {
		srvContext.Errorf("Process Failed,%v", errProcess)
		return nil, errProcess
	}

	//去除id为1的用户信息
	var result = make([]model.SimpleUserDTO, 0)
	for index := range simpleList {
		if !includeAdmin {
			if 1 == simpleList[index].Id {
				continue
			}
		}
		result = append(result, simpleList[index])
	}

	return result, nil
}

//根据模块检查指定用户ID是否在权限范围内
func CheckUserIdPermissionScope(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	targetUserId, errParse := strconv.ParseInt(ctx.Param("targetUserId"), 10, 64)
	module := ctx.Query("module")
	if nil != errParse {
		srvContext.Errorf("parse target User Id failed,%v", errParse)
		return nil, errcode.CerrParamater
	}
	if "" == module {
		module = constant.ModuleUser
	}

	//逻辑处理
	result, errProcess := model.ProcessCheckUserIdPermissionScope(feignKey, targetUserId, module)
	if nil != errProcess {
		srvContext.Errorf("Process Failed,%v", errProcess)
		return nil, errProcess
	}

	return result, nil
}

//模糊搜索查询权限范围内的用户列表，包括用户名、角色名、层级名任意一个匹配（有权限过滤）
func FindLikeNameWithRight(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	includeAdmin, errParse := strconv.ParseBool(ctx.DefaultQuery("includeAdmin", "true"))
	if nil != errParse {
		srvContext.Errorf("parse inculdeAdmin failed, inculdeAdmin:%v", ctx.Query("includeAdmin"))
		return nil, errcode.CerrParamater
	}

	//解析包体
	var searchReq model.UserNameSearchDTO
	if err := ctx.BindJSON(&searchReq); err != nil {
		srvContext.Errorf("BindJson Failed,%v", err)
		return nil, errcode.CerrApiApiBadJsonPayload(err)
	}

	//逻辑处理
	simpleList, errProcess := model.ProcessFindLikeNameWithRight(feignKey, &searchReq, includeAdmin)
	if nil != errProcess {
		srvContext.Errorf("Process Failed,%v", errProcess)
		return nil, errProcess
	}

	return simpleList, nil
}

//模糊搜索查询权限范围内的角色、层级、用户
func FindLikeRoleLevelUserWithRight(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	includeAdmin, errParse := strconv.ParseBool(ctx.DefaultQuery("includeAdmin", "true"))
	if nil != errParse {
		srvContext.Errorf("parse inculdeAdmin failed, inculdeAdmin:%v", ctx.Query("includeAdmin"))
		return nil, errcode.CerrParamater
	}

	//解析包体
	var searchReq model.UserNameSearchDTO
	if err := ctx.BindJSON(&searchReq); err != nil {
		srvContext.Errorf("BindJson Failed,%v", err)
		return nil, errcode.CerrApiApiBadJsonPayload(err)
	}

	//逻辑处理
	simpleList, errProcess := model.ProcessFindRoleLevelUserLikeNameWithRight(feignKey, &searchReq, includeAdmin)
	if nil != errProcess {
		srvContext.Errorf("Process Failed,%v", errProcess)
		return nil, errProcess
	}

	return simpleList, nil
}

//返佣用户查询
func GetSimpleUserCommissionRight(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	//解析包体
	var searchReq model.FuzzyConditionDTO
	if err := ctx.BindJSON(&searchReq); err != nil {
		srvContext.Errorf("BindJson Failed,%v", err)
		return nil, errcode.CerrApiApiBadJsonPayload(err)
	}

	//逻辑处理
	simpleList, errProcess := model.ProcessGetSimpleUserCommissionRight(feignKey, &searchReq)
	if nil != errProcess {
		srvContext.Errorf("Process Failed,%v", errProcess)
		return nil, errProcess
	}

	return simpleList, nil
}

//根据模块和当前用户权限，模糊搜索用户名字和编码信息
func GetSimpleUserByModule(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	name, err := url.QueryUnescape(ctx.Query("name")) //实际上按name和EntitoNo搜索
	if nil != err || "" == name {
		return nil, errcode.CerrParamater
	}
	module, errModule := url.QueryUnescape(ctx.Query("module")) //模块
	if nil != errModule || "" == module {
		return nil, errcode.CerrParamater
	}

	//逻辑处理
	simpleList, errProcess := model.ProcessGetSimpleUserByModuleRight(feignKey, name, module)
	if nil != errProcess {
		srvContext.Errorf("Process Failed,%v", errProcess)
		return nil, errProcess
	}

	return simpleList, nil
}

//
////返佣账户查询，注意：不建议使用，请使用
//func GetSimpleUserAccRight(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
//	name, err := url.QueryUnescape(ctx.Query("name"))			//实际上按name和EntitoNo搜索
//	if nil != err || "" == name {
//		return nil, errcode.CerrParamater
//	}
//
//	//逻辑处理
//	simpleList, errProcess := model.ProcessGetSimpleUserAccRight(feignKey, name)
//	if nil != errProcess {
//		srvContext.Errorf("Process Failed,%v", errProcess)
//		return nil, errProcess
//	}
//
//	return simpleList, nil
//}

//批量归属,  规则
//:不可跨层级归属
//:上级必须归属于当前层级之上
func UpdateParentBatch(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	parentId, errParse := strconv.ParseInt(ctx.Query("parentId"), 10, 64)
	if nil != errParse {
		srvContext.Errorf("parse ParentId failed, parentId:%v", ctx.Query("parentId"))
		return nil, errcode.CerrParamater
	}

	//解析包体
	var ids []int64 = make([]int64, 0)
	if err := ctx.BindJSON(&ids); err != nil {
		srvContext.Errorf("BindJson Failed,%v", err)
		return nil, errcode.CerrApiApiBadJsonPayload(err)
	}

	if parentId <= 0 || len(ids) == 0 {
		srvContext.Errorf("ParentId or ids abnormal")
		return nil, errcode.CerrParamater
	}

	//逻辑处理
	errProcess := model.ProcessUpdateParentBatch(feignKey, parentId, ids)
	if nil != errProcess {
		srvContext.Errorf("Process Failed,%v", errProcess)
		return nil, errProcess
	}

	return nil, nil
}

//根据条件查询用户的简单信息，有权限过滤
func GetUserDetailByPage(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	//解析包体
	var searchReq model.UserSearchDTO
	if err := ctx.BindJSON(&searchReq); err != nil {
		srvContext.Errorf("BindJson Failed,%v", err)
		return nil, errcode.CerrApiApiBadJsonPayload(err)
	}

	//修正参数
	if searchReq.PageNo <= 0 { //页码序号从1开始
		searchReq.PageNo = 1
	}
	if searchReq.Size <= 0 { //每一页的数量
		searchReq.Size = 10
	}

	//逻辑处理
	userPageInfo, errChild := model.ProcessUserDetailByPage(feignKey, &searchReq)
	if nil != errChild {
		srvContext.Errorf("Process Failed,%v", errChild)
		return nil, errChild
	}

	return userPageInfo, nil
}

//用户列表V2
func GetUserDetailByPageV2(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	//解析包体
	var searchReq model.UserSearchDTO
	if err := ctx.BindJSON(&searchReq); err != nil {
		srvContext.Errorf("BindJson Failed,%v", err)
		return nil, errcode.CerrApiApiBadJsonPayload(err)
	}

	//修正参数
	if searchReq.PageNo <= 0 { //页码序号从1开始
		searchReq.PageNo = 1
	}
	if searchReq.Size <= 0 { //每一页的数量
		searchReq.Size = 10
	}

	//逻辑处理
	userPageInfo, errChild := model.ProcessUserDetailByPageV2(feignKey, &searchReq)
	if nil != errChild {
		srvContext.Errorf("Process Failed,%v", errChild)
		return nil, errChild
	}

	return userPageInfo, nil
}

func GetSimpleUserHasRightByPage(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	//解析包体
	var searchReq model.UserSearchDTO
	if err := ctx.BindJSON(&searchReq); err != nil {
		srvContext.Errorf("BindJson Failed,%v", err)
		return nil, errcode.CerrApiApiBadJsonPayload(err)
	}

	//修正参数
	if searchReq.PageNo <= 0 { //页码序号从1开始
		searchReq.PageNo = 1
	}
	if searchReq.Size <= 0 { //每一页的数量
		searchReq.Size = 10
	}

	//逻辑处理
	userDetailPage, errChild := model.ProcessUserDetailByPage(feignKey, &searchReq)
	if nil != errChild {
		srvContext.Errorf("Process Failed,%v", errChild)
		return nil, errChild
	} else if nil == userDetailPage {
		srvContext.Errorf("Process result abnormal")
		return nil, errcode.CerrInternal
	}

	//UserDetailPageDTO转换成SimpleUserPageDTO
	var userSimplePage model.SimpleUserPageDTO
	userSimplePage.List = make([]model.SimpleUserDTO, 0)
	userSimplePage.Size = userDetailPage.Size
	userSimplePage.Pager = userDetailPage.Pager
	userSimplePage.Pages = userDetailPage.Pages
	userSimplePage.Total = userDetailPage.Total
	userSimplePage.Offset = userDetailPage.Offset
	for _, detailInfo := range userDetailPage.List {
		var simpleInfo model.SimpleUserDTO
		simpleInfo.Id = detailInfo.Id
		simpleInfo.Name = detailInfo.Name
		simpleInfo.ParentId = detailInfo.Parent
		simpleInfo.LevelId = detailInfo.LevelId
		simpleInfo.LevelName = detailInfo.LevelName
		simpleInfo.RoleId = detailInfo.RoleId
		simpleInfo.RoleName = detailInfo.RoleName
		simpleInfo.EntityNo = detailInfo.EntityNo
		simpleInfo.PubUserId = detailInfo.PubUserId
		simpleInfo.Login = detailInfo.Login
		simpleInfo.VendorServerId = detailInfo.VendorServerId
		userSimplePage.List = append(userSimplePage.List, simpleInfo)
	}

	return &userSimplePage, nil
}

//根据用户id和角色查询用户,两结果取并集
func ListUserInfoByIdsAndRoles(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	//解析包体
	var pubRoleReq model.UserRoleDTO
	if err := ctx.BindJSON(&pubRoleReq); err != nil {
		srvContext.Errorf("BindJson Failed,%v", err)
		return nil, errcode.CerrApiApiBadJsonPayload(err)
	}

	//逻辑处理
	userInfos, errChild := model.ProcessUserInfoByIdsAndRoles(feignKey, &pubRoleReq)
	if nil != errChild {
		srvContext.Errorf("Process Failed,%v", errChild)
		return nil, errChild
	}

	return userInfos, nil
}

//查询下级用户信息
func ListChildUserIds(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	execBelong, errParseExec := strconv.ParseBool(ctx.Query("execBelong"))
	if nil != errParseExec {
		return nil, errcode.CerrParamater
	}

	currentUserInfo, errGetCurrent := model.GetUserDetailByIdOrPubid(feignKey, feignKey.UserId, 0)
	if errGetCurrent != errGetCurrent {
		srvContext.Errorf("Get current user info failed, error:%v", errGetCurrent)
		return nil, errGetCurrent
	} else if nil == currentUserInfo {
		return nil, errcode.CerrAccountNotFound
	}

	//逻辑处理
	childIds, errChild := model.ProcessListChildUserIds(feignKey, currentUserInfo.Id, execBelong)
	if nil != errChild {
		srvContext.Errorf("Process Failed,%v", errChild)
		return nil, errChild
	}

	return childIds, nil
}

//查询下级用户信息
func ListChildUserIdsById(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	execBelong, errParseExec := strconv.ParseBool(ctx.Query("execBelong"))
	userId, errParseUserId := strconv.ParseInt(ctx.Query("userId"), 10, 64)
	if nil != errParseExec || nil != errParseUserId || userId < 0 {
		return nil, errcode.CerrParamater
	}

	//逻辑处理
	childIds, errChild := model.ProcessListChildUserIds(feignKey, userId, execBelong)
	if nil != errChild {
		srvContext.Errorf("Process Failed,%v", errChild)
		return nil, errChild
	}

	return childIds, nil
}

//查询下级用户信息
func ListChildUserById(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	execBelong, errParseExec := strconv.ParseBool(ctx.Query("execBelong")) //是否过滤直属
	userId, errParseUserId := strconv.ParseInt(ctx.Query("userId"), 10, 64)
	if nil != errParseExec || nil != errParseUserId || userId < 0 {
		return nil, errcode.CerrParamater
	}

	//逻辑处理
	children, errChild := model.ProcessListChildUser(feignKey, userId, execBelong)
	if nil != errChild {
		srvContext.Errorf("Process Failed,%v", errChild)
		return nil, errChild
	}

	//userRecords 转换为userDetails
	userDetails, errCov := model.UserRecords2UserDetails(children)
	if nil != errCov {
		logrus.WithFields(logrus.Fields{"userId": userId}).Warn("Convert user info from []UserDetailTab to []UserDetail failed")
		return userDetails, errcode.CerrInternal
	}
	return userDetails, nil
}

//根据返佣层级或角色ID查询相关用户信息（层级查询支持包含上级层级用户）
func FindUserByType(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	includeParent, errParseInclude := strconv.ParseBool(ctx.Query("includeParent"))
	typeValue, errParseType := strconv.ParseInt(ctx.Query("type"), 10, 64)
	id, errParseId := strconv.ParseInt(ctx.Query("id"), 10, 64)
	if nil != errParseInclude || nil != errParseType || nil != errParseId {
		srvContext.Errorf("Parameter error")
		return nil, errcode.CerrParamater
	}

	//逻辑处理
	userRecords, errChild := model.ProcessFindUserByTypeFuzzy(feignKey, id, typeValue, includeParent, "")
	if nil != errChild {
		srvContext.Errorf("Process Failed,%v", errChild)
		return nil, errChild
	}

	var simpleUsers []model.SimpleUserDTO = make([]model.SimpleUserDTO, 0)
	for index, _ := range userRecords {
		simple, errConv := model.UserRecord2SimpleUser(&userRecords[index])
		if nil == errConv && nil != simple {
			simpleUsers = append(simpleUsers, *simple)
		}
	}

	return simpleUsers, nil
}

//根据返佣层级或角色ID模糊查询查询相关用户信息（层级查询支持包含上级层级用户）
func FindUserByTypeFuzzy(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	includeParent, errParseInclude := strconv.ParseBool(ctx.Query("includeParent"))
	typeValue, errParseType := strconv.ParseInt(ctx.Query("type"), 10, 64)
	id, errParseId := strconv.ParseInt(ctx.Query("id"), 10, 64)
	if nil != errParseInclude || nil != errParseType || nil != errParseId {
		srvContext.Errorf("Parameter error")
		return nil, errcode.CerrParamater
	}

	var fuzzyReq model.FuzzyConditionDTO
	if err := ctx.BindJSON(&fuzzyReq); err != nil {
		srvContext.Errorf("BindJson Failed,%v", err)
		return nil, errcode.CerrApiApiBadJsonPayload(err)
	}

	//逻辑处理
	userRecords, errChild := model.ProcessFindUserByTypeFuzzy(feignKey, id, typeValue, includeParent, fuzzyReq.FuzzyValue)
	if nil != errChild {
		srvContext.Errorf("Process Failed,%v", errChild)
		return nil, errChild
	}

	var simpleUsers []model.SimpleUserDTO = make([]model.SimpleUserDTO, 0)
	for index, _ := range userRecords {
		simple, errConv := model.UserRecord2SimpleUser(&userRecords[index])
		if nil == errConv && nil != simple {
			simpleUsers = append(simpleUsers, *simple)
		}
	}

	return simpleUsers, nil
}

//更新当前用户
func UpdateCurrentUser(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	//解析请求包体
	var bwUserReq model.BWUserDTO
	if err := ctx.BindJSON(&bwUserReq); err != nil {
		srvContext.Errorf("BindJson Failed,%v", err)
		return nil, errcode.CerrApiApiBadJsonPayload(err)
	}

	//格式化参数
	bwUserReq.Name = strings.TrimSpace(bwUserReq.Name)
	bwUserReq.Email = strings.TrimSpace(bwUserReq.Email)
	bwUserReq.Phones.CountryCode = strings.TrimSpace(bwUserReq.Phones.CountryCode)
	bwUserReq.Phones.Phone = strings.TrimSpace(bwUserReq.Phones.Phone)
	bwUserReq.Phone = bwUserReq.Phones.CountryCode + "@-@" + bwUserReq.Phones.Phone
	if "" == bwUserReq.Country {
		bwUserReq.Country = bwUserReq.Region.Country
	}
	if "" == bwUserReq.Province {
		bwUserReq.Province = bwUserReq.Region.Province
	}
	if "" == bwUserReq.City {
		bwUserReq.City = bwUserReq.Region.City
	}

	//逻辑处理
	userDetail, errChild := model.ProcessUpdateCurrentUser(feignKey, &bwUserReq)
	if nil != errChild {
		srvContext.Errorf("Process Failed,%v", errChild)
		return nil, errChild
	}

	return userDetail, nil
}

//全量更新其他用户信息
func UpdateUserV1(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	//解析请求包体
	var bwUserReq model.BWUserDTO
	if err := ctx.BindJSON(&bwUserReq); err != nil {
		srvContext.Errorf("BindJson Failed,%v", err)
		return nil, errcode.CerrApiApiBadJsonPayload(err)
	}

	//格式化参数
	bwUserReq.Name = strings.TrimSpace(bwUserReq.Name)
	bwUserReq.Email = strings.TrimSpace(bwUserReq.Email)
	bwUserReq.Phone = strings.TrimSpace(bwUserReq.Phone)
	bwUserReq.Phones.CountryCode = strings.TrimSpace(bwUserReq.Phones.CountryCode)
	bwUserReq.Phones.Phone = strings.TrimSpace(bwUserReq.Phones.Phone)
	if "" == bwUserReq.Phone {
		bwUserReq.Phone = bwUserReq.Phones.CountryCode + "@-@" + bwUserReq.Phones.Phone
	}
	if "" == bwUserReq.Country {
		bwUserReq.Country = bwUserReq.Region.Country
	}
	if "" == bwUserReq.Province {
		bwUserReq.Province = bwUserReq.Region.Province
	}
	if "" == bwUserReq.City {
		bwUserReq.City = bwUserReq.Region.City
	}
	if nil != bwUserReq.Login {
		if _, err := strconv.ParseInt(*bwUserReq.Login, 10, 64); nil != err {
			*bwUserReq.Login = ""
		}
	}

	if _, err := strconv.ParseInt(bwUserReq.Parent, 10, 64); nil != err {
		bwUserReq.Parent = ""
	}

	//逻辑处理
	detail, errUpdate := model.ProcessUpdateUserV1(feignKey, &bwUserReq)
	if nil != errUpdate {
		srvContext.Errorf("Process Failed,%v", errUpdate)
		return nil, errUpdate
	}

	model.SendUserMsg(feignKey.TenantId, strconv.FormatInt(bwUserReq.Id, 10), constant.BwEvent_UPDATE, detail)
	return nil, nil
}

//获取自定义搜索列下拉菜单
func GetSearchDropDown(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	searchType := ctx.Query("searchType")
	searchLevel := ctx.Query("searchLevel")

	isNew := false
	isNewStr := ctx.Query("isNew")
	if isNewStr != "" {
		isNewTemp, err := strconv.ParseBool(isNewStr)
		if err != nil {
			srvContext.Errorf("Request parameter 'isNew' abnormal")
			return nil, errcode.CerrParamater
		}
		isNew = isNewTemp
	}

	//逻辑处理
	searchResults, errChild := model.ProcessSearchDropDown(feignKey, searchType, searchLevel, isNew)
	if nil != errChild {
		srvContext.Errorf("Process Failed,%v", errChild)
		return nil, errChild
	}

	return searchResults, nil
}

//获取自定义搜索列表
func GetSearchList(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	searchType := ctx.Query("searchType")
	searchLevel := ctx.Query("searchLevel")

	isNew := false
	isNewStr := ctx.Query("isNew")
	if isNewStr != "" {
		isNewTemp, err := strconv.ParseBool(isNewStr)
		if err != nil {
			srvContext.Errorf("Request parameter 'isNew' abnormal")
			return nil, errcode.CerrParamater
		}
		isNew = isNewTemp
	}

	//逻辑处理
	childIds, errChild := model.ProcessSearchList(feignKey, searchType, searchLevel, isNew)
	if nil != errChild {
		srvContext.Errorf("Process Failed,%v", errChild)
		return nil, errChild
	}

	return childIds, nil
}

//获取自定义搜索详情
func GetOneSearchInfo(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	searchId := ctx.Query("id")
	if "" == searchId {
		srvContext.Errorf("Request parameter abnormal")
		return nil, errcode.CerrParamater
	}

	isNew := false
	isNewStr := ctx.Query("isNew")
	if isNewStr != "" {
		isNewTemp, err := strconv.ParseBool(isNewStr)
		if err != nil {
			srvContext.Errorf("Request parameter 'isNew' abnormal")
			return nil, errcode.CerrParamater
		}
		isNew = isNewTemp
	}

	//逻辑处理
	result, errProcess := model.ProcessGetOneSearchInfo(feignKey, searchId, isNew)
	if nil != errProcess {
		srvContext.Errorf("Process Failed,%v", errProcess)
		return nil, errProcess
	}

	return result, nil
}

//新增自定义搜索
func AddSearchInfo(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	//解析请求包体
	var addReq model.AdvanceSearchDTO
	if err := ctx.BindJSON(&addReq); err != nil {
		srvContext.Errorf("BindJson Failed,%v", err)
		return nil, errcode.CerrApiApiBadJsonPayload(err)
	}

	isNew := false
	isNewStr := ctx.Query("isNew")
	if isNewStr != "" {
		isNewTemp, err := strconv.ParseBool(isNewStr)
		if err != nil {
			srvContext.Errorf("Request parameter 'isNew' abnormal")
			return nil, errcode.CerrParamater
		}
		isNew = isNewTemp
	}

	//逻辑处理
	content, errAdd := model.ProcessAddSearchInfo(feignKey, &addReq, isNew)
	if nil != errAdd {
		srvContext.Errorf("Process Failed,%v", errAdd)
		return nil, errAdd
	}

	return content, nil
}

//删除自定义搜索
func DeleteOneSearchInfo(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	searchId := ctx.Param("id")
	//无请求包体

	isNew := false
	isNewStr := ctx.Query("isNew")
	if isNewStr != "" {
		isNewTemp, err := strconv.ParseBool(isNewStr)
		if err != nil {
			srvContext.Errorf("Request parameter 'isNew' abnormal")
			return nil, errcode.CerrParamater
		}
		isNew = isNewTemp
	}

	//逻辑处理
	errAdd := model.ProcessDeleteOneSearchInfo(feignKey, searchId, isNew)
	if nil != errAdd {
		srvContext.Errorf("Process Failed,%v", errAdd)
		return nil, errAdd
	}

	return nil, nil
}

//编辑（更新）自定义搜索
func EditOneSearchInfo(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	//解析请求包体
	var editReq model.AdvanceSearchDTO
	if err := ctx.BindJSON(&editReq); err != nil {
		srvContext.Errorf("BindJson Failed,%v", err)
		return nil, errcode.CerrApiApiBadJsonPayload(err)
	}

	isNew := false
	isNewStr := ctx.Query("isNew")
	if isNewStr != "" {
		isNewTemp, err := strconv.ParseBool(isNewStr)
		if err != nil {
			srvContext.Errorf("Request parameter 'isNew' abnormal")
			return nil, errcode.CerrParamater
		}
		isNew = isNewTemp
	}

	//逻辑处理
	errEdit := model.ProcessEditOneSearchInfo(feignKey, &editReq, isNew)
	if nil != errEdit {
		srvContext.Errorf("Process Failed,%v", errEdit)
		return nil, errEdit
	}

	return nil, nil
}

func MsgReceiversQuery(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	//解析请求包体
	var receiverReq model.MsgReceiversSearchDTO
	if err := ctx.BindJSON(&receiverReq); err != nil {
		srvContext.Errorf("BindJson Failed,%v", err)
		return nil, errcode.CerrApiApiBadJsonPayload(err)
	}

	//逻辑处理
	var msgList []model.MsgReceiversDTO = make([]model.MsgReceiversDTO, 0)
	if constant.ReceiverType_BWUser_User != receiverReq.ReceiverType {
		msgListTmp, errMsg := model.RoleMsgReceivers(feignKey, &receiverReq)
		if nil != errMsg {
			srvContext.Errorf("RoleMsgReceivers Failed,%v", errMsg)
			return nil, errMsg
		}
		msgList = append(msgList, msgListTmp...)
	}

	msgListTmp, errMsg := model.UserMsgReceivers(feignKey, 0, &receiverReq)
	if nil != errMsg {
		srvContext.Errorf("UserMsgReceivers Failed,%v", errMsg)
		return nil, errMsg
	}
	msgList = append(msgList, msgListTmp...)

	return msgList, nil
}

func UpdateEmail(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	//解析请求包体
	var bwUserReq model.BWUserDTO
	if err := ctx.BindJSON(&bwUserReq); err != nil {
		srvContext.Errorf("BindJson Failed,%v", err)
		return nil, errcode.CerrApiApiBadJsonPayload(err)
	}

	bwUserReq.Email = strings.TrimSpace(bwUserReq.Email)
	//逻辑处理
	email, errUpdate := model.ProcessUpdateEmail(feignKey, &bwUserReq)
	if nil != errUpdate {
		srvContext.Errorf("Process Failed,%v", errUpdate)
		return nil, errUpdate
	}
	return email, nil
}

func UpdateTwoFAConfig(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	//解析请求包体
	var bwUserReq model.BWUserDTO
	if err := ctx.BindJSON(&bwUserReq); err != nil {
		srvContext.Errorf("BindJson Failed,%v", err)
		return nil, errcode.CerrApiApiBadJsonPayload(err)
	}
	//逻辑处理
	errUpdate := model.ProcessUpdateTwoFAConfig(feignKey, &bwUserReq)
	if nil != errUpdate {
		srvContext.Errorf("Process Failed,%v", errUpdate)
		return nil, errUpdate
	}
	return nil, nil
}

//增量更新用户信息
func UpdateUserV2(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	//解析请求包体
	var bwUserReq model.BWUserDTOIncrease
	/*
		buf := make([]byte, 4096)
		n, _ := ctx.Request.Body.Read(buf)
		srvContext.Infof(string(buf[0:n]))
		if err := json.Unmarshal(buf[0:n], &bwUserReq); err != nil {
			srvContext.Errorf("BindJson Failed,%v", err)
			return nil, errcode.CerrApiApiBadJsonPayload(err)
		}*/
	if err := ctx.BindJSON(&bwUserReq); err != nil {
		srvContext.Errorf("BindJson Failed,%v", err)
		return nil, errcode.CerrApiApiBadJsonPayload(err)
	}
	//格式化参数
	if nil != bwUserReq.Name {
		*bwUserReq.Name = strings.TrimSpace(*bwUserReq.Name)
	}
	if nil != bwUserReq.Email {
		*bwUserReq.Email = strings.TrimSpace(*bwUserReq.Email)
	}
	if nil != bwUserReq.Phone {
		*bwUserReq.Phone = strings.TrimSpace(*bwUserReq.Phone)
	}
	if nil != bwUserReq.Phones {
		bwUserReq.Phones.CountryCode = strings.TrimSpace(bwUserReq.Phones.CountryCode)
		bwUserReq.Phones.Phone = strings.TrimSpace(bwUserReq.Phones.Phone)
		if strings.ToUpper(bwUserReq.Phones.CountryCode) == "NULL" {
			bwUserReq.Phones.CountryCode = ""
		}
		if strings.ToUpper(bwUserReq.Phones.Phone) == "NULL" {
			bwUserReq.Phones.Phone = ""
		}
		phone := ""
		if "" != bwUserReq.Phones.CountryCode || "" != bwUserReq.Phones.Phone {
			phone = bwUserReq.Phones.CountryCode + "@-@" + bwUserReq.Phones.Phone
		}
		bwUserReq.Phone = &phone
	}

	//逻辑处理
	userDetail, errUpdate := model.ProcessUpdateUserV2(feignKey, &bwUserReq)
	if nil != errUpdate {
		srvContext.Errorf("Process Failed,%v", errUpdate)
		return nil, errUpdate
	}
	model.SendUserMsg(feignKey.TenantId, strconv.FormatInt(bwUserReq.Id, 10), constant.BwEvent_UPDATE, userDetail)
	return userDetail, nil
}

func GetUserFieldsList(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	tableName := ctx.Query("tableName")
	//无请求包体

	formFieldList, errGetField := model.ProcessGetUserFieldsList(feignKey, tableName)
	if nil != errGetField {
		srvContext.Errorf("ProcessGetUserFieldsList Failed,%v", errGetField)
		return nil, errGetField
	}

	return formFieldList, nil
}

func UpdateUserFields(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	var updateReq model.UserFieldsDTO
	if err := ctx.BindJSON(&updateReq); err != nil {
		srvContext.Errorf("BindJson Failed,%v", err)
		return nil, errcode.CerrApiApiBadJsonPayload(err)
	}

	errUpdate := model.ProcessUpdateUserFields(feignKey, &updateReq)
	if nil != errUpdate {
		srvContext.Errorf("ProcessUpdateUserFields Failed,%v", errUpdate)
		return nil, errUpdate
	}

	return nil, nil
}

//根据条件搜索用户
func SearchUser(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	//解析包体
	var searchReq model.SearchDTO
	if err := ctx.BindJSON(&searchReq); err != nil {
		srvContext.Errorf("BindJson Failed,%v", err)
		return nil, errcode.CerrApiApiBadJsonPayload(err)
	}

	//逻辑处理
	userRecords, errChild := model.SearchUserByCondition(feignKey, &searchReq)
	if nil != errChild {
		srvContext.Errorf("Process Failed,%v", errChild)
		return nil, errChild
	}

	response := make([]*model.UserDetail, 0)
	for index := range userRecords {
		detail, errConv := model.UserRecord2UserDetail(userRecords[index])
		if nil != errConv || nil == detail {
			logrus.WithFields(logrus.Fields{"error": errConv}).Warn("UserRecord2SimpleUser abnormal")
			continue
		}
		response = append(response, detail)
	}
	return response, nil
}

// 按角色统计用户数
func UserStatByRoleId(srvContext context.Context, ctx *gin.Context) (interface{}, code.Error) {
	tenantId := ctx.Query("tenantId")
	return model.UserCountGroupByRole(tenantId)
}

//根据条件搜索用户
func SearchSimpleUser(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	//解析包体
	var searchReq model.SearchDTO
	if err := ctx.BindJSON(&searchReq); err != nil {
		srvContext.Errorf("BindJson Failed,%v", err)
		logrus.Error("BindJson Failed,%v", err)
		return nil, errcode.CerrApiApiBadJsonPayload(err)
	}

	//逻辑处理
	userRecords, errChild := model.SearchUserByCondition(feignKey, &searchReq)
	if nil != errChild {
		srvContext.Errorf("Process Failed,%v", errChild)
		return nil, errChild
	}

	response := make([]*model.SimpleUserDTO, 0)
	for index := range userRecords {
		detail, errConv := model.UserRecord2SimpleUser(userRecords[index])
		if nil != errConv || nil == detail {
			logrus.WithFields(logrus.Fields{"error": errConv}).Warn("UserRecord2SimpleUser abnormal")
			continue
		}
		response = append(response, detail)
	}
	return response, nil
}

//TODO 更新用户积分
func UpdateUserPoints(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	//解析请求包体
	var bwUserReq model.BWUserDTO
	if err := ctx.BindJSON(&bwUserReq); err != nil {
		srvContext.Errorf("BindJson Failed,%v", err)
		return false, errcode.CerrApiApiBadJsonPayload(err)
	}

	errProcess := model.ProcessUpdateUserPoints(feignKey, &bwUserReq)
	if errProcess != nil {
		return false, errProcess
	}
	return true, nil
}

//获取用户积分
func GetUserPoints(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	idStr := ctx.Query("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error("获取用户积分GetUserPoints失败")
	}
	userDetail, errProcess := model.ProcessGetUserPoints(feignKey, id)
	return userDetail, errProcess
}

//根据用户email获取用户id
func GetUserIDByEmail(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	email := ctx.Query("email")
	id, errProcess := model.ProcessGetUserIDByEmail(feignKey, email)
	if errProcess != nil {
		logrus.Error("获取用户id GetUserIDByEmail 失败")
		return model.UserIds{}, errProcess
	}
	var userIds *model.UserIds = &model.UserIds{UserId: id, Email: email}
	return userIds, nil
}
