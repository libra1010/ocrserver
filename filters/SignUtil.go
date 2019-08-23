package filters

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type signUtil uint

const SignUtil signUtil = 1

const (

	/*
		保持
	*/
	SIGN_TRY_KEEP = 0
	/**
	原始
	*/
	SIGN_TRY_PLAN = 1
	/**
	加密后
	*/
	SIGN_TRY_ENDOCING = 2
)

func (this signUtil) Sign(identity SignIdentity, signKey string) string {
	return this.SignTry(identity, signKey, SIGN_TRY_KEEP)
}

func (this signUtil) SignTry(identity SignIdentity, signKey string, tryType int) string {
	signKey = strings.TrimSpace(signKey)
	if signKey == "" {
		return ""
	}

	queryString := identity.Url.RawQuery
	md5String := queryString

	if md5String != "" && tryType == SIGN_TRY_PLAN {

		var e error
		md5String, e = url.QueryUnescape(md5String)
		if e != nil {
			md5String = queryString
		}

	} else if md5String != "" && tryType == SIGN_TRY_ENDOCING {
		md5String, e := url.QueryUnescape(md5String)
		if e != nil {
			md5String = queryString
		}
		md5String = url.QueryEscape(md5String)
	}

	orSignStr := signKey + identity.Url.Path + md5String + strconv.FormatInt(identity.TS, 10)
	fmt.Println("原始加密串: ", orSignStr)
	serviceSign := MD5.ToHex(signKey + identity.Url.Path + md5String + strconv.FormatInt(identity.TS, 10))

	return serviceSign
}

func (this signUtil) ValidateSign(identity SignIdentity, signKey string, clientSign string) bool {
	signKey = strings.TrimSpace(signKey)
	clientSign = strings.TrimSpace(clientSign)

	if signKey == "" || clientSign == "" {
		return false
	}

	t := time.Unix(identity.TS, 0)
	if !(t.Year() >= 9999 || t.Year() <= 1970) {
		current := time.Now().Unix()
		if current-identity.TS > 60 || current-identity.TS < 0 {
			return false
		}
	}

	serviceSign := this.Sign(identity, signKey)

	if !strings.EqualFold(serviceSign, clientSign) {
		serviceSign = this.SignTry(identity, signKey, SIGN_TRY_ENDOCING)
		if !strings.EqualFold(serviceSign, clientSign) {
			serviceSign = this.SignTry(identity, signKey, SIGN_TRY_PLAN)
			fmt.Println(serviceSign)
		}
	}

	return strings.EqualFold(serviceSign, clientSign)
}
