package filters

import (
	"crypto/md5"
	"encoding/hex"
)

type cryptoSecurity uint

const MD5 cryptoSecurity = 1

//MD5
func (h cryptoSecurity) ToHex(content string) string {
	hash := md5.New()
	hash.Write([]byte(content))
	cipherStr := hash.Sum(nil)
	return hex.EncodeToString(cipherStr)
}
