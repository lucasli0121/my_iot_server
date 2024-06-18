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
	timeSyncCmd             = 100
	timeSyncRspCmd          = 200
	heartBeatCmd            = 101
	realDataSetCmd          = 102
	realDataSetRspCmd       = 202
	dayReportCmd            = 103
	eventReportCmd          = 104
	setLightDelayCmd        = 105
	getLightDelayCmd        = 106
	cleanEventCmd           = 108
	sleepSwitchCmd          = 109
	nurseModelCmd           = 110
	improveDisturbedCmd     = 111
	breathAbnormalCmd       = 120
	breathAbnormalSwitchCmd = 121
	lightDelayRspCmd        = 206
	otaCmd                  = 107
	otaRspCmd               = 207
)

const (
	timeX1TopicPrefix               = "time/X1/"
	timeX1TopicReplyPrefix          = "timeReply/X1/"
	heartBeatX1TopicPrefix          = "heartBeat/X1/"
	realDataX1TopicPrefix           = "currentData/X1/"
	realDataX1TopicReplyPrefix      = "currentDataReply/X1/"
	dayReportX1TopicPrefix          = "dayReport/X1/"
	eventX1TopicPrefix              = "event/X1/"
	setLedTopicPrefix               = "setLedData/X1/"
	getLedTopicPrefix               = "getLedData/X1/"
	ledReplyTopicPrefix             = "replyLedData/X1/"
	ackVersionTopicPrefix           = "ackVersion/X1/"
	replyVersionTopicPrefix         = "replyVersion/X1/"
	cleanX1EventTopicPrefix         = "cleanEvent/X1/"
	sleepX1SwitchTopicPrefix        = "sleepSwitch/X1/"
	nurseModelTopicPrefix           = "nursingModel/X1/"
	breathAbnormalTopicPrefix       = "breathAbnormal/X1/"
	breathAbnormalSwitchTopicPrefix = "breathAbnormalSwitch/X1/"
	improveDisturbedTopicPrefix     = "improveDisturbSwitch/X1/"
)

func MakeX1TimeTopic(mac string) string {
	mac = strings.ToLower(mac)
	return timeX1TopicPrefix + mac
}
func MakeX1TimeReplayTopic(mac string) string {
	mac = strings.ToLower(mac)
	return timeX1TopicReplyPrefix + mac
}
func MakeX1HeartBeatTopic(mac string) string {
	mac = strings.ToLower(mac)
	return heartBeatX1TopicPrefix + mac
}
func MakeX1RealDataTopic(mac string) string {
	mac = strings.ToLower(mac)
	return realDataX1TopicPrefix + mac
}
func MakeX1RealDataReplayTopic(mac string) string {
	mac = strings.ToLower(mac)
	return realDataX1TopicReplyPrefix + mac
}
func MakeX1DayReportTopic(mac string) string {
	mac = strings.ToLower(mac)
	return dayReportX1TopicPrefix + mac
}
func MakeX1EventTopic(mac string) string {
	mac = strings.ToLower(mac)
	return eventX1TopicPrefix + mac
}
func MakeSetX1LedTopic(mac string) string {
	mac = strings.ToLower(mac)
	return setLedTopicPrefix + mac
}
func MakeGetX1LedTopic(mac string) string {
	mac = strings.ToLower(mac)
	return getLedTopicPrefix + mac
}
func MakeX1LedReplyTopic(mac string) string {
	mac = strings.ToLower(mac)
	return ledReplyTopicPrefix + mac
}
func MakeAckX1VersionTopic(mac string) string {
	mac = strings.ToLower(mac)
	return ackVersionTopicPrefix + mac
}
func MakeX1ReplyVersionTopic(mac string) string {
	mac = strings.ToLower(mac)
	return replyVersionTopicPrefix + mac
}
func MakeX1CleanEventTopic(mac string) string {
	mac = strings.ToLower(mac)
	return cleanX1EventTopicPrefix + mac
}
func MakeX1SleepSwitchTopic(mac string) string {
	mac = strings.ToLower(mac)
	return sleepX1SwitchTopicPrefix + mac
}

func MakeX1NurseModelTopic(mac string) string {
	mac = strings.ToLower(mac)
	return nurseModelTopicPrefix + mac
}
func MakeX1BreathAbnormalTopic(mac string) string {
	mac = strings.ToLower(mac)
	return breathAbnormalTopicPrefix + mac
}
func MakeX1BreathAbnormalSwitchTopic(mac string) string {
	mac = strings.ToLower(mac)
	return breathAbnormalSwitchTopicPrefix + mac
}
func MakeX1ImproveDisturbedTopic(mac string) string {
	return improveDisturbedTopicPrefix + strings.ToLower(mac)
}

func SubscribeX1MqttTopic(mac string) {
	mac = strings.ToLower(mac)
	mylog.Log.Infoln("X1 SubscribeMqttTopic, mac:", mac)
	mq.SubscribeTopic(MakeX1TimeTopic(mac), NewX1MqttMsgProc())
	mq.SubscribeTopic(MakeX1HeartBeatTopic(mac), NewX1MqttMsgProc())
	mq.SubscribeTopic(MakeX1RealDataReplayTopic(mac), NewX1MqttMsgProc())
	mq.SubscribeTopic(MakeX1DayReportTopic(mac), NewX1MqttMsgProc())
	mq.SubscribeTopic(MakeX1EventTopic(mac), NewX1MqttMsgProc())
	mq.SubscribeTopic(MakeX1LedReplyTopic(mac), NewX1MqttMsgProc())
	mq.SubscribeTopic(MakeAckX1VersionTopic(mac), NewX1MqttMsgProc())
}

func SplitX1MqttTopic(topic string) (string, string) {
	idx := strings.LastIndex(topic, "/")
	if idx != -1 {
		return topic[:idx+1], topic[idx+1:]
	}
	return "", ""
}

type X1MqttMsgProc struct {
}

func NewX1MqttMsgProc() *X1MqttMsgProc {
	return &X1MqttMsgProc{}
}

func (me *X1MqttMsgProc) HandleMqttMsg(topic string, payload []byte) {
	prefix, mac := SplitX1MqttTopic(topic)
	if prefix == "" || mac == "" {
		mylog.Log.Errorln("SplitX1MqttTopic failed, topic:", topic)
		return
	}
	mylog.Log.Infoln("X1 HandleMqttMsg, topic:", topic, "prefix:", prefix, "mac:", mac, "payload:", string(payload))

	switch prefix {
	case timeX1TopicPrefix:
		handleX1TimeMqttMsg(mac, payload)
	case heartBeatX1TopicPrefix:
		handleX1HeartBeatMqttMsg(mac, payload)
	case realDataX1TopicReplyPrefix:
		handleX1RealDataMqttMsg(mac, payload)
	case dayReportX1TopicPrefix:
		handleX1DayReportMqttMsg(mac, payload)
	case eventX1TopicPrefix:
		handlerX1EventMqttMsg(mac, payload)
	case ledReplyTopicPrefix:
		handleX1LedReplyMqttMsg(mac, payload)
	case ackVersionTopicPrefix:
		handleX1AckVersionMqttMsg(mac, payload)
	}
}

type X1MsgHeader struct {
	Id  int `json:"id"`
	Ack int `json:"ack"`
}

type X1TimeRspMsg struct {
	X1MsgHeader
	Time int64 `json:"time"`
}

/******************************************************************************
 * function: handleX1TimeMqttMsg
 * description: handle time mqtt message
 * return {*}
********************************************************************************/
func handleX1TimeMqttMsg(mac string, payload []byte) {
	var msgHeader X1MsgHeader
	err := json.Unmarshal(payload, &msgHeader)
	if err != nil {
		mylog.Log.Errorln("handleTimeMqttMsg Unmarshal failed, err:", err)
		return
	}
	if msgHeader.Ack == 1 {
		// reply time
		var replyMsg TimeRspMsg
		replyMsg.Id = timeSyncRspCmd
		replyMsg.Ack = 0
		replyMsg.Time = time.Now().Unix()
		mq.PublishData(MakeX1TimeReplayTopic(mac), replyMsg)
	}
}

/******************************************************************************
 * function: handleX1HeartBeatMqttMsg
 * description:
 * param {string} mac
 * param {[]byte} payload
 * return {*}
********************************************************************************/
func handleX1HeartBeatMqttMsg(mac string, payload []byte) {
	SetDeviceOnline(mac, 1)
}

type X1RealDataJson struct {
	X1MsgHeader
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
 * function: handleX1RealDataMqttMsg
 * description: handle real data message from mqtt
 * param {string} mac
 * param {[]byte} payload
 * return {*}
********************************************************************************/
func handleX1RealDataMqttMsg(mac string, payload []byte) {
	var realDataJson X1RealDataJson
	err := json.Unmarshal(payload, &realDataJson)
	if err != nil {
		mylog.Log.Errorln("handleX1RealDataMqttMsg Unmarshal failed, err:", err)
		return
	}
	if len(realDataJson.HeartRate) == 0 {
		return
	}
	realDataOrigin := NewX1RealDataOrigin()
	realDataOrigin.Mac = mac
	realDataOrigin.Value = string(payload)
	realDataOrigin.CreateTime = time.Now().Format(cfg.TmFmtStr)
	taskPool.Put(&gopool.Task{
		Params: []interface{}{realDataOrigin},
		Do: func(params ...interface{}) {
			var obj = params[0].(*X1RealDataOrigin)
			obj.Insert()
		},
	})

	totalHeartRate := 0
	totalRespRate := 0
	totalBodyMovement := 0
	totalMoveState := 0
	totalBodyStatus := 0
	totalBodyPosition := 0

	realDataSql := NewX1RealDataMysql()
	realDataSql.Mac = mac
	// realDataSql.CreateTime = common.SecondsToTimeStr(realDataJson.GetTime)
	realDataSql.OnbedStatus = realDataJson.OnbedStatus
	for i := 0; i < len(realDataJson.HeartRate); i++ {
		totalHeartRate += realDataJson.HeartRate[i]
	}
	for i := 0; i < len(realDataJson.RespiratoryRate); i++ {
		totalRespRate += realDataJson.RespiratoryRate[i]
	}
	for i := 0; i < len(realDataJson.BodyMovement); i++ {
		totalBodyMovement += realDataJson.BodyMovement[i]
	}
	for i := 0; i < len(realDataJson.MoveState); i++ {
		totalMoveState += realDataJson.MoveState[i]
	}
	for i := 0; i < len(realDataJson.BodyStatus); i++ {
		totalBodyStatus += realDataJson.BodyStatus[i]
	}
	for i := 0; i < len(realDataJson.BodyPosition); i++ {
		totalBodyPosition += realDataJson.BodyPosition[i]
	}
	realDataSql.HeartRate = totalHeartRate / len(realDataJson.HeartRate)
	if len(realDataJson.RespiratoryRate) > 0 {
		realDataSql.RespiratoryRate = totalRespRate / len(realDataJson.RespiratoryRate)
	}
	if len(realDataJson.BodyMovement) > 0 {
		realDataSql.BodyMovement = totalBodyMovement / len(realDataJson.BodyMovement)
	}
	if len(realDataJson.MoveState) > 0 {
		realDataSql.MoveState = totalMoveState / len(realDataJson.MoveState)
	}
	if len(realDataJson.BodyStatus) > 0 {
		realDataSql.BodyStatus = totalBodyStatus / len(realDataJson.BodyStatus)
	}
	if len(realDataJson.BodyPosition) > 0 {
		realDataSql.BodyPosition = totalBodyPosition / len(realDataJson.BodyPosition)
	}
	taskPool.Put(&gopool.Task{
		Params: []interface{}{realDataSql},
		Do: func(params ...interface{}) {
			var obj = params[0].(*X1RealDataMysql)
			obj.Insert()
		},
	})

	heartObj := &HeartRate{
		ID:           0,
		Mac:          mac,
		HeartRate:    realDataSql.HeartRate,
		BreatheRate:  realDataSql.RespiratoryRate,
		ActiveStatus: realDataSql.MoveState,
		PersonStatus: realDataSql.BodyStatus,
		PersonPos:    realDataSql.BodyPosition,
		PhysicalRate: realDataSql.BodyMovement,
		StagesStatus: 0,
		DateTime:     common.GetNowTime(), //common.SecondsToTimeStr(realDataJson.GetTime),
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

// swagger:model X1RealDataMysql
type X1RealDataMysql struct {
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

func NewX1RealDataMysql() *X1RealDataMysql {
	return &X1RealDataMysql{
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
func QueryX1RealDataByCond(filter interface{}, page *common.PageDao, sort interface{}, limited int, results *[]X1RealDataMysql) bool {
	res := false
	backFunc := func(rows *sql.Rows) {
		obj := NewX1RealDataMysql()
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*results = append(*results, *obj)
		}
	}
	if page == nil {
		res = QueryDao(common.DeviceRecordTbl(X1Type), filter, sort, limited, backFunc)
	} else {
		res = QueryPage(common.DeviceRecordTbl(X1Type), page, filter, sort, backFunc)
	}
	return res
}

/******************************************************************************
 * function: QueryX1RealDataToHeartRateByCond
 * description:
 * param {interface{}} filter
 * param {*common.PageDao} page
 * param {interface{}} sort
 * param {int} limited
 * param {*[]HeartRate} results
 * return {*}
********************************************************************************/
func QueryX1RealDataToHeartRateByCond(filter interface{}, page *common.PageDao, sort interface{}, limited int, results *[]HeartRate) bool {
	res := false
	backFunc := func(rows *sql.Rows) {
		obj := NewX1RealDataMysql()
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*results = append(*results, *obj.ConvertX1RealDataToHeartRate())
		}
	}
	if page == nil {
		res = QueryDao(common.DeviceRecordTbl(X1Type), filter, sort, limited, backFunc)
	} else {
		res = QueryPage(common.DeviceRecordTbl(X1Type), page, filter, sort, backFunc)
	}
	return res
}

/******************************************************************************
 * function: ConvertX1RealDataToHeartRate
 * description:
 * return {*}
********************************************************************************/
func (me *X1RealDataMysql) ConvertX1RealDataToHeartRate() *HeartRate {
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
func (me *X1RealDataMysql) DecodeFromGin(c *gin.Context) {
	err := c.ShouldBindBodyWith(me, binding.JSON)
	if err != nil {
		exception.Throw(http.StatusAccepted, err.Error())
	}
}

func (me *X1RealDataMysql) DecodeFromRows(rows *sql.Rows) error {
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
func (me *X1RealDataMysql) DecodeFromRow(row *sql.Row) error {
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

func (me *X1RealDataMysql) QueryByID(id int64) bool {
	me.SetID(id)
	return QueryDaoByID(common.DeviceRecordTbl(X1Type), me.ID, me)
}

func (me *X1RealDataMysql) Insert() bool {
	tblName := common.DeviceRecordTbl(X1Type)
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
	return InsertDao(common.DeviceRecordTbl(X1Type), me)
}
func (me *X1RealDataMysql) Update() bool {
	return UpdateDaoByID(common.DeviceRecordTbl(X1Type), me.ID, me)
}
func (me *X1RealDataMysql) Delete() bool {
	return DeleteDaoByID(common.DeviceRecordTbl(X1Type), me.ID)
}

/*
设置ID
*/
func (me *X1RealDataMysql) SetID(id int64) {
	me.ID = id
}

type X1DayReportJson struct {
	X1MsgHeader
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
 * function: handleX1DayReportMqttMsg
 * description: handle day report message from mqtt
 * param {string} mac
 * param {[]byte} payload
 * return {*}
********************************************************************************/
func TestX1DayReport() {
	str := "{\"sleep_start\":1704218361,\"sleep_end\":1704255552,\"go_bed\":1704207665,\"leave_bed\":1704255552,\"evaluation\":63,\"base_respiratory\":12,\"base_heart_rate\":68,\"base_body_movement\":13,\"sleep_periodization\":[1,1704207665,2,1704218361,3,1704221121,2,1704221361,3,1704221721,2,1704221841,3,1704221961,2,1704222081,3,1704223521,2,1704224001,3,1704224961,2,1704226761,3,1704227001,2,1704227241,3,1704227361,2,1704227481,3,1704229521,2,1704229881,3,1704230121,2,1704231801,3,1704232041,2,1704233361,3,1704234321,2,1704234441,3,1704234561,2,1704234681,3,1704235401,2,1704235881,3,1704236601,2,1704237441,3,1704237681,2,1704238521,1,1704238744,2,1704239168,1,1704239951,2,1704240388,1,1704240486,2,1704240917,1,1704240966,2,1704241343,1,1704241512,2,1704242108,1,1704243178,2,1704243633,1,1704244333,2,1704244749,1,1704244893,2,1704245276,1,1704245321,2,1704245715,0,1704255552],\"sleep_events\":[1,1704211580,1,1704211738,1,1704213506,1,1704214960,1,1704216330,1,1704216398,1,1704221051,1,1704221475,1,1704221612,1,1704221684,1,1704221872,1,1704222830,1,1704222878,1,1704223177,1,1704223297,1,1704223349,1,1704223388,1,1704223427,1,1704223901,1,1704224081,1,1704224137,1,1704224194,1,1704224851,1,1704224898,1,1704226898,1,1704226931,1,1704227324,1,1704227585,1,1704228591,1,1704228874,1,1704229285,1,1704229426,1,1704230024,1,1704230080,1,1704231859,1,1704233456,1,1704233650,1,1704233940,1,1704233964,1,1704234158,1,1704235434,1,1704235812,1,1704236030,1,1704236095,1,1704236228,1,1704236294,1,1704236554,1,1704238681,2,1704238744,1,1704239337,2,1704239951,2,1704240486,2,1704240966,2,1704241512,2,1704243178,2,1704244333,2,1704244893,2,1704245321],\"start\":1704207665,\"end\":1704255552,\"sep\":191,\"respiratory\":[13,14,14,14,14,15,15,14,14,14,14,14,14,14,14,14,14,14,14,14,14,14,14,14,14,14,14,14,14,13,13,13,13,13,13,13,13,13,13,13,13,13,13,13,13,12,12,12,12,12,12,11,11,11,11,12,12,11,11,12,12,12,12,12,12,12,12,12,12,12,12,12,12,12,12,12,12,12,13,13,13,12,12,12,12,12,12,12,12,12,12,12,12,12,12,12,11,11,11,11,12,12,12,12,12,12,12,12,12,12,12,12,12,12,12,12,12,12,11,11,11,11,11,11,11,11,11,11,11,11,11,11,11,11,11,11,11,11,11,12,12,12,12,12,12,12,12,12,12,12,11,11,11,11,11,11,11,11,11,11,11,11,12,12,13,14,14,14,15,15,15,15,15,15,15,15,15,14,14,13,13,13,13,13,13,13,13,13,12,13,14,14,14,14,15,15,15,16,16,16,16,16,16,16,15,16,15,15,15,16,16,16,16,16,16,16,16,16,16,16,16,16,16,16,16,16,16,16,16,16,15,15,15,15,15,14,14,14,14,15,15,15,16,15,15,15,15,15,15,15],\"heart_rate\":[71,72,74,75,77,78,78,77,76,75,74,72,71,70,70,69,69,68,68,68,69,68,69,70,69,68,68,68,67,67,67,67,66,66,66,65,65,65,65,66,66,67,67,66,66,66,66,66,66,66,66,67,67,68,68,69,69,68,69,69,70,71,70,70,69,69,69,68,69,69,69,68,68,67,66,66,66,66,67,67,67,67,66,66,66,65,65,66,67,67,67,66,66,65,65,64,65,64,64,64,64,64,64,64,64,64,65,66,66,66,66,66,66,65,65,64,64,64,63,63,63,63,63,63,63,63,63,63,63,64,65,65,65,65,66,66,66,65,66,66,67,67,67,68,68,67,68,68,67,67,67,66,66,66,66,66,67,66,66,66,65,65,65,68,71,73,73,73,75,79,79,78,79,80,80,78,75,74,73,71,71,70,69,70,70,70,70,69,69,70,72,73,72,72,75,76,78,79,80,79,80,80,80,79,78,78,76,77,78,78,80,81,81,80,80,80,80,81,79,79,79,80,80,80,79,79,80,81,81,80,78,76,76,74,73,72,71,73,74,75,76,78,79,78,77,77,77,77,78,78],\"body_movement\":[42,26,8,0,0,0,8,19,36,13,2,0,0,0,0,0,0,0,0,0,12,13,0,0,0,0,0,0,0,0,10,0,0,0,0,0,0,6,2,0,0,0,0,0,6,20,0,0,0,0,0,0,0,0,0,12,10,5,0,0,0,0,0,0,0,0,0,0,0,9,10,2,18,5,17,0,0,0,0,10,5,17,17,0,5,8,6,0,0,8,7,0,0,0,0,0,0,0,0,0,22,0,5,3,3,0,0,0,0,8,4,4,13,14,0,0,9,6,0,0,0,0,0,0,0,0,9,8,0,0,0,0,0,0,14,10,14,16,4,0,0,0,0,0,2,7,0,5,17,10,13,3,0,0,0,0,19,0,0,0,0,0,10,0,0,2,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],\"id\":1704257505007,\"ack\":0}"
	handleX1DayReportMqttMsg("test", []byte(str))
}
func handleX1DayReportMqttMsg(mac string, payload []byte) {
	var reportData X1DayReportJson
	err := json.Unmarshal(payload, &reportData)
	if err != nil {
		mylog.Log.Errorln("handleDayReportMqttMsg Unmarshal failed, err:", err)
		return
	}
	dayReportOrigin := NewX1DayReportOrigin()
	dayReportOrigin.Mac = mac
	dayReportOrigin.Value = string(payload)
	dayReportOrigin.CreateTime = time.Now().Format(cfg.TmFmtStr)
	if !dayReportOrigin.Insert() {
		mylog.Log.Errorln("handleDayReportMqttMsg insert day report origin failed")
	}
	exception.TryEx{
		Try: func() {
			var dayReportSql = NewX1DayReportSql()
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

type X1OriginJson struct {
	ID         int64  `json:"id" mysql:"id" common:"id" binding:"required"`
	Mac        string `json:"mac" mysql:"mac" size:"32" common:"mac" binding:"required"`
	Value      string `json:"value" mysql:"value" size:"4096" `
	CreateTime string `json:"create_time" mysql:"create_time" binding:"datetime=2006-01-02 15:04:05" `
}

// swagger:model X1DayReportOrigin
type X1DayReportOrigin X1OriginJson

func NewX1DayReportOrigin() *X1DayReportOrigin {
	return &X1DayReportOrigin{
		ID:         0,
		Mac:        "",
		Value:      "",
		CreateTime: time.Now().Format(cfg.TmFmtStr),
	}
}
func (me *X1DayReportOrigin) myTable() string {
	return common.DeviceDayReportJsonTbl(X1Type)
}

func (me *X1DayReportOrigin) DecodeFromGin(c *gin.Context) {
	err := c.ShouldBindBodyWith(me, binding.JSON)
	if err != nil {
		exception.Throw(http.StatusAccepted, err.Error())
	}
}
func (me *X1DayReportOrigin) DecodeFromRows(rows *sql.Rows) error {
	return rows.Scan(
		&me.ID,
		&me.Mac,
		&me.Value,
		&me.CreateTime,
	)
}
func (me *X1DayReportOrigin) DecodeFromRow(row *sql.Row) error {
	return row.Scan(
		&me.ID,
		&me.Mac,
		&me.Value,
		&me.CreateTime,
	)
}
func (me *X1DayReportOrigin) QueryByID(id int64) bool {
	me.ID = id
	return QueryDaoByID(me.myTable(), me.ID, me)
}

func (me *X1DayReportOrigin) Insert() bool {
	if !CheckTableExist(me.myTable()) {
		CreateTableWithStruct(me.myTable(), me)
	}
	return InsertDao(me.myTable(), me)
}
func (me *X1DayReportOrigin) Update() bool {
	return UpdateDaoByID(me.myTable(), me.ID, me)
}
func (me *X1DayReportOrigin) Delete() bool {
	return DeleteDaoByID(me.myTable(), me.ID)
}
func (me *X1DayReportOrigin) SetID(id int64) {
	me.ID = id
}
func QueryX1DayReportJson(mac string, create_date string, result *[]X1DayReportOrigin) bool {
	filter := fmt.Sprintf("mac='%s' and date(create_time)=date('%s')", mac, create_date)
	return QueryDao(common.DeviceDayReportTbl(X1Type), filter, nil, 0, func(rows *sql.Rows) {
		obj := NewX1DayReportOrigin()
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*result = append(*result, *obj)
		}
	})
}

// swagger:model X1RealDataOrigin
type X1RealDataOrigin X1OriginJson

func NewX1RealDataOrigin() *X1RealDataOrigin {
	return &X1RealDataOrigin{
		ID:         0,
		Mac:        "",
		Value:      "",
		CreateTime: time.Now().Format(cfg.TmFmtStr),
	}
}
func (me *X1RealDataOrigin) myTable() string {
	return common.DeviceRecordJsonTbl(X1Type)
}
func (me *X1RealDataOrigin) DecodeFromGin(c *gin.Context) {
	err := c.ShouldBindBodyWith(me, binding.JSON)
	if err != nil {
		exception.Throw(http.StatusAccepted, err.Error())
	}
}
func (me *X1RealDataOrigin) DecodeFromRows(rows *sql.Rows) error {
	return rows.Scan(
		&me.ID,
		&me.Mac,
		&me.Value,
		&me.CreateTime,
	)
}
func (me *X1RealDataOrigin) DecodeFromRow(row *sql.Row) error {
	return row.Scan(
		&me.ID,
		&me.Mac,
		&me.Value,
		&me.CreateTime,
	)
}
func (me *X1RealDataOrigin) QueryByID(id int64) bool {
	me.ID = id
	return QueryDaoByID(me.myTable(), me.ID, me)
}
func (me *X1RealDataOrigin) Insert() bool {
	if !CheckTableExist(me.myTable()) {
		CreateTableWithStruct(me.myTable(), me)
	}
	return InsertDao(me.myTable(), me)
}
func (me *X1RealDataOrigin) Update() bool {
	return UpdateDaoByID(me.myTable(), me.ID, me)
}
func (me *X1RealDataOrigin) Delete() bool {
	return DeleteDaoByID(me.myTable(), me.ID)
}
func (me *X1RealDataOrigin) SetID(id int64) {
	me.ID = id
}

func QueryX1RealDataJson(mac string, create_date string, result *[]X1RealDataOrigin) bool {
	filter := fmt.Sprintf("mac='%s' and date(create_time)=date('%s')", mac, create_date)
	return QueryDao(common.DeviceRecordJsonTbl(X1Type), filter, nil, 0, func(rows *sql.Rows) {
		obj := NewX1RealDataOrigin()
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*result = append(*result, *obj)
		}
	})
}

type X1DayReportSql struct {
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

func NewX1DayReportSql() *X1DayReportSql {
	return &X1DayReportSql{
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
func (me *X1DayReportSql) myTable() string {
	return common.DeviceDayReportTbl(X1Type)
}

func (me *X1DayReportSql) Update() bool {
	return UpdateDaoByID(me.myTable(), me.ID, me)
}
func (me *X1DayReportSql) SetID(id int64) {
	me.ID = id
}
func (me *X1DayReportSql) DecodeFromGin(c *gin.Context) {
	err := c.ShouldBindBodyWith(me, binding.JSON)
	if err != nil {
		exception.Throw(http.StatusAccepted, err.Error())
	}
}
func (me *X1DayReportSql) DecodeFromRows(rows *sql.Rows) error {
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
func (me *X1DayReportSql) DecodeFromRow(row *sql.Row) error {
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

func (me *X1DayReportSql) QueryByID(id int64) bool {
	me.SetID(id)
	return QueryDaoByID(me.myTable(), me.ID, me)
}

func (me *X1DayReportSql) Insert() bool {
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

func (me *X1DayReportSql) Delete() bool {
	return DeleteDaoByID(me.myTable(), me.ID)
}

func QueryX1DayReportByMacAndTime(mac string, startTime, endTime string, results *[]X1DayReportSql) bool {
	filter := fmt.Sprintf("mac='%s' and date(sleep_end_time)>=date('%s') and date(sleep_end_time)<=date('%s') and sleep_periodization > 0", mac, startTime, endTime)
	backFunc := func(rows *sql.Rows) {
		obj := NewX1DayReportSql()
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*results = append(*results, *obj)
		}
	}
	return QueryDao(NewX1DayReportSql().myTable(), filter, "periodization_time", -1, backFunc)
}

/******************************************************************************
 * function:
 * description:
 * param {*} mac
 * param {*} startTime
 * param {string} endTime
 * param {*[]string} results
 * return {*}
********************************************************************************/
func QueryX1DateListInReport(mac, startTime, endTime string, results *[]string) bool {
	filter := fmt.Sprintf("mac='%s' and date(sleep_end_time)>=date('%s') and date(sleep_end_time)<=date('%s') and sleep_periodization > 0", mac, startTime, endTime)
	sql := "select distinct date(sleep_end_time) from " + NewX1DayReportSql().myTable() + " where " + filter
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
type X1EventJson struct {
	X1MsgHeader
	Type           int `json:"type"`
	HeartRate      int `json:"heart_rate"`
	RepiratoryRate int `json:"respiratory_rate"`
}

func handlerX1EventMqttMsg(mac string, payload []byte) {
	var eventJson X1EventJson
	err := json.Unmarshal(payload, &eventJson)
	if err != nil {
		mylog.Log.Errorln("handlerEventMqttMsg Unmarshal failed, err:", err)
		return
	}
	eventSql := NewX1EventSql()
	eventSql.Mac = mac
	eventSql.CreateTime = time.Now().Format(cfg.TmFmtStr)
	eventSql.Type = eventJson.Type
	eventSql.HeartRate = eventJson.HeartRate
	eventSql.RespiratoryRate = eventJson.RepiratoryRate
	taskPool.Put(&gopool.Task{
		Params: []interface{}{eventSql},
		Do: func(params ...interface{}) {
			var obj = params[0].(*X1EventSql)
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

type X1EventSql struct {
	ID              int64  `json:"id" mysql:"id" binding:"omitempty"`
	Mac             string `json:"mac" mysql:"mac"`
	Type            int    `json:"type" mysql:"type"`
	HeartRate       int    `json:"heart_rate" mysql:"heart_rate"`
	RespiratoryRate int    `json:"respiratory_rate" mysql:"respiratory_rate"`
	CreateTime      string `json:"create_time" mysql:"create_time" binding:"datetime=2006-01-02 15:04:05"`
}

func NewX1EventSql() *X1EventSql {
	return &X1EventSql{
		ID:              0,
		Mac:             "",
		Type:            0,
		HeartRate:       0,
		RespiratoryRate: 0,
		CreateTime:      time.Now().Format(cfg.TmFmtStr),
	}
}
func (me *X1EventSql) myTable() string {
	return common.DeviceEventTbl(X1Type)
}
func (me *X1EventSql) Update() bool {
	return UpdateDaoByID(me.myTable(), me.ID, me)
}
func (me *X1EventSql) SetID(id int64) {
	me.ID = id
}
func (me *X1EventSql) DecodeFromGin(c *gin.Context) {
	err := c.ShouldBindBodyWith(me, binding.JSON)
	if err != nil {
		exception.Throw(http.StatusAccepted, err.Error())
	}
}
func (me *X1EventSql) DecodeFromRows(rows *sql.Rows) error {
	return rows.Scan(
		&me.ID,
		&me.Mac,
		&me.Type,
		&me.HeartRate,
		&me.RespiratoryRate,
		&me.CreateTime,
	)
}
func (me *X1EventSql) DecodeFromRow(row *sql.Row) error {
	return row.Scan(
		&me.ID,
		&me.Mac,
		&me.Type,
		&me.HeartRate,
		&me.RespiratoryRate,
		&me.CreateTime,
	)
}
func (me *X1EventSql) QueryByID(id int64) bool {
	me.SetID(id)
	return QueryDaoByID(me.myTable(), me.ID, me)
}
func (me *X1EventSql) Insert() bool {
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
func (me *X1EventSql) Delete() bool {
	return DeleteDaoByID(me.myTable(), me.ID)
}
func (me *X1EventSql) QueryEventByMacAndTime(mac string, startTime, endTime string, results *[]X1EventSql) bool {
	filter := fmt.Sprintf("mac='%s' and create_time>='%s' and create_time<='%s'", mac, startTime, endTime)
	backFunc := func(rows *sql.Rows) {
		obj := NewX1EventSql()
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*results = append(*results, *obj)
		}
	}
	return QueryDao(me.myTable(), filter, nil, -1, backFunc)
}

func AskX1RealData(mac string, freq int, keepPush int) {
	type AskData struct {
		X1MsgHeader
		DeadLine int64 `json:"deadline"`
		Freq     int   `json:"frequency"`
		KeepPush int   `json:"keep_push"`
	}
	var askMsg *AskData = &AskData{}
	askMsg.Id = realDataSetCmd
	askMsg.Ack = 1
	askMsg.DeadLine = time.Now().Add(time.Duration(30) * time.Second).Unix()
	askMsg.Freq = freq
	askMsg.KeepPush = keepPush
	// msg, _ := json.Marshal(askMsg)
	// mylog.Log.Infoln("AskX1RealData, mq, mac:", mac, "msg:", string(msg))
	mq.PublishData(MakeX1RealDataTopic(mac), askMsg)
}

/******************************************************************************
 * function:
 * description:
 * param {string} mac
 * param {[]byte} payload
 * return {*}
********************************************************************************/
type X1LedJson struct {
	X1MsgHeader
	DelayTs int64 `json:"delay_ts"`
}

func handleX1LedReplyMqttMsg(mac string, payload []byte) {
	var ledJson X1LedJson
	err := json.Unmarshal(payload, &ledJson)
	if err != nil {
		mylog.Log.Errorln("handleX1LedReplyMqttMsg Unmarshal failed, err:", err)
		return
	}
	ledSql := NewX1LedSql()
	ledSql.Mac = mac
	ledSql.DelayTs = ledJson.DelayTs
	ledSql.Insert()
}

type X1LedSql struct {
	ID         int64  `json:"id" mysql:"id" binding:"omitempty"`
	Mac        string `json:"mac" mysql:"mac"`
	DelayTs    int64  `json:"delay_ts" mysql:"delay_ts"`
	CreateTime string `json:"create_time" mysql:"create_time" binding:"datetime=2006-01-02 15:04:05"`
}

func NewX1LedSql() *X1LedSql {
	return &X1LedSql{
		ID:         0,
		Mac:        "",
		DelayTs:    0,
		CreateTime: time.Now().Format(cfg.TmFmtStr),
	}
}
func (me *X1LedSql) myTable() string {
	return common.DeviceLedTbl(X1Type)
}
func (me *X1LedSql) Update() bool {
	return UpdateDaoByID(me.myTable(), me.ID, me)
}
func (me *X1LedSql) SetID(id int64) {
	me.ID = id
}
func (me *X1LedSql) DecodeFromGin(c *gin.Context) {
	err := c.ShouldBindBodyWith(me, binding.JSON)
	if err != nil {
		exception.Throw(http.StatusAccepted, err.Error())
	}
}
func (me *X1LedSql) DecodeFromRows(rows *sql.Rows) error {
	return rows.Scan(
		&me.ID,
		&me.Mac,
		&me.DelayTs,
		&me.CreateTime,
	)
}
func (me *X1LedSql) DecodeFromRow(row *sql.Row) error {
	return row.Scan(
		&me.ID,
		&me.Mac,
		&me.DelayTs,
		&me.CreateTime,
	)
}
func (me *X1LedSql) QueryByID(id int64) bool {
	me.SetID(id)
	return QueryDaoByID(me.myTable(), me.ID, me)
}
func (me *X1LedSql) Insert() bool {
	tblName := me.myTable()
	if !CheckTableExist(tblName) {
		sql := `create table ` + tblName + ` (
			id MEDIUMINT NOT NULL AUTO_INCREMENT,
			mac varchar(32) not null comment '设备mac,与设备表关联',
			delay_ts int not null comment '延时时间',
			create_time datetime comment '新增日期',
			PRIMARY KEY (id, mac, create_time)
		)`
		CreateTable(sql)
	}
	return InsertDao(me.myTable(), me)
}
func (me *X1LedSql) Delete() bool {
	return DeleteDaoByID(me.myTable(), me.ID)
}

/******************************************************************************
 * function: handleX1AckVersionMqttMsg
 * description: handle X1 version check and reply message
 * param {string} mac
 * param {[]byte} payload
 * return {*}
********************************************************************************/
type X1AckVersionJson struct {
	X1MsgHeader
	BaseVersion int `json:"base_version"`
	CoreVersion int `json:"core_version"`
}

func handleX1AckVersionMqttMsg(mac string, payload []byte) {
	var ackVersion X1AckVersionJson
	err := json.Unmarshal(payload, &ackVersion)
	if err != nil {
		mylog.Log.Errorln("handleX1AckVersionMqttMsg Unmarshal failed, err:", err)
		return
	}
	var reply []X1VersionReplySql
	QueryVersionReplyByMac(mac, &reply)
	if len(reply) > 0 {
		var replyJson X1VersionReplyJson
		replyJson.Id = otaRspCmd
		replyJson.Ack = 0
		replyJson.Upgrade = reply[0].Upgrade
		replyJson.BaseVersion = reply[0].BaseVersion
		replyJson.BaseFileSize = reply[0].BaseFileSize
		replyJson.BaseUrl = reply[0].BaseUrl
		replyJson.CoreVersion = reply[0].CoreVersion
		replyJson.CoreFileSize = reply[0].CoreFileSize
		replyJson.CoreUrl = reply[0].CoreUrl
		mq.PublishData(MakeX1ReplyVersionTopic(mac), replyJson)
	}
}

type X1VersionReplyJson struct {
	X1MsgHeader
	Upgrade      int    `json:"upgrade"`
	BaseVersion  int    `json:"base_version"`
	BaseFileSize int64  `json:"base_file_size"`
	BaseUrl      string `json:"base_url"`
	CoreVersion  int    `json:"core_version"`
	CoreFileSize int64  `json:"core_file_size"`
	CoreUrl      string `json:"core_url"`
}

type X1VersionReplySql struct {
	ID           int64  `json:"id" mysql:"id" binding:"omitempty"`
	Mac          string `json:"mac" mysql:"mac"`
	Upgrade      int    `json:"upgrade" mysql:"upgrade"`
	BaseVersion  int    `json:"base_version" mysql:"base_version"`
	BaseFileSize int64  `json:"base_file_size" mysql:"base_file_size"`
	BaseUrl      string `json:"base_url" mysql:"base_url"`
	CoreVersion  int    `json:"core_version" mysql:"core_version"`
	CoreFileSize int64  `json:"core_file_size" mysql:"core_file_size"`
	CoreUrl      string `json:"core_url" mysql:"core_url"`
	CreateTime   string `json:"create_time" mysql:"create_time" binding:"datetime=2006-01-02 15:04:05"`
}

func NewX1VersionReplySql() *X1VersionReplySql {
	return &X1VersionReplySql{
		ID:           0,
		Mac:          "",
		Upgrade:      0,
		BaseVersion:  0,
		BaseFileSize: 0,
		BaseUrl:      "",
		CoreVersion:  0,
		CoreFileSize: 0,
		CoreUrl:      "",
		CreateTime:   time.Now().Format(cfg.TmFmtStr),
	}
}
func (me *X1VersionReplySql) myTable() string {
	return common.X1OtaTbl
}
func (me *X1VersionReplySql) Update() bool {
	return UpdateDaoByID(me.myTable(), me.ID, me)
}
func (me *X1VersionReplySql) SetID(id int64) {
	me.ID = id
}
func (me *X1VersionReplySql) DecodeFromGin(c *gin.Context) {
	err := c.ShouldBindBodyWith(me, binding.JSON)
	if err != nil {
		exception.Throw(http.StatusAccepted, err.Error())
	}
}
func (me *X1VersionReplySql) DecodeFromRows(rows *sql.Rows) error {
	return rows.Scan(
		&me.ID,
		&me.Mac,
		&me.Upgrade,
		&me.BaseVersion,
		&me.BaseFileSize,
		&me.BaseUrl,
		&me.CoreVersion,
		&me.CoreFileSize,
		&me.CoreUrl,
		&me.CreateTime,
	)
}
func (me *X1VersionReplySql) DecodeFromRow(row *sql.Row) error {
	return row.Scan(
		&me.ID,
		&me.Mac,
		&me.Upgrade,
		&me.BaseVersion,
		&me.BaseFileSize,
		&me.BaseUrl,
		&me.CoreVersion,
		&me.CoreFileSize,
		&me.CoreUrl,
		&me.CreateTime,
	)
}
func (me *X1VersionReplySql) QueryByID(id int64) bool {
	me.SetID(id)
	return QueryDaoByID(me.myTable(), me.ID, me)
}
func (me *X1VersionReplySql) Insert() bool {
	tblName := me.myTable()
	if !CheckTableExist(tblName) {
		sql := `create table ` + tblName + ` (
			id MEDIUMINT NOT NULL AUTO_INCREMENT,
			mac varchar(32) not null comment '设备mac,与设备表关联',
			upgrade int not null comment '是否升级',
			base_version int not null comment '基础版本号',
			base_file_size bigint not null comment '基础文件大小',
			base_url varchar(256) not null comment '基础文件下载地址',
			core_version int not null comment '核心版本号',
			core_file_size bigint not null comment '核心文件大小',
			core_url varchar(256) not null comment '核心文件下载地址',
			create_time datetime comment '新增日期',
			PRIMARY KEY (id, mac, create_time)
		)`
		CreateTable(sql)
	}
	return InsertDao(me.myTable(), me)
}
func (me *X1VersionReplySql) Delete() bool {
	return DeleteDaoByID(me.myTable(), me.ID)
}

func QueryVersionReplyByMac(mac string, results *[]X1VersionReplySql) bool {
	filter := fmt.Sprintf("upgrade=1 and mac='%s'", mac)
	backFunc := func(rows *sql.Rows) {
		obj := NewX1VersionReplySql()
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*results = append(*results, *obj)
		}
	}
	return QueryDao(common.X1OtaTbl, filter, "create_time desc", 1, backFunc)
}

func CleanX1Event(mac string) {
	type CleanData struct {
		X1MsgHeader
		Type int `json:"type"`
	}
	var cleanMsg *CleanData = &CleanData{}
	cleanMsg.Id = cleanEventCmd
	cleanMsg.Ack = 0
	cleanMsg.Type = 3009
	mq.PublishData(MakeX1CleanEventTopic(mac), cleanMsg)
}

func SleepX1Switch(mac string, s int) {
	type SleepData struct {
		X1MsgHeader
		Switch int `json:"switch"`
	}
	var sleepMsg *SleepData = &SleepData{}
	sleepMsg.Id = sleepSwitchCmd
	sleepMsg.Ack = 0
	sleepMsg.Switch = s
	mq.PublishData(MakeX1SleepSwitchTopic(mac), sleepMsg)
}

func NurseModeX1Switch(mac string, s int) {
	type NurseData struct {
		X1MsgHeader
		Switch int `json:"switch"`
	}
	var nurseMsg *NurseData = &NurseData{}
	nurseMsg.Id = nurseModelCmd
	nurseMsg.Ack = 0
	nurseMsg.Switch = s
	mq.PublishData(MakeX1NurseModelTopic(mac), nurseMsg)
}

func ImproveDisturbedX1Switch(mac string, s int) {
	type ImproveData struct {
		X1MsgHeader
		Switch int `json:"switch"`
	}
	var improveMsg *ImproveData = &ImproveData{}
	improveMsg.Id = improveDisturbedCmd
	improveMsg.Ack = 0
	improveMsg.Switch = s
	mq.PublishData(MakeX1ImproveDisturbedTopic(mac), improveMsg)
}

func BreathAbnormalX1(mac string, ts int) {
	type BreathData struct {
		X1MsgHeader
		Ts int `json:"ts"`
	}
	var breathMsg *BreathData = &BreathData{}
	breathMsg.Id = breathAbnormalCmd
	breathMsg.Ack = 0
	breathMsg.Ts = ts
	mq.PublishData(MakeX1BreathAbnormalTopic(mac), breathMsg)
}
func BreathAbnormalX1Switch(mac string, s int) {
	type BreathData struct {
		X1MsgHeader
		Switch int `json:"switch"`
	}
	var breathMsg *BreathData = &BreathData{}
	breathMsg.Id = breathAbnormalSwitchCmd
	breathMsg.Ack = 0
	breathMsg.Switch = s
	mq.PublishData(MakeX1BreathAbnormalSwitchTopic(mac), breathMsg)
}
