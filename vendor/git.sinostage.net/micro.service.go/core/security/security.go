package koloSecurity

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/base64"
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

type aesCrypto uint
type ecbCrypto uint
type aseStruct struct {
	CBC aesCrypto
	ECB ecbCrypto
}

var AES = aseStruct{}

type padding string
type encoder string
type decoder string
type paddingHandle func(ciphertext []byte, blockSize int) []byte
type unPaddingHandle func(ciphertext []byte) []byte

const (
	Padding_PKCS5 padding = "pkcs5"
	Padding_PKCS7 padding = "pkcs7"

	Encoder_BASE64 encoder = "base64"
	Encoder_HEX    encoder = "hex"
	Encoder_NONE   encoder = "none"

	Decoder_BASE64 decoder = "base64"
	Decoder_HEX    decoder = "hex"
	Decoder_NONE   decoder = "none"

	aes128 = 16
	aes192 = 24
	aes256 = 32
)

func pKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func pKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

//使用PKCS7进行填充，IOS也是7
func pKCS7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func pKCS7UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

func pkcs7Unpad(data []byte) []byte {
	//if blockSize <= 0 {
	//	return nil
	//}
	//if len(data)%blockSize != 0 || len(data) == 0 {
	//	return nil
	//}
	c := data[len(data)-1]
	n := int(c)
	if n == 0 || n > len(data) {
		return nil
	}
	for i := 0; i < n; i++ {
		if data[len(data)-n+i] != c {
			return nil
		}
	}
	return data[:len(data)-n]
}

func decoderString(data string, decoder decoder) ([]byte, error) {
	if decoder == Decoder_BASE64 {
		return base64.StdEncoding.DecodeString(data)
	} else if decoder == Decoder_HEX {
		return hex.DecodeString(data)
	} else if decoder == Decoder_NONE {
		return []byte(data), nil
	}

	return nil, nil
}

func encodingToString(data []byte, encoder encoder) string {
	if encoder == Encoder_BASE64 {
		return base64.StdEncoding.EncodeToString(data)
	} else if encoder == Encoder_HEX {
		return hex.EncodeToString(data)
	} else if encoder == Encoder_NONE {
		return string(data)
	}

	return ""
}

func (this aesCrypto) Decrypt(crypted []byte, key []byte, iv []byte, paddingType padding) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	blockMode := cipher.NewCBCDecrypter(block, iv)
	origData := make([]byte, len(crypted))
	blockMode.CryptBlocks(origData, crypted)
	origData = getUnPaddingHandle(paddingType)(origData)
	return origData, nil
}

func (this aesCrypto) DecryptString(crypted string, key string, paddingType padding, decoder decoder) (string, error) {
	data, err := decoderString(crypted, decoder)
	if err != nil {
		return "", err
	}
	source, err := this.DecryptByteKeyIV(data, key, key, paddingType)
	return string(source), err
}

func (this aesCrypto) EncryptString(source string, key string, paddingType padding, encoder encoder) (string, error) {
	target, err := this.EncryptBytesKeyIV([]byte(source), key, key, paddingType)
	if err != nil {
		return "", err
	}

	result := encodingToString(target, encoder)
	return result, nil
}

func (this aesCrypto) DecryptStringKeyIV(crypted string, key string, iv string, paddingType padding, decoder decoder) (string, error) {
	data, err := decoderString(crypted, decoder)
	if err != nil {
		return "", err
	}
	source, err := this.DecryptByteKeyIV(data, key, iv, paddingType)
	return string(source), err
}

func (this aesCrypto) EncryptStringKeyIV(source string, key string, iv string, paddingType padding, encoder encoder) (string, error) {
	target, err := this.EncryptBytesKeyIV([]byte(source), key, iv, paddingType)
	if err != nil {
		return "", err
	}

	result := encodingToString(target, encoder)
	return result, nil
}

func (this aesCrypto) DecryptByte(crypted []byte, key string, paddingType padding) ([]byte, error) {
	return this.DecryptByteKeyIV(crypted, key, key, paddingType)
}

func (this aesCrypto) EncryptBytes(origData []byte, key string, paddingType padding) ([]byte, error) {
	return this.EncryptBytesKeyIV(origData, key, key, paddingType)
}

//aes加密，填充秘钥key的16位，24,32分别对应AES-128, AES-192, or AES-256.
//
func (this aesCrypto) EncryptBytesKeyIV(origData []byte, key string, iv string, paddingType padding) ([]byte, error) {
	block, err := aes.NewCipher(fillKey(key))
	if err != nil {
		return nil, err
	}

	blockSize := block.BlockSize()
	origData = getPaddingHandle(paddingType)(origData, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, fillIv(iv, blockSize))
	crypted := make([]byte, len(origData))
	blockMode.CryptBlocks(crypted, origData)
	return crypted, nil
}

func (this aesCrypto) DecryptByteKeyIV(crypted []byte, key string, iv string, paddingType padding) ([]byte, error) {
	block, err := aes.NewCipher(fillKey(key))
	if err != nil {
		return nil, err
	}

	blockSize := block.BlockSize()

	blockMode := cipher.NewCBCDecrypter(block, fillIv(iv, blockSize))
	origData := make([]byte, len(crypted))
	blockMode.CryptBlocks(origData, crypted)
	origData = getUnPaddingHandle(paddingType)(origData)
	return origData, nil
}

func getPaddingHandle(p padding) paddingHandle {
	if p == Padding_PKCS5 {
		return pKCS5Padding
	} else if p == Padding_PKCS7 {
		return pKCS7Padding
	}

	return nil
}

func getUnPaddingHandle(p padding) unPaddingHandle {
	if p == Padding_PKCS5 {
		return pKCS5UnPadding
	} else if p == Padding_PKCS7 {
		return pkcs7Unpad
	}

	return nil
}

func fillKey(key string) []byte {
	keyBytes := []byte(key)
	if len(keyBytes) < aes128 {
		newKey := make([]byte, aes128)
		count := copy(newKey, keyBytes)
		for ; count < aes128; count++ {
			newKey[count] = ([]byte(" "))[0]
		}
		return newKey
	} else if len(keyBytes) == aes128 {
		return keyBytes
	}

	if len(keyBytes) < aes192 {
		newKey := make([]byte, aes192)
		count := copy(newKey, keyBytes)
		for ; count < aes192; count++ {
			newKey[count] = ([]byte(" "))[0]
		}
		return newKey
	} else if len(keyBytes) == aes192 {
		return keyBytes
	}

	if len(keyBytes) < aes256 {
		newKey := make([]byte, aes256)
		count := copy(newKey, keyBytes)
		for ; count < aes256; count++ {
			newKey[count] = ([]byte(" "))[0]
		}
		return newKey
	} else if len(keyBytes) == aes256 {
		return keyBytes
	}

	return keyBytes[:aes256]
}

func fillIv(iv string, blockSize int) []byte {

	ivBytes := []byte(iv)
	if len(ivBytes) < blockSize {
		newIv := make([]byte, blockSize)
		count := copy(newIv, ivBytes)
		for ; count < blockSize; count++ {
			newIv[count] = ([]byte(" "))[0]
		}
		return newIv
	} else if len(ivBytes) > blockSize {
		return ivBytes[:blockSize]
	}

	return ivBytes
}
