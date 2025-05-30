/******************************************************************************
 * Author: liguoqiang
 * Date: 2024-09-11 11:15:44
 * LastEditors: liguoqiang
 * LastEditTime: 2024-12-16 11:42:26
 * Description:
********************************************************************************/
package tools

import (
	"fmt"
	"hjyserver/cfg"
	mylog "hjyserver/log"
	"hjyserver/mdb/mysql"
	"hjyserver/redis"
	"testing"
)

func TestSendMiniProgramSubscribeMessage(t *testing.T) {
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
	userId := int64(19)
	nickName := "微信用户"
	reportType := "日报告"
	mac := "f09e9e1f425a"
	score := 80
	startTime := "2021-08-01 00:00:00"
	endTime := "2021-08-01 23:59:59"
	status, msg := SendReportMsgToMiniProgram(userId, nickName, reportType, mac, score, startTime, endTime)
	fmt.Println("status:", status, "msg:", msg)
}

/******************************************************************************
 * function: 测试 发送日报告消息到公众号
 * description:
 * param {*testing.T} t
 * return {*}
********************************************************************************/
func TestSendDayReportMsgToOfficalAccount(t *testing.T) {
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
	userId := int64(19)
	nickName := "微信用户"
	mac := "f09e9e1f425a"
	score := 80
	startTime := "2021-08-01 00:00:00"
	endTime := "2021-08-01 23:59:59"
	status, msg := SendDayReportMsgToOfficalAccount(userId, nickName, mac, score, startTime, endTime)
	fmt.Println("status:", status, "msg:", msg)
}

/******************************************************************************
 * function: 测试次报告模板
 * description:
 * param {*testing.T} t
 * return {*}
********************************************************************************/
func TestSendEveryReportMsgToOfficalAccount(t *testing.T) {
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
	userId := int64(19)
	nickName := "微信用户"
	mac := "f09e9e1f425a"
	startTime := "2021-08-01 00:00:00"
	endTime := "2021-08-01 23:59:59"
	status, msg := SendEveryReportMsgToOfficalAccount(userId, nickName, mac, startTime, endTime)
	fmt.Println("status:", status, "msg:", msg)
}

/******************************************************************************
 * function: 测试设备在线模板
 * description:
 * param {*testing.T} t
 * return {*}
********************************************************************************/
func TestH03DeviceOnlineMsgToOfficalAccount(t *testing.T) {
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
	userId := int64(19)
	nickName := "微信用户"
	mac := "f09e9e1f425a"
	tm := "2024-12-03 14:13:00"
	status, msg := SendH03DeviceOnlineMsgToOfficalAccount(userId, nickName, mac, tm, "使用人已落座")
	fmt.Println("status:", status, "msg:", msg)
}

func TestWxMiniUrl(t *testing.T) {
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

	status, msg := GeneratorWxMiniUrl("", "")
	fmt.Println("status:", status, "msg:", msg)
}
