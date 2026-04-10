package model

import (
	"github.com/lworkltd/kits/service/restful/code"
	"github.com/Sirupsen/logrus"
	"bw/bw-user/errcode"
	"bw/bw-user/constant"
	//"github.com/jinzhu/gorm"
	"gopkg.in/mgo.v2/bson"
	"strconv"
	"time"
	"gopkg.in/mgo.v2"
)

func findAdvanceSearch(tenantId, userId, searchType, searchLevel string, isNew bool) ([]AdvanceSearch, code.Error) {
	mgoSess, errSess := GetMongoConnByTenantID(tenantId)
	if nil != errSess {
		logrus.WithFields(logrus.Fields{"tenant_id": tenantId, "error": errSess}).Error("GetMongoConnByTenantID failed")
		return nil, errSess
	}
	defer mgoSess.Close()
	coll := GetMgoSearchColl(mgoSess, isNew)

	var records []AdvanceSearch = make([]AdvanceSearch, 0)
	condition := bson.M {"tenantId":tenantId,"searchType": searchType,"searchLevel":searchLevel,"enabled":true}
	if "" != userId {
		condition = bson.M {"tenantId":tenantId,"searchType": searchType,"searchLevel":searchLevel,"enabled":true, "createUserId":userId}
	}

	errGet := coll.Find(condition).All(&records)
	if nil != errGet {
		logrus.WithFields(logrus.Fields{"tenant_id": tenantId, "error": errGet}).Error("Get advance search record failed")
		return nil, errcode.CerrOperateMongo
	}
	return records, nil
}


//获取自定义搜索列下拉菜单
func ProcessSearchDropDown(feignKey *XFeignKey, searchType string, searchLevel string, isNew bool) ([]AdvanceSearchDTO, code.Error) {
	if nil == feignKey || "" == feignKey.TenantId {
		return nil, errcode.CerrParamater
	}

	var searchDTOS []AdvanceSearchDTO = make([]AdvanceSearchDTO, 0)
	records, errFind := findAdvanceSearch(feignKey.TenantId, feignKey.UserId, searchType, constant.SearchLevel_USER, isNew)		//获取用户自定义搜索模板
	if nil != errFind {
		logrus.WithFields(logrus.Fields{"tenant_id": feignKey.TenantId, "searchType": searchType,"error": errFind}).Error("findAdvanceSearch by user failed")
		return nil, errFind
	}
	for index, _ := range records {
		var searchDto AdvanceSearchDTO
		searchDto.Name = records[index].Name
		searchDto.SearchId = records[index].SearchId
		searchDto.SearchLevel = records[index].SearchLevel
		searchDto.SearchType = records[index].SearchType
		searchDTOS = append(searchDTOS, searchDto)
	}

	curentUser, errGetUser := GetUserDetailByIdOrPubid(feignKey, feignKey.UserId, 0)
	if errGetUser != nil {
		logrus.WithFields(logrus.Fields{"pubUserId": feignKey.UserId,"error": errGetUser}).Error("GetUserDetailByIdOrPubid failed")
		return nil, errGetUser
	} else if nil == curentUser {
		return nil, errcode.CerrAccountNotFound
	}
	roleIdStr := strconv.FormatInt(curentUser.RoleId, 10)

	//获取系统设置搜索模板
	tenantSearch, errSearch := findAdvanceSearch(feignKey.TenantId, "", searchType, constant.SearchLevel_TENANT, isNew)
	if nil != errSearch {
		logrus.WithFields(logrus.Fields{"tenant_id": feignKey.TenantId, "searchType": searchType,"error": errFind}).Error("findAdvanceSearch by tenant failed")
		return nil, errcode.CerrOperateMongo
	}
	for index, _ := range tenantSearch {
		if len(tenantSearch[index].RoleIds) > 0 && false == StringListIsContain(tenantSearch[index].RoleIds, roleIdStr) {
			continue
		}
		var searchDto AdvanceSearchDTO
		searchDto.Name = tenantSearch[index].Name
		searchDto.SearchId = tenantSearch[index].SearchId
		searchDto.SearchLevel = tenantSearch[index].SearchLevel
		searchDto.SearchType = tenantSearch[index].SearchType
		searchDTOS = append(searchDTOS, searchDto)
	}

	return searchDTOS, nil
}



//获取自定义搜索列表
func ProcessSearchList(feignKey *XFeignKey, searchType string, searchLevel string, isNew bool) ([]AdvanceSearchDTO, code.Error) {
	if nil == feignKey || "" == feignKey.TenantId {
		return nil, errcode.CerrParamater
	}

	var returnResult []AdvanceSearchDTO =make([]AdvanceSearchDTO, 0)
	var findResult []AdvanceSearch
	var errSearch code.Error = nil
	switch searchLevel {
	case constant.SearchLevel_USER:
		findResult, errSearch = findAdvanceSearch(feignKey.TenantId, feignKey.UserId, searchType, searchLevel, isNew)
	case constant.SearchLevel_TENANT:
		findResult, errSearch = findAdvanceSearch(feignKey.TenantId, "", searchType, searchLevel, isNew)
	default:
		return nil, errcode.CerrParamater
	}
	if nil != errSearch {
		logrus.WithFields(logrus.Fields{"tenant_id": feignKey.TenantId, "searchType": searchType,"error": errSearch}).Error("findAdvanceSearch by user failed")
		return nil, errSearch
	} else if len(findResult) == 0 {
		return returnResult, nil
	}

	var hasRoleInfo bool = false
	for index, _ := range findResult {
		var searchDto AdvanceSearchDTO
		searchDto.Name = findResult[index].Name
		searchDto.Condition = findResult[index].Condition
		searchDto.LogicType = findResult[index].LogicType
		searchDto.SearchId = findResult[index].SearchId
		searchDto.SearchLevel = findResult[index].SearchLevel
		searchDto.SearchType = findResult[index].SearchType
		searchDto.RoleIds =  findResult[index].RoleIds
		if len(searchDto.RoleIds) > 0 {
			hasRoleInfo = true			//记录存在roleID信息，后面需要查找roleName
		}
		returnResult = append(returnResult, searchDto)
	}
	if false == hasRoleInfo {
		return returnResult, nil
	}

	//获取全量role信息
	allRoles, errGetRole := GetAllRoleLis(feignKey.TenantId)
	if nil != errGetRole {
		logrus.WithFields(logrus.Fields{"tenant_id": feignKey.TenantId,"error": errGetRole}).Error("GetAllRoleLis failed")
		return nil, errGetRole
	}
	roleMap := RoleTabList2RoleTabMap(allRoles)

	//填充roleName
	for index, _ := range returnResult {
		if len(returnResult[index].RoleIds) == 0 {
			continue
		}
		returnResult[index].RoleNames = make([]string, 0)
		for _, roleIdStr := range returnResult[index].RoleIds {
			var roleName string = ""
			roleId, errParse := strconv.ParseInt(roleIdStr, 10, 64)
			if nil == errParse {
				if roleInfo, ok := roleMap[roleId]; ok {
					roleName = roleInfo.Name
				}
			}
			//每个roleId填充一个Name，即使roleId非法，也要添加一个空串
			returnResult[index].RoleNames = append(returnResult[index].RoleNames, roleName)
		}
	}
	return returnResult, nil
}



//新增自定义搜索
func ProcessAddSearchInfo(feignKey *XFeignKey, addReq *AdvanceSearchDTO, isNew bool) (*AdvanceSearch, code.Error) {
	if nil == feignKey || "" == feignKey.TenantId || nil == addReq {
		return nil, errcode.CerrParamater
	}
	if "" == addReq.Name || "" == addReq.SearchType || "" == addReq.SearchLevel {
		logrus.WithFields(logrus.Fields{"tenant_id": feignKey.TenantId}).Error("ProcessAddSearchInfo parameter name/searchType/searchLevel abnormal")
		return nil, errcode.CerrParamater
	}

	var addsearch AdvanceSearch
	addsearch.SearchId = GenerateNewUUID()
	addsearch.Name = addReq.Name
	addsearch.TenantId = feignKey.TenantId
	addsearch.LogicType = addReq.LogicType
	addsearch.RoleIds = addReq.RoleIds
	addsearch.Condition = addReq.Condition
	addsearch.SearchType = addReq.SearchType
	addsearch.SearchLevel = addReq.SearchLevel
	addsearch.Enabled = true
	addsearch.CreateUserId = feignKey.UserId
	addsearch.ModifyUserId = feignKey.UserId
	addsearch.CreateTime = time.Now().Unix() * 1000
	addsearch.ModifyTime = addsearch.CreateTime


	mgoSess, errSess := GetMongoConnByTenantID(feignKey.TenantId)
	if nil != errSess {
		logrus.WithFields(logrus.Fields{"tenant_id": feignKey.TenantId, "error": errSess}).Error("GetMongoConnByTenantID failed")
		return nil, errSess
	}
	defer mgoSess.Close()

	if errInsert := GetMgoSearchColl(mgoSess, isNew).Insert(&addsearch); nil != errInsert {
		logrus.WithFields(logrus.Fields{"error": errInsert}).Error("Insert search info into mongo failed")
		return nil, errcode.CerrOperateMongo
	}

	return &addsearch, nil
}


//获取自定义搜索详情
func ProcessGetOneSearchRecord(feignKey *XFeignKey, searchId string, isNew bool) (*AdvanceSearch, code.Error) {
	if nil == feignKey || "" == feignKey.TenantId || "" == searchId {
		return nil, errcode.CerrParamater
	}

	mgoSess, errSess := GetMongoConnByTenantID(feignKey.TenantId)
	if nil != errSess {
		logrus.WithFields(logrus.Fields{"tenant_id": feignKey.TenantId, "error": errSess}).Error("GetMongoConnByTenantID failed")
		return nil, errSess
	}
	defer mgoSess.Close()
	coll := GetMgoSearchColl(mgoSess, isNew)

	var record AdvanceSearch
	errGet := coll.Find(bson.M{"tenantId":feignKey.TenantId, "_id": searchId, "enabled": true}).One(&record)
	if nil != errGet {
		if mgo.ErrNotFound == errGet {						//没找到记录
			logrus.WithFields(logrus.Fields{"tenant_id": feignKey.TenantId}).Warn("Not found advance search record")
			return nil, nil									//没找到记录不返回错误
		}
		logrus.WithFields(logrus.Fields{"tenant_id": feignKey.TenantId, "error": errGet}).Error("Get advance search record failed")
		return nil, errcode.CerrOperateMongo
	}

	return &record, nil
}

//获取自定义搜索详情
func ProcessGetOneSearchInfo(feignKey *XFeignKey, searchId string, isNew bool) (*AdvanceSearchDTO, code.Error) {
	record, errGet := ProcessGetOneSearchRecord(feignKey, searchId, isNew)
	if nil != errGet {
		logrus.WithFields(logrus.Fields{"tenant_id": feignKey.TenantId, "searchId":searchId, "error":errGet}).Error("ProcessGetOneSearchRecord failed")
		return nil, errGet
	} else if nil == record {
		logrus.WithFields(logrus.Fields{"tenant_id": feignKey.TenantId, "searchId":searchId}).Error(" searchId not exist")
		return nil, nil						//未找到不返回错误
	}

	var searchDto AdvanceSearchDTO
	searchDto.Name = record.Name
	searchDto.Condition = record.Condition
	searchDto.LogicType = record.LogicType
	searchDto.SearchId = record.SearchId
	searchDto.SearchLevel = record.SearchLevel
	searchDto.SearchType = record.SearchType
	searchDto.RoleIds =  record.RoleIds

	return &searchDto, nil
}


//删除自定义搜索
func ProcessDeleteOneSearchInfo(feignKey *XFeignKey, searchId string, isNew bool) (code.Error) {
	if nil == feignKey || "" == feignKey.TenantId || "" == searchId {
		return errcode.CerrParamater
	}

	mgoSess, errSess := GetMongoConnByTenantID(feignKey.TenantId)
	if nil != errSess {
		logrus.WithFields(logrus.Fields{"tenant_id": feignKey.TenantId, "error": errSess}).Error("GetMongoConnByTenantID failed")
		return errSess
	}
	defer mgoSess.Close()
	coll := GetMgoSearchColl(mgoSess, isNew)

	errRemove := coll.Remove(bson.M{"tenantId":feignKey.TenantId, "_id": searchId, "enabled": true})
	if nil != errRemove && mgo.ErrNotFound != errRemove {
		logrus.WithFields(logrus.Fields{"tenant_id": feignKey.TenantId, "error": errRemove}).Error("Delete one advance search record failed")
		return errcode.CerrOperateMongo
	}

	return nil
}


//编辑（更新）自定义搜索
func ProcessEditOneSearchInfo(feignKey *XFeignKey, editReq *AdvanceSearchDTO, isNew bool) (code.Error) {
	if nil == feignKey || "" == feignKey.TenantId || nil == editReq {
		return errcode.CerrParamater
	}
	if "" == editReq.Name || "" == editReq.SearchType || "" == editReq.SearchLevel || "" == editReq.SearchId{
		logrus.WithFields(logrus.Fields{"tenant_id": feignKey.TenantId}).Error("ProcessAddSearchInfo parameter name/searchType/searchLevel/searchId abnormal")
		return errcode.CerrParamater
	}

	record, errGet := ProcessGetOneSearchRecord(feignKey, editReq.SearchId, isNew)
	if nil != errGet {
		logrus.WithFields(logrus.Fields{"tenant_id": feignKey.TenantId, "searchId":editReq.SearchId, "error":errGet}).Error("ProcessGetOneSearchRecord failed")
		return errGet
	} else if nil == record {
		logrus.WithFields(logrus.Fields{"tenant_id": feignKey.TenantId, "searchId":editReq.SearchId}).Error(" searchId not exist")
		return errcode.CerrAdvanceSearchNotExist
	}

	mgoSess, errSess := GetMongoConnByTenantID(feignKey.TenantId)
	if nil != errSess {
		logrus.WithFields(logrus.Fields{"tenant_id": feignKey.TenantId, "error": errSess}).Error("GetMongoConnByTenantID failed")
		return errSess
	}
	defer mgoSess.Close()

	updateContan := bson.M{"$set":bson.M{"name":editReq.Name, "logicType":editReq.LogicType, "condition": editReq.Condition, "roleIds":editReq.RoleIds, "enabled":true, "modifyUserId":feignKey.UserId, "modifyTime": time.Now().Unix()*1000}}
	errUpdate := GetMgoSearchColl(mgoSess, isNew).Update(bson.M{"_id":record.SearchId}, updateContan)
	if nil != errUpdate {
		logrus.WithFields(logrus.Fields{"error": errUpdate}).Error("Update search info into mongo failed")
		return errcode.CerrOperateMongo
	}

	return nil
}



