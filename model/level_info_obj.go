package model

import (
	//"github.com/lworkltd/kits/service/profile"
	//"github.com/jinzhu/gorm"

	//"github.com/lworkltd/kits/service/restful/code"
	//"github.com/Sirupsen/logrus"
	//"bw/user/errcode"
	"time"
	"database/sql"
)

type LevelTab struct {
	Id                   int64      `gorm:"column:id" gorm:"primary_key"`
	CreateDate           time.Time  `gorm:"column:create_date"`
	CreateUserId         string     `gorm:"column:create_user_id"`
	EntityNo             string     `gorm:"column:entity_no"`
	ModelStatus          string     `gorm:"column:model_status"`		//constant.ModelStatusCreate/ModelStatusLock/ModelStatusSuccess/ModelStatusDelete
	ModifyDate           time.Time  `gorm:"column:modify_date"`
	ModifyUserId         string     `gorm:"column:modify_user_id"`
	ProductId            string     `gorm:"column:product_id"`
	TenantId             string     `gorm:"column:tenant_id"`
	Comment              string     `gorm:"column:comment"`
	Name                 string     `gorm:"column:name"`
	Sid                  int        `gorm:"column:sid"`
	UserCount            int        `gorm:"column:user_count"`
	ParentId             sql.NullInt64  `gorm:"column:parent_id"`
}

func (u LevelTab) TableName() string {
	return "t_level"
}


type LevelDetail struct {
	Id                   int64      `json:"id"`
	EntityNo             string     `json:"entityNo"`
	CreateDate           int64      `json:"createDate"`
	ModifyDate           int64      `json:"modifyDate"`
	TenantId             string     `json:"tenantId"`
	ProductId            string     `json:"productId"`
	CreateUserId         string     `json:"createUserId"`
	ModelStatus          string     `json:"modelStatus"`
	Name                 string     `json:"name"`
	Sid                  int        `json:"sid"`
	UserCount            int        `json:"userCount"`
	Comment              string     `json:"comment"`
}


type LevelDetailSlice [] LevelDetail
func (a LevelDetailSlice) Len() int {   // 重写 Len() 方法
	return len(a)
}
func (a LevelDetailSlice) Swap(i, j int){  // 重写 Swap() 方法
	a[i], a[j] = a[j], a[i]
}
func (a LevelDetailSlice) Less(i, j int) bool { // 重写 Less() 方法， 根据sid从到大排序
	return a[i].Sid < a[j].Sid
}