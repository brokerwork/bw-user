package model

import (
	//"github.com/lworkltd/kits/service/profile"
	//"github.com/jinzhu/gorm"

	//"github.com/lworkltd/kits/service/restful/code"
	//"github.com/Sirupsen/logrus"
	//"bw/user/errcode"
	//"time"
	//"database/sql"
	//"gopkg.in/mgo.v2/bson"
)


type AdvanceSearch struct{
	SearchId     string         `bson:"_id" json:"searchId"`
	TenantId     string         `bson:"tenantId" json:"tenantId"`
	RoleIds      []string       `bson:"roleIds" json:"roleIds"`
	Name         string         `bson:"name" json:"name"`
	LogicType    string         `bson:"logicType" json:"logicType,omitempty"`
	Condition    interface{}   `bson:"condition" json:"condition"`
	SearchType   string         `bson:"searchType" json:"searchType"`
	SearchLevel  string         `bson:"searchLevel" json:"searchLevel,omitempty"`
	Enabled      bool           `bson:"enabled" json:"enabled"`
	CreateUserId string         `bson:"createUserId" json:"createUserId"`
	ModifyUserId string         `bson:"modifyUserId" json:"modifyUserId"`
	CreateTime   int64          `bson:"createTime" json:"createTime"`
	ModifyTime   int64          `bson:"modifyTime" json:"modifyTime"`
}


type AdvanceSearchDTO struct{
	SearchId     string         `json:"searchId"`
	TenantId     string         `json:"tenantId"`
	RoleIds      []string       `json:"roleIds"`
	RoleNames    []string       `json:"roleNames"`
	Name         string         `json:"name"`
	LogicType    string         `json:"logicType,omitempty"`			//LogicType类型
	Condition    interface{}   `json:"condition"`
	SearchType   string         `json:"searchType"`
	SearchLevel  string         `json:"searchLevel,omitempty"`			//SearchLevel类型
	CreateUserId string         `json:"createUserId"`
	ModifyUserId string         `json:"modifyUserId"`
	CreateTime   int64          `json:"createTime"`
	ModifyTime   int64          `json:"modifyTime"`
}