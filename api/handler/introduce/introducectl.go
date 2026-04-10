package introducectl

import (
	"bw/bw-user/api/handler/util"
	"bw/bw-user/constant"
	"bw/bw-user/errcode"
	"strings"

	"bw/bw-user/model"
	"github.com/gin-gonic/gin"
	"github.com/lworkltd/kits/service/context"
	"github.com/lworkltd/kits/service/restful/code"

	"net/url"
	//"encoding/json"
	//"net/url"
	//"github.com/Sirupsen/logrus"
	//"strconv"
	//"time"
	"strconv"
	"time"
)

//格式化SystemIntroduceDTO结构的请求包体enum字段
func formatSystemIntroduceDTO(introduce *model.SystemIntroduceDTO) {
	if nil == introduce {
		return
	}
	introduce.Platform = constant.FormatIntroduceType(introduce.Platform)
	introduce.Type = constant.FormatIntroduceType(introduce.Type)
	introduce.BwUserShow = constant.FormatIntroduceType(introduce.BwUserShow)
	introduce.ParameterType = constant.FormatIntroduceType(introduce.ParameterType)
	introduce.Vendor = constant.FormatVendor(introduce.Vendor)
}

//添加introduce
func AddIntroduce(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	//解析请求包体
	var addIntroduceReq model.SystemIntroduceDTO
	if err := ctx.BindJSON(&addIntroduceReq); err != nil {
		srvContext.Errorf("BindJson Failed,%v", err)
		return nil, errcode.CerrApiApiBadJsonPayload(err)
	}
	formatSystemIntroduceDTO(&addIntroduceReq)

	//逻辑处理
	detail, errChild := model.ProcessAddIntroduce(feignKey, &addIntroduceReq)
	if nil != errChild {
		srvContext.Errorf("Process Failed,%v", errChild)
		return nil, errChild
	}

	return detail, nil
}

func IntroduceListSimple(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	typeValue := ctx.Query("type") //IntroduceType，非必填参数
	var types []string
	if typeValue != "" {
		types = strings.Split(typeValue, ",")
	}

	platform := ctx.Query("platform") //platform，非必填参数

	//无请求包体

	//逻辑处理
	details, errChild := model.FindAllSystemIntroduceSimple(feignKey, platform, types, nil)
	if nil != errChild {
		srvContext.Errorf("Process Failed,%v", errChild)
		return nil, errChild
	}

	return details, nil
}

func IntroduceList(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	typeValue := ctx.Query("type") //IntroduceType，非必填参数
	var types []string
	if typeValue != "" {
		types = strings.Split(typeValue, ",")
	}

	platform := ctx.Query("platform")

	enableStr := ctx.Query("enable") //非必填参数
	var enable *bool = nil
	if "" != enableStr {
		value, err := strconv.ParseBool(enableStr)
		if nil == err {
			enable = new(bool)
			*enable = value
		}
	}
	//无请求包体

	//逻辑处理
	details, errChild := model.FindAllSystemIntroduce(feignKey, platform, types, enable, nil)
	if nil != errChild {
		srvContext.Errorf("Process Failed,%v", errChild)
		return nil, errChild
	}

	return details, nil
}

//根据key和value列表，返回用户详情信息
func GetIntroduceInfosByKey(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	keyValue, errKey := url.QueryUnescape(ctx.Param("keyValue"))
	keyType := ctx.Param("keyType")
	if nil != errKey || "" == keyValue || "" == keyType {
		srvContext.Errorf("GetIntroduceInfosByKey parameter error, keyValue:%v, keyType:%v", ctx.Param("keyValue"), keyType)
		return nil, errcode.CerrParamater
	}

	//keyValue格式：value1,value2,value3
	var intValueList []int64
	var strValueList []string
	var errParseValue code.Error
	var records []model.SystemIntroduceTab
	var errProcess code.Error

	if constant.QueryKeyId == keyType {
		intValueList, errParseValue = model.SplitToIntList(keyValue, ",") //解析keyValue参数
		if nil != errParseValue || len(intValueList) == 0 {
			srvContext.Errorf("GetIntroduceInfosByKey parameter error, keyValue:%v, keyType:%v", ctx.Param("keyValue"), keyType)
			return nil, errcode.CerrParamater
		}
		//获取记录详情
		records, errProcess = model.ListIntroduceRecordsByKey(feignKey, intValueList, keyType)
	} else if constant.QueryKeyEntityNo == keyType || constant.QueryKeyName == keyType {
		strValueList, errParseValue = model.SplitToStringList(keyValue, ",") //解析keyValue参数
		if nil != errParseValue || len(strValueList) == 0 {
			srvContext.Errorf("GetIntroduceInfosByKey parameter error, keyValue:%v, keyType:%v", ctx.Param("keyValue"), keyType)
			return nil, errcode.CerrParamater
		}
		//获取记录详情
		records, errProcess = model.ListIntroduceRecordsByKey(feignKey, strValueList, keyType)
	} else if constant.QueryKeyAllRecords == keyType {
		//获取记录详情
		records, errProcess = model.ListIntroduceRecordsByKey(feignKey, nil, keyType)
	} else {
		srvContext.Errorf("GetIntroduceInfosByKey keytype not support, keyType:%v", keyType)
		return nil, errcode.CerrQueryKeyNotSupport
	}

	if nil != errProcess {
		srvContext.Errorf("ProcessListUserInfosByKey Failed,%v", errProcess)
		return nil, errProcess
	}

	var detailList []model.SystemIntroduceDTO = make([]model.SystemIntroduceDTO, 0)
	for index, _ := range records {
		detail := model.SystemIntroduceTab2SystemIntroduceDTO(&records[index])
		if nil != detail {
			detailList = append(detailList, *detail)
		}
	}

	return detailList, nil
}

//查询截止到今天凌晨0点的Introduce数
func GetIntroduceCountYesterday(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	todayBeginTimeStr := time.Now().Format("2006-01-02")
	todayBeginTime, errParse := time.Parse("2006-01-02", todayBeginTimeStr)
	if nil != errParse {
		srvContext.Errorf("Process Failed,%v", errParse)
		return nil, errcode.CerrInternal
	}
	//逻辑处理
	counter, errChild := model.GetIntroduceCountBeforeOneCreateTime(feignKey.TenantId, todayBeginTime)
	if nil != errChild {
		srvContext.Errorf("Process Failed,%v", errChild)
		return nil, errChild
	}

	return counter, nil
}

//生成默认代理推广链接
func AddDefaultAgentIntroduce(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	return nil, nil
	//language := ctx.Query("lang")			//默认值为zh-CN
	//if "" == language {
	//	language = "zh-CN"
	//}
	////解析请求包体
	//var addIntroduceReq model.SystemIntroduceDTO
	//if err := ctx.BindJSON(&addIntroduceReq); err != nil {
	//	srvContext.Errorf("BindJson Failed,%v", err)
	//	return nil, errcode.CerrApiApiBadJsonPayload(err)
	//}
	//formatSystemIntroduceDTO(&addIntroduceReq)
	//
	////逻辑处理
	//errAdd := model.ProcessAddDefaultAgentIntroduce(feignKey, &addIntroduceReq, language)
	//if nil != errAdd {
	//	srvContext.Errorf("Process Failed,%v", errAdd)
	//	return nil, errAdd
	//}
	//
	//return errAdd, nil
}

//生成默认TW推广链接
func AddDefaultTwIntroduce(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	/*language := ctx.Query("lang")			//默认值为zh-CN
	if "" == language {
		language = "zh-CN"
	}
	//解析请求包体
	var addIntroduceReq model.SystemIntroduceDTO
	if err := ctx.BindJSON(&addIntroduceReq); err != nil {
		srvContext.Errorf("BindJson Failed,%v", err)
		return nil, errcode.CerrApiApiBadJsonPayload(err)
	}
	formatSystemIntroduceDTO(&addIntroduceReq)

	//逻辑处理
	errAdd := model.ProcessAddDefaultTwIntroduce(feignKey, &addIntroduceReq, language)
	if nil != errAdd {
		srvContext.Errorf("Process Failed,%v", errAdd)
		return nil, errAdd
	}

	return errAdd, nil*/
	return nil, nil
}

//更新默认TW推广链接
func UpdateDefaultTwIntroduce(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	/*language := ctx.Query("lang")			//默认值为zh-CN
	if "" == language {
		language = "zh-CN"
	}
	//解析请求包体
	var addIntroduceReq model.SystemIntroduceDTO
	if err := ctx.BindJSON(&addIntroduceReq); err != nil {
		srvContext.Errorf("BindJson Failed,%v", err)
		return nil, errcode.CerrApiApiBadJsonPayload(err)
	}
	formatSystemIntroduceDTO(&addIntroduceReq)

	//逻辑处理
	errAdd := model.ProcessUpdateDefaultTwIntroduce(feignKey, &addIntroduceReq, language)
	if nil != errAdd {
		srvContext.Errorf("Process Failed,%v", errAdd)
		return nil, errAdd
	}

	return errAdd, nil*/
	return nil, nil
}

//添加推广链接点击信息
func AddIntroduceHit(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	//解析请求包体
	var introduceHit model.IntroduceHitDTO
	if err := ctx.BindJSON(&introduceHit); err != nil {
		srvContext.Errorf("BindJson Failed,%v", err)
		return nil, errcode.CerrApiApiBadJsonPayload(err)
	}

	//逻辑处理
	errAdd := model.ProcessAddIntroduceHit(feignKey, &introduceHit)
	if nil != errAdd {
		srvContext.Errorf("Process Failed,%v", errAdd)
		return nil, errAdd
	}

	return errAdd, nil
}

//BW我的推广链接
func GetMyIntroduces(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	//获取当前用户信息
	currentUser, errGetUser := model.GetUserDetailByIdOrPubid(feignKey, feignKey.UserId, 0)
	if nil != errGetUser {
		srvContext.Errorf("Process Failed,%v", errGetUser)
		return nil, errGetUser
	} else if nil == currentUser {
		srvContext.Errorf("Process Failed, user not found")
		return nil, errcode.CerrAccountNotFound
	}
	//获取推广信息
	enable := true
	details, errChild := model.FindAllSystemIntroduce(feignKey, constant.IntroduceType_Web,
		[]string{constant.IntroduceType_StraightGuest, constant.IntroduceType_Agent},
		&enable, map[string]string{"uid": strconv.FormatInt(currentUser.Id, 10)})
	if nil != errChild {
		srvContext.Errorf("Process Failed,%v", errChild)
		return nil, errChild
	}

	strId := strconv.FormatInt(currentUser.Id, 10)
	strLevelId := strconv.FormatInt(currentUser.LevelId, 10)
	strRoleId := strconv.FormatInt(currentUser.RoleId, 10)
	myIntroduce := make([]model.SystemIntroduceDTO, 0)
	for index, _ := range details {
		if constant.IntroduceType_UserAllVisible == details[index].BwUserShow {
			myIntroduce = append(myIntroduce, details[index])
			continue
		}

		if constant.IntroduceType_UserPartVisible == details[index].BwUserShow {
			if model.StringListIsContain(details[index].VisibleUser, strId+"-"+constant.IdType_Id) {
				myIntroduce = append(myIntroduce, details[index])
			} else if model.StringListIsContain(details[index].VisibleUser, strLevelId+"-"+constant.IdType_LevelId) {
				myIntroduce = append(myIntroduce, details[index])
			} else if model.StringListIsContain(details[index].VisibleUser, strRoleId+"-"+constant.IdType_RoleId) {
				myIntroduce = append(myIntroduce, details[index])
			}
			continue
		}

		if constant.IntroduceType_UserInVisible == details[index].BwUserShow {
			if !model.StringListIsContain(details[index].InVisibleUser, strId+"-"+constant.IdType_Id) &&
				!model.StringListIsContain(details[index].InVisibleUser, strLevelId+"-"+constant.IdType_LevelId) &&
				!model.StringListIsContain(details[index].InVisibleUser, strRoleId+"-"+constant.IdType_RoleId) {
				myIntroduce = append(myIntroduce, details[index])
			}
			continue
		}
	}

	return myIntroduce, nil
}

//TW直客推广链接
func GetTWDirectIntroduces(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	//获取客户ID
	customerId := ctx.Query("customerId")
	if customerId == "" {
		srvContext.Errorf("customerId is null")
		return nil, code.New(errcode.ParameterError, "customerId is null")
	}

	//获取推广信息
	enable := true
	details, errChild := model.FindAllSystemIntroduce(feignKey, constant.IntroduceType_Web, []string{constant.IntroduceType_DirectRecommendation},
		&enable, map[string]string{"cid": customerId})
	if nil != errChild {
		srvContext.Errorf("Process Failed,%v", errChild)
		return nil, errChild
	}

	myIntroduce := make([]model.SystemIntroduceDTO, 0)
	for index := range details {
		switch details[index].BwUserShow {
		case constant.IntroduceType_DirectNotVisible:
			continue
		case constant.IntroduceType_DirectAllVisible:
			myIntroduce = append(myIntroduce, details[index])
			break
		case constant.IntroduceType_DirectPartVisible:
			if model.StringListIsContain(details[index].VisibleUser, customerId) {
				myIntroduce = append(myIntroduce, details[index])
			}
			break
		case constant.IntroduceType_DirectPartInvisible:
			if !model.StringListIsContain(details[index].InVisibleUser, customerId) {
				myIntroduce = append(myIntroduce, details[index])
			}
			break
		}
	}

	return myIntroduce, nil
}

//TW代理推广链接
func GetTWAgentIntroduces(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	//获取客户ID
	userId := ctx.Query("bwUserId")
	if userId == "" {
		srvContext.Errorf("bwUserId is null")
		return nil, code.New(errcode.ParameterError, "bwUserId is null")
	}

	//获取推广信息
	enable := true
	details, errChild := model.FindAllSystemIntroduce(feignKey, constant.IntroduceType_Web,
		[]string{constant.IntroduceType_StraightGuest, constant.IntroduceType_Agent},
		&enable, map[string]string{"uid": userId})
	if nil != errChild {
		srvContext.Errorf("Process Failed,%v", errChild)
		return nil, errChild
	}

	myIntroduce := make([]model.SystemIntroduceDTO, 0)
	for index, _ := range details {
		if (constant.IntroduceType_UserPartVisible == details[index].BwUserShow && model.StringListIsContain(details[index].VisibleUser, userId)) ||
			constant.IntroduceType_UserAllVisible == details[index].BwUserShow ||
			constant.IntroduceType_UserInVisible == details[index].BwUserShow && !model.StringListIsContain(details[index].InVisibleUser, userId) {
			myIntroduce = append(myIntroduce, details[index])
		}
	}

	return myIntroduce, nil
}

func GetIntroducesQrcode(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	introduceId, errId := strconv.ParseInt(ctx.Query("introduceId"), 10, 64)
	if nil != errId || introduceId < 0 {
		srvContext.Errorf("Process Failed,%v", errId)
		return nil, errcode.CerrParamater
	}
	isCurrentUserUrl, errIsCurr := strconv.ParseBool(ctx.Query("isCurrentUserUrl")) //可选参数，默认为true
	if nil != errIsCurr {
		isCurrentUserUrl = true
	}

	//获取推广信息
	url, errProcess := model.ProcessIntroducesQrcode(feignKey, introduceId, isCurrentUserUrl)
	if nil != errProcess {
		srvContext.Errorf("Process Failed,%v", errProcess)
		return nil, errProcess
	}

	return url, nil
}

func GetTwIntroducesQrcode(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	introduceId, errId := strconv.ParseInt(ctx.Query("introduceId"), 10, 64)
	if nil != errId || introduceId < 0 {
		srvContext.Errorf("Process Failed,%v", errId)
		return nil, errcode.CerrParamater
	}
	customerId := ctx.Query("customerId")
	bwUserId := ctx.Query("bwUserId")

	//获取推广信息
	url, errProcess := model.ProcessTwIntroducesQrcode(feignKey, introduceId, customerId, bwUserId)
	if nil != errProcess {
		srvContext.Errorf("Process Failed,%v", errProcess)
		return nil, errProcess
	}

	return url, nil
}

func GetIntroducesDetail(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	introduceId, errParseId := strconv.ParseInt(ctx.Query("introduceId"), 10, 64) //introduceId
	statistic, errParseStatic := strconv.ParseBool(ctx.Query("statistic"))
	if nil != errParseId || introduceId < 0 {
		srvContext.Errorf("Parameter ID error,%v", ctx.Query("id"))
		return nil, errcode.CerrParamater
	}
	if nil != errParseStatic {
		statistic = true
	}

	detail, errProcess := model.ProcessIntroducesDetail(feignKey, introduceId, statistic)
	if nil != errProcess {
		srvContext.Errorf("Process failed,%v", errProcess)
		return nil, errProcess
	}

	return detail, nil
}

//切换推广链接启用状态
func SwitchIntroduceState(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	introduceId, errParseId := strconv.ParseInt(ctx.Query("introduceId"), 10, 64) //introduceId
	if nil != errParseId || introduceId < 0 {
		srvContext.Errorf("Parameter ID error,%v", ctx.Query("id"))
		return nil, errcode.CerrParamater
	}

	errProcess := model.ProcessSwitchIntroduceState(feignKey, introduceId)
	if nil != errProcess {
		srvContext.Errorf("Process failed,%v", errProcess)
		return nil, errProcess
	}

	return nil, nil
}

//批量删除推广链接
func DeleteIntroduce(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	//解析请求包体
	var idStrList []string = make([]string, 0)
	if err := ctx.BindJSON(&idStrList); err != nil {
		srvContext.Errorf("BindJson Failed,%v", err)
		return nil, errcode.CerrApiApiBadJsonPayload(err)
	}

	idsList := make([]int64, 0)
	for _, idStr := range idStrList {
		idInt, errParse := strconv.ParseInt(idStr, 10, 64)
		if nil != errParse {
			srvContext.Errorf("parse introduce id Failed, error:%v, id:%v", errParse, idStr)
			return nil, errcode.CerrParamater
		}
		idsList = append(idsList, idInt)
	}

	errProcess := model.ProcessDeleteIntroduce(feignKey, idsList)
	if nil != errProcess {
		srvContext.Errorf("Process failed,%v", errProcess)
		return nil, errProcess
	}

	return nil, nil
}

//更新推广链接
func UpdateSystemIntroduce(srvContext context.Context, args util.ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error) {
	//解析请求包体
	var updateReq model.SystemIntroduceDTO
	if err := ctx.BindJSON(&updateReq); err != nil {
		srvContext.Errorf("BindJson Failed,%v", err)
		return nil, errcode.CerrApiApiBadJsonPayload(err)
	}
	formatSystemIntroduceDTO(&updateReq)

	errProcess := model.ProcessUpdateSystemIntroduce(feignKey, &updateReq)
	if nil != errProcess {
		srvContext.Errorf("Process failed,%v", errProcess)
		return nil, errProcess
	}

	return nil, nil
}
