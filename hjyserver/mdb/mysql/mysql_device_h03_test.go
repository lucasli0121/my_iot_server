/******************************************************************************
 * Author: liguoqiang
 * Date: 2024-08-02 15:57:02
 * LastEditors: liguoqiang
 * LastEditTime: 2025-01-14 23:36:58
 * Description:
********************************************************************************/
package mysql

import (
	"database/sql"
	"fmt"
	"hjyserver/cfg"
	mylog "hjyserver/log"
	"hjyserver/mdb/common"
	"hjyserver/mq"
	"hjyserver/redis"
	"testing"
)

func TestH03SyncRequest(t *testing.T) {
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
	H03SyncRequest("test111")
}

func TestH03RebootRequest(t *testing.T) {
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
	H03RebootRequest("test111", 100)
}

func TestH03AttrRequest(t *testing.T) {
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
	H03AttrRequest("test111", "all")
}

func TestH03SettingRequest(t *testing.T) {
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
	setting := &H03Setting{
		SetOnoffStatus: 0,
		SetControlMode: 1,
		SetColorTemp:   25,
		SetDelayTime:   100,
		SetGestureMode: 1,
	}
	H03SettingRequest("test111", setting)
}

func TestH03Ota(t *testing.T) {
	err := cfg.InitConfig("../../cfg/cfg.yml")
	if err != nil {
		t.Error("initialize config failed, ", err)
		return
	}
	mylog.Init()
	defer mylog.Close()
	Open()
	defer Close()
	H03SyncRequest("d83bda831716")
}

func TestH03Attr(t *testing.T) {
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
		"\"onoff_status\": 1," +
		"\"control_mode\": 0," +
		"\"brightness_val\": 64," +
		"\"color_temp\": 127," +
		"\"delay_time\": 10," +
		"\"gesture_mode\": 1," +
		"\"study_time\": [78,13,60,230]" +
		"}" +
		"}"
	HandleH03MqttMsg("hjy-dev/h03/attr/d83bda831716", []byte(mqStr))
}

func TestH03Event(t *testing.T) {
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
		"\"flow_state\": 3," +
		"\"posture_state\": 3," +
		"\"activity_freq\": 2," +
		"\"warning_event\": 0" +
		"}" +
		"}"
	HandleH03MqttMsg("hjy-dev/h03/event/d83bda831716", []byte(mqStr))
}

func TestH03WarningEvent(t *testing.T) {
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
		"\"flow_state\": 3," +
		"\"posture_state\": 3," +
		"\"activity_freq\": 2," +
		"\"warning_event\": 1" +
		"}" +
		"}"
	HandleH03MqttMsg("hjy-dev/h03/event/00000000", []byte(mqStr))
}

func TestYearWeek(t *testing.T) {
	ts, err := common.StrToTime("2024-11-01 10:10:45")
	if err != nil {
		t.Error("str to time failed, ", err)
		return
	}
	year, week := ts.ISOWeek()
	fmt.Printf("year: %d, week: %d, weekday:%d", year, week, ts.Weekday())
}

func TestWeekReport(t *testing.T) {
	err := cfg.InitConfig("../../cfg/cfg.yml")
	if err != nil {
		t.Error("initialize config failed, ", err)
		return
	}
	mylog.Init()
	defer mylog.Close()
	Open()
	defer Close()
	var studyReportList []H03StudyReport
	// filter := fmt.Sprintf("mac='%s'  and date(end_time) >= '%s' and date(end_time) <= '%s'", "ccba9706727a", "2025-01-06", "2025-01-12")
	QueryDao(NewH03StudyReport().TableName(), nil, nil, -1, func(rows *sql.Rows) {
		obj := NewH03StudyReport()
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			studyReportList = append(studyReportList, *obj)
		}
	})
	for _, v := range studyReportList {
		dailyReport := StatH03DailyReport(&v)
		StatH03WeekReport(dailyReport)
	}
}

func TestQueryWeekReport(t *testing.T) {
	err := cfg.InitConfig("../../cfg/cfg.yml")
	if err != nil {
		t.Error("initialize config failed, ", err)
		return
	}
	mylog.Init()
	defer mylog.Close()
	Open()
	defer Close()
	var weekReportList []H03WeekReport
	QueryH03WeekReportByMac("ccba9706727a", "2024-12-30", &weekReportList)

	for _, v := range weekReportList {
		fmt.Println(v)
	}
}
