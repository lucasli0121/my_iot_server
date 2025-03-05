/******************************************************************************
 * Author: liguoqiang
 * Date: 2024-04-15 22:25:47
 * LastEditors: liguoqiang
 * LastEditTime: 2024-11-21 14:31:12
 * Description:
********************************************************************************/
package common

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	mylog "hjyserver/log"
)

func makeKey() string {
	key := "00000000000000000" + "hjy@123456"
	return key[len(key)-16:]
}
func ConvertBase64ToBytes(data string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(data)
}
func ConvertBytesToBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

/******************************************************************************
 * function:
 * description:
 * param {string} data
 * return {*}
********************************************************************************/
func DecryptDataNoCBCWithDefaultkey(data string) (string, error) {
	key := makeKey()
	return DecryptDataNoCBC([]byte(key), data)
}
func DecryptDataNoCBC(key []byte, data string) (string, error) {
	baseStr, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		mylog.Log.Errorln(err)
		return "", err
	}
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		mylog.Log.Errorln(err)
		return "", err
	}
	blockSize := block.BlockSize()
	origData := []byte(baseStr)
	destData := make([]byte, len(origData))
	for i := 0; i < len(origData); i += blockSize {
		block.Decrypt(destData[i:i+blockSize], origData[i:i+blockSize])
	}
	result := PKCS7UnPadding(destData)
	return result, nil
}
func DecryptDataWithCBC(key []byte, iv []byte, data string) (string, error) {
	baseStr, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		mylog.Log.Errorln(err)
		return "", err
	}
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		mylog.Log.Errorln(err)
		return "", err
	}
	origData := []byte(baseStr)
	if iv == nil {
		blockSize := block.BlockSize()
		iv = origData[:blockSize]
		origData = origData[blockSize:]
	}
	blockMode := cipher.NewCBCDecrypter(block, iv)
	blockMode.CryptBlocks(origData, origData)
	result := PKCS7UnPadding(origData)
	return result, nil
}

func EncryptDataWithDefaultkey(data string) (string, error) {
	key := makeKey()
	return EncryptData([]byte(key), data)
}

func EncryptData(key []byte, data string) (string, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		mylog.Log.Errorln(err)
		return "", err
	}
	blockSize := block.BlockSize()
	origData := []byte(data)
	origData = PKCS7Padding(origData, blockSize)
	destData := make([]byte, len(origData))
	for i := 0; i < len(origData); i += blockSize {
		block.Encrypt(destData[i:i+blockSize], origData[i:i+blockSize])
	}
	// iv := origData[:blockSize]
	// origData = origData[blockSize:]
	// blockMode := cipher.NewCBCEncrypter(block, iv)
	// blockMode.CryptBlocks(origData, origData)
	baseStr := base64.StdEncoding.EncodeToString(destData)
	return baseStr, nil
}

func PKCS7UnPadding(origData []byte) string {
	length := len(origData)
	unpadding := int(origData[length-1])
	return string(origData[:(length - unpadding)])
}

func PKCS7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

/******************************************************************************
 * function: GenerateMQPassword
 * description: 采用HMAC-SHA1算法生成MQTT连接密码
 * param {*} clientId
 * param {string} secretKey
 * return {*}
********************************************************************************/
func GenerateMQPassword(clientId, secretKey string) (string, error) {
	// 创建一个新的 HMAC 使用 SHA1 哈希算法
	h := hmac.New(sha1.New, []byte(secretKey))

	// 写入待签名字符串
	h.Write([]byte(clientId))

	// 计算 HMAC-SHA1 签名
	signature := h.Sum(nil)

	// 对签名结果进行 Base64 编码
	password := base64.StdEncoding.EncodeToString(signature)

	return password, nil
}
