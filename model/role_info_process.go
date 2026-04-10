package model

import (
	"bw/bw-user/constant"
	"bw/bw-user/errcode"
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/jinzhu/gorm"
	"github.com/lworkltd/kits/service/restful/code"
)

const ROLE_TYPE_COLLECTION = "t_role_type"

func ExistInRoleTab(tenant_id string, key string, value interface{}) (bool, code.Error) {
	if "" == tenant_id || "" == key {
		return false, errcode.CerrParamater
	}
	dbConn, errConn := GetDBConnByTenantID(tenant_id)
	if errConn != nil {
		return false, errConn
	}

	var roleRecord RoleTab
	var dbtmp *gorm.DB
	if constant.QueryKeyName == key {
		dbtmp = dbConn.Where("name = ? and model_status != ?", value, constant.ModelStatusDelete).First(&roleRecord)
	} else if constant.QueryKeyId == key {
		dbtmp = dbConn.Where("id = ? and model_status != ?", value, constant.ModelStatusDelete).First(&roleRecord)
	} else if constant.QueryKeyEntityNo == key {
		dbtmp = dbConn.Where("entity_no = ? and model_status != ?", value, constant.ModelStatusDelete).First(&roleRecord)
	} else {
		logrus.WithFields(logrus.Fields{"key": key}).Error("ExistInUserDetailTab Query Key Not Support")
		return false, errcode.CerrQueryKeyNotSupport
	}

	if err := dbtmp.Error; nil != err {
		if err == gorm.ErrRecordNotFound {
			logrus.WithFields(logrus.Fields{"key": key, "value": value}).Debug("Not found role record from DB")
			return false, nil //无记录不返回错误
		}
		logrus.WithFields(logrus.Fields{"error": err}).Error("Failed to get role record from DB by ids")
		return false, errcode.CerrExecuteSQL
	}

	logrus.WithFields(logrus.Fields{"key": key, "value": value}).Debug("Found role records from DB")
	return true, nil
}

//根据roleID获取role信息
func GetRoleInfoByID(tenant_id string, id int64) (*RoleTab, code.Error) {
	if "" == tenant_id {
		return nil, errcode.CerrParamater
	}
	dbConn, errConn := GetDBConnByTenantID(tenant_id)
	if errConn != nil {
		return nil, errConn
	}

	var roleInfo RoleTab
	dbtmp := dbConn.Where("id = ? and model_status != ?", id, constant.ModelStatusDelete).First(&roleInfo)

	if err := dbtmp.Error; nil != err {
		if err == gorm.ErrRecordNotFound {
			logrus.WithFields(logrus.Fields{
				"id": id,
			}).Warn("Not found role record from DB by ids")
			return nil, nil //未找到记录不返回错误
		}
		logrus.WithFields(logrus.Fields{
			"id":    id,
			"error": err,
		}).Error("Failed to get role record from DB by ids")
		return nil, errcode.CerrExecuteSQL
	}
	return &roleInfo, nil
}

//根据roleName获取role信息
func GetRoleInfoByName(tenant_id string, name string) (*RoleTab, code.Error) {
	if "" == tenant_id {
		return nil, errcode.CerrParamater
	}
	dbConn, errConn := GetDBConnByTenantID(tenant_id)
	if errConn != nil {
		return nil, errConn
	}
	var roleInfo RoleTab
	dbtmp := dbConn.Where("name = ? and model_status != ?", name, constant.ModelStatusDelete).First(&roleInfo)

	if err := dbtmp.Error; nil != err {
		if err == gorm.ErrRecordNotFound {
			logrus.WithFields(logrus.Fields{"name": name}).Warn("Not found role record from DB by name")
			return nil, nil //未找到记录不返回错误
		}
		logrus.WithFields(logrus.Fields{"name": name, "error": err}).Error("Failed to get role record from DB by name")
		return nil, errcode.CerrExecuteSQL
	}
	return &roleInfo, nil
}

//根据roleID获取子角色列表
func GetChildRoleListByParentid(tenant_id string, parentID int64) ([]RoleTab, code.Error) {
	if "" == tenant_id {
		return nil, errcode.CerrParamater
	}
	dbConn, errConn := GetDBConnByTenantID(tenant_id)
	if errConn != nil {
		return nil, errConn
	}

	var childList []RoleTab = make([]RoleTab, 0)
	dbtmp := dbConn.Where("parent_id = ? and model_status != ?", parentID, constant.ModelStatusDelete).Find(&childList)

	if err := dbtmp.Error; nil != err {
		if err == gorm.ErrRecordNotFound {
			logrus.Warnf("Not found child role record from DB, parentID:%v", parentID)
			return childList, nil //未找到记录不返回错误
		}
		logrus.Errorf("Failed to get child role record from DB, parentID:%v, error:%v", parentID, err)
		return nil, errcode.CerrExecuteSQL
	}
	return childList, nil
}

func ProcessGetRoleChild(feignKey *XFeignKey, roleID int64) ([]RoleSimple, code.Error) {
	if nil == feignKey || roleID <= 0 {
		return nil, errcode.CerrParamater
	}
	/*
		roleInfo, errGetRole := GetRoleInfoByID(feignKey.TenantId, roleID)
		if nil != errGetRole {
			logrus.Errorf("Failed to get role record from DB, roleID:%v", roleID)
			return nil, errGetRole
		}
		if nil == roleInfo {
			logrus.Errorf("Not found role record from DB, roleID:%v", roleID)
			return nil, errcode.CerrRoleIDNotExist
		}
	*/
	childList, errGetChild := GetChildRoleListByParentid(feignKey.TenantId, roleID)
	if nil != errGetChild {
		logrus.Errorf("GetChildRoleListByParentid failed, roleID:%v, err:%v", roleID, errGetChild)
		return nil, errGetChild
	}

	var roleSimples []RoleSimple = make([]RoleSimple, 0)
	for _, role := range childList {
		var simple RoleSimple
		simple.Id = role.Id
		simple.Name = role.Name
		simple.EntityNo = role.EntityNo
		roleSimples = append(roleSimples, simple)
	}
	return roleSimples, nil
}

//查询单个角色详情 子集
func ProcessGetRoleChildTree(feignKey *XFeignKey, roleID int64) ([]LazyTreeNodeDTO, code.Error) {
	if nil == feignKey || roleID <= 0 {
		return nil, errcode.CerrParamater
	}

	allRole, errGet := GetAllRoleLis(feignKey.TenantId)
	if nil != errGet {
		logrus.Errorf("GetAllRoleLis failed, roleID:%v, err:%v", roleID, errGet)
		return nil, errGet
	}
	//构建map，key为role parent_id, []RoleTab为children
	mapRole := make(map[int64][]RoleTab)
	for index, _ := range allRole {
		var parentId int64 = 0
		if allRole[index].ParentId.Valid {
			parentId = allRole[index].ParentId.Int64
		}
		childs, exist := mapRole[parentId]
		if false == exist {
			childs = make([]RoleTab, 0)
		}
		childs = append(childs, allRole[index])
		mapRole[parentId] = childs
	}

	var nodes []LazyTreeNodeDTO = make([]LazyTreeNodeDTO, 0)
	childList, exist := mapRole[roleID]
	if exist {
		for index, _ := range childList {
			var node LazyTreeNodeDTO
			node.Value = strconv.FormatInt(childList[index].Id, 10)
			node.Label = childList[index].Name
			node.Child = false
			if _, ok := mapRole[childList[index].Id]; ok { //判断是否有child
				node.Child = true
			}
			nodes = append(nodes, node)
		}
	}

	return nodes, nil
}

func getSubChildRoles(parentMapRoles map[int64][]RoleTab, roleId int64) []RoleTab {
	result := make([]RoleTab, 0)
	childs, exist := parentMapRoles[roleId]
	if exist {
		for index, _ := range childs {
			sub := getSubChildRoles(parentMapRoles, childs[index].Id)
			result = append(result, sub...)
		}
		result = append(result, childs...)
	}
	return result
}

//当前用户可以在新增/修改用户时设置的角色信息
func ProcessCurrentSetRole(feignKey *XFeignKey) ([]RoleSimple, code.Error) {
	if nil == feignKey {
		return nil, errcode.CerrParamater
	}
	//获取当前登录用户信息
	userInfo, errGetUser := GetUserDetailByIdOrPubid(feignKey, feignKey.UserId, 0)
	if nil != errGetUser {
		logrus.Errorf("Get current user info failed, error:%v", errGetUser)
		return nil, errGetUser
	}
	if nil == userInfo {
		logrus.Errorf("Not found user info, pubID:%v", feignKey.UserId)
		return nil, errcode.CerrAccountNotFound
	}
	//获取当前用户的role信息
	roleInfo, errGetRole := GetRoleInfoByID(feignKey.TenantId, userInfo.RoleId)
	if nil != errGetRole {
		logrus.Errorf("Failed to get role record from DB, roleID:%v", userInfo.RoleId)
		return nil, errGetRole
	}
	if nil == roleInfo {
		logrus.Errorf("Not found role record from DB, roleID:%v", userInfo.RoleId)
		return nil, errcode.CerrRoleIDNotExist
	}
	//获取当前用户的所有子role信息（含child的child）
	allRoles, errGetAll := GetAllRoleLis(feignKey.TenantId)
	if nil != errGetAll {
		logrus.Errorf("Get all roles failed")
		return nil, errGetAll
	}
	parentMapRoles := RoleTabList2ParentRoleTabMap(allRoles)
	childList := getSubChildRoles(parentMapRoles, roleInfo.Id)

	var roleSimples []RoleSimple = make([]RoleSimple, 0)
	var simple RoleSimple
	simple.Id = roleInfo.Id
	simple.Name = roleInfo.Name
	simple.EntityNo = roleInfo.EntityNo
	roleSimples = append(roleSimples, simple)
	for _, role := range childList {
		simple.Id = role.Id
		simple.Name = role.Name
		simple.EntityNo = role.EntityNo
		roleSimples = append(roleSimples, simple)
	}

	return roleSimples, nil
}

//获取某截止时间前创建的role数
func GetRoleListBeforeOneCreateTime(tenant_id string, deadTime time.Time) ([]RoleTab, code.Error) {
	if "" == tenant_id {
		return nil, errcode.CerrParamater
	}
	dbConn, errConn := GetDBConnByTenantID(tenant_id)
	if errConn != nil {
		return nil, errConn
	}

	var roleList []RoleTab = make([]RoleTab, 0)
	dbtmp := dbConn.Where("create_date < ? and model_status != ?", deadTime, constant.ModelStatusDelete).Find(&roleList)

	if err := dbtmp.Error; nil != err {
		if err == gorm.ErrRecordNotFound {
			logrus.WithFields(logrus.Fields{"deadTime time": deadTime}).Warn("Not found role Record record from DB")
			return roleList, nil //未找到记录不返回错误
		}
		logrus.WithFields(logrus.Fields{"deadTime time": deadTime, "error": err}).Error("Failed to get role record from DB")
		return nil, errcode.CerrExecuteSQL
	}

	return roleList, nil
}

//获取所有的role记录
func GetAllRoleLis(tenant_id string) ([]RoleTab, code.Error) {
	if "" == tenant_id {
		return nil, errcode.CerrParamater
	}
	dbConn, errConn := GetDBConnByTenantID(tenant_id)
	if errConn != nil {
		return nil, errConn
	}

	var allRoles []RoleTab = make([]RoleTab, 0)
	dbtmp := dbConn.Where("model_status != ?", constant.ModelStatusDelete).Find(&allRoles).Order("id ASC")

	if err := dbtmp.Error; nil != err {
		if err == gorm.ErrRecordNotFound {
			logrus.WithFields(logrus.Fields{}).Warn("Not found role record from DB")
			return allRoles, nil //未找到记录不返回错误
		}
		logrus.WithFields(logrus.Fields{"error": err}).Error("Failed to get role record from DB")
		return nil, errcode.CerrExecuteSQL
	}
	return allRoles, nil
}

//RoleTabList转换为RoleTabMap，map的key为roleId
func RoleTabList2RoleTabMap(roletList []RoleTab) map[int64]RoleTab {
	var mapInfo map[int64]RoleTab = make(map[int64]RoleTab)
	for index, _ := range roletList {
		mapInfo[roletList[index].Id] = roletList[index]
	}
	return mapInfo
}

//RoleTabList转换为ParentRoleTabMap，map的key为parentId
func RoleTabList2ParentRoleTabMap(roletList []RoleTab) map[int64][]RoleTab {
	mapInfo := make(map[int64][]RoleTab)
	for index, _ := range roletList {
		parentId := int64(0)
		if roletList[index].ParentId.Valid {
			parentId = roletList[index].ParentId.Int64
		}
		list, exist := mapInfo[parentId]
		if exist && nil != list {
			list = append(list, roletList[index])
		} else {
			list = []RoleTab{roletList[index]}
		}
		mapInfo[parentId] = list
	}
	return mapInfo
}

//根据roleId List获取role记录
func GetRoleLisByRoleIds(tenant_id string, roleIds []int64) ([]RoleTab, code.Error) {
	if "" == tenant_id || len(roleIds) == 0 {
		return nil, errcode.CerrParamater
	}
	dbConn, errConn := GetDBConnByTenantID(tenant_id)
	if errConn != nil {
		return nil, errConn
	}

	var roles []RoleTab = make([]RoleTab, 0)
	dbtmp := dbConn.Where("id in (?) and model_status != ?", roleIds, constant.ModelStatusDelete).Find(&roles)

	if err := dbtmp.Error; nil != err {
		if err == gorm.ErrRecordNotFound {
			logrus.WithFields(logrus.Fields{"roleIds": roleIds}).Warn("Not found role record from DB")
			return roles, nil //未找到记录不返回错误
		}
		logrus.WithFields(logrus.Fields{"error": err, "roleIds": roleIds}).Error("Failed to get role record from DB")
		return nil, errcode.CerrExecuteSQL
	}
	return roles, nil
}

type RoleDetailWithInfoSlice []RoleDetailWithInfo

func (a RoleDetailWithInfoSlice) Len() int {
	return len(a)
}
func (a RoleDetailWithInfoSlice) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a RoleDetailWithInfoSlice) Less(i, j int) bool {
	return a[j].Id < a[i].Id
}

func computeRoleDetailWithRightId(allRoles []RoleTab, allRelation []RoleRightRelationTab, language string) ([]RoleDetailWithInfo, code.Error) {
	if nil == allRoles || nil == allRelation {
		return nil, errcode.CerrParamater
	}
	allRoleType, _ := GetAllRoleType()
	roleTypeMap := make(map[string]RoleType)
	for _, item := range allRoleType {
		roleTypeMap[item.Id] = item
	}
	//由list转换成map，便于后面计算
	var mapRoles map[int64]RoleDetailWithInfo = make(map[int64]RoleDetailWithInfo, 0)
	for _, role := range allRoles {
		var roleDetail RoleDetailWithInfo
		roleDetail.Id = role.Id
		roleDetail.EntityNo = role.EntityNo
		roleDetail.Name = role.Name
		roleDetail.ParentId = role.ParentId.Int64
		roleDetail.SubRoleCount = 0           //后面计算追加
		roleDetail.Rights = make([]string, 0) //后面计算追加
		roleDetail.Comment = role.Comment
		roleDetail.BelongUserCount = role.BelongUserCount
		roleDetail.RoleTypeId = role.RoleTypeId
		if nil != &role.RoleTypeId && "" != role.RoleTypeId {
			roleType := roleTypeMap[role.RoleTypeId]
			if nil != &roleType {
				roleDetail.RoleTypeName = roleType.TypeName[language]
			}
		}
		mapRoles[role.Id] = roleDetail
	}
	//计算SubRoleCount和获取ParentName
	for key, value := range mapRoles {
		if value.ParentId != 0 && value.ParentId != key {
			parent, ok := mapRoles[value.ParentId]
			if ok {
				parent.SubRoleCount += 1
				value.ParentName = parent.Name
				mapRoles[key] = value
				mapRoles[value.ParentId] = parent
			}
		}
	}
	//追加rights
	for _, relation := range allRelation {
		if relation.RightID <= 0 {
			continue
		}
		roleDetail, ok := mapRoles[relation.RoleID]
		if ok {
			roleDetail.Rights = append(roleDetail.Rights, strconv.FormatInt(relation.RightID, 10))
			mapRoles[relation.RoleID] = roleDetail
		}
	}
	//从map转换成切片
	var roleResults []RoleDetailWithInfo = make([]RoleDetailWithInfo, 0)
	for _, roleDetail := range mapRoles {
		roleResults = append(roleResults, roleDetail)
	}
	//重新排序
	sort.Sort(sort.Reverse(RoleDetailWithInfoSlice(roleResults)))
	return roleResults, nil
}

func RoleTab2RoleDetail(record *RoleTab) (*RoleDetail, code.Error) {
	if nil == record {
		return nil, errcode.CerrParamater
	}
	var detail RoleDetail
	detail.Id = record.Id
	detail.EntityNo = record.EntityNo
	detail.CreateDate = record.CreateDate.Unix() * 1000
	if record.CreateDate.Unix() < constant.MIN_TIMESTAMP || record.CreateDate.Unix() > constant.MAX_TIMESTAMP {
		detail.CreateDate = constant.MIN_TIMESTAMP * 1000
	}
	detail.ModifyDate = record.ModifyDate.Unix() * 1000
	if record.ModifyDate.Unix() < constant.MIN_TIMESTAMP || record.ModifyDate.Unix() > constant.MAX_TIMESTAMP {
		detail.ModifyDate = constant.MIN_TIMESTAMP * 1000
	}
	detail.TenantId = record.TenantId
	detail.ProductId = record.TenantId
	detail.CreateUserId = record.CreateUserId
	detail.ModelStatus = record.ModelStatus
	detail.Name = record.Name
	detail.ParentId = record.ParentId.Int64
	detail.RoleRightflags = record.RoleRightFlags
	detail.BelongUserCount = record.BelongUserCount
	detail.SubRoleCount = record.SubRoleCount
	detail.Comment = record.Comment
	return &detail, nil
}

func GetRoleDetailWithRightId(feignKey *XFeignKey) ([]RoleDetailWithInfo, code.Error) {
	tenant_id := feignKey.TenantId
	if "" == tenant_id {
		return nil, errcode.CerrParamater
	}

	allRoles, errGetRole := GetAllRoleLis(tenant_id)
	if nil != errGetRole {
		logrus.WithFields(logrus.Fields{"error": errGetRole}).Error("GetAllRoleLis Failed")
		return nil, errGetRole
	} else if len(allRoles) == 0 {
		return nil, nil
	}

	allRelation, errGetRelation := GeAllRoleRightRelation(tenant_id)
	if nil != errGetRelation {
		logrus.WithFields(logrus.Fields{"error": errGetRelation}).Error("GeAllRoleRightRelation Failed")
		return nil, errGetRelation
	}
	/*
		allRight, errGetRight := GeAllRightRecords(tenant_id)
		if nil != errGetRight {
			logrus.WithFields(logrus.Fields{"error": errGetRight}).Error("GeAllRightRecords Failed")
			return nil, errGetRelation
		}
	*/
	roleDetails, errCompute := computeRoleDetailWithRightId(allRoles, allRelation, feignKey.Language)
	if nil != errCompute {
		logrus.WithFields(logrus.Fields{"error": errCompute}).Error("computeRoleDetailWithRightId Failed")
	}

	return roleDetails, nil
}

func ProcessGetRightIDsByRoleIDs(feignKey *XFeignKey, idSearchReq *IdSearchDTO) ([]int64, code.Error) {
	if nil == feignKey || "" == feignKey.TenantId || nil == idSearchReq || len(idSearchReq.Ids) == 0 {
		return nil, errcode.CerrParamater
	}

	dbConn, errConn := GetDBConnByTenantID(feignKey.TenantId)
	if errConn != nil {
		return nil, errConn
	}

	var roleIds []int64 = make([]int64, 0)
	for _, id := range idSearchReq.Ids {
		idInt, errParse := strconv.ParseInt(id, 10, 64)
		if nil != errParse {
			return nil, errcode.CerrParamater
		}
		roleIds = append(roleIds, idInt)
	}

	var relations []RoleRightRelationTab = make([]RoleRightRelationTab, 0)
	dbtmp := dbConn.Where("t_role in (?)", roleIds).Find(&relations)

	if err := dbtmp.Error; nil != err {
		if err == gorm.ErrRecordNotFound {
			logrus.WithFields(logrus.Fields{"ids": idSearchReq.Ids}).Warn("Not found relation record from DB")
			return nil, nil //未找到记录不返回错误
		}
		logrus.WithFields(logrus.Fields{"error": err, "ids": idSearchReq.Ids}).Error("Failed to get relation record from DB")
		return nil, errcode.CerrExecuteSQL
	}

	var rightList []int64 = make([]int64, 0)
	for _, value := range relations {
		rightList = append(rightList, value.RightID)
	}

	return rightList, nil
}

func ProcessAddRole(feignKey *XFeignKey, upsetReq *RoleDTO, parentId int64) code.Error {
	if nil == feignKey || "" == feignKey.TenantId || nil == upsetReq || "" == upsetReq.Name {
		return errcode.CerrParamater
	}

	//检查Name是否存在
	nameExist, errExist := ExistInRoleTab(feignKey.TenantId, constant.QueryKeyName, upsetReq.Name)
	if nil != errExist {
		logrus.WithFields(logrus.Fields{"name": upsetReq.Name, "error": errExist}).Error("Check exist in role table failed")
		return errExist
	}
	if true == nameExist {
		logrus.WithFields(logrus.Fields{"name": upsetReq.Name}).Error("ProcessAddRole, name have exist")
		return errcode.CerrNameExist
	}
	//检查entityNo是否存在
	if len(upsetReq.EntityNo) > 0 {
		entityExist, errEntity := ExistInRoleTab(feignKey.TenantId, constant.QueryKeyEntityNo, upsetReq.EntityNo)
		if nil != errEntity {
			logrus.WithFields(logrus.Fields{"entityNo": upsetReq.EntityNo, "error": errEntity}).Error("Check exist in role table failed")
			return errEntity
		}
		if true == entityExist {
			logrus.WithFields(logrus.Fields{"entityNo": upsetReq.EntityNo}).Error("ProcessAddRole, entityNo have exist")
			return errcode.CerrEntityNoExist
		}
	}

	var role RoleTab
	role.Name = upsetReq.Name
	role.Comment = upsetReq.Comment
	role.EntityNo = upsetReq.EntityNo
	role.CreateUserId = feignKey.UserId
	role.ModifyUserId = feignKey.UserId
	role.CreateDate = time.Now()
	role.ModifyDate = role.CreateDate
	role.ModelStatus = constant.ModelStatusCreate
	role.ProductId = "BW"
	role.TenantId = feignKey.TenantId
	role.BelongUserCount = 0
	role.RoleTypeId = upsetReq.RoleTypeId
	if parentId != 0 {
		role.ParentId.Int64 = parentId
		role.ParentId.Valid = true
	}

	dbConn, errConn := GetDBConnByTenantID(feignKey.TenantId)
	if errConn != nil {
		return errConn
	}
	//使用transaction保存记录
	tx := dbConn.Begin()
	if err := tx.Create(&role).Error; nil != err { //插入role信息
		logrus.WithFields(logrus.Fields{"name": upsetReq.Name, "parentId": parentId, "error": err}).Error("Add role info to DB failed")
		tx.Rollback()
		return errcode.CerrExecuteSQL
	}
	logrus.WithFields(logrus.Fields{"id": role.Id, "name": upsetReq.Name, "parentId": parentId}).Info("Add role info to DB success, wait commit")

	//生成批量插入relation的sql
	if len(upsetReq.RightIds) > 0 {
		insertRelationSqlStr := "INSERT INTO t_role_right_relation(t_role, role_rights) VALUES"
		roleIDStr := strconv.FormatInt(role.Id, 10)
		for index, rightid := range upsetReq.RightIds {
			rightIDStr := strconv.FormatInt(rightid, 10)
			insertRelationSqlStr += "(" + roleIDStr + ", " + rightIDStr + ")"
			if index < len(upsetReq.RightIds)-1 {
				insertRelationSqlStr += ","
			}
		}
		insertRelationSqlStr += ";"
		logrus.WithFields(logrus.Fields{"insert relation sql": insertRelationSqlStr}).Debug("Add role right relation info to DB sql")

		if err := tx.Exec(insertRelationSqlStr).Error; nil != err { //插入relation，批量插入
			logrus.WithFields(logrus.Fields{"insert relation sql": insertRelationSqlStr, "error": err}).Error("Exec SQL failed")
			tx.Rollback()
			return errcode.CerrExecuteSQL
		}
	}

	if err := tx.Commit().Error; nil != err { //Transaction commit
		logrus.WithFields(logrus.Fields{"error": err, "name": upsetReq.Name}).Error("Add role Transaction commit failed")
		tx.Rollback()
		return errcode.CerrExecuteSQL
	}

	//更新sub_role_count
	go UpdageRoleSubCounter(feignKey)
	logrus.WithFields(logrus.Fields{"id": role.Id, "name": upsetReq.Name, "parentId": parentId}).Info("Add role info to DB commit success")
	return nil
}

//检查role表，若将childId的parent设置成parentId，是否会存在循环，若存在则返回true，否则返回false
func checkRoleParentChildCycle(tenantId string, parentId int64, childId int64) (bool, code.Error) {
	if parentId == 0 || childId == 0 {
		return false, nil
	}
	if childId == parentId {
		return true, nil //存在循环
	}

	allRoll, errGetRole := GetAllRoleLis(tenantId)
	if nil != errGetRole {
		return false, errGetRole
	}
	//转换成map，便于计算
	var mapRole map[int64]RoleTab = make(map[int64]RoleTab, 0)
	for _, role := range allRoll {
		mapRole[role.Id] = role
	}

	parentIdTmp := parentId
	depth := 0
	for parentIdTmp != 0 {
		parent, ok := mapRole[parentIdTmp]
		if false == ok {
			break
		}
		if parent.ParentId.Int64 == childId {
			return true, nil //存在循环
		}
		parentIdTmp = parent.ParentId.Int64
		depth++
		if depth > 200 { //防止DB中的数据有异常导致程序死循环
			logrus.WithFields(logrus.Fields{"childId": childId, "parentId": parentId}).Warn("check role cycle loop 200 times")
			break
		}
	}
	return false, nil //不存在循环
}

func ProcessUpdateRole(feignKey *XFeignKey, upsetReq *RoleDTO, parentId int64) code.Error {
	if nil == feignKey || "" == feignKey.TenantId || nil == upsetReq || "" == upsetReq.Name || upsetReq.Id == 0 {
		return errcode.CerrParamater
	}
	//获取role信息
	oldChild, errGetChild := GetRoleInfoByID(feignKey.TenantId, upsetReq.Id)
	if nil != errGetChild {
		logrus.WithFields(logrus.Fields{"error": errGetChild, "roleID": upsetReq.Id}).Error("Get role info failed")
		return errGetChild
	} else if nil == oldChild {
		logrus.WithFields(logrus.Fields{"roleID": upsetReq.Id}).Error("This role not exist")
		return errcode.CerrRoleIDNotExist
	}

	//检查Name是否存在
	if oldChild.Name != upsetReq.Name {
		nameExist, errExist := ExistInRoleTab(feignKey.TenantId, constant.QueryKeyName, upsetReq.Name)
		if nil != errExist {
			logrus.WithFields(logrus.Fields{"name": upsetReq.Name, "error": errExist}).Error("Check exist in role table failed")
			return errExist
		}
		if true == nameExist {
			logrus.WithFields(logrus.Fields{"name": upsetReq.Name}).Error("ProcessUpdateRole, name have exist")
			return errcode.CerrNameExist
		}
	}

	//检查entityNo是否存在
	if len(upsetReq.EntityNo) > 0 && oldChild.EntityNo != upsetReq.EntityNo {
		entityExist, errEntity := ExistInRoleTab(feignKey.TenantId, constant.QueryKeyEntityNo, upsetReq.EntityNo)
		if nil != errEntity {
			logrus.WithFields(logrus.Fields{"entityNo": upsetReq.EntityNo, "error": errEntity}).Error("Check exist in role table failed")
			return errEntity
		}
		if true == entityExist {
			logrus.WithFields(logrus.Fields{"entityNo": upsetReq.EntityNo}).Error("ProcessUpdateRole, entityNo have exist")
			return errcode.CerrEntityNoExist
		}
	}

	//检查role里parent关系是否存在循环
	isCycle, errCheck := checkRoleParentChildCycle(feignKey.TenantId, parentId, upsetReq.Id)
	if nil != errCheck {
		logrus.WithFields(logrus.Fields{"error": errCheck}).Error("Check role cycle failed")
		return errCheck
	} else if true == isCycle {
		logrus.WithFields(logrus.Fields{"parentId": parentId, "update roleId": upsetReq.Id}).Error("Exist role cycle")
		return errcode.CerrCheckCycleNotPass
	}

	oldChild.EntityNo = upsetReq.EntityNo
	oldChild.Name = upsetReq.Name
	oldChild.Comment = upsetReq.Comment
	oldChild.RoleTypeId = upsetReq.RoleTypeId
	if parentId != 0 {
		oldChild.ParentId.Int64 = parentId
		oldChild.ParentId.Valid = true
	}

	if oldChild.CreateDate.Unix() < constant.MIN_TIMESTAMP || oldChild.CreateDate.Unix() > constant.MAX_TIMESTAMP { //避免原来的时间为null时保持到DB后时间变成'0000-00-00 ***'，然后Java读取数据时报错
		oldChild.CreateDate = time.Unix(constant.MIN_TIMESTAMP, 0)
	}
	oldChild.ModifyDate = time.Now()
	oldChild.ModifyUserId = feignKey.UserId

	dbConn, errConn := GetDBConnByTenantID(feignKey.TenantId)
	if errConn != nil {
		return errConn
	}
	//使用transaction保存记录
	tx := dbConn.Begin()
	if err := tx.Save(oldChild).Error; nil != err { //插入role信息
		logrus.WithFields(logrus.Fields{"name": upsetReq.Name, "parentId": parentId, "error": err}).Error("Update role info to DB failed")
		tx.Rollback()
		return errcode.CerrExecuteSQL
	}

	if err := tx.Where("t_role = ?", upsetReq.Id).Delete(RoleRightRelationTab{}).Error; nil != err { //删除此role的relation信息
		logrus.WithFields(logrus.Fields{"roleID": upsetReq.Id, "error": err}).Error("Delete relation failed")
		tx.Rollback()
		return errcode.CerrExecuteSQL
	}

	//生成批量插入更新relation的sql
	if len(upsetReq.RightIds) > 0 {
		replaceRelationSqlStr := "REPLACE INTO t_role_right_relation(t_role, role_rights) VALUES"
		roleIDStr := strconv.FormatInt(upsetReq.Id, 10)
		for index, rightid := range upsetReq.RightIds {
			rightIDStr := strconv.FormatInt(rightid, 10)
			replaceRelationSqlStr += "(" + roleIDStr + ", " + rightIDStr + ")"
			if index < len(upsetReq.RightIds)-1 {
				replaceRelationSqlStr += ","
			}
		}
		replaceRelationSqlStr += ";"
		logrus.WithFields(logrus.Fields{"sql": replaceRelationSqlStr}).Debug("replace into role right relation info to DB sql")
		if err := tx.Exec(replaceRelationSqlStr).Error; nil != err { //插入relation，批量更新
			logrus.WithFields(logrus.Fields{"relation sql": replaceRelationSqlStr, "error": err}).Error("Exec SQL failed")
			tx.Rollback()
			return errcode.CerrExecuteSQL
		}
	}

	//更新相关role用户的roleName
	errUpdateUser := tx.Table("t_user_detail").Where("role_id = ? and model_status != ?", upsetReq.Id, constant.ModelStatusDelete).Updates(map[string]interface{}{"role_name": upsetReq.Name}).Error
	if nil != errUpdateUser && gorm.ErrRecordNotFound != errUpdateUser {
		logrus.WithFields(logrus.Fields{"roleId": upsetReq.Id, "roleName": upsetReq.Name, "error": errUpdateUser}).Error("update user roleName failed")
		tx.Rollback()
		return errcode.CerrExecuteSQL
	}

	if err := tx.Commit().Error; nil != err { //Transaction commit
		logrus.WithFields(logrus.Fields{"error": err, "name": upsetReq.Name, "roleId": upsetReq.Id}).Error("Update role Transaction commit failed")
		tx.Rollback()
		return errcode.CerrExecuteSQL
	}

	//更新sub_role_count
	go UpdageRoleSubCounter(feignKey)
	return nil
}

func ProcessGetSimpleRoleByRight(feignKey *XFeignKey, rightEntityNo string) ([]RoleDetail, code.Error) {
	tenant_id := feignKey.TenantId
	rightRecords, errGetRight := GeRightRecordsByEntityno(tenant_id, []string{rightEntityNo})
	if nil != errGetRight {
		logrus.WithFields(logrus.Fields{"error": errGetRight, "rightEntityNo": rightEntityNo}).Error("GeRightRecordsByEntityno Failed")
		return nil, errGetRight
	} else if len(rightRecords) == 0 {
		return nil, nil
	}

	//获取相关权限的relation信息
	rightIds := make([]int64, 0)
	for _, right := range rightRecords {
		rightIds = append(rightIds, right.Id)
	}
	relationRecords, errGetRelation := GetRoleRightRelationByRightIds(tenant_id, rightIds)
	if nil != errGetRelation {
		logrus.WithFields(logrus.Fields{"error": errGetRelation}).Error("GeAllRoleRightRelation Failed")
		return nil, errGetRelation
	} else if len(relationRecords) == 0 {
		return nil, nil
	}

	//获取相关的role信息
	roleIds := make([]int64, 0)
	for _, relation := range relationRecords {
		roleIds = append(roleIds, relation.RoleID)
	}
	roleRecords, errGetRole := GetRoleLisByRoleIds(tenant_id, roleIds)
	if nil != errGetRole {
		logrus.WithFields(logrus.Fields{"error": errGetRole}).Error("GetAllRoleLis Failed")
		return nil, errGetRole
	}

	roleDetails := make([]RoleDetail, 0)
	for _, record := range roleRecords {
		detail, errDetail := RoleTab2RoleDetail(&record)
		if nil == errDetail {
			roleDetails = append(roleDetails, *detail)
		}
	}

	return roleDetails, nil
}

//生成AdminRole的基本信息（还未保存到DB）
func generateAdminRole(feignKey *XFeignKey, language string) (*RoleTab, []RightTab, code.Error) {
	allRights, errGetRight := GeAllRightRecords(feignKey.TenantId)
	if nil != errGetRight {
		logrus.WithFields(logrus.Fields{"error": errGetRight}).Error("GeAllRightRecords failed")
		return nil, nil, errGetRight
	}
	if len(allRights) == 0 {
		//需要从配置中获取此租户的所有权限，并添加到t_right表中
		errInitRight := ProcessInitRight(feignKey)
		if nil != errInitRight {
			logrus.WithFields(logrus.Fields{"error": errInitRight}).Error("ProcessInitRight failed")
			return nil, nil, errInitRight
		}

		allRights, errGetRight = GeAllRightRecords(feignKey.TenantId)
		if nil != errGetRight {
			logrus.WithFields(logrus.Fields{"error": errGetRight}).Error("GeAllRightRecords failed")
			return nil, nil, errGetRight
		}
	}

	var role RoleTab
	role.Name = createRoleName(language)
	role.CreateUserId = feignKey.UserId
	role.ModifyUserId = feignKey.UserId
	role.CreateDate = time.Now()
	role.ModifyDate = role.CreateDate
	role.ModelStatus = constant.ModelStatusCreate
	role.ProductId = "BW"
	role.TenantId = feignKey.TenantId
	role.BelongUserCount = 0
	role.ParentId.Valid = false
	role.RoleTypeId = "administrator"
	return &role, allRights, nil
}

//生成自定义Role的基本信息（还未保存到DB）
func generateCustomRole(feignKey *XFeignKey, language string, roleRight []string) (*RoleTab, []RightTab, code.Error) {
	allRights, errGetRight := GeAllRightRecords(feignKey.TenantId)
	if nil != errGetRight {
		logrus.WithFields(logrus.Fields{"error": errGetRight}).Error("GeAllRightRecords failed")
		return nil, nil, errGetRight
	}

	var retainRights []RightTab = make([]RightTab, 0)
	for index, _ := range allRights {
		if RightHaveContains(allRights[index], roleRight) {
			retainRights = append(retainRights, allRights[index])
		}
	}
	var role RoleTab
	role.Name = createCustomRoleName(language, roleRight[0:4])
	role.CreateUserId = feignKey.UserId
	role.ModifyUserId = feignKey.UserId
	role.CreateDate = time.Now()
	role.ModifyDate = role.CreateDate
	role.ModelStatus = constant.ModelStatusCreate
	role.ProductId = "BW"
	role.TenantId = feignKey.TenantId
	role.BelongUserCount = 0
	role.ParentId.Valid = false
	role.RoleTypeId = roleRight[5]
	return &role, retainRights, nil
}

//模糊查询角色，目前按名称，编号查询
func RoleMsgReceivers(feignKey *XFeignKey, receiverReq *MsgReceiversSearchDTO) ([]MsgReceiversDTO, code.Error) {
	if nil == feignKey || "" == feignKey.TenantId || nil == receiverReq {
		return nil, errcode.CerrParamater
	}
	var msgList []MsgReceiversDTO = make([]MsgReceiversDTO, 0)
	dbConn, errConn := GetDBConnByTenantID(feignKey.TenantId)
	if errConn != nil {
		return nil, errConn
	}

	likeFuzzyVal := "%" + escapeSql(receiverReq.FuzzyVal) + "%"
	var roleList []RoleTab = make([]RoleTab, 0)
	dbtmp := dbConn.Where("(name like ? or entity_no like ?) and model_status != ?", likeFuzzyVal, likeFuzzyVal, constant.ModelStatusDelete).Find(&roleList)
	if err := dbtmp.Error; nil != err {
		if err == gorm.ErrRecordNotFound {
			logrus.Warnf("Not found role record from DB, FuzzyVal:%v", likeFuzzyVal)
			return msgList, nil //未找到记录不返回错误
		}
		logrus.Errorf("Failed to get role record from DB, FuzzyVal:%v, error:%v", likeFuzzyVal, err)
		return nil, errcode.CerrExecuteSQL
	}

	for index, _ := range roleList {
		var msg MsgReceiversDTO
		msg.Id = strconv.FormatInt(roleList[index].Id, 10)
		msg.Name = roleList[index].Name
		msg.IdType = constant.IdType_RoleId
		msgList = append(msgList, msg)
	}
	return msgList, nil
}

//模糊查询角色，目前按名称查询
func SearchRoleByFuzzyName(feignKey *XFeignKey, fuzzyValue string) ([]*RoleTab, code.Error) {
	if nil == feignKey || "" == feignKey.TenantId || "" == fuzzyValue {
		return nil, errcode.CerrParamater
	}
	dbConn, errConn := GetDBConnByTenantID(feignKey.TenantId)
	if errConn != nil {
		return nil, errConn
	}

	likeFuzzyVal := "%" + escapeSql(fuzzyValue) + "%"
	var roleList = make([]*RoleTab, 0)
	dbtmp := dbConn.Where("name like ? and model_status != ?", likeFuzzyVal, constant.ModelStatusDelete).Find(&roleList)
	if err := dbtmp.Error; nil != err {
		if err == gorm.ErrRecordNotFound {
			logrus.Warnf("Not found role record from DB, FuzzyVal:%v", likeFuzzyVal)
			return roleList, nil //未找到记录不返回错误
		}
		logrus.Errorf("Failed to get role record from DB, FuzzyVal:%v, error:%v", likeFuzzyVal, err)
		return nil, errcode.CerrExecuteSQL
	}
	return roleList, nil
}

func ProcessRemoveRole(feignKey *XFeignKey, roleId int64) code.Error {
	tenant_id := feignKey.TenantId
	if "" == tenant_id || roleId == 0 {
		return errcode.CerrParamater
	}
	//检查是否存在此角色的用户
	userList, errGetUser := ProcessListUserRecordsByKey(feignKey, []int64{roleId}, constant.QueryKeyRoleId)
	if nil != errGetUser {
		logrus.WithFields(logrus.Fields{"error": errGetUser}).Error("ProcessListUserRecordsByKey failed")
		return errGetUser
	} else if len(userList) > 0 {
		logrus.WithFields(logrus.Fields{"roleId": roleId, "user Count": len(userList)}).Error("Remove role exist user failed")
		return errcode.CerrRoleIDExistUser
	}

	//检查此角色是否有子角色
	allRole, errGetAll := GetAllRoleLis(feignKey.TenantId)
	if nil != errGetAll {
		logrus.WithFields(logrus.Fields{"error": errGetAll}).Error("GetAllRoleLis failed")
		return errGetAll
	}
	for index, _ := range allRole {
		if allRole[index].ParentId.Valid && allRole[index].ParentId.Int64 == roleId {
			logrus.WithFields(logrus.Fields{"roleId": roleId, "subRoleId": allRole[index].Id}).Error("Remove role exist sub role failed")
			return errcode.CerrRoleIDExistSubRole
		}
	}

	//在t_role表中删除用户信息
	//暂时采用硬删除DB记录，后面通过修改状态实现////////////////
	dbConn, errConn := GetDBConnByTenantID(tenant_id)
	if errConn != nil {
		return errConn
	}
	tx := dbConn.Begin()

	if errDelete := tx.Where("t_role = ?", roleId).Delete(RoleRightRelationTab{}).Error; nil != errDelete {
		logrus.WithFields(logrus.Fields{"t_role": roleId, "error": errDelete}).Errorf("Delete role right relation record failed")
		tx.Rollback()
		return errcode.CerrExecuteSQL
	}

	var roleRecord RoleTab
	roleRecord.Id = roleId
	if errDelete := tx.Delete(&roleRecord).Error; nil != errDelete {
		logrus.Errorf("Delete user record failed, roleId:%v, error:%v", roleId, errDelete)
		tx.Rollback()
		return errcode.CerrExecuteSQL
	}

	if err := tx.Commit().Error; nil != err {
		logrus.WithFields(logrus.Fields{"t_role": roleId, "error": err}).Errorf("Delete role commit failed")
		return errcode.CerrExecuteSQL
	}

	//更新sub_role_count
	go UpdageRoleSubCounter(feignKey)
	logrus.WithFields(logrus.Fields{"roleId": roleId, "operate User": feignKey.UserId}).Infof("Remove roleId")
	return nil
}

//全量计算用户的下级用户数，并更新到DB
func UpdageRoleSubCounter(feignKey *XFeignKey) code.Error {
	allRole, errGetAll := GetAllRoleLis(feignKey.TenantId)
	if nil != errGetAll {
		logrus.WithFields(logrus.Fields{"error": errGetAll}).Error("GetAllRoleLis failed")
		return errGetAll
	}

	type subCount struct {
		oldSubCount int
		newSubCount int
	}
	type roleCount struct {
		oldRoleCount int
		newRoleCount int
	}

	roleIdSubCountMap := make(map[int64]subCount) //计算得到的下级角色数
	idParentMap := make(map[int64]int64)          //key为userId，value为parentId
	subRoleMap := make(map[int64][]int64)         //key为parentid, value为直接下级RoleID
	for index, _ := range allRole {
		var counter subCount
		counter.oldSubCount = allRole[index].SubRoleCount //保存原始的下级角色数
		counter.newSubCount = 0
		roleIdSubCountMap[allRole[index].Id] = counter
		var parentId int64 = 0
		if allRole[index].ParentId.Valid {
			parentId = allRole[index].ParentId.Int64
		}
		//构造parentid为key，subroleList为value的map
		subList, exist := subRoleMap[parentId]
		if false == exist {
			subList = make([]int64, 0)
		}
		subList = append(subList, allRole[index].Id)
		subRoleMap[parentId] = subList
		idParentMap[allRole[index].Id] = parentId
	}

	for parentId, subList := range subRoleMap {
		depth := 0
		for parentId != 0 && depth < 100 {
			if counter, exist := roleIdSubCountMap[parentId]; exist {
				counter.newSubCount += len(subList)
				roleIdSubCountMap[parentId] = counter
			}
			existParent := false
			if parentId, existParent = idParentMap[parentId]; false == existParent {
				parentId = 0
			}
			depth++
		}
	}

	dbConn, errConn := GetDBConnByTenantID(feignKey.TenantId)
	if errConn != nil {
		return errConn
	}
	//更新下级用户数到DB
	var roleRecord RoleTab
	for roleId, counter := range roleIdSubCountMap {
		if counter.newSubCount == counter.oldSubCount {
			continue
		}
		roleRecord.Id = roleId
		err := dbConn.Model(&roleRecord).Where("id = ?", roleId).Update("sub_role_count", counter.newSubCount).Error
		if nil != err {
			logrus.WithFields(logrus.Fields{"error": err, "roleId": roleId, "subRoleCount": counter.newSubCount}).Error("update sub_role_count failed")
			//不返回错误
		}
		logrus.WithFields(logrus.Fields{"roleId": roleId, "subRoleCount": counter.newSubCount}).Debug("update sub_role_count success")
	}
	return nil
}

// 获取所有角色类型
func GetAllRoleType() ([]RoleType, code.Error) {
	roleTypes := make([]RoleType, 0)
	if errGet := MongoFindAll(ROLE_TYPE_COLLECTION, map[string]string{}, &roleTypes); nil != errGet {
		return nil, errcode.CerrOperateMongo
	}
	return roleTypes, nil
}

// 获取指定ID的角色类型
func GetRoleTypeById(roleTypeId string) (*RoleType, code.Error) {
	var roleType RoleType
	if errGet := MongoFindOne(ROLE_TYPE_COLLECTION, bson.M{"_id": roleTypeId}, &roleType); nil != errGet && errGet != mgo.ErrNotFound {
		return nil, errcode.CerrOperateMongo
	}
	return &roleType, nil
}

// 根据角色名称获取角色类型
func GetRoleTypeInfo4User(tenantId string, roleId int64) (string, string) {
	roleInfo, errGetRole := GetRoleInfoByID(tenantId, roleId)
	if nil == errGetRole && nil != roleInfo {
		roleType, errGetRoleType := GetRoleTypeById(roleInfo.RoleTypeId)
		if nil == errGetRoleType && nil != roleType {
			return roleType.Id, roleType.TypeName["zh-CN"]
		}
	}
	return "", ""
}

// 刷新角色
func FreshRoleType(tenantId string) code.Error {
	dbConn, errConn := GetDBConnByTenantID(tenantId)
	if errConn != nil {
		return errConn
	}
	roles, errGetRole := GetAllRoleLis(tenantId)
	if nil != errGetRole {
		return errGetRole
	}
	var roleRecord RoleTab
	for _, item := range roles {
		if nil != &item.Name && "" != item.Name && (item.RoleTypeId == "" || nil == &item.RoleTypeId) {
			roleType := MatchRoleTypeAccordingRoleName(item.Name)
			logrus.WithFields(logrus.Fields{"tenantId": tenantId, "roleId": item.Id, "roleName": item.Name, "roleTypeId": roleType}).Debug("fresh role type params")
			if "" != roleType {
				err := dbConn.Model(&roleRecord).Where("id = ?", item.Id).Update("role_type_id", roleType).Error
				if nil != err {
					logrus.WithFields(logrus.Fields{"error": err, "tenantId": tenantId, "roleId": item.Id, "roleTypeId": roleType}).Error("fresh role type failed")
					return errcode.CerrExecuteSQL
				}
				logrus.WithFields(logrus.Fields{"tenantId": tenantId, "roleId": item.Id, "roleTypeId": roleType}).Debug("fresh role type success")
			}
		}
	}
	return nil
}

//根据条件查询用户
func SearchRoleByCondition(feignKey *XFeignKey, searchReq *SearchDTO) ([]*RoleTab, code.Error) {
	if nil == feignKey || "" == feignKey.TenantId {
		return nil, errcode.CerrParamater
	}
	//自定义查询角色
	whereSql, advanceSearchErr := getAdvanceSearchConditionForRole(feignKey, searchReq.Conditions)
	if nil != advanceSearchErr {
		logrus.WithFields(logrus.Fields{"error": advanceSearchErr}).Error("getAdvanceSearchConditionForRole failed")
		return nil, advanceSearchErr
	}
	dbConn, errConn := GetDBConnByTenantID(feignKey.TenantId)
	if errConn != nil {
		return nil, errConn
	}

	roleRecords := make([]*RoleTab, 0)
	var dbtmp *gorm.DB = nil
	if len(searchReq.Fields) == 0 {
		dbtmp = dbConn.Where("1 = 1" + whereSql).Find(&roleRecords)
	} else {
		dbtmp = dbConn.Select(strings.Join(searchReq.Fields, ",")).Where("1 = 1" + whereSql).Find(&roleRecords)
	}

	if err := dbtmp.Error; nil != err {
		if err == gorm.ErrRecordNotFound {
			logrus.WithFields(logrus.Fields{"whereSql": whereSql}).Debug("Not found role records from DB")
			return roleRecords, nil //未找到记录不返回错误
		}
		logrus.WithFields(logrus.Fields{"whereSql": whereSql, "error": err}).Error("Get role records from DB failed")
		return nil, errcode.CerrExecuteSQL
	}
	return roleRecords, nil
}

// 根据角色名称匹配角色类型
func MatchRoleTypeAccordingRoleName(roleName string) string {
	if StringListIsContainIgnoreCase([]string{"老板", "总经理", "总裁", "CEO", "董事长", "boss", "老闆", "總經理", "總裁", "董事長"}, roleName) {
		return "boss"
	}
	if StringListContainContain([]string{"管理员", "admin", "后台管理", "backoffice", "管理員", "後台管理"}, roleName) {
		return "administrator"
	}
	if StringListContainContain([]string{"销售总监", "销售主管", "业务总监", "业务总管", "銷售總監", "業務總監", "業務總管", "Sales Leader", "Sales Director"}, roleName) {
		return "sales_leader"
	}
	if StringListContainContain([]string{"销售", "电销", "业务", "sale", "客户经理", "銷售", "電銷", "業務", "客戶經理", "business"}, roleName) {
		return "sales"
	}
	if StringListContainContain([]string{"客服主管", "客服经理", "客服总管", "客服总监", "客服經理", "客服總管", "客服總監", "Customer Service Director", "CS Director"}, roleName) {
		return "customer_service_leader"
	}
	if StringListContainContain([]string{"客服", "客诉", "service", "客訴", "CS"}, roleName) {
		return "customer_service"
	}
	if StringListContainContain([]string{"财务", "会计", "财会", "finance", "財務", "會計", "財會"}, roleName) {
		return "finance_manager"
	}
	if StringListContainContain([]string{"市场总监", "市场主管", "营销总监", "行销总监", "市場總監", "市場主管", "營銷總監", "行銷總監", "Marketing Director", "MKT Director"}, roleName) {
		return "market_leader"
	}
	if StringListContainContain([]string{"市场", "市場", "行销", "行銷", "market", "Marketing", "MKT"}, roleName) {
		return "market"
	}
	if StringListContainContain([]string{"运营", "運營", "Operation"}, roleName) {
		return "operating"
	}
	if StringListContainContain([]string{"风控", "风险", "風控", "風險", "risk"}, roleName) {
		return "risk"
	}
	if StringListContainContain([]string{"监管", "合规", "監管", "合規", "Compliance"}, roleName) {
		return "compliance"
	}
	if StringListContainContain([]string{"代理", "IB", "经纪", "經紀", "Broker"}, roleName) {
		return "ib"
	}
	return ""
}

func getAdvanceSearchConditionForRole(feignKey *XFeignKey, searchReq []*AdvanceCondition) (string, code.Error) {
	if len(searchReq) == 0 {
		return "", nil
	}
	var conditionList []string
	for index := range searchReq {
		oneCondition := searchReq[index]
		switch oneCondition.Field {
		case constant.SearchField_Id:
			conditionList = append(conditionList, fmt.Sprintf(" id in (%v) ", escapeSql(oneCondition.Value)))
			break
		case constant.SearchField_Name:
			conditionList = append(conditionList, fmt.Sprintf(" name like '%%%v%%' ", escapeSql(oneCondition.Value)))
			break
		case constant.SearchField_EntityNo:
			conditionList = append(conditionList, fmt.Sprintf(" entity_no like '%%%v%%' ", escapeSql(oneCondition.Value)))
			break
		case constant.SearchField_NameOrEntityNo:
			conditionList = append(conditionList, fmt.Sprintf(" (name like '%%%v%%' or entity_no like '%%%v%%') ", escapeSql(oneCondition.Value), escapeSql(oneCondition.Value)))
			break
		case constant.SearchField_Right:
			rightEntityNo := oneCondition.Value
			roleIdList, err := GetRoleIdListByRightEntityNo(feignKey.TenantId, rightEntityNo)
			if nil != err {
				logrus.WithFields(logrus.Fields{"error": err, "TenantId": feignKey.TenantId, "rightEntityNo": rightEntityNo}).Error("GetRoleIdListByRightEntityNo Failed")
				return "", err
			}
			roleIdStr := "'-2838383'" //添加一个不可能存在的ID，避免传空集合
			for _, id := range roleIdList {
				roleIdStr += fmt.Sprintf(",'%v'", id)
			}
			conditionList = append(conditionList, fmt.Sprintf(" id in (%v) ", roleIdStr))
			break
		default:
			logrus.WithFields(logrus.Fields{"Field": oneCondition.Field}).Error("getAdvanceSearchConditionForRole keyType not support")
			return "", errcode.CerrQueryKeyNotSupport
		}

	}
	if len(conditionList) > 0 {
		return " and " + strings.Join(conditionList, " and "), nil
	}
	return "", nil
}
