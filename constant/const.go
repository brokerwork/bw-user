package constant

const (
	// 本品产品命名
	ProductId = "BW"
)

// 定义常用Http常用传参字段
const (
	// HeaderApiTenantId 租户传递租户ID的头部字段名称
	HeaderApiTenantId = "X-Api-Tenantid"
	// HeaderApiServerId 租户传递MT4服务器ID的头部字段名称
	HeaderApiServerId = "X-Api-Serverid"
	// HeaderApiToken 租户传递校验令牌的头部字段名称
	HeaderApiToken = "X-Api-Token"
	// HeaderApiAccountToken 交易账户的令牌的头部字段
	HeaderApiAccountToken = "X-Api-Account-Token"
	// QueryApiAccount 交易账户的ID查询字段
	//QueryAccount = "account"
	//HeaderAPIXFeignKey 请求的用户信息
	HeaderAPIXFeigKey = "X-Feign-Key"
)

const (
	PwdStrengthMiddle      = 0
	PwdStrengthStrong      = 1
	PwdStrengthSuperStrong = 2

	LOOP_MAX_DEPTH = 200

	MIN_TIMESTAMP   = 60          //1970/1/1 8:1:0
	MAX_TIMESTAMP   = 32503564800 //2999/12/31 00:00:00
	MAX_COMM_LENGTH = 250
)

type ParameterType int32

const (
	Header = 1
	Route  = 2
	Query  = 3
	Json   = 4
)

func (pt ParameterType) String() string {
	switch pt {
	case Header:
		return "header"
	case Route:
		return "route"
	case Query:
		return "query"
	case Json:
		return "json body"
	}

	return ""
}

const (
	ModelStatusCreate  = "CREATE"
	ModelStatusLock    = "LOCK"
	ModelStatusSuccess = "SUCCESS"
	ModelStatusDelete  = "DELETE"
)

const (
	ServiceNamePubAuth = "pub-auth"
	PathPubAuthAddUser = "/v1/pub/user/add"
	PathPubTrashUser   = "/v1/pub/user/trash" //管理员删除用户（软删除，移至回收站）
	PathPubUserSet     = "/v1/pub/user/set"
	PathPubUserNumber  = "/v1/pub/user/number"

	ServiceNameBwReport     = "bw-report-service"
	PathBwReportAddOrUpdate = "/v1/report/user/addOrUpdate"              //用户佣金规则详细-添加修改
	PathBwReportUserDelete  = "/v1/report/user/delete/{userId}"          //用户佣金规则详细-删除
	PathIsLevelNotUsed      = "/v1/report/setting/isLevelUsed/{levelId}" //层级是否不在使用用

	ServiceNameBwMessage                = "bw-message"
	PathBwMessageSendMailByTemplateType = "/api/v1/message/sendMailByTemplateType"

	ServiceNameTenant     = "tenants-org"
	PathGetTenantInfo     = "/v1/ops/tenants/{tenantId}"
	PathGetGenantOption   = "/v1/ops/tenants/metadata/field/option/{fieldId}"
	PathTenantFieldSimple = "/v1/ops/tenants/metadata/form-field/simple"
	PathTenantField       = "/v1/ops/tenants/metadata/form-field/list"

	ServiceNameAccountFeign   = "bw-account"
	PathCheckAccountExist     = "/v1/account/manage/checkAccountExist/{login}"
	PathStatisticDeposit      = "/v2/account/statistic/deposit"
	PathGetCustomerAccountMap = "/v2/account/statistic/accountMap"
	PathAccountBaseInfo       = "/v1/account/manage/baseInfo/byCustomers"
	PathCheckUserAsOwner      = "/v1/account/manage/checkUserAsOwner"

	ServiceNameProducts   = "products-org"
	PathTenantByKey       = "/v1/ops/product/tp-key"
	PathProducePermission = "/v1/ops/product/permission"
	PathGetProductDeploy  = "/v1/ops/product/deploy/mysql"
	PathProductUserLimit  = "/v1/ops/product/user/limit"

	ServiceNameTaskJob = "tasks-job"
	PathTaskAgencyList = "/v1/tasks/stats/agency/list"

	ServiceNameCustom         = "bw-custom"
	PathLincCustom            = "/v2/custom/profiles/link/{customSource}"
	PathCustomerOpportunities = "/v1/custom/opportunity/{salesStage}/batch/list"
	PathCustomQuery           = "/v2/custom/query/list"

	ServiceNamePic = "bw-pic"
)

const (
	QueryKeyId           = "id"
	QueryKeyEmail        = "email"
	QueryKeyPhone        = "phone"
	QueryKeyEntityNo     = "entityNo"
	QueryKeyName         = "name"
	QueryKeyLogin        = "login"
	QueryKeyRoleId       = "roleId"
	QueryKeyLevelId      = "levelId"
	QueryKeypubUserIds   = "pubUserIds"
	QueryKeypubUserIdsIB = "pubUserIdsIB"
	QueryKeyIdNum        = "idNum"
	QueryKeyAllRecords   = "allRecords"
)

const (
	QueryType_ENTITY_NO = "ENTITY_NO"
	QueryType_ROLE      = "ROLE"
	QueryType_NAME      = "NAME"
	QueryType_PHONE     = "PHONE"
	QueryType_EMAIL     = "EMAIL"
	QueryType_LEVEL     = "LEVEL"
	QueryType_ID        = "ID"
	QueryType_PARENT    = "PARENT"
	QueryType_LOGIN     = "LOGIN"
)

const (
	ModuleUser             = "User"
	ModuleCustomer         = "Customer"
	ModuleAccount          = "Account"
	ModuleAccountReport    = "AccountReport"
	ModuleCommissionReport = "CommissionReport"
	ModuleEarningReport    = "EarningReport"
)

//用户权限
const (
	AUTHORITY_ACCOUNT_OPEN            = "ACCOUNT_OPEN"            //开户
	AUTHORITY_ACCOUNT_MODIFY_PERSONAL = "ACCOUNT_MODIFY-PERSONAL" //修改账户所有人
	AUTHORITY_ACCOUNT_MODIFY_LEVER    = "ACCOUNT_MODIFY-LEVER"    //修改杠杆
	AUTHORITY_ACCOUNT_MODIFY_PWD      = "ACCOUNT_MODIFY-PWD"      //修改密码
	AUTHORITY_ACCOUNT_MODIFY_DW       = "ACCOUNT_MODIFY-DW"       //出入金
	AUTHORITY_ACCOUNT_MODIFY_CREDIT   = "ACCOUNT_MODIFY-CREDIT"   //修改信用
	AUTHORITY_ACCOUNT_DELETE          = "ACCOUNT_DELETE"          //删除账户
	AUTHORITY_ACCOUNT_MODIFY_DATA     = "ACCOUNT_MODIFY-DATA"     //修改账户信息
	AUTHORITY_ACCOUNT_ADDCUSTOMER     = "ACCOUNT_ADDCUSTOMER"     //创建客户

	//查询账户权限
	AUTHORITY_ACCOUNT_SELECT             = "ACCOUNT_SELECT"             //查看账户，下面的父级
	AUTHORITY_ACCOUNT_SELECT_DIRECTLY    = "ACCOUNT_SELECT_DIRECTLY"    //查看归属给我的账户
	AUTHORITY_ACCOUNT_SELECT_SUBORDINATE = "ACCOUNT_SELECT_SUBORDINATE" //查看归属给下级的账户
	AUTHORITY_ACCOUNT_SELECT_ALL         = "ACCOUNT_SELECT_ALL"         //查看平台所有账户
	AUTHORITY_ACCOUNT_SELECT_WILD        = "ACCOUNT_SELECT_WILD"        //查看无归属账户

	//用户权限
	AUTHORITY_USER_SELECT_DIRECTLY    = "USER_SELECT_DIRECTLY"    //查看归属给我的账户
	AUTHORITY_USER_SELECT_SUBORDINATE = "USER_SELECT_SUBORDINATE" //查看归属给下级的账户
	AUTHORITY_USER_SELECT_ALL         = "USER_SELECT_ALL"         //查看平台所有账户
	AUTHORITY_USER_SELECT_WILD        = "USER_SELECT_WILD"        //查看无归属账户

	//账户报表权限
	AUTHORITY_ACCOUNT_REPORT_SELECT_MY   = "STAT_VIEW_ACC_RANGE_MY"  //查看归属给我的账户
	AUTHORITY_ACCOUNT_REPORT_SELECT_SUB  = "STAT_VIEW_ACC_RANGE_SUB" //查看归属给下级的账户
	AUTHORITY_ACCOUNT_REPORT_SELECT_ALL  = "STAT_VIEW_ACC_RANGE_ALL" //查看平台所有账户
	AUTHORITY_ACCOUNT_REPORT_SELECT_WILD = "STAT_VIEW_ACC_RANGE_NO"  //查看无归属账户

	//佣金报表
	AUTHORITY_COMMISSION_REPORT_SELECT_SUB = "STAT_VIEW_COMMISSION_RANGE_SUB" //查看下级的返佣
	AUTHORITY_COMMISSION_REPORT_SELECT_ALL = "STAT_VIEW_COMMISSION_RANGE_ALL" //查看所有的返佣

	//业绩报表
	AUTHORITY_EARNING_REPORT_SELECT_DIRECT   = "STAT_VIEW_ACHIEVEMENT_RANGE_DIRECTLY" //查看直属
	AUTHORITY_EARNING_REPORT_SELECT_INDIRECT = "STAT_VIEW_ACHIEVEMENT_RANGE_SUB"      //查看非直属
	AUTHORITY_EARNING_REPORT_SELECT_ALL      = "STAT_VIEW_ACHIEVEMENT_RANGE_ALL"      //查看所有用户
	AUTHORITY_EARNING_REPORT_SELECT_WILD     = "STAT_VIEW_ACHIEVEMENT_RANGE_SUPERIOR" //查看无归属

	//客户
	AUTHORITY_CUSTOMER_SELECT_SELECT_MY = "CUSTOMER_SELECT_DIRECTLY"    //归属给我的客户
	AUTHORITY_CUSTOMER_SELECT_SUB       = "CUSTOMER_SELECT_SUBORDINATE" //归属给我下属的客户
	AUTHORITY_CUSTOMER_SELECT_ALL       = "CUSTOMER_SELECT_ALL"         //平台所有客户
	AUTHORITY_CUSTOMER_SELECT_WILD      = "CUSTOMER_SELECT_WILD"        //查看无归属
)

const (
	NORMALIZED_RIGHT_ALL         = 1
	NORMALIZED_RIGHT_NO_PARENT   = 2
	NORMALIZED_RIGHT_MY          = 3
	NORMALIZED_RIGHT_DIRECT      = 4
	NORMALIZED_RIGHT_SUBORDINATE = 5
)

const (
	SearchUserScope_ALL             = "ALL"
	SearchUserScope_BELONG          = "BELONG"
	SearchUserScope_NOT_BELONG      = "NOT_BELONG"      //所有的非直属
	SearchUserScope_ONLY_NOT_BELONG = "ONLY_NOT_BELONG" //仅一级非直属
)

const (
	SEARCH_TYPE_ROLE  = 0
	SEARCH_TYPE_LEVEL = 1
)

const (
	CommissionMode_MULTI_AGENT  = "MULTI_AGENT"
	CommissionMode_DISTRIBUTION = "DISTRIBUTION"
)

const (
	IntroduceType_uid    = "uid"
	IntroduceType_pid    = "pid"
	IntroduceType_eid    = "eid"
	IntroduceType_Web    = "Web"
	IntroduceType_Mobile = "Mobile"

	IntroduceType_UserAllVisible  = "UserAllVisible"  //所有用户可见
	IntroduceType_UserPartVisible = "UserPartVisible" //部分用户可见
	IntroduceType_UserInVisible   = "UserInVisible"   //部分用户不可见
	IntroduceType_UserNotVisible  = "UserNotVisible"  //所有用户不可见

	IntroduceType_DirectAllVisible    = "DirectAllVisible"    //所有直客可见
	IntroduceType_DirectPartVisible   = "DirectPartVisible"   //部分直客可见
	IntroduceType_DirectPartInvisible = "DirectPartInvisible" //部分直客不可见
	IntroduceType_DirectNotVisible    = "DirectNotVisible"    //所有直客不可见

	IntroduceType_StraightGuest        = "StraightGuest"        //直客
	IntroduceType_Agent                = "Agent"                //代理
	IntroduceType_DirectRecommendation = "DirectRecommendation" //直客推荐直客

	IntroduceType_NoOwner       = "NoOwner"       //无归属
	IntroduceType_OtherOwner    = "OtherOwner"    //其他
	IntroduceType_CustomerOwner = "CustomerOwner" //链接所属客户归属
)

const (
	LogicType_AND = "AND"
	LogicType_OR  = "OR"
)

const (
	Vendor_SAXO    = "SAXO"
	Vendor_PingAn  = "PingAn"
	Vendor_MT4     = "MT4"
	Vendor_MT5     = "MT5"
	Vendor_Esunny  = "Esunny"
	Vendor_SAXOFIX = "SAXOFIX"
	Vendor_IGFIX   = "IGFIX"
	Vendor_LMAXFIX = "LMAXFIX"
	Vendor_CTRADER = "CTRADER"
)

const (
	DEFAULT_TW_INTRODUCE_KEY    = "DEFAULT_TW_INTRODUCE"
	DEFAULT_AGENT_INTRODUCE_KEY = "DEFAULT_AGENT_INTRODUCE"
	Language_CN                 = "zh-CN"
	Language_EN                 = "en-US"
	Language_TW                 = "zh-TW"
	Language_HK                 = "zh-HK"
	Language_JP                 = "ja-JP"
)

const (
	SearchLevel_TENANT = "TENANT"
	SearchLevel_USER   = "USER"
)

const (
	IdType_Id             = "Id"
	IdType_RoleId         = "RoleId"
	IdType_LevelId        = "LevelId"
	IdType_UserGroupId    = "UserGroupId"
	IdType_AccountGroupId = "AccountGroupId"
)

const (
	ReceiverType_BwUser       = "BwUser"
	ReceiverType_Account      = "Account"
	ReceiverType_TwUser       = "TwUser"
	ReceiverType_BwCustomer   = "BwCustomer"
	ReceiverType_MyBwCustomer = "MyBwCustomer"
	ReceiverType_BWUser_User  = "BWUser_User"
)

const (
	OwnerType_RoleId = "RoleId"
	OwnerType_Id     = "Id"
)

const (
	TaskState_Submited = "Submited"
	TaskState_Rejected = "Rejected"
	TaskState_Refused  = "Refused"
	TaskState_Dealed   = "Dealed"
	TaskState_Finished = "Finished"
	TaskState_Closed   = "Closed"
	TaskState_Handling = "Handling"
)

const (
	OwnerType_all         = "all"
	OwnerType_noParent    = "noParent"
	OwnerType_sub         = "sub"
	OwnerType_subBelong   = "subBelong"
	OwnerType_participant = "participant"
)

const (
	IntroduceStatistic_HIT                  = "HIT"
	IntroduceStatistic_CUSTOMER             = "CUSTOMER"
	IntroduceStatistic_CUSTOMER_HAS_ACCOUNT = "CUSTOMER_HAS_ACCOUNT"
	IntroduceStatistic_CUSTOMER_HAS_DEPOSIT = "CUSTOMER_HAS_DEPOSIT"
	//下面2个为LWORK专用
	IntroduceStatistic_CUSTOMER_WIN    = "CUSTOMER_WIN"
	IntroduceStatistic_CONTRACT_AMOUNT = "CONTRACT_AMOUNT"
	//下面3个为代理专用
	IntroduceStatistic_APPLY          = "APPLY"
	IntroduceStatistic_APPLY_PASS     = "APPLY_PASS"
	IntroduceStatistic_APPLY_NOT_PASS = "APPLY_NOT_PASS"
)

const (
	MessageType_ALL          = "ALL"
	MessageType_MAIL         = "MAIL"
	MessageType_WEB          = "WEB"
	MessageType_WEB_ALERT    = "WEB_ALERT"
	MessageType_WEB_ANNOUNCE = "WEB_ANNOUNCE"
	MessageType_SMS          = "SMS"
)

const (
	BwEvent_ADD               = "ADD"
	BwEvent_UPDATE            = "UPDATE"
	BwEvent_DELETE            = "DELETE"
	BwEvent_DEPOSITE          = "DEPOSITE"
	BwEvent_WITHDRAWAL        = "WITHDRAWAL"
	BwEvent_OPEN_ACCOUNT      = "OPEN_ACCOUNT"
	BwEvent_OPEN_SAME_ACCOUNT = "OPEN_SAME_ACCOUNT"
)

const (
	SearchCondition_EQ      = "EQ"
	SearchCondition_NEQ     = "NEQ"
	SearchCondition_REGEX   = "REGEX"
	SearchCondition_CONTAIN = "CONTAIN"
	SearchCondition_IN      = "IN"
	SearchCondition_NIN     = "NIN"
	SearchCondition_BETWEEN = "BETWEEN"
	SearchCondition_GT      = "GT"
	SearchCondition_LT      = "LT"
	SearchCondition_EMPTY   = "EMPTY"

	SearchField_Id             = "id"
	SearchField_Login          = "login"
	SearchField_Name           = "name"
	SearchField_EntityNo       = "entityNo"
	SearchField_NameOrEntityNo = "nameOrEntityNo"
	SearchField_LevelId        = "levelId"
	SearchField_RoleName       = "roleName"
	SearchField_ExcludeIbRole  = "excludeIbRole"
	SearchField_ParentName     = "parentName"
	SearchField_Email          = "email"
	SearchField_Phone          = "phones"
	SearchField_Right          = "right"
)

var DefaultFieldKeys = []string{"id", "subUserCount", "createDate", "active", "ownAccounts", "ownCustomers",
	"balance", "profit", "equity", "margin", "marginFree", "marginLevel", "credit",
	"accounts", "customerState", "recommendedCustomerNum", "openState", "dealState"}

var AES_KEY = "22222highlow2222"
