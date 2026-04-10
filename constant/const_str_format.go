package constant


func FormatIntroduceType(introduceStr string) string {
	switch introduceStr {
	case IntroduceType_uid:
	case IntroduceType_pid:
	case IntroduceType_eid:
	case IntroduceType_Web:
	case IntroduceType_Mobile:
	case IntroduceType_UserAllVisible:
	case IntroduceType_UserPartVisible:
	case IntroduceType_UserInVisible:
	case IntroduceType_UserNotVisible:
	case IntroduceType_DirectAllVisible:
	case IntroduceType_DirectPartVisible:
	case IntroduceType_DirectPartInvisible:
	case IntroduceType_DirectNotVisible:
	case IntroduceType_StraightGuest:
	case IntroduceType_Agent:
	case IntroduceType_DirectRecommendation:
	default:
		return ""
	}
	return introduceStr
}

func FormatVendor(vendor string) string {
	switch vendor {
	case Vendor_SAXO:
	case Vendor_PingAn:
	case Vendor_MT4:
	case Vendor_MT5:
	case Vendor_Esunny:
	case Vendor_SAXOFIX:
	case Vendor_IGFIX:
	case Vendor_LMAXFIX:
	case Vendor_CTRADER:
	case IntroduceType_Agent:
	default:
		return ""
	}
	return vendor
}