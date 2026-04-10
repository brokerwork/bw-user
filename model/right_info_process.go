package model

import (
	"bw/bw-user/constant"
	"bw/bw-user/errcode"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/jinzhu/gorm"
	"github.com/lworkltd/kits/service/restful/code"
)

//获取所有的role right relation
func GeAllRoleRightRelation(tenant_id string) ([]RoleRightRelationTab, code.Error) {
	if "" == tenant_id {
		return nil, errcode.CerrParamater
	}
	dbConn, errConn := GetDBConnByTenantID(tenant_id)
	if errConn != nil {
		return nil, errConn
	}

	var allRelation []RoleRightRelationTab = make([]RoleRightRelationTab, 0)
	dbtmp := dbConn.Find(&allRelation)

	if err := dbtmp.Error; nil != err {
		if err == gorm.ErrRecordNotFound {
			logrus.WithFields(logrus.Fields{}).Warn("Not found relation record from DB")
			return allRelation, nil //未找到记录不返回错误
		}
		logrus.WithFields(logrus.Fields{"error": err}).Error("Failed to get relation record from DB")
		return nil, errcode.CerrExecuteSQL
	}
	return allRelation, nil
}

//根据rightid列表获取其所有的role right relation
func GetRoleRightRelationByRightIds(tenant_id string, rightIds []int64) ([]RoleRightRelationTab, code.Error) {
	if "" == tenant_id || len(rightIds) == 0 {
		return nil, errcode.CerrParamater
	}
	dbConn, errConn := GetDBConnByTenantID(tenant_id)
	if errConn != nil {
		return nil, errConn
	}

	var allRelation []RoleRightRelationTab = make([]RoleRightRelationTab, 0)
	dbtmp := dbConn.Where("role_rights in (?)", rightIds).Find(&allRelation)

	if err := dbtmp.Error; nil != err {
		if err == gorm.ErrRecordNotFound {
			logrus.WithFields(logrus.Fields{"rightIds": rightIds}).Warn("Not found relation record from DB")
			return allRelation, nil //未找到记录不返回错误
		}
		logrus.WithFields(logrus.Fields{"error": err, "rightIds": rightIds}).Error("Failed to get relation record from DB")
		return nil, errcode.CerrExecuteSQL
	}
	return allRelation, nil
}

//获取所有的right记录信息
func GeAllRightRecords(tenant_id string) ([]RightTab, code.Error) {
	if "" == tenant_id {
		return nil, errcode.CerrParamater
	}
	dbConn, errConn := GetDBConnByTenantID(tenant_id)
	if errConn != nil {
		return nil, errConn
	}

	var allRight []RightTab = make([]RightTab, 0)
	dbtmp := dbConn.Where("model_status != ?", constant.ModelStatusDelete).Find(&allRight)

	if err := dbtmp.Error; nil != err {
		if err == gorm.ErrRecordNotFound {
			logrus.WithFields(logrus.Fields{}).Warn("Not found right record from DB")
			return allRight, nil //未找到记录不返回错误
		}
		logrus.WithFields(logrus.Fields{"error": err}).Error("Failed to get right record from DB")
		return nil, errcode.CerrExecuteSQL
	}
	return allRight, nil
}

//RightTabList转换为RightTabMap，map的key为rightId
func RightTabList2RightTabMap(rightList []RightTab) map[int64]RightTab {
	var mapInfo map[int64]RightTab = make(map[int64]RightTab)
	for index, _ := range rightList {
		mapInfo[rightList[index].Id] = rightList[index]
	}
	return mapInfo
}

func RightTabList2RightCodeTabMap(rightList []RightTab) map[string]RightTab {
	var mapInfo = make(map[string]RightTab)
	for index, _ := range rightList {
		mapInfo[rightList[index].EntityNo] = rightList[index]
	}
	return mapInfo
}

//根据rightEntityNo获取其所有的right记录信息
func GeRightRecordsByEntityno(tenant_id string, rightEntityNoList []string) ([]RightTab, code.Error) {
	if "" == tenant_id || len(rightEntityNoList) == 0 {
		return nil, errcode.CerrParamater
	}
	dbConn, errConn := GetDBConnByTenantID(tenant_id)
	if errConn != nil {
		return nil, errConn
	}

	rightEntityNoStr := "'xx'"
	for _, rightEntityNo := range rightEntityNoList {
		rightEntityNoStr += fmt.Sprintf(",'%v'", rightEntityNo)
	}

	query := fmt.Sprintf("entity_no in (%v) and model_status != '%v'", rightEntityNoStr, constant.ModelStatusDelete)

	allRight := make([]RightTab, 0)
	//dbtmp := dbConn.Where("entity_no in (?) and model_status != ?", rightEntityNoStr, constant.ModelStatusDelete).Find(&allRight)
	dbtmp := dbConn.Where(query).Find(&allRight)

	if err := dbtmp.Error; nil != err {
		if err == gorm.ErrRecordNotFound {
			logrus.WithFields(logrus.Fields{"entityNo": rightEntityNoList}).Warn("Not found right record from DB")
			return allRight, nil //未找到记录不返回错误
		}
		logrus.WithFields(logrus.Fields{"error": err, "entityNo": rightEntityNoList}).Error("Failed to get right record from DB")
		return nil, errcode.CerrExecuteSQL
	}
	return allRight, nil
}

/*
//根据type获取其所有的right记录信息，若"" == typeStr则获取所有记录
func GeRightRecordsByType(tenant_id string, typeStr string) ([]RightTab, code.Error) {
	if "" == tenant_id {
		return nil, errcode.CerrParamater
	}
	dbConn, errConn := GetDBConnByTenantID(tenant_id)
	if errConn != nil {
		return nil, errConn
	}

	var rightRecord []RightTab = make([]RightTab, 0)
	var dbtmp *gorm.DB
	if "" == typeStr {
		dbtmp = dbConn.Where("model_status != ?", typeStr, constant.ModelStatusDelete).Find(&rightRecord)
	} else {
		dbtmp = dbConn.Where("type = ? and model_status != ?", typeStr, constant.ModelStatusDelete).Find(&rightRecord)
	}

	if err := dbtmp.Error; nil != err {
		if err == gorm.ErrRecordNotFound {
			logrus.WithFields(logrus.Fields{"typeStr": typeStr}).Warn("Not found right record from DB")
			return rightRecord, nil				//未找到记录不返回错误
		}
		logrus.WithFields(logrus.Fields{"error": err, "typeStr": typeStr}).Error("Failed to get right record from DB")
		return nil, errcode.CerrExecuteSQL
	}
	return rightRecord, nil
}
*/

//根据RoleID获取其right信息
func GetOneRoleRightList(feignKey *XFeignKey, roleID int64) ([]RightTab, code.Error) {
	tenant_id := feignKey.TenantId
	if "" == tenant_id {
		return nil, errcode.CerrParamater
	}
	dbConn, errConn := GetDBConnByTenantID(tenant_id)
	if errConn != nil {
		return nil, errConn
	}

	var relations []RoleRightRelationTab = make([]RoleRightRelationTab, 0)
	dbtmp := dbConn.Where("t_role = ?", roleID).Find(&relations) //按sid排序
	if err := dbtmp.Error; nil != err {
		if err == gorm.ErrRecordNotFound {
			logrus.Warnf("Not found role right relation record from DB, roleID:%v", roleID)
			return nil, nil //未找到记录不返回错误
		}
		logrus.Errorf("Failed to get role right relation record from DB, error:%v", err)
		return nil, errcode.CerrExecuteSQL
	}

	var rightIdList []int64 = make([]int64, 0)
	for _, relation := range relations {
		rightIdList = append(rightIdList, relation.RightID)
	}
	//logrus.Debugf("roleID:%v, rightIDList:%v", roleID, rightIdList)

	var rightRecords []RightTab = make([]RightTab, 0)
	dbtmp = dbConn.Where("id in (?) and model_status != ?", rightIdList, constant.ModelStatusDelete).Find(&rightRecords)
	if err := dbtmp.Error; nil != err {
		if err == gorm.ErrRecordNotFound {
			logrus.Warnf("Not found right record from DB, roleID:%v", roleID)
			return rightRecords, nil //未找到记录不返回错误
		}
		logrus.Errorf("Failed to get right record from DB, roleID:%v, error:%v", roleID, err)
		return nil, errcode.CerrExecuteSQL
	}

	return rightRecords, nil
}

func RightsHaveCotain(rightRecords []RightTab, rightEntityNo string) bool {
	for index, _ := range rightRecords {
		if rightEntityNo == rightRecords[index].EntityNo {
			//logrus.Debugf("rightEntityNo:%v, rightID:%v", rightEntityNo, right.Id)
			return true
		}
	}
	return false
}

func RightHaveContains(rightRecord RightTab, rightEntityNos []string) bool {
	for index, _ := range rightEntityNos {
		if rightRecord.EntityNo == rightEntityNos[index] {
			return true
		}
	}
	return false
}

func RightTab2RightDetail(record *RightTab) (*RightDetail, code.Error) {
	if nil == record {
		return nil, errcode.CerrParamater
	}
	var detail RightDetail
	detail.Id = record.Id
	detail.EntityNo = record.EntityNo
	detail.CreateDate = record.CreateDate.Unix() * 1000
	if record.CreateDate.Unix() < constant.MIN_TIMESTAMP {
		detail.CreateDate = constant.MIN_TIMESTAMP * 1000
	}
	detail.ModifyDate = record.ModifyDate.Unix() * 1000
	if record.ModifyDate.Unix() < constant.MIN_TIMESTAMP {
		detail.ModifyDate = constant.MIN_TIMESTAMP * 1000
	}
	detail.TenantId = record.TenantId
	detail.ProductId = record.ProductId
	detail.CreateUserId = record.CreateUserId
	detail.ModelStatus = record.ModelStatus
	detail.Name = record.Name
	detail.ParentId = 0
	if true == record.ParentId.Valid {
		detail.ParentId = record.ParentId.Int64
	}
	detail.Flag = record.Flag
	detail.Comment = record.Comment
	detail.Type = record.Type
	detail.Children = make([]RightDetail, 0)
	return &detail, nil
}

//mapChildrenRight, 有parent的Right信息，其中key为parentId
//depth，递归深度，避免特殊错误数据递归调用死循环
func addChildrenFoRightDetail(detail *RightDetail, mapChildrenRight map[int64][]RightDetail, depth int) {
	if nil == detail || nil == mapChildrenRight || depth > 100 { //最大递归深度100，避免特殊异常数据死循环
		return
	}
	childRights, ok := mapChildrenRight[detail.Id] //找到直接children
	if ok && nil != childRights {
		detail.Children = childRights
	}
	for index, _ := range detail.Children {
		addChildrenFoRightDetail(&detail.Children[index], mapChildrenRight, depth+1) //对每个child去寻找下一级child
	}
}

func extractRightNameByLanguage(name string, language string) string {
	if "" == name {
		return name
	}
	strList := strings.Split(name, "/")
	switch language {
	case "en-US":
		if len(strList) > 1 {
			return strList[1]
		} else {
			return strList[0]
		}
	case "zh-HK":
	case "zh-TW":
		if len(strList) > 3 {
			return strList[3]
		} else {
			return strList[0]
		}
	case "ja-JP":
		if len(strList) > 4 {
			return strList[4]
		} else {
			return strList[0]
		}
	case "ko-KR":
		if len(strList) > 5 {
			return strList[5]
		} else {
			return strList[0]
		}
	case "id-ID":
		if len(strList) > 6 {
			return strList[6]
		} else {
			return strList[0]
		}
	case "vi-VN":
		if len(strList) > 7 {
			return strList[7]
		} else {
			return strList[0]
		}
	default:
		return strList[0]
	}
	return name
}

func ModifyChildrenLanguageForRightDetail(detail *RightDetail, language string) {
	if nil == detail {
		return
	}
	for index, _ := range detail.Children {
		detail.Children[index].Name = extractRightNameByLanguage(detail.Children[index].Name, language)
		ModifyChildrenLanguageForRightDetail(&detail.Children[index], language) //对每个child去修改下一级child
	}
	return
}

func ProcessListTopRights(feignKey *XFeignKey) ([]RightDetail, code.Error) {
	tenant_id := feignKey.TenantId
	if "" == tenant_id {
		return nil, errcode.CerrParamater
	}
	dbConn, errConn := GetDBConnByTenantID(tenant_id)
	if errConn != nil {
		return nil, errConn
	}

	var rightRecords []RightTab = make([]RightTab, 0)
	dbtmp := dbConn.Where("model_status != ?", constant.ModelStatusDelete).Find(&rightRecords)
	if err := dbtmp.Error; nil != err {
		if err == gorm.ErrRecordNotFound {
			logrus.WithFields(logrus.Fields{}).Warn("Not found right record from DB")
			return nil, nil //未找到记录不返回错误
		}
		logrus.WithFields(logrus.Fields{"error": err}).Error("Failed to get right record from DB")
		return nil, errcode.CerrExecuteSQL
	}

	var noParentRights []RightDetail = make([]RightDetail, 0)                       //无parent的right信息
	var mapChildrenRight map[int64][]RightDetail = make(map[int64][]RightDetail, 0) //有parent的Right信息，其中key为parentId
	for _, record := range rightRecords {
		detail, errConv := RightTab2RightDetail(&record)
		if nil != errConv || nil == detail {
			logrus.WithFields(logrus.Fields{"error": errConv}).Error("Failed to convert right info")
			continue
		}
		if detail.ParentId > 0 { //添加到mapChildrenRight中
			chilDetails, ok := mapChildrenRight[detail.ParentId]
			if false == ok || nil == chilDetails {
				chilDetails = make([]RightDetail, 0)
			}
			chilDetails = append(chilDetails, *detail)
			mapChildrenRight[detail.ParentId] = chilDetails
		} else { //添加到noParentRight中
			noParentRights = append(noParentRights, *detail)
		}
	}

	//添加child right信息
	for index, _ := range noParentRights {
		addChildrenFoRightDetail(&noParentRights[index], mapChildrenRight, 1)
	}

	//处理language
	for index, _ := range noParentRights {
		noParentRights[index].Name = extractRightNameByLanguage(noParentRights[index].Name, feignKey.Language)
		ModifyChildrenLanguageForRightDetail(&noParentRights[index], feignKey.Language)
	}

	return noParentRights, nil
}

func GetCurrentUserRight(feignKey *XFeignKey) ([]string, code.Error) {
	tenant_id := feignKey.TenantId
	if "" == tenant_id {
		return nil, errcode.CerrParamater
	}

	//获取当前登录用户信息
	userInfo, errGetUser := GetUserDetailByIdOrPubid(feignKey, feignKey.UserId, 0)
	if nil != errGetUser {
		logrus.WithFields(logrus.Fields{"error": errGetUser, "pubId": feignKey.UserId}).Error("Get current user info failed")
		return nil, errGetUser
	}
	if nil == userInfo {
		logrus.WithFields(logrus.Fields{"pubID": feignKey.UserId}).Error("Not found user info, pubID:%v", feignKey.UserId)
		return nil, errcode.CerrAccountNotFound
	}

	rightrecords, errGetOneRoleRight := GetOneRoleRightList(feignKey, userInfo.RoleId)
	if nil != errGetOneRoleRight {
		logrus.WithFields(logrus.Fields{"pubID": feignKey.UserId, "roleID": userInfo.RoleId, "error": errGetOneRoleRight}).Error("Get one user right list failed")
		return nil, errGetOneRoleRight
	}

	var rightStrList []string = make([]string, 0)
	for index, _ := range rightrecords {
		rightStrList = append(rightStrList, rightrecords[index].EntityNo)
	}
	return rightStrList, nil
}

func GetRoleRight(feignKey *XFeignKey, roleId int64) ([]string, code.Error) {
	tenant_id := feignKey.TenantId
	if "" == tenant_id {
		return nil, errcode.CerrParamater
	}

	rightrecords, errGetOneRoleRight := GetOneRoleRightList(feignKey, roleId)
	if nil != errGetOneRoleRight {
		logrus.WithFields(logrus.Fields{"pubID": feignKey.UserId, "roleID": roleId, "error": errGetOneRoleRight}).Error("Get one user right list failed")
		return nil, errGetOneRoleRight
	}

	var rightStrList []string = make([]string, 0)
	for index, _ := range rightrecords {
		rightStrList = append(rightStrList, rightrecords[index].EntityNo)
	}
	return rightStrList, nil
}

func createRoleName(language string) string {
	//roleName := "Admin"
	//logrus.WithFields(logrus.Fields{"language": language, "roleName":roleName}).Debug("createRoleName")
	return "Admin"
}

func createCustomRoleName(language string, roleName []string) string {
	if "zh-CN" == language {
		return roleName[0]
	} else if "zh-TW" == language {
		return roleName[1]
	} else if "en-US" == language {
		return roleName[2]
	} else if "ja-JP" == language {
		return roleName[3]
	} else if "ko-KR" == language {
		return roleName[4]
	} else {
		return roleName[2]
	}
}

//根据用户Right权限，提取用户的Normalized权限信息，返回Normalized权限信息的切片
func normalizeUserTreeRight(userRights []RightTab, module string) ([]int, code.Error) {
	var normalizedList = make([]int, 0)

	switch module {
	case constant.ModuleUser:
		normalizedList = append(normalizedList, constant.NORMALIZED_RIGHT_MY)
		if RightsHaveCotain(userRights, constant.AUTHORITY_USER_SELECT_ALL) {
			normalizedList = append(normalizedList, constant.NORMALIZED_RIGHT_ALL)
		}
		if RightsHaveCotain(userRights, constant.AUTHORITY_USER_SELECT_WILD) {
			normalizedList = append(normalizedList, constant.NORMALIZED_RIGHT_NO_PARENT)
		}
		if RightsHaveCotain(userRights, constant.AUTHORITY_USER_SELECT_DIRECTLY) {
			normalizedList = append(normalizedList, constant.NORMALIZED_RIGHT_DIRECT)
		}
		if RightsHaveCotain(userRights, constant.AUTHORITY_USER_SELECT_SUBORDINATE) {
			normalizedList = append(normalizedList, constant.NORMALIZED_RIGHT_SUBORDINATE)
		}
	case constant.ModuleAccount:
		if RightsHaveCotain(userRights, constant.AUTHORITY_ACCOUNT_SELECT_ALL) {
			normalizedList = append(normalizedList, constant.NORMALIZED_RIGHT_ALL)
		}
		if RightsHaveCotain(userRights, constant.AUTHORITY_ACCOUNT_SELECT_DIRECTLY) {
			normalizedList = append(normalizedList, constant.NORMALIZED_RIGHT_MY)
		}
		if RightsHaveCotain(userRights, constant.AUTHORITY_ACCOUNT_SELECT_SUBORDINATE) {
			normalizedList = append(normalizedList, constant.NORMALIZED_RIGHT_DIRECT, constant.NORMALIZED_RIGHT_SUBORDINATE)
		}
	case constant.ModuleAccountReport:
		if RightsHaveCotain(userRights, constant.AUTHORITY_ACCOUNT_REPORT_SELECT_ALL) {
			normalizedList = append(normalizedList, constant.NORMALIZED_RIGHT_ALL)
		}
		if RightsHaveCotain(userRights, constant.AUTHORITY_ACCOUNT_REPORT_SELECT_MY) {
			normalizedList = append(normalizedList, constant.NORMALIZED_RIGHT_MY)
		}
		if RightsHaveCotain(userRights, constant.AUTHORITY_ACCOUNT_REPORT_SELECT_SUB) {
			normalizedList = append(normalizedList, constant.NORMALIZED_RIGHT_DIRECT, constant.NORMALIZED_RIGHT_SUBORDINATE)
		}
	case constant.ModuleCommissionReport:
		normalizedList = append(normalizedList, constant.NORMALIZED_RIGHT_MY)
		if RightsHaveCotain(userRights, constant.AUTHORITY_COMMISSION_REPORT_SELECT_ALL) {
			normalizedList = append(normalizedList, constant.NORMALIZED_RIGHT_ALL)
		}
		if RightsHaveCotain(userRights, constant.AUTHORITY_COMMISSION_REPORT_SELECT_SUB) {
			normalizedList = append(normalizedList, constant.NORMALIZED_RIGHT_DIRECT, constant.NORMALIZED_RIGHT_SUBORDINATE)
		}
	case constant.ModuleEarningReport:
		normalizedList = append(normalizedList, constant.NORMALIZED_RIGHT_MY)
		if RightsHaveCotain(userRights, constant.AUTHORITY_EARNING_REPORT_SELECT_ALL) {
			normalizedList = append(normalizedList, constant.NORMALIZED_RIGHT_ALL)
		}
		if RightsHaveCotain(userRights, constant.AUTHORITY_EARNING_REPORT_SELECT_WILD) {
			normalizedList = append(normalizedList, constant.NORMALIZED_RIGHT_NO_PARENT)
		}
		if RightsHaveCotain(userRights, constant.AUTHORITY_EARNING_REPORT_SELECT_DIRECT) {
			normalizedList = append(normalizedList, constant.NORMALIZED_RIGHT_DIRECT)
		}
		if RightsHaveCotain(userRights, constant.AUTHORITY_EARNING_REPORT_SELECT_INDIRECT) {
			normalizedList = append(normalizedList, constant.NORMALIZED_RIGHT_SUBORDINATE)
		}
	case constant.ModuleCustomer:
		if RightsHaveCotain(userRights, constant.AUTHORITY_CUSTOMER_SELECT_ALL) {
			normalizedList = append(normalizedList, constant.NORMALIZED_RIGHT_ALL)
		}
		if RightsHaveCotain(userRights, constant.AUTHORITY_CUSTOMER_SELECT_SELECT_MY) {
			normalizedList = append(normalizedList, constant.NORMALIZED_RIGHT_MY)
		}
		if RightsHaveCotain(userRights, constant.AUTHORITY_CUSTOMER_SELECT_SUB) {
			normalizedList = append(normalizedList, constant.NORMALIZED_RIGHT_DIRECT, constant.NORMALIZED_RIGHT_SUBORDINATE)
		}
	default:
		logrus.WithFields(logrus.Fields{"module": module}).Error("normalizeUserTreeRight parameter error")
		return nil, errcode.CerrParamater
	}
	return normalizedList, nil
}

func parseTenantPermission(parentNode *PermissionNode, node *PermissionNode) []RightTab {
	var rights []RightTab = make([]RightTab, 0)
	var right RightTab
	right.EntityNo = strings.ToUpper(node.AuthNo)
	right.Flag = node.Flag
	right.Name = node.AuthName
	right.Dependence = strings.ToUpper(node.Dependence)
	if strings.Contains(node.AuthNo, "_") {
		types := strings.Split(node.AuthNo, "_")
		right.Type = strings.ToUpper(strings.Replace(node.AuthNo, "_"+types[len(types)-1], "", -1))
	} else {
		right.Type = node.AuthNo
	}

	if nil != parentNode {
		right.Comment = strings.ToUpper(parentNode.AuthNo) //暂时把parent的EntityNo存入comment中，后面要转换成parent_id
	}
	rights = append(rights, right)

	for index, _ := range node.Children {
		rightSub := parseTenantPermission(node, &node.Children[index])
		rights = append(rights, rightSub...)
	}
	return rights
}

func tenantPermissionDTOList2RightTabList(permissionList []TenantPermissionDTO) []RightTab {
	var rights []RightTab = make([]RightTab, 0)
	for index, _ := range permissionList {
		for _, node := range permissionList[index].Permissions {
			node.AuthNo = strings.ToUpper(permissionList[index].ModuleCode)
			node.AuthName = permissionList[index].ModuleName
			rightSub := parseTenantPermission(nil, &node)
			rights = append(rights, rightSub...)
		}
	}
	return rights
}

//初始化权限树
func ProcessInitRight(feignKey *XFeignKey) code.Error {
	if nil == feignKey || "" == feignKey.TenantId {
		return errcode.CerrParamater
	}

	dbRights, errGetDbRight := GeAllRightRecords(feignKey.TenantId)
	if nil != errGetDbRight {
		logrus.WithFields(logrus.Fields{"error": errGetDbRight}).Error("GeAllRightRecords error")
		return errGetDbRight
	}
	//从外部获取租户的产品权限信息
	allProductPermission, errPermission := ClientProductPermission(feignKey, feignKey.TenantId, feignKey.ProductId, "zh-CN")
	if nil != errPermission {
		logrus.WithFields(logrus.Fields{"error": errPermission}).Error("ClientProductPermission error")
		return errPermission
	}
	allProductRights := tenantPermissionDTOList2RightTabList(allProductPermission) //结构转换

	var mapRights map[string]RightTab = make(map[string]RightTab) //key为entityNo，用户记录有有效的db right记录
	//计算需要从DB中删除的rightsId
	deleteRightIds := make([]int64, 0)
	for index, _ := range dbRights {
		if false == RightsHaveCotain(allProductRights, dbRights[index].EntityNo) {
			deleteRightIds = append(deleteRightIds, dbRights[index].Id)
		} else {
			mapRights[dbRights[index].EntityNo] = dbRights[index] //记录有用的db right记录
		}
	}
	logrus.WithFields(logrus.Fields{"deleteRightIds": deleteRightIds}).Info("right record will delete")

	dbConn, errConn := GetDBConnByTenantID(feignKey.TenantId)
	if errConn != nil {
		return errConn
	}
	tx := dbConn.Begin() //开始transaction，后续所有的写操作均使用本transaction

	//DB中多余的权限删除，删除上下级关系（批量操作）
	if len(deleteRightIds) > 0 {
		if err := tx.Exec("UPDATE t_right SET parent_id = NULL WHERE parent_id IN (?)", deleteRightIds).Error; nil != err {
			logrus.WithFields(logrus.Fields{"error": err, "rightId": deleteRightIds}).Error("reset parent_id error")
			tx.Rollback()
			return errcode.CerrExecuteSQL
		}
		//删除role right relation（批量操作）
		if err := tx.Exec("DELETE FROM t_role_right_relation WHERE role_rights IN (?)", deleteRightIds).Error; nil != err {
			logrus.WithFields(logrus.Fields{"error": err, "rightId": deleteRightIds}).Error("delete from t_role_right_relation error")
			tx.Rollback()
			return errcode.CerrExecuteSQL
		}
		//删除right记录（批量操作）
		if err := tx.Exec(" DELETE FROM t_right WHERE id IN (?)", deleteRightIds).Error; nil != err {
			logrus.WithFields(logrus.Fields{"error": err, "rightId": deleteRightIds}).Error("delete from t_right error")
			tx.Rollback()
			return errcode.CerrExecuteSQL
		}
	}

	//新添加权限记录
	var newRights []RightTab
	for index, _ := range allProductRights {
		newRight := allProductRights[index]
		if false == RightsHaveCotain(dbRights, newRight.EntityNo) { //db中还不存在，添加记录
			rightRecord := allProductRights[index]
			rightRecord.CreateDate = time.Now()
			rightRecord.ModifyDate = rightRecord.CreateDate
			rightRecord.ModelStatus = constant.ModelStatusCreate
			rightRecord.Comment = "" //allProductRights[index]中的comment暂时存放有parent的entityNo
			if err := tx.Create(&rightRecord).Error; nil != err {
				logrus.WithFields(logrus.Fields{"error": err, "entityNo": rightRecord.EntityNo}).Error("add right record error")
				tx.Rollback()
				return errcode.CerrExecuteSQL
			}
			mapRights[rightRecord.EntityNo] = rightRecord //记录有用的db right记录
			logrus.WithFields(logrus.Fields{"entityNo": rightRecord.EntityNo}).Info("add right record success, waite for submit")

			//添加至新增权限列表
			newRight.Id = rightRecord.Id
			newRights = append(newRights, newRight)
		}
	}

	//更新权限字段和上下级关系
	for index, _ := range allProductRights {
		record, exist := mapRights[allProductRights[index].EntityNo]
		if false == exist {
			logrus.WithFields(logrus.Fields{"entityNo": allProductRights[index].EntityNo}).Warn("mapRights not this right record")
			continue
		}
		if record.CreateDate.Unix() < constant.MIN_TIMESTAMP || record.CreateDate.Unix() > constant.MAX_TIMESTAMP {
			record.CreateDate = time.Unix(constant.MIN_TIMESTAMP, 0)
		}
		record.Name = allProductRights[index].Name
		record.ModifyDate = time.Now()
		record.ParentId.Valid = false
		if "" != allProductRights[index].Comment { //allProductRights[index]中的comment暂时存放有parent的entityNo
			if tmpRecord, tmpExist := mapRights[allProductRights[index].Comment]; true == tmpExist {
				record.ParentId.Int64 = tmpRecord.Id
				record.ParentId.Valid = true
			}
		}
		if err := tx.Save(&record).Error; nil != err {
			logrus.WithFields(logrus.Fields{"error": err, "entityNo": record.EntityNo, "id": record.Id}).Error("update right record error")
			tx.Rollback()
			return errcode.CerrExecuteSQL
		}
	}

	//重置AdminUser的role权限（roleId为1，有所有的权限）
	mapRightsSize := len(mapRights)
	if mapRightsSize > 0 {
		roleInfo, errGetRole := GetRoleInfoByID(feignKey.TenantId, 1)
		if nil == errGetRole && nil != roleInfo { //roleId为1的角色存在
			if err := tx.Exec(" DELETE FROM t_role_right_relation WHERE t_role = 1").Error; nil != err {
				logrus.WithFields(logrus.Fields{"error": err, "t_role": 1}).Error("delete from t_role_right_relation error")
				tx.Rollback()
				return errcode.CerrExecuteSQL
			}
			insertRelationSql := "INSERT INTO t_role_right_relation(t_role,role_rights) VALUES" //批量插入sql
			counter := 0
			for _, record := range mapRights {
				if counter < mapRightsSize-1 {
					insertRelationSql += fmt.Sprintf("(1, %v),", record.Id)
				} else {
					insertRelationSql += fmt.Sprintf("(1, %v)", record.Id) //无逗号
				}
				counter++
			}
			if err := tx.Exec(insertRelationSql).Error; nil != err {
				logrus.WithFields(logrus.Fields{"error": err, "insertRelationSql": insertRelationSql}).Error("insert t_role_right_relation error")
				tx.Rollback()
				return errcode.CerrExecuteSQL
			}
			logrus.WithFields(logrus.Fields{"insertRelationSql": insertRelationSql}).Debug("Insert t_role_right_relation success")
		}
	}

	//事务提交
	if err := tx.Commit().Error; nil != err {
		logrus.WithFields(logrus.Fields{"error": err}).Error("initRight commit transaction error")
		return errcode.CerrExecuteSQL
	}

	//刷新角色新增权限的勾选状态
	go refreshAllRoleNewRight(feignKey, newRights)

	return nil
}

func refreshAllRoleNewRight(feignKey *XFeignKey, newRights []RightTab) {
	if len(newRights) == 0 {
		return
	}

	allRoles, allRolesErr := GetAllRoleLis(feignKey.TenantId)
	if allRolesErr != nil {
		logrus.WithFields(logrus.Fields{"tenant": feignKey.TenantId, "error": allRolesErr}).Error("Failed to GetAllRoleLis")
		return
	}

	allRight, allRightErr := GeAllRightRecords(feignKey.TenantId)
	if allRightErr != nil {
		logrus.WithFields(logrus.Fields{"tenant": feignKey.TenantId, "error": allRightErr}).Error("Failed to GeAllRightRecords")
		return
	}

	allRightMap := RightTabList2RightTabMap(allRight)

	for index := range allRoles {
		role := allRoles[index]
		rights, rightErr := GetOneRoleRightList(feignKey, role.Id)
		if rightErr != nil {
			logrus.WithFields(logrus.Fields{"tenant": feignKey.TenantId, "role": role.Id, "error": rightErr}).Error("Failed to GetOneRoleRightList")
			continue
		}

		oldRightMap := RightTabList2RightCodeTabMap(rights)

		for newRightIndex := range newRights {
			//为了信息准确，用数据库的信息
			newRight := allRightMap[newRights[newRightIndex].Id]

			dependence := newRights[newRightIndex].Dependence

			if _, contain := oldRightMap[newRight.EntityNo]; contain {
				continue
			}

			switch dependence {
			case "OWN":
				addRoleRight(feignKey.TenantId, &role, &newRight)
				break
			case "PARENT":
				parent, hasParent := allRightMap[newRight.ParentId.Int64]
				if hasParent {
					_, hasParentRight := oldRightMap[parent.EntityNo]
					if hasParentRight {
						addRoleRight(feignKey.TenantId, &role, &newRight)
					}
				}
				break
			default:
				if dependence != "" {
					if _, hasDependenceRight := oldRightMap[dependence]; hasDependenceRight {
						addRoleRight(feignKey.TenantId, &role, &newRight)
					}
				}
				break
			}
		}

	}
}

func addRoleRight(tenantId string, role *RoleTab, right *RightTab) {
	dbConn, errConn := GetDBConnByTenantID(tenantId)
	if errConn != nil {
		logrus.WithFields(logrus.Fields{"tenant": tenantId, "error": errConn}).Error("Failed to GetDBConnByTenantID")
		return
	}

	db := dbConn.Exec("INSERT INTO t_role_right_relation(t_role, role_rights) VALUES (?, ?);", role.Id, right.Id)
	if db.Error != nil {
		logrus.WithFields(logrus.Fields{"tenant": tenantId, "roleId": role.Id, "right": right.Id, "error": db.Error}).Error("Failed to INSERT role right")
	}
}

func addRoleRightV2(tenantId string, roleId int64, rightId int64) {
	logrus.WithFields(logrus.Fields{"tenant": tenantId, "roleId": roleId, "right": rightId}).Info("INSERT role right")
	dbConn, errConn := GetDBConnByTenantID(tenantId)
	if errConn != nil {
		logrus.WithFields(logrus.Fields{"tenant": tenantId, "error": errConn}).Error("Failed to GetDBConnByTenantID")
		return
	}
	db := dbConn.Exec("INSERT INTO t_role_right_relation(t_role, role_rights) VALUES (?, ?);", roleId, rightId)
	if db.Error != nil {
		logrus.WithFields(logrus.Fields{"tenant": tenantId, "roleId": roleId, "right": rightId, "error": db.Error}).Error("Failed to INSERT role right")
	}
}

func GetRightByEntityNos(entityNos []string, tenantId string) ([]RightTab, code.Error) {
	dbConn, errConn := GetDBConnByTenantID(tenantId)
	if errConn != nil {
		logrus.WithFields(logrus.Fields{"tenant": tenantId, "error": errConn}).Error("Failed to GetDBConnByTenantID")
		return nil, errConn
	}
	rightList := make([]RightTab, 0)
	where := "entity_no in ("
	for _, item := range entityNos {
		where += "'" + item + "',"
	}
	where = strings.TrimRight(where, ",") + ")"
	dbtmp := dbConn.Where(where).Find(&rightList)
	if err := dbtmp.Error; nil != err {
		if err == gorm.ErrRecordNotFound {
			logrus.Warnf("Not found right record from DB, entityNos:%v, tenantId:%v", entityNos, tenantId)
			return rightList, nil //未找到记录不返回错误
		}
		logrus.Errorf("Failed to get right record from DB, entityNos:%v, tenantId:%v, error:%v", entityNos, tenantId, err)
		return nil, errcode.CerrExecuteSQL
	}
	return rightList, nil
}

func GetRoleIdListByRightEntityNo(tenantId string, rightEntityNo string) ([]int64, code.Error) {
	rightRecords, errGetRight := GeRightRecordsByEntityno(tenantId, []string{rightEntityNo})
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
	relationRecords, errGetRelation := GetRoleRightRelationByRightIds(tenantId, rightIds)
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
	return roleIds, nil
}

func getAlreadyRightRoleSet(feignKey *XFeignKey, keyIdRightDTO KeyIdRightDTO) (map[string]bool, code.Error) {
	//1.获取拥有该权限的角色Ids
	roleIds, err := GetRoleIdListByRightEntityNo(feignKey.TenantId, keyIdRightDTO.RightEntityNo)
	if nil != err {
		logrus.WithFields(logrus.Fields{"error": err, "TenantId": feignKey.TenantId, "rightEntityNo": keyIdRightDTO.RightEntityNo}).Error("GetRoleIdListByRightEntityNo Failed!")
		return nil, err
	}
	//2.转换为map
	roleRightMap := make(map[string]bool)
	for _, roleId := range roleIds {
		roleRightMap[strconv.FormatInt(roleId, 10)] = true
	}
	return roleRightMap, nil
}

func CheckIdsRight(feignKey *XFeignKey, keyIdRightDTO KeyIdRightDTO, outRoleId bool) (map[string]bool, code.Error) {
	//1.获取拥有权限的决定Id
	roleRightMap, err := getAlreadyRightRoleSet(feignKey, keyIdRightDTO)
	if err != nil {
		logrus.WithFields(logrus.Fields{"error": err, "TenantId": feignKey.TenantId, "rightEntityNo": keyIdRightDTO.RightEntityNo}).Error("getAlreadyRightRoleSet Failed!")
	}
	idCheckMap := make(map[string]bool)
	switch keyIdRightDTO.KeyType {
	case constant.QueryKeypubUserIds:
		//2.获取用户角色id
		pubIdToRoleIdMap, err := GetUserIdToRoleIdMap(feignKey, keyIdRightDTO.Ids, constant.QueryKeypubUserIds)
		if nil != err {
			logrus.WithFields(logrus.Fields{"error": err, "TenantId": feignKey.TenantId, "rightEntityNo": keyIdRightDTO.RightEntityNo}).Error("ProcessListUserRecordsByKey Failed")
			return nil, err
		}
		//3.权限判断结果
		if outRoleId == false {
			for _, id := range keyIdRightDTO.Ids {
				roleId, ok := pubIdToRoleIdMap[id]
				if ok {
					_, ok1 := roleRightMap[roleId]
					idCheckMap[id] = ok1
				} else {
					logrus.WithFields(logrus.Fields{"error": err, "TenantId": feignKey.TenantId, "rightEntityNo": keyIdRightDTO.RightEntityNo}).Error("pubUserId No RoleId!", id)
					idCheckMap[id] = false
				}
			}
		} else {

			for _, roleId := range pubIdToRoleIdMap {
				idCheckMap[roleId] = false
				_, ok := roleRightMap[roleId]
				if ok {
					idCheckMap[roleId] = true
				} else if keyIdRightDTO.RightEntityNo == "TASK_TRADER_SET" || keyIdRightDTO.RightEntityNo == "TASK_IB_SET" { //判断这个角色所有人员是否都在待勾选列表中,这个代码很恶心，测试好了不要动
					roleIdInt64, _ := strconv.ParseInt(roleId, 10, 64)
					userList, errGetUser := ProcessListUserRecordsByKey(feignKey, []int64{roleIdInt64}, constant.QueryKeyRoleId)
					if nil != errGetUser {
						logrus.WithFields(logrus.Fields{"error": errGetUser, "roleId": roleId}).Error("ProcessListUserRecordsByKey failed")
						return nil, errGetUser
					}
					for _, userInfo := range userList {
						_, isExist := pubIdToRoleIdMap[userInfo.PubUserId]
						//用户列表不是全部包含,就不添加这个角色的权限
						if isExist == false {
							logrus.WithFields(logrus.Fields{"info": userInfo.PubUserId, "roleId": roleId}).Error("TaskSetUserList Not All Role User ")
							idCheckMap[roleId] = true
						}
					}
				}
			}
		}
		break
	case constant.QueryKeypubUserIdsIB:
		//2.获取用户角色id
		pubIdToRoleIdMap, err := GetUserIdToRoleIdMap(feignKey, keyIdRightDTO.Ids, constant.QueryKeypubUserIds)
		if nil != err {
			logrus.WithFields(logrus.Fields{"error": err, "TenantId": feignKey.TenantId}).Error("ProcessListUserRecordsByKey Failed")
			return nil, err
		}
		roleIdStr := "-2838383" //添加一个不可能存在的ID，避免传空集合
		for _, roleId := range pubIdToRoleIdMap {
			roleIdStr += fmt.Sprintf(",%v", roleId)
		}
		advanceCondition := &AdvanceCondition{
			Condition: "",
			Field:     constant.SearchField_Id,
			Value:     roleIdStr,
		}
		searchReq := &SearchDTO{}
		searchReq.Conditions = append(searchReq.Conditions, advanceCondition)
		roleTabList, err := SearchRoleByCondition(feignKey, searchReq)
		if nil != err {
			logrus.WithFields(logrus.Fields{"error": err, "TenantId": feignKey.TenantId}).Error("SearchRoleByCondition Failed")
			return nil, err
		}
		roleMap := make(map[string]bool)
		for _, roleTab := range roleTabList {
			isIb := true
			if roleTab.RoleTypeId == "ib" {
				isIb = false
			}
			roleMap[strconv.FormatInt(roleTab.Id, 10)] = isIb
		}
		for pubId, roleId := range pubIdToRoleIdMap {
			idCheckMap[pubId] = roleMap[roleId]
		}
		break
	case constant.QueryKeyRoleId:
		//4.权限判断结果
		for _, id := range keyIdRightDTO.Ids {
			_, ok := roleRightMap[id]
			idCheckMap[id] = ok
		}
		break
	}
	return idCheckMap, nil
}

//	临时性手动添加父权限
func getPreRightId(feignKey *XFeignKey, keyIdRightDTO KeyIdRightDTO) ([]int64, code.Error) {
	var rightEntityNoPre string
	if strings.Contains(keyIdRightDTO.RightEntityNo, "TASK_TRADER") {
		rightEntityNoPre = "TASK_TRADER"
	} else {
		rightEntityNoPre = "TASK_IB"
	}
	rightTabList, err := GeRightRecordsByEntityno(feignKey.TenantId, []string{rightEntityNoPre, "TASK"})
	if nil != err {
		logrus.WithFields(logrus.Fields{"error": err, "TenantId": feignKey.TenantId, "rightEntityNo": keyIdRightDTO.RightEntityNo}).Error("GeRightRecordsByEntityno Failed")
		return nil, err
	}
	ids := make([]int64, 0)
	for _, data := range rightTabList {
		ids = append(ids, data.Id)
	}
	return ids, nil
}

func AddIdsRight(feignKey *XFeignKey, keyIdRightDTO KeyIdRightDTO) code.Error {
	//1.获取权限id
	rightTabList, err := GeRightRecordsByEntityno(feignKey.TenantId, []string{keyIdRightDTO.RightEntityNo})
	if nil != err {
		logrus.WithFields(logrus.Fields{"error": err, "TenantId": feignKey.TenantId, "rightEntityNo": keyIdRightDTO.RightEntityNo}).Error("GeRightRecordsByEntityno Failed")
		return err
	}
	if len(rightTabList) == 0 {
		logrus.WithFields(logrus.Fields{"error": err, "TenantId": feignKey.TenantId, "rightEntityNo": keyIdRightDTO.RightEntityNo}).Error("GeRightRecordsByEntityno Is empty")
		return errcode.CerrExecuteSQL
	}
	rightId := rightTabList[0].Id

	//2.获取已经拥有权限的ids
	idCheckMap, err := CheckIdsRight(feignKey, keyIdRightDTO, true)
	if nil != err {
		logrus.WithFields(logrus.Fields{"error": err, "TenantId": feignKey.TenantId, "rightEntityNo": keyIdRightDTO.RightEntityNo}).Error("CheckIdsRight Failed")
		return err
	}
	//3.删除已经拥有权限的ids
	for id, isHaveRight := range idCheckMap {
		if !isHaveRight {
			roleId, err := strconv.ParseInt(id, 10, 64)
			if nil != err {
				logrus.WithFields(logrus.Fields{"error": err, "TenantId": feignKey.TenantId, "rightEntityNo": keyIdRightDTO.RightEntityNo}).Error(" strconv.ParseInt(id, 10, 64) Failed")
				return errcode.CerrParamater
			}

			addRoleRightV2(feignKey.TenantId, roleId, rightId)

			//添加父权限
			preRightIds, err := getPreRightId(feignKey, keyIdRightDTO)
			if nil != err {
				logrus.WithFields(logrus.Fields{"error": err, "TenantId": feignKey.TenantId, "rightEntityNo": keyIdRightDTO.RightEntityNo}).Error(" strconv.ParseInt(id, 10, 64) Failed")
				return errcode.CerrParamater
			}
			for _, preRightId := range preRightIds {
				addRoleRightV2(feignKey.TenantId, roleId, preRightId)
			}
		}

	}
	return nil
}

func AddRightDependOtherRightForRole(feignKey *XFeignKey, depend string, right string) code.Error {
	rightRecords, errGetRight := GeRightRecordsByEntityno(feignKey.TenantId, []string{depend, right})
	if nil != errGetRight {
		return errGetRight
	}
	if len(rightRecords) != 2 {
		return errcode.CerrRightNotExist
	}
	var dependId int64 = 0
	var rightId int64 = 0
	for _, rightRecord := range rightRecords {
		if rightRecord.EntityNo == depend {
			dependId = rightRecord.Id
		} else if rightRecord.EntityNo == right {
			rightId = rightRecord.Id
		}
	}
	relationRecords, errGetRelation := GetRoleRightRelationByRightIds(feignKey.TenantId, []int64{dependId, rightId})
	if nil != errGetRelation {
		return errGetRelation
	}
	mapRelation := make(map[int64][]int64)
	for _, relation := range relationRecords {
		childs, exist := mapRelation[relation.RoleID]
		if false == exist {
			childs = make([]int64, 0)
		}
		childs = append(childs, relation.RightID)
		mapRelation[relation.RoleID] = childs
	}
	for k, v := range mapRelation {
		if len(v) == 1 && v[0] == dependId {
			addRoleRightV2(feignKey.TenantId, k, rightId)
		}
	}
	return nil
}
