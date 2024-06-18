/******************************************************************************
 * Author: liguoqiang
 * Date: 2024-04-15 22:25:47
 * LastEditors: liguoqiang
 * LastEditTime: 2024-04-17 09:20:34
 * Description:
********************************************************************************/
package common

import (
	"bytes"
	"crypto/aes"
	"encoding/base64"
	mylog "hjyserver/log"
)

func makeKey() string {
	key := "00000000000000000" + "hjy@123456"
	return key[len(key)-16:]
}
func DecryptData(data string) (string, error) {
	key := makeKey()
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
	// iv := origData[:blockSize]
	// origData = origData[blockSize:]
	// blockMode := cipher.NewCBCDecrypter(block, iv)
	// blockMode.CryptBlocks(origData, origData)
	result := PKCS7UnPadding(destData)
	return result, nil
}

func EncryptData(data string) (string, error) {
	key := makeKey()
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

func TestEncriptData() {
	data := "888888"
	result, err := EncryptData(data)
	if err != nil {
		mylog.Log.Errorln("TestEncriptData error", err)
		return
	}
	mylog.Log.Infoln("TestEncriptData success", result)
}
