/******************************************************************************
 * Author: liguoqiang
 * Date: 2024-07-31 11:42:24
 * LastEditors: liguoqiang
 * LastEditTime: 2024-07-31 14:26:00
 * Description:
********************************************************************************/
package common

import (
	"fmt"
	"hjyserver/cfg"
	"testing"
)

func TestHttpPostJson(t *testing.T) {
	err := cfg.InitConfig("../../cfg/cfg.yml")
	if err != nil {
		t.Error("initialize config failed, ", err)
		return
	}
	uri := cfg.This.Wx.AccessTokenUri
	params := "{" +
		"\"grant_type\": \"client_credential\"," +
		"\"appid\": \"wxb44db5b8185402bd\"," +
		"\"secret\": \"bcfa116a917c6f660845ea401e12faac\"," +
		"\"force_refresh\": false" + "}"
	result, err := HttpPostJson(uri, []byte(params))
	if err != nil {
		t.Error("http post json failed, ", err)
		return
	}
	fmt.Println("result: ", string(result))
	t.Log("result: ", string(result))
}
