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

type TenantContacts struct {
	ContactsId   string `json:"contactsId"`
	ContactsName string `json:"contactsName"`
	Email        string `json:"email"`
	Phone        string `json:"phone"`
	ContactsType string `json:"contactsType"`
	UserId       string `json:"userId"`
	ModifyTime   int64  `json:"modifyTime"`
}

type LworkTenantDTO struct {
	TenantId   string  `json:"tenantId"`
	TenantName string  `json:"tenantName"`
	Balance    float64 `json:"balance"`
	Credit     float64 `json:"credit"`
	Email      string  `json:"email"`
	Phone      string  `json:"phone"`
	//managers          Set<String>
	Password      string           `json:"password"`
	Comments      string           `json:"comments"`
	ScUserId      string           `json:"scUserId"`
	ScUserEnabled bool             `json:"scUserEnabled"`
	ModifyTime    int64            `json:"modifyTime"`
	CreateTime    int64            `json:"createTime"`
	Enabled       bool             `json:"enabled"`
	Type          string           `json:"type"` //TenantType
	Language      string           `json:"language"`
	Mode          string           `json:"mode"` //CommissionMode
	ContactsList  []TenantContacts `json:"contactsList"`
	Product       []string         `json:"product"`
	ShowPoweredBy string           `json:"showPoweredBy"`
}

type FormFieldDTO struct {
	Key              string            `json:"key"`
	Label            string            `json:"label"`
	Readonly         bool              `json:"readonly"`
	Overuse          bool              `json:"overuse"`
	Enable           bool              `json:"enable"`
	Unique           bool              `json:"unique"`
	FieldType        string            `json:"fieldType"`
	Columns          string            `json:"columns"`
	SysDefault       bool              `json:"sysDefault"`
	Size             int               `json:"size"`
	OrderNo          int               `json:"orderNo"`
	OptionList       []OptionDTO       `json:"optionList"`
	DefaultValue     string            `json:"defaultValue"`
	TwDefaultValue   string            `json:"twDefaultValue"`
	DefaultCity      map[string]string `json:"defaultCity"`
	DefaultCheckbox  []string          `json:"defaultCheckbox"`
	PlaceHolder      string            `json:"placeHolder"`
	ErrorCode        string            `json:"errorCode"`
	Searchable       bool              `json:"searchable"`
	Sensitive        bool              `json:"sensitive"`
	LongField        bool              `json:"longField"`
	BusinessSelected bool              `json:"businessSelected"`
	ValidateType     json.RawMessage   `json:"validateType"`
	Show             bool              `json:"show"`
}

type OptionDTO struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

func ClientFindTenantById(feignKey *XFeignKey, tenant_id string) (*LworkTenantDTO, code.Error) {
	feignkeyBypes, _ := json.Marshal(feignKey)
	rsp, err := invoke.Name(constant.ServiceNameTenant).
		Get(constant.PathGetTenantInfo).
		Header("X-Feign-Key", string(feignkeyBypes)).
		Route("tenantId", tenant_id).
		Response()

	var response LworkTenantDTO
	cerr := invokeutils.ExtractHttpResponse(constant.ServiceNameTenant, err, rsp, &response)
	if cerr != nil {
		logrus.WithFields(logrus.Fields{
			"error": cerr,
		}).Error("Request Find Tenant By Id failed")
		return nil, errcode.CerrRequestTenantFeign
	}

	return &response, nil
}

func ClientGetTenantOption(feignKey *XFeignKey, fieldId string) ([]OptionDTO, code.Error) {
	feignkeyBypes, _ := json.Marshal(feignKey)
	rsp, err := invoke.Name(constant.ServiceNameTenant).
		Get(constant.PathGetGenantOption).
		Header("X-Feign-Key", string(feignkeyBypes)).
		Route("fieldId", fieldId).
		Response()

	var response []OptionDTO
	cerr := invokeutils.ExtractHttpResponse(constant.ServiceNameTenant, err, rsp, &response)
	if cerr != nil {
		logrus.WithFields(logrus.Fields{
			"error": cerr,
		}).Error("Request Get Tenant Option failed")
		return nil, errcode.CerrRequestTenantFeign
	}

	return response, nil
}

func ClientFormfieldList(feignKey *XFeignKey, tableName string) ([]FormFieldDTO, code.Error) {
	feignkeyBypes, _ := json.Marshal(feignKey)
	rsp, err := invoke.Name(constant.ServiceNameTenant).
		Get(constant.PathTenantField).
		Header("X-Feign-Key", string(feignkeyBypes)).
		Query("tableName", tableName).
		Response()

	var response []FormFieldDTO
	cerr := invokeutils.ExtractHttpResponse(constant.ServiceNameTenant, err, rsp, &response)
	if cerr != nil {
		logrus.WithFields(logrus.Fields{
			"error": cerr,
		}).Error("Request Get Form Field failed")
		return nil, errcode.CerrRequestTenantFeign
	}

	return response, nil
}
