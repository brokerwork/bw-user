package model

import (
	"bw/bw-user/constant"
	"bw/bw-user/errcode"
	"github.com/Sirupsen/logrus"
	"github.com/jinzhu/gorm"
	"github.com/lworkltd/kits/service/restful/code"
	"sort"
	"time"
)

//根据roleID获取role信息
func GetLevelInfoByID(tenant_id string, id int64) (*LevelTab, code.Error) {
	if "" == tenant_id {
		return nil, errcode.CerrParamater
	}
	dbConn, errConn := GetDBConnByTenantID(tenant_id)
	if errConn != nil {
		return nil, errConn
	}

	var levelInfo LevelTab
	dbtmp := dbConn.Where("id = ? and model_status != ?", id, constant.ModelStatusDelete).First(&levelInfo)

	if err := dbtmp.Error; nil != err {
		if err == gorm.ErrRecordNotFound {
			logrus.WithFields(logrus.Fields{
				"id": id,
			}).Warn("Not found level record from DB by id")
			return nil, nil
		}
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Error("Failed to get level record from DB by ids")
		return nil, errcode.CerrExecuteSQL
	}
	return &levelInfo, nil
}

//获取所有Level信息
func ProcessGetLevelList(feignKey *XFeignKey) ([]LevelDetail, code.Error) {
	tenant_id := feignKey.TenantId
	if "" == tenant_id {
		return nil, errcode.CerrParamater
	}
	dbConn, errConn := GetDBConnByTenantID(tenant_id)
	if errConn != nil {
		return nil, errConn
	}

	var levelRecords []LevelTab = make([]LevelTab, 0)
	dbtmp := dbConn.Where("model_status != ?", constant.ModelStatusDelete).Order("sid asc").Find(&levelRecords) //按sid排序
	if err := dbtmp.Error; nil != err {
		if err == gorm.ErrRecordNotFound {
			logrus.Warnf("Not found level record from DB")
			return nil, nil //未找到记录不返回错误
		}
		logrus.Errorf("Failed to get level record from DB, error:%v", err)
		return nil, errcode.CerrExecuteSQL
	}

	var levelDetails []LevelDetail = make([]LevelDetail, 0)
	for _, record := range levelRecords {
		detail, errConvert := ConvertLevelTal2LevelDetail(&record)
		if nil != errConvert || nil == detail {
			return nil, errcode.CerrInternal
		}
		levelDetails = append(levelDetails, *detail)
	}
	return levelDetails, nil
}

func ProcessAddlevel(feignKey *XFeignKey, addReq *LevelDetail) code.Error {
	if nil == feignKey || "" == feignKey.TenantId || nil == addReq {
		return errcode.CerrParamater
	}
	var levelTab LevelTab
	levelTab.Name = addReq.Name
	levelTab.Sid = addReq.Sid
	levelTab.UserCount = 0
	return addLevel(feignKey, &levelTab)
}

func ProcessDeletelevel(feignKey *XFeignKey, levelId int64) code.Error {
	if nil == feignKey || "" == feignKey.TenantId || levelId <= 0 {
		return errcode.CerrParamater
	}
	//判断是否存在用户
	userRecords, errGet := ProcessListUserRecordsByKey(feignKey, []int64{levelId}, constant.QueryKeyLevelId)
	if nil != errGet {
		logrus.WithFields(logrus.Fields{"error": errGet, "levelId": levelId}).Error("get user list by levelId failed")
		return errGet
	}
	if len(userRecords) > 0 {
		logrus.WithFields(logrus.Fields{"user count": len(userRecords), "levelId": levelId}).Error("delete level exist user")
		return errcode.CerrDeleteLevelExistUser
	}
	//判断是否为结算层级level
	result, errIsUsed := clienIsLevelNotUsed(feignKey, levelId) //result为isLevelNotUsed
	if nil != errIsUsed {
		logrus.WithFields(logrus.Fields{"levelId": levelId}).Error("request is level used failed")
		return errIsUsed
	}
	if false == result {
		logrus.WithFields(logrus.Fields{"levelId": levelId}).Error("request is level used failed")
		return errcode.CerrLevelIsReportUsed
	}

	levelInfo, errGetLevel := GetLevelInfoByID(feignKey.TenantId, levelId)
	if nil != errGetLevel {
		logrus.WithFields(logrus.Fields{"levelId": levelId}).Error("get level info failed")
		return errGetLevel
	} else if nil == levelInfo {
		logrus.WithFields(logrus.Fields{"levelId": levelId}).Error("delete level not exist")
		return errcode.CerrLevelIDNotExist
	}

	dbConn, errConn := GetDBConnByTenantID(feignKey.TenantId)
	if errConn != nil {
		return errConn
	}
	tx := dbConn.Begin()
	if err := tx.Exec("update t_level set sid = sid -1 where sid >= ?", levelInfo.Sid).Error; nil != err {
		logrus.WithFields(logrus.Fields{"levelId": levelId}).Error("delete level update sid failed")
		tx.Rollback()
		return errcode.CerrExecuteSQL
	}
	if errDelete := tx.Delete(levelInfo).Error; nil != errDelete {
		logrus.WithFields(logrus.Fields{"levelId": levelId}).Error("delete level failed")
		tx.Rollback()
		return errcode.CerrExecuteSQL
	}

	if err := tx.Commit().Error; nil != err {
		logrus.WithFields(logrus.Fields{"levelId": levelId}).Error("delete level commit failed")
		return errcode.CerrExecuteSQL
	}
	return nil
}

func ProcessUpdatelevel(feignKey *XFeignKey, updateReq *LevelDetail) code.Error {
	if nil == feignKey || "" == feignKey.TenantId || nil == updateReq || updateReq.Id <= 0 || "" == updateReq.Name {
		return errcode.CerrParamater
	}

	levelInfo, errGetLevel := GetLevelInfoByID(feignKey.TenantId, updateReq.Id)
	if nil != errGetLevel {
		logrus.WithFields(logrus.Fields{"levelId": updateReq.Id}).Error("get level info failed")
		return errGetLevel
	} else if nil == levelInfo {
		logrus.WithFields(logrus.Fields{"levelId": updateReq.Id}).Error("update level not exist")
		return errcode.CerrLevelIDNotExist
	}

	if levelInfo.Name != updateReq.Name { //检查名字是否重复
		allLevel, errGetLevel := ProcessGetLevelList(feignKey)
		if nil != errGetLevel {
			logrus.WithFields(logrus.Fields{"error": errGetLevel}).Error("ProcessGetLevelList failed")
			return errGetLevel
		}
		for index, _ := range allLevel {
			if allLevel[index].Name == updateReq.Name {
				logrus.WithFields(logrus.Fields{"name": updateReq.Name}).Error("update level name exists")
				return errcode.CerrNameExist
			}
		}
	}

	dbConn, errConn := GetDBConnByTenantID(feignKey.TenantId)
	if errConn != nil {
		return errConn
	}
	tx := dbConn.Begin()

	levelInfo.Name = updateReq.Name
	levelInfo.ModifyUserId = feignKey.UserId
	levelInfo.ModifyDate = time.Now()
	if levelInfo.CreateDate.Unix() < constant.MIN_TIMESTAMP || levelInfo.CreateDate.Unix() > constant.MAX_TIMESTAMP {
		levelInfo.CreateDate = time.Unix(constant.MIN_TIMESTAMP, 0)
	}
	errUpdate := tx.Save(levelInfo).Error
	if nil != errUpdate {
		logrus.WithFields(logrus.Fields{"levelId": updateReq.Id, "roleName": updateReq.Name, "error": errUpdate}).Error("update level failed")
		tx.Rollback()
		return errcode.CerrExecuteSQL
	}
	//更新相关level用户的levelName
	errUpdateUser := tx.Table("t_user_detail").Where("level_id = ? and model_status != ?", updateReq.Id, constant.ModelStatusDelete).Updates(map[string]interface{}{"level_name": updateReq.Name}).Error
	if nil != errUpdateUser && gorm.ErrRecordNotFound != errUpdateUser {
		logrus.WithFields(logrus.Fields{"levelId": updateReq.Id, "levelName": updateReq.Name, "error": errUpdateUser}).Error("update user levelName failed")
		tx.Rollback()
		return errcode.CerrExecuteSQL
	}
	if err := tx.Commit().Error; nil != err {
		logrus.WithFields(logrus.Fields{"levelId": updateReq.Id}).Error("update level commit failed")
		return errcode.CerrExecuteSQL
	}
	return nil
}

func LevelDetailList2LevelDetailMap(levelList []LevelDetail) map[int64]LevelDetail {
	var mapInfo map[int64]LevelDetail = make(map[int64]LevelDetail)
	for index, _ := range levelList {
		mapInfo[levelList[index].Id] = levelList[index]
	}
	return mapInfo
}

func ConvertLevelTal2LevelDetail(record *LevelTab) (*LevelDetail, code.Error) {
	if nil == record {
		return nil, errcode.CerrParamater
	}
	var detail LevelDetail
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
	detail.ProductId = record.ProductId
	detail.CreateUserId = record.CreateUserId
	detail.ModelStatus = record.ModelStatus
	detail.Name = record.Name
	detail.Sid = record.Sid
	detail.UserCount = record.UserCount
	detail.Comment = record.Comment
	return &detail, nil
}

func addLevel(feignKey *XFeignKey, level *LevelTab) code.Error {
	if nil == feignKey || nil == level || "" == level.Name || len(level.Name) > 40 {
		return errcode.CerrParamater
	}

	allLevel, errGetLevel := ProcessGetLevelList(feignKey)
	if nil != errGetLevel {
		logrus.WithFields(logrus.Fields{"error": errGetLevel}).Error("ProcessGetLevelList failed")
		return errGetLevel
	}

	var maxSid int = 0
	for index, _ := range allLevel {
		if allLevel[index].Name == level.Name {
			logrus.WithFields(logrus.Fields{"name": level.Name}).Error("Add level name exists")
			return errcode.CerrNameExist
		}
		if allLevel[index].Sid > maxSid {
			maxSid = allLevel[index].Sid
		}
	}

	if level.Sid < 1 {
		level.Sid = maxSid + 1
	}
	if len(allLevel) > 0 && level.Sid > maxSid+1 { //添加level的sid不能大于当前最大sid+1
		logrus.WithFields(logrus.Fields{"Current maxSid": maxSid, "this Sid": level.Sid}).Error("Add level Sid error")
		return errcode.CerrSidIllegal
	}

	level.CreateDate = time.Now()
	level.CreateUserId = feignKey.UserId
	level.ModifyDate = level.CreateDate
	level.ModifyUserId = feignKey.UserId
	level.ModelStatus = constant.ModelStatusCreate

	dbConn, errConn := GetDBConnByTenantID(feignKey.TenantId)
	if errConn != nil {
		return errConn
	}

	tx := dbConn.Begin()
	if err := tx.Exec("update t_level set sid = sid + 1 where sid >= ?", level.Sid).Error; nil != err {
		logrus.WithFields(logrus.Fields{"level Sid": level.Sid}).Error("Add level update sid failed")
		tx.Rollback()
		return errcode.CerrExecuteSQL
	}

	if errSave := tx.Create(&level).Error; nil != errSave {
		logrus.WithFields(logrus.Fields{"name": level.Name, "sid": level.Sid, "error": errSave}).Error("Add level failed")
		tx.Rollback()
		return errcode.CerrExecuteSQL
	}

	if err := tx.Commit().Error; nil != err {
		logrus.WithFields(logrus.Fields{"name": level.Name, "sid": level.Sid, "error": err}).Error("Add level failed")
		return errcode.CerrExecuteSQL
	}
	return nil
}

func initLevel(feignKey *XFeignKey) {
	var level LevelTab
	level.Name = "IB1"
	level.Sid = 1
	addLevel(feignKey, &level)

	var level2 LevelTab
	level2.Name = "IB2"
	level2.Sid = 2
	addLevel(feignKey, &level2)

	var level3 LevelTab
	level3.Name = "IB3"
	level3.Sid = 3
	addLevel(feignKey, &level3)
}

func ProcessGetLevelListByAuthority(feignKey *XFeignKey) ([]LevelDetail, code.Error) {
	tenant_id := feignKey.TenantId
	if "" == tenant_id {
		return nil, errcode.CerrParamater
	}

	userDetail, errGetUser := GetUserDetailByIdOrPubid(feignKey, feignKey.UserId, 0)
	if nil != errGetUser {
		logrus.Errorf("GetUserDetailByIdOrPubid failed, pubId:%v, error:%v", feignKey.UserId, errGetUser)
		return nil, errGetUser
	}

	rightRecords, errGetRight := GetOneRoleRightList(feignKey, userDetail.RoleId)
	if nil != errGetRight {
		logrus.Errorf("GetOneRoleRightList failed, pubId:%v, error:%v", feignKey.UserId, errGetRight)
		return nil, errGetRight
	}
	logrus.Debugf("GetOneRoleRightList, result size:%v, pubId:%v, roleID:%v", len(rightRecords), feignKey.UserId, userDetail.RoleId)

	levelAll, errGetLevel := ProcessGetLevelList(feignKey)
	if nil != errGetRight {
		logrus.Errorf("ProcessGetLevelList failed, pubId:%v, error:%v", feignKey.UserId, errGetLevel)
		return nil, errGetRight
	}
	logrus.Debugf("ProcessGetLevelList result size:%v", len(levelAll))

	//获取用户的level信息
	var userLevel *LevelTab = nil
	if userDetail.LevelId > 0 {
		var errUserLevel code.Error
		userLevel, errUserLevel = GetLevelInfoByID(feignKey.TenantId, userDetail.LevelId)
		if nil != errUserLevel {
			logrus.Errorf("GetLevelInfoByID failed, useID:%v, error:%v", userDetail.Id, errGetLevel)
			return nil, errUserLevel
		}
	}

	if nil == userLevel || RightsHaveCotain(rightRecords, constant.AUTHORITY_USER_SELECT_ALL) || RightsHaveCotain(rightRecords, constant.AUTHORITY_USER_SELECT_WILD) {
		return levelAll, nil
	}
	sort.Sort(LevelDetailSlice(levelAll))
	var filterLevels []LevelDetail = make([]LevelDetail, 0)
	//归属给下级的账户权限（此时userLevel不会为nil），找sid>用户sid的记录,跳过一条
	if RightsHaveCotain(rightRecords, constant.AUTHORITY_USER_SELECT_SUBORDINATE) {
		flag := false
		//如果有直属权限，则不跳过
		if RightsHaveCotain(rightRecords, constant.AUTHORITY_USER_SELECT_DIRECTLY) {
			flag = true
		}
		for _, record := range levelAll {
			if record.Sid > userLevel.Sid && false == flag {
				flag = true
			} else if record.Sid > userLevel.Sid && true == flag {
				filterLevels = append(filterLevels, record)
			}
		}
		return filterLevels, nil
	}

	//查看归属给我的账户, 找sid>用户sid的第一条记录
	if RightsHaveCotain(rightRecords, constant.AUTHORITY_USER_SELECT_DIRECTLY) {
		var findUserLevel bool = false
		for _, record := range levelAll {
			if true == findUserLevel {
				filterLevels = append(filterLevels, record)
				return filterLevels, nil
			}
			if record.Sid == userLevel.Sid {
				findUserLevel = true
			}
		}
	}

	return filterLevels, nil
}

func ProcessGetEarningReportByAuthority(feignKey *XFeignKey) ([]LevelDetail, code.Error) {
	tenant_id := feignKey.TenantId
	if "" == tenant_id {
		return nil, errcode.CerrParamater
	}

	userDetail, errGetUser := GetUserDetailByIdOrPubid(feignKey, feignKey.UserId, 0)
	if nil != errGetUser {
		logrus.Errorf("GetUserDetailByIdOrPubid failed, pubId:%v, error:%v", feignKey.UserId, errGetUser)
		return nil, errGetUser
	} else if nil == userDetail {
		return nil, errcode.CerrAccountNotFound
	}
	rightRecords, errGetRight := GetOneRoleRightList(feignKey, userDetail.RoleId)
	if nil != errGetRight {
		logrus.Errorf("GetOneRoleRightList failed, pubId:%v, error:%v", feignKey.UserId, errGetRight)
		return nil, errGetRight
	}
	logrus.Debugf("GetOneRoleRightList, result size:%v, pubId:%v, roleID:%v", len(rightRecords), feignKey.UserId, userDetail.RoleId)
	levelAll, errGetLevel := ProcessGetLevelList(feignKey)
	if nil != errGetRight {
		logrus.Errorf("ProcessGetLevelList failed, pubId:%v, error:%v", feignKey.UserId, errGetLevel)
		return nil, errGetRight
	}
	logrus.Debugf("ProcessGetLevelList result size:%v", len(levelAll))

	var userLevelInfo *LevelTab = nil
	if userDetail.LevelId > 0 {
		var errGetLevel code.Error
		userLevelInfo, errGetLevel = GetLevelInfoByID(feignKey.TenantId, userDetail.LevelId)
		if nil != errGetLevel {
			logrus.WithFields(logrus.Fields{"pubId": feignKey.UserId, "levelId": userDetail.LevelId, "error": errGetLevel}).Error("get level info failed")
			return nil, errGetRight
		}
	}

	if nil == userLevelInfo || RightsHaveCotain(rightRecords, constant.AUTHORITY_EARNING_REPORT_SELECT_ALL) || RightsHaveCotain(rightRecords, constant.AUTHORITY_EARNING_REPORT_SELECT_WILD) {
		return levelAll, nil
	}
	sort.Sort(LevelDetailSlice(levelAll))
	var levelResult []LevelDetail = make([]LevelDetail, 0)
	if RightsHaveCotain(rightRecords, constant.AUTHORITY_EARNING_REPORT_SELECT_INDIRECT) {
		//寻找比用户level的sid大两个层级的level信息（跳过一个）
		flag := false
		//如果有直属权限，则不跳过
		if RightsHaveCotain(rightRecords, constant.AUTHORITY_EARNING_REPORT_SELECT_DIRECT) {
			flag = true
		}
		for index, _ := range levelAll {
			if levelAll[index].Sid > userLevelInfo.Sid && false == flag {
				flag = true
			} else if levelAll[index].Sid > userLevelInfo.Sid && true == flag {
				levelResult = append(levelResult, levelAll[index])
			}
		}
		return levelResult, nil
	}

	if RightsHaveCotain(rightRecords, constant.AUTHORITY_EARNING_REPORT_SELECT_DIRECT) {
		//寻找用户level的sid的下一级level信息（一个level Info）
		flag := false
		for index, _ := range levelAll {
			if levelAll[index].Sid > userLevelInfo.Sid && false == flag {
				flag = true
			} else if levelAll[index].Sid > userLevelInfo.Sid && true == flag {
				levelResult = append(levelResult, levelAll[index])
				return levelResult, nil
			}
		}
	}
	return levelResult, nil
}

//模糊查询层级，目前按名称查询
func SearchLevelByFuzzyName(feignKey *XFeignKey, fuzzyValue string) ([]*LevelTab, code.Error) {
	if nil == feignKey || "" == feignKey.TenantId || "" == fuzzyValue {
		return nil, errcode.CerrParamater
	}
	dbConn, errConn := GetDBConnByTenantID(feignKey.TenantId)
	if errConn != nil {
		return nil, errConn
	}

	likeFuzzyVal := "%" + escapeSql(fuzzyValue) + "%"
	var levelList = make([]*LevelTab, 0)
	dbtmp := dbConn.Where("name like ? and model_status != ?", likeFuzzyVal, constant.ModelStatusDelete).Find(&levelList)
	if err := dbtmp.Error; nil != err {
		if err == gorm.ErrRecordNotFound {
			logrus.Warnf("Not found level record from DB, FuzzyVal:%v", likeFuzzyVal)
			return levelList, nil //未找到记录不返回错误
		}
		logrus.Errorf("Failed to get level record from DB, FuzzyVal:%v, error:%v", likeFuzzyVal, err)
		return nil, errcode.CerrExecuteSQL
	}
	return levelList, nil
}

//获取某截止时间前创建的层级数
func GetLevelListBeforeOneCreateTime(tenant_id string, deadTime time.Time) ([]LevelTab, code.Error) {
	if "" == tenant_id {
		return nil, errcode.CerrParamater
	}
	dbConn, errConn := GetDBConnByTenantID(tenant_id)
	if errConn != nil {
		return nil, errConn
	}

	var levelList []LevelTab = make([]LevelTab, 0)
	dbtmp := dbConn.Where("create_date < ? and model_status != ?", deadTime, constant.ModelStatusDelete).Find(&levelList)

	if err := dbtmp.Error; nil != err {
		if err == gorm.ErrRecordNotFound {
			logrus.WithFields(logrus.Fields{"deadTime time": deadTime}).Warn("Not found Level Record record from DB")
			return levelList, nil //未找到记录不返回错误
		}
		logrus.WithFields(logrus.Fields{"deadTime time": deadTime, "error": err}).Error("Failed to get level record from DB")
		return nil, errcode.CerrExecuteSQL
	}
	return levelList, nil
}

//根据名称查询层级
func GetLevelByName(feignKey *XFeignKey, name string) (*LevelTab, code.Error) {
	if nil == feignKey || "" == feignKey.TenantId || "" == name {
		return nil, errcode.CerrParamater
	}
	dbConn, errConn := GetDBConnByTenantID(feignKey.TenantId)
	if errConn != nil {
		return nil, errConn
	}

	var levelInfo LevelTab
	dbtmp := dbConn.Where("name = ? and model_status != ?", name, constant.ModelStatusDelete).Find(&levelInfo)
	if err := dbtmp.Error; nil != err {
		if err == gorm.ErrRecordNotFound {
			logrus.WithFields(logrus.Fields{"name": name}).Warn("Not found level record from DB by name")
			return nil, nil //未找到记录不返回错误
		}
		logrus.WithFields(logrus.Fields{"name": name, "error": err}).Error("Failed to get level record from DB by name")
		return nil, errcode.CerrExecuteSQL
	}
	return &levelInfo, nil
}
