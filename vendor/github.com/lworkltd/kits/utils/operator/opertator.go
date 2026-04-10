package operator




// 代替三元运算
func TernaryOperatorString(condition bool, trueResult string, falaeResult string) string {
	if condition == true {
		return trueResult
	}
	return falaeResult;
}



// 代替三元运算
func TernaryOperatorInt(condition bool, trueResult int, falaeResult int) int {
	if condition == true {
		return trueResult
	}
	return falaeResult;
}


// 代替三元运算
func TernaryOperatorInt64(condition bool, trueResult int64, falaeResult int64) int64 {
	if condition == true {
		return trueResult
	}
	return falaeResult;
}