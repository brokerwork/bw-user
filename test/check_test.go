package test

import (
	"bw/bw-user/constant"
	"bw/bw-user/model"
	"testing"
)

func Test_newUserTree(t *testing.T) {

	var aeskey = []byte(constant.AES_KEY)

	var checkInfo = "bk.goldwellcap.com@-@1767196799000"
	encryptInfo, err := model.AesEncrypt(checkInfo, aeskey)
	if err != nil {
		t.Fatal(err)
	}


	println(encryptInfo)
	result, err := model.AesDecrypt(encryptInfo, aeskey)
	if err != nil {
		t.Fatal(err)
	}

	if string(result) != checkInfo{
		t.Fatal("not same")
	}
}
