package model

import (
	"bw/bw-user/constant"
	"bw/bw-user/errcode"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"github.com/Sirupsen/logrus"
	"github.com/lworkltd/kits/service/restful/code"
	"strconv"
	"strings"
	"time"

	//"github.com/jinzhu/gorm"
	"gopkg.in/mgo.v2/bson"
)

func CheckTenant(feignKey *XFeignKey, domain string) (bool, code.Error) {
	mgoSess, errSess := GetMongoConnByTenantID(feignKey.TenantId)
	if nil != errSess {
		logrus.WithFields(logrus.Fields{"tenant_id": feignKey.TenantId, "error": errSess}).Error("GetMongoConnByTenantID failed")
		return false, errSess
	}
	defer mgoSess.Close()
	coll := GetMgoCheckColl(mgoSess)

	var record CheckInfo
	condition := bson.M{"tenantId": feignKey.TenantId}

	errGet := coll.Find(condition).One(&record)
	if nil != errGet {
		logrus.WithFields(logrus.Fields{"tenant_id": feignKey.TenantId, "error": errGet}).Error("Get tenant info record failed")
		return false, errcode.CerrOperateMongo
	}

	//AES秘钥
	var aeskey = []byte(constant.AES_KEY)

	tenantInfo, aesErr := AesDecrypt(record.Info, aeskey)
	if aesErr != nil {
		return false, code.Newf(1111, "Decrypt wrong,%v", aesErr)
	}

	infos := strings.Split(string(tenantInfo), "@-@")
	validDomains := strings.Split(infos[0], ",")
	expire, err := strconv.ParseInt(infos[1], 10, 64)
	if nil != errGet {
		return false,  code.Newf(1111, "Decrypt wrong,%v", err)
	}

	contain := false
	for index := range validDomains {
		if validDomains[index] == domain {
			contain = true
		}
	}

	now := time.Now().UnixNano() / 1e6
	return contain && expire > now, nil
}



func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

func AesEncrypt(orig string, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	origData := []byte(orig)


	blockSize := block.BlockSize()
	origData = PKCS5Padding(origData, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	crypted := make([]byte, len(origData))
	blockMode.CryptBlocks(crypted, origData)
	return base64.StdEncoding.EncodeToString(crypted), nil
}

func AesDecrypt(cryptedstr string, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	crypted, err := base64.StdEncoding.DecodeString(cryptedstr)
	if err != nil {
		return "", err
	}

	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	origData := make([]byte, len(crypted))
	blockMode.CryptBlocks(origData, crypted)
	origData = PKCS5UnPadding(origData)
	return string(origData), nil
}