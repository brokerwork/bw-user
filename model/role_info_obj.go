package model

import (
	//"github.com/lworkltd/kits/service/profile"
	//"github.com/jinzhu/gorm"

	"database/sql"
	//"github.com/lworkltd/kits/service/restful/code"
	//"github.com/Sirupsen/logrus"
	//"bw/user/errcode"
	"time"
)

type RoleTab struct {
	Id              int64         `gorm:"column:id" gorm:"primary_key"`
	CreateDate      time.Time     `gorm:"column:create_date"`
	CreateUserId    string        `gorm:"column:create_user_id"`
	EntityNo        string        `gorm:"column:entity_no"`
	ModelStatus     string        `gorm:"column:model_status"` //constant.ModelStatusCreate/ModelStatusLock/ModelStatusSuccess/ModelStatusDelete
	ModifyDate      time.Time     `gorm:"column:modify_date"`
	ModifyUserId    string        `gorm:"column:modify_user_id"`
	ProductId       string        `gorm:"column:product_id"`
	TenantId        string        `gorm:"column:tenant_id"`
	BelongUserCount int           `gorm:"column:belong_user_count"`
	Comment         string        `gorm:"column:comment"`
	Name            string        `gorm:"column:name"`
	RoleRightFlags  int64         `gorm:"column:role_right_flags"`
	SubRoleCount    int           `gorm:"column:sub_role_count"`
	ParentId        sql.NullInt64 `gorm:"column:parent_id"`
	RoleTypeId      string        `gorm:"column:role_type_id"`
}

// ALTER TABLE t_role ADD COLUMN role_type_id varchar(32) NOT NULL DEFAULT '' COMMENT '角色类型ID';
type RoleType struct {
	Id       string            `bson:"_id" json:"id"`
	TypeName map[string]string `json:"typeName" bson:"typeName"`
	EntityNo []string          `bson:"entityNo" json:"entityNo"`
}

type RoleTypeDTO struct {
	Id       string `json:"id"`
	TypeName string `json:"typeName"`
}

func (u RoleTab) TableName() string {
	return "t_role"
}

type RoleSimple struct {
	Id           int64  `json:"id"`
	Name         string `json:"name"`
	EntityNo     string `json:"entityNo"`
	RoleTypeId   string `json:"roleTypeId"`
	RoleTypeName string `json:"roleTypeName"`
}

type RoleDetailWithInfo struct {
	Id              int64                `json:"id"`
	EntityNo        string               `json:"entityNo"`
	Name            string               `json:"name"`
	ParentId        int64                `json:"parentId"`
	ParentName      string               `json:"parentName"`
	SubRoleCount    int                  `json:"subRoleCount"`
	Rights          []string             `json:"rights"`
	Children        []RoleDetailWithInfo `json:"children"`
	Comment         string               `json:"comment"`
	BelongUserCount int                  `json:"belongUserCount"`
	RoleTypeId      string               `json:"roleTypeId"`
	RoleTypeName    string               `json:"roleTypeName"`
}

type RoleDetail struct {
	Id              int64        `json:"id"`
	EntityNo        string       `json:"entityNo"`
	CreateDate      int64        `json:"createDate"`
	ModifyDate      int64        `json:"modifyDate"`
	TenantId        string       `json:"tenantId"`
	ProductId       string       `json:"productId"`
	CreateUserId    string       `json:"createUserId"`
	ModelStatus     string       `json:"modelStatus"`
	Name            string       `json:"name"`
	ParentId        int64        `json:"parentId"`
	Children        []RoleDetail `json:"children"`
	RoleRightflags  int64        `json:"roleRightflags"`
	BelongUserCount int          `json:"belongUserCount"`
	SubRoleCount    int          `json:"subRoleCount"`
	Comment         string       `json:"comment"`
}

type IdSearchDTO struct {
	Ids    []string `json:"ids"`
	IdType string   `json:"idType"`
}

//用于新增/修改role
type RoleDTO struct {
	Id         int64   `json:"id"`
	EntityNo   string  `json:"entityNo"`
	Name       string  `json:"name"`
	RightIds   []int64 `json:"rightIds"`
	Comment    string  `json:"comment"`
	RoleTypeId string  `json:"roleTypeId"`
}
