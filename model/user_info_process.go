package model

import (
	"bw/bw-user/conf"
	"bw/bw-user/constant"
	"bw/bw-user/errcode"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/lworkltd/kits/service/restful/code"
	"github.com/lworkltd/kits/utils/operator"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"strconv"
	"strings"
	"time"
)

//根据t_user_detail表中的keyType列获取值在values中的记录记录, values为切片
func ProcessListUserRecordsByKey(feignKey *XFeignKey, values interface{}, keyType string) ([]UserDetailTab, code.Error) {
	dbConn, errConn := GetDBConnByTenantID(feignKey.TenantId)
	if errConn != nil {
		return nil, errConn
	}

	var userRecords []UserDetailTab = make([]UserDetailTab, 0)
	var dbtmp *gorm.DB
	switch keyType {
	case constant.QueryKeyId:
		dbtmp = dbConn.Where("id in (?) and model_status != ?", values, constant.ModelStatusDelete).Find(&userRecords)
	case constant.QueryKeyLevelId:
		dbtmp = dbConn.Where("level_id in (?) and model_status != ?", values, constant.ModelStatusDelete).Find(&userRecords)
	case constant.QueryKeyRoleId:
		dbtmp = dbConn.Where("role_id in (?) and model_status != ?", values, constant.ModelStatusDelete).Find(&userRecords)
	case constant.QueryKeyName:
		dbtmp = dbConn.Where("name in (?) and model_status != ?", values, constant.ModelStatusDelete).Find(&userRecords)
	case constant.QueryKeyEmail:
		dbtmp = dbConn.Where("email in (?) and model_status != ?", values, constant.ModelStatusDelete).Find(&userRecords)
	case constant.QueryKeyEntityNo:
		dbtmp = dbConn.Where("entity_no in (?) and model_status != ?", values, constant.ModelStatusDelete).Find(&userRecords)
	case constant.QueryKeyLogin:
		dbtmp = dbConn.Where("login in (?) and model_status != ?", values, constant.ModelStatusDelete).Find(&userRecords)
	case constant.QueryKeypubUserIds:
		dbtmp = dbConn.Where("pub_user_id in (?) and model_status != ?", values, constant.ModelStatusDelete).Find(&userRecords)
	case constant.QueryKeyIdNum:
		dbtmp = dbConn.Where("id_num in (?) and model_status != ?", values, constant.ModelStatusDelete).Find(&userRecords)
	case constant.QueryKeyAllRecords:
		dbtmp = dbConn.Where(" model_status != ?", constant.ModelStatusDelete).Find(&userRecords)
	default:
		logrus.WithFields(logrus.Fields{"keyType": keyType}).Error("ProcessListUserInfosByKey keyType not support")
		return nil, errcode.CerrQueryKeyNotSupport
	}

	if err := dbtmp.Error; nil != err {
		if err == gorm.ErrRecordNotFound {
			logrus.WithFields(logrus.Fields{"keyType": keyType, "value": values}).Warn("Not found user records from DB by")
			return userRecords, nil //未找到记录不返回错误
		}
		logrus.WithFields(logrus.Fields{"error": err, "keyType": keyType, "value": values}).Error("Failed to get records from DB by")
		return nil, errcode.CerrExecuteSQL
	}

	return userRecords, nil
}

//根据t_user_detail表中的keyType列获取值在values中的记录记录, values为切片
func ProcessListUserInfosByKey(feignKey *XFeignKey, values interface{}, keyType string) ([]UserDetail, code.Error) {
	userRecords, errGetRecords := ProcessListUserRecordsByKey(feignKey, values, keyType)
	if nil != errGetRecords {
		logrus.WithFields(logrus.Fields{"error": errGetRecords, "keyType": keyType, "value": values}).Error("ProcessListUserRecordsByKey failed")
		return nil, errGetRecords
	}

	var userDetails []UserDetail = make([]UserDetail, 0)
	//userRecords 转换为userDetails
	var errCov code.Error
	userDetails, errCov = UserRecords2UserDetails(userRecords)
	if nil != errCov {
		logrus.WithFields(logrus.Fields{"keyType": keyType, "value": values}).Warn("Convert user info from []UserDetailTab to []UserDetail failed")
		return userDetails, errcode.CerrInternal
	}

	ReviseRoleNameAndLevelName(userDetails, feignKey)
	return userDetails, nil
}

//获取无parent的用户信息
func ProcessListNoParentUsers(feignKey *XFeignKey) ([]UserDetailTab, code.Error) {
	if nil == feignKey || "" == feignKey.TenantId {
		return nil, errcode.CerrParamater
	}
	dbConn, errConn := GetDBConnByTenantID(feignKey.TenantId)
	if errConn != nil {
		return nil, errConn
	}

	var userRecords []UserDetailTab = make([]UserDetailTab, 0)
	dbtmp := dbConn.Where("(parent_id is null or parent_id = '' or parent_id = '0') and id !=1 and model_status != ?", constant.ModelStatusDelete).Find(&userRecords)

	if err := dbtmp.Error; nil != err {
		if err == gorm.ErrRecordNotFound {
			logrus.WithFields(logrus.Fields{}).Warn("Not found no parent user records from DB")
			return userRecords, nil //未找到记录不返回错误
		}
		logrus.WithFields(logrus.Fields{"error": err}).Error("Failed to get no parent user records from DB")
		return nil, errcode.CerrExecuteSQL
	}

	return userRecords, nil
}

//查询所有用户的简单信息（含Iid，name，entityNo）
func ProcessListSimpleUser(feignKey *XFeignKey) ([]IdNameDTO, code.Error) {
	if nil == feignKey || "" == feignKey.TenantId {
		return nil, errcode.CerrParamater
	}
	dbConn, errConn := GetDBConnByTenantID(feignKey.TenantId)
	if errConn != nil {
		return nil, errConn
	}

	var simplesRecords []IdNameDTO = make([]IdNameDTO, 0)

	rows, errRaw := dbConn.Raw("select id, name, entity_no from t_user_detail where model_status != ?", constant.ModelStatusDelete).Rows() // (*sql.Rows, error)
	defer rows.Close()
	if nil != errRaw {
		if errRaw == gorm.ErrRecordNotFound {
			logrus.WithFields(logrus.Fields{}).Warn("Not found user records from DB")
			return simplesRecords, nil //未找到记录不返回错误
		}
		logrus.WithFields(logrus.Fields{"error": errRaw}).Error("Failed to get user records from DB")
		return nil, errcode.CerrExecuteSQL
	}

	for rows.Next() {
		var record IdNameDTO
		dbConn.ScanRows(rows, &record)
		simplesRecords = append(simplesRecords, record)
	}

	return simplesRecords, nil
}

//查询有返佣账户的所有用户的简单信息（含Iid，name，entityNo）
func ProcessListSimpleUserHasAccountUser(feignKey *XFeignKey) ([]IdNameDTO, code.Error) {
	if nil == feignKey || "" == feignKey.TenantId {
		return nil, errcode.CerrParamater
	}
	dbConn, errConn := GetDBConnByTenantID(feignKey.TenantId)
	if errConn != nil {
		return nil, errConn
	}

	var simplesRecords []IdNameDTO = make([]IdNameDTO, 0)

	rows, errRaw := dbConn.Raw("select id, name, entity_no from t_user_detail where model_status != ? and !ISNULL(login) and login != ''", constant.ModelStatusDelete).Rows() // (*sql.Rows, error)
	defer rows.Close()
	if nil != errRaw {
		if errRaw == gorm.ErrRecordNotFound {
			logrus.WithFields(logrus.Fields{}).Warn("Not found user records from DB")
			return simplesRecords, nil //未找到记录不返回错误
		}
		logrus.WithFields(logrus.Fields{"error": errRaw}).Error("Failed to get user records from DB")
		return nil, errcode.CerrExecuteSQL
	}

	for rows.Next() {
		var record IdNameDTO
		dbConn.ScanRows(rows, &record)
		simplesRecords = append(simplesRecords, record)
	}

	return simplesRecords, nil
}

//查询所有用户的简单信息（含id,name,parent,levelId,levelName,roleId,roleName,entityNo,pubUserId,login,vendorServerId）
func ProcessListUserAndLevel(feignKey *XFeignKey) ([]SimpleUserDTO, code.Error) {
	if nil == feignKey || "" == feignKey.TenantId {
		return nil, errcode.CerrParamater
	}
	dbConn, errConn := GetDBConnByTenantID(feignKey.TenantId)
	if errConn != nil {
		return nil, errConn
	}

	var simplesRecords []SimpleUserDTO = make([]SimpleUserDTO, 0)

	rows, errRaw := dbConn.Raw("select id, name, parent_id, level_id, level_name, role_id, role_name, entity_no, pub_user_id, login, vendor_server_id from t_user_detail where model_status != ?", constant.ModelStatusDelete).Rows() // (*sql.Rows, error)
	defer rows.Close()
	if nil != errRaw {
		if errRaw == gorm.ErrRecordNotFound {
			logrus.WithFields(logrus.Fields{}).Warn("Not found user records from DB")
			return simplesRecords, nil //未找到记录不返回错误
		}
		logrus.WithFields(logrus.Fields{"error": errRaw}).Error("Failed to get user records from DB")
		return nil, errcode.CerrExecuteSQL
	}

	for rows.Next() {
		var record SimpleUserDTO
		dbConn.ScanRows(rows, &record)
		simplesRecords = append(simplesRecords, record)
	}

	return simplesRecords, nil
}

//根据pub_user_id或者ID，获取一个用户信息，若pubId不是空串则使用pubId查询，否则使用idNum查询
func GetUserRecordByIdOrPubid(feignKey *XFeignKey, pubId string, idNum int64) (*UserDetailTab, code.Error) {
	if nil == feignKey || ("" == pubId && idNum <= 0) {
		return nil, errcode.CerrParamater
	}
	dbConn, errConn := GetDBConnByTenantID(feignKey.TenantId)
	if errConn != nil {
		return nil, errConn
	}

	var userRecord UserDetailTab
	var dbtmp *gorm.DB
	if "" != pubId {
		dbtmp = dbConn.Where("pub_user_id = ? and model_status != ?", pubId, constant.ModelStatusDelete).First(&userRecord)
	} else {
		dbtmp = dbConn.Where("id = ? and model_status != ?", idNum, constant.ModelStatusDelete).First(&userRecord)
	}

	if err := dbtmp.Error; nil != err {
		if err == gorm.ErrRecordNotFound {
			logrus.Warnf("Not found user record from DB, pub_user_id:%v, idNum:%v", pubId, idNum)
			return nil, nil //未找到记录不返回错误
		}
		logrus.Errorf("Failed to get records from DB, pub_user_id:%v, idNum:%v, error:%v", pubId, idNum, err)
		return nil, errcode.CerrExecuteSQL
	}
	userRecord.TenantId = feignKey.TenantId
	return &userRecord, nil
}

//根据pub_user_id或者ID，获取一个用户信息，若pubId不是空串则使用pubId查询，否则使用idNum查询
func GetUserDetailByIdOrPubid(feignKey *XFeignKey, pubId string, idNum int64) (*UserDetail, code.Error) {
	userRecord, errGetUser := GetUserRecordByIdOrPubid(feignKey, pubId, idNum)
	if nil != errGetUser {
		logrus.WithFields(logrus.Fields{"errGetUser": errGetUser}).Error("GetUserRecordByIdOrPubid failed")
		return nil, errGetUser
	} else if nil == userRecord {
		return nil, nil //用户不存在不返回错误
	}

	//userRecords 转换为userDetails
	return UserRecord2UserDetail(userRecord)
}

//UserDetailTab 转换为UserDetail
func UserRecord2UserDetail(record *UserDetailTab) (*UserDetail, code.Error) {
	if nil == record {
		return nil, errcode.CerrParamater
	}
	var detail UserDetail
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
	detail.ProduceId = record.ProductId
	detail.CreateUserId = record.CreateUserId
	detail.ModelStatus = record.ModelStatus
	detail.Name = record.Name
	detail.UserName = record.Username
	detail.Email = record.Email
	detail.Phone = record.Phone
	detail.Address = record.Address
	detail.Country = record.Country
	detail.Province = record.Province
	detail.City = record.City
	detail.Postcode = record.Postcode
	detail.RoleId = record.RoleId
	detail.RoleName = record.RoleName
	detail.LevelId = record.LevelId
	detail.LevelName = record.LevelName
	detail.Parent = record.ParentId
	detail.Sex = record.Sex
	detail.VendorServerId = record.VendorServerId
	detail.Comment = record.Comment
	detail.Nickname = record.Nickname
	detail.HeadImage = record.HeadImage
	detail.Birthday = record.Birthday
	detail.PubUserId = record.PubUserId
	detail.SubUserCount = record.SubUserCount
	detail.Version = record.Version
	detail.Active = record.Active
	detail.Login = record.Login
	if len(record.NeedInitPass) > 0 && 1 == record.NeedInitPass[0] {
		detail.NeedInitPass = true
	} else {
		detail.NeedInitPass = false
	}
	detail.Region.Country = record.Country
	detail.Region.Province = record.Province
	detail.Region.City = record.City
	detail.Phones.CountryCode, detail.Phones.Phone, detail.Phones.PhoneStr = splitPhoneValue(record.Phone)
	detail.IdType = record.IdType
	detail.IdNum = record.IdNum
	detail.IdUrl1 = record.IdUrl1
	detail.IdUrl2 = record.IdUrl2
	detail.BankAccount = record.BankAccount
	detail.BankBranch = record.BankBranch
	detail.AccountNo = record.AccountNo
	detail.BankCardFile1 = record.BankCardFile1
	detail.BankCardFile2 = record.BankCardFile2
	detail.DoAgencyBusiness = record.DoAgencyBusiness
	detail.InvestExperience = record.InvestExperience
	detail.Agent = record.Agent
	detail.TwoFactorAuth = record.TwoFactorAuth
	detail.Field01 = record.Field01
	detail.Field02 = record.Field02
	detail.Field03 = record.Field03
	detail.Field04 = record.Field04
	detail.Field05 = record.Field05
	detail.Field06 = record.Field06
	detail.Field07 = record.Field07
	detail.Field08 = record.Field08
	detail.Field09 = record.Field09
	detail.Field10 = record.Field10
	detail.Field11 = record.Field11
	detail.Field12 = record.Field12
	detail.Field13 = record.Field13
	detail.Field14 = record.Field14
	detail.Field15 = record.Field15
	detail.Field16 = record.Field16
	detail.Field17 = record.Field17
	detail.Field18 = record.Field18
	detail.Field19 = record.Field19
	detail.Field20 = record.Field20

	detail.Points1 = record.Points1
	detail.Points2 = record.Points2
	detail.Points3 = record.Points3
	detail.Points4 = record.Points4
	detail.Points5 = record.Points5
	detail.Points6 = record.Points6
	detail.Points7 = record.Points7
	return &detail, nil
}

//UserDetailTab 转换为UserDetail
func UserRecord2SimpleUser(record *UserDetailTab) (*SimpleUserDTO, code.Error) {
	if nil == record {
		return nil, errcode.CerrParamater
	}
	var simpleUser SimpleUserDTO
	simpleUser.Id = record.Id
	simpleUser.EntityNo = record.EntityNo
	simpleUser.Name = record.Name
	simpleUser.RoleId = record.RoleId
	simpleUser.RoleName = record.RoleName
	simpleUser.LevelId = record.LevelId
	simpleUser.LevelName = record.LevelName
	simpleUser.ParentId = record.ParentId
	simpleUser.VendorServerId = record.VendorServerId
	simpleUser.PubUserId = record.PubUserId
	simpleUser.Login = record.Login
	simpleUser.TwoFactorAuth = record.TwoFactorAuth
	return &simpleUser, nil
}

//[]UserDetailTab 转换为 []UserDetail
func UserRecords2UserDetails(userRecords []UserDetailTab) ([]UserDetail, code.Error) {
	userDetails := make([]UserDetail, 0)
	for _, record := range userRecords {
		detail, errConvert := UserRecord2UserDetail(&record)
		if nil != errConvert || nil == detail {
			return userDetails, errcode.CerrInternal
		}
		userDetails = append(userDetails, *detail)
	}
	return userDetails, nil
}

//phoneStr，参数："+86@-@1111111111"，返回："+86", "1111111111", "+86 1111111111"
func splitPhoneValue(phoneValue string) (string, string, string) {
	phone := ""
	countryCode := ""
	phoneStr := ""
	if "" == phoneValue {
		return countryCode, phone, phoneStr
	}

	strTmp := strings.Split(phoneValue, "@-@")
	if len(strTmp) == 2 {
		countryCode = strTmp[0]
		phone = strTmp[1]
		phoneStr = strTmp[0] + " " + strTmp[1]
	} else {
		phoneStr = phoneValue
		phone = phoneValue
	}
	if strings.ToUpper(countryCode) == "NULL" {
		countryCode = ""
	}
	if strings.ToUpper(phone) == "NULL" {
		phone = ""
		phoneStr = ""
	}
	return countryCode, phone, phoneStr
}

//添加用户处理
func ProcessAddUser(userInfo *BWUserDTO, feignKey *XFeignKey, domain string) (*UserDetail, code.Error) {
	//用户数判断
	if reachMaxUserAmount(feignKey) == true {
		return nil, code.NewMcode(
			fmt.Sprintf("BW_USER_MAX"),
			"BW_USER_MAX",
		)
	}

	tenant_id := feignKey.TenantId
	if err := checkParameterForAddUser(userInfo, feignKey); nil != err {
		return nil, err
	}

	userInfo.Email = strings.TrimSpace(userInfo.Email)
	//邮箱检验
	isEmail, _ := IsEmail(userInfo.Email)

	if isEmail == false {
		return nil, code.NewMcode(
			fmt.Sprintf("EMAIL_FORMAT_ERROR"),
			"EMAIL_FORMAT_ERROR",
		)
	}

	//向pub-auth请求添加用户
	var pubUserInfo PubUserInfoDTO = PubUserInfoDTO{
		Mail:     userInfo.Email,
		Phone:    userInfo.Phones.Phone,
		Password: userInfo.Password,
	}
	pubUserId, errPubAuth := ClientAddPubAuthUser(feignKey, &pubUserInfo)
	if nil != errPubAuth {
		return nil, errPubAuth
	}
	userInfo.PubUserId = pubUserId
	pubUserInfo.UserId = pubUserId

	//保存用户基本信息到DB
	userTab, errTab := BWUserDTO2UserDetailTabForAdd(userInfo, feignKey)
	if nil != errTab {
		return nil, errTab
	}
	errUserSave := SaveUserInfoToDB(tenant_id, userTab)
	if nil != errUserSave {
		ClientDeletePubAuthUser(feignKey, &pubUserInfo)
		return nil, errUserSave
	}

	userInfo.Id = userTab.Id
	logrus.WithFields(logrus.Fields{"user_id": userTab.Id}).Debug("After save user info")

	//保存黑白名单到DB（暂未实现）	/////////////

	//请求 用户佣金规则详细-添加修改
	if nil == userInfo.Commission {
		userInfo.Commission = new(UserCommissionUpdateRequestDTO)
	}
	userInfo.Commission.UserId = userTab.Id
	userInfo.Commission.LevelId, _ = strconv.ParseInt(userInfo.LevelId, 10, 64)
	userInfo.Commission.ParentId, _ = strconv.ParseInt(userInfo.Parent, 10, 64)
	_, errCommission := ClientAddOrUpdateCommission(feignKey, userInfo.Commission)
	if nil != errCommission {
		logrus.WithFields(logrus.Fields{"error": errCommission}).Warn("AddOrUpdateCommission failed")
		ClientDeletePubAuthUser(feignKey, &pubUserInfo) //添加Commission失败,删除pubuserId
		dbConn, errConn := GetDBConnByTenantID(tenant_id)
		if errConn != nil {
			return nil, errConn
		}
		dbConn.Delete(userTab) //添加Commission失败，删除用户
		return nil, errCommission
	}

	//计算更新相关的用户数（比如下级用户数，层级用户数等）
	go UpdageSubCounter(feignKey)

	//发送邮件
	errSendEmail := SendAddUserEmail(feignKey, userInfo, domain)
	if nil != errSendEmail {
		logrus.WithFields(logrus.Fields{"error": errSendEmail}).Warn("SendAddUserEmail failed")
	}
	logrus.Infof("Add one user success, email:%v, userID:%v, operate user:%v", userTab.Email, userTab.Id, feignKey.UserId)

	result, err := UserRecord2UserDetail(userTab)
	result.Password = userInfo.Password
	return result, err
}

func reachMaxUserAmount(feignKey *XFeignKey) bool {
	count, _ := ClientGetPubAuthUserNumber(feignKey)
	limit, _ := ClientGetProductUserLimit(feignKey)
	logrus.Infof("reachMaxUserAmount count:%v, limit:%v", count, limit)
	return count >= limit
}

func SaveUserInfoToDB(tenant_id string, userTab *UserDetailTab) code.Error {
	if "" == tenant_id || nil == userTab {
		return errcode.CerrParamater
	}
	dbConn, errConn := GetDBConnByTenantID(tenant_id)
	if errConn != nil {
		return errConn
	}

	dbtmp := dbConn.Create(userTab)
	if err := dbtmp.Error; nil != err {
		logrus.WithFields(logrus.Fields{
			"error": err,
			"email": userTab.Email,
		}).Error("Failed save user info to DB")
		return errcode.CerrExecuteSQL
	}

	return nil
}

func SendAddUserEmail(feignKey *XFeignKey, userInfo *BWUserDTO, domain string) code.Error {
	if nil == feignKey || "" == feignKey.TenantId || nil == userInfo {
		return errcode.CerrParamater
	}
	if false == userInfo.SendEmail {
		return nil
	}

	tenantProduct, err := ClientProductTenantByKey(feignKey, feignKey.TenantId, "BW")
	if err != nil {
		logrus.Warn("get bw product info fail {} {}", feignKey.TenantId, err)
		tenantProduct = &TenantProductDTO{}
	}

	expireTime := time.Now().Add(30 * time.Minute)
	expire := expireTime.Format("2006-01-02 15:04:05")

	var email EmailDTO
	email.SendUser.UserId = "-1"
	email.SendUser.UserName = "system"
	email.ToUser = make([]UserDTO, 0)
	var touser = UserDTO{UserId: "-1", UserName: userInfo.Email}
	email.ToUser = append(email.ToUser, touser)
	if userInfo.ForTask {
		email.TemplateType = "BW_TASK_ADD_A_USER"
	} else {
		email.TemplateType = "ADD_A_USER"
	}

	email.Message.Vars = make(map[string]string)
	email.Message.Vars["usl"] = domain
	email.Message.Vars["email"] = userInfo.Email
	email.Message.Vars["password"] = userInfo.Password

	email.Message.Vars["name"] = userInfo.Name
	email.Message.Vars["phone"] = strings.Replace(userInfo.Phone, "@-@", " ", -1)
	email.Message.Vars["id_num"] = userInfo.IdNum
	email.Message.Vars["address"] = userInfo.Address
	email.Message.Vars["account_no"] = userInfo.AccountNo
	email.Message.Vars["expire"] = expire

	email.Message.Vars["companyName"] = tenantProduct.CompanyName
	email.Message.Vars["CompanyName"] = tenantProduct.CompanyName
	email.Message.Vars["company_name"] = tenantProduct.CompanyName
	email.Message.Vars["company_site"] = tenantProduct.CompanySite
	email.Message.Vars["company_email"] = tenantProduct.CompanyEmail
	email.Message.Vars["company_phone"] = tenantProduct.CompanyPhone
	email.Message.Vars["company_address"] = tenantProduct.CompanyAddress
	email.TenantInfo.TenantId = feignKey.TenantId
	email.TenantInfo.ProductId = "BW"
	email.Lang = userInfo.Lang

	_, errSendEmail := ClientSendAddUserEmail(feignKey, &email)
	return errSendEmail
}

func checkParameterForAddUser(userInfo *BWUserDTO, feignKey *XFeignKey) code.Error {
	if nil == feignKey || "" == feignKey.TenantId || nil == userInfo {
		return code.New(errcode.ParameterError, "feign key or tenantId is null")
	}
	if "" == userInfo.Email {
		return code.New(errcode.ParameterError, "email is null")
	}
	if "" == userInfo.Name {
		return code.New(errcode.ParameterError, "user name is null")
	}

	roleIdPara, errParseRole := strconv.ParseInt(userInfo.RoleId, 10, 64)
	if nil != errParseRole || roleIdPara <= 0 {
		return code.New(errcode.ParameterError, "role is invalid")
	}
	if "" == userInfo.Password {
		userInfo.Password = GeneratePassword(conf.GetApplication().PwdStrength)
	}

	userInfo.CreateDate = time.Now().Unix() * 1000
	//loginCheck
	if nil != userInfo.Login {
		loginNum, error := strconv.Atoi(*userInfo.Login)
		if nil != error {
			userInfo.Login = nil
		} else if loginNum > 0 && (nil == userInfo.VendorServerId || false == strings.Contains(*userInfo.VendorServerId, "_")) {
			return errcode.CerrParamater
		}
	}

	//检查或者生成EntityNo
	var errEntityNo code.Error
	userInfo.EntityNo, errEntityNo = GenerateEntityNo(feignKey.TenantId, userInfo.EntityNo)
	if nil != errEntityNo {
		logrus.WithFields(logrus.Fields{
			"Error": errEntityNo,
		}).Error("Generate EntityNo failed")
		return errEntityNo
	}

	//检查email
	emailExist, errEmail := ExistInUserDetailTab(feignKey.TenantId, constant.QueryKeyEmail, userInfo.Email)
	if nil != errEmail {
		logrus.WithFields(logrus.Fields{
			"email": userInfo.Email,
		}).Error("ExistInUserDetailTab for email failed")
		return errEmail
	} else if true == emailExist {
		logrus.WithFields(logrus.Fields{
			"email": userInfo.Email,
		}).Error("email have exist")
		return errcode.CerrEmailExist
	}

	//获取roleInfo
	roleInfo, errRole := GetRoleInfoByID(feignKey.TenantId, roleIdPara)
	if nil != errRole {
		logrus.WithFields(logrus.Fields{
			"roleID": userInfo.RoleId,
		}).Error("GetRoleInfoByID failed")
		return errRole
	} else if nil == roleInfo {
		logrus.WithFields(logrus.Fields{
			"roleID": userInfo.RoleId,
		}).Error("role haven't exist")
		return errcode.CerrRoleIDNotExist
	}
	userInfo.RoleName = roleInfo.Name //用DB中查询出的名字替换

	//获取level信息
	levelIdPara, errParseLevel := strconv.ParseInt(userInfo.LevelId, 10, 64)
	if nil == errParseLevel && levelIdPara > 0 {
		levelInfo, errLevel := GetLevelInfoByID(feignKey.TenantId, levelIdPara)
		if nil != errLevel {
			logrus.WithFields(logrus.Fields{
				"levelID": userInfo.LevelId,
			}).Error("GetLevelInfoByID failed")
			return errLevel
		} else if nil == levelInfo {
			logrus.WithFields(logrus.Fields{
				"levelID": userInfo.LevelId,
			}).Error("level haven't exist")
			return errcode.CerrLevelIDNotExist
		}
		userInfo.LevelName = levelInfo.Name
	} else {
		userInfo.LevelId = "0"
		userInfo.LevelName = ""
	}

	//用户层级不能大于上级用户层级
	parentIdPara, errParseParent := strconv.ParseInt(userInfo.Parent, 10, 64)
	if nil == errParseParent && parentIdPara > 0 {
		if userInfo.LevelId != "0" {
			//检查parent的level是否合适
			userRecords, err := ProcessFindUserByTypeFuzzy(feignKey, levelIdPara, constant.SEARCH_TYPE_LEVEL, true, "")
			if nil != err {
				logrus.WithFields(logrus.Fields{"error": err}).Error("ProcessFindUserByTypeFuzzy failed")
				return err
			}
			counter := 0
			for index := range userRecords {
				if userRecords[index].Id == parentIdPara {
					counter++
				}
			}
			if 0 == counter {
				logrus.WithFields(logrus.Fields{"parentId": userInfo.Parent, "levelId": userInfo.LevelId}).Error("add user, parent level not fit")
				return errcode.CerrParentLevelNotFit
			}
		} else {
			//代理用户层级为0（没有层级），有上级用户； 不能创建
			logrus.WithFields(logrus.Fields{"parentId": userInfo.Parent, "levelId": userInfo.LevelId}).Error("add user, parent level not fit")
			return errcode.CerrParentLevelNotFit
		}
	} else {
		userInfo.Parent = ""
		userInfo.ParentName = ""
	}

	if strings.ToUpper(userInfo.Phones.CountryCode) == "NULL" {
		userInfo.Phones.CountryCode = ""
	}
	if strings.ToUpper(userInfo.Phones.Phone) == "NULL" {
		userInfo.Phones.Phone = ""
	}
	if "" != userInfo.Phones.CountryCode || "" != userInfo.Phones.Phone {
		userInfo.Phone = userInfo.Phones.CountryCode + "@-@" + userInfo.Phones.Phone
	}

	return nil
}

//pwdStrength value is:PwdStrengthMiddle/PwdStrengthStrong/PwdStrengthSuperStrong
func GeneratePassword(pwdStrength int) string {
	var passwdLength uint
	switch pwdStrength {
	case constant.PwdStrengthMiddle, constant.PwdStrengthStrong:
		passwdLength = 8
	case constant.PwdStrengthSuperStrong:
		passwdLength = 8
	default:
		passwdLength = 8
	}
	return "Ab&" + GetRandomStringEnhance(passwdLength-2, 2)
}

/*
//生成随机字符串
func GetRandomString(randLen int) string{
	str := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < randLen; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}
*/
func GenerateEntityNo(tenant_id string, entityNo string) (string, code.Error) {
	if len(entityNo) > 0 {
		exist, errExist := ExistInUserDetailTab(tenant_id, constant.QueryKeyEntityNo, entityNo)
		if nil == errExist && false == exist {
			return entityNo, nil
		} else if nil != errExist {
			return "", errExist
		} else {
			return "", errcode.CerrEntityNoExist
		}
	}
	entityNo = strings.ToUpper(GetRandomStringEnhance(3, 1))

	counter := 0
	for counter < 3 {
		exist, errExist := ExistInUserDetailTab(tenant_id, constant.QueryKeyEntityNo, entityNo)
		if nil == errExist && false == exist {
			return entityNo, nil
		}
		if true == exist {
			entityNo = strings.ToUpper(GetRandomStringEnhance(3, 1))
		}
		counter++
	}
	return "", errcode.CerrEntityNoExist
}

func ExistInUserDetailTab(tenant_id string, key string, value interface{}) (bool, code.Error) {
	if "" == tenant_id || "" == key {
		return false, errcode.CerrParamater
	}
	dbConn, errConn := GetDBConnByTenantID(tenant_id)
	if errConn != nil {
		return false, errConn
	}

	var userRecord UserDetailTab
	var dbtmp *gorm.DB
	if constant.QueryKeyEmail == key {
		dbtmp = dbConn.Where("email = ? and model_status != ?", value, constant.ModelStatusDelete).First(&userRecord)
	} else if constant.QueryKeyPhone == key {
		dbtmp = dbConn.Where("phone = ? and model_status != ?", value, constant.ModelStatusDelete).First(&userRecord)
	} else if constant.QueryKeyEntityNo == key {
		dbtmp = dbConn.Where("entity_no = ? and model_status != ?", value, constant.ModelStatusDelete).First(&userRecord)
	} else if constant.QueryKeyLogin == key {
		dbtmp = dbConn.Where("login = ? and model_status != ?", value, constant.ModelStatusDelete).First(&userRecord)
	} else {
		logrus.WithFields(logrus.Fields{"key": key}).Error("Query Key Not Support")
		return false, errcode.CerrQueryKeyNotSupport
	}

	if err := dbtmp.Error; nil != err {
		if err == gorm.ErrRecordNotFound {
			logrus.WithFields(logrus.Fields{"key": key, "value": value}).Debug("Not found user record from DB")
			return false, nil //无记录不返回错误
		}
		logrus.WithFields(logrus.Fields{"error": err}).Error("Failed to get user record from DB by ids")
		return false, errcode.CerrExecuteSQL
	}

	logrus.WithFields(logrus.Fields{"key": key, "value": value}).Debug("Found records from DB")
	return true, nil
}

func BWUserDTO2UserDetailTabForAdd(userInfo *BWUserDTO, feignKey *XFeignKey) (*UserDetailTab, code.Error) {
	if nil == userInfo || nil == feignKey {
		logrus.WithFields(logrus.Fields{}).Debug("BWUserDTO2UserDetailTabForAdd parameter error")
		return nil, errcode.CerrParamater
	}
	var userTab UserDetailTab
	userTab.CreateDate = time.Now()
	userTab.CreateUserId = feignKey.UserId
	userTab.ModelStatus = constant.ModelStatusCreate
	userTab.ModifyDate = userTab.CreateDate
	userTab.ModifyUserId = feignKey.UserId
	userTab.Active = 1
	userTab.Birthday = userInfo.Birthday
	userTab.City = userInfo.City
	userTab.Comment = userInfo.Comment
	userTab.Country = userInfo.Country
	userTab.Address = userInfo.Address
	userTab.Email = userInfo.Email
	userTab.EntityNo = userInfo.EntityNo
	userTab.HeadImage = userInfo.HeadImage
	userTab.Nickname = userInfo.Nickname
	parentInt, errParse := strconv.ParseInt(userInfo.Parent, 10, 64)
	if nil == errParse && parentInt > 0 {
		userTab.ParentId = strconv.FormatInt(parentInt, 10)
	}
	userTab.Phone = userInfo.Phone
	if "" != userInfo.Phones.CountryCode && "" != userInfo.Phones.Phone {
		userTab.Phone = userInfo.Phones.CountryCode + "@-@" + userInfo.Phones.Phone
	}
	userTab.Postcode = userInfo.Postcode
	userTab.ProductId = "BW"
	userTab.Province = userInfo.Province
	userTab.RoleId, _ = strconv.ParseInt(userInfo.RoleId, 10, 64)
	userTab.RoleName = userInfo.RoleName
	userTab.Username = userInfo.UserName
	userTab.Name = userInfo.Name
	userTab.PubUserId = userInfo.PubUserId
	userTab.TenantId = userInfo.TenantId
	userTab.Sex = userInfo.Sex
	userTab.LevelId, _ = strconv.ParseInt(userInfo.LevelId, 10, 64)
	userTab.LevelName = userInfo.LevelName
	userTab.Login = ""
	if nil != userInfo.Login {
		userTab.Login = *userInfo.Login
	}
	userTab.VendorServerId = ""
	if nil != userInfo.VendorServerId {
		userTab.VendorServerId = *userInfo.VendorServerId
	}
	userTab.NeedInitPass = make([]uint8, 1)
	if true == userInfo.NeedInitPass {
		userTab.NeedInitPass[0] = 1
	} else {
		userTab.NeedInitPass[0] = 0
	}
	userTab.IdType = userInfo.IdType
	userTab.IdNum = userInfo.IdNum
	userTab.IdUrl1 = userInfo.IdUrl1
	userTab.IdUrl2 = userInfo.IdUrl2
	userTab.BankAccount = userInfo.BankAccount
	userTab.BankBranch = userInfo.BankBranch
	userTab.AccountNo = userInfo.AccountNo
	userTab.BankCardFile1 = userInfo.BankCardFile1
	userTab.BankCardFile2 = userInfo.BankCardFile2
	userTab.DoAgencyBusiness = userInfo.DoAgencyBusiness
	userTab.InvestExperience = userInfo.InvestExperience
	userTab.Agent = userInfo.Agent
	userTab.Field01 = userInfo.Field01
	userTab.Field02 = userInfo.Field02
	userTab.Field03 = userInfo.Field03
	userTab.Field04 = userInfo.Field04
	userTab.Field05 = userInfo.Field05
	userTab.Field06 = userInfo.Field06
	userTab.Field07 = userInfo.Field07
	userTab.Field08 = userInfo.Field08
	userTab.Field09 = userInfo.Field09
	userTab.Field10 = userInfo.Field10
	userTab.Field11 = userInfo.Field11
	userTab.Field12 = userInfo.Field12
	userTab.Field13 = userInfo.Field13
	userTab.Field14 = userInfo.Field14
	userTab.Field15 = userInfo.Field15
	userTab.Field16 = userInfo.Field16
	userTab.Field17 = userInfo.Field17
	userTab.Field18 = userInfo.Field18
	userTab.Field19 = userInfo.Field19
	userTab.Field20 = userInfo.Field20
	return &userTab, nil
}

//更新单个用户的单个字段, key: user_detail colume name, such as :key="active", value=1  /  key="name", value="name_test"
func ProcessUserUpdateOneFieldById(userIDNum int64, key string, value interface{}, feignKey *XFeignKey) code.Error {
	if userIDNum <= 0 || nil == feignKey || "" == key || nil == value {
		return errcode.CerrParamater
	}
	dbConn, errConn := GetDBConnByTenantID(feignKey.TenantId)
	if errConn != nil {
		return errConn
	}

	var userRecord UserDetailTab
	userRecord.Id = userIDNum
	dbtmp := dbConn.Model(&userRecord).Where("model_status != ?", constant.ModelStatusDelete).Update(key, value)
	if err := dbtmp.Error; nil != err {
		logrus.Errorf("Failed to update user record, userIDNum:%v, key:%v, value:%v, error:%v", userIDNum, key, value, err)
		return errcode.CerrExecuteSQL
	}
	return nil
}

//根据t_user_detail表中的主键id删除用户
func ProcessDeleteUserById(feignKey *XFeignKey, userIdNum int64) code.Error {
	tenant_id := feignKey.TenantId
	if "" == tenant_id || userIdNum < 0 {
		return errcode.CerrParamater
	}

	logrus.Infof("Start delete user:%v, operate user:%v", userIdNum, feignKey.UserId)
	if 1 == userIdNum {
		logrus.Errorf("userId = 1 is supper admin user ,not allow delete")
		return errcode.CerrForbidDelete
	}

	//获取直属下级用户信息
	childrenRecords, errGetChildren := GetChildrenUserByID(tenant_id, userIdNum)
	if nil != errGetChildren {
		logrus.Errorf("GetChildrenUserByID failed, userID:%v, error:%v", userIdNum, errGetChildren)
		return errGetChildren
	}
	if len(childrenRecords) > 0 {
		logrus.Errorf("userID:%v, have children user, not allow delete", userIdNum)
		return errcode.CerrForbidDeleteWithSubUser
	}

	//获取待删除的用户信息
	userDetail, errGetDetail := GetUserDetailByIdOrPubid(feignKey, "", userIdNum)
	if nil != errGetDetail {
		logrus.Errorf("Get user detail info failed, userId:%v, error:%v", userIdNum, errGetDetail)
		return errGetDetail
	}
	if nil == userDetail {
		logrus.Infof("Delete user:%v not exist, operate user:%v", userIdNum, feignKey.UserId)
		return nil
	}
	//向pub-auth请求删除用户
	var pubUser PubUserInfoDTO
	pubUser.UserId = userDetail.PubUserId
	_, errDeletePub := ClientDeletePubAuthUser(feignKey, &pubUser)
	if nil != errDeletePub {
		logrus.Errorf("ClientDeletePubAuthUser failed, userId:%v, pubUserId:%v, error:%v", userIdNum, userDetail.PubUserId, errDeletePub)
		return errDeletePub
	}

	//向report service请求删除返佣规则
	resultCommission, errCommission := ClienDeleteUserCommission(feignKey, userIdNum)
	if nil != errCommission || false == resultCommission {
		logrus.Warnf("Delete user commission failed, result:%v, error:%v", resultCommission, errCommission)
		//not return error
	}

	//在t_user_detail表中删除用户信息
	//暂时采用硬删除DB记录，后面通过修改状态实现////////////////
	dbConn, errConn := GetDBConnByTenantID(tenant_id)
	if errConn != nil {
		return errConn
	}
	var userRecordDelete UserDetailTab
	userRecordDelete.Id = userIdNum
	userRecordDelete.PubUserId = userDetail.PubUserId
	if errDelete := dbConn.Delete(&userRecordDelete).Error; nil != errDelete {
		logrus.Errorf("Delete user record failed, userID:%v, error:%v", userIdNum, errDelete)
		return errcode.CerrExecuteSQL
	}

	//更新下级用户数等计数
	go UpdageSubCounter(feignKey)
	logrus.Infof("Delete user:%v, operate user:%v", userIdNum, feignKey.UserId)
	return nil
}

//获取一个用户的直接下级用户信息
func GetChildrenUserByID(tenant_id string, userIdNum int64) ([]UserDetailTab, code.Error) {
	if "" == tenant_id || userIdNum <= 0 {
		return nil, errcode.CerrParamater
	}

	dbConn, errConn := GetDBConnByTenantID(tenant_id)
	if errConn != nil {
		return nil, errConn
	}

	userIdStr := strconv.FormatInt(userIdNum, 10)
	var userRecords []UserDetailTab = make([]UserDetailTab, 0)
	dbtmp := dbConn.Where("parent_id = ? and model_status != ?", userIdStr, constant.ModelStatusDelete).Find(&userRecords)

	if err := dbtmp.Error; nil != err {
		if err == gorm.ErrRecordNotFound {
			logrus.Infof("Not found childre records from DB, userID:%v", userIdNum)
			return nil, nil //未找到记录不返回错误
		}
		logrus.Errorf("Failed to get children records from DB, userId:%v", userIdNum)
		return nil, errcode.CerrExecuteSQL
	}

	return userRecords, nil
}

//添加用户处理，返回新添加用户的id
func ProcessAddAdminUser(feignKey *XFeignKey, name string, email string, password string, language string) (int64, code.Error) {
	//向pub-auth请求添加用户
	var pubUserInfo PubUserInfoDTO = PubUserInfoDTO{
		Mail:     email,
		Password: password,
	}
	pubUserId, errPubAuth := ClientAddPubAuthUser(feignKey, &pubUserInfo)
	if nil != errPubAuth {
		return 0, errPubAuth
	}

	//初始化Level
	initLevel(feignKey)
	//获取Admin初始化的role信息
	roleInfo, rightRecords, errGen := generateAdminRole(feignKey, language)
	if nil != errGen {
		logrus.WithFields(logrus.Fields{"error": errGen}).Error("generateAdminRole failed")
		return 0, errGen
	} else if nil == roleInfo || len(rightRecords) == 0 {
		logrus.WithFields(logrus.Fields{}).Error("generateAdminRole abnormal")
		return 0, errcode.CerrInternal
	}

	dbConn, errConn := GetDBConnByTenantID(feignKey.TenantId)
	if errConn != nil {
		return 0, errConn
	}
	//使用transaction保存记录
	tx := dbConn.Begin()
	if err := tx.Create(roleInfo).Error; nil != err { //插入role信息
		logrus.WithFields(logrus.Fields{"name": roleInfo.Name, "error": err}).Error("Add role info to DB failed")
		tx.Rollback()
		return 0, errcode.CerrExecuteSQL
	}
	//生成批量插入relation的sql
	logrus.WithFields(logrus.Fields{"id": roleInfo.Id, "name": roleInfo.Name}).Info("Add role info for admin user success, waite commit")
	insertRelationSqlStr := "INSERT INTO t_role_right_relation(t_role, role_rights) VALUES"
	roleIDStr := strconv.FormatInt(roleInfo.Id, 10)
	for index, right := range rightRecords {
		rightIDStr := strconv.FormatInt(right.Id, 10)
		insertRelationSqlStr += "(" + roleIDStr + ", " + rightIDStr + ")"
		if index < len(rightRecords)-1 {
			insertRelationSqlStr += ","
		}
	}
	insertRelationSqlStr += ";"
	logrus.WithFields(logrus.Fields{"insert relation sql": insertRelationSqlStr}).Debug("Add role right relation info to DB sql")

	if err := tx.Exec(insertRelationSqlStr).Error; nil != err { //插入relation，批量插入
		logrus.WithFields(logrus.Fields{"insert relation sql": insertRelationSqlStr, "error": err}).Error("Exec SQL failed")
		tx.Rollback()
		return 0, errcode.CerrExecuteSQL
	}
	//添加用户信息
	var userInfo UserDetailTab
	userInfo.PubUserId = pubUserId
	userInfo.Email = email
	userInfo.Name = name
	userInfo.RoleId = roleInfo.Id
	userInfo.RoleName = roleInfo.Name
	userInfo.NeedInitPass = []uint8{1}
	userInfo.Active = 1
	userInfo.CreateDate = time.Now()
	userInfo.ModifyDate = userInfo.CreateDate
	userInfo.ModifyUserId = feignKey.UserId
	userInfo.ModelStatus = constant.ModelStatusCreate
	userInfo.ProductId = "BW"
	enttityNo, errGenEntityNo := GenerateEntityNo(feignKey.TenantId, "")
	if nil != errGenEntityNo {
		logrus.WithFields(logrus.Fields{"error": errGenEntityNo}).Error("GenerateEntityNo failed")
		return 0, errGenEntityNo
	}
	userInfo.EntityNo = enttityNo
	if err := tx.Create(&userInfo).Error; nil != err { //插入admin用户信息
		logrus.WithFields(logrus.Fields{"name": name, "error": err}).Error("Add user info to DB failed")
		tx.Rollback()
		return 0, errcode.CerrExecuteSQL
	}

	if err := tx.Commit().Error; nil != err { //Transaction commit
		logrus.WithFields(logrus.Fields{"error": err, "name": name, "email": email}).Error("Add admin user Transaction commit failed")
		tx.Rollback()
		return 0, errcode.CerrExecuteSQL
	}

	logrus.WithFields(logrus.Fields{"name": name, "email": email, "id": userInfo.Id}).Info("Add admin user success")
	return userInfo.Id, nil
}

// 初始化角色列表
func ProcessInitRoleList(feignKey *XFeignKey, language string) code.Error {
	confRoleInfo := conf.GetApplication().RoleRights
	for index, _ := range confRoleInfo {
		roleInfo, rightRecords, errGen := generateCustomRole(feignKey, language, confRoleInfo[index])
		if nil != errGen {
			logrus.WithFields(logrus.Fields{"error": errGen}).Error("generateAdminRole failed")
			return errGen
		} else if nil == roleInfo || len(rightRecords) == 0 {
			logrus.WithFields(logrus.Fields{}).Error("generateAdminRole abnormal")
			return errcode.CerrInternal
		}
		dbConn, errConn := GetDBConnByTenantID(feignKey.TenantId)
		if errConn != nil {
			return errConn
		}
		//使用transaction保存记录
		tx := dbConn.Begin()
		if err := tx.Create(roleInfo).Error; nil != err { //插入role信息
			logrus.WithFields(logrus.Fields{"name": roleInfo.Name, "error": err}).Error("Add role info to DB failed")
			tx.Rollback()
			return errcode.CerrExecuteSQL
		}
		//生成批量插入relation的sql
		logrus.WithFields(logrus.Fields{"id": roleInfo.Id, "name": roleInfo.Name}).Info("Add role info for admin user success, waite commit")
		insertRelationSqlStr := "INSERT INTO t_role_right_relation(t_role, role_rights) VALUES"
		roleIDStr := strconv.FormatInt(roleInfo.Id, 10)
		for index, right := range rightRecords {
			rightIDStr := strconv.FormatInt(right.Id, 10)
			insertRelationSqlStr += "(" + roleIDStr + ", " + rightIDStr + ")"
			if index < len(rightRecords)-1 {
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

		if err := tx.Commit().Error; nil != err { //Transaction commit
			logrus.WithFields(logrus.Fields{"error": err, "name": roleInfo.Name}).Error("Add role Transaction commit failed")
			tx.Rollback()
			return errcode.CerrExecuteSQL
		}

		logrus.WithFields(logrus.Fields{"name": roleInfo.Name}).Info("Add role success")
	}
	return nil
}

//根据user表中的id列表，获取直接下级的用户信息
func GetBelongUserRecord(feignKey *XFeignKey, parentIds []int64) ([]UserDetailTab, code.Error) {
	if nil == feignKey || len(parentIds) == 0 {
		return nil, errcode.CerrParamater
	}
	dbConn, errConn := GetDBConnByTenantID(feignKey.TenantId)
	if errConn != nil {
		return nil, errConn
	}

	var userRecords []UserDetailTab = make([]UserDetailTab, 0)
	dbtmp := dbConn.Where("parent_id in (?) and model_status != ?", parentIds, constant.ModelStatusDelete).Find(&userRecords)
	if err := dbtmp.Error; nil != err {
		if err == gorm.ErrRecordNotFound {
			logrus.WithFields(logrus.Fields{"parentIds": parentIds}).Warn("Not found user records from DB by parentIds")
			return nil, nil //未找到记录不返回错误
		}
		logrus.WithFields(logrus.Fields{"error": err}).Error("Failed to get records from DB by parentIds")
		return nil, errcode.CerrExecuteSQL
	}

	return userRecords, nil
}

//获取某截止时间前创建的用户数,返回用户数和错误
func GetUserCountBeforeOneCreateTime(tenant_id string, deadTime time.Time) (int64, code.Error) {
	if "" == tenant_id {
		return 0, errcode.CerrParamater
	}
	dbConn, errConn := GetDBConnByTenantID(tenant_id)
	if errConn != nil {
		return 0, errConn
	}

	var total int64 = 0
	errDB := dbConn.Model(UserDetailTab{}).Where(" create_date < ? AND model_status != ?", deadTime, constant.ModelStatusDelete).Count(&total).Error
	if nil != errDB {
		logrus.WithFields(logrus.Fields{"deadTime time": deadTime, "error": errDB}).Error("Failed to get user count from DB")
		return 0, errcode.CerrExecuteSQL
	}
	return total, nil
}

//获取用户总数
func GetUserCount(tenant_id string) (int64, code.Error) {
	if "" == tenant_id {
		return 0, errcode.CerrParamater
	}
	dbConn, errConn := GetDBConnByTenantID(tenant_id)
	if errConn != nil {
		return 0, errConn
	}

	var total int64 = 0
	errDB := dbConn.Model(UserDetailTab{}).Where(" model_status != ?", constant.ModelStatusDelete).Count(&total).Error
	if nil != errDB {
		logrus.WithFields(logrus.Fields{"error": errDB}).Error("Failed to get user count from DB")
		return 0, errcode.CerrExecuteSQL
	}
	return total, nil
}

//获取所有的用户信息（数据较大，尽可能避免使用）
func GetAllUserRecords(feignKey *XFeignKey) ([]UserDetailTab, code.Error) {
	dbConn, errConn := GetDBConnByTenantID(feignKey.TenantId)
	if errConn != nil {
		return nil, errConn
	}

	var userRecords = make([]UserDetailTab, 0)
	dbtmp := dbConn.Where("model_status != ?", constant.ModelStatusDelete).Find(&userRecords)
	if err := dbtmp.Error; nil != err {
		if err == gorm.ErrRecordNotFound {
			logrus.WithFields(logrus.Fields{}).Warn("Not found user records from DB")
			return userRecords, nil //未找到记录不返回错误
		}
		logrus.WithFields(logrus.Fields{"error": err}).Error("Failed to get user records from DB")
		return nil, errcode.CerrExecuteSQL
	}
	return userRecords, nil
}

//获取所有的用户信息（数据较大，尽可能避免使用），返回结果map[int64][]UserDetailTab，key为parent_id, []UserDetailTab为相同parent的user信息
func getAllUserParentMapInfo(feignKey *XFeignKey) (map[int64][]UserDetailTab, code.Error) {
	userRecords, getUserErr := GetAllUserRecords(feignKey)
	if err := getUserErr; nil != err {
		logrus.WithFields(logrus.Fields{"error": err}).Error("Failed to get user records from DB")
		return nil, errcode.CerrExecuteSQL
	}

	return userListToUserParentMap(userRecords)
}

func userListToUserParentMap(userRecords []UserDetailTab) (map[int64][]UserDetailTab, code.Error) {
	var mapUserInfo = make(map[int64][]UserDetailTab, 0)
	//切片转换成map
	for index := range userRecords {
		var parentId int64 = 0
		if "" != userRecords[index].ParentId {
			id, errParse := strconv.ParseInt(userRecords[index].ParentId, 10, 64)
			if nil != errParse || id < 0 {
				logrus.WithFields(logrus.Fields{"error": errParse, "parentId": userRecords[index].ParentId, "user id": userRecords[index].Id}).Warn("one user parentId abnormal")
			}
			parentId = id
		}
		childList, ok := mapUserInfo[parentId]
		if false == ok {
			childList = make([]UserDetailTab, 0)
		}
		childList = append(childList, userRecords[index])
		mapUserInfo[parentId] = childList
	}

	return mapUserInfo, nil
}

//parentMapUserInfo的key为parentID
//返回的map，其key为用户的id
func ParentMapUser2IdMapUser(parentMapUserInfo map[int64][]UserDetailTab) map[int64]UserDetailTab {
	idMapUserInfo := make(map[int64]UserDetailTab, 0)
	for _, childList := range parentMapUserInfo {
		for _, userinfo := range childList {
			idMapUserInfo[userinfo.Id] = userinfo
		}
	}
	return idMapUserInfo
}

/*
//获取所有的用户信息（数据较大，尽可能避免使用），返回结果map[int64]UserDetailTab，key为UserId, UserDetailTab为用户信息
func getAllUserIdMapInfo(feignKey *XFeignKey) (map[int64]UserDetailTab, code.Error) {
	dbConn, errConn := GetDBConnByTenantID(feignKey.TenantId)
	if errConn != nil {
		return nil, errConn
	}

	var userRecords []UserDetailTab = make([]UserDetailTab, 0)
	var idMapUserInfo map[int64]UserDetailTab = make(map[int64]UserDetailTab, 0)
	dbtmp := dbConn.Where("model_status != ?", constant.ModelStatusDelete).Find(&userRecords)

	if err := dbtmp.Error; nil != err {
		if err == gorm.ErrRecordNotFound {
			logrus.WithFields(logrus.Fields{}).Warn("Not found user records from DB")
			return idMapUserInfo, nil				//未找到记录不返回错误
		}
		logrus.WithFields(logrus.Fields{"error": err,}).Error("Failed to get user records from DB")
		return nil, errcode.CerrExecuteSQL
	}

	//切片转换成map
	for index, _ := range userRecords {
		idMapUserInfo[userRecords[index].Id] = userRecords[index]
	}

	return idMapUserInfo, nil
}
*/

//用户列表信息转换为LazyTreeNodeDTO列表
func transitUserInfoToTreeNodeList(userList []UserDetailTab) []*LazyTreeNodeDTO {
	var treeNodes = make([]*LazyTreeNodeDTO, 0)
	for index, _ := range userList {
		var node LazyTreeNodeDTO
		node.Value = strconv.FormatInt(userList[index].Id, 10)
		node.Label = userList[index].Name
		node.Parent = userList[index].ParentId
		node.Child = false //false
		treeNodes = append(treeNodes, &node)
	}
	return treeNodes
}
func transitUserInfoToTreeNode(user *UserDetailTab) *LazyTreeNodeDTO {
	var node LazyTreeNodeDTO
	node.Value = strconv.FormatInt(user.Id, 10)
	node.Label = user.Name
	node.Parent = user.ParentId
	node.Child = false //false
	return &node
}
func transitUserDetailToTreeNode(user *UserDetail) *LazyTreeNodeDTO {
	var node LazyTreeNodeDTO
	node.Value = strconv.FormatInt(user.Id, 10)
	node.Label = user.Name
	node.Parent = user.Parent
	node.Child = false //false
	return &node
}

//设置treeNodes中的是否有child信息，mapUserInfo为所有用户的信息，其key为parentid
func setTreeNodesHasChild(treeNodes []*LazyTreeNodeDTO, mapUserInfo map[int64][]UserDetailTab) {
	for index, _ := range treeNodes {
		userId, errParse := strconv.ParseInt(treeNodes[index].Value, 0, 64)
		if errParse != nil {
			logrus.WithFields(logrus.Fields{"value": treeNodes[index].Value}).Warn("setTreeNodesHasChild one value abnormal")
			continue
		}
		_, existChild := mapUserInfo[userId] //查找当前节点的用户ID，是否存在child用户
		if existChild {
			treeNodes[index].Child = true
		} else {
			treeNodes[index].Child = false
		}
	}
}

//获取用户的下级用户信息信息，userIds，需获取的用户ID列表，mapUserInfo为所有用户的信息，其key为parentid
//searchScope取值范围SearchUserScope_ALL、SearchUserScope_NOT_BELONG、SearchUserScope_BELONG
func getChildrenUserList(userIds []int64, searchScope string, mapUserInfo map[int64][]UserDetailTab, depth int) ([]UserDetailTab, code.Error) {
	if depth > constant.LOOP_MAX_DEPTH {
		logrus.WithFields(logrus.Fields{"depth": depth}).Error("getChildrenUserList Loop Depth abnormal")
		return nil, errcode.CerrInternal
	}

	belongUsers := make([]UserDetailTab, 0) //直接直属下级用户信息
	for _, userid := range userIds {
		curChildUsers, exist := mapUserInfo[userid]
		if exist {
			belongUsers = append(belongUsers, curChildUsers...)
		}
	}
	if len(belongUsers) == 0 { //无直接下级信息，直接返回
		return belongUsers, nil
	}

	var belongIds []int64 = make([]int64, 0)
	for index, _ := range belongUsers {
		belongIds = append(belongIds, belongUsers[index].Id)
	}

	if constant.SearchUserScope_ALL == searchScope {
		notBelongChild, errGet := getChildrenUserList(belongIds, constant.SearchUserScope_ALL, mapUserInfo, depth+1) //非直属的所有下级用户
		if nil != errGet {
			return nil, errGet
		}
		return append(belongUsers, notBelongChild...), nil //直接下级+非直属下级
	} else if constant.SearchUserScope_NOT_BELONG == searchScope {
		notBelongChild, errGet := getChildrenUserList(belongIds, constant.SearchUserScope_ALL, mapUserInfo, depth+1) //非直属的所有下级用户
		if nil != errGet {
			return nil, errGet
		}
		return notBelongChild, nil
	} else if constant.SearchUserScope_BELONG == searchScope {
		return belongUsers, nil
	} else if constant.SearchUserScope_ONLY_NOT_BELONG == searchScope { //仅第一级非直属
		return getChildrenUserList(belongIds, constant.SearchUserScope_BELONG, mapUserInfo, depth+1) //直属下级用户
	}

	logrus.WithFields(logrus.Fields{"searchScope": searchScope}).Error("getChildrenUserList parameter error")
	return nil, errcode.CerrParamater
}

//mapUserInfo为所有的用户信息（key为parentId），currentRight为用户Normalized权限信息
func findUserTreeRoot(currentUser *UserDetail, currentRight []int, mapUserInfo map[int64][]UserDetailTab) ([]*LazyTreeNodeDTO, code.Error) {
	noParentUser, exist := mapUserInfo[0] //mapUserInfo[0]，即为无parent的用户信息
	if false == exist {
		noParentUser = make([]UserDetailTab, 0)
	}

	//所有权限，返回所有无parent的信息
	if IntListIsCotain(currentRight, constant.NORMALIZED_RIGHT_ALL) {
		result := transitUserInfoToTreeNodeList(noParentUser)
		setTreeNodesHasChild(result, mapUserInfo)
		return result, nil
	} else if IntListIsCotain(currentRight, constant.NORMALIZED_RIGHT_NO_PARENT) {
		result := transitUserInfoToTreeNodeList(noParentUser) //无parent的用户信息

		//如果有查看自己的权限
		if IntListIsCotain(currentRight, constant.NORMALIZED_RIGHT_MY) {
			var noParentHasCurrent = false
			var myNode *LazyTreeNodeDTO
			for index := range result {
				if result[index].Value == strconv.FormatInt(currentUser.Id, 10) {
					myNode = result[index]
					noParentHasCurrent = true
				}
			}

			if noParentHasCurrent == false {
				myNode = transitUserDetailToTreeNode(currentUser)
			}

			//根据是否有查看直属下级权限设置child字段
			if IntListIsCotain(currentRight, constant.NORMALIZED_RIGHT_DIRECT) {
				myNode.Child = len(mapUserInfo[currentUser.Id]) > 0
			} else {
				myNode.Child = false
			}

			if noParentHasCurrent == false {
				result = append(result, myNode)
			}

			//如果没有直属权限，但是有非直属权限，则需要把非直属放到根节点
			if !IntListIsCotain(currentRight, constant.NORMALIZED_RIGHT_DIRECT) && IntListIsCotain(currentRight, constant.NORMALIZED_RIGHT_SUBORDINATE) {
				noDirectUser, errGetNoDirect := getChildrenUserList([]int64{currentUser.Id}, constant.SearchUserScope_ONLY_NOT_BELONG, mapUserInfo, 1) //仅仅一级非直属下级信息
				if nil != errGetNoDirect {
					logrus.WithFields(logrus.Fields{"error": errGetNoDirect, "userId": currentUser.Id}).Error("getChildrenUserList failed")
					return nil, errGetNoDirect
				}
				noDirectUserTreeNode := transitUserInfoToTreeNodeList(noDirectUser)
				setTreeNodesHasChild(noDirectUserTreeNode, mapUserInfo)
				result = append(result, noDirectUserTreeNode...)
			}
		} else {
			//如果没有查看自己的权限，则只需要找直属和非直属放到根目录
			if IntListIsCotain(currentRight, constant.NORMALIZED_RIGHT_DIRECT) {
				directUser, errGetDirect := getChildrenUserList([]int64{currentUser.Id}, constant.SearchUserScope_BELONG, mapUserInfo, 1)
				if nil != errGetDirect {
					logrus.WithFields(logrus.Fields{"error": errGetDirect, "userId": currentUser.Id}).Error("getChildrenUserList failed")
					return nil, errGetDirect
				}

				directUserTreeNode := transitUserInfoToTreeNodeList(directUser)
				if IntListIsCotain(currentRight, constant.NORMALIZED_RIGHT_SUBORDINATE) {
					setTreeNodesHasChild(directUserTreeNode, mapUserInfo)
				} else {
					for index := range directUserTreeNode {
						directUserTreeNode[index].Child = false
					}
				}
				result = append(result, directUserTreeNode...)
			} else if IntListIsCotain(currentRight, constant.NORMALIZED_RIGHT_SUBORDINATE) {
				noDirectUser, errGetNoDirect := getChildrenUserList([]int64{currentUser.Id}, constant.SearchUserScope_ONLY_NOT_BELONG, mapUserInfo, 1) //仅仅一级非直属下级信息
				if nil != errGetNoDirect {
					logrus.WithFields(logrus.Fields{"error": errGetNoDirect, "userId": currentUser.Id}).Error("getChildrenUserList failed")
					return nil, errGetNoDirect
				}
				noDirectUserTreeNode := transitUserInfoToTreeNodeList(noDirectUser)
				setTreeNodesHasChild(noDirectUserTreeNode, mapUserInfo)
				result = append(result, noDirectUserTreeNode...)
			}
		}
		return result, nil
	} else if IntListIsCotain(currentRight, constant.NORMALIZED_RIGHT_MY) { //没有 没有 无归属权限，没有 所有权限，由My权限
		result := []*LazyTreeNodeDTO{transitUserDetailToTreeNode(currentUser)}
		if IntListIsCotain(currentRight, constant.NORMALIZED_RIGHT_DIRECT) {
			setTreeNodesHasChild(result, mapUserInfo)
		} else if IntListIsCotain(currentRight, constant.NORMALIZED_RIGHT_SUBORDINATE) {
			for index := range result {
				result[index].Child = false
			}
			noDirectUser, errGetNoDirect := getChildrenUserList([]int64{currentUser.Id}, constant.SearchUserScope_ONLY_NOT_BELONG, mapUserInfo, 1) //仅仅一级非直属下级信息
			if nil != errGetNoDirect {
				logrus.WithFields(logrus.Fields{"error": errGetNoDirect, "userId": currentUser.Id}).Error("getChildrenUserList failed")
				return nil, errGetNoDirect
			}
			noDirectUserTreeNode := transitUserInfoToTreeNodeList(noDirectUser)
			setTreeNodesHasChild(noDirectUserTreeNode, mapUserInfo)
			result = append(result, noDirectUserTreeNode...)
		}
		return result, nil
	} else if IntListIsCotain(currentRight, constant.NORMALIZED_RIGHT_DIRECT) {
		noDirectUser, errGetDirect := getChildrenUserList([]int64{currentUser.Id}, constant.SearchUserScope_BELONG, mapUserInfo, 1)
		if nil != errGetDirect {
			logrus.WithFields(logrus.Fields{"error": errGetDirect, "userId": currentUser.Id}).Error("getChildrenUserList failed")
			return nil, errGetDirect
		}

		result := transitUserInfoToTreeNodeList(noDirectUser)
		if IntListIsCotain(currentRight, constant.NORMALIZED_RIGHT_SUBORDINATE) {
			setTreeNodesHasChild(result, mapUserInfo)
		}
		return result, nil
	} else if IntListIsCotain(currentRight, constant.NORMALIZED_RIGHT_SUBORDINATE) {
		childUsers, errGetChild := getChildrenUserList([]int64{currentUser.Id}, constant.SearchUserScope_ONLY_NOT_BELONG, mapUserInfo, 1) //第一级非直属下级信息
		if nil != errGetChild {
			logrus.WithFields(logrus.Fields{"error": errGetChild, "userId": currentUser.Id}).Error("getChildrenUserList failed")
			return nil, errGetChild
		}
		childTreeNode := transitUserInfoToTreeNodeList(childUsers)
		setTreeNodesHasChild(childTreeNode, mapUserInfo)
		return childTreeNode, nil
	}

	return []*LazyTreeNodeDTO{}, nil
}

//获取当前登录用户的可查看权限ID范围，若返回nil，则当前用户有所有ID的权限
//mapUserInfo为所有的用户信息（key为parentId），currentRight为用户Normalized权限信息
func getUserIdPermissionScope(currentUserId int64, currentRight []int, mapUserInfo map[int64][]UserDetailTab) ([]int64, code.Error) {
	if IntListIsCotain(currentRight, constant.NORMALIZED_RIGHT_ALL) {
		return nil, nil
	}

	var permissionIds []int64 = make([]int64, 0)
	if IntListIsCotain(currentRight, constant.NORMALIZED_RIGHT_NO_PARENT) {
		noParentUser, exist := mapUserInfo[0] //mapUserInfo[0]，即为无parent的用户信息
		if exist {
			for index, _ := range noParentUser {
				permissionIds = append(permissionIds, noParentUser[index].Id) //添加无parent的用户ID
			}
		}
	}
	if IntListIsCotain(currentRight, constant.NORMALIZED_RIGHT_MY) {
		permissionIds = append(permissionIds, currentUserId)
	}
	if IntListIsCotain(currentRight, constant.NORMALIZED_RIGHT_DIRECT) {
		childUsers, errGetChild := getChildrenUserList([]int64{currentUserId}, constant.SearchUserScope_BELONG, mapUserInfo, 1) //直属下级信息
		if nil != errGetChild {
			logrus.WithFields(logrus.Fields{"error": errGetChild, "userId": currentUserId}).Error("getChildrenUserList failed")
			return nil, errGetChild
		}
		for index, _ := range childUsers {
			permissionIds = append(permissionIds, childUsers[index].Id) //添加直属下级用户
		}
	}
	if IntListIsCotain(currentRight, constant.NORMALIZED_RIGHT_SUBORDINATE) {
		childUsers, errGetChild := getChildrenUserList([]int64{currentUserId}, constant.SearchUserScope_NOT_BELONG, mapUserInfo, 1) //非直属下级
		if nil != errGetChild {
			logrus.WithFields(logrus.Fields{"error": errGetChild, "userId": currentUserId}).Error("getChildrenUserList failed")
			return nil, errGetChild
		}
		for index, _ := range childUsers {
			permissionIds = append(permissionIds, childUsers[index].Id) //添加非直属下级用户
		}
	}
	//logrus.WithFields(logrus.Fields{"permissionIds": permissionIds, "currentUserId": currentUserId}).Debug("getUserIdPermissionScope")
	return permissionIds, nil
}

//用户树查询下级信息接口, 根据不同权限和模块显示不同返回
func ProcessUserTreeChildByModuleRight(feignKey *XFeignKey, targetUserId int64, module string) ([]*LazyTreeNodeDTO, code.Error) {
	if nil == feignKey || "" == feignKey.TenantId || "" == feignKey.UserId || "" == module {
		return nil, errcode.CerrParamater
	}
	//获取当前用户信息
	currentUser, errCurrent := GetUserDetailByIdOrPubid(feignKey, feignKey.UserId, 0)
	if nil != errCurrent {
		logrus.WithFields(logrus.Fields{"pubUserid": feignKey.UserId, "error": errCurrent}).Error("Get current user info failed")
		return nil, errCurrent
	} else if nil == currentUser {
		logrus.WithFields(logrus.Fields{"pubUserid": feignKey.UserId}).Error("Get current user info not exist")
		return nil, errcode.CerrUserNotFound
	}
	//获取当前用户角色的right信息
	userRights, errGetRight := GetOneRoleRightList(feignKey, currentUser.RoleId)
	if nil != errGetRight {
		logrus.WithFields(logrus.Fields{"roleID": currentUser.RoleId, "error": errGetRight}).Error("Get current user right info failed")
		return nil, errGetRight
	}

	//转换为Normalized权限信息
	currentRight, errNormalize := normalizeUserTreeRight(userRights, module)
	if nil != errNormalize {
		return nil, errNormalize
	}
	//logrus.WithFields(logrus.Fields{"currentRight":currentRight, "roleName": currentUser.RoleName}).Error("current user normalize Right")/////
	//获取所有的用户信息
	mapUserInfo, errGetAllUser := getAllUserParentMapInfo(feignKey)
	if nil != errGetAllUser {
		return nil, errGetAllUser
	}

	return getChildrenByRight(currentUser, currentRight, mapUserInfo, targetUserId)
}

func getChildrenByRight(currentUser *UserDetail, normalizeRights []int, allUserChildrenMap map[int64][]UserDetailTab, targetUserId int64) ([]*LazyTreeNodeDTO, code.Error) {

	//查询根节点
	if 0 == targetUserId {
		treeNodes, errFindTree := findUserTreeRoot(currentUser, normalizeRights, allUserChildrenMap)
		if nil != errFindTree {
			logrus.WithFields(logrus.Fields{"error": errFindTree, "currentRight": normalizeRights, "currentUserId": currentUser.Id}).Error("findUserTreeRoot failed")
			return nil, errFindTree
		}
		return treeNodes, nil
	}

	//子节点查询
	//获取当前登录用户的可查看权限ID范围，若permissionIds为nil，则当前用户有所有ID的权限
	permissionIds, errGetPermission := getUserIdPermissionScope(currentUser.Id, normalizeRights, allUserChildrenMap)
	if nil != errGetPermission {
		logrus.WithFields(logrus.Fields{"error": errGetPermission}).Error("getUserIdPermissionScope failed")
		return nil, errGetPermission
	}
	if permissionIds != nil && false == Int64ListIsCotain(permissionIds, targetUserId) {
		return []*LazyTreeNodeDTO{}, nil
	}
	//查询targetUserId的直接下级用户
	targetChildUsers, errGetChild := getChildrenUserList([]int64{targetUserId}, constant.SearchUserScope_BELONG, allUserChildrenMap, 1) //直属下级信息
	if nil != errGetChild {
		logrus.WithFields(logrus.Fields{"error": errGetChild, "userId": targetUserId}).Error("getChildrenUserList failed")
		return nil, errGetChild
	}

	var havePermissionTargetChild = make([]UserDetailTab, 0)
	for index := range targetChildUsers {
		if nil == permissionIds || Int64ListIsCotain(permissionIds, targetChildUsers[index].Id) { //permissionIds为nil表示当前用户对所有id有权限
			havePermissionTargetChild = append(havePermissionTargetChild, targetChildUsers[index])
		}
	}
	childTreeNodes := transitUserInfoToTreeNodeList(havePermissionTargetChild) //childTreeNodes是直属下级，若有非直属下级权限才检查是否有child
	if IntListIsCotain(normalizeRights, constant.NORMALIZED_RIGHT_SUBORDINATE) || IntListIsCotain(normalizeRights, constant.NORMALIZED_RIGHT_ALL) {
		setTreeNodesHasChild(childTreeNodes, allUserChildrenMap)
	}
	return childTreeNodes, nil
}

//用户树查询下级信息接口
func ProcessUserTree(feignKey *XFeignKey, targetUserId int64) ([]*LazyTreeNodeDTO, code.Error) {
	if nil == feignKey || "" == feignKey.TenantId || "" == feignKey.UserId {
		return nil, errcode.CerrParamater
	}

	mapParentUserInfo, errGetAllUser := getAllUserParentMapInfo(feignKey)
	if nil != errGetAllUser {
		return nil, errGetAllUser
	}

	return getSubUserTree(targetUserId, mapParentUserInfo), nil
}

func BuildTargetUserTreeByRight(feignKey *XFeignKey, targetUserId int64, module string) ([]*LazyTreeNodeDTO, code.Error) {
	if nil == feignKey || "" == feignKey.TenantId || "" == feignKey.UserId {
		return nil, errcode.CerrParamater
	}

	//获取当前用户信息
	currentUser, errCurrent := GetUserDetailByIdOrPubid(feignKey, feignKey.UserId, 0)
	if nil != errCurrent {
		logrus.WithFields(logrus.Fields{"pubUserid": feignKey.UserId, "error": errCurrent}).Error("Get current user info failed")
		return nil, errCurrent
	} else if nil == currentUser {
		logrus.WithFields(logrus.Fields{"pubUserid": feignKey.UserId}).Error("Get current user info not exist")
		return nil, errcode.CerrUserNotFound
	}

	//获取当前用户角色的right信息
	userRights, errGetRight := GetOneRoleRightList(feignKey, currentUser.RoleId)
	if nil != errGetRight {
		logrus.WithFields(logrus.Fields{"roleID": currentUser.RoleId, "error": errGetRight}).Error("Get current user right info failed")
		return nil, errGetRight
	}

	//转换为Normalized权限信息
	normalizeRights, errNormalize := normalizeUserTreeRight(userRights, module)
	if nil != errNormalize {
		return nil, errNormalize
	}

	mapParentUserInfo, errGetAllUser := getAllUserParentMapInfo(feignKey)
	if nil != errGetAllUser {
		return nil, errGetAllUser
	}

	rootNodes, errFindTree := findUserTreeRoot(currentUser, normalizeRights, mapParentUserInfo)
	if nil != errFindTree {
		logrus.WithFields(logrus.Fields{"error": errFindTree, "currentRight": normalizeRights, "currentUserId": currentUser.Id}).Error("findUserTreeRoot failed")
		return nil, errFindTree
	}

	var rootContain = false
	for index := range rootNodes {
		if rootNodes[index].Value == strconv.FormatInt(targetUserId, 10) {
			rootContain = true
			rootNodes[index].Selected = true
		}
	}

	//如果ROOT包含，则返回根节点
	if rootContain {
		return rootNodes, nil
	}

	userIdToInfoMap := ParentMapUser2IdMapUser(mapParentUserInfo)

	//搜索指定用户所有上级, ID从前到后从底层到高层
	allParentIds := make([]string, 0)
	allParentIds = append(allParentIds, strconv.FormatInt(targetUserId, 10))
	getAllParentIds(userIdToInfoMap, targetUserId, &allParentIds)

	buildChildrenTree(currentUser, normalizeRights, mapParentUserInfo, allParentIds, 0, rootNodes)

	return rootNodes, nil
}

func getAllParentIds(idInfoMap map[int64]UserDetailTab, targetUserId int64, parentIds *[]string) {
	if user, ok := idInfoMap[targetUserId]; ok {
		if user.ParentId != "" {
			*parentIds = append(*parentIds, user.ParentId)

			parentIdInt, _ := strconv.ParseInt(user.ParentId, 10, 64)
			getAllParentIds(idInfoMap, parentIdInt, parentIds)
		}
	}
}

func buildChildrenTree(currentUser *UserDetail, normalizeRights []int, mapParentUserInfo map[int64][]UserDetailTab,
	allPathIds []string, curPathDepth int, root []*LazyTreeNodeDTO) {
	curIndex := len(allPathIds) - curPathDepth - 1
	if curIndex < 0 {
		return
	}

	targetId := allPathIds[curIndex]
	for index := range root {
		if root[index].Value == targetId {
			root[index].Selected = true

			targetUserIdInt, _ := strconv.ParseInt(targetId, 10, 64)
			children, getChildrenErr := getChildrenByRight(currentUser, normalizeRights, mapParentUserInfo, targetUserIdInt)
			if getChildrenErr != nil {
				return
			}
			root[index].Children = children

			buildChildrenTree(currentUser, normalizeRights, mapParentUserInfo, allPathIds, curPathDepth+1, children)
		}
	}
}

func getSubUserTree(targetUserId int64, mapParentUserInfo map[int64][]UserDetailTab) []*LazyTreeNodeDTO {
	var result []*LazyTreeNodeDTO
	if children, ok := mapParentUserInfo[targetUserId]; ok {
		for index := range children {
			user := children[index]
			child := transitUserInfoToTreeNode(&user)
			child.Children = getSubUserTree(user.Id, mapParentUserInfo)
			result = append(result, child)
		}
	}
	return result
}

//为用户筛选生成where语句，返回where部分的sql
func getSimpleWhereConditionForUserSearch(feignKey *XFeignKey, searchReq *UserSearchDTO) (string, code.Error) {
	whereSql := "model_status != '" + constant.ModelStatusDelete + "' "
	if "" == searchReq.QueryContent {
		return whereSql, nil
	}
	var intQueryContent int64 = -1
	if constant.QueryType_ID == searchReq.QueryType || constant.QueryType_LEVEL == searchReq.QueryType {
		var errParse error
		intQueryContent, errParse = strconv.ParseInt(searchReq.QueryContent, 10, 64)
		if nil != errParse || intQueryContent < 0 {
			logrus.WithFields(logrus.Fields{"error": errParse, "QueryContent": searchReq.QueryContent, "QueryType": searchReq.QueryType}).Error("getSimpleWhereConditionForUserSearch parameter queryContent error")
			return "", errcode.CerrParamater
		}
	}

	switch searchReq.QueryType {
	case constant.QueryType_ID:
		whereSql += fmt.Sprintf(" and id = %v ", intQueryContent)
	case constant.QueryType_LEVEL:
		whereSql += fmt.Sprintf(" and level_id = %v ", intQueryContent)
	case constant.QueryType_ENTITY_NO:
		whereSql += fmt.Sprintf(" and entity_no like '%%%v%%' ", escapeSql(searchReq.QueryContent))
	case constant.QueryType_NAME:
		whereSql += fmt.Sprintf(" and name like '%%%v%%' ", escapeSql(searchReq.QueryContent))
	case constant.QueryType_PHONE:
		whereSql += fmt.Sprintf(" and phone like '%%%v%%' ", escapeSql(searchReq.QueryContent))
	case constant.QueryType_EMAIL:
		whereSql += fmt.Sprintf(" and email like '%%%v%%' ", escapeSql(searchReq.QueryContent))
	case constant.QueryType_ROLE: //根据roleName筛选
		whereSql += fmt.Sprintf(" and role_name like '%%%v%%' ", escapeSql(searchReq.QueryContent))
	case constant.QueryType_LOGIN: //根据roleName筛选
		whereSql += fmt.Sprintf(" and login like '%%%v%%' ", escapeSql(searchReq.QueryContent))
	case constant.QueryType_PARENT: //根据parent的name筛选
		dbConn, errConn := GetDBConnByTenantID(feignKey.TenantId)
		if errConn != nil {
			return "", errConn
		}
		var userRecords []UserDetailTab = make([]UserDetailTab, 0)
		//根据like parent_name查询用户信息
		dbtmp := dbConn.Where("name like ? and model_status != ?", "%"+escapeSql(searchReq.QueryContent)+"%", constant.ModelStatusDelete).Find(&userRecords)
		if err := dbtmp.Error; nil != err {
			if err == gorm.ErrRecordNotFound {
				logrus.WithFields(logrus.Fields{"name": searchReq.QueryContent}).Infof("Not found user records by like name")
				//未找到记录不返回
			} else {
				logrus.WithFields(logrus.Fields{"name": searchReq.QueryContent, "error": err}).Error("Qyery user records by like name failed")
				return "", errcode.CerrExecuteSQL
			}
		}

		parentIdList := "'-2838383'" //添加一个不可能存在的ID，避免传空集合
		for index, _ := range userRecords {
			parentIdList += fmt.Sprintf(",'%v'", userRecords[index].Id)
		}
		whereSql += fmt.Sprintf(" and parent_id in (%v) ", parentIdList)
	default:

	}
	logrus.WithFields(logrus.Fields{"QueryContent": searchReq.QueryContent, "QueryType": searchReq.QueryType, "whereSql": whereSql}).Debug("Generate where sql")
	return whereSql, nil
}

//为用户筛选生成where语句，返回where部分的sql
func getSimpleWhereConditionForUserSearchV2(feignKey *XFeignKey, searchReq *UserSearchDTO) (string, code.Error) {
	whereSql := "model_status != '" + constant.ModelStatusDelete + "' "
	if "" == searchReq.QueryContent {
		return whereSql, nil
	}

	if searchReq.QueryContent != "" {
		whereSql += "and ("
		//编码
		whereSql += fmt.Sprintf(" entity_no like '%%%v%%' ", escapeSql(searchReq.QueryContent))
		//姓名
		whereSql += fmt.Sprintf(" or name like '%%%v%%' ", escapeSql(searchReq.QueryContent))
		//上级用户姓名
		dbConn, errConn := GetDBConnByTenantID(feignKey.TenantId)
		if errConn != nil {
			return "", errConn
		}
		var userRecords []UserDetailTab = make([]UserDetailTab, 0)
		//根据like parent_name查询用户信息
		dbtmp := dbConn.Where("name like ? and model_status != ?", "%"+escapeSql(searchReq.QueryContent)+"%", constant.ModelStatusDelete).Find(&userRecords)
		if err := dbtmp.Error; nil != err {
			if err == gorm.ErrRecordNotFound {
				logrus.WithFields(logrus.Fields{"name": searchReq.QueryContent}).Infof("Not found user records by like name")
				//未找到记录不返回
			}
			logrus.WithFields(logrus.Fields{"name": searchReq.QueryContent, "error": err}).Error("Qyery user records by like name failed")
			return "", errcode.CerrExecuteSQL
		}

		parentIdList := "'-2838383'" //添加一个不可能存在的ID，避免传空集合
		for index, _ := range userRecords {
			parentIdList += fmt.Sprintf(",'%v'", userRecords[index].Id)
		}
		whereSql += fmt.Sprintf(" or parent_id in (%v) ", parentIdList)

		whereSql += ")"
	}
	logrus.WithFields(logrus.Fields{"QueryContent": searchReq.QueryContent, "QueryType": searchReq.QueryType, "whereSql": whereSql}).Debug("Generate where sql")
	return whereSql, nil
}

func escapeSql(input string) string {
	result := strings.Replace(input, "'", "\\'", -1)
	result = strings.Replace(result, "\"", "\\\"", -1)
	return result
}

//根据条件查询用户的简单信息（含id,name,parent,levelId,levelName,roleId,roleName,entityNo,pubUserId,login,vendorServerId）
//目前只支持QueryType、QueryContent、PageNo、Size四个字段搜索
func ProcessSimpleUserByPage(feignKey *XFeignKey, searchReq *UserSearchDTO) (*SimpleUserPageDTO, code.Error) {
	if nil == feignKey || "" == feignKey.TenantId || nil == searchReq || searchReq.PageNo <= 0 || searchReq.Size <= 0 {
		return nil, errcode.CerrParamater
	}

	whereSql, errGetSql := getSimpleWhereConditionForUserSearch(feignKey, searchReq)
	if nil != errGetSql {
		logrus.WithFields(logrus.Fields{"error": errGetSql}).Error("getSimpleWhereConditionForUserSearch failed")
		return nil, errGetSql
	}
	dbConn, errConn := GetDBConnByTenantID(feignKey.TenantId)
	if errConn != nil {
		return nil, errConn
	}

	var total int64 = 0
	if errCount := dbConn.Model(&UserDetailTab{}).Where(whereSql).Count(&total).Error; nil != errCount {
		logrus.WithFields(logrus.Fields{"error": errCount, "whereSql": whereSql}).Error("get user count failed")
		return nil, errGetSql
	}
	logrus.WithFields(logrus.Fields{"whereSql": whereSql, "tenantId": feignKey.TenantId, "total": total}).Debug("total user count")

	var response SimpleUserPageDTO
	response.Total = total
	response.Size = searchReq.Size
	response.Pager = searchReq.PageNo
	response.Offset = (searchReq.PageNo - 1) * searchReq.Size
	if int(total)%searchReq.Size == 0 {
		response.Pages = int(total) / searchReq.Size
	} else {
		response.Pages = int(total)/searchReq.Size + 1
	}
	response.List = make([]SimpleUserDTO, 0)

	var userRecords []UserDetailTab = make([]UserDetailTab, 0)
	dbtmp := dbConn.Where(whereSql).Offset(response.Offset).Limit(response.Size).Find(&userRecords)
	if err := dbtmp.Error; nil != err {
		if err == gorm.ErrRecordNotFound {
			logrus.WithFields(logrus.Fields{"whereSql": whereSql, "offset": response.Offset, "limit": response.Size}).Infof("Not found user records from DB")
			return &response, nil //未找到记录不返回错误
		}
		logrus.WithFields(logrus.Fields{"whereSql": whereSql, "offset": response.Offset, "limit": response.Size}).Infof("Get user records from DB failed")
		return nil, errcode.CerrExecuteSQL
	}

	for index, _ := range userRecords {
		simpleUser, errConv := UserRecord2SimpleUser(&userRecords[index])
		if nil != errConv || nil == simpleUser {
			logrus.WithFields(logrus.Fields{"error": errConv}).Warn("UserRecord2SimpleUser abnormal")
			continue
		}
		response.List = append(response.List, *simpleUser)
	}

	return &response, nil
}

func getWhereConditionFindUserByField(feignKey *XFeignKey, searchReq *UserFieldSearchDTO) (string, code.Error) {
	whereSql := "model_status != '" + constant.ModelStatusDelete + "' and (1 != 1 "
	for _, field := range searchReq.FieldTypes {
		switch field {
		case constant.QueryKeyEmail:
			whereSql += fmt.Sprintf(" or email like '%%%v%%' ", escapeSql(searchReq.FuzzyValue))
		case constant.QueryKeyName:
			whereSql += fmt.Sprintf(" or name like '%%%v%%' ", escapeSql(searchReq.FuzzyValue))
		case constant.QueryKeyEntityNo:
			whereSql += fmt.Sprintf(" or entity_no like '%%%v%%' ", escapeSql(searchReq.FuzzyValue))
		case constant.QueryKeyLogin:
			whereSql += fmt.Sprintf(" or login like '%%%v%%' ", escapeSql(searchReq.FuzzyValue))
		case constant.QueryKeyId:
			intValue, errparse := strconv.ParseInt(searchReq.FuzzyValue, 10, 64)
			if nil != errparse {
				logrus.WithFields(logrus.Fields{"error": errparse, "FuzzyValue": searchReq.FuzzyValue, "field": field}).Error("getWhereConditionFindUserByField parameter abnormal")
				return "", errcode.CerrParamater
			}
			whereSql += fmt.Sprintf(" or id = %v ", intValue)
		case constant.QueryKeyRoleId:
			intValue, errparse := strconv.ParseInt(searchReq.FuzzyValue, 10, 64)
			if nil != errparse {
				logrus.WithFields(logrus.Fields{"error": errparse, "FuzzyValue": searchReq.FuzzyValue, "field": field}).Error("getWhereConditionFindUserByField parameter abnormal")
				return "", errcode.CerrParamater
			}
			whereSql += fmt.Sprintf(" or role_id = %v ", intValue)
		case constant.QueryKeyLevelId:
			intValue, errparse := strconv.ParseInt(searchReq.FuzzyValue, 10, 64)
			if nil != errparse {
				logrus.WithFields(logrus.Fields{"error": errparse, "FuzzyValue": searchReq.FuzzyValue, "field": field}).Error("getWhereConditionFindUserByField parameter abnormal")
				return "", errcode.CerrParamater
			}
			whereSql += fmt.Sprintf(" or level_id = %v ", intValue)
		default:
			logrus.WithFields(logrus.Fields{"FuzzyValue": searchReq.FuzzyValue, "field": field}).Error("getWhereConditionFindUserByField parameter abnormal")
			return "", errcode.CerrParamater
		}
	}
	whereSql += ")"

	return whereSql, nil
}

//模糊搜索查询用户指定字段值
func ProcessFindUserByField(feignKey *XFeignKey, searchReq *UserFieldSearchDTO, module string) ([]SimpleUserDTO, code.Error) {
	if nil == feignKey || "" == feignKey.TenantId || nil == searchReq || "" == searchReq.FuzzyValue || len(searchReq.FieldTypes) == 0 {
		return nil, errcode.CerrParamater
	}

	currentUser, errGetUser := GetUserDetailByIdOrPubid(feignKey, feignKey.UserId, 0)
	if nil != errGetUser {
		logrus.WithFields(logrus.Fields{"pubUserId": feignKey.UserId, "error": errGetUser}).Error("Get current user info failed")
		return nil, errGetUser
	} else if nil == currentUser {
		return nil, errcode.CerrAccountNotFound
	}
	//获取当前用户角色的right信息
	userRights, errGetRight := GetOneRoleRightList(feignKey, currentUser.RoleId)
	if nil != errGetRight {
		logrus.WithFields(logrus.Fields{"roleID": currentUser.RoleId, "error": errGetRight}).Error("Get current user right info failed")
		return nil, errGetRight
	}
	//获取所有的用户信息, map的key为parent_id
	mapUserInfo, errGetAllUser := getAllUserParentMapInfo(feignKey)
	if nil != errGetAllUser {
		return nil, errGetAllUser
	}
	//转换为Normalized权限信息
	currentRight, errNormalize := normalizeUserTreeRight(userRights, module)
	if nil != errNormalize {
		return nil, errNormalize
	}
	//用户权限可见范围
	permissionIds, errGetPermission := getUserIdPermissionScope(currentUser.Id, currentRight, mapUserInfo)
	if nil != errGetPermission {
		logrus.WithFields(logrus.Fields{"error": errGetPermission}).Error("getUserIdPermissionScope failed")
		return nil, errGetPermission
	}
	if nil != permissionIds { //permissionIds为nil表示有所有用户权限
		permissionIds = append(permissionIds, currentUser.Id) //添加用户自己有权限
	}

	whereSql, errGetSql := getWhereConditionFindUserByField(feignKey, searchReq)
	if nil != errGetSql {
		logrus.WithFields(logrus.Fields{"error": errGetSql}).Error("getWhereConditionFindUserByField failed")
		return nil, errGetSql
	}
	dbConn, errConn := GetDBConnByTenantID(feignKey.TenantId)
	if errConn != nil {
		return nil, errConn
	}

	var userRecords = make([]UserDetailTab, 0)
	dbtmp := dbConn.Where(whereSql).Find(&userRecords)
	if err := dbtmp.Error; nil != err {
		if err == gorm.ErrRecordNotFound {
			logrus.WithFields(logrus.Fields{"whereSql": whereSql}).Infof("Not found user records from DB")
			return nil, nil //未找到记录不返回错误
		}
		logrus.WithFields(logrus.Fields{"whereSql": whereSql, "error": err}).Infof("Get user records from DB failed")
		return nil, errcode.CerrExecuteSQL
	}

	var simpleList = make([]SimpleUserDTO, 0)
	for index, _ := range userRecords {
		if nil != permissionIds && false == Int64ListIsCotain(permissionIds, userRecords[index].Id) {
			continue //permissionIds为nil表示有所有用户权限
		}
		simpleUser, errConv := UserRecord2SimpleUser(&userRecords[index])
		if nil != errConv || nil == simpleUser {
			logrus.WithFields(logrus.Fields{"error": errConv}).Warn("UserRecord2SimpleUser abnormal")
			continue
		}
		simpleList = append(simpleList, *simpleUser)
	}

	return simpleList, nil
}

//根据模块检查指定用户ID是否在权限范围内
func ProcessCheckUserIdPermissionScope(feignKey *XFeignKey, targetUserId int64, module string) (bool, code.Error) {
	if nil == feignKey || "" == feignKey.TenantId || targetUserId <= 0 || "" == module {
		return false, errcode.CerrParamater
	}

	currentUser, errGetUser := GetUserRecordByIdOrPubid(feignKey, feignKey.UserId, 0)
	if nil != errGetUser {
		logrus.WithFields(logrus.Fields{"pubUserId": feignKey.UserId, "error": errGetUser}).Error("Get current user info failed")
		return false, errGetUser
	} else if nil == currentUser {
		return false, errcode.CerrAccountNotFound
	}
	//获取当前用户角色的right信息
	userRights, errGetRight := GetOneRoleRightList(feignKey, currentUser.RoleId)
	if nil != errGetRight {
		logrus.WithFields(logrus.Fields{"roleID": currentUser.RoleId, "error": errGetRight}).Error("Get current user right info failed")
		return false, errGetRight
	}
	//获取所有的用户信息, map的key为parent_id
	mapUserInfo, errGetAllUser := getAllUserParentMapInfo(feignKey)
	if nil != errGetAllUser {
		return false, errGetAllUser
	}
	//转换为Normalized权限信息
	currentRight, errNormalize := normalizeUserTreeRight(userRights, module)
	if nil != errNormalize {
		return false, errNormalize
	}
	//用户权限可见范围
	permissionIds, errGetPermission := getUserIdPermissionScope(currentUser.Id, currentRight, mapUserInfo)
	if nil != errGetPermission {
		logrus.WithFields(logrus.Fields{"error": errGetPermission}).Error("getUserIdPermissionScope failed")
		return false, errGetPermission
	}

	if nil == permissionIds || Int64ListIsCotain(permissionIds, targetUserId) { //permissionIds为nil，表示有所有权限
		return true, nil //有权限
	}
	return false, nil
}

func getWhereConditionFindLikeNameWithRight(feignKey *XFeignKey, searchReq *UserNameSearchDTO) (string, code.Error) {
	if nil == searchReq || nil == searchReq || "" == searchReq.FuzzyValue ||
		(false == searchReq.SearchRole && false == searchReq.SearchLevel && false == searchReq.SearchUser) {
		return "", errcode.CerrParamater
	}
	whereSql := "model_status != '" + constant.ModelStatusDelete + "' and (1 != 1 "
	if searchReq.SearchUser {
		whereSql += fmt.Sprintf(" or name like '%%%v%%' ", escapeSql(searchReq.FuzzyValue))
	}
	if searchReq.SearchLevel {
		whereSql += fmt.Sprintf(" or level_name like '%%%v%%' ", escapeSql(searchReq.FuzzyValue))
	}
	if searchReq.SearchRole {
		whereSql += fmt.Sprintf(" or role_name like '%%%v%%' ", escapeSql(searchReq.FuzzyValue))
	}
	whereSql += ")"

	return whereSql, nil
}

//模糊搜索查询权限范围内的用户列表，包括用户名、角色名、层级名任意一个匹配（有权限过滤）
func ProcessFindLikeNameWithRight(feignKey *XFeignKey, searchReq *UserNameSearchDTO, includeAdmin bool) ([]SimpleUserDTO, code.Error) {
	if nil == feignKey || "" == feignKey.TenantId || nil == searchReq || "" == searchReq.FuzzyValue {
		return nil, errcode.CerrParamater
	}

	currentUser, errGetUser := GetUserDetailByIdOrPubid(feignKey, feignKey.UserId, 0)
	if nil != errGetUser {
		logrus.WithFields(logrus.Fields{"pubUserId": feignKey.UserId, "error": errGetUser}).Error("Get current user info failed")
		return nil, errGetUser
	} else if nil == currentUser {
		return nil, errcode.CerrAccountNotFound
	}
	//获取当前用户角色的right信息
	userRights, errGetRight := GetOneRoleRightList(feignKey, currentUser.RoleId)
	if nil != errGetRight {
		logrus.WithFields(logrus.Fields{"roleID": currentUser.RoleId, "error": errGetRight}).Error("Get current user right info failed")
		return nil, errGetRight
	}

	//获取所有的用户信息, map的key为parent_id
	mapUserInfo, errGetAllUser := getAllUserParentMapInfo(feignKey)
	if nil != errGetAllUser {
		return nil, errGetAllUser
	}
	//转换为Normalized权限信息
	currentRight, errNormalize := normalizeUserTreeRight(userRights, constant.ModuleUser)
	if nil != errNormalize {
		return nil, errNormalize
	}

	//用户权限可见范围
	permissionIds, errGetPermission := getUserIdPermissionScope(currentUser.Id, currentRight, mapUserInfo)
	if nil != errGetPermission {
		logrus.WithFields(logrus.Fields{"error": errGetPermission}).Error("getUserIdPermissionScope failed")
		return nil, errGetPermission
	}
	//logrus.WithFields(logrus.Fields{"userId": currentUser.Id, "permissionIds":permissionIds}).Info("user Permission")////////////
	whereSql, errGetSql := getWhereConditionFindLikeNameWithRight(feignKey, searchReq)
	if nil != errGetSql {
		logrus.WithFields(logrus.Fields{"error": errGetSql}).Error("getWhereConditionFindLikeNameWithRight failed")
		return nil, errGetSql
	}
	dbConn, errConn := GetDBConnByTenantID(feignKey.TenantId)
	if errConn != nil {
		return nil, errConn
	}

	var userRecords []UserDetailTab = make([]UserDetailTab, 0)
	dbtmp := dbConn.Where(whereSql).Find(&userRecords)
	if err := dbtmp.Error; nil != err {
		if err == gorm.ErrRecordNotFound {
			logrus.WithFields(logrus.Fields{"whereSql": whereSql}).Infof("Not found user records from DB")
			return nil, nil //未找到记录不返回错误
		}
		logrus.WithFields(logrus.Fields{"whereSql": whereSql, "error": err}).Infof("Get user records from DB failed")
		return nil, errcode.CerrExecuteSQL
	}

	var simpleList []SimpleUserDTO = make([]SimpleUserDTO, 0)
	for index, _ := range userRecords {
		if nil != permissionIds && false == Int64ListIsCotain(permissionIds, userRecords[index].Id) {
			continue //permissionIds为nil表示有所有用户权限
		}

		//去除ADMIN用户
		if !includeAdmin && userRecords[index].Id == 1 {
			continue
		}

		simpleUser, errConv := UserRecord2SimpleUser(&userRecords[index])
		if nil != errConv || nil == simpleUser {
			logrus.WithFields(logrus.Fields{"error": errConv}).Warn("UserRecord2SimpleUser abnormal")
			continue
		}
		simpleList = append(simpleList, *simpleUser)
	}

	return simpleList, nil
}

//模糊搜索查询权限范围内的角色、权限、用户（有权限过滤）
func ProcessFindRoleLevelUserLikeNameWithRight(feignKey *XFeignKey, searchReq *UserNameSearchDTO, includeAdmin bool) ([]*MsgReceiversDTO, code.Error) {
	if nil == feignKey || "" == feignKey.TenantId || nil == searchReq || "" == searchReq.FuzzyValue {
		return nil, errcode.CerrParamater
	}

	roles, roleErr := SearchRoleByFuzzyName(feignKey, searchReq.FuzzyValue)
	if roleErr != nil {
		return nil, roleErr
	}

	levels, levelErr := SearchLevelByFuzzyName(feignKey, searchReq.FuzzyValue)
	if levelErr != nil {
		return nil, levelErr
	}

	userList, userErr := ProcessFindLikeNameWithRight(feignKey, &UserNameSearchDTO{
		FuzzyValue:  searchReq.FuzzyValue,
		SearchUser:  true,
		SearchRole:  false,
		SearchLevel: false,
	}, includeAdmin)

	if userErr != nil {
		return nil, userErr
	}

	result := make([]*MsgReceiversDTO, 0)
	for index := range roles {
		result = append(result, &MsgReceiversDTO{
			Id:       strconv.FormatInt(roles[index].Id, 10),
			Name:     roles[index].Name,
			IdType:   constant.IdType_RoleId,
			EntityNo: roles[index].EntityNo,
		})
	}

	for index := range levels {
		result = append(result, &MsgReceiversDTO{
			Id:       strconv.FormatInt(levels[index].Id, 10),
			Name:     levels[index].Name,
			IdType:   constant.IdType_LevelId,
			EntityNo: levels[index].EntityNo,
		})
	}

	for index := range userList {
		result = append(result, &MsgReceiversDTO{
			Id:       strconv.FormatInt(userList[index].Id, 10),
			Name:     userList[index].Name,
			IdType:   constant.IdType_Id,
			EntityNo: userList[index].EntityNo,
		})
	}

	return result, nil
}

//返佣用户查询
func ProcessGetSimpleUserCommissionRight(feignKey *XFeignKey, searchReq *FuzzyConditionDTO) ([]IdNameDTO, code.Error) {
	if nil == feignKey || "" == feignKey.TenantId || nil == searchReq || "" == searchReq.FuzzyValue {
		return nil, errcode.CerrParamater
	}

	currentUser, errGetUser := GetUserDetailByIdOrPubid(feignKey, feignKey.UserId, 0)
	if nil != errGetUser {
		logrus.WithFields(logrus.Fields{"pubUserId": feignKey.UserId, "error": errGetUser}).Error("Get current user info failed")
		return nil, errGetUser
	} else if nil == currentUser {
		return nil, errcode.CerrAccountNotFound
	}
	//获取当前用户角色的right信息
	userRights, errGetRight := GetOneRoleRightList(feignKey, currentUser.RoleId)
	if nil != errGetRight {
		logrus.WithFields(logrus.Fields{"roleID": currentUser.RoleId, "error": errGetRight}).Error("Get current user right info failed")
		return nil, errGetRight
	}

	dbConn, errConn := GetDBConnByTenantID(feignKey.TenantId)
	if errConn != nil {
		return nil, errConn
	}
	var userRecords []UserDetailTab = make([]UserDetailTab, 0)
	var dbtmp *gorm.DB = nil
	searchReq.FuzzyValue = escapeSql(searchReq.FuzzyValue)
	if RightsHaveCotain(userRights, constant.AUTHORITY_COMMISSION_REPORT_SELECT_ALL) {
		dbtmp = dbConn.Where("name like ? and model_status != ?", "%"+searchReq.FuzzyValue+"%", constant.ModelStatusDelete).Find(&userRecords)
	} else if RightsHaveCotain(userRights, constant.AUTHORITY_COMMISSION_REPORT_SELECT_SUB) {
		//获取所有的用户信息, map的key为parent_id
		mapUserInfo, errGetAllUser := getAllUserParentMapInfo(feignKey)
		if nil != errGetAllUser {
			return nil, errGetAllUser
		}

		childUsers, errGetChild := getChildrenUserList([]int64{currentUser.Id}, constant.SearchUserScope_ALL, mapUserInfo, 1) //获取所有下级
		if nil != errGetChild {
			logrus.WithFields(logrus.Fields{"userId": currentUser.Id, "error": errGetChild}).Error("getChildrenUserList failed")
			return nil, errGetChild
		}
		childList := []int64{-11} //添加一个不可能存在的ID，避免为空
		for index, _ := range childUsers {
			childList = append(childList, childUsers[index].Id)
		}
		dbtmp = dbConn.Where("name like ? and id in (?) and model_status != ?", "%"+searchReq.FuzzyValue+"%", childList, constant.ModelStatusDelete).Find(&userRecords)
	}

	if err := dbtmp.Error; nil != err {
		if err == gorm.ErrRecordNotFound {
			logrus.WithFields(logrus.Fields{"FuzzyValue": searchReq.FuzzyValue}).Infof("Not found user records from DB")
			return nil, nil //未找到记录不返回错误
		}
		logrus.WithFields(logrus.Fields{"FuzzyValue": searchReq.FuzzyValue, "error": err}).Infof("Get user records from DB failed")
		return nil, errcode.CerrExecuteSQL
	}

	var idNameList []IdNameDTO = make([]IdNameDTO, 0)
	for index, _ := range userRecords {
		var node IdNameDTO
		node.Id = userRecords[index].Id
		node.Name = userRecords[index].Name
		node.EntityNo = userRecords[index].EntityNo
		idNameList = append(idNameList, node)
	}

	return idNameList, nil
}

//根据模块权限，按Name和EntityNumber模糊搜索用户信息
func ProcessGetSimpleUserByModuleRight(feignKey *XFeignKey, searchKey string, module string) ([]IdNameDTO, code.Error) {
	searchKey = escapeSql(searchKey)
	if nil == feignKey || "" == feignKey.TenantId || "" == searchKey || "" == module {
		return nil, errcode.CerrParamater
	}

	currentUser, errGetUser := GetUserRecordByIdOrPubid(feignKey, feignKey.UserId, 0)
	if nil != errGetUser {
		logrus.WithFields(logrus.Fields{"pubUserId": feignKey.UserId, "error": errGetUser}).Error("Get current user info failed")
		return nil, errGetUser
	} else if nil == currentUser {
		return nil, errcode.CerrAccountNotFound
	}
	//获取当前用户角色的right信息
	userRights, errGetRight := GetOneRoleRightList(feignKey, currentUser.RoleId)
	if nil != errGetRight {
		logrus.WithFields(logrus.Fields{"roleID": currentUser.RoleId, "error": errGetRight}).Error("Get current user right info failed")
		return nil, errGetRight
	}

	moduleRights, errNormalizeRight := normalizeUserTreeRight(userRights, module)
	if nil != errNormalizeRight {
		logrus.WithFields(logrus.Fields{"roleID": currentUser.RoleId, "module": module, "error": errNormalizeRight}).Error("Get current user normalized right info failed")
		return nil, errNormalizeRight
	}

	dbConn, errConn := GetDBConnByTenantID(feignKey.TenantId)
	if errConn != nil {
		return nil, errConn
	}

	var userRecords []UserDetailTab = make([]UserDetailTab, 0)
	var errOpedb error = nil
	if IntListIsCotain(moduleRights, constant.NORMALIZED_RIGHT_ALL) {
		errOpedb = dbConn.Where("(name like ? or entity_no like ?) and model_status != ?", "%"+searchKey+"%", "%"+searchKey+"%", constant.ModelStatusDelete).Find(&userRecords).Error
	} else {
		//获取所有的用户信息, map的key为parent_id
		mapUserInfo, errGetAllUser := getAllUserParentMapInfo(feignKey)
		if nil != errGetAllUser {
			return nil, errGetAllUser
		}

		//获取可见用户ID
		permissionIds, errGetPermission := getUserIdPermissionScope(currentUser.Id, moduleRights, mapUserInfo)
		if nil != errGetPermission {
			logrus.WithFields(logrus.Fields{"error": errGetPermission}).Error("getUserIdPermissionScope failed")
			return nil, errGetPermission
		}

		childList := []int64{-11} //添加一个不可能存在的ID，避免为空
		for index, _ := range permissionIds {
			childList = append(childList, permissionIds[index])
		}
		errOpedb = dbConn.Where("(name like ? or entity_no like ?) and id in (?) and model_status != ?", "%"+searchKey+"%", "%"+searchKey+"%", childList, constant.ModelStatusDelete).Find(&userRecords).Error
	}

	if nil != errOpedb {
		if errOpedb == gorm.ErrRecordNotFound {
			logrus.WithFields(logrus.Fields{"name": searchKey}).Infof("Not found user records from DB")
			return nil, nil //未找到记录不返回错误
		}
		logrus.WithFields(logrus.Fields{"name": searchKey, "error": errOpedb}).Infof("Get user records from DB failed")
		return nil, errcode.CerrExecuteSQL
	}

	var idNameList []IdNameDTO = make([]IdNameDTO, 0)
	for index, _ := range userRecords {
		var node IdNameDTO
		node.Id = userRecords[index].Id
		node.Name = userRecords[index].Name
		node.EntityNo = userRecords[index].EntityNo
		idNameList = append(idNameList, node)
	}

	return idNameList, nil
}

/*
//返佣账户查询,//实际上按name和EntitoNo搜索
func ProcessGetSimpleUserAccRight(feignKey *XFeignKey, name string) ([]IdNameDTO, code.Error) {
		return nil, errcode.CerrParamater
	}

	currentUser, errGetUser := GetUserRecordByIdOrPubid(feignKey, feignKey.UserId, 0)
	if nil != errGetUser {
		logrus.WithFields(logrus.Fields{"pubUserId": feignKey.UserId, "error": errGetUser}).Error("Get current user info failed")
		return nil, errGetUser
	} else if nil == currentUser {
		return nil, errcode.CerrAccountNotFound
	}
	//获取当前用户角色的right信息
	userRights, errGetRight := GetOneRoleRightList(feignKey, currentUser.RoleId)
	if nil != errGetRight {
		logrus.WithFields(logrus.Fields{"roleID": currentUser.RoleId, "error": errGetRight}).Error("Get current user right info failed")
		return nil, errGetRight
	}

	dbConn, errConn := GetDBConnByTenantID(feignKey.TenantId)
	if errConn != nil {
		return nil, errConn
	}
	var userRecords []UserDetailTab = make([]UserDetailTab, 0)
	var errOpedb error = nil
	if RightsHaveCotain(userRights, constant.AUTHORITY_ACCOUNT_REPORT_SELECT_ALL) {
		errOpedb = dbConn.Where("(name like ? or entity_no like ?) and model_status != ?", "%"+name+"%", "%"+name+"%", constant.ModelStatusDelete).Find(&userRecords).Error
	} else if RightsHaveCotain(userRights, constant.AUTHORITY_ACCOUNT_REPORT_SELECT_SUB) {
		//获取所有的用户信息, map的key为parent_id
		mapUserInfo, errGetAllUser := getAllUserParentMapInfo(feignKey)
		if nil != errGetAllUser {
			return nil, errGetAllUser
		}

		childUsers, errGetChild := getChildrenUserList([]int64{currentUser.Id}, constant.SearchUserScope_ALL, mapUserInfo, 1) //获取所有下级
		if nil != errGetChild {
			logrus.WithFields(logrus.Fields{"userId": currentUser.Id, "error": errGetChild}).Error("getChildrenUserList failed")
			return nil, errGetChild
		}
		childList := []int64{-11} //添加一个不可能存在的ID，避免为空
		for index, _ := range childUsers {
			childList = append(childList, childUsers[index].Id)
		}
		errOpedb = dbConn.Where("(name like ? or entity_no like ?) and id in (?) and model_status != ?", "%"+name+"%", "%"+name+"%", childList, constant.ModelStatusDelete).Find(&userRecords).Error
	} else if RightsHaveCotain(userRights, constant.AUTHORITY_ACCOUNT_REPORT_SELECT_MY) && strings.Contains(currentUser.Name, name) {
		userRecords = append(userRecords, *currentUser)
	}

	if nil != errOpedb {
		if errOpedb == gorm.ErrRecordNotFound {
			logrus.WithFields(logrus.Fields{"name": name}).Infof("Not found user records from DB")
			return nil, nil //未找到记录不返回错误
		}
		logrus.WithFields(logrus.Fields{"name": name, "error": errOpedb}).Infof("Get user records from DB failed")
		return nil, errcode.CerrExecuteSQL
	}

	var idNameList []IdNameDTO = make([]IdNameDTO, 0)
	for index, _ := range userRecords {
		var node IdNameDTO
		node.Id = userRecords[index].Id
		node.Name = userRecords[index].Name
		node.EntityNo = userRecords[index].EntityNo
		idNameList = append(idNameList, node)
	}

	return idNameList, nil
}*/

//为用户筛选生成where语句，返回where部分的sql
//若includeCurrentUser为false，则排除此用户ID
func getDetailWhereConditionForUserSearch(feignKey *XFeignKey, searchReq *UserSearchDTO, currentUser *UserDetail, includeCurrentUser bool) (string, code.Error) {
	simpleConditionWhere, errGetSimple := getSimpleWhereConditionForUserSearch(feignKey, searchReq)
	if nil != errGetSimple {
		logrus.WithFields(logrus.Fields{"error": errGetSimple}).Error("getSimpleWhereConditionForUserSearch failed")
		return "", errGetSimple
	}

	userWhere, errGetUse := getWhereUserConditionSearch(feignKey, searchReq, currentUser, includeCurrentUser)
	if nil != errGetUse {
		logrus.WithFields(logrus.Fields{"error": errGetUse}).Error("getWhereUserConditionSearch failed")
		return "", errGetSimple
	}

	return simpleConditionWhere + userWhere, nil
}

func getDetailWhereConditionForUserSearchV2(feignKey *XFeignKey, searchReq *UserSearchDTO, currentUser *UserDetail, includeCurrentUser bool) (string, code.Error) {
	simpleConditionWhere, errGetSimple := getSimpleWhereConditionForUserSearchV2(feignKey, searchReq)
	if nil != errGetSimple {
		logrus.WithFields(logrus.Fields{"error": errGetSimple}).Error("getSimpleWhereConditionForUserSearch failed")
		return "", errGetSimple
	}

	userWhere, errGetUse := getWhereUserConditionSearch(feignKey, searchReq, currentUser, includeCurrentUser)
	if nil != errGetUse {
		logrus.WithFields(logrus.Fields{"error": errGetUse}).Error("getWhereUserConditionSearch failed")
		return "", errGetUse
	}

	advanceSearchWhere, advanceSearchErr := getAdvanceSearchCondition(feignKey, searchReq.AdvanceConditions)
	if nil != advanceSearchErr {
		logrus.WithFields(logrus.Fields{"error": advanceSearchErr}).Error("getAdvanceSearchCondition failed")
		return "", advanceSearchErr
	}

	return simpleConditionWhere + userWhere + advanceSearchWhere, nil
}

func getWhereUserConditionSearch(feignKey *XFeignKey, searchReq *UserSearchDTO, currentUser *UserDetail, includeCurrentUser bool) (string, code.Error) {
	whereSql := ""
	if searchReq.StartDate > 0 {
		whereSql += fmt.Sprintf(" AND UNIX_TIMESTAMP(create_date) >= %v ", searchReq.StartDate/1000)
	}
	if searchReq.EndDate > 0 {
		whereSql += fmt.Sprintf(" AND UNIX_TIMESTAMP(create_date) <= %v ", searchReq.EndDate/1000)
	}
	if searchReq.LevelId > 0 {
		whereSql += fmt.Sprintf(" AND level_id = %v ", searchReq.LevelId)
	}
	if false == includeCurrentUser {
		whereSql += fmt.Sprintf(" AND id != %v ", currentUser.Id)
	}

	//获取当前用户角色的right信息
	userRights, errGetRight := GetOneRoleRightList(feignKey, currentUser.RoleId)
	if nil != errGetRight {
		logrus.WithFields(logrus.Fields{"roleID": currentUser.RoleId, "error": errGetRight}).Error("Get current user right info failed")
		return "", errGetRight
	}
	//获取所有的用户信息, map的key为parent_id
	mapUserInfo, errGetAllUser := getAllUserParentMapInfo(feignKey)
	if nil != errGetAllUser {
		return "", errGetAllUser
	}

	if "" == searchReq.UserId {
		if "all" == searchReq.UserSearchType {
			if false == RightsHaveCotain(userRights, constant.AUTHORITY_USER_SELECT_ALL) {
				logrus.WithFields(logrus.Fields{"UserSearchType": searchReq.UserSearchType}).Error("User not right to search")
				return "", errcode.CerrUserNotRight
			}
		} else if "sub" == searchReq.UserSearchType {
			if false == RightsHaveCotain(userRights, constant.AUTHORITY_USER_SELECT_DIRECTLY) && false == RightsHaveCotain(userRights, constant.AUTHORITY_USER_SELECT_ALL) {
				logrus.WithFields(logrus.Fields{"UserSearchType": searchReq.UserSearchType}).Error("User not right to search")
				return "", errcode.CerrUserNotRight
			}
			belongIds := "'-2838383'"                         //添加一个不可能存在的ID，避免传空集合
			belongUsers, exist := mapUserInfo[currentUser.Id] //当前用户的直接下属用户
			if exist {
				for index, _ := range belongUsers {
					belongIds += fmt.Sprintf(",%v", belongUsers[index].Id)
				}
			}
			whereSql += fmt.Sprintf(" AND id in (%v) ", belongIds)
		} else if "subBelong" == searchReq.UserSearchType {
			if false == RightsHaveCotain(userRights, constant.AUTHORITY_USER_SELECT_SUBORDINATE) && false == RightsHaveCotain(userRights, constant.AUTHORITY_USER_SELECT_ALL) {
				logrus.WithFields(logrus.Fields{"UserSearchType": searchReq.UserSearchType}).Error("User not right to search")
				return "", errcode.CerrUserNotRight
			}
			belongIds := "'-2838383'"                                                                                                     //添加一个不可能存在的ID，避免传空集合
			belongUsers, errGetchild := getChildrenUserList([]int64{currentUser.Id}, constant.SearchUserScope_NOT_BELONG, mapUserInfo, 1) //当前用户的非直接下属用户
			if nil == errGetchild {
				for index, _ := range belongUsers {
					belongIds += fmt.Sprintf(",%v", belongUsers[index].Id)
				}
			}
			whereSql += fmt.Sprintf(" AND id in (%v) ", belongIds)
		} else if "noParent" == searchReq.UserSearchType {
			if false == RightsHaveCotain(userRights, constant.AUTHORITY_USER_SELECT_WILD) && false == RightsHaveCotain(userRights, constant.AUTHORITY_USER_SELECT_ALL) {
				logrus.WithFields(logrus.Fields{"UserSearchType": searchReq.UserSearchType}).Error("User not right to search")
				return "", errcode.CerrUserNotRight
			}
			noParentIds := "'-2838383'"            //添加一个不可能存在的ID，避免传空集合
			noParentUsers, exist := mapUserInfo[0] //无parent的用户
			if exist {
				for index, _ := range noParentUsers {
					if 1 == noParentUsers[index].Id { //无上级用户不展示Admin用户
						continue
					}
					noParentIds += fmt.Sprintf(",%v", noParentUsers[index].Id)
				}
			}
			whereSql += fmt.Sprintf(" AND id in (%v) ", noParentIds)
		} else if "allSee" == searchReq.UserSearchType {
			if false == RightsHaveCotain(userRights, constant.AUTHORITY_USER_SELECT_WILD) && false == RightsHaveCotain(userRights, constant.AUTHORITY_USER_SELECT_ALL) &&
				false == RightsHaveCotain(userRights, constant.AUTHORITY_USER_SELECT_DIRECTLY) && false == RightsHaveCotain(userRights, constant.AUTHORITY_USER_SELECT_SUBORDINATE) {
				logrus.WithFields(logrus.Fields{"UserSearchType": searchReq.UserSearchType}).Error("User not right to search")
				return "", errcode.CerrUserNotRight
			}
			if RightsHaveCotain(userRights, constant.AUTHORITY_USER_SELECT_ALL) {
				//不添加过滤条件
			} else {
				var inUsers []UserDetailTab = make([]UserDetailTab, 0)
				if RightsHaveCotain(userRights, constant.AUTHORITY_USER_SELECT_DIRECTLY) {
					belongUsers, exist := mapUserInfo[currentUser.Id] //当前用户的直接下属用户
					if exist {
						inUsers = append(inUsers, belongUsers...)
					}
				}
				if RightsHaveCotain(userRights, constant.AUTHORITY_USER_SELECT_SUBORDINATE) {
					belongUsers, errGetchild := getChildrenUserList([]int64{currentUser.Id}, constant.SearchUserScope_NOT_BELONG, mapUserInfo, 1) //当前用户的非直接下属用户
					if nil == errGetchild {
						inUsers = append(inUsers, belongUsers...)
					}
				}
				if RightsHaveCotain(userRights, constant.AUTHORITY_USER_SELECT_WILD) {
					noParentUsers, exist := mapUserInfo[0] //无parent的用户
					if exist {
						inUsers = append(inUsers, noParentUsers...)
					}
				}
				inUserIds := strconv.FormatInt(currentUser.Id, 10) //添加用户自己
				for index, _ := range inUsers {
					inUserIds += fmt.Sprintf(",%v", inUsers[index].Id)
				}
				whereSql += fmt.Sprintf(" AND id in (%v) ", inUserIds)
			}
			//过滤默认生成的管理员用户
			whereSql += fmt.Sprintf(" AND id != 1 ")
		} else {
			logrus.WithFields(logrus.Fields{"UserSearchType": searchReq.UserSearchType}).Error("SearchType is not correct")
			return "", errcode.CerrParamater
		}
	} else { //"" != searchReq.UserId
		//转换为Normalized权限信息
		currentRight, errNormalize := normalizeUserTreeRight(userRights, constant.ModuleUser)
		if nil != errNormalize {
			return "", errNormalize
		}
		targetUserId, _ := strconv.ParseInt(searchReq.UserId, 10, 64)
		//用户权限可见范围
		permissionIds, errGetPermission := getUserIdPermissionScope(currentUser.Id, currentRight, mapUserInfo)
		if nil != errGetPermission {
			logrus.WithFields(logrus.Fields{"error": errGetPermission}).Error("getUserIdPermissionScope failed")
			return "", errGetPermission
		}
		if len(permissionIds) != 0 && false == Int64ListIsCotain(permissionIds, targetUserId) {
			return "", errcode.CerrUserNotRight
		}
		//若permissionIds为nil，则有所有权限
		inUsers := make([]UserDetailTab, 0) //存储targetUserId的child
		inUserIds := "-2838383"             //添加一个不可能存在的ID，避免传空集合
		if "sub" == searchReq.UserSearchType {
			belongUsers, exist := mapUserInfo[targetUserId] //targetUserId的直接下属用户
			if exist {
				inUsers = belongUsers
			}
		} else {
			belongUsers, errGetchild := getChildrenUserList([]int64{targetUserId}, constant.SearchUserScope_ALL, mapUserInfo, 1) //targetUserId的直接下属用户和非直接下级用户
			if nil == errGetchild {
				inUsers = belongUsers
			}
			inUserIds += fmt.Sprintf(",%v", targetUserId)
		}
		for index, _ := range inUsers {
			if nil == permissionIds || Int64ListIsCotain(permissionIds, inUsers[index].Id) {
				inUserIds += fmt.Sprintf(",%v", inUsers[index].Id)
			}
		}
		whereSql += fmt.Sprintf(" AND id in (%v) ", inUserIds)
	}

	logrus.WithFields(logrus.Fields{"QueryContent": searchReq.QueryContent, "QueryType": searchReq.QueryType, "whereSql": whereSql}).Debug("Generate where sql")
	return whereSql, nil
}

func getAdvanceSearchCondition(feignKey *XFeignKey, searchReq []*AdvanceCondition) (string, code.Error) {
	if len(searchReq) == 0 {
		return "", nil
	}

	var conditionList []string
	for index := range searchReq {
		oneCondition := searchReq[index]
		switch oneCondition.Field {
		case constant.SearchField_Login:
			conditionList = append(conditionList, fmt.Sprintf(" login like '%%%v%%' ", escapeSql(oneCondition.Value)))
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
		case constant.SearchField_LevelId:
			if oneCondition.Value != "" && oneCondition.Value != "all" {
				conditionList = append(conditionList, fmt.Sprintf(" level_id = %v ", escapeSql(oneCondition.Value)))
			}
			break
		case constant.SearchField_ExcludeIbRole:
			dbConn, errConn := GetDBConnByTenantID(feignKey.TenantId)
			if errConn != nil {
				return "", errConn
			}
			var roles = make([]RoleTab, 0)
			//根据like parent_name查询用户信息
			var dbtmp *gorm.DB
			if strings.Contains(strings.ToLower(oneCondition.Value), "1") {
				dbtmp = dbConn.Model(&RoleTab{}).Where(" role_type_id != 'ib'").Find(&roles)
			} else {
				continue
			}
			if err := dbtmp.Error; nil != err {
				if err == gorm.ErrRecordNotFound {
					return " 1 = -1 ", nil
				} else {
					logrus.WithFields(logrus.Fields{"name": oneCondition.Value, "error": err}).Error("query role ib name fail")
					return "", errcode.CerrExecuteSQL
				}
			}

			roleIdList := "'-2838383'" //添加一个不可能存在的ID，避免传空集合
			for index := range roles {
				roleIdList += fmt.Sprintf(",'%v'", roles[index].Id)
			}
			conditionList = append(conditionList, fmt.Sprintf(" role_id in (%v) ", roleIdList))
			break
		case constant.SearchField_RoleName:
			dbConn, errConn := GetDBConnByTenantID(feignKey.TenantId)
			if errConn != nil {
				return "", errConn
			}
			var roles = make([]RoleTab, 0)
			//根据like parent_name查询用户信息
			dbtmp := dbConn.Model(&RoleTab{}).Where(" name like ? ", "%"+oneCondition.Value+"%").Find(&roles)
			if err := dbtmp.Error; nil != err {
				if err == gorm.ErrRecordNotFound {
					return " 1 = -1 ", nil
				} else {
					logrus.WithFields(logrus.Fields{"name": oneCondition.Value, "error": err}).Error("query role like name fail")
					return "", errcode.CerrExecuteSQL
				}
			}

			roleIdList := "'-2838383'" //添加一个不可能存在的ID，避免传空集合
			for index := range roles {
				roleIdList += fmt.Sprintf(",'%v'", roles[index].Id)
			}
			conditionList = append(conditionList, fmt.Sprintf(" role_id in (%v) ", roleIdList))
			break
		case constant.SearchField_ParentName:
			dbConn, errConn := GetDBConnByTenantID(feignKey.TenantId)
			if errConn != nil {
				return "", errConn
			}
			var userRecords = make([]UserDetailTab, 0)
			//根据like parent_name查询用户信息
			dbtmp := dbConn.Where("name like ? and model_status != ?", "%"+oneCondition.Value+"%", constant.ModelStatusDelete).Find(&userRecords)
			if err := dbtmp.Error; nil != err {
				if err == gorm.ErrRecordNotFound {
					logrus.WithFields(logrus.Fields{"name": oneCondition.Value}).Infof("Not found user records by like name")
					//未找到记录不返回
				} else {
					logrus.WithFields(logrus.Fields{"name": oneCondition.Value, "error": err}).Error("Query user records by like name failed")
					return "", errcode.CerrExecuteSQL
				}
			}
			parentIdList := "'-2838383'" //添加一个不可能存在的ID，避免传空集合
			for index, _ := range userRecords {
				parentIdList += fmt.Sprintf(",'%v'", userRecords[index].Id)
			}
			conditionList = append(conditionList, fmt.Sprintf(" parent_id in (%v) ", parentIdList))
			break
		case constant.SearchField_Email:
			conditionList = append(conditionList, fmt.Sprintf(" email like '%%%v%%' ", oneCondition.Value))
			break
		case constant.SearchField_Phone:
			conditionList = append(conditionList, fmt.Sprintf(" phone like '%%%v%%' ", oneCondition.Value))
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
			conditionList = append(conditionList, fmt.Sprintf(" role_id in (%v) ", roleIdStr))
			break
		}
	}

	if len(conditionList) > 0 {
		return " and " + strings.Join(conditionList, " and "), nil
	}
	return "", nil
}

//根据条件查询用户的详情信息（有权限过滤）
func ProcessUserDetailByPage(feignKey *XFeignKey, searchReq *UserSearchDTO) (*UserDetailPageDTO, code.Error) {
	if nil == feignKey || "" == feignKey.TenantId || nil == searchReq || searchReq.PageNo <= 0 || searchReq.Size <= 0 {
		return nil, errcode.CerrParamater
	}

	//获取当前用户信息
	currentUser, errCurrent := GetUserDetailByIdOrPubid(feignKey, feignKey.UserId, 0)
	if nil != errCurrent {
		logrus.WithFields(logrus.Fields{"pubUserid": feignKey.UserId, "error": errCurrent}).Error("Get current user info failed")
		return nil, errCurrent
	} else if nil == currentUser {
		logrus.WithFields(logrus.Fields{"pubUserid": feignKey.UserId}).Error("Get current user info not exist")
		return nil, errcode.CerrAccountNotFound
	}

	whereSql, errGetSql := getDetailWhereConditionForUserSearch(feignKey, searchReq, currentUser, false)
	if nil != errGetSql {
		logrus.WithFields(logrus.Fields{"error": errGetSql}).Error("getDetailWhereConditionForUserSearch failed")
		return nil, errGetSql
	}
	dbConn, errConn := GetDBConnByTenantID(feignKey.TenantId)
	if errConn != nil {
		return nil, errConn
	}

	var total int64 = 0
	if errCount := dbConn.Model(&UserDetailTab{}).Where(whereSql).Count(&total).Error; nil != errCount {
		logrus.WithFields(logrus.Fields{"error": errCount, "whereSql": whereSql}).Error("get user count failed")
		return nil, errGetSql
	}
	logrus.WithFields(logrus.Fields{"whereSql": whereSql, "tenantId": feignKey.TenantId}).Debug("total user count")

	var response UserDetailPageDTO
	response.Total = total
	response.Size = searchReq.Size
	response.Pager = searchReq.PageNo
	response.Offset = (searchReq.PageNo - 1) * searchReq.Size
	if int(total)%searchReq.Size == 0 {
		response.Pages = int(total) / searchReq.Size
	} else {
		response.Pages = int(total)/searchReq.Size + 1
	}
	response.List = make([]UserDetail, 0)

	var userRecords []UserDetailTab = make([]UserDetailTab, 0)
	var dbtmp *gorm.DB = nil
	if "" != searchReq.Sortby { //排序处理
		var orderStr string
		switch searchReq.Sortby {
		case "createDate":
			orderStr = "create_date"
		case "roleName":
			orderStr = "role_name"
		case "levelName":
			orderStr = "level_name"
		case "modifyDate":
			orderStr = "modify_date"
		case "entityNo":
			orderStr = "entity_no"
		default:
			orderStr = "create_date"
		}
		if true == searchReq.OrderDesc {
			orderStr += " DESC"
		}
		dbtmp = dbConn.Where(whereSql).Order(orderStr).Offset(response.Offset).Limit(response.Size).Find(&userRecords)
	} else {
		dbtmp = dbConn.Where(whereSql).Offset(response.Offset).Limit(response.Size).Find(&userRecords)
	}
	if err := dbtmp.Error; nil != err {
		if err == gorm.ErrRecordNotFound {
			logrus.WithFields(logrus.Fields{"whereSql": whereSql, "offset": response.Offset, "limit": response.Size}).Debug("Not found user records from DB")
			return &response, nil //未找到记录不返回错误
		}
		logrus.WithFields(logrus.Fields{"whereSql": whereSql, "offset": response.Offset, "limit": response.Size, "error": err}).Error("Get user records from DB failed")
		return nil, errcode.CerrExecuteSQL
	}

	//if "level" == searchReq.Sortby {}		/////////对于sorttype为level要重新排序，排序规则相对复杂，可否产品优化？

	for index, _ := range userRecords {
		detail, errConv := UserRecord2UserDetail(&userRecords[index])
		if nil != errConv || nil == detail {
			logrus.WithFields(logrus.Fields{"error": errConv}).Warn("UserRecord2SimpleUser abnormal")
			continue
		}
		response.List = append(response.List, *detail)
	}

	ReviseRoleNameAndLevelName(response.List, feignKey)
	return &response, nil
}

func ReviseRoleNameAndLevelName(userList []UserDetail, feignKey *XFeignKey) code.Error {
	if len(userList) == 0 || nil == feignKey {
		return nil
	}
	//填充roleName，levelName
	roleList, errGetRole := GetAllRoleLis(feignKey.TenantId)
	if nil != errGetRole {
		logrus.WithFields(logrus.Fields{"error": errGetRole}).Error("GetAllRoleLis failed")
		return errGetRole
	}
	roleMap := RoleTabList2RoleTabMap(roleList)
	levelList, errGetLevel := ProcessGetLevelList(feignKey)
	if nil != errGetLevel {
		logrus.WithFields(logrus.Fields{"error": errGetLevel}).Error("ProcessGetLevelList failed")
		return errGetLevel
	}
	levelMap := LevelDetailList2LevelDetailMap(levelList)

	for index, _ := range userList {
		roleInfo, exist := roleMap[userList[index].RoleId]
		userList[index].RoleName = operator.TernaryOperatorString(exist, roleInfo.Name, "")

		levelInfo, exist := levelMap[userList[index].LevelId]
		userList[index].LevelName = operator.TernaryOperatorString(exist, levelInfo.Name, "")
	}
	return nil
}

//根据条件查询用户的详情信息（有权限过滤）
func ProcessUserDetailByPageV2(feignKey *XFeignKey, searchReq *UserSearchDTO) (*UserDetailPageDTO, code.Error) {
	if nil == feignKey || "" == feignKey.TenantId || nil == searchReq || searchReq.PageNo <= 0 || searchReq.Size <= 0 {
		return nil, errcode.CerrParamater
	}

	//获取当前用户信息
	currentUser, errCurrent := GetUserDetailByIdOrPubid(feignKey, feignKey.UserId, 0)
	if nil != errCurrent {
		logrus.WithFields(logrus.Fields{"pubUserid": feignKey.UserId, "error": errCurrent}).Error("Get current user info failed")
		return nil, errCurrent
	} else if nil == currentUser {
		logrus.WithFields(logrus.Fields{"pubUserid": feignKey.UserId}).Error("Get current user info not exist")
		return nil, errcode.CerrAccountNotFound
	}

	whereSql, errGetSql := getDetailWhereConditionForUserSearchV2(feignKey, searchReq, currentUser, false)
	if nil != errGetSql {
		logrus.WithFields(logrus.Fields{"error": errGetSql}).Error("getDetailWhereConditionForUserSearch failed")
		return nil, errGetSql
	}
	dbConn, errConn := GetDBConnByTenantID(feignKey.TenantId)
	if errConn != nil {
		return nil, errConn
	}

	var total int64 = 0
	if errCount := dbConn.Model(&UserDetailTab{}).Where(whereSql).Count(&total).Error; nil != errCount {
		logrus.WithFields(logrus.Fields{"error": errCount, "whereSql": whereSql}).Error("get user count failed")
		return nil, errGetSql
	}
	logrus.WithFields(logrus.Fields{"whereSql": whereSql, "tenantId": feignKey.TenantId}).Debug("total user count")

	var response UserDetailPageDTO
	response.Total = total
	response.Size = searchReq.Size
	response.Pager = searchReq.PageNo
	response.Offset = (searchReq.PageNo - 1) * searchReq.Size
	if int(total)%searchReq.Size == 0 {
		response.Pages = int(total) / searchReq.Size
	} else {
		response.Pages = int(total)/searchReq.Size + 1
	}
	response.List = make([]UserDetail, 0)

	var userRecords []UserDetailTab = make([]UserDetailTab, 0)
	var dbtmp *gorm.DB = nil
	if "" != searchReq.Sortby { //排序处理
		var orderStr string
		switch searchReq.Sortby {
		case "createDate":
			orderStr = "create_date"
		case "roleName":
			orderStr = "role_name"
		case "levelName":
			orderStr = "level_name"
		case "modifyDate":
			orderStr = "modify_date"
		case "entityNo":
			orderStr = "entity_no"
		default:
			orderStr = "create_date"
		}
		if true == searchReq.OrderDesc {
			orderStr += " DESC"
		}
		dbtmp = dbConn.Where(whereSql).Order(orderStr).Offset(response.Offset).Limit(response.Size).Find(&userRecords)
	} else {
		dbtmp = dbConn.Where(whereSql).Offset(response.Offset).Limit(response.Size).Find(&userRecords)
	}
	if err := dbtmp.Error; nil != err {
		if err == gorm.ErrRecordNotFound {
			logrus.WithFields(logrus.Fields{"whereSql": whereSql, "offset": response.Offset, "limit": response.Size}).Debug("Not found user records from DB")
			return &response, nil //未找到记录不返回错误
		}
		logrus.WithFields(logrus.Fields{"whereSql": whereSql, "offset": response.Offset, "limit": response.Size, "error": err}).Error("Get user records from DB failed")
		return nil, errcode.CerrExecuteSQL
	}

	//if "level" == searchReq.Sortby {}		/////////对于sorttype为level要重新排序，排序规则相对复杂，可否产品优化？

	for index, _ := range userRecords {
		detail, errConv := UserRecord2UserDetail(&userRecords[index])
		if nil != errConv || nil == detail {
			logrus.WithFields(logrus.Fields{"error": errConv}).Warn("UserRecord2SimpleUser abnormal")
			continue
		}
		response.List = append(response.List, *detail)
	}

	ReviseRoleNameAndLevelName(response.List, feignKey)
	return &response, nil
}

//根据条件查询用户
func SearchUserByCondition(feignKey *XFeignKey, searchReq *SearchDTO) ([]*UserDetailTab, code.Error) {
	if nil == feignKey || "" == feignKey.TenantId {
		return nil, errcode.CerrParamater
	}
	whereSql, advanceSearchErr := getAdvanceSearchCondition(feignKey, searchReq.Conditions)
	if nil != advanceSearchErr {
		logrus.WithFields(logrus.Fields{"error": advanceSearchErr}).Error("getAdvanceSearchCondition failed")
		return nil, advanceSearchErr
	}
	dbConn, errConn := GetDBConnByTenantID(feignKey.TenantId)
	if errConn != nil {
		return nil, errConn
	}
	var userRecords = make([]*UserDetailTab, 0)
	var dbtmp *gorm.DB = nil
	if len(searchReq.Fields) == 0 {
		dbtmp = dbConn.Where("1 = 1" + whereSql).Find(&userRecords)
	} else {
		dbtmp = dbConn.Select(strings.Join(searchReq.Fields, ",")).Where("1 = 1" + whereSql).Find(&userRecords)
	}
	if err := dbtmp.Error; nil != err {
		if err == gorm.ErrRecordNotFound {
			logrus.WithFields(logrus.Fields{"whereSql": whereSql}).Debug("Not found user records from DB")
			return userRecords, nil //未找到记录不返回错误
		}
		logrus.WithFields(logrus.Fields{"whereSql": whereSql, "error": err}).Error("Get user records from DB failed")
		return nil, errcode.CerrExecuteSQL
	}
	return userRecords, nil
}

//根据用户id和角色查询用户,两结果取并集
func ProcessUserInfoByIdsAndRoles(feignKey *XFeignKey, pubRoleReq *UserRoleDTO) ([]UserDetail, code.Error) {
	if nil == feignKey || "" == feignKey.TenantId || nil == pubRoleReq {
		return nil, errcode.CerrParamater
	}
	var userRecords []UserDetailTab = make([]UserDetailTab, 0)
	var userDetails []UserDetail = make([]UserDetail, 0)
	if len(pubRoleReq.PubUserIds) <= 0 && len(pubRoleReq.Roles) <= 0 { //请求列表为空的时候返回空，不要报错
		return userDetails, nil
	}

	var rolesInt []int64 = make([]int64, 0)
	if len(pubRoleReq.Roles) > 0 {
		for _, roleStr := range pubRoleReq.Roles {
			roleInt, errParse := strconv.ParseInt(roleStr, 10, 64)
			if nil != errParse {
				logrus.WithFields(logrus.Fields{"error": errParse, "roleStr": roleStr}).Error("Parse roles to int failed")
				return nil, errcode.CerrParamater
			}
			rolesInt = append(rolesInt, roleInt)
		}
	} else {
		rolesInt = append(rolesInt, -11) //添加一个不可能levelId, 避免为空
	}

	if len(pubRoleReq.PubUserIds) == 0 {
		pubRoleReq.PubUserIds = []string{"-11"} //添加一个不可能levelId, 避免为空
	}

	dbConn, errConn := GetDBConnByTenantID(feignKey.TenantId)
	if errConn != nil {
		return nil, errConn
	}
	dbtmp := dbConn.Where("(pub_user_id in (?) or role_id in (?)) and model_status != ?", pubRoleReq.PubUserIds, rolesInt, constant.ModelStatusDelete).Find(&userRecords)
	if err := dbtmp.Error; nil != err {
		if err == gorm.ErrRecordNotFound {
			logrus.WithFields(logrus.Fields{"pub_user_id": pubRoleReq.PubUserIds, "role_id": rolesInt}).Warn("Not found user records")
			return userDetails, nil //未找到记录不返回错误
		}
		logrus.WithFields(logrus.Fields{"error": err, "pub_user_id": pubRoleReq.PubUserIds, "role_id": rolesInt}).Error("Failed to get user records")
		return nil, errcode.CerrExecuteSQL
	}

	for index, _ := range userRecords {
		detail, errConv := UserRecord2UserDetail(&userRecords[index])
		if nil == errConv && nil != detail {
			userDetails = append(userDetails, *detail)
		}
	}

	ReviseRoleNameAndLevelName(userDetails, feignKey)
	return userDetails, nil
}

/*
//根据用户id和角色查询用户,两结果取并集
func ProcessUserInfoByIdsAndRoles(feignKey *XFeignKey, pubRoleReq *UserRoleDTO) ([]UserDetail, code.Error) {
	if nil == feignKey || "" == feignKey.TenantId || nil == pubRoleReq || (len(pubRoleReq.PubUserIds) <= 0 && len(pubRoleReq.Roles) <= 0) {
		return nil, errcode.CerrParamater
	}

	var userInfos []UserDetail = make([]UserDetail, 0)
	if len(pubRoleReq.PubUserIds) > 0 {
		userInfoByPub, errByPub := ProcessListUserInfosByKey(feignKey, pubRoleReq.PubUserIds, constant.QueryKeypubUserIds)
		if nil != errByPub {
			logrus.WithFields(logrus.Fields{"error": errByPub, "PubUserIds":pubRoleReq.PubUserIds}).Error("ProcessListUserInfosByKey failed")
			return nil, errByPub
		}
		userInfos = append(userInfos, userInfoByPub...)
	}

	if len(pubRoleReq.Roles) > 0 {
		var rolesInt []int64 = make([]int64, 0)
		for _, roleStr := range pubRoleReq.Roles {
			roleInt, errParse := strconv.ParseInt(roleStr, 10, 64)
			if nil != errParse {
				logrus.WithFields(logrus.Fields{"error": errParse, "roleStr":roleStr}).Error("Parse roles to int failed")
				return nil, errcode.CerrParamater
			}
			rolesInt = append(rolesInt, roleInt)
		}
		userInfoByRole, errByRole := ProcessListUserInfosByKey(feignKey, rolesInt, constant.QueryKeyRoleId)
		if nil != errByRole {
			logrus.WithFields(logrus.Fields{"error": errByRole, "Roles":pubRoleReq.Roles}).Error("ProcessListUserInfosByKey failed")
			return nil, errByRole
		}
		userInfos = append(userInfos, userInfoByRole...)
	}

	return userInfos, nil
}
*/

//返回child用户列表
func ProcessListChildUser(feignKey *XFeignKey, userId int64, execBelong bool) ([]UserDetailTab, code.Error) {
	if nil == feignKey || userId <= 0 || "" == feignKey.TenantId {
		return nil, errcode.CerrParamater
	}

	allUserMap, errGetAll := getAllUserParentMapInfo(feignKey)
	if nil != errGetAll {
		logrus.WithFields(logrus.Fields{"error": errGetAll}).Error("Get all user info failed")
		return nil, errGetAll
	}

	var childUsers []UserDetailTab
	var errGetChild code.Error
	if execBelong {
		childUsers, errGetChild = getChildrenUserList([]int64{userId}, constant.SearchUserScope_NOT_BELONG, allUserMap, 1) //获取非直属下级
	} else {
		childUsers, errGetChild = getChildrenUserList([]int64{userId}, constant.SearchUserScope_ALL, allUserMap, 1) //获取所有下级
	}
	if nil != errGetChild {
		logrus.WithFields(logrus.Fields{"error": errGetChild, "execBelong": execBelong}).Error("Get child user info failed")
		return nil, errGetChild
	}

	return childUsers, nil
}

//返回child用户的id列表
func ProcessListChildUserIds(feignKey *XFeignKey, userId int64, execBelong bool) ([]int64, code.Error) {
	childUsers, err := ProcessListChildUser(feignKey, userId, execBelong)
	if err != nil {
		return nil, err
	}

	var childIds []int64 = make([]int64, 0)
	for index, _ := range childUsers {
		childIds = append(childIds, childUsers[index].Id)
	}

	return childIds, nil
}

func ProcessFindUserByTypeFuzzy(feignKey *XFeignKey, id int64, typeValue int64, includeParent bool, fuzzyValue string) ([]UserDetailTab, code.Error) {
	if nil == feignKey || "" == feignKey.TenantId {
		return nil, errcode.CerrParamater
	}

	whereSql := "model_status != '" + constant.ModelStatusDelete + "' "
	if "" != fuzzyValue {
		whereSql += fmt.Sprintf(" and name like '%%%v%%' ", escapeSql(fuzzyValue))
	}
	switch typeValue {
	case constant.SEARCH_TYPE_ROLE:
		whereSql += fmt.Sprintf(" and role_id = %v ", id)
	case constant.SEARCH_TYPE_LEVEL:
		if id > 0 {
			if includeParent {
				mode := constant.CommissionMode_MULTI_AGENT
				tenantInfo, errGetTenant := ClientFindTenantById(feignKey, feignKey.TenantId)
				if nil != errGetTenant {
					logrus.WithFields(logrus.Fields{"errGetTenant": errGetTenant}).Warn("ClientFindTenantById failed")
				} else if nil != tenantInfo {
					mode = tenantInfo.Mode
				}

				if constant.CommissionMode_DISTRIBUTION == mode { //分销模式可查看所有层级用户
					whereSql += " and level_id > 0 "
				} else {
					//查找比当前level_id的sid小的所有level的id列表
					currentLevel, errCurentLevel := GetLevelInfoByID(feignKey.TenantId, id)
					if nil != errCurentLevel {
						logrus.WithFields(logrus.Fields{"errCurentLevel": errCurentLevel}).Error("Get current level info failed")
						return nil, errCurentLevel
					} else if nil == currentLevel {
						logrus.WithFields(logrus.Fields{"level_id": id}).Error("Get current level info not found")
						return nil, errcode.CerrLevelIDNotExist
					}
					allLevel, errGetAllLevel := ProcessGetLevelList(feignKey)
					if nil != errGetAllLevel || nil != errCurentLevel {
						logrus.WithFields(logrus.Fields{"errGetAllLevel": errGetAllLevel}).Error("Get all level info failed")
						return nil, errGetAllLevel
					}

					levelIds := "-11" //先添加一个不可能的ID，避免为空情况
					for index, _ := range allLevel {
						if allLevel[index].Sid < currentLevel.Sid {
							levelIds += fmt.Sprintf(",%v", allLevel[index].Id)
						}
					}
					whereSql += fmt.Sprintf(" and level_id in (%v) ", levelIds)
				}
			} else { //includeParent == false
				whereSql += fmt.Sprintf(" and level_id = %v ", id)
			}
		}
	default:
		logrus.WithFields(logrus.Fields{"typeValue": typeValue}).Error("ProcessFindUserByTypeFuzzy parameter typeValue abnormal")
		return nil, errcode.CerrParamater
	}
	logrus.WithFields(logrus.Fields{"whereSql": whereSql}).Debug("FindUserByTypeFuzzy sql")

	dbConn, errConn := GetDBConnByTenantID(feignKey.TenantId)
	if errConn != nil {
		return nil, errConn
	}
	var userRecords []UserDetailTab = make([]UserDetailTab, 0)
	dbtmp := dbConn.Where(whereSql).Find(&userRecords)

	if err := dbtmp.Error; nil != err {
		if err == gorm.ErrRecordNotFound {
			logrus.WithFields(logrus.Fields{"whereSql": whereSql}).Warn("Not found user records from DB by ids")
			return userRecords, nil //未找到记录不返回错误
		}
		logrus.WithFields(logrus.Fields{"whereSql": whereSql}).Error("Failed to get user records from DB by ids")
		return nil, errcode.CerrExecuteSQL
	}

	return userRecords, nil
}

//检查参数
func loginCheck(feignKey *XFeignKey, loginStr *string, vendorServerId *string) (bool, code.Error) {
	if nil == feignKey || nil == loginStr || "" == *loginStr {
		return true, nil
	}
	login, errParse := strconv.ParseInt(*loginStr, 10, 64)
	if nil != errParse {
		logrus.WithFields(logrus.Fields{"loginStr": loginStr}).Error("Failed to ParseInt for login")
		return false, errcode.CerrParamater
	}
	if (nil == vendorServerId || "" == *vendorServerId || false == strings.Contains(*vendorServerId, "_")) && login > 0 {
		return false, nil
	}
	var Ids []string
	if nil != vendorServerId {
		Ids = strings.Split(*vendorServerId, "_")
	}
	if len(Ids) >= 2 {
		return ClientCheckAccountExist(feignKey, login, Ids[1], Ids[0]) /////////////后面保留此行，删掉下面两行
		//ClientCheckAccountExist(login, Ids[1], Ids[0])								//////////////////后面删掉此行，保留上面的一行
		//return true, nil					////////////后面删掉此行，保留上面的一行
	} else {
		return false, nil
	}
}

//更新当前用户
func ProcessUpdateCurrentUser(feignKey *XFeignKey, bwUserReq *BWUserDTO) (*UserDetail, code.Error) {
	if nil == feignKey || "" == feignKey.TenantId || nil == bwUserReq {
		return nil, errcode.CerrParamater
	}

	currentUser, errGetCurrent := GetUserRecordByIdOrPubid(feignKey, feignKey.UserId, 0) //获取当前用户信息
	if nil != errGetCurrent {
		logrus.WithFields(logrus.Fields{"pubUserid": feignKey.UserId}).Error("Failed to get current user")
		return nil, errGetCurrent
	} else if nil == currentUser {
		logrus.WithFields(logrus.Fields{"pubUserid": feignKey.UserId}).Error("Not found current user")
		return nil, errcode.CerrAccountNotFound
	}

	if "" == bwUserReq.Name {
		logrus.WithFields(logrus.Fields{"pubUserid": feignKey.UserId}).Error("ProcessUpdateCurrentUser, Name is empty")
		return nil, errcode.CerrParamater
	}
	checkResult, errLoginCheck := loginCheck(feignKey, bwUserReq.Login, bwUserReq.VendorServerId)
	if nil != errLoginCheck || false == checkResult {
		logrus.WithFields(logrus.Fields{"Login": bwUserReq.Login, "VendorServerId": bwUserReq.VendorServerId}).Error("ProcessUpdateCurrentUser, loginCheck failed")
		return nil, errcode.CerrLoginCheckNotPass
	}

	if "" == currentUser.Name {
		currentUser.Name = bwUserReq.Name
	}
	currentUser.HeadImage = bwUserReq.HeadImage
	currentUser.Phone = bwUserReq.Phone
	currentUser.Sex = bwUserReq.Sex
	currentUser.Birthday = bwUserReq.Birthday
	currentUser.Country = bwUserReq.Country
	currentUser.Province = bwUserReq.Province
	currentUser.City = bwUserReq.City
	currentUser.Address = bwUserReq.Address
	currentUser.Comment = bwUserReq.Comment
	currentUser.NeedInitPass = []uint8{0}
	currentUser.Login = ""
	if nil != bwUserReq.Login {
		currentUser.Login = *bwUserReq.Login
	}
	currentUser.VendorServerId = ""
	if nil != bwUserReq.VendorServerId {
		currentUser.VendorServerId = *bwUserReq.VendorServerId
	}
	currentUser.Username = bwUserReq.UserName
	currentUser.Name = bwUserReq.Name
	if currentUser.CreateDate.Unix() < constant.MIN_TIMESTAMP || currentUser.CreateDate.Unix() > constant.MAX_TIMESTAMP {
		currentUser.CreateDate = time.Unix(constant.MIN_TIMESTAMP, 0)
	}
	currentUser.ModifyDate = time.Now()
	currentUser.ModifyUserId = feignKey.UserId

	//其它信息
	currentUser.IdType = bwUserReq.IdType
	currentUser.IdNum = bwUserReq.IdNum
	currentUser.IdUrl1 = bwUserReq.IdUrl1
	currentUser.IdUrl2 = bwUserReq.IdUrl2
	currentUser.BankAccount = bwUserReq.BankAccount
	currentUser.BankBranch = bwUserReq.BankBranch
	currentUser.AccountNo = bwUserReq.AccountNo
	currentUser.BankCardFile1 = bwUserReq.BankCardFile1
	currentUser.BankCardFile2 = bwUserReq.BankCardFile2
	currentUser.DoAgencyBusiness = bwUserReq.DoAgencyBusiness
	currentUser.InvestExperience = bwUserReq.InvestExperience
	currentUser.Agent = bwUserReq.Agent
	currentUser.Field01 = bwUserReq.Field01
	currentUser.Field02 = bwUserReq.Field02
	currentUser.Field03 = bwUserReq.Field03
	currentUser.Field04 = bwUserReq.Field04
	currentUser.Field05 = bwUserReq.Field05
	currentUser.Field06 = bwUserReq.Field06
	currentUser.Field07 = bwUserReq.Field07
	currentUser.Field08 = bwUserReq.Field08
	currentUser.Field09 = bwUserReq.Field09
	currentUser.Field10 = bwUserReq.Field10
	currentUser.Field11 = bwUserReq.Field11
	currentUser.Field12 = bwUserReq.Field12
	currentUser.Field13 = bwUserReq.Field13
	currentUser.Field14 = bwUserReq.Field14
	currentUser.Field15 = bwUserReq.Field15
	currentUser.Field16 = bwUserReq.Field16
	currentUser.Field17 = bwUserReq.Field17
	currentUser.Field18 = bwUserReq.Field18
	currentUser.Field19 = bwUserReq.Field19
	currentUser.Field20 = bwUserReq.Field20

	dbConn, errConn := GetDBConnByTenantID(feignKey.TenantId)
	if errConn != nil {
		return nil, errConn
	}
	dbtmp := dbConn.Save(currentUser)
	if err := dbtmp.Error; nil != err {
		logrus.WithFields(logrus.Fields{"error": err}).Error("Failed to save user info")
		return nil, errcode.CerrExecuteSQL
	}

	return UserRecord2UserDetail(currentUser)
}

//检查User表，若将childId的parent设置成parentId，是否会存在循环，若存在则返回true，否则返回false
//parentMapUserInfo，全量用户信息，key为parent_id, []UserDetailTab为相同parent的user信息
func checkUserParentChildCycle(parentMapUserInfo map[int64][]UserDetailTab, parentId int64, childId int64) (bool, code.Error) {
	if parentId < 0 || childId <= 0 {
		return false, errcode.CerrParamater
	}
	if childId == parentId {
		return true, nil //存在循环
	}

	idMapUser := ParentMapUser2IdMapUser(parentMapUserInfo) //转换为以用户id为key的map
	parentIdTmp := parentId
	depth := 0
	for parentIdTmp > 0 {
		parent, ok := idMapUser[parentIdTmp]
		if false == ok {
			break
		}

		var errParse error
		if "" == parent.ParentId {
			parentIdTmp = 0
		} else {
			parentIdTmp, errParse = strconv.ParseInt(parent.ParentId, 10, 64)
			if nil != errParse {
				logrus.WithFields(logrus.Fields{"error": errParse, "userId": parent.Id, "parentId": parent.ParentId}).Warn("Parse Parent failed")
				break
			}
		}

		if parentIdTmp == childId {
			return true, nil //存在循环
		}
		depth++
		if depth > 200 { //防止DB中的数据有异常导致程序死循环
			logrus.WithFields(logrus.Fields{"childId": childId, "parentId": parentId}).Warn("check user cycle loop 200 times")
			break
		}
	}
	return false, nil //不存在循环
}

//parentMapUserInfo，全量用户信息，key为parent_id, []UserDetailTab为相同parent的user信息
func parameterCheckForUpdateUser(feignKey *XFeignKey, bwUserReq *BWUserDTO, dbUser *UserDetailTab, parentMapUserInfo map[int64][]UserDetailTab) code.Error {
	roleIdPara, errParseRoleId := strconv.ParseInt(bwUserReq.RoleId, 10, 64)
	if nil != errParseRoleId || roleIdPara <= 0 || "" == bwUserReq.Name || "" == bwUserReq.Email || "" == bwUserReq.EntityNo {
		logrus.WithFields(logrus.Fields{"roleId": bwUserReq.RoleId}).Error("parameterCheckForUpdateUser roleId abnormal")
		return errcode.CerrParamater
	}
	if dbUser.PubUserId == feignKey.UserId {
		logrus.WithFields(logrus.Fields{"currentUserPubID": feignKey.UserId, "reqUserPubId": bwUserReq.PubUserId}).Error("only allow update other user")
		return errcode.CerrOperateForbidden
	}

	roleInfo, errGetRoleInfo := GetRoleInfoByID(feignKey.TenantId, roleIdPara)
	if nil != errGetRoleInfo {
		logrus.WithFields(logrus.Fields{"roleId": bwUserReq.RoleId, "error": errGetRoleInfo}).Error("GetRoleInfoByID failed")
		return errGetRoleInfo
	} else if nil == roleInfo {
		logrus.WithFields(logrus.Fields{"roleId": bwUserReq.RoleId}).Error("parameterCheckForUpdateUser roleId not exist")
		return errcode.CerrRoleIDNotExist
	}
	bwUserReq.RoleName = roleInfo.Name

	levelIdPara, errParseLevel := strconv.ParseInt(bwUserReq.LevelId, 10, 64)
	if nil == errParseLevel && levelIdPara > 0 {
		levelInfo, errGetLevelInfo := GetLevelInfoByID(feignKey.TenantId, levelIdPara)
		if nil != errGetLevelInfo {
			logrus.WithFields(logrus.Fields{"levelId": bwUserReq.LevelId, "error": errGetLevelInfo}).Error("GetLevelInfoByID failed")
			return errGetLevelInfo
		} else if nil == levelInfo {
			logrus.WithFields(logrus.Fields{"levelId": bwUserReq.LevelId}).Error("parameterCheckForUpdateUser levelId not exist")
			return errcode.CerrLevelIDNotExist
		}
		bwUserReq.LevelName = levelInfo.Name
	} else {
		bwUserReq.LevelId = "0"
		bwUserReq.LevelName = ""
	}

	if 1 == bwUserReq.Id && roleIdPara != dbUser.RoleId { //超级管理员不可以修改角色
		logrus.WithFields(logrus.Fields{"roleId": bwUserReq.RoleId, "dbRoleID": dbUser.RoleId}).Error("supper admin user not allow modify roleId")
		return errcode.CerrOperateForbidden
	}

	if bwUserReq.Email != dbUser.Email {
		exist, errCheckEmail := ExistInUserDetailTab(feignKey.TenantId, constant.QueryKeyEmail, bwUserReq.Email)
		if nil != errCheckEmail {
			logrus.WithFields(logrus.Fields{"email": bwUserReq.Email, "error": errCheckEmail}).Error("Check email exist failed")
			return errCheckEmail
		} else if true == exist {
			logrus.WithFields(logrus.Fields{"email": bwUserReq.Email}).Error("Check email have exist")
			return errcode.CerrEmailExist
		}
	}

	if bwUserReq.EntityNo != dbUser.EntityNo {
		exist, errCheckEntityNo := ExistInUserDetailTab(feignKey.TenantId, constant.QueryKeyEntityNo, bwUserReq.EntityNo)
		if nil != errCheckEntityNo {
			logrus.WithFields(logrus.Fields{"entityNo": bwUserReq.EntityNo, "error": errCheckEntityNo}).Error("Check entityNo exist failed")
			return errCheckEntityNo
		} else if true == exist {
			logrus.WithFields(logrus.Fields{"entityNo": bwUserReq.EntityNo}).Error("Check email have exist")
			return errcode.CerreEntityExist
		}
	}

	if nil != bwUserReq.Login && "" != *bwUserReq.Login && *bwUserReq.Login != dbUser.Login {
		exist, errCheckLogin := ExistInUserDetailTab(feignKey.TenantId, constant.QueryKeyLogin, bwUserReq.Login)
		if nil != errCheckLogin {
			logrus.WithFields(logrus.Fields{"login": bwUserReq.Login, "error": errCheckLogin}).Error("Check Login exist failed")
			return errCheckLogin
		} else if true == exist {
			logrus.WithFields(logrus.Fields{"email": bwUserReq.Email}).Error("Check login have exist")
			return errcode.CerreEntityExist
		}
	}
	//检查parent信息
	parentId, errParse := strconv.ParseInt(bwUserReq.Parent, 10, 64)
	if nil == errParse && parentId > 0 {
		existCycle, errCheck := checkUserParentChildCycle(parentMapUserInfo, parentId, bwUserReq.Id) //检查是否存在parent循环
		if nil == errCheck && true == existCycle {
			logrus.WithFields(logrus.Fields{"parentId": parentId, "userId": bwUserReq.Id}).Error("Check parentid have cycle")
			return errcode.CerrCheckCycleNotPass
		}
		//检查parent的level是否合适
		mode := constant.CommissionMode_MULTI_AGENT
		tenantInfo, errGetTenant := ClientFindTenantById(feignKey, feignKey.TenantId)
		if nil != errGetTenant {
			logrus.WithFields(logrus.Fields{"errGetTenant": errGetTenant}).Error("ClientFindTenantById failed")
			return errGetTenant
		} else if nil != tenantInfo {
			mode = tenantInfo.Mode
		}
		if constant.CommissionMode_DISTRIBUTION == mode {
			userRecords, err := ProcessFindUserByTypeFuzzy(feignKey, levelIdPara, constant.SEARCH_TYPE_LEVEL, true, "")
			if nil != err {
				logrus.WithFields(logrus.Fields{"error": err}).Error("ProcessFindUserByTypeFuzzy failed")
				return err
			}
			counter := 0
			for index, _ := range userRecords {
				if userRecords[index].Id == parentId {
					counter++
				}
			}
			if 0 == counter {
				logrus.WithFields(logrus.Fields{"parentId": parentId, "levelId": bwUserReq.LevelId, "userId": bwUserReq.Id}).Error("Update user, parent level not fit")
				return errcode.CerrParentLevelNotFit
			}
		}
	} else {
		bwUserReq.Parent = ""
	}

	checkResult, errLoginCheck := loginCheck(feignKey, bwUserReq.Login, bwUserReq.VendorServerId)
	if nil != errLoginCheck || false == checkResult {
		logrus.WithFields(logrus.Fields{"Login": bwUserReq.Login, "VendorServerId": bwUserReq.VendorServerId}).Error("ProcessUpdateCurrentUser, loginCheck failed")
		return errcode.CerrLoginCheckNotPass
	}
	return nil
}

func updatePass(feignKey *XFeignKey, pubUserId, email, phone, password string) code.Error {
	var userInfoDTO PubUserInfoDTO
	userInfoDTO.UserId = pubUserId
	userInfoDTO.Mail = email
	userInfoDTO.Phone = phone
	userInfoDTO.Password = password
	errSetPub := clientAdminSetUser(feignKey, &userInfoDTO)
	return errSetPub
}

//更新bwUserReq中的信息到dbUser，并保存到数据库
func updateDBUserInfoByReq(feignKey *XFeignKey, dbUser *UserDetailTab, bwUserReq *BWUserDTO) code.Error {
	dbUser.Birthday = bwUserReq.Birthday
	dbUser.City = bwUserReq.City
	dbUser.Comment = bwUserReq.Comment
	dbUser.Country = bwUserReq.Country
	dbUser.Address = bwUserReq.Address
	dbUser.Email = bwUserReq.Email
	dbUser.EntityNo = bwUserReq.EntityNo
	dbUser.HeadImage = bwUserReq.HeadImage
	dbUser.Nickname = bwUserReq.Nickname
	dbUser.ParentId = bwUserReq.Parent
	dbUser.Phone = bwUserReq.Phone
	dbUser.Postcode = bwUserReq.Postcode
	dbUser.Province = bwUserReq.Province
	dbUser.RoleId, _ = strconv.ParseInt(bwUserReq.RoleId, 10, 64)
	dbUser.RoleName = bwUserReq.RoleName
	dbUser.Username = bwUserReq.UserName
	dbUser.Name = bwUserReq.Name
	dbUser.Sex = bwUserReq.Sex
	dbUser.LevelId, _ = strconv.ParseInt(bwUserReq.LevelId, 10, 64)
	dbUser.LevelName = bwUserReq.LevelName
	dbUser.Login = ""
	if nil != bwUserReq.Login {
		dbUser.Login = *bwUserReq.Login
	}
	dbUser.VendorServerId = ""
	if nil != bwUserReq.VendorServerId {
		dbUser.VendorServerId = *bwUserReq.VendorServerId
	}
	dbUser.NeedInitPass = make([]uint8, 1)
	dbUser.NeedInitPass[0] = 0
	if true == bwUserReq.NeedInitPass {
		dbUser.NeedInitPass[0] = 1
	}
	if dbUser.CreateDate.Unix() < constant.MIN_TIMESTAMP || dbUser.CreateDate.Unix() > constant.MAX_TIMESTAMP {
		dbUser.CreateDate = time.Unix(constant.MIN_TIMESTAMP, 0)
	}
	dbUser.ModifyDate = time.Now()
	dbUser.ModifyUserId = feignKey.UserId

	dbConn, errConn := GetDBConnByTenantID(feignKey.TenantId)
	if errConn != nil {
		return errConn
	}
	if err := dbConn.Save(dbUser).Error; nil != err {
		logrus.WithFields(logrus.Fields{"userId": dbUser.Id, "error": err}).Error("Update user into to DB failed")
		return errcode.CerrExecuteSQL
	}
	return nil
}

//更新bwUserReq中的信息到dbUser，并保存到数据库
func updateDBUserInfoByReqV2(feignKey *XFeignKey, dbUser *UserDetailTab, bwUserReq *BWUserDTOIncrease) code.Error {
	dbUser.City = IfEmptyChoseOther(bwUserReq.City, dbUser.City)
	dbUser.Comment = IfEmptyChoseOther(bwUserReq.Comment, dbUser.Comment)
	dbUser.Country = IfEmptyChoseOther(bwUserReq.Country, dbUser.Country)
	dbUser.Address = IfEmptyChoseOther(bwUserReq.Address, dbUser.Address)
	dbUser.Email = IfEmptyChoseOther(bwUserReq.Email, dbUser.Email)
	dbUser.EntityNo = IfEmptyChoseOther(bwUserReq.EntityNo, dbUser.EntityNo)

	dbUser.ParentId = IfEmptyChoseOther(bwUserReq.Parent, dbUser.ParentId)
	dbUser.Phone = IfEmptyChoseOther(bwUserReq.Phone, dbUser.Phone)
	dbUser.Province = IfEmptyChoseOther(bwUserReq.Province, dbUser.Province)
	if nil != bwUserReq.RoleId && nil != bwUserReq.RoleName {
		dbUser.RoleId, _ = strconv.ParseInt(*bwUserReq.RoleId, 10, 64)
		dbUser.RoleName = operator.TernaryOperatorString(dbUser.RoleId != 0, *bwUserReq.RoleName, "")
	}
	dbUser.Username = IfEmptyChoseOther(bwUserReq.UserName, dbUser.Username)
	dbUser.Name = IfEmptyChoseOther(bwUserReq.Name, dbUser.Name)
	if nil != bwUserReq.LevelId && nil != bwUserReq.LevelName {
		dbUser.LevelId, _ = strconv.ParseInt(*bwUserReq.LevelId, 10, 64)
		dbUser.LevelName = operator.TernaryOperatorString(dbUser.LevelId != 0, *bwUserReq.LevelName, "")
	}
	if nil != bwUserReq.Login {
		dbUser.Login = *bwUserReq.Login
	}
	if nil != bwUserReq.VendorServerId {
		dbUser.VendorServerId = *bwUserReq.VendorServerId
	}
	if dbUser.CreateDate.Unix() < constant.MIN_TIMESTAMP || dbUser.CreateDate.Unix() > constant.MAX_TIMESTAMP {
		dbUser.CreateDate = time.Unix(constant.MIN_TIMESTAMP, 0)
	}
	dbUser.ModifyDate = time.Now()
	dbUser.ModifyUserId = feignKey.UserId

	//其它信息
	dbUser.IdType = IfEmptyChoseOther(bwUserReq.IdType, dbUser.IdType)
	dbUser.IdNum = IfEmptyChoseOther(bwUserReq.IdNum, dbUser.IdNum)
	dbUser.IdUrl1 = IfEmptyChoseOther(bwUserReq.IdUrl1, dbUser.IdUrl1)
	dbUser.IdUrl2 = IfEmptyChoseOther(bwUserReq.IdUrl2, dbUser.IdUrl2)
	dbUser.BankAccount = IfEmptyChoseOther(bwUserReq.BankAccount, dbUser.BankAccount)
	dbUser.BankBranch = IfEmptyChoseOther(bwUserReq.BankBranch, dbUser.BankBranch)
	dbUser.AccountNo = IfEmptyChoseOther(bwUserReq.AccountNo, dbUser.AccountNo)
	dbUser.BankCardFile1 = IfEmptyChoseOther(bwUserReq.BankCardFile1, dbUser.BankCardFile1)
	dbUser.BankCardFile2 = IfEmptyChoseOther(bwUserReq.BankCardFile2, dbUser.BankCardFile2)
	dbUser.DoAgencyBusiness = IfEmptyChoseOther(bwUserReq.DoAgencyBusiness, dbUser.DoAgencyBusiness)
	dbUser.InvestExperience = IfEmptyChoseOther(bwUserReq.InvestExperience, dbUser.InvestExperience)
	dbUser.Agent = bwUserReq.Agent
	dbUser.Field01 = IfEmptyChoseOther(bwUserReq.Field01, dbUser.Field01)
	dbUser.Field02 = IfEmptyChoseOther(bwUserReq.Field02, dbUser.Field02)
	dbUser.Field03 = IfEmptyChoseOther(bwUserReq.Field03, dbUser.Field03)
	dbUser.Field04 = IfEmptyChoseOther(bwUserReq.Field04, dbUser.Field04)
	dbUser.Field05 = IfEmptyChoseOther(bwUserReq.Field05, dbUser.Field05)
	dbUser.Field06 = IfEmptyChoseOther(bwUserReq.Field06, dbUser.Field06)
	dbUser.Field07 = IfEmptyChoseOther(bwUserReq.Field07, dbUser.Field07)
	dbUser.Field08 = IfEmptyChoseOther(bwUserReq.Field08, dbUser.Field08)
	dbUser.Field09 = IfEmptyChoseOther(bwUserReq.Field09, dbUser.Field09)
	dbUser.Field10 = IfEmptyChoseOther(bwUserReq.Field10, dbUser.Field10)
	dbUser.Field11 = IfEmptyChoseOther(bwUserReq.Field11, dbUser.Field11)
	dbUser.Field12 = IfEmptyChoseOther(bwUserReq.Field12, dbUser.Field12)
	dbUser.Field13 = IfEmptyChoseOther(bwUserReq.Field13, dbUser.Field13)
	dbUser.Field14 = IfEmptyChoseOther(bwUserReq.Field14, dbUser.Field14)
	dbUser.Field15 = IfEmptyChoseOther(bwUserReq.Field15, dbUser.Field15)
	dbUser.Field16 = IfEmptyChoseOther(bwUserReq.Field16, dbUser.Field16)
	dbUser.Field17 = IfEmptyChoseOther(bwUserReq.Field17, dbUser.Field17)
	dbUser.Field18 = IfEmptyChoseOther(bwUserReq.Field18, dbUser.Field18)
	dbUser.Field19 = IfEmptyChoseOther(bwUserReq.Field19, dbUser.Field19)
	dbUser.Field20 = IfEmptyChoseOther(bwUserReq.Field20, dbUser.Field20)

	if bwUserReq.Points1 != nil && "" != *bwUserReq.Points1 {
		if _, errStr := strconv.ParseFloat(*bwUserReq.Points1, 64); errStr == nil {
			dbUser.Points1 = *bwUserReq.Points1
		}
	}
	if bwUserReq.Points2 != nil && "" != *bwUserReq.Points2 {
		if _, errStr := strconv.ParseFloat(*bwUserReq.Points2, 64); errStr == nil {
			dbUser.Points2 = *bwUserReq.Points2
		}
	}
	if bwUserReq.Points3 != nil && *bwUserReq.Points3 != "" {
		if _, errStr := strconv.ParseFloat(*bwUserReq.Points3, 64); errStr == nil {
			dbUser.Points3 = *bwUserReq.Points3
		}
	}
	if bwUserReq.Points4 != nil && *bwUserReq.Points4 != "" {
		if _, errStr := strconv.ParseFloat(*bwUserReq.Points4, 64); errStr == nil {
			dbUser.Points4 = *bwUserReq.Points4
		}
	}
	if bwUserReq.Points5 != nil && *bwUserReq.Points5 != "" {
		if _, errStr := strconv.ParseFloat(*bwUserReq.Points5, 64); errStr == nil {
			dbUser.Points5 = *bwUserReq.Points5
		}
	}
	if bwUserReq.Points6 != nil && *bwUserReq.Points6 != "" {
		if _, errStr := strconv.ParseFloat(*bwUserReq.Points6, 64); errStr == nil {
			dbUser.Points6 = *bwUserReq.Points6
		}
	}
	if bwUserReq.Points7 != nil && *bwUserReq.Points7 != "" {
		if _, errStr := strconv.ParseFloat(*bwUserReq.Points7, 64); errStr == nil {
			dbUser.Points7 = *bwUserReq.Points7
		}
	}

	dbConn, errConn := GetDBConnByTenantID(feignKey.TenantId)
	if errConn != nil {
		return errConn
	}
	if err := dbConn.Save(dbUser).Error; nil != err {
		logrus.WithFields(logrus.Fields{"userId": dbUser.Id, "error": err}).Error("Update user into to DB failed")
		return errcode.CerrExecuteSQL
	}
	return nil
}

func breakUserAccountRelation(feignKey *XFeignKey, vendor string, serverId string, accountId string) code.Error {
	dbConn, errConn := GetDBConnByTenantID(feignKey.TenantId)
	if errConn != nil {
		return errConn
	}

	if err := dbConn.Model(&UserDetailTab{}).Where("vendor_server_id = (?) and login = (?)", vendor+"_"+serverId, accountId).Update("login", nil).Error; nil != err {
		logrus.WithFields(logrus.Fields{"serverId": serverId, "vendor": vendor, "accountId": accountId, "error": err}).Error("breakUserAccountRelation failed")
		return errcode.CerrExecuteSQL
	}

	return nil
}

// 判断当前用户是否能查看用户
// userId 查看的用户ID    return true:具有权限,false:不具有权限
//调用的ProcessUserDetailByPage逻辑很重，是否可以优化？///////////////////////
func currentUserCanSeeOneUser(feignKey *XFeignKey, userId int64) (bool, code.Error) {
	var userSerach UserSearchDTO
	userSerach.QueryType = constant.QueryType_ID
	userSerach.UserSearchType = "allSee"
	userSerach.QueryContent = strconv.FormatInt(userId, 10)
	userSerach.PageNo = 1
	userSerach.Size = 10
	userPageInfo, err := ProcessUserDetailByPage(feignKey, &userSerach) //若当前用户对userID有权限， 则结果集不为空
	if nil != err {
		logrus.WithFields(logrus.Fields{"error": err}).Error("ProcessUserDetailByPage failed")
		return false, err
	}
	if nil != userPageInfo && len(userPageInfo.List) > 0 {
		return true, nil //有权限查看
	}
	return false, nil //无权限查看
}

//全量更新当前用户信息
func ProcessUpdateUserV1(feignKey *XFeignKey, bwUserReq *BWUserDTO) (*UserDetail, code.Error) {
	if nil == feignKey || "" == feignKey.TenantId || nil == bwUserReq || bwUserReq.Id <= 0 {
		return nil, errcode.CerrParamater
	}

	bwUserReq.Email = strings.TrimSpace(bwUserReq.Email)
	//邮箱检验
	isEmail, _ := IsEmail(bwUserReq.Email)

	if isEmail == false {
		return nil, code.NewMcode(
			fmt.Sprintf("EMAIL_FORMAT_ERROR"),
			"EMAIL_FORMAT_ERROR",
		)
	}

	parentMapUserInfo, errGetAll := getAllUserParentMapInfo(feignKey)
	if nil != errGetAll {
		logrus.WithFields(logrus.Fields{"pubUserid": feignKey.UserId}).Error("Failed to get all user info")
		return nil, errGetAll
	}

	dbUser, errGetUser := GetUserRecordByIdOrPubid(feignKey, "", bwUserReq.Id)
	if nil != errGetUser {
		logrus.WithFields(logrus.Fields{"userId": bwUserReq.Id}).Error("get user info failed")
		return nil, errGetUser
	} else if nil == dbUser {
		logrus.WithFields(logrus.Fields{"userId": bwUserReq.Id}).Error("User info for update not exist")
		return nil, errcode.CerrAccountNotFound
	}

	errCheckPara := parameterCheckForUpdateUser(feignKey, bwUserReq, dbUser, parentMapUserInfo)
	if nil != errCheckPara {
		logrus.WithFields(logrus.Fields{"update userId": bwUserReq.Id, "error": errCheckPara}).Error("parameterCheckForUpdateUser failed")
		return nil, errCheckPara
	}

	canSee, errCheck := currentUserCanSeeOneUser(feignKey, bwUserReq.Id) //检查当前用户是否有权限更新本次操作
	if nil != errCheck {
		logrus.WithFields(logrus.Fields{"update userId": bwUserReq.Id, "error": errCheck}).Error("currentUserCanSeeOneUser failed")
		return nil, errCheck
	} else if false == canSee {
		logrus.WithFields(logrus.Fields{"update userId": bwUserReq.Id, "currentUserPubId": feignKey.UserId}).Error("Current user not permission to update this user")
		return nil, errcode.CerrOperateForbidden
	}

	if errUpdatePass := updatePass(feignKey, dbUser.PubUserId, bwUserReq.Email, bwUserReq.Phones.Phone, bwUserReq.Password); nil != errUpdatePass {
		logrus.WithFields(logrus.Fields{"update userId": bwUserReq.Id, "error": errUpdatePass}).Error("updatePass failed")
		return nil, errUpdatePass
	}

	//更新用户基本信息到DB
	errUpdate := updateDBUserInfoByReq(feignKey, dbUser, bwUserReq)
	if nil != errUpdate {
		logrus.WithFields(logrus.Fields{"userId": bwUserReq.Id, "error": errUpdate}).Error("updateDBUserInfoByReq failed")
		return nil, errUpdate
	}
	logrus.WithFields(logrus.Fields{"user_id": bwUserReq.Id, "operate userPubId": feignKey.UserId}).Info("Updatee user info success")

	//保存黑白名单到DB（暂未实现）	/////////////

	//请求 用户佣金规则详细-添加修改
	if nil == bwUserReq.Commission {
		bwUserReq.Commission = new(UserCommissionUpdateRequestDTO)
	}
	bwUserReq.Commission.UserId = dbUser.Id /////需确认ID是否已是插入记录到DB后的值
	bwUserReq.Commission.LevelId = dbUser.LevelId
	bwUserReq.Commission.ParentId, _ = strconv.ParseInt(dbUser.ParentId, 10, 64)
	_, errCommission := ClientAddOrUpdateCommission(feignKey, bwUserReq.Commission)
	if nil != errCommission {
		logrus.WithFields(logrus.Fields{"error": errCommission}).Warn("AddOrUpdateCommission failed")
		//return errCommission					//添加Commission失败，不返回错误
	}

	//计算更新相关的用户数（比如下级用户数，层级用户数等）
	go UpdageSubCounter(feignKey)
	//UserMqSender.sendUserMessage(authInfo.tenantId, userDto.getId(), BwEventType.UPDATE);///////////////

	return UserRecord2UserDetail(dbUser)
}

func isLoginBindByUser(tenant_id string, vendorServer string, login string) (bool, code.Error) {
	dbConn, errConn := GetDBConnByTenantID(tenant_id)
	if errConn != nil {
		return false, errConn
	}

	var userRecord UserDetailTab
	var dbtmp *gorm.DB
	dbtmp = dbConn.Where("vendor_server_id = ? and login = ? and model_status != ?", vendorServer, login, constant.ModelStatusDelete).First(&userRecord)

	if err := dbtmp.Error; nil != err {
		if err == gorm.ErrRecordNotFound {
			logrus.WithFields(logrus.Fields{"vendorServer": vendorServer, "login": login}).Debug("Not found user record from DB")
			return false, nil //无记录不返回错误
		}
		logrus.WithFields(logrus.Fields{"error": err}).Error("Failed to get user record from DB by ids")
		return false, errcode.CerrExecuteSQL
	}

	logrus.WithFields(logrus.Fields{"vendorServer": vendorServer, "login": login}).Debug("Found records from DB")
	return true, nil
}

//parentMapUserInfo，全量用户信息，key为parent_id, []UserDetailTab为相同parent的user信息
func parameterCheckForUpdateUserV2(feignKey *XFeignKey, bwUserReq *BWUserDTOIncrease, dbUser *UserDetailTab, parentMapUserInfo map[int64][]UserDetailTab) code.Error {
	if !isEmptyOrNum(bwUserReq.RoleId) || !isEmptyOrNum(bwUserReq.LevelId) || !isEmptyOrNum(bwUserReq.Parent) {
		logrus.WithFields(logrus.Fields{}).Error("RoleId/LevelId/Parent abnormal")
		return errcode.CerrParamater
	}
	if nil != bwUserReq.Email && "" == *bwUserReq.Email {
		logrus.WithFields(logrus.Fields{}).Error("Email abnormal")
		return errcode.CerrParamater
	}
	if nil != bwUserReq.EntityNo && "" == *bwUserReq.EntityNo {
		logrus.WithFields(logrus.Fields{}).Error("entityNo abnormal")
		return errcode.CerrParamater
	}

	if nil != bwUserReq.RoleId {
		if "" != *bwUserReq.RoleId {
			roleIdPara, errParseRole := strconv.ParseInt(*bwUserReq.RoleId, 10, 64)
			if nil == errParseRole && roleIdPara > 0 {
				roleInfo, errGet := GetRoleInfoByID(feignKey.TenantId, roleIdPara)
				if nil != errGet {
					logrus.WithFields(logrus.Fields{"error": errGet, "roleId": *bwUserReq.RoleId}).Error("GetRoleInfoByID failed")
					return errGet
				} else if nil == roleInfo {
					logrus.WithFields(logrus.Fields{"tenant_id": feignKey.TenantId, "error": errGet, "roleId": *bwUserReq.RoleId}).Error("role not exist failed")
					return errcode.CerrRoleIDNotExist
				}
				roleName := roleInfo.Name
				bwUserReq.RoleName = &roleName
			}
		} else { // "" == *bwUserReq.RoleId
			roleName := ""
			bwUserReq.RoleName = &roleName
		}

	}

	levelUpdate := dbUser.LevelId
	if nil != bwUserReq.LevelId {
		if "" != *bwUserReq.LevelId {
			levelIdPara, errParseLevel := strconv.ParseInt(*bwUserReq.LevelId, 10, 64)
			if nil == errParseLevel && levelIdPara > 0 {
				levelInfo, errGetLevelInfo := GetLevelInfoByID(feignKey.TenantId, levelIdPara)
				if nil != errGetLevelInfo {
					logrus.WithFields(logrus.Fields{"levelId": *bwUserReq.LevelId, "error": errGetLevelInfo}).Error("GetLevelInfoByID failed")
					return errGetLevelInfo
				} else if nil == levelInfo {
					logrus.WithFields(logrus.Fields{"levelId": *bwUserReq.LevelId}).Error("parameterCheckForUpdateUser levelId not exist")
					return errcode.CerrLevelIDNotExist
				}
				levelUpdate = levelIdPara
				levelName := levelInfo.Name
				bwUserReq.LevelName = &levelName
			}
		} else { //"" == *bwUserReq.LevelId
			levelUpdate = 0
			levelName := ""
			bwUserReq.LevelName = &levelName
		}
	}

	if nil != bwUserReq.Email && *bwUserReq.Email != dbUser.Email { //邮箱不能重复
		exist, errCheckEmail := ExistInUserDetailTab(feignKey.TenantId, constant.QueryKeyEmail, *bwUserReq.Email)
		if nil != errCheckEmail {
			logrus.WithFields(logrus.Fields{"email": *bwUserReq.Email, "error": errCheckEmail}).Error("Check email exist failed")
			return errCheckEmail
		} else if true == exist {
			logrus.WithFields(logrus.Fields{"email": *bwUserReq.Email}).Error("Check email have exist")
			return errcode.CerrEmailExist
		}
	}

	if nil != bwUserReq.EntityNo && *bwUserReq.EntityNo != dbUser.EntityNo { //编码不能重复
		exist, errCheckEntityNo := ExistInUserDetailTab(feignKey.TenantId, constant.QueryKeyEntityNo, *bwUserReq.EntityNo)
		if nil != errCheckEntityNo {
			logrus.WithFields(logrus.Fields{"entityNo": *bwUserReq.EntityNo, "error": errCheckEntityNo}).Error("Check entityNo exist failed")
			return errCheckEntityNo
		} else if true == exist {
			logrus.WithFields(logrus.Fields{"entityNo": *bwUserReq.EntityNo}).Error("Check email have exist")
			return errcode.CerreEntityExist
		}
	}

	if 1 == bwUserReq.Id && nil != bwUserReq.RoleId { //超级管理员不可以修改角色
		roleIdPara, _ := strconv.ParseInt(*bwUserReq.RoleId, 10, 64)
		if roleIdPara != dbUser.RoleId {
			logrus.WithFields(logrus.Fields{"roleId": *bwUserReq.RoleId, "dbRoleID": dbUser.RoleId}).Error("supper admin user not allow modify roleId")
			return errcode.CerrOperateForbidden
		}
	}

	//检查parent信息
	if nil != bwUserReq.Parent {
		parentId, errParse := strconv.ParseInt(*bwUserReq.Parent, 10, 64)
		if nil == errParse && parentId > 0 {
			existCycle, errCheck := checkUserParentChildCycle(parentMapUserInfo, parentId, bwUserReq.Id) //检查是否存在parent循环
			if nil == errCheck && true == existCycle {
				logrus.WithFields(logrus.Fields{"parentId": parentId, "userId": bwUserReq.Id}).Error("Check parentid have cycle")
				return errcode.CerrCheckCycleNotPass
			}

			if levelUpdate == 0 {
				logrus.WithFields(logrus.Fields{"parentId": parentId, "levelId": levelUpdate, "userId": bwUserReq.Id}).Error("Update user, parent level not fit")
				return errcode.CerrParentLevelNotFit
			}

			//检查parent的level是否合适
			userRecords, err := ProcessFindUserByTypeFuzzy(feignKey, levelUpdate, constant.SEARCH_TYPE_LEVEL, true, "")
			if nil != err {
				logrus.WithFields(logrus.Fields{"error": err}).Error("ProcessFindUserByTypeFuzzy failed")
				return err
			}
			counter := 0
			for index, _ := range userRecords {
				if userRecords[index].Id == parentId {
					counter++
				}
			}
			if 0 == counter {
				logrus.WithFields(logrus.Fields{"parentId": parentId, "levelId": bwUserReq.LevelId, "userId": bwUserReq.Id}).Error("Update user, parent level not fit")
				return errcode.CerrParentLevelNotFit
			}
		}
	}

	//校验server和login合法性
	finalVendorServer := dbUser.VendorServerId
	if nil != bwUserReq.VendorServerId {
		finalVendorServer = *bwUserReq.VendorServerId
	}
	finalLogin := dbUser.Login
	if nil != bwUserReq.Login {
		finalLogin = *bwUserReq.Login
	}
	if finalLogin != "" && finalVendorServer == "" {
		logrus.WithFields(logrus.Fields{"finalLogin": finalLogin, "finalVendorServer": finalVendorServer}).Error("Update user, login serverId not fit")
		return errcode.CerrLoginCheckNotPass
	}

	if (finalVendorServer != dbUser.VendorServerId || finalLogin != dbUser.Login) && finalLogin != "" {
		checkResult, errLoginCheck := loginCheck(feignKey, &finalLogin, &finalVendorServer)
		if nil != errLoginCheck || false == checkResult {
			logrus.WithFields(logrus.Fields{"Login": bwUserReq.Login, "VendorServerId": *bwUserReq.VendorServerId}).Error("ProcessUpdateCurrentUser, loginCheck failed")
			return errcode.CerrLoginCheckNotPass
		}
		// 2018-3-26 支持返佣账号绑定一个
		//isBind, errCheck := isLoginBindByUser(feignKey.TenantId, finalVendorServer, finalLogin)
		//if nil != errCheck || true == isBind {
		//	logrus.WithFields(logrus.Fields{"Login": bwUserReq.Login, "VendorServerId": *bwUserReq.VendorServerId, "isBind": isBind}).Error("isLoginBindByOtherUser error")
		//	return errcode.CerrLoginCheckNotPass
		//}
	}

	return nil
}

//更新邮箱
func ProcessUpdateEmail(feignKey *XFeignKey, bwUserReq *BWUserDTO) (string, code.Error) {
	if nil == feignKey || "" == feignKey.TenantId || nil == bwUserReq || bwUserReq.Id <= 0 || "" == bwUserReq.Email {
		return "", errcode.CerrParamater
	}

	bwUserReq.Email = strings.TrimSpace(bwUserReq.Email)
	//邮箱检验
	isEmail, _ := IsEmail(bwUserReq.Email)

	if isEmail == false {
		return "", code.NewMcode(
			fmt.Sprintf("EMAIL_FORMAT_ERROR"),
			"EMAIL_FORMAT_ERROR",
		)
	}

	userRecord, errGetUser := GetUserRecordByIdOrPubid(feignKey, "", bwUserReq.Id)
	if nil != errGetUser {
		logrus.WithFields(logrus.Fields{"errGetUser": errGetUser}).Error("GetUserRecordByIdOrPubid failed")
		return "", errGetUser
	} else if nil == userRecord {
		return "", errcode.CerrAccountNotFound
	}

	dbConn, errConn := GetDBConnByTenantID(feignKey.TenantId)
	if errConn != nil {
		return "", errConn
	}

	errUpdate := dbConn.Model(userRecord).Update("email", bwUserReq.Email).Error
	if nil != errUpdate {
		logrus.WithFields(logrus.Fields{"errGetUser": errUpdate, "tenant_id": feignKey.TenantId, "userId": bwUserReq.Id}).Error("update email failed")
		return "", errGetUser
	}
	return bwUserReq.Email, nil
}

//增量更新用户信息
func ProcessUpdateUserV2(feignKey *XFeignKey, bwUserReq *BWUserDTOIncrease) (*UserDetail, code.Error) {
	if nil == feignKey || "" == feignKey.TenantId || nil == bwUserReq || bwUserReq.Id <= 0 {
		return nil, errcode.CerrParamater
	}

	if bwUserReq.Email != nil {
		*bwUserReq.Email = strings.TrimSpace(*bwUserReq.Email)

		isEmail, _ := IsEmail(*bwUserReq.Email)
		//邮箱检验
		if isEmail == false {
			return nil, code.NewMcode(
				fmt.Sprintf("EMAIL_FORMAT_ERROR"),
				"EMAIL_FORMAT_ERROR",
			)
		}
	}

	parentMapUserInfo, errGetAll := getAllUserParentMapInfo(feignKey)
	if nil != errGetAll {
		logrus.WithFields(logrus.Fields{"pubUserid": feignKey.UserId}).Error("Failed to get all user info")
		return nil, errGetAll
	}

	dbUser, errGetUser := GetUserRecordByIdOrPubid(feignKey, "", bwUserReq.Id)
	if nil != errGetUser {
		logrus.WithFields(logrus.Fields{"userId": bwUserReq.Id}).Error("get user info failed")
		return nil, errGetUser
	} else if nil == dbUser {
		logrus.WithFields(logrus.Fields{"userId": bwUserReq.Id}).Error("User info for update not exist")
		return nil, errcode.CerrAccountNotFound
	}

	errCheckPara := parameterCheckForUpdateUserV2(feignKey, bwUserReq, dbUser, parentMapUserInfo)
	if nil != errCheckPara {
		logrus.WithFields(logrus.Fields{"update userId": bwUserReq.Id, "error": errCheckPara}).Error("parameterCheckForUpdateUser failed")
		return nil, errCheckPara
	}

	if (bwUserReq.Email != nil && *bwUserReq.Email != dbUser.Email) || (bwUserReq.Password != nil && *bwUserReq.Password != "") {
		email := IfEmptyChoseOther(bwUserReq.Email, dbUser.Email)
		phone := ""
		if nil != bwUserReq.Phones {
			phone = bwUserReq.Phones.Phone
		} else {
			_, phone, _ = splitPhoneValue(dbUser.Phone)
		}
		password := ""
		if bwUserReq.Password != nil {
			password = *bwUserReq.Password
		}
		if errUpdatePass := updatePass(feignKey, dbUser.PubUserId, email, phone, password); nil != errUpdatePass {
			logrus.WithFields(logrus.Fields{"update userId": bwUserReq.Id, "error": errUpdatePass}).Error("updatePass failed")
			return nil, errUpdatePass
		}
	}

	//请求 用户佣金规则详细-添加修改
	if nil != bwUserReq.Commission {
		bwUserReq.Commission.UserId = dbUser.Id
		bwUserReq.Commission.LevelId = dbUser.LevelId
		if nil != bwUserReq.LevelId {
			bwUserReq.Commission.LevelId, _ = strconv.ParseInt(*bwUserReq.LevelId, 10, 64)
		}
		bwUserReq.Commission.OldParentId, _ = strconv.ParseInt(dbUser.ParentId, 10, 64)
		if nil != bwUserReq.Parent {
			bwUserReq.Commission.ParentId, _ = strconv.ParseInt(*bwUserReq.Parent, 10, 64)
		}
		_, errCommission := ClientAddOrUpdateCommission(feignKey, bwUserReq.Commission)
		if nil != errCommission {
			logrus.WithFields(logrus.Fields{"error": errCommission}).Warn("AddOrUpdateCommission failed")
			return nil, errCommission //添加Commission失败，不返回错误
		}
	}

	//更新用户基本信息到DB
	errUpdate := updateDBUserInfoByReqV2(feignKey, dbUser, bwUserReq)
	if nil != errUpdate {
		logrus.WithFields(logrus.Fields{"userId": bwUserReq.Id, "error": errUpdate}).Error("updateDBUserInfoByReq failed")
		return nil, errUpdate
	}
	logrus.WithFields(logrus.Fields{"user_id": bwUserReq.Id, "operate userPubId": feignKey.UserId}).Info("Updatee user info success")

	//保存黑白名单到DB（暂未实现）	/////////////

	//计算更新相关的用户数（比如下级用户数，层级用户数等）
	go UpdageSubCounter(feignKey)
	//UserMqSender.sendUserMessage(authInfo.tenantId, userDto.getId(), BwEventType.UPDATE);///////////////

	return UserRecord2UserDetail(dbUser)
}

//模糊查询用户，目前按邮箱，名称，用户名，编号查询
func UserMsgReceivers(feignKey *XFeignKey, roleId int64, receiverReq *MsgReceiversSearchDTO) ([]MsgReceiversDTO, code.Error) {
	if nil == feignKey || "" == feignKey.TenantId || nil == receiverReq {
		return nil, errcode.CerrParamater
	}
	var msgList []MsgReceiversDTO = make([]MsgReceiversDTO, 0)
	dbConn, errConn := GetDBConnByTenantID(feignKey.TenantId)
	if errConn != nil {
		return nil, errConn
	}

	var userList []UserDetailTab = make([]UserDetailTab, 0)
	var dbtmp *gorm.DB
	likeFuzzyVal := "%" + escapeSql(receiverReq.FuzzyVal) + "%"
	if roleId > 0 {
		dbtmp = dbConn.Where("(email like ? or phone like ? or name like ? or username like ? or entity_no like ?) and roleId = ? and model_status != ?",
			likeFuzzyVal, likeFuzzyVal, likeFuzzyVal, likeFuzzyVal, likeFuzzyVal, roleId, constant.ModelStatusDelete).Find(&userList)
	} else {
		dbtmp = dbConn.Where("(email like ? or phone like ? or name like ? or username like ? or entity_no like ?) and model_status != ?",
			likeFuzzyVal, likeFuzzyVal, likeFuzzyVal, likeFuzzyVal, likeFuzzyVal, constant.ModelStatusDelete).Find(&userList)
	}
	if err := dbtmp.Error; nil != err {
		if err == gorm.ErrRecordNotFound {
			logrus.Warnf("Not found user record from DB, FuzzyVal:%v", receiverReq.FuzzyVal)
			return msgList, nil //未找到记录不返回错误
		}
		logrus.Errorf("Failed to get user record from DB, FuzzyVal:%v, error:%v", receiverReq.FuzzyVal, err)
		return nil, errcode.CerrExecuteSQL
	}

	for index, _ := range userList {
		if constant.MessageType_SMS == receiverReq.Type { //对于SMS，若无手机号则不返回
			_, phone, phoneStr := splitPhoneValue(userList[index].Phone)
			if "" == phone || "" == phoneStr {
				continue
			}
		} else if constant.MessageType_MAIL == receiverReq.Type {
			if "" == userList[index].Email {
				continue
			}
		}
		var msg MsgReceiversDTO
		msg.Id = userList[index].PubUserId
		msg.Name = userList[index].Name
		msg.IdType = constant.IdType_Id
		msg.RoleName = userList[index].RoleName
		msg.LevelName = userList[index].LevelName
		msg.EntityNo = userList[index].EntityNo
		msgList = append(msgList, msg)
	}
	return msgList, nil
}

// 批量归属,  规则
//1:不可跨层级归属
//2:上级必须归属于当前层级之上
func ProcessUpdateParentBatch(feignKey *XFeignKey, parentId int64, ids []int64) code.Error {
	if nil == feignKey || "" == feignKey.TenantId || parentId <= 0 || len(ids) == 0 {
		return errcode.CerrParamater
	}

	parentUser, errGetParent := GetUserDetailByIdOrPubid(feignKey, "", parentId)
	if nil != errGetParent {
		logrus.WithFields(logrus.Fields{"error": errGetParent}).Error("Get parent user info failed")
		return errGetParent
	} else if nil == parentUser {
		logrus.WithFields(logrus.Fields{"parentId": parentId}).Error("Get parent user info not found")
		return errcode.CerrAccountNotFound
	}

	users, errGetUsers := ProcessListUserRecordsByKey(feignKey, ids, constant.QueryKeyId)
	if nil != errGetUsers {
		logrus.WithFields(logrus.Fields{"error": errGetUsers, "ids": ids}).Error("Get users info failed")
		return errGetUsers
	} else if len(users) == 0 {
		logrus.WithFields(logrus.Fields{"ids": ids}).Error("Get users info not found")
		return errcode.CerrAccountNotFound
	}
	//检查用户level层级
	firstLevelId := users[0].LevelId
	for index, _ := range users {
		if firstLevelId == 0 || users[index].LevelId != firstLevelId {
			logrus.WithFields(logrus.Fields{"level": firstLevelId, "parentId": parentId}).Error("user level not equal")
			return errcode.CerrUserLevelNotEqual
		}
	}

	//检查是否允许设置为当前parent
	allowAsParent, errFind := ProcessFindUserByTypeFuzzy(feignKey, firstLevelId, constant.SEARCH_TYPE_LEVEL, true, "")
	if nil != errFind {
		logrus.WithFields(logrus.Fields{"level": firstLevelId}).Error("ProcessFindUserByTypeFuzzy, find allow as parent failed")
		return errFind
	}
	containParent := false
	for index, _ := range allowAsParent {
		if allowAsParent[index].Id == parentId {
			containParent = true
			break
		}
	}
	if false == containParent {
		logrus.WithFields(logrus.Fields{"level": firstLevelId, "parentId": parentId}).Error("not allow set this parent")
		return errcode.CerrOperateForbidden
	}

	dbConn, errConn := GetDBConnByTenantID(feignKey.TenantId)
	if errConn != nil {
		return errConn
	}
	//更新parent
	if err := dbConn.Model(&UserDetailTab{}).Where("id in (?)", ids).Update("parent_id", strconv.FormatInt(parentId, 10)).Error; nil != err {
		logrus.WithFields(logrus.Fields{"ids": ids, "parentId": parentId, "error": err}).Error("update user parent failed")
		return errcode.CerrExecuteSQL
	}

	return nil
}

//全量计算用户的下级用户数，并更新到DB
func UpdageSubCounter(feignKey *XFeignKey) code.Error {
	allUser, errGetAll := GetAllUserRecords(feignKey)
	if nil != errGetAll {
		logrus.WithFields(logrus.Fields{"error": errGetAll}).Error("getAllUserParentMapInfo failed")
		return errGetAll
	}
	allLevel, errGetLevel := ProcessGetLevelList(feignKey)
	if nil != errGetLevel {
		logrus.WithFields(logrus.Fields{"error": errGetLevel}).Error("ProcessGetLevelList failed")
		return errGetLevel
	}

	type subCount struct {
		oldSubCount int
		newSubCount int
	}
	type userCount struct {
		oldUserCount int
		newUserCount int
	}

	leverCountMap := make(map[int64]userCount) //key为levelId
	for index, _ := range allLevel {
		var counter userCount
		counter.oldUserCount = allLevel[index].UserCount
		counter.newUserCount = 0
		leverCountMap[allLevel[index].Id] = counter
	}

	userIdSubCountMap := make(map[int64]subCount) //计算得到的下级用户数
	idParentMap := make(map[int64]int64)          //key为userId，value为parentId
	subUserMap := make(map[int64][]int64)         //key为parentid, value为直接下级用户ID
	for index, _ := range allUser {
		var counter subCount
		counter.oldSubCount = allUser[index].SubUserCount //保存原始的下级用户数
		counter.newSubCount = 0
		userIdSubCountMap[allUser[index].Id] = counter
		parentId, errParse := strconv.ParseInt(allUser[index].ParentId, 10, 64)
		if nil != errParse {
			parentId = 0
		}
		//构造parentid为key，subuserList为value的map
		subList, exist := subUserMap[parentId]
		if false == exist {
			subList = make([]int64, 0)
		}

		subList = append(subList, allUser[index].Id)
		subUserMap[parentId] = subList
		idParentMap[allUser[index].Id] = parentId

		leverCounter, existLevel := leverCountMap[allUser[index].LevelId] //更新level计数
		if existLevel {
			leverCounter.newUserCount++
			leverCountMap[allUser[index].LevelId] = leverCounter
		}
	}

	for parentId, subList := range subUserMap {
		depth := 0
		for parentId > 0 && depth < 100 {
			if counter, exist := userIdSubCountMap[parentId]; exist {
				counter.newSubCount += len(subList)
				userIdSubCountMap[parentId] = counter
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
	var userRecord UserDetailTab
	for userId, counter := range userIdSubCountMap {
		if counter.newSubCount == counter.oldSubCount {
			continue
		}
		userRecord.Id = userId
		err := dbConn.Model(&userRecord).Where("id = ?", userId).Update("sub_user_count", counter.newSubCount).Error
		if nil != err {
			logrus.WithFields(logrus.Fields{"error": err, "userId": userId, "subUserCount": counter.newSubCount}).Error("update sub_user_count failed")
			//不返回错误
		}
		logrus.WithFields(logrus.Fields{"userId": userId, "subUserCount": counter.newSubCount}).Debug("update sub_user_count success")
	}
	//更新level的用户数到DB
	var levelRecord LevelTab
	for levelId, counter := range leverCountMap {
		if counter.oldUserCount == counter.newUserCount {
			continue
		}
		levelRecord.Id = levelId
		err := dbConn.Model(&levelRecord).Where("id = ?", levelId).Update("user_count", counter.newUserCount).Error
		if nil != err {
			logrus.WithFields(logrus.Fields{"error": err, "userId": levelId, "subUserCount": counter.newUserCount}).Error("update level user count failed")
			//不返回错误
		}
		logrus.WithFields(logrus.Fields{"levelId": levelId, "subUserCount": counter.newUserCount}).Debug("update level user count success")
	}

	return nil
}

func GetUserFiledsFromMongo(tenantId, userId, tableName string) (*UserFields, code.Error) {
	mgoSess, errSess := GetMongoConnByTenantID(tenantId)
	if nil != errSess {
		logrus.WithFields(logrus.Fields{"tenant_id": tenantId, "error": errSess}).Error("GetMongoConnByTenantID failed")
		return nil, errSess
	}
	defer mgoSess.Close()
	coll := GetMgoUserFieldColl(mgoSess)

	var record UserFields
	condition := bson.M{"tenantId": tenantId, "createUserId": userId, "tableName": tableName, "enabled": true}
	errGet := coll.Find(condition).One(&record)
	if nil != errGet {
		if mgo.ErrNotFound == errGet {
			logrus.WithFields(logrus.Fields{"tenant_id": tenantId, "tableName": tableName, "userId": userId}).Warn("Get user fields record not found")
			return nil, nil
		}
		logrus.WithFields(logrus.Fields{"tenant_id": tenantId, "error": errGet}).Error("Get user fields record failed")
		return nil, errcode.CerrOperateMongo
	}
	return &record, nil
}

func ProcessGetUserFieldsList(feignKey *XFeignKey, tableName string) ([]FormFieldDTO, code.Error) {
	if nil == feignKey || "" == feignKey.TenantId || "" == tableName {
		return nil, errcode.CerrParamater
	}

	fieldsForm, errGetField := ClientFormfieldList(feignKey, tableName)
	if nil != errGetField {
		logrus.WithFields(logrus.Fields{"tableName": tableName, "error": errGetField}).Error("ClientFormfieldSimpleList failed")
		return nil, errGetField
	}
	logrus.Infof("ProcessGetUserFieldsList   fieldsForm: %s", fieldsForm)
	userFields, errGetFromDB := GetUserFiledsFromMongo(feignKey.TenantId, feignKey.UserId, tableName)
	if nil != errGetFromDB {
		logrus.WithFields(logrus.Fields{"tableName": tableName, "error": errGetFromDB}).Error("GetUserFiledsFromMongo failed")
		return nil, errGetFromDB
	}
	logrus.Infof("ProcessGetUserFieldsList   userFields:%s", userFields)
	//过滤非常用字段
	var systemFields []FormFieldDTO
	for index := range fieldsForm {
		if fieldsForm[index].Overuse == true {
			systemFields = append(systemFields, fieldsForm[index])
		}
	}

	list := make([]FormFieldDTO, 0)
	if len(systemFields) > 0 && nil != userFields {
		formMap := make(map[string]FormFieldDTO)
		for index, _ := range systemFields {
			formMap[systemFields[index].Key] = systemFields[index]
		}

		//按设定顺序排列
		for _, f := range userFields.UserFields {
			if formFieldDTO, ok := formMap[f.Key]; !ok {
				//特殊字段
				if StringListIsContain(constant.DefaultFieldKeys, f.Key) {
					list = append(list, FormFieldDTO{Key: f.Key, Label: f.Label, Show: f.Show})
				}
			} else {
				//定义字段
				formFieldDTO.Show = f.Show
				list = append(list, formFieldDTO)
			}
		}

		//补充新增字段（因为自定义字段可能会新增，或者之前关闭又打开等）
		fieldsMap := make(map[string]UserField)
		for index, _ := range userFields.UserFields {
			fieldsMap[userFields.UserFields[index].Key] = userFields.UserFields[index]
		}
		for _, f := range systemFields {
			if _, ok := fieldsMap[f.Key]; !ok {
				list = append(list, f)
			}
		}

		//上线后新增字段：
		if strings.Contains(tableName, "t_customer_profiles,t_customer_follow") {
			//上线后新增字段：
			if _, ok := fieldsMap["accounts"]; !ok {
				list = append(list, FormFieldDTO{Key: "accounts", Label: "accounts", Show: true})
			}

			if _, ok := fieldsMap["customerState"]; !ok {
				list = append(list, FormFieldDTO{Key: "customerState", Label: "customerState", Show: true})
			}
			if _, ok := fieldsMap["openState"]; !ok {
				list = append(list, FormFieldDTO{Key: "openState", Label: "openState", Show: true})
			}
			if _, ok := fieldsMap["dealState"]; !ok {
				list = append(list, FormFieldDTO{Key: "dealState", Label: "dealState", Show: true})
			}
			if _, ok := fieldsMap["recommendedCustomerNum"]; !ok {
				list = append(list, FormFieldDTO{Key: "recommendedCustomerNum", Label: "recommendedCustomerNum", Show: true})
			}
		}
		//加上特殊固定字段
		if strings.Contains(tableName, "t_account_account,t_account_profiles,t_account_finacial,t_account_id_info") {
			if _, ok := fieldsMap["balance"]; !ok {
				list = append(list, FormFieldDTO{Key: "balance", Label: "balance", Show: true})
			}
			if _, ok := fieldsMap["profit"]; !ok {
				list = append(list, FormFieldDTO{Key: "profit", Label: "profit", Show: true})
			}
			if _, ok := fieldsMap["equity"]; !ok {
				list = append(list, FormFieldDTO{Key: "equity", Label: "equity", Show: true})
			}
			if _, ok := fieldsMap["margin"]; !ok {
				list = append(list, FormFieldDTO{Key: "margin", Label: "margin", Show: true})
			}
			if _, ok := fieldsMap["marginFree"]; !ok {
				list = append(list, FormFieldDTO{Key: "marginFree", Label: "marginFree", Show: true})
			}
			if _, ok := fieldsMap["marginLevel"]; !ok {
				list = append(list, FormFieldDTO{Key: "marginLevel", Label: "marginLevel", Show: true})
			}
			if _, ok := fieldsMap["credit"]; !ok {
				list = append(list, FormFieldDTO{Key: "credit", Label: "credit", Show: true})
			}
		}
		//加上特殊固定字段
		if strings.Contains(tableName, "t_account_cbroker,t_account_profiles,t_account_finacial,t_account_id_info") {
			if _, ok := fieldsMap["balance"]; !ok {
				list = append(list, FormFieldDTO{Key: "balance", Label: "balance", Show: true})
			}
			if _, ok := fieldsMap["equity"]; !ok {
				list = append(list, FormFieldDTO{Key: "equity", Label: "equity", Show: true})
			}
		}
		logrus.Infof("1. ProcessGetUserFieldsList   list:%s", list)

		return list, nil
	} else {
		list = append(list, systemFields...)
		//加上特殊固定字段
		if "t_user_profiles" == tableName {
			list = append(list, FormFieldDTO{Key: "id", Label: "ID", Show: true})
			list = append(list, FormFieldDTO{Key: "subUserCount", Label: "lower", Show: true})
			list = append(list, FormFieldDTO{Key: "createDate", Label: "create_time", Show: true})
			list = append(list, FormFieldDTO{Key: "active", Label: "login_status", Show: true})
			list = append(list, FormFieldDTO{Key: "ownAccounts", Label: "own_accounts", Show: true})
			list = append(list, FormFieldDTO{Key: "ownCustomers", Label: "own_customers", Show: true})
		}
		//加上特殊固定字段
		if strings.Contains(tableName, "t_account_account,t_account_profiles,t_account_finacial,t_account_id_info") {
			list = append(list, FormFieldDTO{Key: "balance", Label: "balance", Show: true})
			list = append(list, FormFieldDTO{Key: "profit", Label: "profit", Show: true})
			list = append(list, FormFieldDTO{Key: "equity", Label: "equity", Show: true})
			list = append(list, FormFieldDTO{Key: "margin", Label: "margin", Show: true})
			list = append(list, FormFieldDTO{Key: "marginFree", Label: "marginFree", Show: true})
			list = append(list, FormFieldDTO{Key: "marginLevel", Label: "marginLevel", Show: true})
			list = append(list, FormFieldDTO{Key: "credit", Label: "credit", Show: true})
		}
		//加上特殊固定字段
		if strings.Contains(tableName, "t_account_cbroker,t_account_profiles,t_account_finacial,t_account_id_info") {
			list = append(list, FormFieldDTO{Key: "balance", Label: "balance", Show: true})
			list = append(list, FormFieldDTO{Key: "equity", Label: "equity", Show: true})
		}
		//加上特殊固定字段
		if strings.Contains(tableName, "t_customer_profiles,t_customer_follow") {
			list = append(list, FormFieldDTO{Key: "accounts", Label: "accounts", Show: true})
			list = append(list, FormFieldDTO{Key: "customerState", Label: "customerState", Show: true})
			list = append(list, FormFieldDTO{Key: "recommendedCustomerNum", Label: "recommendedCustomerNum", Show: true})
			list = append(list, FormFieldDTO{Key: "openState", Label: "openState", Show: true})
			list = append(list, FormFieldDTO{Key: "dealState", Label: "dealState", Show: true})
		}
	}
	logrus.Infof("2. ProcessGetUserFieldsList   list:%s", list)
	return list, nil
}

func UpserFiledsToMongo(record *UserFields) code.Error {
	if nil == record || "" == record.FieldsId || "" == record.TenantId {
		return errcode.CerrParamater
	}
	mgoSess, errSess := GetMongoConnByTenantID(record.TenantId)
	if nil != errSess {
		logrus.WithFields(logrus.Fields{"tenant_id": record.TenantId, "error": errSess}).Error("GetMongoConnByTenantID failed")
		return errSess
	}
	defer mgoSess.Close()

	selector := bson.M{"_id": record.FieldsId}
	updator := bson.M{
		"_id":          record.FieldsId,
		"tenantId":     record.TenantId,
		"tableName":    record.TableName,
		"userFields":   record.UserFields,
		"enabled":      record.Enabled,
		"createUserId": record.CreateUserId,
		"modifyUserId": record.ModifyUserId,
		"createTime":   record.CreateTime,
		"modifyTime":   record.ModifyTime,
	}
	if _, errUpsert := GetMgoUserFieldColl(mgoSess).Upsert(selector, updator); nil != errUpsert {
		logrus.WithFields(logrus.Fields{"error": errUpsert}).Error("Insert into mongo failed")
		return errcode.CerrOperateMongo
	}

	return nil
}

func getUserFieldLabel(key, label string) string {
	switch key {
	//用户特殊字段
	case "id":
		return "ID"
	case "subUserCount":
		return "lower"
	case "createDate":
		return "create_time"
	case "active":
		return "login_status"
		//账户特殊字段
	case "balance":
		return "balance"
	case "profit":
		return "profit"
	case "equity":
		return "equity"
	case "ownAccounts":
		return "own_accounts"
	case "ownCustomers":
		return "own_customers"
		//客户特殊字段
	case "accounts":
		return "accounts"
	case "customerState":
		return "customerState"
	case "recommendedCustomerNum":
		return "recommendedCustomerNum"
	case "openState":
		return "openState"
	case "dealState":
		return "dealState"
	default:
		return label
	}
}

func ProcessUpdateUserFields(feignKey *XFeignKey, updateReq *UserFieldsDTO) code.Error {
	if nil == feignKey || "" == feignKey.TenantId || nil == updateReq {
		return errcode.CerrParamater
	}

	fieldDB, errGetFromDB := GetUserFiledsFromMongo(feignKey.TenantId, feignKey.UserId, updateReq.TableName)
	if nil != errGetFromDB {
		logrus.WithFields(logrus.Fields{"tableName": updateReq.TableName, "error": errGetFromDB}).Error("GetUserFiledsFromMongo failed")
		return errGetFromDB
	}
	if nil == fieldDB {
		fieldDB = &UserFields{
			FieldsId:     GenerateNewUUID(),
			TenantId:     feignKey.TenantId,
			TableName:    updateReq.TableName,
			UserFields:   updateReq.UserFields,
			Enabled:      true,
			CreateUserId: feignKey.UserId,
			ModifyUserId: feignKey.UserId,
			CreateTime:   time.Now().Unix() * 1000,
			ModifyTime:   time.Now().Unix() * 1000,
		}
	} else {
		fieldDB.UserFields = updateReq.UserFields
		fieldDB.ModifyUserId = feignKey.UserId
		fieldDB.ModifyTime = time.Now().Unix() * 1000
	}

	if "t_user_profiles" == updateReq.TableName ||
		strings.Contains(updateReq.TableName, "t_customer_profiles,t_customer_follow") ||
		strings.Contains(updateReq.TableName, "t_account_cbroker,t_account_profiles,t_account_finacial,t_account_id_info") ||
		strings.Contains(updateReq.TableName, "t_account_account,t_account_profiles,t_account_finacial,t_account_id_info") {
		for index, _ := range updateReq.UserFields {
			updateReq.UserFields[index].Label = getUserFieldLabel(updateReq.UserFields[index].Key, updateReq.UserFields[index].Label)
		}
	}
	return UpserFiledsToMongo(fieldDB)
}

// 按角色统计用户数
func UserCountGroupByRole(tenantId string) ([]UserRoleStat, code.Error) {
	dbConn, errConn := GetDBConnByTenantID(tenantId)
	if errConn != nil {
		return nil, errConn
	}
	stats := make([]UserRoleStat, 0)
	rows, errRaw := dbConn.Table("t_user_detail").Select("count(*) as user_count, role_id").Group("role_id").Rows()
	defer rows.Close()
	if nil != errRaw {
		if errRaw == gorm.ErrRecordNotFound {
			logrus.WithFields(logrus.Fields{}).Warn("Not found user records from DB")
			return stats, nil //未找到记录不返回错误
		}
		logrus.WithFields(logrus.Fields{"error": errRaw}).Error("Failed to stat user count from DB")
		return nil, errcode.CerrExecuteSQL
	}
	for rows.Next() {
		var record UserRoleStat
		dbConn.ScanRows(rows, &record)
		stats = append(stats, record)
	}
	return stats, nil
}

func GetUserIdToRoleIdMap(feignKey *XFeignKey, ids []string, key string) (map[string]string, code.Error) {
	userDetailTabList, err := ProcessListUserRecordsByKey(feignKey, ids, key)
	if nil != err {
		logrus.WithFields(logrus.Fields{"error": err, "TenantId": feignKey.TenantId, "ids": ids}).Error("ProcessListUserRecordsByKey Failed")
		return nil, err
	}
	pubIdToRoleIdMap := make(map[string]string)
	for _, userDetailTab := range userDetailTabList {
		pubIdToRoleIdMap[userDetailTab.PubUserId] = strconv.FormatInt(userDetailTab.RoleId, 10)
	}
	return pubIdToRoleIdMap, nil
}

//更新二次验证配置
func ProcessUpdateTwoFAConfig(feignKey *XFeignKey, bwUserReq *BWUserDTO) code.Error {
	if nil == feignKey || "" == feignKey.TenantId || nil == bwUserReq {
		logrus.WithFields(logrus.Fields{"feignKey": feignKey}).Error("parameters failed")
		return errcode.CerrParamater
	}
	userRecord, errGetUser := GetUserRecordByIdOrPubid(feignKey, bwUserReq.PubUserId, 0)
	if nil != errGetUser {
		logrus.WithFields(logrus.Fields{"errGetUser": errGetUser}).Error("GetUserRecordByIdOrPubid failed")
		return errGetUser
	} else if nil == userRecord {
		return errcode.CerrAccountNotFound
	}
	userRecord.TwoFactorAuth = bwUserReq.TwoFactorAuth
	dbConn, errConn := GetDBConnByTenantID(feignKey.TenantId)
	if errConn != nil {
		return errConn
	}
	if err := dbConn.Save(userRecord).Error; nil != err {
		logrus.WithFields(logrus.Fields{"errGetUser": err, "tenant_id": feignKey.TenantId, "userId": bwUserReq.Id}).Error("update two fa config failed")
		return errcode.CerrExecuteSQL
	}
	return nil
}

//TODO 更新用户积分
func ProcessUpdateUserPoints(feignKey *XFeignKey, bwUserReq *BWUserDTO) code.Error {

	dbConn, errConn := GetDBConnByTenantID(feignKey.TenantId)
	if errConn != nil {
		return errConn
	}

	var userDetail UserDetail

	//sqlMap := map[string]interface{}{"points1": bwUserReq.Points1, "points2": bwUserReq.Points2, "points3": bwUserReq.Points3, "points4": bwUserReq.Points4, "points5": bwUserReq.Points5, "points6": bwUserReq.Points6, "points7": bwUserReq.Points7}
	sqlMap := getSqlMap(bwUserReq)
	dbConn.Table("t_user_detail").Model(&userDetail).Where("id = ?", bwUserReq.Id).Updates(sqlMap)
	return nil
}

func getSqlMap(dto *BWUserDTO) map[string]interface{} {
	sqlMap := make(map[string]interface{})
	if dto.Points1 != "" {
		if _, errStr := strconv.ParseFloat(dto.Points1, 64); errStr == nil {
			sqlMap["points1"] = dto.Points1
		}
	}
	if dto.Points2 != "" {
		if _, errStr := strconv.ParseFloat(dto.Points2, 64); errStr == nil {
			sqlMap["points2"] = dto.Points2
		}
	}
	if dto.Points3 != "" {
		if _, errStr := strconv.ParseFloat(dto.Points3, 64); errStr == nil {
			sqlMap["points3"] = dto.Points3
		}
	}
	if dto.Points4 != "" {
		if _, errStr := strconv.ParseFloat(dto.Points4, 64); errStr == nil {
			sqlMap["points4"] = dto.Points4
		}
	}
	if dto.Points5 != "" {
		if _, errStr := strconv.ParseFloat(dto.Points5, 64); errStr == nil {
			sqlMap["points5"] = dto.Points5
		}
	}
	if dto.Points6 != "" {
		if _, errStr := strconv.ParseFloat(dto.Points6, 64); errStr == nil {
			sqlMap["points6"] = dto.Points6
		}
	}
	if dto.Points7 != "" {
		if _, errStr := strconv.ParseFloat(dto.Points7, 64); errStr == nil {
			sqlMap["points7"] = dto.Points7
		}
	}
	return sqlMap
}

//获取用户积分
func ProcessGetUserPoints(feignKey *XFeignKey, id int) (UserDetail, code.Error) {
	dbConn, errConn := GetDBConnByTenantID(feignKey.TenantId)
	if errConn != nil {
		return UserDetail{}, errConn
	}
	var userDetail UserDetail
	dbConn.Table("t_user_detail").Select("points1,points2,points3,points4,points5,points6,points7,active,name").Where("id = ?", id).Find(&userDetail)
	return userDetail, nil
}

func ProcessGetUserIDByEmail(feignKey *XFeignKey, email string) (int64, code.Error) {
	dbConn, errConn := GetDBConnByTenantID(feignKey.TenantId)
	if errConn != nil {
		return 0, errConn
	}
	var userDetail UserDetail
	dbConn.Table("t_user_detail").Select("id").Where("email = ?", email).Find(&userDetail)
	return userDetail.Id, nil
}
