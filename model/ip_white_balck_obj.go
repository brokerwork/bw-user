package model

import (
	//"github.com/lworkltd/kits/service/profile"
	//"github.com/jinzhu/gorm"

	//"github.com/lworkltd/kits/service/restful/code"
	//"github.com/Sirupsen/logrus"
	//"bw/user/errcode"
	"time"
)

type IpWhiteBlackTab struct {
	Id                   int64      `gorm:"column:id" gorm:"primary_key"`
	CreateDate           time.Time  `gorm:"column:create_date"`
	CreateUserId         string     `gorm:"column:create_user_id"`
	EntityNo             string     `gorm:"column:entity_no"`
	ModelStatus          string     `gorm:"column:model_status"`		//constant.ModelStatusCreate/ModelStatusLock/ModelStatusSuccess/ModelStatusDelete
	ModifyDate           time.Time  `gorm:"column:modify_date"`
	ModifyUserId         string     `gorm:"column:modify_user_id"`
	ProductId            string     `gorm:"column:product_id"`
	TenantId             string     `gorm:"column:tenant_id"`
	Enable               bool       `gorm:"colume:enable"`
	FromIp               string     `gorm:"columne:from_ip"`
	ToIp                 string     `gorm:"colume:to_ip"`
	User                 string      `gorm:"colume:user"`
	White                bool        `gorm:"colume:white"`
}

func (u IpWhiteBlackTab) TableName() string {
	return "t_user_detail_ip_white_black"
}

