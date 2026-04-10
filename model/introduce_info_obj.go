package model

import (
	//"github.com/lworkltd/kits/service/profile"
	//"github.com/jinzhu/gorm"

	"bw/bw-user/constant"
	"bw/bw-user/errcode"
	"github.com/lworkltd/kits/service/restful/code"
	//"github.com/lworkltd/kits/service/restful/code"
	//"github.com/Sirupsen/logrus"
	//"bw/user/errcode"
	"time"

	//"database/sql"
	"gopkg.in/mgo.v2/bson"
	"strings"
)

type SystemIntroduceTab struct {
	Id                int64     `gorm:"column:id" gorm:"primary_key"`
	CreateDate        time.Time `gorm:"column:create_date"`
	CreateUserId      string    `gorm:"column:create_user_id"`
	EntityNo          string    `gorm:"column:entity_no"`
	ModelStatus       string    `gorm:"column:model_status"` //constant.ModelStatusCreate/ModelStatusLock/ModelStatusSuccess/ModelStatusDelete
	ModifyDate        time.Time `gorm:"column:modify_date"`
	ModifyUserId      string    `gorm:"column:modify_user_id"`
	ProductId         string    `gorm:"column:product_id"`
	TenantId          string    `gorm:"column:tenant_id"`
	Enable            int8      `gorm:"column:enable"`
	Platform          string    `gorm:"column:platform"`
	Type              string    `gorm:"column:type"`
	Name              string    `gorm:"column:name"`
	Url               string    `gorm:"column:url"`
	DisplayUrl        string    `gorm:"column:display_url"`
	QrCode            string    `gorm:"column:qr_code"`
	ParameterType     string    `gorm:"column:parameter_type"`
	BwUserShow        string    `gorm:"column:bw_user_show"`
	VisibleUser       string    `gorm:"column:visible_user"`
	VisibleUserName   string    `gorm:"column:visible_user_name"`
	Participants      string    `gorm:"column:participant_user"`
	ParticipantNames  string    `gorm:"column:participant_names"`
	ServerId          string    `gorm:"column:server_id"`
	Vendor            string    `gorm:"column:vendor"`
	MtGroup           string    `gorm:"column:mt_group"`
	Leverage          int       `gorm:"column:leverage"`
	AccountGroup      string    `gorm:"column:account_group"`
	OwnerId           string    `gorm:"column:owner_id"`
	OwnerType         string    `gorm:"column:owner_type"`
	BusinessCode      string    `gorm:"column:business_code"`
	Os                string    `gorm:"column:os"`
	SoftwarePackage   string    `gorm:"column:software_package"`
	InVisibleUser     string    `gorm:"column:invisible_user"`
	InVisibleUserName string    `gorm:"column:invisible_user_name"`
}

func (record *SystemIntroduceTab) updateFromSystemIntroduceDTO(dto *SystemIntroduceDTO) code.Error {
	record.BwUserShow = dto.BwUserShow
	if "" == record.BwUserShow {
		record.BwUserShow = constant.IntroduceType_UserNotVisible
	}
	if constant.IntroduceType_UserPartVisible == dto.BwUserShow || constant.IntroduceType_DirectPartVisible == dto.BwUserShow {
		//部分可见必须有用户信息
		if len(dto.VisibleUser) == 0 || len(dto.VisibleUserName) == 0 {
			return errcode.CerrParamater
		}
		record.VisibleUser = strings.Join(dto.VisibleUser, "@-@") //Id-IdType
		record.VisibleUserName = strings.Join(dto.VisibleUserName, "@-@")
		record.InVisibleUser = ""
		record.InVisibleUserName = ""
	} else if constant.IntroduceType_UserInVisible == dto.BwUserShow || constant.IntroduceType_DirectPartInvisible == dto.BwUserShow {
		//部分不可见必须有用户信息
		if len(dto.InVisibleUser) == 0 || len(dto.InVisibleUserName) == 0 {
			return errcode.CerrParamater
		}
		record.InVisibleUser = strings.Join(dto.InVisibleUser, "@-@") //Id-IdType
		record.InVisibleUserName = strings.Join(dto.InVisibleUserName, "@-@")
		record.VisibleUser = ""
		record.VisibleUserName = ""
	} else if constant.IntroduceType_UserAllVisible == dto.BwUserShow || constant.IntroduceType_UserNotVisible == dto.BwUserShow ||
		constant.IntroduceType_DirectAllVisible == dto.BwUserShow || constant.IntroduceType_DirectNotVisible == dto.BwUserShow {
		record.VisibleUser = ""
		record.VisibleUserName = ""
		record.InVisibleUser = ""
		record.InVisibleUserName = ""
	}

	if constant.IntroduceType_StraightGuest == record.Type || constant.IntroduceType_DirectRecommendation == record.Type {
		record.ParameterType = dto.ParameterType
		record.ServerId = dto.ServerId
		record.Vendor = dto.Vendor
		record.MtGroup = dto.MtGroup
		record.AccountGroup = dto.AccountGroup
		record.Leverage = dto.Leverage
	} else {
		record.ParameterType = ""
		record.ServerId = ""
		record.Vendor = ""
		record.MtGroup = ""
		record.AccountGroup = ""
		record.Leverage = 0
	}
	record.OwnerType = dto.OwnerType
	record.OwnerId = dto.OwnerId
	if dto.Participants == nil {
		record.Participants = ""
	} else {
		record.Participants = strings.Join(dto.Participants, "@-@")
	}
	if dto.ParticipantNames == nil {
		record.ParticipantNames = ""
	} else {
		record.ParticipantNames = strings.Join(dto.ParticipantNames, "@-@")
	}
	return nil
}

func (u SystemIntroduceTab) TableName() string {
	return "t_system_introduce"
}

type SystemIntroduceDTO struct {
	TenantId          string   `json:"tenantId"`
	Id                int64    `json:"id"`
	Creator           string   `json:"creator"`
	CreateTime        int64    `json:"createTime"`
	EntityNo          string   `json:"entityNo"`           //推广编号
	Platform          string   `json:"platform,omitempty"` //IntroduceType 平台, Web/Mobile
	Name              string   `json:"name"`               //推广名称
	Enable            bool     `json:"enable"`
	Type              string   `json:"type,omitempty"`          //IntroduceType 推广类型，代理Agent/直客StraightGuest
	BwUserShow        string   `json:"bwUserShow,omitempty"`    //IntroduceType 用户可见范围,UserAllVisible, UserPartVisible, UserNotVisible,UserInVisible
	VisibleUser       []string `json:"visibleUser"`             //可见用户，代理Agent/直客StraightGuest
	VisibleUserName   []string `json:"visibleUserName"`         //可见用户Name，代理Agent/直客StraightGuest
	InVisibleUser     []string `json:"inVisibleUser"`           //可见用户，代理Agent/直客StraightGuest
	InVisibleUserName []string `json:"inVisibleUserName"`       //可见用户Name，代理Agent/直客StraightGuest
	Participants      []string `json:"participants"`            //参与人
	ParticipantNames  []string `json:"participantNames"`        //参与人
	Url               string   `json:"url"`                     //用户设置的推广地址
	DisplayUrl        string   `json:"displayUrl"`              //包装后展示的推广地址
	QrCode            string   `json:"qrCode"`                  //二维码
	ParameterType     string   `json:"parameterType,omitempty"` //IntroduceType 参数类型
	ServerId          string   `json:"serverId"`                //交易服务器
	Vendor            string   `json:"vendor,omitempty"`
	MtGroup           string   `json:"mtGroup"`
	AccountGroup      string   `json:"accountGroup"`        //账户组
	Leverage          int      `json:"leverage,omitempty"`  //杠杆
	OwnerType         string   `json:"ownerType,omitempty"` //默认归属类型，只有WEB和不展示BW用户的才有该值，值为RoleId/Id
	OwnerId           string   `json:"ownerId"`             //默认归属ID，只有WEB和不展示BW用户的才有该值
	OwnerName         string   `json:"ownerName"`           //默认归属名，冗余字段，用于展示
	BusinessCode      string   `json:"businessCode"`        //编号
	Os                string   `json:"os"`                  //应用平台，比如iOS/Android
	SoftwarePackage   string   `json:"softwarePackage"`     //软件包名称
}

type SystemIntroduceHit struct {
	Id          bson.ObjectId `bson:"_id"`
	Url         string        `bson:"url"`
	IntroduceId string        `bson:"introduceId"`
	UserId      string        `bson:"userId"`
	TenantId    string        `bson:"tenantId"`
	ClientIp    string        `bson:"clientIp"`
	Device      string        `bson:"device"`
	HitTime     time.Time     `bson:"time"`
}

type IntroduceHitDTO struct {
	Url         string `json:"url"`         //原始访问url
	IntroduceId string `json:"introduceId"` //推广链接Entity ID
	UserId      string `json:"userId"`      //推广用户ID
	TenantId    string `json:"tenantId"`
	ClientIp    string `json:"clientIp"`
	Device      string `json:"device"`
	HitTime     int64  `json:"time"`
}

//直客类型推广链接统计详情DTO
type IntroduceStatisticDTO struct {
	SystemIntroduceDTO
	HitNumber         int64                   `json:"hitNumber"`         //点击人数
	NewCustomerNumber int64                   `json:"newCustomerNumber"` //新增客户
	OpenAccountNumber int64                   `json:"openAccountNumber"` //开户客户数
	DepositeNumber    int64                   `json:"depositeNumber"`    //入金客户数
	WinCustomerNumber int64                   `json:"winCustomerNumber"` //赢单客户数
	ContractAmount    int64                   `json:"contractAmount"`    //合同金额
	ApplyNumber       int64                   `json:"applyNumber"`       //申请人数
	PassNumber        int64                   `json:"passNumber"`        //通过人数
	NotPassNumber     int64                   `json:"notPassNumber"`     //未通过人数
	UserStatistic     []IntroduceStatisticDTO `json:"userStatistic"`     //用户级统计明细
}

func (static *IntroduceStatisticDTO) copyFromSystemIntroduceTab(record *SystemIntroduceTab) {
	if nil == record {
		return
	}
	static.Id = record.Id
	static.TenantId = record.TenantId
	static.Creator = record.CreateUserId
	static.CreateTime = record.CreateDate.Unix()
	static.EntityNo = record.EntityNo
	static.Platform = record.Platform
	static.Name = record.Name
	static.Enable = false
	if record.Enable != 0 {
		static.Enable = true
	}
	static.Type = record.Type
	static.BwUserShow = record.BwUserShow
	if "" != record.VisibleUser {
		static.VisibleUser = strings.Split(record.VisibleUser, "@-@")
		for index := range static.VisibleUser {
			idAndIdType := strings.Split(static.VisibleUser[index], "-")
			if len(idAndIdType) == 1 {
				static.VisibleUser[index] = static.VisibleUser[index] + "-Id"
			}
		}
	}
	if "" != record.VisibleUserName {
		static.VisibleUserName = strings.Split(record.VisibleUserName, "@-@")
	}
	if "" != record.InVisibleUser {
		static.InVisibleUser = strings.Split(record.InVisibleUser, "@-@")
		for index := range static.InVisibleUser {
			idAndIdType := strings.Split(static.InVisibleUser[index], "-")
			if len(idAndIdType) == 1 {
				static.InVisibleUser[index] = static.InVisibleUser[index] + "-Id"
			}
		}
	}
	if "" != record.InVisibleUserName {
		static.InVisibleUserName = strings.Split(record.InVisibleUserName, "@-@")
	}
	if "" != record.Participants {
		static.Participants = strings.Split(record.Participants, "@-@")
	}
	if "" != record.ParticipantNames {
		static.ParticipantNames = strings.Split(record.ParticipantNames, "@-@")
	}
	static.Url = record.Url
	static.DisplayUrl = record.DisplayUrl
	static.QrCode = record.QrCode
	static.ParameterType = record.ParameterType
	static.ServerId = record.ServerId
	static.Vendor = record.Vendor
	static.MtGroup = record.MtGroup
	static.AccountGroup = record.AccountGroup
	static.Leverage = record.Leverage
	static.OwnerType = record.OwnerType
	static.OwnerId = record.OwnerId
	static.BusinessCode = record.BusinessCode
	static.Os = record.Os
	static.SoftwarePackage = record.SoftwarePackage
}

type Account struct {
	ServerId   string `json:"serverId"`
	AccountId  string `json:"accountId"`
	CustomerId string `json:"customerId"`
	HasDeposit bool   `json:"hasDeposit"`
}
