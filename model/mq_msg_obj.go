package model

type BwMsgEvent struct {
	TenantId string `json:"tenantId"`

	Type    string `json:"type"`//事件类型
	Id      string `json:"id"`
	ObjType string `json:"objType"`//对象类型

	Detail interface{} `json:"detail"`
}
