package model

import (
	"bw/bw-user/constant"
	"bw/bw-user/errcode"
	"encoding/json"
	"github.com/Sirupsen/logrus"
	"github.com/lworkltd/kits/service/invoke"
	"github.com/lworkltd/kits/service/restful/code"
	invokeutils "github.com/lworkltd/kits/utils/invoke"
)

type AgencyRegisterStatsDTO struct {
	JobId        string `json:"jobId"`
	JobNo        string `json:"jobNo"`
	TenantId     string `json:"tenantId"`
	TaskState    string `json:"jobState"`
	CustomSource string `json:"customSource"`
	Uid          string `json:"uid"`
}

func ClientTaskAgencyList(feignKey *XFeignKey, customerSource string) ([]AgencyRegisterStatsDTO, code.Error) {
	feignkeyBypes, _ := json.Marshal(feignKey)
	rsp, err := invoke.Name(constant.ServiceNameTaskJob).
		Post(constant.PathTaskAgencyList).
		Header("X-Feign-Key", string(feignkeyBypes)).
		Query("customerSource", customerSource).
		Response()

	var response []AgencyRegisterStatsDTO
	cerr := invokeutils.ExtractHttpResponse(constant.ServiceNameTaskJob, err, rsp, &response)
	if cerr != nil {
		logrus.WithFields(logrus.Fields{
			"error": cerr,
		}).Error("Task Agency List failed")
		return nil, errcode.CerrRequestTaskAgencyList
	}

	return response, nil
}
