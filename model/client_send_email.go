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

type UserDTO struct{
	UserId           string              `json:"userId,omitempty"`
	UserName         string              `json:"userName,omitempty"`
}

type TenantDTO struct {
	TenantId         string             `json:"tenantId"`
	ProductId        string             `json:"productId"`
}

type MessageBase struct{
	Content          string             `json:"content,omitempty"`		//内容，长度5000，内容与模板id不能同时为空
	Title            string             `json:"title,omitempty"`
	TemplateId       int                `json:"templateId,omitempty"`
	TemplateName     string             `json:"templateName,omitempty"`
	MessageConfigId  int                `json:"messageConfigId,omitempty"`
	Vars             map[string]string  `json:"vars"`
}

type EmailDTO struct{
	ToUser          []UserDTO           `json:"toUser"`
	SendUser        UserDTO             `json:"sendUser"`				//发件人姓名（邮箱）          为空则为系统用户，为userId:-1,userName:system"
	From            string              `json:"from,omitempty"`
	Message         MessageBase         `json:"message"`
	TemplateType    string              `json:"templateType,omitempty"`	//模板类型，默认为通用
	TenantInfo      TenantDTO           `json:"tenantDTO"`
	AdditionInfo    string              `json:"additionInfo,omitempty"`
	Lang            string              `json:"lang,omitempty"`
	Mobile          string              `json:"mobile,omitempty"`
}



func ClientSendAddUserEmail(feignKey *XFeignKey, email *EmailDTO ) (string, code.Error){
	feignkeyBypes,_ := json.Marshal(feignKey)
	rsp, err := invoke.Name(constant.ServiceNameBwMessage).
	Post(constant.PathBwMessageSendMailByTemplateType).
	Header("X-Feign-Key", string(feignkeyBypes)).
	Json(email).
	Response()

	var response string
	cerr := invokeutils.ExtractHttpResponse(constant.ServiceNameBwMessage, err, rsp, &response)
	if cerr != nil {
		logrus.WithFields(logrus.Fields{
			"toUser": email.ToUser,
			"error": cerr,
		}).Error("Send email info to message server failed")
		return "", errcode.CerrReuestMailSer
	}

	logrus.WithFields(logrus.Fields{
		"toUser": email.ToUser,
		"data": response,
	}).Debug("Send email info to message server success")
	return response, nil
}
