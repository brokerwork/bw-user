package util

import (
	//"bw/user/api/constant"
	//"bw/user/api/errcode"
	//"strconv"
	//"strings"

	"github.com/gin-gonic/gin"
	"github.com/lworkltd/kits/service/restful/code"
)

type tenantApiReader struct {
	*gin.Context
}

func (reader *tenantApiReader) CheckBase() code.Error {

	return nil
}


// ArgsReaderFromGinContext 使用gin.Context 创建一个API接口参数读取器
func ArgsReaderFromGinContext(ctx *gin.Context) ApiArgsReader {
	return &tenantApiReader{
		Context: ctx,
	}
}
