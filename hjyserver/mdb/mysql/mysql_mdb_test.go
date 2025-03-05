/******************************************************************************
 * Author: liguoqiang
 * Date: 2024-08-06 20:27:22
 * LastEditors: liguoqiang
 * LastEditTime: 2024-08-06 23:56:13
 * Description:
********************************************************************************/
package mysql

import (
	"hjyserver/cfg"
	mylog "hjyserver/log"
	"hjyserver/redis"
	"testing"
)

func TestGetUserToken(t *testing.T) {
	err := cfg.InitConfig("../../cfg/cfg.yml")
	if err != nil {
		t.Error("initialize config failed, ", err)
		return
	}
	mylog.Init()
	defer mylog.Close()
	redis.InitRedis()
	defer redis.CloseRedis()
	user := NewUser()
	user.SetID(1)
	token, _ := GetUserToken(user)
	t.Log("token:", token)
}

func TestVerifyUserToken(t *testing.T) {
	err := cfg.InitConfig("../../cfg/cfg.yml")
	if err != nil {
		t.Error("initialize config failed, ", err)
		return
	}
	mylog.Init()
	defer mylog.Close()
	redis.InitRedis()
	defer redis.CloseRedis()
	token := "wcLchXqPCSVAJyCBd7leQ5rxhQ4O3euQZlMXk60pKuM="
	result := VerifyUserToken(token)
	t.Log("result:", result)
}
