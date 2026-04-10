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

type RightTab struct {
	Id           int64         `gorm:"column:id" gorm:"primary_key"`
	CreateDate   time.Time     `gorm:"column:create_date"`
	CreateUserId string        `gorm:"column:create_user_id"`
	EntityNo     string        `gorm:"column:entity_no"`
	ModelStatus  string        `gorm:"column:model_status"` //constant.ModelStatusCreate/ModelStatusLock/ModelStatusSuccess/ModelStatusDelete
	ModifyDate   time.Time     `gorm:"column:modify_date"`
	ModifyUserId string        `gorm:"column:modify_user_id"`
	ProductId    string        `gorm:"column:product_id"`
	TenantId     string        `gorm:"column:tenant_id"`
	Comment      string        `gorm:"column:comment"`
	Flag         int64         `gorm:"column:flag"`
	Name         string        `gorm:"column:name"`
	Type         string        `gorm:"column:type"`
	ParentId     sql.NullInt64 `gorm:"column:parent_id"`
	/**
	 * 默认选中: OWN
	 * 如果拥有上级全新责备选中: PARENT
	 * 如果拥有某种权限则被选中: 指定权限authNo
	 */
	Dependence string `gorm:"-"`
}

func (u RightTab) TableName() string {
	return "t_right"
}

//无主键
type RoleRightRelationTab struct {
	RoleID  int64 `gorm:"column:t_role"`
	RightID int64 `gorm:"column:role_rights"`
}

func (u RoleRightRelationTab) TableName() string {
	return "t_role_right_relation"
}

type RightDetail struct {
	Id           int64         `json:"id"`
	EntityNo     string        `json:"entityNo"`
	CreateDate   int64         `json:"createDate"`
	ModifyDate   int64         `json:"modifyDate"`
	TenantId     string        `json:"tenantId"`
	ProductId    string        `json:"productId"`
	CreateUserId string        `json:"createUserId"`
	ModelStatus  string        `json:"modelStatus"`
	Name         string        `json:"name"`
	ParentId     int64         `json:"parentId"`
	Flag         int64         `json:"flag"`
	Comment      string        `json:"comment"`
	Type         string        `json:"type"`
	Children     []RightDetail `json:"children"`
}

type PermissionNode struct {
	ModulePermissionId string           `json:"modulePermissionId"`
	AuthNo             string           `json:"authNo"`
	AuthName           string           `json:"authName"`
	Flag               int64            `json:"flag"`
	Parent             *PermissionNode  `json:"parent"`
	Children           []PermissionNode `json:"children"`
	/**
	 * 默认选中: OWN
	 * 如果拥有上级全新责备选中: PARENT
	 * 如果拥有某种权限则被选中: 指定权限authNo
	 */
	Dependence string `json:"dependence"`
}

type TenantPermissionDTO struct {
	ModuleCode  string           `json:"moduleCode"`
	ModuleName  string           `json:"moduleName"`
	Permissions []PermissionNode `json:"permissions"`
}

type KeyIdRightDTO struct {
	KeyType       string   `json:"keyType"`
	RightEntityNo string   `json:"rightEntityNo"`
	Ids           []string `json:"ids"`
}
