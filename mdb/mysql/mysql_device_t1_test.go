/******************************************************************************
 * Author: liguoqiang
 * Date: 2024-08-02 15:57:02
 * LastEditors: liguoqiang
 * LastEditTime: 2025-03-08 22:49:04
 * Description:
********************************************************************************/
package mysql

import (
	"database/sql"
	"fmt"
	"hjyserver/cfg"
	mylog "hjyserver/log"
	"hjyserver/mq"
	"hjyserver/redis"
	"testing"
)

func TestT1SyncRequest(t *testing.T) {
	err := cfg.InitConfig("../../cfg/cfg.yml")
	if err != nil {
		t.Error("initialize config failed, ", err)
		return
	}
	mylog.Init()
	defer mylog.Close()
	Open()
	defer Close()
	// init mqtt object
	if !mq.InitMqtt() {
		mylog.Log.Error("init mqtt failed exit!")
		return
	}
	defer mq.CloseMqtt()
	T1SyncRequest("test111")
}

func TestT1RebootRequest(t *testing.T) {
	err := cfg.InitConfig("../../cfg/cfg.yml")
	if err != nil {
		t.Error("initialize config failed, ", err)
		return
	}
	mylog.Init()
	defer mylog.Close()
	Open()
	defer Close()
	// init mqtt object
	if !mq.InitMqtt() {
		mylog.Log.Error("init mqtt failed exit!")
		return
	}
	defer mq.CloseMqtt()
	T1RebootRequest("test111", 100)
}

func TestT1AttrRequest(t *testing.T) {
	err := cfg.InitConfig("../../cfg/cfg.yml")
	if err != nil {
		t.Error("initialize config failed, ", err)
		return
	}
	mylog.Init()
	defer mylog.Close()
	Open()
	defer Close()
	// init mqtt object
	if !mq.InitMqtt() {
		mylog.Log.Error("init mqtt failed exit!")
		return
	}
	defer mq.CloseMqtt()
	T1AttrRequest("test111", "all")
}

func TestT1SettingRequest(t *testing.T) {
	err := cfg.InitConfig("../../cfg/cfg.yml")
	if err != nil {
		t.Error("initialize config failed, ", err)
		return
	}
	mylog.Init()
	defer mylog.Close()
	Open()
	defer Close()
	// init mqtt object
	if !mq.InitMqtt() {
		mylog.Log.Error("init mqtt failed exit!")
		return
	}
	defer mq.CloseMqtt()
	setting := &T1Setting{
		SetNlMode:       1,
		SetNlBrightness: 100,
		SetBlMode:       1,
		SetBlBrightness: 100,
		SetBlDelay:      10,
		SetAlarmMode:    1,
		SetAlarmTime: []int{
			10,
			30,
		},
		SetAlarmVol:    10,
		SetGestureMode: 0,
	}
	T1SettingRequest("test111", setting)
}

func TestT1Ota(t *testing.T) {
	err := cfg.InitConfig("../../cfg/cfg.yml")
	if err != nil {
		t.Error("initialize config failed, ", err)
		return
	}
	mylog.Init()
	defer mylog.Close()
	Open()
	defer Close()
	T1SyncRequest("d83bda831716")
}

func TestT1Attr(t *testing.T) {
	err := cfg.InitConfig("../../cfg/cfg.yml")
	if err != nil {
		t.Error("initialize config failed, ", err)
		return
	}
	mylog.Init()
	defer mylog.Close()
	Open()
	defer Close()
	redis.InitRedis()
	defer redis.CloseRedis()
	mq.InitMqtt()
	defer mq.CloseMqtt()
	mqStr := "{" +
		"\"cmd\": 104," +
		"\"s\":  3927," +
		"\"time\": 1718185943," +
		"\"id\": \"d83bdaa605d2\"," +
		"\"data\": {" +
		"\"respiratory\": 15," +
		"\"heart_rate\": 73," +
		"\"body_movement\": 255," +
		"\"body_angle\": 30," +
		"\"body_distance\": 40," +
		"\"flow_state\":9," +
		"\"position_interval\":4," +
		"\"nl_mode\":1," +
		"\"nl_brightness\":255," +
		"\"bl_mode\":1," +
		"\"bl_brightness\":255," +
		"\"bl_delay\": 60," +
		"\"alarm_mode\": 2," +
		"\"alarm_time\": [13,30]," +
		"\"alarm_vol\":220," +
		"\"gesture_mode\": 1," +
		"\"study_time\": [78,13,60,230]" +
		"}" +
		"}"
	HandleT1MqttMsg("hjy-dev/t1/attr/d83bda831716", []byte(mqStr))
}

func TestT1Event(t *testing.T) {
	err := cfg.InitConfig("../../cfg/cfg.yml")
	if err != nil {
		t.Error("initialize config failed, ", err)
		return
	}
	mylog.Init()
	defer mylog.Close()
	Open()
	defer Close()
	redis.InitRedis()
	defer redis.CloseRedis()
	mq.InitMqtt()
	defer mq.CloseMqtt()
	mqStr := "{" +
		"\"cmd\": 105," +
		"\"s\":  3927," +
		"\"time\": 1718185943," +
		"\"id\": \"d83bdaa605d2\"," +
		"\"data\": {" +
		"\"body_status\": 1," +
		"\"posture_state\": 3," +
		"\"activity_freq\": 2," +
		"\"warning_event\": 0," +
		"\"alarm_rang\": 1" +
		"}" +
		"}"
	HandleT1MqttMsg("hjy-dev/t1/event/d83bda831716", []byte(mqStr))
}

func TestT1WarningEvent(t *testing.T) {
	err := cfg.InitConfig("../../cfg/cfg.yml")
	if err != nil {
		t.Error("initialize config failed, ", err)
		return
	}
	mylog.Init()
	defer mylog.Close()
	Open()
	defer Close()
	redis.InitRedis()
	defer redis.CloseRedis()
	mq.InitMqtt()
	defer mq.CloseMqtt()
	mqStr := "{" +
		"\"cmd\": 105," +
		"\"s\":  3927," +
		"\"time\": 1718185943," +
		"\"id\": \"00000000\"," +
		"\"data\": {" +
		"\"body_status\": 1," +
		"\"posture_state\": 3," +
		"\"activity_freq\": 2," +
		"\"warning_event\": 1" +
		"}" +
		"}"
	HandleT1MqttMsg("hjy-dev/t1/event/00000000", []byte(mqStr))
}

func TestT1WeekReport(t *testing.T) {
	err := cfg.InitConfig("../../cfg/cfg.yml")
	if err != nil {
		t.Error("initialize config failed, ", err)
		return
	}
	mylog.Init()
	defer mylog.Close()
	Open()
	defer Close()
	var studyReportList []T1StudyReport
	// filter := fmt.Sprintf("mac='%s'  and date(end_time) >= '%s' and date(end_time) <= '%s'", "ccba9706727a", "2025-01-06", "2025-01-12")
	QueryDao(NewT1StudyReport().TableName(), nil, nil, -1, func(rows *sql.Rows) {
		obj := NewT1StudyReport()
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			studyReportList = append(studyReportList, *obj)
		}
	})
	for _, v := range studyReportList {
		dailyReport := StatT1DailyReport(&v)
		StatT1WeekReport(dailyReport)
	}
}

func TestT1QueryWeekReport(t *testing.T) {
	err := cfg.InitConfig("../../cfg/cfg.yml")
	if err != nil {
		t.Error("initialize config failed, ", err)
		return
	}
	mylog.Init()
	defer mylog.Close()
	Open()
	defer Close()
	var weekReportList []T1WeekReport
	QueryT1WeekReportByMac("ccba9706727a", "2024-12-30", &weekReportList)

	for _, v := range weekReportList {
		fmt.Println(v)
	}
}
