package model

import (
	//"fmt"
	//"net/http"
	"testing"
)

func Test_EmailValidator(t *testing.T) {
	if result,_ := IsEmail("abcd1234") ; false != result {
		t.Errorf("EmailValidator: %s \n", "abcd1234")
	}

	if result,_ := IsEmail("abcd1234@qq.com") ; true != result {
		t.Errorf("EmailValidator: %s \n", "abcd1234")
	}

	if result,_ := IsEmail("abcd1234 @qq.com") ; false != result {
		t.Errorf("EmailValidator: %s \n", "abcd1234")
	}

	if result,_ := IsEmail("abcd1234@qq.com") ; true != result {
		t.Errorf("EmailValidator: %s \n", "abcd1234")
	}

	if result,_ := IsEmail("@@@@qq.com") ; false != result {
		t.Errorf("EmailValidator: %s \n", "abcd1234")
	}

	if result,_ := IsEmail("@12312@qq.com") ; true != result {
		t.Errorf("EmailValidator: %s \n", "abcd1234")
	}
}
