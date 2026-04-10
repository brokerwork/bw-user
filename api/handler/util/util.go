package util

import (
	//"bw/user/api/constant"
	//"bw/user/api/errcode"
	//"strconv"
	//"strings"

	"github.com/lworkltd/kits/service/context"
	"github.com/lworkltd/kits/service/restful/wrap"

	"github.com/gin-gonic/gin"
	"github.com/lworkltd/kits/service/restful/code"
	"bw/bw-user/constant"
	"bw/bw-user/model"
	"encoding/json"
	"bw/bw-user/errcode"
	"net/url"
)

// ApiArgsReader 是一个参数阅读器
type ApiArgsReader interface {
	CheckBase() code.Error
}





// TenantArgsHandlerFunc 包含了ApiArgsReader的API Handler函数类型
type TenantArgsHandlerFunc func(srvContext context.Context, argsReader ApiArgsReader, ctx *gin.Context) (interface{}, code.Error)

// TenantCheck 租户授权校验封装
func TenantCheck(f TenantArgsHandlerFunc) wrap.WrappedFunc {
	return func(srvContext context.Context, ctx *gin.Context) (interface{}, code.Error) {
		args := ArgsReaderFromGinContext(ctx)
		if cerr := args.CheckBase(); cerr != nil {
			return nil, cerr
		}

		return f(srvContext, args, ctx)
	}
}


// TenantArgsHandlerFunc 包含了ApiArgsReader的API Handler函数类型
type TenantArgsHandlerFuncV2 func(srvContext context.Context, argsReader ApiArgsReader, ctx *gin.Context, feignKey *model.XFeignKey) (interface{}, code.Error)

// TenantCheck 租户授权校验封装，解析并传递XFeignKey
func TenantCheckV2(f TenantArgsHandlerFuncV2) wrap.WrappedFunc {
	return func(srvContext context.Context, ctx *gin.Context) (interface{}, code.Error) {
		args := ArgsReaderFromGinContext(ctx)
		if cerr := args.CheckBase(); cerr != nil {
			return nil, cerr
		}

		//解析header中的FeigKey
		feignKeyStr, err := url.QueryUnescape(ctx.GetHeader(constant.HeaderAPIXFeigKey))
		if nil != err {
			feignKeyStr = ctx.GetHeader(constant.HeaderAPIXFeigKey)
		}
		var feignKey model.XFeignKey
		if err := json.Unmarshal([]byte(feignKeyStr), &feignKey); nil != err || "" == feignKey.TenantId{
			srvContext.Errorf("BindJson Failed,%v", err)
			return nil, errcode.CerrApiApiBadJsonPayload(err)
		}

		return f(srvContext, args, ctx, &feignKey)
	}
}
