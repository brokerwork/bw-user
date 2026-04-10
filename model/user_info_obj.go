package model

import (
	//"github.com/lworkltd/kits/service/profile"
	//"github.com/jinzhu/gorm"

	"strings"
	//"github.com/lworkltd/kits/service/restful/code"
	//"github.com/Sirupsen/logrus"
	//"bw/user/errcode"
	"time"
)

type UserDetailTab struct {
	Id               int64     `gorm:"column:id" gorm:"primary_key"`
	CreateDate       time.Time `gorm:"column:create_date"`
	CreateUserId     string    `gorm:"column:create_user_id"`
	EntityNo         string    `gorm:"column:entity_no"`
	ModelStatus      string    `gorm:"column:model_status"` //constant.ModelStatusCreate/ModelStatusLock/ModelStatusSuccess/ModelStatusDelete
	ModifyDate       time.Time `gorm:"column:modify_date"`
	ModifyUserId     string    `gorm:"column:modify_user_id"`
	ProductId        string    `gorm:"column:product_id"`
	TenantId         string    `gorm:"column:tenant_id"`
	Active           int       `gorm:"column:active" gorm:"not null"` //default is 1
	Address          string    `gorm:"column:address"`
	Birthday         string    `gorm:"column:birthday"`
	City             string    `gorm:"column:city"`
	Comment          string    `gorm:"column:comment"`
	Country          string    `gorm:"column:country"`
	Email            string    `gorm:"column:email"`
	HeadImage        string    `gorm:"column:head_image"`
	LevelId          int64     `gorm:"column:level_id" gorm:"not null"`
	LevelName        string    `gorm:"column:level_name"`
	Login            string    `gorm:"column:login"`
	Name             string    `gorm:"column:name"`
	NeedInitPass     []uint8   `gorm:"column:need_init_pass" gorm:"not null"` //value only allow 0 or 1
	Nickname         string    `gorm:"column:nickname"`
	ParentId         string    `gorm:"column:parent_id"`
	Phone            string    `gorm:"column:phone"`
	Postcode         string    `gorm:"column:postcode"`
	Province         string    `gorm:"column:province"`
	PubUserId        string    `gorm:"column:pub_user_id"`
	RoleId           int64     `gorm:"column:role_id" gorm:"not null"`
	RoleName         string    `gorm:"column:role_name"`
	Sex              string    `gorm:"column:sex"`
	SubUserCount     int       `gorm:"column:sub_user_count" gorm:"not null"`
	Username         string    `gorm:"column:username"`
	VendorServerId   string    `gorm:"column:vendor_server_id"`
	Version          int       `gorm:"column:version" gorm:"not null"`
	IdType           string    `gorm:"column:id_type"`
	IdNum            string    `gorm:"column:id_num"`
	IdUrl1           string    `gorm:"column:id_url1"`
	IdUrl2           string    `gorm:"column:id_url2"`
	BankAccount      string    `gorm:"column:bank_account"`
	BankBranch       string    `gorm:"column:bank_branch"`
	AccountNo        string    `gorm:"column:account_no"`
	BankCardFile1    string    `gorm:"column:bank_card_file1"`
	BankCardFile2    string    `gorm:"column:bank_card_file2"`
	DoAgencyBusiness string    `gorm:"column:do_agency_business"`
	InvestExperience string    `gorm:"column:invest_experience"`
	Agent            bool      `gorm:"column:agent"`
	TwoFactorAuth    string    `gorm:"column:two_factor_auth"`
	Field01          string    `gorm:"column:field01"`
	Field02          string    `gorm:"column:field02"`
	Field03          string    `gorm:"column:field03"`
	Field04          string    `gorm:"column:field04"`
	Field05          string    `gorm:"column:field05"`
	Field06          string    `gorm:"column:field06"`
	Field07          string    `gorm:"column:field07"`
	Field08          string    `gorm:"column:field08"`
	Field09          string    `gorm:"column:field09"`
	Field10          string    `gorm:"column:field10"`
	Field11          string    `gorm:"column:field11"`
	Field12          string    `gorm:"column:field12"`
	Field13          string    `gorm:"column:field13"`
	Field14          string    `gorm:"column:field14"`
	Field15          string    `gorm:"column:field15"`
	Field16          string    `gorm:"column:field16"`
	Field17          string    `gorm:"column:field17"`
	Field18          string    `gorm:"column:field18"`
	Field19          string    `gorm:"column:field19"`
	Field20          string    `gorm:"column:field20"`

	//积分
	Points1 string `json:"points1"`
	Points2 string `json:"points2"`
	Points3 string `json:"points3"`
	Points4 string `json:"points4"`
	Points5 string `json:"points5"`
	Points6 string `json:"points6"`
	Points7 string `json:"points7"`
}

func (u UserDetailTab) TableName() string {
	return "t_user_detail"
}

type RegionInfo struct {
	Country  string `json:"country"`
	Province string `json:"province"`
	City     string `json:"city"`
}

type PhoneInfo struct {
	Phone       string `json:"phone"`       //such as "1111111111",
	CountryCode string `json:"countryCode"` //such as "+86"
	PhoneStr    string `json:"phoneStr"`    //such as "+86 1111111111"
}

type IpWhiteBlackDTO struct {
	White  bool   `json:"white"`
	FromIp string `json:"fromIp"`
	ToIp   string `json:"toIp"`
	Enable bool   `json:"enable"`
}

type UserDetail struct {
	Id               int64             `json:"id"`
	EntityNo         string            `json:"entityNo"`
	CreateDate       int64             `json:"createDate"`
	ModifyDate       int64             `json:"modifyDate"`
	TenantId         string            `json:"tenantId"`  //租户ID
	ProduceId        string            `json:"productId"` //产品ID
	CreateUserId     string            `json:"createUserId"`
	ModifyUserId     string            `json:"modifyUserId"`
	ModelStatus      string            `json:"modelStatus"` //constant.ModelStatusCreate/ModelStatusLock/ModelStatusSuccess/ModelStatusDelete
	Name             string            `json:"name"`
	UserName         string            `json:"username"`
	Email            string            `json:"email"`
	Phone            string            `json:"phone"` // such as "+86@-@1111111111",
	Address          string            `json:"address"`
	Country          string            `json:"country"`
	Province         string            `json:"province"`
	City             string            `json:"city"`
	Postcode         string            `json:"postcode"`
	RoleId           int64             `json:"roleId"`
	RoleName         string            `json:"roleName"`
	RoleTypeId       string            `json:"roleTypeId"`
	RoleTypeName     string            `json:"roleTypeName"`
	LevelId          int64             `json:"levelId"`
	LevelName        string            `json:"levelName"`
	Parent           string            `json:"parent"`
	Sex              string            `json:"sex"`
	VendorServerId   string            `json:"vendorServerId"`
	Password         string            `json:"password"`
	Comment          string            `json:"comment"`
	Nickname         string            `json:"nickname"`
	HeadImage        string            `json:"headImage"`
	Birthday         string            `json:"birthday"`
	PubUserId        string            `json:"pubUserId"`
	SubUserCount     int               `json:"subUserCount"`
	Version          int               `json:"version"`
	Active           int               `json:"active"`
	Login            string            `json:"login"`
	NeedInitPass     bool              `json:"needInitPass"`
	Region           RegionInfo        `json:"region"`
	Phones           PhoneInfo         `json:"phones"`
	IpWhiteBlacks    []IpWhiteBlackDTO `json:"ipWhiteBlacks"`
	IdType           string            `json:"idType"`
	IdNum            string            `json:"idNum"`
	IdUrl1           string            `json:"idUrl1"`
	IdUrl2           string            `json:"idUrl2"`
	BankAccount      string            `json:"bankAccount"`
	BankBranch       string            `json:"bankBranch"`
	AccountNo        string            `json:"accountNo"`
	BankCardFile1    string            `json:"bankCardFile1"`
	BankCardFile2    string            `json:"bankCardFile2"`
	DoAgencyBusiness string            `json:"doAgencyBusiness"`
	InvestExperience string            `json:"investExperience"`
	Agent            bool              `json:"agent"`
	TwoFactorAuth    string            `json:"twoFactorAuth"`

	//积分
	Points1 string `json:"points1"`
	Points2 string `json:"points2"`
	Points3 string `json:"points3"`
	Points4 string `json:"points4"`
	Points5 string `json:"points5"`
	Points6 string `json:"points6"`
	Points7 string `json:"points7"`

	Field01 string `json:"field01"`
	Field02 string `json:"field02"`
	Field03 string `json:"field03"`
	Field04 string `json:"field04"`
	Field05 string `json:"field05"`
	Field06 string `json:"field06"`
	Field07 string `json:"field07"`
	Field08 string `json:"field08"`
	Field09 string `json:"field09"`
	Field10 string `json:"field10"`
	Field11 string `json:"field11"`
	Field12 string `json:"field12"`
	Field13 string `json:"field13"`
	Field14 string `json:"field14"`
	Field15 string `json:"field15"`
	Field16 string `json:"field16"`
	Field17 string `json:"field17"`
	Field18 string `json:"field18"`
	Field19 string `json:"field19"`
	Field20 string `json:"field20"`
}

type UserCommissionDetailValueDTO struct {
	Value    string `json:"value"`
	DetailId int    `json:"detailId"`
}

type UserRuleDetailDTO struct {
	Id              int     `json:"id"`
	UserId          int     `json:"userId"`
	RuleType        int     `json:"ruleType"`
	RuleId          int     `json:"ruleId"`
	RuleDetailId    int     `json:"ruleDetailId"`
	CommissionValue float64 `json:"commissionValue"`
}

type UserCommissionSettingDTO struct {
	RuleId             int                            `json:"ruleId"`
	RuleType           *int                           `json:"ruleType"`
	Name               *string                        `json:"name,omitempty"`
	DetailId           int                            `json:"detailId"`
	CommissionValue    *float64                       `json:"commissionValue"`
	MaxCommissionValue *float64                       `json:"maxCommissionValue"`
	Unit               *string                        `json:"unit,omitempty"`
	Values             []UserCommissionDetailValueDTO `json:"values,omitempty"`
	//OldValue             UserRuleDetailDTO               `json:"oldValue"`
}

type UserCommissionUpdateRequestDTO struct {
	List        []UserCommissionSettingDTO `json:"list"`
	LevelId     int64                      `json:"levelId"`
	ParentId    int64                      `json:"parentId"`
	OldParentId int64                      `json:"oldParentId"`
	UserId      int64                      `json:"userId"`
}

//全量用户数据
type BWUserDTO struct {
	Id               int64                           `json:"id"`
	Name             string                          `json:"name"`
	UserName         string                          `json:"username"`
	Email            string                          `json:"email"`
	Phone            string                          `json:"phone"` // such as "+86@-@1111111111"
	EntityNo         string                          `json:"entityNo"`
	RoleId           string                          `json:"roleId"`
	RoleName         string                          `json:"roleName"`
	LevelId          string                          `json:"levelId"`
	LevelName        string                          `json:"levelName"`
	Parent           string                          `json:"parent"`
	ParentName       string                          `json:"parentName"`
	Password         string                          `json:"password"`
	Comment          string                          `json:"comment"`
	Nickname         string                          `json:"nickname"`
	HeadImage        string                          `json:"headImage"`
	Birthday         string                          `json:"birthday"`
	Address          string                          `json:"address"`
	TenantId         string                          `json:"tenantId"` //租户ID
	PubUserId        string                          `json:"pubUserId"`
	Country          string                          `json:"country"`
	Province         string                          `json:"province"`
	City             string                          `json:"city"`
	Postcode         string                          `json:"postcode"`
	CreateDate       int64                           `json:"createDate"`
	ModifyDate       int64                           `json:"modifyDate"`
	Sex              string                          `json:"sex"`
	SubUserCount     int                             `json:"subUserCount"`
	Active           int                             `json:"active"`
	Login            *string                         `json:"login"`
	VendorServerId   *string                         `json:"vendorServerId"`
	SendEmail        bool                            `json:"sendEmail"`
	NeedInitPass     bool                            `json:"needInitPass"`
	Commission       *UserCommissionUpdateRequestDTO `json:"commission"`
	IpWhiteBlacks    []IpWhiteBlackDTO               `json:"ipWhiteBlacks"`
	Region           RegionInfo                      `json:"region"`
	Phones           PhoneInfo                       `json:"phones"`
	Lang             string                          `json:"lang"`
	OwnerType        string                          `json:"ownerType,omitempty"`
	MessageCode      string                          `json:"messageCode"`
	IdType           string                          `json:"idType"`
	IdNum            string                          `json:"idNum"`
	IdUrl1           string                          `json:"idUrl1"`
	IdUrl2           string                          `json:"idUrl2"`
	BankAccount      string                          `json:"bankAccount"`
	BankBranch       string                          `json:"bankBranch"`
	AccountNo        string                          `json:"accountNo"`
	BankCardFile1    string                          `json:"bankCardFile1"`
	BankCardFile2    string                          `json:"bankCardFile2"`
	DoAgencyBusiness string                          `json:"doAgencyBusiness"`
	InvestExperience string                          `json:"investExperience"`
	Agent            bool                            `json:"agent"`
	TwoFactorAuth    string                          `json:"twoFactorAuth"`

	//积分
	Points1 string `json:"points1"`
	Points2 string `json:"points2"`
	Points3 string `json:"points3"`
	Points4 string `json:"points4"`
	Points5 string `json:"points5"`
	Points6 string `json:"points6"`
	Points7 string `json:"points7"`

	Field01 string `json:"field01"`
	Field02 string `json:"field02"`
	Field03 string `json:"field03"`
	Field04 string `json:"field04"`
	Field05 string `json:"field05"`
	Field06 string `json:"field06"`
	Field07 string `json:"field07"`
	Field08 string `json:"field08"`
	Field09 string `json:"field09"`
	Field10 string `json:"field10"`
	Field11 string `json:"field11"`
	Field12 string `json:"field12"`
	Field13 string `json:"field13"`
	Field14 string `json:"field14"`
	Field15 string `json:"field15"`
	Field16 string `json:"field16"`
	Field17 string `json:"field17"`
	Field18 string `json:"field18"`
	Field19 string `json:"field19"`
	Field20 string `json:"field20"`
	ForTask bool   `json:"forTask"`
}

type BWUserDTOIncrease struct {
	Id               int64                           `json:"id"`
	Name             *string                         `json:"name"`
	UserName         *string                         `json:"username"`
	Email            *string                         `json:"email"`
	Phone            *string                         `json:"phone"` // such as "+86@-@1111111111"
	EntityNo         *string                         `json:"entityNo"`
	RoleId           *string                         `json:"roleId"`
	RoleName         *string                         `json:"roleName"`
	LevelId          *string                         `json:"levelId"`
	LevelName        *string                         `json:"levelName"`
	Parent           *string                         `json:"parent"`
	ParentName       *string                         `json:"parentName"`
	Password         *string                         `json:"password"`
	Comment          *string                         `json:"comment"`
	Nickname         *string                         `json:"nickname"`
	HeadImage        *string                         `json:"headImage"`
	Birthday         *string                         `json:"birthday"`
	Address          *string                         `json:"address"`
	TenantId         *string                         `json:"tenantId"` //租户ID
	PubUserId        *string                         `json:"pubUserId"`
	Country          *string                         `json:"country"`
	Province         *string                         `json:"province"`
	City             *string                         `json:"city"`
	Postcode         *string                         `json:"postcode"`
	Sex              *string                         `json:"sex"`
	Login            *string                         `json:"login"`
	VendorServerId   *string                         `json:"vendorServerId"`
	CreateDate       int64                           `json:"createDate"`
	ModifyDate       int64                           `json:"modifyDate"`
	SubUserCount     int                             `json:"subUserCount"`
	Active           int                             `json:"active"`
	SendEmail        bool                            `json:"sendEmail"`
	NeedInitPass     bool                            `json:"needInitPass"`
	Commission       *UserCommissionUpdateRequestDTO `json:"commission"`
	IpWhiteBlacks    []IpWhiteBlackDTO               `json:"ipWhiteBlacks"`
	Region           *RegionInfo                     `json:"region"`
	Phones           *PhoneInfo                      `json:"phones"`
	Lang             *string                         `json:"lang"`
	OwnerType        *string                         `json:"ownerType,omitempty"`
	MessageCode      *string                         `json:"messageCode"`
	IdType           *string                         `json:"idType"`
	IdNum            *string                         `json:"idNum"`
	IdUrl1           *string                         `json:"idUrl1"`
	IdUrl2           *string                         `json:"idUrl2"`
	BankAccount      *string                         `json:"bankAccount"`
	BankBranch       *string                         `json:"bankBranch"`
	AccountNo        *string                         `json:"accountNo"`
	BankCardFile1    *string                         `json:"bankCardFile1"`
	BankCardFile2    *string                         `json:"bankCardFile2"`
	DoAgencyBusiness *string                         `json:"doAgencyBusiness"`
	InvestExperience *string                         `json:"investExperience"`
	Agent            bool                            `json:"agent"`
	TwoFactorAuth    string                          `json:"twoFactorAuth"`
	Field01          *string                         `json:"field01"`
	Field02          *string                         `json:"field02"`
	Field03          *string                         `json:"field03"`
	Field04          *string                         `json:"field04"`
	Field05          *string                         `json:"field05"`
	Field06          *string                         `json:"field06"`
	Field07          *string                         `json:"field07"`
	Field08          *string                         `json:"field08"`
	Field09          *string                         `json:"field09"`
	Field10          *string                         `json:"field10"`
	Field11          *string                         `json:"field11"`
	Field12          *string                         `json:"field12"`
	Field13          *string                         `json:"field13"`
	Field14          *string                         `json:"field14"`
	Field15          *string                         `json:"field15"`
	Field16          *string                         `json:"field16"`
	Field17          *string                         `json:"field17"`
	Field18          *string                         `json:"field18"`
	Field19          *string                         `json:"field19"`
	Field20          *string                         `json:"field20"`
	//积分
	Points1 *string `json:"points1"`
	Points2 *string `json:"points2"`
	Points3 *string `json:"points3"`
	Points4 *string `json:"points4"`
	Points5 *string `json:"points5"`
	Points6 *string `json:"points6"`
	Points7 *string `json:"points7"`
}

type IdNameDTO struct {
	Id       int64  `gorm:"column:id" gorm:"primary_key" json:"id"`
	Name     string `gorm:"column:name" json:"name"`
	EntityNo string `gorm:"column:entity_no" json:"entityNo"`
}

type SimpleUserDTO struct {
	Id             int64  `gorm:"column:id" gorm:"primary_key" json:"id"`
	Name           string `gorm:"column:name" json:"name"`
	ParentId       string `gorm:"column:parent_id" json:"parent"`
	LevelId        int64  `gorm:"column:level_id" gorm:"not null" json:"levelId"`
	LevelName      string `gorm:"column:level_name" json:"levelName"`
	RoleId         int64  `gorm:"column:role_id" gorm:"not null" json:"roleId"`
	RoleName       string `gorm:"column:role_name" json:"roleName"`
	EntityNo       string `gorm:"column:entity_no" json:"entityNo"`
	PubUserId      string `gorm:"column:pub_user_id" json:"pubUserId"`
	Login          string `gorm:"column:login" json:"login"`
	VendorServerId string `gorm:"column:vendor_server_id" json:"vendorServerId"`
	TwoFactorAuth  string `gorm:"column:two_factor_auth" json:"twoFactorAuth"`
}

type LazyTreeNodeDTO struct {
	Label    string             `json:"label"`    //名称
	Value    string             `json:"value"`    //ID
	Child    bool               `json:"child"`    //是否有下级
	Selected bool               `json:"selected"` //选中状态
	Parent   string             `json:"parent"`   //父ID
	Children []*LazyTreeNodeDTO `json:"children"`
}

type LazyTreeNodeDTOSlice []*LazyTreeNodeDTO

func (a LazyTreeNodeDTOSlice) Len() int { // 重写 Len() 方法
	return len(a)
}
func (a LazyTreeNodeDTOSlice) Swap(i, j int) { // 重写 Swap() 方法
	a[i], a[j] = a[j], a[i]
}
func (a LazyTreeNodeDTOSlice) Less(i, j int) bool { // 重写 Less() 方法， 根据sid从到大排序
	return strings.ToUpper(a[i].Label) < strings.ToUpper(a[j].Label)
}

type UserSearchDTO struct {
	TenantId       string `json:"tenantId"`
	StartDate      int64  `json:"startDate"`
	EndDate        int64  `json:"endDate"`
	PageNo         int    `json:"pageNo"`
	Size           int    `json:"size"`
	Sortby         string `json:"sortby"`
	OrderDesc      bool   `json:"orderDesc"`
	LevelId        int64  `json:"levelId"`
	QueryType      string `json:"queryType,omitempty"` //QueryType类型
	QueryContent   string `json:"queryContent"`
	UserSearchType string `json:"userSearchType"`
	UserId         string `json:"userId"`

	AdvanceConditions []*AdvanceCondition `json:"advanceConditions"`
}

type SearchDTO struct {
	Conditions []*AdvanceCondition `json:"condition"`
	Fields     []string            `json:"resultFields"`
}

type AdvanceCondition struct {
	Condition string `json:"condition"`
	Field     string `json:"field"`
	Value     string `json:"value"`
}

type SimpleUserPageDTO struct {
	Pager  int             `json:"pager"`
	Pages  int             `json:"pages"`
	Size   int             `json:"size"`
	Total  int64           `json:"total"`
	Offset int             `json:"offset"`
	List   []SimpleUserDTO `json:"list"`
}

type UserDetailPageDTO struct {
	Pager  int          `json:"pager"`
	Pages  int          `json:"pages"`
	Size   int          `json:"size"`
	Total  int64        `json:"total"`
	Offset int          `json:"offset"`
	List   []UserDetail `json:"list"`
}

type UserRoleDTO struct {
	PubUserIds []string `json:"pubUserIds"`
	Roles      []string `json:"roles"`
}

type FuzzyConditionDTO struct {
	FuzzyValue string `json:"fuzzyValue"`
}

type MsgReceiversSearchDTO struct {
	TenantId     string `json:"tenantId"`
	FuzzyVal     string `json:"fuzzyVal"`
	ReceiverType string `json:"receiverType"`
	Type         string `json:"type,omitempty"` //MessageType类型
}

type MsgReceiversDTO struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	IdType    string `json:"idType,omitempty"` //IdType
	RoleName  string `json:"roleName"`
	LevelName string `json:"levelName"`
	EntityNo  string `json:"entityNo"`
}

type UserFieldSearchDTO struct {
	FuzzyValue string   `json:"fuzzyValue"`
	FieldTypes []string `json:"fieldType"`
}

type UserNameSearchDTO struct {
	FuzzyValue  string `json:"fuzzyValue"`
	SearchUser  bool   `json:"searchUser"`
	SearchRole  bool   `json:"searchRole"`
	SearchLevel bool   `json:"searchLevel"`
}

//mongo DB,t_user_fields
type UserField struct {
	Key   string `bson:"key"`
	Label string `bson:"label"`
	Show  bool   `bson:"show"`
}
type UserFields struct {
	FieldsId     string      `bson:"_id"`
	TenantId     string      `bson:"tenantId"`
	TableName    string      `bson:"tableName"`
	UserFields   []UserField `bson:"userFields"`
	Enabled      bool        `bson:"enabled"`
	CreateUserId string      `bson:"createUserId"`
	ModifyUserId string      `bson:"modifyUserId"`
	CreateTime   int64       `bson:"createTime"`
	ModifyTime   int64       `bson:"modifyTime"`
}

type UserFieldsDTO struct {
	TableName  string      `json:"tableName"`
	UserFields []UserField `json:"userFields"`
}

type UserRoleStat struct {
	UserCount int64 `gorm:"column:user_count" json:"userCount"`
	RoleId    int64 `gorm:"column:role_id" json:"roleId"`
}

type UserKeyDTO struct {
	KeyType  string   `json:"keyType"`
	KeyValue []string `json:"keyValue"`
}

type UserIds struct {
	UserId int64  `json:"userId"`
	Email  string `json:"email"`
}
