package model

import (
	"bw/bw-user/constant"
	"bw/bw-user/errcode"
	"encoding/json"
	"github.com/Sirupsen/logrus"
	"github.com/lworkltd/kits/service/invoke"
	"github.com/lworkltd/kits/service/restful/code"
	invokeutils "github.com/lworkltd/kits/utils/invoke"
	"strconv"
	"strings"
)

type AccountStatisticDTO struct {
	ServerId string   `json:"serverId"`
	Accounts []string `json:"accounts"` //set
}

type AccountAndOwnerRelation struct {
	Account  string `json:"account"`
	Vendor   string `json:"vendor"`
	ServerId string `json:"serverId"`
	BindTime int64  `json:"bindTime"`
	IsActive bool   `json:"isActive"`
}

type AbstractAccountOwner struct {
	//Serializable        未写
	Id         string                    `json:"id"`
	TenantId   string                    `json:"tenantId"`
	CustomerId string                    `json:"customerId"`
	Accounts   []AccountAndOwnerRelation `json:"accounts"`
}

type AccountOwnerBaseDTO struct {
	AbstractAccountOwner
	AccountName          string     `json:"accountName"`
	Gender               string     `json:"gender"`
	Birthday             string     `json:"birthday"`
	Phones               PhoneInfo  `json:"phones"`
	StandbyTelephone     PhoneInfo  `json:"standbyTelephone"`
	Email                string     `json:"email"`
	Nationality          string     `json:"nationality"`
	Residence            RegionInfo `json:"residence"`
	HomePlace            RegionInfo `json:"homePlace"`
	Postcode             string     `json:"postcode"`
	Address              string     `json:"address"`
	Company              string     `json:"company"`
	WorkAge              string     `json:"workAge"`
	Remark               string     `json:"remark"`
	IdAddress            string     `json:"idAddress"`
	Im                   string     `json:"im"`
	Introducer           string     `json:"introducer"`
	FamilyName           string     `json:"familyName"`
	Appellation          string     `json:"appellation"`
	AccountType          string     `json:"accountType"`
	FirstName            string     `json:"firstName"`
	LastName             string     `json:"lastName"`
	Title                string     `json:"title"`
	Isusa                string     `json:"isusa"`
	EmploymentStatus     string     `json:"employmentStatus"`
	Employer             string     `json:"employer"`
	JobTitle             string     `json:"jobTitle"`
	AverageTaxableIncome string     `json:"averageTaxableIncome"`
	TotalMoney           string     `json:"totalMoney"`
	StreetAddress        string     `json:"streetAddress"`
	UnitNum              string     `json:"unitNum"`
	AreaTownState        string     `json:"areaTownState"`
	HomeNumber           string     `json:"homeNumber"`
}

func ClientCheckAccountExist(feignKey *XFeignKey, login int64, serverId string, vendor string) (bool, code.Error) {
	feignkeyBypes, _ := json.Marshal(feignKey)
	rsp, err := invoke.Name(constant.ServiceNameAccountFeign).
		Get(constant.PathCheckAccountExist).
		Route("login", strconv.FormatInt(login, 10)).
		Header("X-Feign-Key", string(feignkeyBypes)).
		Header("x-api-serverid", serverId).
		Header("x-api-vendor", vendor).
		Response()

	var response bool
	cerr := invokeutils.ExtractHttpResponse(constant.ServiceNameAccountFeign, err, rsp, &response)
	if cerr != nil {
		logrus.WithFields(logrus.Fields{
			"error": cerr,
		}).Error("Request Find Tenant By Id failed")
		return false, errcode.CerrRequestAccountFeign
	}

	return response, nil
}

//根据客户查询账户所有人基本信息
func ClientFindAccountBaseInfoByCustomer(feignKey *XFeignKey, searchReq *IdSearchDTO) ([]AccountOwnerBaseDTO, code.Error) {
	feignkeyBypes, _ := json.Marshal(feignKey)
	rsp, err := invoke.Name(constant.ServiceNameAccountFeign).
		Post(constant.PathAccountBaseInfo).
		Header("X-Feign-Key", string(feignkeyBypes)).
		Json(searchReq).
		Response()

	var response []AccountOwnerBaseDTO
	cerr := invokeutils.ExtractHttpResponse(constant.ServiceNameAccountFeign, err, rsp, &response)
	if cerr != nil {
		logrus.WithFields(logrus.Fields{
			"error": cerr,
		}).Error("ClientFindAccountBaseInfoByCustomer failed")
		return nil, errcode.CerrRequestAccountFeign
	}

	return response, nil
}

//获取租户有入金记录的账户统计信息
func ClientGetStatisticHasDeposit(feignKey *XFeignKey) ([]AccountStatisticDTO, code.Error) {
	feignkeyBypes, _ := json.Marshal(feignKey)
	rsp, err := invoke.Name(constant.ServiceNameAccountFeign).
		Get(constant.PathStatisticDeposit).
		Header("X-Feign-Key", string(feignkeyBypes)).
		Response()

	var response []AccountStatisticDTO
	cerr := invokeutils.ExtractHttpResponse(constant.ServiceNameAccountFeign, err, rsp, &response)
	if cerr != nil {
		logrus.WithFields(logrus.Fields{
			"error": cerr,
		}).Error("ClientGetStatisticHasDeposit failed")
		return nil, errcode.CerrRequestAccountFeign
	}

	return response, nil
}

func GetCustomerAccountMap(feignKey *XFeignKey, searchReq *IdSearchDTO) (map[string][]Account, code.Error) {
	feignkeyBypes, _ := json.Marshal(feignKey)
	rsp, err := invoke.Name(constant.ServiceNameAccountFeign).
		Get(constant.PathGetCustomerAccountMap).
		Header("X-Feign-Key", string(feignkeyBypes)).
		Json(searchReq).
		Response()

	var response map[string][]Account
	cerr := invokeutils.ExtractHttpResponse(constant.ServiceNameAccountFeign, err, rsp, &response)
	if cerr != nil {
		logrus.WithFields(logrus.Fields{
			"error": cerr,
		}).Error("GetCustomerAccountMap failed")
		return nil, errcode.CerrRequestAccountFeign
	}

	return response, nil
}

//检查bw用户是否被归属
func ClientCheckUserAsOwner(userIds []string, feignKey *XFeignKey) (bool, code.Error) {
	feignkeyBypes, _ := json.Marshal(feignKey)
	rsp, err := invoke.Name(constant.ServiceNameAccountFeign).
		Get(constant.PathCheckUserAsOwner).
		Header("X-Feign-Key", string(feignkeyBypes)).
		Query("userIds", strings.Join(userIds, ",")).
		Response()

	var response bool
	cerr := invokeutils.ExtractHttpResponse(constant.ServiceNameAccountFeign, err, rsp, &response)
	if cerr != nil {
		logrus.WithFields(logrus.Fields{
			"error": cerr,
		}).Error("ClientCheckUserAsOwner failed")
		return false, errcode.CerrRequestAccountFeign
	}

	return response, nil
}
