/******************************************************************************
 * Author: liguoqiang
 * Date: 2024-07-31 14:58:54
 * LastEditors: liguoqiang
 * LastEditTime: 2024-08-31 11:04:28
 * Description:
********************************************************************************/
package mdbwx

import (
	"encoding/json"
	"fmt"
	"hjyserver/cfg"
	mylog "hjyserver/log"
	"hjyserver/mdb/common"
	"hjyserver/mdb/mysql"
	"hjyserver/redis"
	wxtools "hjyserver/wx/tools"
	"testing"
)

func TestHttpGetWxUserBaseInfo(t *testing.T) {
	uri := "https://api.weixin.qq.com/cgi-bin/user/info"
	var params map[string]string = make(map[string]string)
	params["access_token"] = "82_kMxqn-85XgpQKGd0QKY2rQyYBIHTbAO0H-JCfBkECDDVUiw_TYVGUO1DoBMuK0-Kp_2yuZsWnymfktqqicHFzUjQPjOAdBQpJg2Ld7rZzC9SIsnW6qlbghcTdq4DOUjAGAPAG"
	params["openid"] = "o_YBY6jQkAH49234gmqhEpOsRVFI"
	params["lang"] = "zh_CN"
	result, err := common.HttpGet(uri, params)
	if err != nil {
		t.Errorf("HttpGetUserInfo failed, %v", err)
		return
	}

	userInfo := &wxtools.WxUserBaseInfo{}
	err = json.Unmarshal(result, &userInfo)
	if err != nil {
		t.Errorf("HttpGetUserInfo failed, %v", err)
		return
	}
	fmt.Println("result: ", result)
	fmt.Println("userInfo: ", userInfo)
}

func TestWxSssionKey(t *testing.T) {
	err := cfg.InitConfig("../../cfg/cfg.yml")
	if err != nil {
		t.Error("initialize config failed, ", err)
		return
	}
	mylog.Init()
	defer mylog.Close()
	code := "0b3Lvk0w3uCtg332Fo3w30FYZf0Lvk0M"
	session, err := getWxMiniSessionByCode(code)
	if err != nil {
		t.Error("getWxMiniSessionByCode failed, ", err)
		return
	}
	fmt.Println("session: ", session)
}

func TestWxGetPhoneNumber(t *testing.T) {
	err := cfg.InitConfig("../../cfg/cfg.yml")
	if err != nil {
		t.Error("initialize config failed, ", err)
		return
	}
	mylog.Init()
	defer mylog.Close()
	if !redis.InitRedis() {
		fmt.Println("init redis failed exit!")
		return
	}
	defer redis.CloseRedis()
	loginCode := "0f3sFvFa1NxvVH0XOuIa1J26Zk1sFvFp"
	session, err := getWxMiniSessionByCode(loginCode)
	if err != nil {
		t.Error("getWxMiniSessionByCode failed, ", err)
		return
	}
	phoneCode := "d4e17df134b50fff0f325602cb34cb6e31b35e6f7b20da477a9216f783e0c55c"
	openId := session.OpenId
	phoneInfo, err := getWxMiniPhoneNumber(phoneCode, openId)
	if err != nil {
		t.Error("getWxMiniPhoneNumber failed, ", err)
		return
	}
	fmt.Println("phoneInfo: ", phoneInfo)
}

func TestGetWxOfficalAccessToken(t *testing.T) {
	err := cfg.InitConfig("../../cfg/cfg.yml")
	if err != nil {
		t.Error("initialize config failed, ", err)
		return
	}
	mylog.Init()
	defer mylog.Close()
	if !redis.InitRedis() {
		fmt.Println("init redis failed exit!")
		return
	}
	defer redis.CloseRedis()
	accessToken, err := wxtools.QueryWxOfficalAccessToken()
	if err != nil {
		t.Error("getWxOfficalAccessToken failed, ", err)
		return
	}
	fmt.Println("accessToken: ", accessToken)
}

func TestSendReportCard(t *testing.T) {
	err := cfg.InitConfig("../../cfg/cfg.yml")
	if err != nil {
		t.Error("initialize config failed, ", err)
		return
	}
	mylog.Init()
	defer mylog.Close()
	if !redis.InitRedis() {
		fmt.Println("init redis failed exit!")
		return
	}
	defer redis.CloseRedis()
	mysql.Open()
	defer mysql.Close()
	wxtools.SendCustomerReportMsgToOfficalAccount(1, "test", "f09e9e1f42ca", "2024-08-24")
}

func TestDecriptSessionToJson(t *testing.T) {
	err := cfg.InitConfig("../../cfg/cfg.yml")
	if err != nil {
		t.Error("initialize config failed, ", err)
		return
	}
	mylog.Init()
	defer mylog.Close()
	if !redis.InitRedis() {
		fmt.Println("init redis failed exit!")
		return
	}
	defer redis.CloseRedis()
	jscode := "0c3ATUGa1bBWZH0FXgIa1CEQRA1ATUGn"
	session, err := getWxMiniSessionByCode(jscode)
	if err != nil {
		t.Errorf("TestDecriptSessionToJson error:%v", err)
		return
	}
	key, err := common.ConvertBase64ToBytes(session.SessionKey)
	if err != nil {
		t.Errorf("TestDecriptSessionToJson error:%v", err)
		return
	}
	encryptedData := "sabZL2HgnVjz5RsXgtxUnCGEXWIt3XrUTZ67JPg8D+n9hvfbAVSOtQdV4vfRll662ewpxUa0bcO+94m0ZpONOvKERPNEs9n1hU8jj6MH9ZGaNT5Ck6iBS0j9OBcoJqwYdaPZUST8MSUcnN7KcJpxQQB+xVBSf77ogSOjh5ZYmwJ9qJ7CCsGLeAIKRPM1Sc/NrJgw1N1NeYwbwJkMkT0b4A=="
	result, err := common.DecryptDataNoCBC(key, encryptedData)
	if err != nil {
		t.Errorf("TestDecriptSessionToJson error:%v", err)
		return
	}
	t.Errorf("TestDecriptSessionToJson success, %v", result)
}
