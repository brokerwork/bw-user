package model

import (
	"bw/bw-user/constant"
	"bw/bw-user/errcode"
	"bytes"
	"github.com/Sirupsen/logrus"
	"github.com/lworkltd/kits/service/restful/code"
	invokeutils "github.com/lworkltd/kits/utils/invoke"
	"mime/multipart"
	"net/http"
)

func SaveFileToAliyun(uri string, params map[string]string, fileByte []byte) (string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "file")
	if err != nil {
		return "", err
	}
	_, err = part.Write(fileByte)
	if err != nil {
		return "", err
	}
	for key, val := range params {
		_ = writer.WriteField(key, val)
	}
	err = writer.Close()
	if err != nil {
		return "", err
	}

	if err != nil {
		return "", nil
	}
	client := &http.Client{}
	request, err := http.NewRequest("POST", uri, body)
	if err != nil {
		return "", nil
	}

	request.Header.Add("Content-Type", writer.FormDataContentType())
	rsp, err := client.Do(request)
	if err != nil {
		return "", nil
	}
	var response string
	cerr := invokeutils.ExtractHttpResponse(constant.ServiceNamePic, err, rsp, &response)
	if cerr != nil {
		logrus.WithFields(logrus.Fields{
			"error": cerr,
		}).Error("link custom failed")
		return "", code.Newf(errcode.UploadPicErr, "upload err %v", cerr)
	}
	return "//" + response, nil
}
