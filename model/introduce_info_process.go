package model

import (
	"bw/bw-user/constant"
	"bw/bw-user/errcode"
	"github.com/Sirupsen/logrus"
	"github.com/jinzhu/gorm"
	"github.com/lworkltd/kits/service/restful/code"
	"time"

	"bw/bw-user/conf"
	"bytes"
	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
	"gopkg.in/mgo.v2/bson"
	"image/png"
	"strings"
	//"net/url"
	"strconv"
)

func ExistInIntroduceTab(tenant_id string, key string, value interface{}) (bool, code.Error) {
	if "" == tenant_id || "" == key {
		return false, errcode.CerrParamater
	}
	dbConn, errConn := GetDBConnByTenantID(tenant_id)
	if errConn != nil {
		return false, errConn
	}

	var introduceRecord SystemIntroduceTab
	var dbtmp *gorm.DB
	if constant.QueryKeyEntityNo == key {
		dbtmp = dbConn.Where("entity_no = ? and model_status != ?", value, constant.ModelStatusDelete).First(&introduceRecord)
	} else if constant.QueryKeyName == key {
		dbtmp = dbConn.Where("name = ? and model_status != ?", value, constant.ModelStatusDelete).First(&introduceRecord)
	} else {
		logrus.WithFields(logrus.Fields{"key": key}).Error("Query Key Not Support")
		return false, errcode.CerrQueryKeyNotSupport
	}

	if err := dbtmp.Error; nil != err {
		if err == gorm.ErrRecordNotFound {
			logrus.WithFields(logrus.Fields{"key": key, "value": value}).Debug("Not found introduce record from DB")
			return false, nil //无记录不返回错误
		}
		logrus.WithFields(logrus.Fields{"error": err}).Error("Failed to get introduce record from DB by ids")
		return false, errcode.CerrExecuteSQL
	}

	logrus.WithFields(logrus.Fields{"key": key, "value": value}).Debug("Found introduce records from DB")
	return true, nil
}

func GetIntroduceRecordsById(feignKey *XFeignKey, introduceId int64) (*SystemIntroduceTab, code.Error) {
	dbConn, errConn := GetDBConnByTenantID(feignKey.TenantId)
	if errConn != nil {
		return nil, errConn
	}

	var introduceRecord SystemIntroduceTab
	dbtmp := dbConn.Where("id = ? and model_status != ?", introduceId, constant.ModelStatusDelete).First(&introduceRecord)

	if err := dbtmp.Error; nil != err {
		if err == gorm.ErrRecordNotFound {
			logrus.WithFields(logrus.Fields{"introduceId": introduceId}).Warn("Not found introduce record from DB")
			return nil, nil //未找到记录不返回错误
		}
		logrus.WithFields(logrus.Fields{"error": err, "introduceId": introduceId}).Error("Failed to get introduces record from DB")
		return nil, errcode.CerrExecuteSQL
	}

	return &introduceRecord, nil
}

//根据t_system_introduce表中的keyType列获取值在values中的记录记录, values为切片
//若keyType为QueryKeyAllRecords，则返回所有记录
func ListIntroduceRecordsByKey(feignKey *XFeignKey, values interface{}, keyType string) ([]SystemIntroduceTab, code.Error) {
	dbConn, errConn := GetDBConnByTenantID(feignKey.TenantId)
	if errConn != nil {
		return nil, errConn
	}

	var introduceRecords []SystemIntroduceTab = make([]SystemIntroduceTab, 0)
	var dbtmp *gorm.DB
	switch keyType {
	case constant.QueryKeyId:
		dbtmp = dbConn.Where("id in (?) and model_status != ?", values, constant.ModelStatusDelete).Find(&introduceRecords)
	case constant.QueryKeyName:
		dbtmp = dbConn.Where("name in (?) and model_status != ?", values, constant.ModelStatusDelete).Find(&introduceRecords)
	case constant.QueryKeyEntityNo:
		dbtmp = dbConn.Where("entity_no in (?) and model_status != ?", values, constant.ModelStatusDelete).Find(&introduceRecords)
	case constant.QueryKeyAllRecords:
		dbtmp = dbConn.Where(" model_status != ?", constant.ModelStatusDelete).Find(&introduceRecords)
	default:
		logrus.WithFields(logrus.Fields{"keyType": keyType}).Error("ListIntroduceRecordsByKey keyType not support")
		return nil, errcode.CerrQueryKeyNotSupport
	}

	if err := dbtmp.Error; nil != err {
		if err == gorm.ErrRecordNotFound {
			logrus.WithFields(logrus.Fields{"keyType": keyType, "value": values}).Warn("Not found introduce records from DB")
			return introduceRecords, nil //未找到记录不返回错误
		}
		logrus.WithFields(logrus.Fields{"error": err, "keyType": keyType, "value": values}).Error("Failed to get introduces records from DB")
		return nil, errcode.CerrExecuteSQL
	}

	return introduceRecords, nil
}

func introduceParameterCheck(tenant_id string, record *SystemIntroduceTab, checkNameExist bool) code.Error {
	if "" == record.Name || len(record.Name) >= 255 {
		logrus.WithFields(logrus.Fields{}).Error("Add introduce info, name abnormal")
		return errcode.CerrParamater
	}
	if constant.IntroduceType_Mobile != record.Platform && constant.IntroduceType_Web != record.Platform {
		logrus.WithFields(logrus.Fields{"platform": record.Platform}).Error("Platform abnormal")
		return errcode.CerrParamater
	}

	if checkNameExist {
		exist, errCheck := ExistInIntroduceTab(tenant_id, constant.QueryKeyName, record.Name)
		if nil != errCheck {
			logrus.WithFields(logrus.Fields{"error": errCheck}).Error("ExistInIntroduceTab failed")
			return errCheck
		} else if exist {
			logrus.WithFields(logrus.Fields{"name": record.Name}).Error("Add introduce info, name have exist")
			return errcode.CerrNameExist
		}
	}

	if constant.IntroduceType_Web == record.Platform {
		if "" == record.Url || len(record.Url) >= 255 || "" == record.BwUserShow {
			logrus.WithFields(logrus.Fields{"Platform": record.Platform}).Error("Add introduce info, url/bwUserShow abnormal")
			return errcode.CerrParamater
		}
		//参数不可包括pid,eid,id
		urlSplits := strings.Split(record.Url, "?")
		if len(urlSplits) >= 2 {
			paras := strings.Split(urlSplits[1], "&")
			for _, para := range paras {
				if strings.Index(para, "pid=") == 0 || strings.Index(para, "eid=") == 0 || strings.Index(para, "id=") == 0 {
					logrus.WithFields(logrus.Fields{"url": record.Url}).Error("Add introduce info, url parameter error")
					return errcode.CerrParamater
				}
			}
		}

		if constant.IntroduceType_UserPartVisible == record.BwUserShow && (len(record.VisibleUser) == 0 || len(record.VisibleUserName) == 0) {
			logrus.WithFields(logrus.Fields{}).Error("Add introduce info, UserPartVisible, VisibleUser/VisibleUserName is empty")
			return errcode.CerrParamater
		}
	}
	return nil
}

func generateIntroduceDisplayUrl(productInfo *TenantProductDTO, introduceEntityNo string, parameters map[string]string) (string, string, code.Error) {
	domain := productInfo.CustomerDomain

	if "" == domain {
		return "", "", errcode.CerrProductDomainNotSet
	}

	if "" == introduceEntityNo {
		logrus.WithFields(logrus.Fields{"CustomerDomain": productInfo.CustomerDomain, "ProductDomain": productInfo.ProductDomain, "introduceEntityNo": introduceEntityNo}).Error("domain/introduceEntityNo is empty")
		return "", "", errcode.CerrProductInfoAbnormal
	}

	scheme := productInfo.CustomerDomainScheme
	if scheme == "" {
		scheme = "http"
	} else {
		scheme = strings.TrimRight(scheme, "://")
		scheme = strings.TrimRight(scheme, ":")
	}

	domain = strings.TrimSpace(domain)
	domain = strings.Replace(domain, "\\\\", "/", -1)
	if strings.HasPrefix(domain, "//") {
		domain = scheme + ":" + domain
	} else if false == strings.HasPrefix(domain, "http") {
		domain = scheme + "://" + domain
	}

	var agentApplyUrl string
	var introduceUrl string
	if strings.HasSuffix(domain, "/") {
		agentApplyUrl = domain + "agentApply"
		introduceUrl = domain + "introduce?"
	} else {
		agentApplyUrl = domain + "/agentApply"
		introduceUrl = domain + "/introduce?"
	}
	introduceUrl += "iid=" + introduceEntityNo
	if parameters != nil {
		for parameter := range parameters {
			introduceUrl += "&" + parameter + "=" + parameters[parameter]
		}
	}

	return agentApplyUrl, introduceUrl, nil
}

//新添加Introduce
func ProcessAddIntroduce(feignKey *XFeignKey, addReq *SystemIntroduceDTO) (*SystemIntroduceDTO, code.Error) {
	if nil == feignKey || "" == feignKey.TenantId || nil == addReq {
		return nil, errcode.CerrParamater
	}
	addReq.Name = strings.TrimSpace(addReq.Name)
	addReq.Url = strings.TrimSpace(addReq.Url)
	if "" == addReq.Name || len(addReq.Name) > constant.MAX_COMM_LENGTH || len(addReq.Url) > constant.MAX_COMM_LENGTH || ("" == addReq.Url && constant.IntroduceType_StraightGuest == addReq.Type) {
		return nil, errcode.CerrParamater
	}

	record, errConv := SystemIntroduceDTO2SystemIntroduceTab(addReq)
	if nil != errConv {
		return nil, errConv
	}

	//推广链接编码为空时自动生成，如果外部传入则判断是否重复
	if record.EntityNo == "" {
		record.EntityNo = generateIntroduceEntityNo(feignKey.TenantId)
	} else {
		exist, errExist := ExistInIntroduceTab(feignKey.TenantId, constant.QueryKeyEntityNo, record.EntityNo)
		if errExist != nil {
			return nil, errExist
		}

		if exist == true {
			return nil, errcode.CerreEntityExist
		}
	}

	//生成代理自动填充目标URL和展示的URL
	product, errGetProduct := ClientProductTenantByKey(feignKey, feignKey.TenantId, "BW")
	if nil != errGetProduct {
		logrus.WithFields(logrus.Fields{"tenant_id": feignKey.TenantId}).Error("Request client product tenant by key failed")
		return nil, errGetProduct
	}
	agentApplyUrl, introduceUrl, err := generateIntroduceDisplayUrl(product, record.EntityNo, nil)
	if nil != err {
		logrus.WithFields(logrus.Fields{"error": err}).Error("generateIntroduceDisplayUrl failed")
		return nil, err
	}
	record.DisplayUrl = introduceUrl //展示的URL
	if constant.IntroduceType_Web == addReq.Platform && constant.IntroduceType_Agent == addReq.Type {
		record.Url = agentApplyUrl //代理自动填充目标URL
	}

	if err := introduceParameterCheck(feignKey.TenantId, record, true); nil != err {
		return nil, err
	}

	record.CreateDate = time.Now()
	record.CreateUserId = feignKey.UserId
	record.ModelStatus = constant.ModelStatusCreate
	record.ModifyDate = record.CreateDate
	record.ModifyUserId = feignKey.UserId

	dbConn, errConn := GetDBConnByTenantID(feignKey.TenantId)
	if errConn != nil {
		return nil, errConn
	}
	if err := dbConn.Create(record).Error; nil != err {
		logrus.WithFields(logrus.Fields{"error": err}).Error("Save introduce into to DB failed")
		return nil, errcode.CerrExecuteSQL
	}

	detail := SystemIntroduceTab2SystemIntroduceDTO(record)

	return detail, nil
}

func SystemIntroduceDTO2SystemIntroduceTab(addReq *SystemIntroduceDTO) (*SystemIntroduceTab, code.Error) {
	var record SystemIntroduceTab
	record.Id = addReq.Id
	record.TenantId = addReq.TenantId
	record.Name = addReq.Name
	record.Platform = addReq.Platform
	record.BwUserShow = addReq.BwUserShow
	record.EntityNo = addReq.EntityNo
	if "" == addReq.BwUserShow {
		record.BwUserShow = constant.IntroduceType_UserNotVisible
	}
	record.ProductId = constant.ProductId
	record.Enable = 1
	if addReq.Platform == constant.IntroduceType_Web {
		record.Type = addReq.Type
		record.Url = addReq.Url
		switch addReq.BwUserShow {
		case constant.IntroduceType_UserPartVisible, constant.IntroduceType_DirectPartVisible:
			if len(addReq.VisibleUser) == 0 || len(addReq.VisibleUserName) == 0 {
				logrus.WithFields(logrus.Fields{}).Error("VisibleUser or VisibleUserName is empty")
				return nil, errcode.CerrParamater
			}
			record.VisibleUser = strings.Join(addReq.VisibleUser, "@-@")
			record.VisibleUserName = strings.Join(addReq.VisibleUserName, "@-@")
			break
		case constant.IntroduceType_UserInVisible, constant.IntroduceType_DirectPartInvisible:
			if len(addReq.InVisibleUser) == 0 || len(addReq.InVisibleUserName) == 0 {
				logrus.WithFields(logrus.Fields{}).Error("InVisibleUser or InVisibleUserName is empty")
				return nil, errcode.CerrParamater
			}
			record.InVisibleUser = strings.Join(addReq.InVisibleUser, "@-@")
			record.InVisibleUserName = strings.Join(addReq.InVisibleUserName, "@-@")
			break
		default:
		}
		if len(addReq.Participants) != 0 || len(addReq.ParticipantNames) != 0 {
			record.Participants = strings.Join(addReq.Participants, "@-@")
			record.ParticipantNames = strings.Join(addReq.ParticipantNames, "@-@")
		}

		if constant.IntroduceType_StraightGuest == addReq.Type {
			record.ParameterType = addReq.ParameterType
			record.ServerId = addReq.ServerId
			record.Vendor = addReq.Vendor
			record.MtGroup = addReq.MtGroup
			record.AccountGroup = addReq.AccountGroup
			record.Leverage = addReq.Leverage
		} else if constant.IntroduceType_DirectRecommendation == addReq.Type {
			record.ServerId = addReq.ServerId
			record.Vendor = addReq.Vendor
			record.MtGroup = addReq.MtGroup
			record.AccountGroup = addReq.AccountGroup
			record.Leverage = addReq.Leverage
		}

		record.OwnerType = addReq.OwnerType
		record.OwnerId = addReq.OwnerId
	}
	if constant.IntroduceType_Mobile == addReq.Platform {
		record.BwUserShow = constant.IntroduceType_UserNotVisible
		record.BusinessCode = addReq.BusinessCode
		record.Os = addReq.Os
		record.SoftwarePackage = addReq.SoftwarePackage
	}
	return &record, nil
}

func SystemIntroduceTab2SystemIntroduceDTO(record *SystemIntroduceTab) *SystemIntroduceDTO {
	var detail SystemIntroduceDTO
	detail.Id = record.Id
	detail.TenantId = record.TenantId
	detail.Creator = record.CreateUserId
	detail.CreateTime = record.CreateDate.Unix() * 1000
	detail.EntityNo = record.EntityNo
	detail.Platform = record.Platform
	detail.Name = record.Name
	detail.Enable = false
	if record.Enable != 0 {
		detail.Enable = true
	}
	detail.Type = record.Type
	detail.BwUserShow = record.BwUserShow
	if "" != record.VisibleUser {
		detail.VisibleUser = strings.Split(record.VisibleUser, "@-@")
		//处理老数据，只有用户可见范围的需要添加
		if strings.Contains(detail.BwUserShow, "User") {
			for index := range detail.VisibleUser {
				if !strings.Contains(detail.VisibleUser[index], "-") {
					detail.VisibleUser[index] = detail.VisibleUser[index] + "-" + constant.IdType_Id
				}
			}
		}
	}
	if "" != record.VisibleUserName {
		detail.VisibleUserName = strings.Split(record.VisibleUserName, "@-@")
	}
	if "" != record.InVisibleUser {
		detail.InVisibleUser = strings.Split(record.InVisibleUser, "@-@")
		//处理老数据，只有用户可见范围的需要添加
		if strings.Contains(detail.BwUserShow, "User") {
			for index := range detail.InVisibleUser {
				if !strings.Contains(detail.InVisibleUser[index], "-") {
					detail.InVisibleUser[index] = detail.InVisibleUser[index] + "-" + constant.IdType_Id
				}
			}
		}
	}
	if "" != record.InVisibleUserName {
		detail.InVisibleUserName = strings.Split(record.InVisibleUserName, "@-@")
	}

	if "" != record.Participants {
		detail.Participants = strings.Split(record.Participants, "@-@")
	}
	if "" != record.ParticipantNames {
		detail.ParticipantNames = strings.Split(record.ParticipantNames, "@-@")
	}
	detail.Url = record.Url
	detail.DisplayUrl = record.DisplayUrl
	detail.QrCode = record.QrCode
	detail.ParameterType = record.ParameterType
	detail.ServerId = record.ServerId
	detail.Vendor = record.Vendor
	detail.MtGroup = record.MtGroup
	detail.AccountGroup = record.AccountGroup
	detail.Leverage = record.Leverage
	detail.OwnerType = record.OwnerType
	detail.OwnerId = record.OwnerId
	detail.BusinessCode = record.BusinessCode
	detail.Os = record.Os
	detail.SoftwarePackage = record.SoftwarePackage
	return &detail
}

func generateIntroduceEntityNo(tenant_id string) string {
	entityNo := GetRandomStringEnhance(4, 4)
	counter := 0
	for counter < 5 {
		exist, errExist := ExistInIntroduceTab(tenant_id, constant.QueryKeyEntityNo, entityNo)
		if nil == errExist && false == exist {
			return entityNo
		}
		if true == exist {
			entityNo = GetRandomStringEnhance(4, 4)
		}
		counter++
	}
	return entityNo
}

//获取Introduce列表，typeValue和enable是过滤参数
func FindAllSystemIntroduceSimple(feignKey *XFeignKey, platform string, types []string, enable *bool) ([]SystemIntroduceDTO, code.Error) {
	if nil == feignKey || "" == feignKey.TenantId {
		return nil, errcode.CerrParamater
	}

	records, errList := ListIntroduceRecordsByKey(feignKey, nil, constant.QueryKeyAllRecords)
	if nil != errList {
		logrus.WithFields(logrus.Fields{}).Error("List All Introduce Records failed")
		return nil, errList
	}

	var detailList = make([]SystemIntroduceDTO, 0)
	for index := range records {
		if "" != platform && records[index].Platform != platform {
			continue
		}
		if len(types) > 0 && !StringListIsContain(types, records[index].Type) {
			continue
		}

		recordEnable := false
		if records[index].Enable != 0 {
			recordEnable = true
		}
		if nil != enable && *enable != recordEnable {
			continue
		}
		detail := SystemIntroduceTab2SystemIntroduceDTO(&records[index])
		detailList = append(detailList, *detail)
	}
	return detailList, nil
}

//获取Introduce列表，typeValue和enable是过滤参数,userId为推广用户的id
func FindAllSystemIntroduce(feignKey *XFeignKey, platform string, types []string, enable *bool, parameters map[string]string) ([]SystemIntroduceDTO, code.Error) {

	detailList, errList := FindAllSystemIntroduceSimple(feignKey, platform, types, enable)
	if nil != errList {
		logrus.WithFields(logrus.Fields{"error": errList}).Error("FindAllSystemIntroduceSimple failed")
		return nil, errList
	}
	if len(detailList) == 0 {
		return detailList, nil
	}

	productInfo, errGetProduct := ClientProductTenantByKey(feignKey, feignKey.TenantId, constant.ProductId)
	if nil != errGetProduct {
		logrus.WithFields(logrus.Fields{"tenant_id": feignKey.TenantId}).Error("Request client product tenant by key failed")
		return nil, errGetProduct
	}

	var pubUserIds = make([]string, 0)
	for index := range detailList {
		if "" != detailList[index].Creator {
			pubUserIds = append(pubUserIds, detailList[index].Creator)
		}
		if constant.IntroduceType_Web != detailList[index].Platform {
			continue
		}
		//生成代理自动填充目标URL和展示的URL
		_, introduceUrl, err := generateIntroduceDisplayUrl(productInfo, detailList[index].EntityNo, parameters)
		if nil != err {
			logrus.WithFields(logrus.Fields{"id": detailList[index].Id, "entityNo": detailList[index].EntityNo}).Warn("generateIntroduceDisplayUrl failed")
			continue
		}
		detailList[index].DisplayUrl = introduceUrl
	}

	if len(pubUserIds) < 0 {
		return detailList, nil
	}
	//根据pubUserIds批量获取名字，并更新到结果集中
	var mapUserName = make(map[string]string)
	users, errGetUser := ProcessListUserRecordsByKey(feignKey, pubUserIds, constant.QueryKeypubUserIds)
	if nil != errGetUser {
		logrus.WithFields(logrus.Fields{"error": errGetUser}).Warn("ProcessListUserRecordsByKey failed")
	} else if len(users) > 0 {
		for index := range users {
			mapUserName[users[index].PubUserId] = users[index].Name
		}

		for index := range detailList {
			creator := detailList[index].Creator
			if "" == creator {
				continue
			}
			name, ok := mapUserName[creator]
			if ok {
				detailList[index].Creator = name
			}
		}
	}

	return detailList, nil
}

//获取某截止时间前创建的introduce数,返回计数和错误
func GetIntroduceCountBeforeOneCreateTime(tenant_id string, deadTime time.Time) (int64, code.Error) {
	if "" == tenant_id {
		return 0, errcode.CerrParamater
	}
	dbConn, errConn := GetDBConnByTenantID(tenant_id)
	if errConn != nil {
		return 0, errConn
	}

	var total int64 = 0
	if errCount := dbConn.Model(&SystemIntroduceTab{}).Where("create_date < ? AND model_status != ?", deadTime, constant.ModelStatusDelete).Count(&total).Error; nil != errCount {
		logrus.WithFields(logrus.Fields{"error": errCount, "deadTime": deadTime}).Error("get introduce count failed")
		return 0, errcode.CerrExecuteSQL
	}

	return total, nil
}

//添加推广链接点击信息
func ProcessAddIntroduceHit(feignKey *XFeignKey, addReq *IntroduceHitDTO) code.Error {
	if nil == feignKey || "" == feignKey.TenantId || nil == addReq {
		return errcode.CerrParamater
	}

	var introduceHit SystemIntroduceHit
	introduceHit.Id = bson.NewObjectId()
	introduceHit.Url = addReq.Url
	introduceHit.IntroduceId = addReq.IntroduceId
	introduceHit.UserId = addReq.UserId
	introduceHit.TenantId = addReq.TenantId
	introduceHit.ClientIp = addReq.ClientIp
	introduceHit.Device = addReq.Device
	introduceHit.HitTime = time.Unix(addReq.HitTime/1000, addReq.HitTime%1000*1000000)

	mgoSess, errSess := GetMongoConnByTenantID(feignKey.TenantId)
	if nil != errSess {
		logrus.WithFields(logrus.Fields{"tenant_id": feignKey.TenantId, "error": errSess}).Error("GetMongoConnByTenantID failed")
		return errSess
	}
	defer mgoSess.Close()
	if errInsert := GetMgoHitColl(mgoSess).Insert(&introduceHit); nil != errInsert {
		logrus.WithFields(logrus.Fields{"error": errInsert}).Error("Insert into mongo failed")
		return errcode.CerrOperateMongo
	}

	return nil
}

//生成推广链接二维码并保存到阿里云,返回阿里云上的图片链接
func genQRCodeToAliyunOss(tenantId, fileName, displayUrl string) (string, code.Error) {
	//生成二维码图片
	code, errEncode := qr.Encode(displayUrl, qr.L, qr.Unicode)
	if nil != errEncode {
		logrus.WithFields(logrus.Fields{"error": errEncode, "displayUrl": displayUrl}).Error("qr.Encode failed")
		return "", errcode.CerrGenerateQrCode
	}
	codeScale, errScale := barcode.Scale(code, 300, 300)
	if nil != errScale {
		logrus.WithFields(logrus.Fields{"error": errEncode, "displayUrl": displayUrl}).Error("qr Scale failed")
		return "", errcode.CerrGenerateQrCode
	}
	pngBuffer := new(bytes.Buffer)
	errToBuffer := png.Encode(pngBuffer, codeScale)
	if nil != errToBuffer {
		logrus.WithFields(logrus.Fields{"error": errEncode, "displayUrl": displayUrl}).Error("png.Encode failed")
		return "", errcode.CerrGenerateQrCode
	}
	pngData := pngBuffer.Bytes()
	logrus.WithFields(logrus.Fields{"pngData len": len(pngData), "displayUrl": displayUrl, "codeContent": code.Content(), "codeScale": codeScale.Content()}).Debug("generate QR Code success, wait for send to cloud")

	//上传到阿里云
	uploadParams := map[string]string{
		"tenantId": tenantId,
		"name":     fileName,
		"dir":      "QRCode",
	}
	ossFilePath, err := SaveFileToAliyun(conf.GetApplication().PicServerDomain+"/v1/aliyun/oss/bw/upload", uploadParams, pngData)
	if nil != err {
		logrus.WithFields(logrus.Fields{"pngData len": len(pngData), "url": conf.GetApplication().PicServerDomain, "uploadParams": uploadParams,
			"dir": "QRCode"}).Error("save file to bw-pic error")

	}
	return ossFilePath, nil

	/*client, errOssNew := oss.New(conf.GetApplication().AliyunEndPoint, conf.GetApplication().AliyunAccessId, conf.GetApplication().AliyunAccessKey)
	if errOssNew != nil {
		logrus.WithFields(logrus.Fields{"error": errOssNew, "AliyunEndPoint": conf.GetApplication().AliyunEndPoint}).Error("aliyun oss new failed")
		return "", errcode.CerrAliyunAccess
	}
	bucket, errBucke := client.Bucket(conf.GetApplication().AliyunBucket)
	if errBucke != nil {
		logrus.WithFields(logrus.Fields{"error": errBucke, "AliyunBucket": conf.GetApplication().AliyunBucket}).Error("aliyun oss Bucke failed")
		return "", errcode.CerrAliyunAccess
	}
	errPut := bucket.PutObject(filePath, bytes.NewReader(pngData))
	if errPut != nil {
		logrus.WithFields(logrus.Fields{"error": errPut, "filePath": filePath}).Error("aliyun oss Bucke failed")
		return "", errcode.CerrAliyunAccess
	}

	if conf.GetApplication().PicCustomerDomainEnable != "" && conf.GetApplication().PicCustomerDomainEnable == "true" {
		url := conf.GetApplication().PicCustomerDomain + "/" + filePath
		if false == strings.HasPrefix(url, "//") {
			url = "//" + url
		}
		return url, nil
	} else {
		url := conf.GetApplication().AliyunBucket + "." + conf.GetApplication().AliyunEndPoint + "/" + filePath
		if false == strings.HasPrefix(url, "//") {
			url = "//" + url
		}
		return url, nil
	}*/
}

//生成BW推广链接二维码
func ProcessIntroducesQrcode(feignKey *XFeignKey, introduceId int64, isCurrentUserUrl bool) (string, code.Error) {
	if nil == feignKey || "" == feignKey.TenantId || introduceId < 0 {
		return "", errcode.CerrParamater
	}

	introduceInfo, errGet := GetIntroduceRecordsById(feignKey, introduceId)
	if nil != errGet {
		logrus.WithFields(logrus.Fields{"tenant_id": feignKey.TenantId, "error": errGet, "introduceId": introduceId}).Error("GetIntroduceRecordsById failed")
		return "", errGet
	} else if nil == introduceInfo {
		logrus.WithFields(logrus.Fields{"tenant_id": feignKey.TenantId, "introduceId": introduceId}).Error("GetIntroduceRecordsById not exist")
		return "", errcode.CerrIntroduceNotExist
	}

	productInfo, errGetProduct := ClientProductTenantByKey(feignKey, feignKey.TenantId, constant.ProductId)
	if nil != errGetProduct {
		logrus.WithFields(logrus.Fields{"tenant_id": feignKey.TenantId, "error": errGetProduct}).Error("Request client product tenant by key failed")
		return "", errGetProduct
	}

	parameters := make(map[string]string, 0)

	if introduceInfo.Type == constant.IntroduceType_DirectRecommendation {
		if isCurrentUserUrl {
			//bw不支持获取当前用户直客推广链接二维码
			return "", code.New(errcode.ParameterError, "Getting BW direct client qr code is not support")
		}
	} else {
		var userId int64 = 0
		if isCurrentUserUrl {
			userInfo, errUser := GetUserDetailByIdOrPubid(feignKey, feignKey.UserId, 0)
			if nil != errUser {
				logrus.WithFields(logrus.Fields{"tenant_id": feignKey.TenantId, "error": errUser, "pubUserId": feignKey.UserId}).Error("GetUserDetailByIdOrPubid failed")
				return "", errUser
			} else if nil == introduceInfo {
				logrus.WithFields(logrus.Fields{"tenant_id": feignKey.TenantId, "pubUserId": feignKey.UserId}).Error("GetUserDetailByIdOrPubid not exist")
				return "", errcode.CerrAccountNotFound
			}
			userId = userInfo.Id
		}

		if userId > 0 {
			parameters["uid"] = strconv.FormatInt(userId, 10)
		}
	}

	_, introduceUrl, errGenerate := generateIntroduceDisplayUrl(productInfo, introduceInfo.EntityNo, parameters)
	if nil != errGenerate {
		logrus.WithFields(logrus.Fields{"tenant_id": feignKey.TenantId, "error": errGenerate}).Error("generateIntroduceDisplayUrl failed")
		return "", errGenerate
	}

	if parameters["uid"] == "" {
		return genQRCodeToAliyunOss(feignKey.TenantId, introduceInfo.EntityNo+"_"+".png", introduceUrl)
	} else {
		return genQRCodeToAliyunOss(feignKey.TenantId, introduceInfo.EntityNo+"_"+parameters["uid"]+".png", introduceUrl)
	}

}

//生成TW推广链接二维码
func ProcessTwIntroducesQrcode(feignKey *XFeignKey, introduceId int64, customerId string, bwUserId string) (string, code.Error) {
	if nil == feignKey || "" == feignKey.TenantId || introduceId < 0 {
		return "", errcode.CerrParamater
	}

	introduceInfo, errGet := GetIntroduceRecordsById(feignKey, introduceId)
	if nil != errGet {
		logrus.WithFields(logrus.Fields{"tenant_id": feignKey.TenantId, "error": errGet, "introduceId": introduceId}).Error("GetIntroduceRecordsById failed")
		return "", errGet
	} else if nil == introduceInfo {
		logrus.WithFields(logrus.Fields{"tenant_id": feignKey.TenantId, "introduceId": introduceId}).Error("GetIntroduceRecordsById not exist")
		return "", errcode.CerrIntroduceNotExist
	}

	productInfo, errGetProduct := ClientProductTenantByKey(feignKey, feignKey.TenantId, constant.ProductId)
	if nil != errGetProduct {
		logrus.WithFields(logrus.Fields{"tenant_id": feignKey.TenantId, "error": errGetProduct}).Error("Request client product tenant by key failed")
		return "", errGetProduct
	}

	parameters := make(map[string]string, 0)

	if introduceInfo.Type == constant.IntroduceType_DirectRecommendation {
		if customerId == "" {
			return "", code.New(errcode.ParameterError, "customerId is empty")
		}
		parameters["cid"] = customerId
	} else {
		if bwUserId == "" {
			return "", code.New(errcode.ParameterError, "bwUserId is empty")
		}
		parameters["uid"] = bwUserId
	}

	_, introduceUrl, errGenerate := generateIntroduceDisplayUrl(productInfo, introduceInfo.EntityNo, parameters)
	if nil != errGenerate {
		logrus.WithFields(logrus.Fields{"tenant_id": feignKey.TenantId, "error": errGenerate}).Error("generateIntroduceDisplayUrl failed")
		return "", errGenerate
	}

	if parameters["uid"] == "" {
		return genQRCodeToAliyunOss(feignKey.TenantId, introduceInfo.EntityNo+"_"+parameters["cid"]+".png", introduceUrl)
	} else {
		return genQRCodeToAliyunOss(feignKey.TenantId, introduceInfo.EntityNo+"_"+parameters["uid"]+".png", introduceUrl)
	}
}

func getIntroduceHitInfoByIntroduceCode(tenantId, introduceCode string) ([]SystemIntroduceHit, code.Error) {
	if "" == tenantId || "" == introduceCode {
		return nil, errcode.CerrParamater
	}
	mgoSess, errSess := GetMongoConnByTenantID(tenantId)
	if nil != errSess {
		logrus.WithFields(logrus.Fields{"tenant_id": tenantId, "error": errSess}).Error("GetMongoConnByTenantID failed")
		return nil, errSess
	}
	defer mgoSess.Close()
	coll := GetMgoHitColl(mgoSess)

	var records []SystemIntroduceHit = make([]SystemIntroduceHit, 0)
	condition := bson.M{"tenantId": tenantId, "introduceId": introduceCode}
	errGet := coll.Find(condition).All(&records)
	if nil != errGet {
		logrus.WithFields(logrus.Fields{"tenant_id": tenantId, "error": errGet}).Error("Get introduce hit record failed")
		return nil, errcode.CerrOperateMongo
	}
	return records, nil
}

func introduceStaticGroupUserData(feignKey *XFeignKey, hitList []SystemIntroduceHit, applyList []AgencyRegisterStatsDTO) []IntroduceStatisticDTO {
	mapStatic := make(map[int64]IntroduceStatisticDTO, 0) //key为userId
	userIds := make([]int64, 0)
	for index, _ := range hitList {
		userId, errParse := strconv.ParseInt(hitList[index].UserId, 10, 64)
		if nil != errParse {
			continue
		}
		static, exist := mapStatic[userId]
		if false == exist {
			static.Id = userId
			static.NotPassNumber = 0
			static.PassNumber = 0
			static.ApplyNumber = 0
			static.HitNumber = 0
		}
		static.HitNumber++
		mapStatic[userId] = static
		userIds = append(userIds, userId)
	}

	for index, _ := range applyList {
		userId, errParse := strconv.ParseInt(applyList[index].Uid, 10, 64)
		if nil != errParse {
			continue
		}
		static, exist := mapStatic[userId]
		if false == exist {
			static.Id = userId
			static.NotPassNumber = 0
			static.PassNumber = 0
			static.ApplyNumber = 0
			static.HitNumber = 0
		}
		if constant.TaskState_Finished == applyList[index].TaskState {
			static.PassNumber++
		} else if constant.TaskState_Refused == applyList[index].TaskState {
			static.NotPassNumber++
		}
		static.ApplyNumber++
		mapStatic[userId] = static
		userIds = append(userIds, userId)
	}

	staticList := make([]IntroduceStatisticDTO, 0)
	if len(mapStatic) == 0 {
		return staticList
	}
	userRecords, errGetUser := ProcessListUserRecordsByKey(feignKey, userIds, constant.QueryKeyId)
	for userId, static := range mapStatic {
		//填充用户名称
		if nil == errGetUser {
			for index, _ := range userRecords {
				if userId == userRecords[index].Id {
					static.Name = userRecords[index].Name
					break
				}
			}
		}
		staticList = append(staticList, static)
	}
	return staticList
}

func hasDeposit(serverDepositMap map[string]AccountStatisticDTO, serverId, accountId string) bool {
	info, exist := serverDepositMap[serverId]
	if false == exist {
		return false
	}
	logrus.WithFields(logrus.Fields{"serverId": serverId, "accounts": info.Accounts, "accountId": accountId}).Debug("") /////////////////////////
	return StringListIsContain(info.Accounts, accountId)
}

func fillUserData(userStatisticMap map[string]IntroduceStatisticDTO, introduceType, userId string, count int64) {
	userIdInt, errParse := strconv.ParseInt(userId, 10, 64)
	if userIdInt < 0 || nil != errParse || count < 0 {
		return
	}

	var statistic IntroduceStatisticDTO
	var exist bool
	statistic, exist = userStatisticMap[userId]
	if false == exist {
		statistic.HitNumber = 0
		statistic.NewCustomerNumber = 0
		statistic.OpenAccountNumber = 0
		statistic.DepositeNumber = 0
	}
	switch introduceType {
	case constant.IntroduceStatistic_HIT:
		statistic.HitNumber += count
	case constant.IntroduceStatistic_CUSTOMER:
		statistic.NewCustomerNumber += count
	case constant.IntroduceStatistic_CUSTOMER_HAS_ACCOUNT:
		statistic.OpenAccountNumber += count
	case constant.IntroduceStatistic_CUSTOMER_HAS_DEPOSIT:
		statistic.DepositeNumber += count
	}
	userStatisticMap[userId] = statistic
}

func calcIntroduceLworkStatistic(feignKey *XFeignKey, result *IntroduceStatisticDTO, introduceInfo *SystemIntroduceTab) code.Error {
	hitList, errGetHit := getIntroduceHitInfoByIntroduceCode(feignKey.TenantId, introduceInfo.EntityNo)
	if nil != errGetHit {
		logrus.WithFields(logrus.Fields{"tenantId": feignKey.TenantId, "EntityNo": introduceInfo.EntityNo, "error": errGetHit}).Error("getIntroduceHitInfoByIntroduceCode error")
		return errGetHit
	}
	customerList, errGetCustom := ClientLinkCustom(feignKey, introduceInfo.EntityNo)
	if nil != errGetCustom {
		logrus.WithFields(logrus.Fields{"tenantId": feignKey.TenantId, "EntityNo": introduceInfo.EntityNo, "error": errGetCustom}).Error("ClientLinkCustom error")
		return errGetCustom
	}

	opportunityMap := make(map[string][]CustomerOpportunitiesDTO)
	if len(customerList) > 0 {
		options, errGetOption := ClientGetTenantOption(feignKey, "salesStage")
		if nil != errGetOption {
			logrus.WithFields(logrus.Fields{"tenantId": feignKey.TenantId, "error": errGetOption}).Error("ClientGetTenantOption error")
			return errGetOption
		}
		winFieldId := ""
		for index, _ := range options {
			if "赢单" == options[index].Label {
				winFieldId = options[index].Value
			}
		}
		if "" != winFieldId {
			customerIds := make([]string, 0)
			for index, _ := range customerList {
				customerIds = append(customerIds, customerList[index].CustomerId)
			}
			opportunityList, errGetOpportunity := ClientGetCustomerOpportunities(feignKey, customerIds, winFieldId)
			if nil != errGetOpportunity {
				logrus.WithFields(logrus.Fields{"tenantId": feignKey.TenantId, "error": errGetOption}).Error("ClientGetCustomerOpportunities error")
				return errGetOption
			}
			for index, _ := range opportunityList {
				customerId := opportunityList[index].CustomerId
				oppList, exist := opportunityMap[customerId]
				if false == exist {
					oppList = make([]CustomerOpportunitiesDTO, 0)
				}
				oppList = append(oppList, opportunityList[index])
				opportunityMap[customerId] = oppList
			}
		}
	}

	result.HitNumber = int64(len(hitList))
	result.NewCustomerNumber = int64(len(customerList))
	result.WinCustomerNumber = int64(len(opportunityMap))
	result.ContractAmount = 0
	return nil
}

func calcIntroduceCommonStatistic(feignKey *XFeignKey, result *IntroduceStatisticDTO, introduceInfo *SystemIntroduceTab) code.Error {
	hitList, errGetHit := getIntroduceHitInfoByIntroduceCode(feignKey.TenantId, introduceInfo.EntityNo)
	if nil != errGetHit {
		logrus.WithFields(logrus.Fields{"tenantId": feignKey.TenantId, "EntityNo": introduceInfo.EntityNo, "error": errGetHit}).Error("getIntroduceHitInfoByIntroduceCode error")
		return errGetHit
	}
	customerList, errGetCustom := ClientLinkCustom(feignKey, introduceInfo.EntityNo)
	if nil != errGetCustom {
		logrus.WithFields(logrus.Fields{"tenantId": feignKey.TenantId, "EntityNo": introduceInfo.EntityNo, "error": errGetCustom}).Error("ClientLinkCustom error")
		return errGetCustom
	}

	customerAccountMap := make(map[string][]Account)
	if len(customerList) > 0 {
		var searchDTO IdSearchDTO
		for index, _ := range customerList {
			searchDTO.Ids = append(searchDTO.Ids, customerList[index].CustomerId)
		}
		customerAccountMap, _ = GetCustomerAccountMap(feignKey, &searchDTO)
		//accountStacustomerAccountMaptic, errgetStatic := ClientGetStatisticHasDeposit(feignKey)
		//if nil != errgetStatic {
		//	logrus.WithFields(logrus.Fields{"tenantId": feignKey.TenantId, "error":errgetStatic}).Error("ClientGetStatisticHasDeposit error")
		//	return errgetStatic
		//}
		//serverDepositMap := make(map[string]AccountStatisticDTO)
		//for index, _ := range accountStatic {
		//	serverDepositMap[accountStatic[index].ServerId] = accountStatic[index]
		//}
		//
		//var searchDTO IdSearchDTO
		//for index, _ := range customerList {
		//	searchDTO.Ids = append(searchDTO.Ids, customerList[index].CustomerId)
		//}
		//accountOwners, errGetAccount := ClientFindAccountBaseInfoByCustomer(feignKey, &searchDTO)
		//if nil != errGetAccount {
		//	logrus.WithFields(logrus.Fields{"tenantId": feignKey.TenantId, "error":errGetAccount,"ids":searchDTO.Ids}).Error("ClientFindAccountBaseInfoByCustomer error")
		//	return errGetAccount
		//}
		//
		//for index, _ := range accountOwners {
		//	for _, relation := range accountOwners[index].Accounts {
		//		var act Account
		//		act.ServerId = relation.ServerId
		//		act.AccountId = relation.Account
		//		act.CustomerId = accountOwners[index].CustomerId
		//		act.HasDeposit = hasDeposit(serverDepositMap, relation.ServerId, relation.Account)
		//
		//		actList, exist := customerAccountMap[act.CustomerId]
		//		if false == exist {
		//			actList = make([]Account, 0)
		//		}
		//		actList = append(actList, act)
		//		customerAccountMap[act.CustomerId] = actList
		//	}
		//}
	}

	logrus.WithFields(logrus.Fields{"--->>> customerAccountMap": customerAccountMap}).Info("CustomerAccountMap")

	result.HitNumber = int64(len(hitList))
	result.NewCustomerNumber = int64(len(customerList))
	result.PassNumber = 0
	result.NotPassNumber = 0
	result.DepositeNumber = 0
	result.OpenAccountNumber = 0

	userStatisticMap := make(map[string]IntroduceStatisticDTO)
	if introduceInfo.BwUserShow != constant.IntroduceType_UserNotVisible {
		for index, _ := range hitList {
			fillUserData(userStatisticMap, constant.IntroduceStatistic_HIT, hitList[index].UserId, 1)
		}
		customerIdMap := make(map[string]CustomerPropertiesDTO)
		for index, _ := range customerList {
			fillUserData(userStatisticMap, constant.IntroduceStatistic_CUSTOMER, customerList[index].Uid, 1)
			customerIdMap[customerList[index].CustomerId] = customerList[index]
		}
		for customerId, accountList := range customerAccountMap {
			customer, existCus := customerIdMap[customerId]
			logrus.WithFields(logrus.Fields{"--->>> customer": customer}).Info("customer")

			if existCus {
				fillUserData(userStatisticMap, constant.IntroduceStatistic_CUSTOMER_HAS_ACCOUNT, customer.Uid, 1)
				result.OpenAccountNumber += 1
			}
			result.PassNumber += 1
			haveDeposit := false
			for index, _ := range accountList {
				if accountList[index].HasDeposit {
					fillUserData(userStatisticMap, constant.IntroduceStatistic_CUSTOMER_HAS_DEPOSIT, customer.Uid, 1)
					haveDeposit = true
				}
			}
			if haveDeposit {
				result.DepositeNumber += 1
			}
		}
	}
	//logrus.WithFields(logrus.Fields{"1.---> userStatisticMap": userStatisticMap}).Info("userStatisticMap")
	if len(userStatisticMap) > 0 {
		var userIdList []int64 = make([]int64, 0)
		for userIdStr, _ := range userStatisticMap {
			userIdInt, _ := strconv.ParseInt(userIdStr, 10, 64)
			userIdList = append(userIdList, userIdInt)
		}
		userList, errGetUser := ProcessListUserRecordsByKey(feignKey, userIdList, constant.QueryKeyId)
		//logrus.WithFields(logrus.Fields{"---> userList": userList}).Info("userList")
		if nil == errGetUser {
			for index, _ := range userList {
				userIdStr := strconv.FormatInt(userList[index].Id, 10)
				static, exist := userStatisticMap[userIdStr]
				if exist {
					static.Id = userList[index].Id
					static.Name = userList[index].Name
					userStatisticMap[userIdStr] = static
				}
			}
		} else {
			logrus.WithFields(logrus.Fields{"tenantId": feignKey.TenantId, "error": errGetUser, "ids": userIdList}).Warn("ProcessListUserRecordsByKey error")
			//不返回错误
		}
		//logrus.WithFields(logrus.Fields{"2.---> userStatisticMap": userStatisticMap}).Info("userStatisticMap")

		result.UserStatistic = make([]IntroduceStatisticDTO, 0)
		for _, static := range userStatisticMap {
			result.UserStatistic = append(result.UserStatistic, static)
		}
	}
	return nil
}

func ProcessIntroducesDetail(feignKey *XFeignKey, introduceId int64, statistic bool) (*IntroduceStatisticDTO, code.Error) {
	if nil == feignKey || "" == feignKey.TenantId || introduceId <= 0 {
		return nil, errcode.CerrParamater
	}

	introduceInfo, errGetIntroduce := GetIntroduceRecordsById(feignKey, introduceId)
	if nil != errGetIntroduce {
		logrus.WithFields(logrus.Fields{"error": errGetIntroduce, "introduceId": introduceId}).Error("GetIntroduceRecordsById failed")
		return nil, errGetIntroduce
	} else if nil == introduceInfo {
		logrus.WithFields(logrus.Fields{"introduceId": introduceId}).Error("GetIntroduceRecordsById not exist")
		return nil, errcode.CerrIntroduceNotExist
	}

	var result IntroduceStatisticDTO
	result.copyFromSystemIntroduceTab(introduceInfo)

	if statistic && constant.IntroduceType_DirectRecommendation != result.Type { //统计计算
		if conf.GetApplication().LworkOfficialTenantId == feignKey.TenantId {
			errCalc := calcIntroduceLworkStatistic(feignKey, &result, introduceInfo)
			if nil != errCalc {
				logrus.WithFields(logrus.Fields{"tenantId": feignKey.TenantId, "error": errCalc}).Error("calcIntroduceLworkStatistic error")
				return nil, errCalc
			}
		} else if constant.IntroduceType_Agent == introduceInfo.Type {
			hitList, errGetHit := getIntroduceHitInfoByIntroduceCode(feignKey.TenantId, introduceInfo.EntityNo)
			if nil != errGetHit {
				logrus.WithFields(logrus.Fields{"tenantId": feignKey.TenantId, "EntityNo": introduceInfo.EntityNo, "error": errGetHit}).Error("getIntroduceHitInfoByIntroduceCode error")
				return nil, errGetHit
			}
			applyList, errGetAgency := ClientTaskAgencyList(feignKey, introduceInfo.EntityNo)
			if nil != errGetAgency {
				logrus.WithFields(logrus.Fields{"tenantId": feignKey.TenantId, "EntityNo": introduceInfo.EntityNo}).Error("ClientTaskAgencyList error")
				return nil, errGetAgency
			}
			//logrus.WithFields(logrus.Fields{"--->>> applyList": applyList}).Info("applyList")

			result.HitNumber = int64(len(hitList))
			result.ApplyNumber = int64(len(applyList))
			result.PassNumber = 0
			result.NotPassNumber = 0
			for index, _ := range applyList {
				//logrus.WithFields(logrus.Fields{"--->>> TaskState": applyList[index].TaskState}).Info("TaskState")
				if constant.TaskState_Finished == applyList[index].TaskState {
					result.PassNumber += 1
				} else if constant.TaskState_Refused == applyList[index].TaskState {
					result.NotPassNumber += 1
				}
			}
			//logrus.WithFields(logrus.Fields{"--->>> 1.result": result}).Info("result")

			result.UserStatistic = introduceStaticGroupUserData(feignKey, hitList, applyList)
		} else {
			errCalc := calcIntroduceCommonStatistic(feignKey, &result, introduceInfo)
			if nil != errCalc {
				logrus.WithFields(logrus.Fields{"tenantId": feignKey.TenantId, "error": errCalc}).Error("calcIntroduceCommonStatistic error")
				return nil, errCalc
			}
		}
	}
	//logrus.WithFields(logrus.Fields{"--->>> 2.result": result}).Info("result")
	//填充归属信息
	if constant.OwnerType_RoleId == result.OwnerType {
		roleId, errParse := strconv.ParseInt(result.OwnerId, 10, 64)
		if nil == errParse {
			roleInfo, errGetRole := GetRoleInfoByID(feignKey.TenantId, roleId)
			if nil == errGetRole && nil != roleInfo {
				result.OwnerName = roleInfo.Name
			} else {
				logrus.WithFields(logrus.Fields{"roleId": roleId, "errGetRole": errGetRole, "roleInfo": roleInfo}).Warn("get role info abnormal")
			}
		} else {
			logrus.WithFields(logrus.Fields{"OwnerId": result.OwnerId, "OwnerType": result.OwnerType, "introduceId": introduceId}).Warn("introduce info ownerId abnormal")
		}
	} else if constant.OwnerType_Id == result.OwnerType && "" != result.OwnerId {
		userInfo, errGetUser := GetUserRecordByIdOrPubid(feignKey, result.OwnerId, 0)
		if nil == errGetUser && nil != userInfo {
			result.OwnerName = userInfo.Name
		} else {
			logrus.WithFields(logrus.Fields{"OwnerId": result.OwnerId, "errGetUser": errGetUser, "userInfo": userInfo}).Warn("GetUserRecordByIdOrPubid abnormal")
		}
	}

	logrus.WithFields(logrus.Fields{"--->>> 3.result": result}).Info("result")

	return &result, nil
}

//切换推广链接启用状态
func ProcessSwitchIntroduceState(feignKey *XFeignKey, introduceId int64) code.Error {
	if nil == feignKey || "" == feignKey.TenantId || introduceId <= 0 {
		return errcode.CerrParamater
	}

	introduceInfo, errGetIntroduce := GetIntroduceRecordsById(feignKey, introduceId)
	if nil != errGetIntroduce {
		logrus.WithFields(logrus.Fields{"error": errGetIntroduce, "introduceId": introduceId}).Error("GetIntroduceRecordsById failed")
		return errGetIntroduce
	} else if nil == introduceInfo {
		logrus.WithFields(logrus.Fields{"introduceId": introduceId}).Error("GetIntroduceRecordsById not exist")
		return errcode.CerrIntroduceNotExist
	}

	toStatus := 1
	if introduceInfo.Enable > 0 {
		toStatus = 0
	} else {
		toStatus = 1
	}

	dbConn, errConn := GetDBConnByTenantID(feignKey.TenantId)
	if errConn != nil {
		return errConn
	}
	errUpdate := dbConn.Model(&introduceInfo).Where("id = ?", introduceId).Update("enable", toStatus).Error
	if nil != errUpdate {
		logrus.WithFields(logrus.Fields{"introduceId": introduceId, "errUpdate": errUpdate}).Error("update introduce enable status failed")
		return errcode.CerrIntroduceNotExist
	}
	logrus.WithFields(logrus.Fields{"introduceId": introduceId, "errUpdate": errUpdate, "toStatus": toStatus}).Info("update introduce enable status success")

	return nil
}

//批量删除推广链接
func ProcessDeleteIntroduce(feignKey *XFeignKey, idsList []int64) code.Error {
	tenant_id := feignKey.TenantId
	if "" == tenant_id || len(idsList) <= 0 {
		return errcode.CerrParamater
	}

	//在t_system_introduce表中删除用户信息
	//暂时采用硬删除DB记录，后面通过修改状态实现////////////////
	dbConn, errConn := GetDBConnByTenantID(tenant_id)
	if errConn != nil {
		return errConn
	}
	if errDelete := dbConn.Where("id in (?)", idsList).Delete(SystemIntroduceTab{}).Error; nil != errDelete {
		logrus.WithFields(logrus.Fields{"ids": idsList, "error": errDelete}).Errorf("Delete introduce record failed")
		return errcode.CerrExecuteSQL
	}

	logrus.WithFields(logrus.Fields{"ids": idsList, "operate User": feignKey.UserId}).Infof("Remove introduce")
	return nil
}

//更新推广链接
func ProcessUpdateSystemIntroduce(feignKey *XFeignKey, updateReq *SystemIntroduceDTO) code.Error {
	if nil == feignKey || "" == feignKey.TenantId || nil == updateReq || updateReq.Id < 0 {
		return errcode.CerrParamater
	}

	introduceInfo, errGetIntroduce := GetIntroduceRecordsById(feignKey, updateReq.Id)
	if nil != errGetIntroduce {
		logrus.WithFields(logrus.Fields{"id": updateReq.Id, "error": errGetIntroduce}).Errorf("GetIntroduceRecordsById failed")
		return errGetIntroduce
	} else if nil == introduceInfo {
		logrus.WithFields(logrus.Fields{"id": updateReq.Id}).Errorf("System introduce not exist")
		return errcode.CerrIntroduceNotExist
	}

	errConv := introduceInfo.updateFromSystemIntroduceDTO(updateReq)
	if nil != errConv {
		logrus.WithFields(logrus.Fields{"error": errConv}).Errorf("updateFromSystemIntroduceDTO failed")
		return errConv
	}
	if introduceInfo.CreateDate.Unix() < constant.MIN_TIMESTAMP || introduceInfo.CreateDate.Unix() > constant.MAX_TIMESTAMP {
		introduceInfo.CreateDate = time.Unix(constant.MIN_TIMESTAMP, 0)
	}
	introduceInfo.ModifyDate = time.Now()
	introduceInfo.ModifyUserId = feignKey.UserId

	//在t_system_introduce表中删除用户信息
	//暂时采用硬删除DB记录，后面通过修改状态实现////////////////
	dbConn, errConn := GetDBConnByTenantID(feignKey.TenantId)
	if errConn != nil {
		return errConn
	}
	if errSave := dbConn.Save(introduceInfo).Error; nil != errSave {
		logrus.WithFields(logrus.Fields{"id": updateReq.Id, "error": errSave}).Errorf("update introduce record failed")
		return errcode.CerrExecuteSQL
	}

	logrus.WithFields(logrus.Fields{"id": updateReq.Id, "operate User": feignKey.UserId}).Infof("update introduce success")
	return nil
}
