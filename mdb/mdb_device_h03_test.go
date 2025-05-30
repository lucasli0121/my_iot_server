/******************************************************************************
 * Author: liguoqiang
 * Date: 2024-08-20 13:05:02
 * LastEditors: liguoqiang
 * LastEditTime: 2024-10-25 20:12:16
 * Description:
********************************************************************************/
package mdb

import (
	"fmt"
	"hjyserver/cfg"
	mylog "hjyserver/log"
	"hjyserver/mdb/common"
	"hjyserver/mdb/mysql"
	"hjyserver/mq"
	"hjyserver/redis"
	"testing"
)

func TestH03CheckDayReport(t *testing.T) {
	err := cfg.InitConfig("../cfg/cfg.yml")
	if err != nil {
		t.Error("initialize config failed, ", err)
		return
	}
	mylog.Init()
	defer mylog.Close()
	mysql.Open()
	defer mysql.Close()
	// init mqtt object
	if !mq.InitMqtt() {
		mylog.Log.Error("init mqtt failed exit!")
		return
	}
	defer mq.CloseMqtt()
	checkDayReportTimer()
}

func TestSubscribeMessage(t *testing.T) {
	err := cfg.InitConfig("../cfg/cfg.yml")
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
	reportType := "次报告"
	mac := "f09e9e1f425a"
	score := 80
	startTime := "2021-08-01 00:00:00"
	endTime := "2021-08-01 23:59:59"

	notifyProc := &H03ReportNotifyProc{}
	status, msg := notifyProc.NotifyEveryReportToOfficalAccount(userId, nickName, mac, startTime, endTime)
	if status != common.Success {
		status, msg = notifyProc.NotifyToMiniProgram(userId, nickName, reportType, mac, score, startTime, endTime)
	}
	fmt.Println("status:", status, "msg:", msg)
}

func TestCalculateBeatUsers(t *testing.T) {
	val := CalculateBeatUsers(340)
	fmt.Println("val:", val)
	val = CalculateBeatUsers(100)
	fmt.Println("val:", val)
	val = CalculateBeatUsers(450)
	fmt.Println("val:", val)
	val = CalculateBeatUsers(500)
	fmt.Println("val:", val)
	val = CalculateBeatUsers(980)
	fmt.Println("val:", val)
	val = CalculateBeatUsers(1000)
	fmt.Println("val:", val)
	val = CalculateBeatUsers(0)
	fmt.Println("val:", val)
	val = CalculateBeatUsers(1)
	fmt.Println("val:", val)
}
