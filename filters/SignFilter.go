package filters

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type SignFilter struct {
	Logger *log.Logger
	Next   http.Handler
}

// ServeHTTP ...
func (f *SignFilter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Access-Control-Allow-Origin,X-UUID,X-TS,X-RANDOM,X-SIGN,X-ACCESS-TOKEN,X-USERAGENT,Origin,X-Requested-With,content-type,x-access-token,X-PLATFORM,X-LANGUAGE,X-VERSION,X-DEBUG")
	w.Header().Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Add("Access-Control-Expose-Headers", "Access-Control-Allow-Origin,X-UUID,X-TS,X-RANDOM,X-SIGN,X-ACCESS-TOKEN,X-USERAGENT,Origin,X-Requested-With,content-type,x-access-token,X-PLATFORM,X-LANGUAGE,X-VERSION,X-DEBUG")
	w.Header().Add("Access-Control-Allow-Credentials", "true")

	//放行所有OPTIONS方法
	if strings.ToUpper(r.Method) == "OPTIONS" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	clientSign := r.Header.Get("X-SIGN")
	if clientSign == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	timestampStr := r.Header.Get("X-TS")
	if timestampStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ts, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		fmt.Println("时间错错误", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	identity := SignIdentity{TS: ts, Url: r.URL}

	if !SignUtil.ValidateSign(identity, "ocr.signkey", clientSign) {
		fmt.Println("签名错误 404")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	f.Next.ServeHTTP(w, r)
}

// SetNext ...
func (f *SignFilter) SetNext(next http.Handler) {
	f.Next = next
}
