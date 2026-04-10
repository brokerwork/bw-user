package conf

type Application struct {
	// 添加你的配置
	TenantId    string `toml:"tenantId"`
	ProductId   string `toml:"productId"`
	Registrable bool   `toml:"registrable"`
	AllowEmail  bool   `toml:"allowEmail"`
	AllowPhone  bool   `toml:"allowPhone"`
	PwdStrength int    `toml:"pwdStrength"`
	//PwdRegexMap                 map[int]string   `toml:"pwdRegexMap"`
	VerificationLoginFailTimes int64  `toml:"verificationLoginFailTimes"`
	LockLoginFailTimes         int64  `toml:"lockLoginFailTimes"`
	LogoutTime                 int64  `toml:"logoutTime"`
	IntroduceHitDBName         string `toml:"introduce_hit_db_name"`
	AdvanceSearchDBName        string `toml:"advance_search_db_name"`
	UserFieldsDBName           string `toml:"user_fields_db_name"`

	//私有化部署功能，返回OSS自定义域名
	PicCustomerDomainEnable string   `toml:"pic_customer_domain_enable"`
	PicCustomerDomain       string `toml:"pic_customer_domain"`
	//统一的图片上传服务器
	PicServerDomain string `toml:"picServerDomain"`

	LworkOfficialTenantId   string `toml:"lworkOfficialTenantId"`

	MqUrl      string     `toml:"Mq_Url"`
	RoleRights [][]string `toml:"roleRights"`
}

func GetApplication() *Application {
	return &configuration.Application
}
