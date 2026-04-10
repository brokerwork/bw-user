package model

// 处理字符串校验
import (
	"regexp"
	"strings"
)

const EMAIL_PATTERN  = "[\\w!#$%&'*+/=?^_`{|}~-]+(?:\\.[\\w!#$%&'*+/=?^_`{|}~-]+)*@(?:[\\w](?:[\\w-]*[\\w])?\\.)+[\\w](?:[\\w-]*[\\w])?"

// 校验是否是邮箱
func IsEmail(val string) (bool, error) {
	if strings.Contains(val, " ") {
		return false, nil
	}

	return regexp.MatchString(EMAIL_PATTERN, val)
}

