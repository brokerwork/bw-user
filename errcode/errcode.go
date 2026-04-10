package errcode

import (
	//"fmt"
	"bw/bw-user/constant"

	"github.com/lworkltd/kits/service/restful/code"
)

const (
	// 读取缓存失败
	ConnectMySQLFailed              = 1001
	ConnectMongoFailed              = 1002
	ParameterError                  = 1003 // 函数传参错误
	ExecuteSQLError                 = 1004
	OperateMongoFailed              = 1005
	InternalError                   = 1006
	GeneratEntityNoError            = 1007
	EmailHaveExist                  = 1008
	RoleIDNotExist                  = 1009
	LevelIDNotExist                 = 1010
	QueryKeyNotSupport              = 1011
	ForbidDelete                    = 1012
	NameHaveExist                   = 1013
	CheckCycleNotPass               = 1014
	NotRightInfo                    = 1015
	SidIllegal                      = 1016
	UserNotRight                    = 1017
	LoginCheckNotPass               = 1018
	EntityHaveExist                 = 1019
	ParentLevelNotFit               = 1020
	ProductInfoAbnormal             = 1021
	LoginHaveExist                  = 1022
	AdvanceSearchNotExist           = 1023
	IntroduceNotExist               = 1024
	GenerateQrCodeFailed            = 1025
	AliyunAccessFailed              = 1026
	UserLevelNotEqual               = 1027
	RequestTaskAgencyListFailed     = 1028
	RequestCustomFeignFailed        = 1029
	ReferenceCustomerNotAllowDelete = 1030
	ReferenceAccountNotAllowDelete  = 1031
	RoleIDExistUser                 = 1032
	RoleIDExistSubRole              = 1033
	DeleteLevelExistUser            = 1034
	LevelIsReportUsed               = 1035 //结算层级不可删除
	ForbidDeleteWithSubUser         = 1036 //有下级用户禁止删除

	GetDashboardConfigFail = 1037

	ProductDomainNotSet = 1038

	Forbiden = 1039

	UploadPicErr   = 1040
	RightNotExists = 1041
	UserHaveExist  = 1042

	// 缺少必要消息参数
	ApiLackParameter = 1130
	OperateForbidden = 1131

	// 参数校验校验错误
	// 格式，类型，值域
	ApiBadJsonPayload = 1400
	// 条件不可达
	AccountNotFound = 1203
	UserNotFound    = 1204
	RoleNotFound    = 1205
	LevelNotFound   = 1206

	//请求其他服务异常
	RequestPubAuthFailed      = 1300
	RequestMailSerFailed      = 1301
	RequestCommissionFailed   = 1302
	RequestTenantFeignFailed  = 1303
	RequestAccountFeignFailed = 1304
	RequestProductFeignFailed = 1305
	ApiBadDatetime            = 1306
)

var (
	CerrAccountNotFound = code.New(AccountNotFound, "account not found")
	CerrUserNotFound    = code.New(UserNotFound, "user not found")
	CerrRoleNotFound    = code.New(RoleNotFound, "role not found")
	CerrLevelNotFound   = code.New(LevelNotFound, "level not found")

	CerrConnMySQL          = code.New(ConnectMySQLFailed, "Failed to connect mysql server")
	CerrConnMongo          = code.New(ConnectMongoFailed, "Failed to connect mongo server")
	CerrParamater          = code.New(ParameterError, "Parameter Error")
	CerrExecuteSQL         = code.New(ExecuteSQLError, "ExecuteSQLError")
	CerrOperateMongo       = code.New(OperateMongoFailed, "Operate Mongo failed")
	CerrInternal           = code.New(InternalError, "Internal Error")
	CerrEntityNoExist      = code.New(GeneratEntityNoError, "Generate EntityNo Error")
	CerrEmailExist         = code.New(EmailHaveExist, "Email have exist")
	CerrRoleIDNotExist     = code.New(RoleIDNotExist, "Role ID not exist")
	CerrLevelIDNotExist    = code.New(LevelIDNotExist, "Level ID not exist")
	CerrQueryKeyNotSupport = code.New(QueryKeyNotSupport, "Queyr Key Not Support")

	CerrRequestPubAuth                  = code.New(RequestPubAuthFailed, "Request PubAuth Failed")
	CerrReuestMailSer                   = code.New(RequestMailSerFailed, "Request Mail Server Failed")
	CerrRequestCommissionSer            = code.New(RequestCommissionFailed, "Request Commission Server Failed")
	CerrForbidDelete                    = code.New(ForbidDelete, "Forbid Delete")
	CerrNameExist                       = code.New(NameHaveExist, "Name have Exist")
	CerrCheckCycleNotPass               = code.New(CheckCycleNotPass, "Check Cycle Not Pass")
	CerrSidIllegal                      = code.New(SidIllegal, " Sid is illegal")
	CerrUserNotRight                    = code.New(UserNotRight, "User Not Right")
	CerrRequestTenantFeign              = code.New(RequestTenantFeignFailed, "Request Tenant Feign Failed")
	CerrRequestAccountFeign             = code.New(RequestAccountFeignFailed, "Request Account Feign Failed")
	CerrLoginCheckNotPass               = code.New(LoginCheckNotPass, "Login Check Not Pass")
	CerrOperateForbidden                = code.New(OperateForbidden, "Operate Forbidden")
	CerreEntityExist                    = code.New(EntityHaveExist, "EntityNo have exist")
	CerrParentLevelNotFit               = code.New(ParentLevelNotFit, "Parent Level Not Fit")
	CerrRequestProductFeign             = code.New(RequestProductFeignFailed, "Request Product Feign Failed")
	CerrProductInfoAbnormal             = code.New(ProductInfoAbnormal, "Product Info Abnormal")
	CerrProductDomainNotSet             = code.New(ProductDomainNotSet, "Product Domain is not config")
	CerrAdvanceSearchNotExist           = code.New(AdvanceSearchNotExist, "Advance Search Not Exist")
	CerrIntroduceNotExist               = code.New(IntroduceNotExist, "Introduce Not Exist")
	CerrGenerateQrCode                  = code.New(GenerateQrCodeFailed, "Generate QrCode failed")
	CerrAliyunAccess                    = code.New(AliyunAccessFailed, "Aliyun Access failed")
	CerrUserLevelNotEqual               = code.New(UserLevelNotEqual, "User Level Not Equal")
	CerrRequestTaskAgencyList           = code.New(RequestTaskAgencyListFailed, "Request Task Agency List Failed")
	CerrRequestCustomFeign              = code.New(RequestCustomFeignFailed, "Request Custom Feign Failed")
	CerrRoleIDExistUser                 = code.New(RoleIDExistUser, "RoleID Exist User")
	CerrRoleIDExistSubRole              = code.New(RoleIDExistSubRole, "RoleID Exist Sub Roler")
	CerrReferenceCustomerNotAllowDelete = code.New(ReferenceCustomerNotAllowDelete, "ReferenceOrBindCustomerNotAllowDelete")
	CerrReferenceAccountNotAllowDelete  = code.New(ReferenceAccountNotAllowDelete, "ReferenceAccountNotAllowDelete")
	CerrDeleteLevelExistUser            = code.New(DeleteLevelExistUser, "Delete Level Exist User")
	CerrLevelIsReportUsed               = code.New(LevelIsReportUsed, "Level Is Report Used")
	CerrForbidDeleteWithSubUser         = code.New(ForbidDeleteWithSubUser, "Forbid Delete With Sub User")
	CerrUserHaveExist                   = code.New(UserHaveExist, "User exists")
	CerrRightNotExist                   = code.New(RightNotExists, "right not exists")
	CerrGetDashboardConfigFail          = code.New(GetDashboardConfigFail, "Find Config Fail")
	Forbidden                           = code.New(Forbiden, "Forbidden")
)

func CerrApiApiBadJsonPayload(err error) code.Error {
	return code.Newf(ApiBadJsonPayload, "json payload wrong,%v", err)
}

func CerrApiBadDatetime(f string) code.Error {
	return code.Newf(ApiBadDatetime, "parameter [%s] is not a valid data time", f)
}

func CerrApiLengthOverLimit(name string, pt constant.ParameterType, max, got int) code.Error {
	return code.Newf(ApiBadDatetime, "%s parameter [%v] over length, max is %d got %d ", name, pt, max, got)
}

func CerrApiLengthNotEnough(name string, pt constant.ParameterType, min, got int) code.Error {
	return code.Newf(ApiBadDatetime, "%v parameter [%s] over length, min is %d got %d ", name, pt, min, got)
}

func CerrApiIntegerValueOverRange(name string, pt constant.ParameterType, min, max int) code.Error {
	return code.Newf(ApiBadDatetime, "%v parameter [%s] over length, range [%d,%d] ", pt, name, min, max)
}

func CerrApiFloatValueOverRange(name string, pt constant.ParameterType, min, max float64) code.Error {
	return code.Newf(ApiBadDatetime, "%s parameter [%v] over length, range [%d,%d] ", name, pt, min, max)
}

func CerrApiLackParameterWithCandidates(name string, pt constant.ParameterType, candidates ...interface{}) code.Error {
	return code.Newf(ApiLackParameter, "missing %v parameter [%s] , candidates %v ", pt, name, candidates)
}

func CerrApiLackParameter(name string, pt constant.ParameterType) code.Error {
	return code.Newf(ApiLackParameter, "missing %v parameter [%s]", pt, name)
}

func CerrApiBadParameterWithCandidates(name string, pt constant.ParameterType, candidates ...interface{}) code.Error {
	return code.Newf(ApiLackParameter, "bad %v parameter [%s] , candidates %v ", pt, name, candidates)
}
