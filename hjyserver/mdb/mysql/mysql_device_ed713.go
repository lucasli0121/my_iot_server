package mysql

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"hjyserver/cfg"
	"hjyserver/exception"
	"hjyserver/gopool"
	mylog "hjyserver/log"
	"hjyserver/mdb/common"
	"hjyserver/mq"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

const (
	timeTopicPrefix          = "time/ED713/"
	timeTopicReplyPrefix     = "timeReply/ED713/"
	heartBeatTopicPrefix     = "heartBeat/ED713/"
	realDataTopicPrefix      = "currentData/ED713/"
	realDataTopicReplyPrefix = "currentDataReply/ED713/"
	dayReportTopicPrefix     = "dayReport/ED713/"
	eventTopicPrefix         = "event/ED713/"
)

func MakeEd713TimeTopic(mac string) string {
	mac = strings.ToLower(mac)
	return timeTopicPrefix + mac
}
func MakeEd713TimeReplayTopic(mac string) string {
	mac = strings.ToLower(mac)
	return timeTopicReplyPrefix + mac
}
func MakeEd713HeartBeatTopic(mac string) string {
	mac = strings.ToLower(mac)
	return heartBeatTopicPrefix + mac
}
func MakeEd713RealDataTopic(mac string) string {
	mac = strings.ToLower(mac)
	return realDataTopicPrefix + mac
}
func MakeEd713RealDataReplayTopic(mac string) string {
	mac = strings.ToLower(mac)
	return realDataTopicReplyPrefix + mac
}
func MakeEd713DayReportTopic(mac string) string {
	mac = strings.ToLower(mac)
	return dayReportTopicPrefix + mac
}
func MakeEd713EventTopic(mac string) string {
	mac = strings.ToLower(mac)
	return eventTopicPrefix + mac
}

func SubscribeEd713MqttTopic(mac string) {
	mac = strings.ToLower(mac)
	mylog.Log.Infoln("E713 SubscribeMqttTopic, mac:", mac)
	mq.SubscribeTopic(MakeEd713TimeTopic(mac), NewEd713MqttMsgProc())
	mq.SubscribeTopic(MakeEd713HeartBeatTopic(mac), NewEd713MqttMsgProc())
	mq.SubscribeTopic(MakeEd713RealDataReplayTopic(mac), NewEd713MqttMsgProc())
	mq.SubscribeTopic(MakeEd713DayReportTopic(mac), NewEd713MqttMsgProc())
	mq.SubscribeTopic(MakeEd713EventTopic(mac), NewEd713MqttMsgProc())
}

func SplitEd713MqttTopic(topic string) (string, string) {
	idx := strings.LastIndex(topic, "/")
	if idx != -1 {
		return topic[:idx+1], topic[idx+1:]
	}
	return "", ""
}

type Ed713MqttMsgProc struct {
}

func NewEd713MqttMsgProc() *Ed713MqttMsgProc {
	return &Ed713MqttMsgProc{}
}

func (me *Ed713MqttMsgProc) HandleMqttMsg(topic string, payload []byte) {
	prefix, mac := SplitEd713MqttTopic(topic)
	if prefix == "" || mac == "" {
		mylog.Log.Errorln("SplitEd713MqttTopic failed, topic:", topic)
		return
	}
	mylog.Log.Infoln("Ed713 HandleMqttMsg, topic:", topic, "prefix:", prefix, "mac:", mac, "payload:", string(payload))

	switch prefix {
	case timeTopicPrefix:
		handleTimeMqttMsg(mac, payload)
	case heartBeatTopicPrefix:
		handleHeartBeatMqttMsg(mac, payload)
	case realDataTopicReplyPrefix:
		handleRealDataMqttMsg(mac, payload)
	case dayReportTopicPrefix:
		handleDayReportMqttMsg(mac, payload)
	case eventTopicPrefix:
		handlerEventMqttMsg(mac, payload)
	}
}

type MsgHeader struct {
	Id  int `json:"id"`
	Ack int `json:"ack"`
}

type TimeRspMsg struct {
	MsgHeader
	Time int64 `json:"time"`
}

/******************************************************************************
 * function: handleTimeMqttMsg
 * description: handle time mqtt message
 * return {*}
********************************************************************************/
func handleTimeMqttMsg(mac string, payload []byte) {
	var msgHeader MsgHeader
	err := json.Unmarshal(payload, &msgHeader)
	if err != nil {
		mylog.Log.Errorln("handleTimeMqttMsg Unmarshal failed, err:", err)
		return
	}
	if msgHeader.Ack == 1 {
		// reply time
		var replyMsg TimeRspMsg
		replyMsg.Id = msgHeader.Id
		replyMsg.Ack = 0
		replyMsg.Time = time.Now().Unix()
		mq.PublishData(MakeEd713TimeReplayTopic(mac), replyMsg)
	}
}

/******************************************************************************
 * function: SetDeviceOnline
 * description:
 * param {string} mac
 * param {[]byte} payload
 * return {*}
********************************************************************************/
func handleHeartBeatMqttMsg(mac string, payload []byte) {
	SetDeviceOnline(mac, 1)

}

type Ed713RealDataJson struct {
	MsgHeader
	KeepPush        int   `json:"keep_push"`
	GetTime         int64 `json:"gettime"`
	HeartRate       []int `json:"heart_rate"`
	RespiratoryRate []int `json:"respiratory_rate"`
	BodyMovement    []int `json:"body_movement"`
	MoveState       []int `json:"move_state"`
	BodyStatus      []int `json:"body_status"`
	BodyPosition    []int `json:"body_position"`
	OnbedStatus     int   `json:"onbed_status"`
}

/******************************************************************************
 * function: handleRealDataMqttMsg
 * description: handle real data message from mqtt
 * param {string} mac
 * param {[]byte} payload
 * return {*}
********************************************************************************/
func handleRealDataMqttMsg(mac string, payload []byte) {
	var realDataJson Ed713RealDataJson
	err := json.Unmarshal(payload, &realDataJson)
	if err != nil {
		mylog.Log.Errorln("handleRealDataMqttMsg Unmarshal failed, err:", err)
		return
	}
	if len(realDataJson.HeartRate) == 0 {
		return
	}
	heartLen := len(realDataJson.HeartRate)
	totalHeartRate := 0
	totalRespRate := 0
	totalBodyMovement := 0
	totalMoveState := 0
	totalBodyStatus := 0
	totalBodyPosition := 0

	realDataSql := NewEd713RealDataMysql()
	realDataSql.Mac = mac
	realDataSql.CreateTime = common.SecondsToTimeStr(realDataJson.GetTime)
	realDataSql.OnbedStatus = realDataJson.OnbedStatus
	for i := 0; i < heartLen; i++ {
		totalHeartRate += realDataJson.HeartRate[i]
		totalRespRate += realDataJson.RespiratoryRate[i]
		totalBodyMovement += realDataJson.BodyMovement[i]
		totalMoveState += realDataJson.MoveState[i]
		totalBodyStatus += realDataJson.BodyStatus[i]
		totalBodyPosition += realDataJson.BodyPosition[i]
	}
	realDataSql.HeartRate = totalHeartRate / heartLen
	realDataSql.RespiratoryRate = totalRespRate / heartLen
	realDataSql.BodyMovement = totalBodyMovement / heartLen
	realDataSql.MoveState = totalMoveState / heartLen
	realDataSql.BodyStatus = totalBodyStatus / heartLen
	realDataSql.BodyPosition = totalBodyPosition / heartLen
	taskPool.Put(&gopool.Task{
		Params: []interface{}{realDataSql},
		Do: func(params ...interface{}) {
			var obj = params[0].(*Ed713RealDataMysql)
			obj.Insert()
		},
	})

	heartObj := &HeartRate{
		ID:           0,
		Mac:          mac,
		HeartRate:    totalHeartRate / heartLen,
		BreatheRate:  totalRespRate / heartLen,
		ActiveStatus: totalMoveState / heartLen,
		PersonStatus: totalBodyStatus / heartLen,
		PersonPos:    totalBodyPosition / heartLen,
		PhysicalRate: totalBodyMovement / heartLen,
		StagesStatus: 0,
		DateTime:     common.SecondsToTimeStr(realDataJson.GetTime),
	}
	if heartObj.HeartRate > 0 && heartObj.BreatheRate > 0 && heartObj.ActiveStatus > 0 {
		heartObj.PersonNum = 1
		heartObj.StagesStatus = 1
	} else {
		heartObj.PersonNum = 0
	}
	if heartObj.ActiveStatus == 3 && heartObj.PersonStatus == 2 {
		heartObj.StagesStatus = 2
	}
	if heartObj.ActiveStatus == 3 && heartObj.PersonStatus == 3 {
		heartObj.StagesStatus = 3
	}
	mq.PublishData(common.MakeHeartRateTopic(mac), heartObj)
}

// swagger:model Ed713RealDataMysql
type Ed713RealDataMysql struct {
	ID              int64  `json:"id" mysql:"id" binding:"omitempty"`
	Mac             string `json:"mac" mysql:"mac"`
	HeartRate       int    `json:"heart_rate" mysql:"heart_rate"`
	RespiratoryRate int    `json:"respiratory_rate" mysql:"respiratory_rate"`
	BodyMovement    int    `json:"body_movement" mysql:"body_movement"`
	MoveState       int    `json:"move_state" mysql:"move_state"`
	BodyStatus      int    `json:"body_status" mysql:"body_status"`
	BodyPosition    int    `json:"body_position" mysql:"body_position"`
	OnbedStatus     int    `json:"onbed_status" mysql:"onbed_status"`
	CreateTime      string `json:"create_time" mysql:"create_time" binding:"datetime=2006-01-02 15:04:05"`
}

func (me *Ed713RealDataMysql) ConvertEd713RealDataToHeartRate() *HeartRate {
	heartObj := &HeartRate{
		ID:           0,
		Mac:          me.Mac,
		HeartRate:    me.HeartRate,
		BreatheRate:  me.RespiratoryRate,
		ActiveStatus: me.MoveState,
		PersonStatus: me.BodyStatus,
		PersonPos:    me.BodyPosition,
		PhysicalRate: me.BodyMovement,
		StagesStatus: 0,
		DateTime:     me.CreateTime,
	}
	if heartObj.HeartRate > 0 && heartObj.BreatheRate > 0 && heartObj.ActiveStatus > 0 {
		heartObj.PersonNum = 1
	} else {
		heartObj.PersonNum = 0
	}
	if heartObj.HeartRate == 0 {
		heartObj.StagesStatus = 0
	} else {
		heartObj.StagesStatus = 1
	}
	return heartObj
}
func NewEd713RealDataMysql() *Ed713RealDataMysql {
	return &Ed713RealDataMysql{
		ID:              0,
		Mac:             "",
		HeartRate:       0,
		RespiratoryRate: 0,
		BodyMovement:    0,
		MoveState:       0,
		BodyStatus:      0,
		BodyPosition:    0,
		OnbedStatus:     0,
		CreateTime:      time.Now().Format(cfg.TmFmtStr),
	}
}
func QueryEd713RealDataByCond(filter interface{}, page *common.PageDao, sort interface{}, limited int, results *[]Ed713RealDataMysql) bool {
	res := false
	backFunc := func(rows *sql.Rows) {
		obj := NewEd713RealDataMysql()
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*results = append(*results, *obj)
		}
	}
	if page == nil {
		res = QueryDao(common.DeviceRecordTbl(Ed713Type), filter, sort, limited, backFunc)
	} else {
		res = QueryPage(common.DeviceRecordTbl(Ed713Type), page, filter, sort, backFunc)
	}
	return res
}

/******************************************************************************
 * function:
 * description:
 * param {interface{}} filter
 * param {*common.PageDao} page
 * param {interface{}} sort
 * param {int} limited
 * param {*[]HeartRate} results
 * return {*}
********************************************************************************/
func QueryEd713RealDataToHeartRateByCond(filter interface{}, page *common.PageDao, sort interface{}, limited int, results *[]HeartRate) bool {
	res := false
	backFunc := func(rows *sql.Rows) {
		obj := NewEd713RealDataMysql()
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*results = append(*results, *obj.ConvertEd713RealDataToHeartRate())
		}
	}
	if page == nil {
		res = QueryDao(common.DeviceRecordTbl(Ed713Type), filter, sort, limited, backFunc)
	} else {
		res = QueryPage(common.DeviceRecordTbl(Ed713Type), page, filter, sort, backFunc)
	}
	return res
}

func (me *Ed713RealDataMysql) DecodeFromGin(c *gin.Context) {
	err := c.ShouldBindBodyWith(me, binding.JSON)
	if err != nil {
		exception.Throw(http.StatusAccepted, err.Error())
	}
}

func (me *Ed713RealDataMysql) DecodeFromRows(rows *sql.Rows) error {
	return rows.Scan(
		&me.ID,
		&me.Mac,
		&me.HeartRate,
		&me.RespiratoryRate,
		&me.BodyMovement,
		&me.MoveState,
		&me.BodyStatus,
		&me.BodyPosition,
		&me.OnbedStatus,
		&me.CreateTime,
	)
}
func (me *Ed713RealDataMysql) DecodeFromRow(row *sql.Row) error {
	return row.Scan(
		&me.ID,
		&me.Mac,
		&me.HeartRate,
		&me.RespiratoryRate,
		&me.BodyMovement,
		&me.MoveState,
		&me.BodyStatus,
		&me.BodyPosition,
		&me.OnbedStatus,
		&me.CreateTime,
	)
}

func (me *Ed713RealDataMysql) QueryByID(id int64) bool {
	me.SetID(id)
	return QueryDaoByID(common.DeviceRecordTbl(Ed713Type), me.ID, me)
}

func (me *Ed713RealDataMysql) Insert() bool {
	tblName := common.DeviceRecordTbl(Ed713Type)
	if !CheckTableExist(tblName) {
		sql := `create table ` + tblName + ` (
			id MEDIUMINT NOT NULL AUTO_INCREMENT,
			mac varchar(32) not null comment '设备mac,与设备表关联',
			heart_rate int not null comment '心率',
			respiratory_rate int not null comment '呼吸率',
			body_movement int not null comment '体动',
			move_state int not null comment '移动状态',
			body_status int not null comment '体位状态',
			body_position int not null comment '体位',
			onbed_status int not null comment '在床状态',
			create_time datetime comment '新增日期',
			PRIMARY KEY (id, mac, create_time)
		)`
		CreateTable(sql)
	}
	return InsertDao(common.DeviceRecordTbl(Ed713Type), me)
}
func (me *Ed713RealDataMysql) Update() bool {
	return UpdateDaoByID(common.DeviceRecordTbl(Ed713Type), me.ID, me)
}
func (me *Ed713RealDataMysql) Delete() bool {
	return DeleteDaoByID(common.DeviceRecordTbl(Ed713Type), me.ID)
}

/*
设置ID
*/
func (me *Ed713RealDataMysql) SetID(id int64) {
	me.ID = id
}

type Ed713DayReportJson struct {
	MsgHeader
	SleepStart         int64   `json:"sleep_start"`
	SleepEnd           int64   `json:"sleep_end"`
	GoBed              int64   `json:"go_bed"`
	LeaveBed           int64   `json:"leave_bed"`
	SleepPeriodization []int64 `json:"sleep_periodization"`
	SleepEvents        []int64 `json:"sleep_events"`
	Evaluation         int     `json:"evaluation"`
	BaseRespiratory    int     `json:"base_respiratory"`
	BaseHeartRate      int     `json:"base_heart_rate"`
	BaseBodyMovement   int     `json:"base_body_movement"`
	Start              int64   `json:"start"`
	End                int64   `json:"end"`
	Sep                int     `json:"sep"`
	Respiratory        []int   `json:"respiratory"`
	HeartRate          []int   `json:"heart_rate"`
	BodyMovement       []int   `json:"body_movement"`
}

/******************************************************************************
 * function: handleDayReportMqttMsg
 * description: handle day report message from mqtt
 * param {string} mac
 * param {[]byte} payload
 * return {*}
********************************************************************************/
func TestEd713DayReport() {
	str := "{\"sleep_start\":1704218361,\"sleep_end\":1704255552,\"go_bed\":1704207665,\"leave_bed\":1704255552,\"evaluation\":63,\"base_respiratory\":12,\"base_heart_rate\":68,\"base_body_movement\":13,\"sleep_periodization\":[1,1704207665,2,1704218361,3,1704221121,2,1704221361,3,1704221721,2,1704221841,3,1704221961,2,1704222081,3,1704223521,2,1704224001,3,1704224961,2,1704226761,3,1704227001,2,1704227241,3,1704227361,2,1704227481,3,1704229521,2,1704229881,3,1704230121,2,1704231801,3,1704232041,2,1704233361,3,1704234321,2,1704234441,3,1704234561,2,1704234681,3,1704235401,2,1704235881,3,1704236601,2,1704237441,3,1704237681,2,1704238521,1,1704238744,2,1704239168,1,1704239951,2,1704240388,1,1704240486,2,1704240917,1,1704240966,2,1704241343,1,1704241512,2,1704242108,1,1704243178,2,1704243633,1,1704244333,2,1704244749,1,1704244893,2,1704245276,1,1704245321,2,1704245715,0,1704255552],\"sleep_events\":[1,1704211580,1,1704211738,1,1704213506,1,1704214960,1,1704216330,1,1704216398,1,1704221051,1,1704221475,1,1704221612,1,1704221684,1,1704221872,1,1704222830,1,1704222878,1,1704223177,1,1704223297,1,1704223349,1,1704223388,1,1704223427,1,1704223901,1,1704224081,1,1704224137,1,1704224194,1,1704224851,1,1704224898,1,1704226898,1,1704226931,1,1704227324,1,1704227585,1,1704228591,1,1704228874,1,1704229285,1,1704229426,1,1704230024,1,1704230080,1,1704231859,1,1704233456,1,1704233650,1,1704233940,1,1704233964,1,1704234158,1,1704235434,1,1704235812,1,1704236030,1,1704236095,1,1704236228,1,1704236294,1,1704236554,1,1704238681,2,1704238744,1,1704239337,2,1704239951,2,1704240486,2,1704240966,2,1704241512,2,1704243178,2,1704244333,2,1704244893,2,1704245321],\"start\":1704207665,\"end\":1704255552,\"sep\":191,\"respiratory\":[13,14,14,14,14,15,15,14,14,14,14,14,14,14,14,14,14,14,14,14,14,14,14,14,14,14,14,14,14,13,13,13,13,13,13,13,13,13,13,13,13,13,13,13,13,12,12,12,12,12,12,11,11,11,11,12,12,11,11,12,12,12,12,12,12,12,12,12,12,12,12,12,12,12,12,12,12,12,13,13,13,12,12,12,12,12,12,12,12,12,12,12,12,12,12,12,11,11,11,11,12,12,12,12,12,12,12,12,12,12,12,12,12,12,12,12,12,12,11,11,11,11,11,11,11,11,11,11,11,11,11,11,11,11,11,11,11,11,11,12,12,12,12,12,12,12,12,12,12,12,11,11,11,11,11,11,11,11,11,11,11,11,12,12,13,14,14,14,15,15,15,15,15,15,15,15,15,14,14,13,13,13,13,13,13,13,13,13,12,13,14,14,14,14,15,15,15,16,16,16,16,16,16,16,15,16,15,15,15,16,16,16,16,16,16,16,16,16,16,16,16,16,16,16,16,16,16,16,16,16,15,15,15,15,15,14,14,14,14,15,15,15,16,15,15,15,15,15,15,15],\"heart_rate\":[71,72,74,75,77,78,78,77,76,75,74,72,71,70,70,69,69,68,68,68,69,68,69,70,69,68,68,68,67,67,67,67,66,66,66,65,65,65,65,66,66,67,67,66,66,66,66,66,66,66,66,67,67,68,68,69,69,68,69,69,70,71,70,70,69,69,69,68,69,69,69,68,68,67,66,66,66,66,67,67,67,67,66,66,66,65,65,66,67,67,67,66,66,65,65,64,65,64,64,64,64,64,64,64,64,64,65,66,66,66,66,66,66,65,65,64,64,64,63,63,63,63,63,63,63,63,63,63,63,64,65,65,65,65,66,66,66,65,66,66,67,67,67,68,68,67,68,68,67,67,67,66,66,66,66,66,67,66,66,66,65,65,65,68,71,73,73,73,75,79,79,78,79,80,80,78,75,74,73,71,71,70,69,70,70,70,70,69,69,70,72,73,72,72,75,76,78,79,80,79,80,80,80,79,78,78,76,77,78,78,80,81,81,80,80,80,80,81,79,79,79,80,80,80,79,79,80,81,81,80,78,76,76,74,73,72,71,73,74,75,76,78,79,78,77,77,77,77,78,78],\"body_movement\":[42,26,8,0,0,0,8,19,36,13,2,0,0,0,0,0,0,0,0,0,12,13,0,0,0,0,0,0,0,0,10,0,0,0,0,0,0,6,2,0,0,0,0,0,6,20,0,0,0,0,0,0,0,0,0,12,10,5,0,0,0,0,0,0,0,0,0,0,0,9,10,2,18,5,17,0,0,0,0,10,5,17,17,0,5,8,6,0,0,8,7,0,0,0,0,0,0,0,0,0,22,0,5,3,3,0,0,0,0,8,4,4,13,14,0,0,9,6,0,0,0,0,0,0,0,0,9,8,0,0,0,0,0,0,14,10,14,16,4,0,0,0,0,0,2,7,0,5,17,10,13,3,0,0,0,0,19,0,0,0,0,0,10,0,0,2,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],\"id\":1704257505007,\"ack\":0}"
	handleDayReportMqttMsg("test", []byte(str))
}
func handleDayReportMqttMsg(mac string, payload []byte) {
	var reportData Ed713DayReportJson
	err := json.Unmarshal(payload, &reportData)
	if err != nil {
		mylog.Log.Errorln("handleDayReportMqttMsg Unmarshal failed, err:", err)
		return
	}
	exception.TryEx{
		Try: func() {
			var dayReportSql = NewEd713DayReportSql()
			dayReportSql.Mac = mac
			dayReportSql.SleepStartTime = common.SecondsToTimeStr(reportData.SleepStart)
			dayReportSql.SleepEndTime = common.SecondsToTimeStr(reportData.SleepEnd)
			dayReportSql.GoBedTime = common.SecondsToTimeStr(reportData.GoBed)
			dayReportSql.LeaveBedTime = common.SecondsToTimeStr(reportData.LeaveBed)
			dayReportSql.Evaluation = reportData.Evaluation
			dayReportSql.BaseRespiratory = reportData.BaseRespiratory
			dayReportSql.BaseHeartRate = reportData.BaseHeartRate
			dayReportSql.BaseBodyMovement = reportData.BaseBodyMovement
			dayReportSql.InBedStartTime = common.SecondsToTimeStr(reportData.Start)
			dayReportSql.InBedEndTime = common.SecondsToTimeStr(reportData.End)
			dayReportSql.InBedSep = reportData.Sep
			dayReportSql.CreateTime = time.Now().Format(cfg.TmFmtStr)
			for i := 0; i < len(reportData.SleepPeriodization); i += 2 {
				dayReportSql.SleepPeriodization = int(reportData.SleepPeriodization[i])
				dayReportSql.PeriodizationTime = common.SecondsToTimeStr(reportData.SleepPeriodization[i+1])
				if i < len(reportData.SleepEvents) {
					dayReportSql.SleepEvents = int(reportData.SleepEvents[i])
					dayReportSql.SleepEventsTime = common.SecondsToTimeStr(reportData.SleepEvents[i+1])
				}
				idx := i / 2
				if idx < len(reportData.Respiratory) {
					dayReportSql.Respiratory = int(reportData.Respiratory[idx])
					dayReportSql.HeartRate = int(reportData.HeartRate[idx])
					dayReportSql.BodyMovement = int(reportData.BodyMovement[idx])
				}
				dayReportSql.Insert()
			}
		},
		Catch: func(e exception.Exception) {
			mylog.Log.Errorln("handleDayReportMqttMsg catch exception, err:", e.Error())
		},
	}.Run()
}

type Ed713DayReportSql struct {
	ID                 int64  `json:"id" mysql:"id" binding:"omitempty"`
	Mac                string `json:"mac" mysql:"mac"`
	SleepStartTime     string `json:"sleep_start_time" mysql:"sleep_start_time"`
	SleepEndTime       string `json:"sleep_end_time" mysql:"sleep_end_time"`
	GoBedTime          string `json:"go_bed_time" mysql:"go_bed_time"`
	LeaveBedTime       string `json:"leave_bed_time" mysql:"leave_bed_time"`
	SleepPeriodization int    `json:"sleep_periodization" mysql:"sleep_periodization"`
	PeriodizationTime  string `json:"periodization_time" mysql:"periodization_time"`
	SleepEvents        int    `json:"sleep_events" mysql:"sleep_events"`
	SleepEventsTime    string `json:"sleep_events_time" mysql:"sleep_events_time"`
	Evaluation         int    `json:"evaluation" mysql:"evaluation"`
	BaseRespiratory    int    `json:"base_respiratory" mysql:"base_respiratory"`
	BaseHeartRate      int    `json:"base_heart_rate" mysql:"base_heart_rate"`
	BaseBodyMovement   int    `json:"base_body_movement" mysql:"base_body_movement"`
	InBedStartTime     string `json:"inbed_start_time" mysql:"inbed_start_time"`
	InBedEndTime       string `json:"inbed_end_time" mysql:"inbed_end_time"`
	InBedSep           int    `json:"inbed_sep" mysql:"inbed_sep"`
	Respiratory        int    `json:"respiratory" mysql:"respiratory"`
	HeartRate          int    `json:"heart_rate" mysql:"heart_rate"`
	BodyMovement       int    `json:"body_movement" mysql:"body_movement"`
	CreateTime         string `json:"create_time" mysql:"create_time" binding:"datetime=2006-01-02 15:04:05"`
}

func NewEd713DayReportSql() *Ed713DayReportSql {
	return &Ed713DayReportSql{
		ID:                 0,
		Mac:                "",
		SleepStartTime:     "",
		SleepEndTime:       "",
		GoBedTime:          "",
		LeaveBedTime:       "",
		SleepPeriodization: 0,
		PeriodizationTime:  "",
		SleepEvents:        0,
		SleepEventsTime:    "",
		Evaluation:         0,
		BaseRespiratory:    0,
		BaseHeartRate:      0,
		BaseBodyMovement:   0,
		InBedStartTime:     "",
		InBedEndTime:       "",
		InBedSep:           0,
		Respiratory:        0,
		HeartRate:          0,
		BodyMovement:       0,
		CreateTime:         time.Now().Format(cfg.TmFmtStr),
	}
}
func (me *Ed713DayReportSql) myTable() string {
	return common.DeviceDayReportTbl(Ed713Type)
}

func (me *Ed713DayReportSql) Update() bool {
	return UpdateDaoByID(me.myTable(), me.ID, me)
}
func (me *Ed713DayReportSql) SetID(id int64) {
	me.ID = id
}
func (me *Ed713DayReportSql) DecodeFromGin(c *gin.Context) {
	err := c.ShouldBindBodyWith(me, binding.JSON)
	if err != nil {
		exception.Throw(http.StatusAccepted, err.Error())
	}
}
func (me *Ed713DayReportSql) DecodeFromRows(rows *sql.Rows) error {
	return rows.Scan(
		&me.ID,
		&me.Mac,
		&me.SleepStartTime,
		&me.SleepEndTime,
		&me.GoBedTime,
		&me.LeaveBedTime,
		&me.SleepPeriodization,
		&me.PeriodizationTime,
		&me.SleepEvents,
		&me.SleepEventsTime,
		&me.Evaluation,
		&me.BaseRespiratory,
		&me.BaseHeartRate,
		&me.BaseBodyMovement,
		&me.InBedStartTime,
		&me.InBedEndTime,
		&me.InBedSep,
		&me.Respiratory,
		&me.HeartRate,
		&me.BodyMovement,
		&me.CreateTime,
	)
}
func (me *Ed713DayReportSql) DecodeFromRow(row *sql.Row) error {
	return row.Scan(
		&me.ID,
		&me.Mac,
		&me.SleepStartTime,
		&me.SleepEndTime,
		&me.GoBedTime,
		&me.LeaveBedTime,
		&me.SleepPeriodization,
		&me.PeriodizationTime,
		&me.SleepEvents,
		&me.SleepEventsTime,
		&me.Evaluation,
		&me.BaseRespiratory,
		&me.BaseHeartRate,
		&me.BaseBodyMovement,
		&me.InBedStartTime,
		&me.InBedEndTime,
		&me.InBedSep,
		&me.Respiratory,
		&me.HeartRate,
		&me.BodyMovement,
		&me.CreateTime,
	)
}

func (me *Ed713DayReportSql) QueryByID(id int64) bool {
	me.SetID(id)
	return QueryDaoByID(me.myTable(), me.ID, me)
}

func (me *Ed713DayReportSql) Insert() bool {
	tblName := me.myTable()
	if !CheckTableExist(tblName) {
		sql := `create table ` + tblName + ` (
			id MEDIUMINT NOT NULL AUTO_INCREMENT,
			mac varchar(32) not null comment '设备mac,与设备表关联',
			sleep_start_time datetime not null comment '睡眠开始时间',
			sleep_end_time datetime not null comment '睡眠结束时间',
			go_bed_time datetime not null comment '上床时间',
			leave_bed_time datetime not null comment '离床时间',
			sleep_periodization int not null comment '睡眠分期',
			periodization_time datetime comment '睡眠分期时间',
			sleep_events int not null comment '睡眠事件',
			sleep_events_time datetime comment '睡眠事件时间',
			evaluation int not null comment '睡眠评估',
			base_respiratory int not null comment '基础呼吸',
			base_heart_rate int not null comment '基础心率',
			base_body_movement int not null comment '基础体动',
			inbed_start_time datetime not null comment '在床开始时间',
			inbed_end_time datetime not null comment '在床结束时间',
			inbed_sep int not null comment '在床间隔',
			respiratory int not null comment '呼吸',
			heart_rate int not null comment '心率',
			body_movement int not null comment '体动',
			create_time datetime comment '新增日期',
			PRIMARY KEY (id, mac, create_time)
		)`
		CreateTable(sql)
	}
	return InsertDao(me.myTable(), me)
}

func (me *Ed713DayReportSql) Delete() bool {
	return DeleteDaoByID(me.myTable(), me.ID)
}

/******************************************************************************
 * function:
 * description: 根据mac地址以及开始，结束时间查询ed713的睡眠报告，只查询有分期的数据
 * param {string} mac
 * param {*} startTime
 * param {string} endTime
 * param {*[]Ed713DayReportSql} results
 * return {*}
********************************************************************************/
func QueryEd713DayReportByMacAndTime(mac string, startTime, endTime string, results *[]Ed713DayReportSql) bool {
	filter := fmt.Sprintf("mac='%s' and date(sleep_end_time)>=date('%s') and date(sleep_end_time)<=date('%s') and sleep_periodization > 0", mac, startTime, endTime)
	backFunc := func(rows *sql.Rows) {
		obj := NewEd713DayReportSql()
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*results = append(*results, *obj)
		}
	}
	return QueryDao(NewEd713DayReportSql().myTable(), filter, "periodization_time", -1, backFunc)
}
func QueryEd713DateListInReport(mac, startTime, endTime string, results *[]string) bool {
	filter := fmt.Sprintf("mac='%s' and date(sleep_end_time)>=date('%s') and date(sleep_end_time)<=date('%s') and sleep_periodization > 0", mac, startTime, endTime)
	sql := "select distinct date(sleep_end_time) from " + NewEd713DayReportSql().myTable() + " where " + filter
	sql += " order by date(sleep_end_time)"
	rows, err := mDb.Query(sql)
	if err != nil {
		mylog.Log.Errorln(err)
		return false
	}
	defer rows.Close()
	for rows.Next() {
		var date string
		err := rows.Scan(&date)
		if err != nil {
			mylog.Log.Errorln(err)
			return false
		}
		*results = append(*results, date)
	}
	return true
}

/******************************************************************************
 * function:
 * description:
 * return {*}
********************************************************************************/
type Ed713EventJson struct {
	MsgHeader
	Type           int `json:"type"`
	HeartRate      int `json:"heart_rate"`
	RepiratoryRate int `json:"respiratory_rate"`
}

func handlerEventMqttMsg(mac string, payload []byte) {
	var eventJson Ed713EventJson
	err := json.Unmarshal(payload, &eventJson)
	if err != nil {
		mylog.Log.Errorln("handlerEventMqttMsg Unmarshal failed, err:", err)
		return
	}
	eventSql := NewEd713EventSql()
	eventSql.Mac = mac
	eventSql.CreateTime = time.Now().Format(cfg.TmFmtStr)
	eventSql.Type = eventJson.Type
	eventSql.HeartRate = eventJson.HeartRate
	eventSql.RespiratoryRate = eventJson.RepiratoryRate
	taskPool.Put(&gopool.Task{
		Params: []interface{}{eventSql},
		Do: func(params ...interface{}) {
			var obj = params[0].(*Ed713EventSql)
			obj.Insert()
		},
	})

	var heartEvent = &HeartEvent{}
	heartEvent.Type = eventJson.Type
	heartEvent.HeartRate = eventJson.HeartRate
	heartEvent.RespiratoryRate = eventJson.RepiratoryRate
	heartEvent.CreateTime = eventSql.CreateTime
	var userDevices []UserDeviceDetail
	QueryUserDeviceDetailByMac(mac, &userDevices)
	if len(userDevices) > 0 {
		heartEvent.UserDeviceDetail = userDevices[0]
		mq.PublishData(common.MakeHeartEventTopic(mac), heartEvent)
		// send sms
		SendSleepAlarmSms(heartEvent)
	}
}

type Ed713EventSql struct {
	ID              int64  `json:"id" mysql:"id" binding:"omitempty"`
	Mac             string `json:"mac" mysql:"mac"`
	Type            int    `json:"type" mysql:"type"`
	HeartRate       int    `json:"heart_rate" mysql:"heart_rate"`
	RespiratoryRate int    `json:"respiratory_rate" mysql:"respiratory_rate"`
	CreateTime      string `json:"create_time" mysql:"create_time" binding:"datetime=2006-01-02 15:04:05"`
}

func NewEd713EventSql() *Ed713EventSql {
	return &Ed713EventSql{
		ID:              0,
		Mac:             "",
		Type:            0,
		HeartRate:       0,
		RespiratoryRate: 0,
		CreateTime:      time.Now().Format(cfg.TmFmtStr),
	}
}
func (me *Ed713EventSql) myTable() string {
	return common.DeviceEventTbl(Ed713Type)
}
func (me *Ed713EventSql) Update() bool {
	return UpdateDaoByID(me.myTable(), me.ID, me)
}
func (me *Ed713EventSql) SetID(id int64) {
	me.ID = id
}
func (me *Ed713EventSql) DecodeFromGin(c *gin.Context) {
	err := c.ShouldBindBodyWith(me, binding.JSON)
	if err != nil {
		exception.Throw(http.StatusAccepted, err.Error())
	}
}
func (me *Ed713EventSql) DecodeFromRows(rows *sql.Rows) error {
	return rows.Scan(
		&me.ID,
		&me.Mac,
		&me.Type,
		&me.HeartRate,
		&me.RespiratoryRate,
		&me.CreateTime,
	)
}
func (me *Ed713EventSql) DecodeFromRow(row *sql.Row) error {
	return row.Scan(
		&me.ID,
		&me.Mac,
		&me.Type,
		&me.HeartRate,
		&me.RespiratoryRate,
		&me.CreateTime,
	)
}
func (me *Ed713EventSql) QueryByID(id int64) bool {
	me.SetID(id)
	return QueryDaoByID(me.myTable(), me.ID, me)
}
func (me *Ed713EventSql) Insert() bool {
	tblName := me.myTable()
	if !CheckTableExist(tblName) {
		sql := `create table ` + tblName + ` (
			id MEDIUMINT NOT NULL AUTO_INCREMENT,
			mac varchar(32) not null comment '设备mac,与设备表关联',
			type int not null comment '事件类型',
			heart_rate int not null comment '心率',
			respiratory_rate int not null comment '呼吸率',
			create_time datetime comment '新增日期',
			PRIMARY KEY (id, mac, create_time)
		)`
		CreateTable(sql)
	}
	return InsertDao(me.myTable(), me)
}
func (me *Ed713EventSql) Delete() bool {
	return DeleteDaoByID(me.myTable(), me.ID)
}
func (me *Ed713EventSql) QueryEventByMacAndTime(mac string, startTime, endTime string, results *[]Ed713EventSql) bool {
	filter := fmt.Sprintf("mac='%s' and create_time>='%s' and create_time<='%s'", mac, startTime, endTime)
	backFunc := func(rows *sql.Rows) {
		obj := NewEd713EventSql()
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*results = append(*results, *obj)
		}
	}
	return QueryDao(me.myTable(), filter, nil, -1, backFunc)
}

func AskEd713RealData(mac string, freq int, keepPush int) {
	type AskData struct {
		MsgHeader
		DeadLine int64 `json:"deadline"`
		Freq     int   `json:"frequency"`
		KeepPush int   `json:"keep_push"`
	}
	var askMsg *AskData = &AskData{}
	askMsg.Id = 1
	askMsg.Ack = 0
	askMsg.DeadLine = time.Now().Add(time.Duration(30) * time.Second).Unix()
	askMsg.Freq = freq
	askMsg.KeepPush = keepPush
	// msg, _ := json.Marshal(askMsg)
	// mylog.Log.Infoln("AskEd713RealData, mq, mac:", mac, "msg:", string(msg))
	mq.PublishData(MakeEd713RealDataTopic(mac), askMsg)
}
