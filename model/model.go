package model

import (
	"bw/bw-user/conf"
	"bw/bw-user/constant"
	"bw/bw-user/errcode"
	"fmt"
	"github.com/Sirupsen/logrus"
	//"github.com/lworkltd/kits/service/profile"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/lworkltd/kits/service/restful/code"
	"gopkg.in/mgo.v2"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"
)

type XFeignKey struct {
	ClientIp    string `json:"clientIp,omitempty"`
	CompanyName string `json:"companyName,omitempty"`
	Device      string `json:"device,omitempty"`
	Email       string `json:"email,omitempty"`
	Language    string `json:"language,omitempty"`
	Phone       string `json:"phone,omitempty"`
	ProductId   string `json:"productId,omitempty"`
	TenantId    string `json:"tenantId,omitempty"`
	UserId      string `json:"userId,omitempty"` //public user id
}

type tenantDBConnInfo struct {
	conn     *gorm.DB
	connTime int64 //重新建立链接的的时间戳
}

//User Tenant map, key is tenant_id
var dbConnMap map[string]tenantDBConnInfo = make(map[string]tenantDBConnInfo)
var dbConnMapMutex sync.RWMutex

var mgoSession *mgo.Session = nil
var dbMgoConnMutex sync.RWMutex

/*
func GetCommResponse(result bool, code int, data interface{}) *CommResonse {
	var response CommResonse
	response.Result = result
	response.Mcode = fmt.Sprintf("m%07d", code)
	response.Data = data
	return &response
}
*/

func getTenantMysqlConnStr(tenant_id string) (string, code.Error) {
	feignKey := XFeignKey{TenantId: tenant_id, ProductId: constant.ProductId}
	deploy, errGetDeploy := ClientGetProductMysql(&feignKey)
	if nil != errGetDeploy {
		logrus.WithFields(logrus.Fields{"error": errGetDeploy,}).Error("Failed to ClientGetProductMysql")
		return "", errGetDeploy
	} else if nil == deploy {
		return "", errcode.CerrInternal
	}

	if conf.GetMode() == "local" {
		deploy.Host = conf.GetMysql().Host
	}

	//"root:password@tcp(10.25.100.164:3306)/brokerwork_t001117?charset=utf8&parseTime=True&loc=Local"
	connStr := fmt.Sprintf("%v:%v@tcp(%v)/%v?charset=utf8&parseTime=True&loc=Local", deploy.User, deploy.Password, deploy.Host, deploy.Scheme)
	logrus.WithFields(logrus.Fields{"connStr": connStr, "tenant_id": tenant_id}).Info("generate mysql conn str")
	return connStr, nil
}

func InitMysqlConn(tenant_id string) (*gorm.DB, code.Error) {
	//db, errConn := gorm.Open("mysql", "user:password@/dbname?charset=utf8&parseTime=True&loc=Local")

	connUrl, errGetConnStr := getTenantMysqlConnStr(tenant_id)
	if nil != errGetConnStr {
		logrus.WithFields(logrus.Fields{"error": errGetConnStr,}).Error("Failed to getTenantMysqlConnStr")
		return nil, errGetConnStr
	}
	if conf.GetMode() == "local" {
		connUrl = conf.GetMysql().Url
	}
	db, errConn := gorm.Open("mysql", connUrl)
	if errConn != nil {
		logrus.WithFields(logrus.Fields{
			"error": errConn,
		}).Error("Failed to connect mysql server")
		return nil, errcode.CerrConnMySQL
	}

	db.DB().SetMaxIdleConns(0)
	db.DB().SetMaxOpenConns(4)
	db.DB().SetConnMaxLifetime(60 * time.Second)

	return db, nil
}

//This function may be calling by multiple coroutine at the same time
//Use tenant_id to get the DB conn
func GetDBConnByTenantID(tenant_id string) (*gorm.DB, code.Error) {
	if "" == tenant_id {
		return nil, errcode.CerrParamater
	}

	timeNow := time.Now().Unix()

	dbConnMapMutex.RLock()
	if connInfo, ok := dbConnMap[tenant_id]; true == ok {
		dbConnMapMutex.RUnlock()
		return connInfo.conn, nil
	}
	dbConnMapMutex.RUnlock()

	newConn, err := InitMysqlConn(tenant_id)
	if err != nil {
		logrus.WithFields(logrus.Fields{"tenant_id": tenant_id,}).Error("Failed to connect mysql server")
		return nil, err
	}

	dbConnMapMutex.Lock()
	//if oldInfo, ok := dbConnMap[tenant_id]; true == ok {
	//	defer oldInfo.conn.Close()
	//}
	newInfo := tenantDBConnInfo{conn: newConn, connTime: timeNow}
	dbConnMap[tenant_id] = newInfo
	dbConnMapMutex.Unlock()
	return newConn, nil
}

//获取mongoDB的连接，获取到后需defer sess.Close()
func GetMongoConnByTenantID(tenant_id string) (*mgo.Session, code.Error) {
	if "" == tenant_id {
		return nil, errcode.CerrParamater
	}

	dbMgoConnMutex.RLock()
	if nil != mgoSession {
		sess := mgoSession.Clone()
		dbMgoConnMutex.RUnlock()
		return sess, nil
	}
	dbMgoConnMutex.RUnlock()

	session, err := mgo.Dial(conf.GetMongo().Url)
	if nil != err || nil == session {
		logrus.WithFields(logrus.Fields{"tenant_id": tenant_id, "error": err}).Error("Failed to connect mongo server")
		return nil, errcode.CerrConnMongo
	}
	dbMgoConnMutex.Lock()
	mgoSession = session
	sess := mgoSession.Clone()
	dbMgoConnMutex.Unlock()
	return sess, nil
}

func GetMgoHitColl(sess *mgo.Session) *mgo.Collection {
	dbName := conf.GetApplication().IntroduceHitDBName
	collName := "t_introduce_hit"
	if "" == dbName {
		dbName = "msc-db"
	}
	return sess.DB(dbName).C(collName)
}

func GetMgoUserFieldColl(sess *mgo.Session) *mgo.Collection {
	dbName := conf.GetApplication().UserFieldsDBName
	collName := "t_user_fields"
	if "" == dbName {
		dbName = "msc-db"
	}
	return sess.DB(dbName).C(collName)
}

func GetMgoSearchColl(sess *mgo.Session, isNew bool) *mgo.Collection {
	dbName := conf.GetApplication().AdvanceSearchDBName
	collName := "t_advance_search"
	if isNew == true {
		collName = "t_advance_search_new"
	}
	if "" == dbName {
		dbName = "msc-db"
	}
	return sess.DB(dbName).C(collName)
}

func GetMgoCheckColl(sess *mgo.Session) *mgo.Collection {
	dbName := conf.GetApplication().UserFieldsDBName
	collName := "tenant_info"
	return sess.DB(dbName).C(collName)
}

//srcValue：value1,value2,value3
func SplitToIntList(srcValue string, sep string) ([]int64, code.Error) {
	var intValueList []int64 = make([]int64, 0)
	if "" == srcValue {
		return intValueList, nil
	}

	valueList := strings.Split(srcValue, sep)
	for _, value := range valueList {
		if "" == value {
			continue
		}
		intValue, err := strconv.ParseInt(value, 10, 64)
		if nil != err {
			logrus.WithFields(logrus.Fields{"srcValue": srcValue}).Error("SplitToIntList, srcValue illegal")
			return nil, errcode.CerrParamater
		}
		intValueList = append(intValueList, intValue)
	}
	return intValueList, nil
}

//srcValue：value1,value2,value3
func SplitToStringList(srcValue string, sep string) ([]string, code.Error) {
	var strValueList []string = make([]string, 0)
	if "" == srcValue {
		return strValueList, nil
	}

	valueList := strings.Split(srcValue, sep)
	for _, value := range valueList {
		if "" != value {
			strValueList = append(strValueList, value)
		}
	}
	return strValueList, nil
}

func IntListIsCotain(list []int, element int) bool {
	for _, value := range list {
		if value == element {
			return true
		}
	}
	return false
}

func Int64ListIsCotain(list []int64, element int64) bool {
	for _, value := range list {
		if value == element {
			return true
		}
	}
	return false
}

func StringListIsContain(list []string, element string) bool {
	for index := range list {
		if element == list[index] {
			return true
		}
	}
	return false
}

func StringListIsContainIgnoreCase(list []string, element string) bool {
	for index := range list {
		src := strings.ToLower(list[index])
		dest := strings.ToLower(element)
		if src == dest {
			return true
		}
	}
	return false
}

func StringListContainContain(list []string, element string) bool {
	for index := range list {
		src := strings.ToLower(list[index])
		dest := strings.ToLower(element)
		if strings.Contains(dest, src) {
			return true
		}
	}
	return false
}

//生成随机字符串
func GetRandomStringEnhance(charCount uint, numCount uint) string {
	strChar := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	strNum := "0123456789"
	bytesChar := []byte(strChar)
	bytesNum := []byte(strNum)
	result := []byte{}
	genCharCount := uint(0)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := uint(0); i < charCount+numCount; i++ {
		isNum := (genCharCount >= charCount) || (r.Intn(2)%2 == 0)
		if !isNum {
			result = append(result, bytesChar[r.Intn(len(bytesChar))])
			genCharCount++
		} else {
			result = append(result, bytesNum[r.Intn(len(bytesNum))])
		}
	}
	return string(result)
}

func GenerateNewUUID() string {
	str := "0123456789abcdef"
	bytesStr := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < 8; i++ {
		result = append(result, bytesStr[r.Intn(len(bytesStr))])
	}
	result = append(result, []byte("-")...)
	for i := 0; i < 4; i++ {
		result = append(result, bytesStr[r.Intn(len(bytesStr))])
	}
	result = append(result, []byte("-")...)
	for i := 0; i < 4; i++ {
		result = append(result, bytesStr[r.Intn(len(bytesStr))])
	}
	result = append(result, []byte("-")...)
	for i := 0; i < 4; i++ {
		result = append(result, bytesStr[r.Intn(len(bytesStr))])
	}
	result = append(result, []byte("-")...)
	for i := 0; i < 12; i++ {
		result = append(result, bytesStr[r.Intn(len(bytesStr))])
	}
	return string(result)
}

//判断是否为空串或者数字
func isEmptyOrNum(strValue *string) (bool) {
	if nil == strValue {
		return true
	}
	if "" == *strValue {
		return true
	}
	_, err := strconv.ParseInt(*strValue, 10, 64)
	if nil != err {
		return false
	}
	return true
}

func IfEmptyChoseOther(source *string, new string) string {
	if nil == source {
		return new
	}
	return *source
}

func Int64ListJoinToString(list []int64, spec string) string {
	if len(list) == 0 {
		return ""
	}
	result := ""
	for index, value := range list {
		result += strconv.FormatInt(value, 10)
		if index < len(list)-1 {
			result += spec
		}
	}
	return result
}
