package server

import (
	"bw/bw-user/api/handler/introduce"
	"bw/bw-user/api/handler/levelctl"
	"bw/bw-user/api/handler/rightctl"
	"bw/bw-user/api/handler/rolectl"
	"bw/bw-user/api/handler/statistic"
	"bw/bw-user/api/handler/userctl"
	"bw/bw-user/api/handler/util"
	"context"
	"github.com/gin-gonic/gin"
	"github.com/lworkltd/kits/service/profile"
	"github.com/lworkltd/kits/service/restful/wrap"
	"io/ioutil"
	"time"

	"bw/bw-user/api/handler/dashboardctl"
	discoveryutils "github.com/lworkltd/kits/utils/discovery"
)

var wrapper *wrap.Wrapper

// TODO: 在gin所监听的接口同时处理pprof
func initService(_ *gin.Engine, option *profile.Service) error {
	wrapper = wrap.New(&wrap.Option{
		Prefix:      option.McodePrefix,
		LogLevel:    option.LogLevel,
		LogFilePath: option.LogFilePath,
	})

	if option.Reportable {
		if err := discoveryutils.RegisterServerWithProfile("/ping", option); err != nil {
			return err
		}
	}

	if option.PprofEnabled {
		// TODO:handle for pprof
	}

	return nil
}

func Setup(option *profile.Service) error {
	r := gin.New()
	r.Use(gin.Recovery())

	if err := initService(r, option); err != nil {
		return err
	}

	root := r.Group("/")
	root.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	if option.PathPrefix != "" {
		root = root.Group(option.PathPrefix)
	}

	root.GET("/release/info", func(c *gin.Context) {
		dat, _ := ioutil.ReadFile("release.version")
		c.JSON(200, gin.H{
			"time":    time.Time(time.Now()).Format("2006-01-02 15:04:05"),
			"version": string(dat),
			"name":    option.ReportName,
		})
	})

	routeV1User(root.Group("/v1/user"))
	//bw-api接口
	routeV1ApiUser(root.Group("/v1/api/user"))
	routeV2User(root.Group("/v2/user"))
	routeV1UserIntroduce(root.Group("/v1/user/introduce"))
	//routeV1UserSearch(root.Group("/v1/user/search"))
	routeV1Level(root.Group("/v1/level"))
	routeV1Role(root.Group("/v1/role"))
	routeV1Right(root.Group("/v1/right"))
	routeV1Dashboard(root.Group("/v1/dashboard"))

	routeV1Statistic(root.Group("/v1/statistic"))

	return r.Run(option.Host)
}

func routeV1User(v1User *gin.RouterGroup) {
	wrapper.Post(v1User, "/check", util.TenantCheckV2(userctl.Check))
	wrapper.Post(v1User, "/listByIds", util.TenantCheckV2(userctl.ListUserByIds))
	wrapper.Post(v1User, "/exists", util.TenantCheckV2(userctl.UserExists))
	wrapper.Post(v1User, "/currentUser", util.TenantCheckV2(userctl.ObtainCurrentUserInfo))
	wrapper.Post(v1User, "/updateActiveById", util.TenantCheckV2(userctl.UpdateActiveById))
	wrapper.Post(v1User, "/delete", util.TenantCheckV2(userctl.DeleteUser))
	wrapper.Post(v1User, "/bypubUser", util.TenantCheckV2(userctl.GetUserInfoByPubid))
	wrapper.Post(v1User, "/addAdminUser", util.TenantCheckV2(userctl.AddAdminUser))
	wrapper.Get(v1User, "/current/child/direct/id", util.TenantCheckV2(userctl.GetBelongUserId)) //查询当前用户直属下级用户
	wrapper.Get(v1User, "/child/direct/info", util.TenantCheckV2(userctl.GetUserBelongUser))     //查询当前用户直属下级用户
	//根据keyType（email,id,name,entityNo,login,roleId,levelId），查询value属于keyValue的用户详细信息
	wrapper.Get(v1User, "/findByKey/:keyValue/:keyType", util.TenantCheckV2(userctl.GetUserInfosByKey)) //根据KeyType 获取用户信息--用下面一种方式调用
	wrapper.Post(v1User, "/findByKey", util.TenantCheckV2(userctl.FindUserInfosByKey))                  //根据KeyType获取用户信息
	wrapper.Get(v1User, "/list/noparent/id", util.TenantCheckV2(userctl.GetNoParentuserIds))            //查询无parent的用户id
	//listUserAndLevel接口包含了listSimpleUser所需要的内容
	wrapper.Get(v1User, "/listSimpleUser", util.TenantCheckV2(userctl.ListSimpleUser))     //查询所有用户的简单信息（含Iid，name，entityNo）
	wrapper.Get(v1User, "/listUserAndLevel", util.TenantCheckV2(userctl.ListUserAndLevel)) //查询所有用户的简单信息（含id,name,parent,levelId,levelName,roleId,roleName,entityNo,pubUserId,login,vendorServerId）
	wrapper.Get(v1User, "/listSimpleUser/haveAccount", util.TenantCheckV2(userctl.ListSimpleUserHasAccountUser))
	wrapper.Post(v1User, "/findSimpleByKeyAndType", util.TenantCheckV2(userctl.GetOneSimpleUser))                         //查询一个用户的简单信息（含id,name,parent,levelId,levelName,roleId,roleName,entityNo,pubUserId,login,vendorServerId）
	wrapper.Post(v1User, "/findSimpleByPage", util.TenantCheckV2(userctl.GetSimpleUserByPage))                            //根据条件查询用户的简单信息，无权限过滤（含id,name,parent,levelId,levelName,roleId,roleName,entityNo,pubUserId,login,vendorServerId）
	wrapper.Post(v1User, "/findSimpleByPage/hasRight", util.TenantCheckV2(userctl.GetSimpleUserHasRightByPage))           //根据条件查询用户的简单信息，有权限过滤（含id,name,parent,levelId,levelName,roleId,roleName,entityNo,pubUserId,login,vendorServerId）
	wrapper.Post(v1User, "/list", util.TenantCheckV2(userctl.GetUserDetailByPage))                                        //根据条件查询用户的详情信息
	wrapper.Get(v1User, "/report/count/yesterday", util.TenantCheckV2(userctl.GetUserCountYesterday))                     //查询截止到今天凌晨0点的用户数
	wrapper.Get(v1User, "/report/introduce/count/yesterday", util.TenantCheckV2(introducectl.GetIntroduceCountYesterday)) //查询截止到今天凌晨0点的introduce数
	wrapper.Get(v1User, "/tree/child", util.TenantCheckV2(userctl.GetUserTreeChild))                                      //用户树查询下级信息接口, 根据不同权限和模块显示不同返回
	wrapper.Get(v1User, "/tree", util.TenantCheckV2(userctl.GetUserTree))                                                 //用户树查询下级完整用户数信息接口
	wrapper.Get(v1User, "/tree/search", util.TenantCheckV2(userctl.SearchUserTree))                                       //用户树查询下级完整用户数信息接口, 根据不同权限和模块显示不同返回
	wrapper.Post(v1User, "/countTenat", util.TenantCheckV2(userctl.GetTenantUserCount))                                   //查询登录租户的用户数
	wrapper.Post(v1User, "/listByIdsAndRoles", util.TenantCheckV2(userctl.ListUserInfoByIdsAndRoles))                     //目前用于根据用户id和角色查询用户,两结果取并集
	wrapper.Post(v1User, "/list/childIds", util.TenantCheckV2(userctl.ListChildUserIds))                                  //查询当前用户下级用户信息
	wrapper.Get(v1User, "/list/userId/childIds", util.TenantCheckV2(userctl.ListChildUserIdsById))                        //查询下级用户信息
	wrapper.Get(v1User, "/list/userId/child", util.TenantCheckV2(userctl.ListChildUserById))                              //查询下级用户信息
	wrapper.Post(v1User, "/list/type", util.TenantCheckV2(userctl.FindUserByType))                                        //根据返佣层级或角色ID查询相关用户信息（层级查询支持包含上级层级用户）
	wrapper.Post(v1User, "/list/type/fuzzy", util.TenantCheckV2(userctl.FindUserByTypeFuzzy))                             //根据返佣层级或角色ID模糊查询查询相关用户信息（层级查询支持包含上级层级用户）
	wrapper.Post(v1User, "/updateCurrentUser", util.TenantCheckV2(userctl.UpdateCurrentUser))                             //更新当前用户
	wrapper.Post(v1User, "/updateUser", util.TenantCheckV2(userctl.UpdateUserV1))                                         //全量更新其他用户信息
	wrapper.Post(v1User, "/msgReceiversQuery", util.TenantCheckV2(userctl.MsgReceiversQuery))                             //
	wrapper.Post(v1User, "/listAll", util.TenantCheckV2(userctl.ListAllUserDetail))                                       //查询全量用户详细信息
	wrapper.Post(v1User, "/findUserByField", util.TenantCheckV2(userctl.FindUserByField))                                 //模糊搜索查询用户指定字段值（有权限过滤）
	wrapper.Post(v1User, "/findLikeName/hasRight", util.TenantCheckV2(userctl.FindLikeNameWithRight))                     //模糊搜索查询权限范围内的用户列表，包括用户名、角色名、层级名任意一个匹配（有权限过滤）
	wrapper.Post(v1User, "/simpleUserCommissionRight", util.TenantCheckV2(userctl.GetSimpleUserCommissionRight))          //返佣用户查询
	//wrapper.Get(v1User, "/simpleUserAccRight", util.TenantCheckV2(userctl.GetSimpleUserAccRight))			//返佣账户查询
	wrapper.Get(v1User, "/simpleUserModuleRight", util.TenantCheckV2(userctl.GetSimpleUserByModule)) //返佣账户查询
	//链接由/v1/user/{parentId}/updateParentBatch改为/v1/user/updateParentBatch?parentId={parentId}
	wrapper.Post(v1User, "/updateParentBatch", util.TenantCheckV2(userctl.UpdateParentBatch))                      //批量归属
	wrapper.Get(v1User, "/permission/check/:targetUserId", util.TenantCheckV2(userctl.CheckUserIdPermissionScope)) //根据模块检查指定用户ID是否在权限范围内

	wrapper.Get(v1User, "/fields/list", util.TenantCheckV2(userctl.GetUserFieldsList))
	wrapper.Post(v1User, "/fields/update", util.TenantCheckV2(userctl.UpdateUserFields))
	wrapper.Post(v1User, "/updateEmail", util.TenantCheckV2(userctl.UpdateEmail)) //修改用户邮箱
	wrapper.Post(v1User, "/update2FAConfig", util.TenantCheckV2(userctl.UpdateTwoFAConfig))
	//url由/v1/user/search/{id}改成/v1/user/search/byId?id={id}
	wrapper.Get(v1User, "/search/byId", util.TenantCheckV2(userctl.GetOneSearchInfo))                          //获取自定义搜索详情
	wrapper.Get(v1User, "/search/dropdown", util.TenantCheckV2(userctl.GetSearchDropDown))                     //获取自定义搜索列下拉菜单
	wrapper.Get(v1User, "/search/list", util.TenantCheckV2(userctl.GetSearchList))                             //获取自定义搜索列表
	wrapper.Post(v1User, "/search/add", util.TenantCheckV2(userctl.AddSearchInfo))                             //新增自定义搜索
	wrapper.Post(v1User, "/search/delete/:id", util.TenantCheckV2(userctl.DeleteOneSearchInfo))                //删除自定义搜索
	wrapper.Post(v1User, "/search/edit", util.TenantCheckV2(userctl.EditOneSearchInfo))                        //编辑（更新）自定义搜索
	wrapper.Post(v1User, "/findLikeRoleLevelUser", util.TenantCheckV2(userctl.FindLikeRoleLevelUserWithRight)) //模糊搜索查询权限范围内的角色、层级、用户

	wrapper.Get(v1User, "/getUserPoints", util.TenantCheckV2(userctl.GetUserPoints)) //获取用户积分
}

func routeV1ApiUser(v1ApiUser *gin.RouterGroup) {
	wrapper.Get(v1ApiUser, "/ids", util.TenantCheckV2(userctl.GetUserIDByEmail))              //根据用户email获取用户id
	wrapper.Get(v1ApiUser, "/updateUserPoints", util.TenantCheckV2(userctl.UpdateUserPoints)) //更新用户积分
}

func routeV2User(v2User *gin.RouterGroup) {
	wrapper.Post(v2User, "/list", util.TenantCheckV2(userctl.GetUserDetailByPageV2)) //根据条件查询用户的详情信息
	wrapper.Post(v2User, "/add", util.TenantCheckV2(userctl.AddUser))
	wrapper.Post(v2User, "/updateUser", util.TenantCheckV2(userctl.UpdateUserV2))
	wrapper.Post(v2User, "/search", util.TenantCheckV2(userctl.SearchUser))
	wrapper.Post(v2User, "/search/simpleUser", util.TenantCheckV2(userctl.SearchSimpleUser))
	wrapper.Get(v2User, "/statByRole", userctl.UserStatByRoleId)
}

func routeV1Statistic(v1Statistic *gin.RouterGroup) {
	wrapper.Get(v1Statistic, "/user/count/remain", util.TenantCheckV2(statistic.RemainUserCount)) //根据条件查询用户的详情信息
}

func routeV1UserIntroduce(v1UserIntrudoce *gin.RouterGroup) {
	wrapper.Post(v1UserIntrudoce, "/addIntroduce", util.TenantCheckV2(introducectl.AddIntroduce))
	wrapper.Get(v1UserIntrudoce, "/list/simple", util.TenantCheckV2(introducectl.IntroduceListSimple))
	wrapper.Get(v1UserIntrudoce, "/list", util.TenantCheckV2(introducectl.IntroduceList))
	wrapper.Get(v1UserIntrudoce, "/findByKey/:keyValue/:keyType", util.TenantCheckV2(introducectl.GetIntroduceInfosByKey))
	//wrapper.Post(v1UserIntrudoce, "/addDefaultAgentIntroduce", util.TenantCheckV2(introducectl.AddDefaultAgentIntroduce))		//生成默认代理推广链接
	//wrapper.Post(v1UserIntrudoce, "/addDefaultIntroduce", util.TenantCheckV2(introducectl.AddDefaultTwIntroduce))				//生成默认TW推广链接
	//wrapper.Post(v1UserIntrudoce, "/updateDefaultIntroduce", util.TenantCheckV2(introducectl.UpdateDefaultTwIntroduce))				//更新默认TW推广链接
	wrapper.Post(v1UserIntrudoce, "/hit/add", util.TenantCheckV2(introducectl.AddIntroduceHit))                 //添加推广链接点击信息
	wrapper.Get(v1UserIntrudoce, "/myIntroduces", util.TenantCheckV2(introducectl.GetMyIntroduces))             //BW我的推广链接
	wrapper.Get(v1UserIntrudoce, "/twDirectIntroduces", util.TenantCheckV2(introducectl.GetTWDirectIntroduces)) //TW直客推广链接
	wrapper.Get(v1UserIntrudoce, "/twAgentIntroduces", util.TenantCheckV2(introducectl.GetTWAgentIntroduces))   //TW代理推广链接
	//访问链接由/v1/user/introduce/60/qrcode?isCurrentUserUrl=true改成/v1/user/introduce/qrcode?isCurrentUserUrl=true&introduceId=40
	wrapper.Get(v1UserIntrudoce, "/qrcode", util.TenantCheckV2(introducectl.GetIntroducesQrcode))
	wrapper.Get(v1UserIntrudoce, "/twQrcode", util.TenantCheckV2(introducectl.GetTwIntroducesQrcode)) //生成推广二维码
	//访问链接由/v1/user/introduce/{id}/detail?statistic=true改成/v1/user/introduce/detail?introduceId={introduceId}&statistic=true
	wrapper.Get(v1UserIntrudoce, "/detail", util.TenantCheckV2(introducectl.GetIntroducesDetail))
	//访问路径由/v1/user/introduce/{introduceId}/switch改为/v1/user/introduce/switch?introduceId={introduceId}
	wrapper.Post(v1UserIntrudoce, "/switch", util.TenantCheckV2(introducectl.SwitchIntroduceState))  //切换推广链接启用状态
	wrapper.Post(v1UserIntrudoce, "/delete", util.TenantCheckV2(introducectl.DeleteIntroduce))       //批量删除推广链接
	wrapper.Post(v1UserIntrudoce, "/update", util.TenantCheckV2(introducectl.UpdateSystemIntroduce)) //更新推广链接
}

func routeV1Level(v1Level *gin.RouterGroup) {
	wrapper.Get(v1Level, "/list", util.TenantCheckV2(levelctl.GetLevelList))
	wrapper.Get(v1Level, "/list/byAuthority", util.TenantCheckV2(levelctl.GetLevelListByAuthority))
	wrapper.Get(v1Level, "/list/earningReport/byAuthority", util.TenantCheckV2(levelctl.GetEarningReportByAuthority))
	wrapper.Post(v1Level, "/add", util.TenantCheckV2(levelctl.Addlevel))       //添加层级
	wrapper.Post(v1Level, "/delete", util.TenantCheckV2(levelctl.Deletelevel)) //删除层级
	wrapper.Post(v1Level, "/update", util.TenantCheckV2(levelctl.UpdateLevel)) //更新层级
}

func routeV1Role(v1Role *gin.RouterGroup) {
	wrapper.Get(v1Role, "/child", util.TenantCheckV2(rolectl.GetRoleChild))                    //路径由"/v1/role/:roleId/child"改成""/v1/role/child?roleId={roleId}"
	wrapper.Get(v1Role, "/findCurrentSetRole", util.TenantCheckV2(rolectl.FindCurrentSetRole)) //当前用户可以在新增/修改用户时设置的角色信息
	wrapper.Get(v1Role, "/report/level/count/yesterday", util.TenantCheckV2(rolectl.GetLevelCountYesterday))
	wrapper.Get(v1Role, "/report/role/count/yesterday", util.TenantCheckV2(rolectl.GetRoleCountYesterday))
	wrapper.Post(v1Role, "/currentRight", util.TenantCheckV2(rolectl.GetCurrentRight)) //当前用户角色具有的权限集合
	wrapper.Post(v1Role, "/listDetail", util.TenantCheckV2(rolectl.GetRoleDetails))    //查询结果不含child
	wrapper.Post(v1Role, "/list", util.TenantCheckV2(rolectl.GetRoleSimple))           //查询role列表，输出只有id、name、entityNo
	wrapper.Post(v1Role, "/right/id", util.TenantCheckV2(rolectl.GetRightIDsByRoleIDs))
	wrapper.Post(v1Role, "/upsert", util.TenantCheckV2(rolectl.UpsertRole))                       //新增/修改role
	wrapper.Post(v1Role, "/listSimpleHasRight", util.TenantCheckV2(rolectl.GetSimpleRoleByRight)) //查询具体任务处理权限角色
	//访问链接由/v1/role/{roleId}/right/id改成/v1/role/roleId/right/id?roleId={roleId}
	wrapper.Get(v1Role, "/roleId/right/id", util.TenantCheckV2(rolectl.GetRightIDsByRoleID))  //根据一个RoleId获取其权限ID列表
	wrapper.Get(v1Role, "/roleId/right/key", util.TenantCheckV2(rolectl.GetRightKeyByRoleID)) //根据一个RoleId获取其权限KEY列表
	//访问链接由/v1/role/{roleId}/child/tree改成/v1/role/child/tree?roleId={roleId}
	wrapper.Get(v1Role, "/child/tree", util.TenantCheckV2(rolectl.GetRoleChildTree))                                                //查询单个角色详情 子集
	wrapper.Post(v1Role, "/remove", util.TenantCheckV2(rolectl.RemoveRole))                                                         //删除角色
	wrapper.Get(v1Role, "/type/list", util.TenantCheckV2(rolectl.GetRoleTypeList))                                                  //获取角色类型列表
	wrapper.Get(v1Role, "/type/right/:roleTypeId", util.TenantCheckV2(rolectl.GetRightByRoleTypeId))                                //获取指定角色类型的权限ID
	wrapper.Get(v1Role, "/fresh/role/type", util.TenantCheckV2(rolectl.FreshRole4RoleType))                                         //刷新角色
	wrapper.Post(v1Role, "/search", util.TenantCheckV2(rolectl.SearchRoleSimpleInfo))                                               //自定义获取角色基本信息
	wrapper.Post(v1Role, "/addRoleRightDependOtherRight/:depend/:right", util.TenantCheckV2(rightctl.AddRoleRightDependOtherRight)) // 根据角色是否有某个权限设置所有租户各个角色指定权限勾选状态
}

func routeV1Right(v1Right *gin.RouterGroup) {
	wrapper.Post(v1Right, "/listTopRights", util.TenantCheckV2(rightctl.GetListTopRights)) //获取无上级的权限，其结果集包括child
	wrapper.Get(v1Right, "/initRight", util.TenantCheckV2(rightctl.InitRight))             //初始化权限树
	wrapper.Post(v1Right, "/checkIdsRight", util.TenantCheckV2(rightctl.CheckIdsRight))    //判断ids权限
	wrapper.Post(v1Right, "/addIdsRight", util.TenantCheckV2(rightctl.AddIdsRight))        //按照角色ids 添加权限
}

func routeV1Dashboard(v1DashBoard *gin.RouterGroup) {
	wrapper.Post(v1DashBoard, "/config/save", util.TenantCheckV2(dashboardctl.SaveUserConfig))          //保存用户我的仪表盘配置
	wrapper.Get(v1DashBoard, "/config/user/:keyUserId", util.TenantCheckV2(dashboardctl.GetUserConfig)) //获取指定用户我的仪表盘配置
	wrapper.Post(v1DashBoard, "/config/delete   ", util.TenantCheckV2(dashboardctl.DeleteUserConfig))   //删除指定用户我的仪表盘配置
}

func ctxFromGinContext(ctx *gin.Context) context.Context {
	// TODO:pending the context from gin conetext
	return context.Background()
}
