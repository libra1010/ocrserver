package filters

import "net/url"

type SignIdentity struct {
	//uuid
	UUID string
	//TS
	TS int64
	//语言
	Language string

	//平台
	Platform string

	//调试状态
	Debug bool

	Env string

	//签名
	Sign string

	//系统信息
	Agent string

	//调用者
	Caller string

	SkipSign bool

	Token string

	Version string

	Url *url.URL
}
