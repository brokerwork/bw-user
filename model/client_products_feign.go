package model

import (
	"github.com/lworkltd/kits/service/restful/code"
	"bw/bw-user/constant"
	"github.com/lworkltd/kits/service/invoke"
	invokeutils "github.com/lworkltd/kits/utils/invoke"
	"github.com/Sirupsen/logrus"
	"bw/bw-user/errcode"
	"encoding/json"
)

type PlatformLimit struct{
	Vendor                string        `json:"vendor"`
	SecondLimited         int64         `json:"secondLimited"`
}

type TenantVasDTO struct {				//还有ValueAddSrvDTO的成员可以加入
	OrderQty              int           `json:"orderQty"`
	OrderedTime           int64         `json:"orderedTime"`
	ExpiredTime           int64         `json:"expiredTime"`
	Amount                float64       `json:"amount"`
	PayAmount             float64       `json:"payAmount"`
	TotalQty              int64         `json:"totalQty"`
}

type TenantProductDTO struct{
	TpId                  string        `json:"tpId"`
	TenantId              string        `json:"tenantId"`
	TenantName            string        `json:"tenantName"`
	ProductId             string        `json:"productId"`
	VersionId             string        `json:"versionId"`
	VersionName           string        `json:"versionName"`
	SiteName              string        `json:"siteName"`
	ProductDomain         string        `json:"productDomain"`
	CustomerDomain        string        `json:"customerDomain"`
	CustomerDomainScheme        string        `json:"customerDomainScheme"`
	ProductLogo           string        `json:"productLogo"`
	ProductIcon           string        `json:"productIcon"`
	MbileLogo            string        `json:"mobileLogo"`
	CompanyName           string        `json:"companyName"`
	CompanyEmail          string        `json:"companyEmail"`
	CompanyPhone          string        `json:"companyPhone"`
	CompanyAddress        string        `json:"companyAddress"`
	CompanySite           string        `json:"companySite"`
	OlCustomerSrv         string        `json:"olCustomerSrv"`
	Feedback              string        `json:"feedback"`
	StatsCode             string        `json:"statsCode"`
	PeriodType            int           `json:"periodType"`
	Period                int           `json:"period"`
	Started               int64         `json:"started"`
	Expired               int64         `json:"expired"`
	NumLimited            int64         `json:"numLimited"`
	Allowed               bool          `json:"allowed"`
	MinuteLimited         int64         `json:"minuteLimited"`
	PlatformLimitList     []PlatformLimit `json:"platformLimitList"`
	PlatformId            string        `json:"platformId"`
	PlatformKey           string        `json:"platformKey"`
	PayNotifyUrl          string        `json:"payNotifyUrl"`
	PayResultUrl          string        `json:"payResultUrl"`
	PayUrl                string        `json:"payUrl"`
	ProductFee            float64       `json:"productFee"`
	Actived               bool          `json:"actived"`
	Token                 string        `json:"token"`
	PublicKey             string        `json:"publicKey"`
	PublicKeyUpdateTime   int64         `json:"publicKeyUpdateTime"`
	ThemeId               string        `json:"themeId"`
	TokenExpired          int64         `json:"tokenExpired"`
	Vass                  []TenantVasDTO `json:"vass"`
	Language4mail         string        `json:"language4mail"`
	CreateTime            int64         `json:"createTime"`
	ModifyTime            int64         `json:"modifyTime"`
	//ProductVersionDTO version;  暂时未加入
}


type VendorConnector struct {
	Name      string `json:"name"`
	Host      string `json:"host"`
	Port      int    `json:"port"`
	Login     string `json:"login"`
	Password  string `json:"password"`
	Status    int    `json:"status"`
	BeginNo   int64  `json:"beginNo"`
	EndNo     int64  `json:"endNo"`
	CurrentNo int64  `json:"currentNo"`
	Vendor    string `json:"vendor"`
	Type      string `json:"type"`
	Enabled   bool   `json:"enabled"`
}
type ProductDeployDTO struct {
	ServerId        int64           `json:"serverId"`
	SchemeId        int64           `json:"schemeId"`
	TenantId        string          `json:"tenantId"`
	ProductId       string          `json:"productId"`
	Host            string          `json:"host"`
	Url             string          `json:"url"`
	Scheme          string          `json:"scheme"`
	User            string          `json:"user"`
	Password        string          `json:"password"`
	NumLimit        int64           `json:"numLimit"`
	VendorConnector VendorConnector `json:"vendorConnector"`
	UpdateTime      int64           `json:"updateTime"`
}


func ClientProductTenantByKey(feignKey *XFeignKey, tenant_id string, productId string) (*TenantProductDTO, code.Error){
	feignkeyBypes,_ := json.Marshal(feignKey)
	rsp, err := invoke.Name(constant.ServiceNameProducts).
	Get(constant.PathTenantByKey).
	Header("X-Feign-Key", string(feignkeyBypes)).
	Query("tenantId", tenant_id).
	Query("productId", productId).
	Response()

	var response TenantProductDTO
	cerr := invokeutils.ExtractHttpResponse(constant.ServiceNameProducts, err, rsp, &response)
	if cerr != nil {
		logrus.WithFields(logrus.Fields{"error": cerr,}).Error("Request product tenant failed")
		return nil, errcode.CerrRequestProductFeign
	}

	return &response, nil
}

func ClientProductPermission(feignKey *XFeignKey, tenantId string, productId string, language string) ([]TenantPermissionDTO, code.Error){
	if "" == language {
		language = "zh-CN"
	}
	feignkeyBypes,_ := json.Marshal(feignKey)
	rsp, err := invoke.Name(constant.ServiceNameProducts).
		Get(constant.PathProducePermission).
		Header("X-Feign-Key", string(feignkeyBypes)).
		Query("tenantId", tenantId).
		Query("productId", productId).
		Query("language", language).
		Response()

	var response []TenantPermissionDTO
	cerr := invokeutils.ExtractHttpResponse(constant.ServiceNameProducts, err, rsp, &response)
	if cerr != nil {
		logrus.WithFields(logrus.Fields{"error": cerr,}).Error("Request product tenant failed")
		return nil, errcode.CerrRequestProductFeign
	}

	return response, nil
}



func ClientGetProductMysql(feignKey *XFeignKey) (*ProductDeployDTO, code.Error){
	feignkeyBypes,_ := json.Marshal(feignKey)
	rsp, err := invoke.Name(constant.ServiceNameProducts).
		Get(constant.PathGetProductDeploy).
		Header("X-Feign-Key", string(feignkeyBypes)).
		Response()

	var response ProductDeployDTO
	cerr := invokeutils.ExtractHttpResponse(constant.ServiceNameProducts, err, rsp, &response)
	if cerr != nil {
		logrus.WithFields(logrus.Fields{"error": cerr,}).Error("Request product tenant failed")
		return nil, errcode.CerrRequestProductFeign
	}

	return &response, nil
}


func ClientGetProductUserLimit(feignKey *XFeignKey) (int, code.Error){
	feignkeyBypes,_ := json.Marshal(feignKey)
	rsp, err := invoke.Name(constant.ServiceNameProducts).
		Get(constant.PathProductUserLimit).
		Header("X-Feign-Key", string(feignkeyBypes)).
		Query("tenantId", feignKey.TenantId).
		Query("productId", feignKey.ProductId).
		Response()

	var response int
	cerr := invokeutils.ExtractHttpResponse(constant.ServiceNameProducts, err, rsp, &response)
	if cerr != nil {
		logrus.WithFields(logrus.Fields{"error": cerr,}).Error("Request product tenant failed")
		return 0, errcode.CerrRequestProductFeign
	}

	return response, nil
}
