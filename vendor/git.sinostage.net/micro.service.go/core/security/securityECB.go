package koloSecurity

import (
	"crypto/aes"
)

func (this ecbCrypto) DecryptString(crypted string, key string, paddingType padding, decoder decoder) (string, error) {
	data, err := decoderString(crypted, decoder)
	if err != nil {
		return "", err
	}
	source, err := this.DecryptBytes(data, key, paddingType)
	return string(source), err
}

func (this ecbCrypto) EncryptString(source string, key string, paddingType padding, encoder encoder) (string, error) {
	target, err := this.EncryptBytes([]byte(source), key, paddingType)
	if err != nil {
		return "", err
	}

	result := encodingToString(target, encoder)
	return result, nil
}

//aes加密，填充秘钥key的16位，24,32分别对应AES-128, AES-192, or AES-256.
//
func (this ecbCrypto) EncryptBytes(origData []byte, key string, paddingType padding) ([]byte, error) {

	cipher, _ := aes.NewCipher(fillKey(key))
	length := (len(origData) + aes.BlockSize) / aes.BlockSize
	plain := make([]byte, length*aes.BlockSize)
	copy(plain, origData)
	pad := byte(len(plain) - len(origData))
	for i := len(origData); i < len(plain); i++ {
		plain[i] = pad
	}
	encrypted := make([]byte, len(plain))
	// 分组分块加密
	for bs, be := 0, cipher.BlockSize(); bs <= len(origData); bs, be = bs+cipher.BlockSize(), be+cipher.BlockSize() {
		cipher.Encrypt(encrypted[bs:be], plain[bs:be])
	}

	return encrypted, nil

	//block, err := aes.NewCipher(fillKey(key))
	//if err != nil {
	//	return nil, err
	//}
	//
	//blockSize := block.BlockSize()
	//origData = getPaddingHandle(paddingType)(origData, blockSize)
	//
	////返回加密结果
	//encryptData := make([]byte, len(origData))
	////存储每次加密的数据
	//tmpData := make([]byte, blockSize)
	//
	////分组分块加密
	//for index := 0; index < len(origData); index += blockSize {
	//	block.Encrypt(tmpData, origData[index:index+blockSize])
	//	copy(encryptData, tmpData)
	//}
	//
	//return encryptData, nil
}

func (this ecbCrypto) DecryptBytes(crypted []byte, key string, paddingType padding) ([]byte, error) {
	block, err := aes.NewCipher(fillKey(key))
	if err != nil {
		return nil, err
	}

	blockSize := block.BlockSize()

	decrypted := make([]byte, len(crypted))
	for bs, be := 0, blockSize; bs < len(crypted); bs, be = bs+blockSize, be+blockSize {
		block.Decrypt(decrypted[bs:be], crypted[bs:be])
	}

	return getUnPaddingHandle(paddingType)(decrypted), nil
}
