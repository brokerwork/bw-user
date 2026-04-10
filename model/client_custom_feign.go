package model

import (
	"bw/bw-user/constant"
	"bw/bw-user/errcode"
	"encoding/json"
	"github.com/Sirupsen/logrus"
	"github.com/lworkltd/kits/service/invoke"
	"github.com/lworkltd/kits/service/restful/code"
	invokeutils "github.com/lworkltd/kits/utils/invoke"
)

type CustomerPropertiesDTO struct {
	CustomerId       string     `json:"customerId"`
	TenantId         string     `json:"tenantId"`
	CustomName       string     `json:"customName"`      //客户名称
	OweId            string     `json:"oweId"`           //主归属人
	BindUserId       string     `json:"bindUserId"`      //绑定BW用户ID
	BindUserName     string     `json:"bindUserName"`    //绑定BW用户名
	CommendId        string     `json:"commendId"`       //推荐客户ID
	CommendName      string     `json:"commendName"`     //推荐客户名
	OweName          string     `json:"oweName"`         //主归属人名称
	Participant      []string   `json:"participant"`     //参与人
	ParticipantName  []string   `json:"participantName"` //参与人名称
	CustomNo         string     `json:"customNo"`        //客户编号
	CustomerType     string     `json:"customerType"`
	Ambitious        string     `json:"ambitious"`    //客户规模: 指客户人数
	CustomSource     string     `json:"customSource"` //客户来源
	SourceType       string     `json:"sourceType"`   //
	Introducer       string     `json:"introducer"`
	CustomerLevel    string     `json:"customerLevel"`
	Country          RegionInfo `json:"country"`
	Address          string     `json:"address"`
	Postcode         string     `json:"postcode"`
	Site             string     `json:"site"` //网址
	Social           string     `json:"social"`
	Im               string     `json:"im"`
	Phones           PhoneInfo  `json:"phones"`
	Faxes            string     `json:"faxes"`
	IdType           string     `json:"idType"`   //身份证明类型
	IdNum            string     `json:"idNum"`    //身份证明号码
	Comments         string     `json:"comments"` //备注
	IdUrl1           string     `json:"idUrl1"`   //身份证明A
	IdUrl2           string     `json:"idUrl2"`   //身份证明2
	Email            string     `json:"email"`
	ProductId        string     `json:"productId"`
	Creator          string     `json:"creator"`
	Enabled          bool       `json:"enabled"`
	CreateTime       int64      `json:"createTime"`
	ModifyTime       int64      `json:"modifyTime"`
	Isfollow         bool       `json:"isfollow"`
	FollowContent    string     `json:"followContent"`
	FollowWay        string     `json:"followWay"`
	FollowTime       int64      `json:"followTime"`
	Birthday         string     `json:"birthday"`
	IdAddress        string     `json:"idAddress"`
	StandbyTelephone PhoneInfo  `json:"standbyTelephone"`
	RevisitTime      int64      `json:"revisitTime"`
	Uid              string     `json:"uid"`
	OwnerType        string     `json:"ownerType"`
	Gender           string     `json:"gender"`
}

type CustomerOpportunitiesDTO struct {
	OpportunityId   string  `json:"opportunityId"` //"销售机会ID
	CustomerId      string  `json:"customerId"`    //客户ID
	CustomName      string  `json:"customName"`
	OweId           string  `json:"oweId"`   //归属ID
	OweName         string  `json:"oweName"` //归属名称
	OpportunityName string  `json:"opportunityName"`
	OpportunityType string  `json:"opportunityType"` //机会类型
	SalesStage      string  `json:"salesStage"`      //销售阶段
	ExpectAmount    float64 `json:"expectAmount"`    //预计成交金额
	ExpectTime      int64   `json:"expectTime"`      //预计成交时间
	LoseCause       string  `json:"loseCause"`       //输单原因
	Comments        string  `json:"comments"`
	TenantId        string  `json:"tenantId"`
	Creator         string  `json:"creator"`
	Enabled         bool    `json:"enabled"`
	CreateTime      int64   `json:"createTime"`
	ModifyTime      int64   `json:"modifyTime"`
	IsLose          bool    `json:"isLose"`
	StatementDate   int64   `json:"statementDate"`
	OwnerType       string  `json:"ownerType"`
}

//根据ID查询模型,ID与归属不同时过滤
type QueryByIdDTO struct {
	Ids         []string `json:"ids"`
	OwnIds      []string `json:"ownIds"`
	BindUserIds []string `json:"bindUserIds"`
	Columns     []string `json:"columns"`
}

//根据客户来源获取客户
func ClientLinkCustom(feignKey *XFeignKey, customerSource string) ([]CustomerPropertiesDTO, code.Error) {
	feignkeyBypes, _ := json.Marshal(feignKey)
	rsp, err := invoke.Name(constant.ServiceNameCustom).
		Get(constant.PathLincCustom).
		Header("X-Feign-Key", string(feignkeyBypes)).
		Route("customSource", customerSource).
		Response()

	var response = make([]CustomerPropertiesDTO, 0)
	cerr := invokeutils.ExtractHttpResponse(constant.ServiceNameCustom, err, rsp, &response)
	if cerr != nil {
		logrus.WithFields(logrus.Fields{
			"error": cerr,
		}).Error("link custom failed")
		return nil, errcode.CerrRequestCustomFeign
	}

	return response, nil
}

//批量获取销售机会
func ClientGetCustomerOpportunities(feignKey *XFeignKey, customerIds []string, salesStage string) ([]CustomerOpportunitiesDTO, code.Error) {
	feignkeyBypes, _ := json.Marshal(feignKey)
	rsp, err := invoke.Name(constant.ServiceNameCustom).
		Post(constant.PathCustomerOpportunities).
		Header("X-Feign-Key", string(feignkeyBypes)).
		Route("salesStage", salesStage).
		Json(&customerIds).
		Response()

	var response []CustomerOpportunitiesDTO = make([]CustomerOpportunitiesDTO, 0)
	cerr := invokeutils.ExtractHttpResponse(constant.ServiceNameCustom, err, rsp, &response)
	if cerr != nil {
		logrus.WithFields(logrus.Fields{
			"error": cerr,
		}).Error("link custom failed")
		return nil, errcode.CerrRequestCustomFeign
	}

	return response, nil
}

//根据ID取客户列表
func ClientCustomerPropertirsById(qyeryById *QueryByIdDTO, feignKey *XFeignKey) ([]CustomerPropertiesDTO, code.Error) {
	feignkeyBypes, _ := json.Marshal(feignKey)
	rsp, err := invoke.Name(constant.ServiceNameCustom).
		Post(constant.PathCustomQuery).
		Header("X-Feign-Key", string(feignkeyBypes)).
		Json(qyeryById).
		Response()

	var response = make([]CustomerPropertiesDTO, 0)
	cerr := invokeutils.ExtractHttpResponse(constant.ServiceNameCustom, err, rsp, &response)
	if cerr != nil {
		logrus.WithFields(logrus.Fields{
			"error": cerr,
		}).Error("custom query failed")
		return nil, errcode.CerrRequestCustomFeign
	}

	return response, nil
}

//根据ID取客户
func GetCustomerById(feignKey *XFeignKey, customerId string) (*CustomerPropertiesDTO, code.Error) {
	if customerId == "" {
		return nil, nil
	}

	queryBody := QueryByIdDTO{
		Ids:     []string{customerId},
		Columns: []string{"customName"},
	}

	feignkeyBypes, _ := json.Marshal(feignKey)
	rsp, err := invoke.Name(constant.ServiceNameCustom).
		Post(constant.PathCustomQuery).
		Header("X-Feign-Key", string(feignkeyBypes)).
		Json(queryBody).
		Response()

	var response = make([]CustomerPropertiesDTO, 0)
	cerr := invokeutils.ExtractHttpResponse(constant.ServiceNameCustom, err, rsp, &response)
	if cerr != nil {
		logrus.WithFields(logrus.Fields{
			"error": cerr,
		}).Error("custom query failed")
		return nil, errcode.CerrRequestCustomFeign
	}

	if len(response) > 0 {
		return &response[0], nil
	}
	return nil, nil
}
