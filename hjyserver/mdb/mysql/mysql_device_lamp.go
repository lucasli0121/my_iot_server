/*
 * @Author: liguoqiang
 * @Date: 2022-06-15 14:27:42
 * @LastEditors: liguoqiang
 * @LastEditTime: 2024-04-30 15:26:33
 * @Description:
 */

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
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

const (
	DeviceSyncCmd     = 101
	DeviceSyncCmdRsp  = 201
	KeepAlive         = 102
	RealDataSet       = 104
	RealDataSetRsp    = 204
	EventReport       = 105
	EventReportRsp    = 205
	ReportSubmit      = 106
	ReadLampStatus    = 107
	ReadLampStatusRsp = 207
	ControlLamp       = 108
	ControlLampRsp    = 208
)

const (
	LampSuccess       = 0
	LampFail          = 5000
	LampInvalidDevice = 5001
	LampOffline       = 5002
	LampParamError    = 5003
	LampTimeout       = 5004
	LampInvalidCmd    = 5005
	LampUnknowError   = 5006
)

const (
	LearnReportEvent = 1
)

var snVal int = 0

func makeSn() int {
	snVal++
	if snVal >= math.MaxInt32 {
		snVal = 0
	}
	return snVal
}

/******************************************************************************
 * function: MakeHl77DeliverTopicByMac
 * description: define Hl77 deliver topic by mac
 * return {*}
********************************************************************************/
func MakeHl77DeliverTopicByMac(mac string) string {
	mac = strings.ToLower(mac)
	mylog.Log.Infoln("HL77 SubscribeMqttTopic, mac:", mac)
	return fmt.Sprintf("HL77/upRaw/%s/data", mac)
}

/******************************************************************************
 * function: MakeHl77PublishTopicByMac
 * description: define publish topic by mac
 * return {*}
********************************************************************************/
func MakeHl77PublishTopicByMac(mac string) string {
	mac = strings.ToLower(mac)
	return fmt.Sprintf("HL77/downRaw/%s/data", mac)
}

/******************************************************************************
 * function: HandleLampMqttMsg
 * description: handle all mqtt message
 * message is json format,
 * decode message and desicde what to do
 * return {*}
********************************************************************************/
type LampMqttMsgProc struct {
}

func NewLampMqttMsgProc() *LampMqttMsgProc {
	return &LampMqttMsgProc{}
}

func (me *LampMqttMsgProc) HandleMqttMsg(topic string, payload []byte) {
	mylog.Log.Infoln("HandleMqttMsg:", topic, string(payload))
	HandleLampMqttMsg(payload)
}

func HandleLampMqttMsg(payload []byte) {
	var lampMqttMsg *LampMqttMsg = NewLampMqttMsg()
	err := json.Unmarshal([]byte(payload), &lampMqttMsg)
	if err != nil {
		mylog.Log.Errorln(err)
	}
	switch lampMqttMsg.Cmd {
	case DeviceSyncCmd:
		handleDeviceSyncCmd(lampMqttMsg)
	case KeepAlive:
		handleKeepAlive(lampMqttMsg)
	case RealDataSetRsp:
		handleRealDataSetRsp(lampMqttMsg)
	case EventReport:
		handleEventReportRsp(lampMqttMsg)
	case ReportSubmit:
		handleReportSubmit(lampMqttMsg)
	case ControlLampRsp:
		handleControlLampRsp(lampMqttMsg)
	case ReadLampStatusRsp:
		handleControlLampRsp(lampMqttMsg)
	default:
	}
}

/******************************************************************************
 * description: define lamp mqtt message struct, this struct has two fields
 *              one is message head, another is message body
********************************************************************************/
type LampMqttMsg struct {
	Cmd  int    `json:"cmd"`
	Sn   int    `json:"sn"`
	Ts   int64  `json:"ts"`
	Mac  string `json:"mac"`
	Data string `json:"data"`
}

func NewLampMqttMsg() *LampMqttMsg {
	return &LampMqttMsg{
		Cmd:  0,
		Sn:   0,
		Ts:   0,
		Mac:  "",
		Data: "",
	}
}

/******************************************************
* device response version information
*******************************************************/
type LampVersionRsp struct {
	ErrorCode         int    `json:"errorCode"`
	SystemTime        int64  `json:"systemTime"`
	Upgrade           int    `json:"upgrade"`
	RemoteBaseVersion int    `json:"remoteBaseVersion"`
	BaseOtaUrl        string `json:"baseOtaUrl"`
	BaseFileSize      int    `json:"baseFileSize"`
	RemoteCoreVersion int    `json:"remoteCoreVersion"`
	CoreOtaUrl        string `json:"coreOtaUrl"`
	CoreFileSize      int    `json:"coreFileSize"`
}

func NewLampVersionRsp() *LampVersionRsp {
	return &LampVersionRsp{
		ErrorCode:         0,
		SystemTime:        0,
		Upgrade:           0,
		RemoteBaseVersion: 0,
		BaseOtaUrl:        "",
		BaseFileSize:      0,
		RemoteCoreVersion: 0,
		CoreOtaUrl:        "",
		CoreFileSize:      0,
	}
}

type LampOtaSql struct {
	ID                int64  `json:"id" mysql:"id"`
	Upgrade           int    `json:"upgrade" mysql:"upgrade"`
	RemoteBaseVersion int    `json:"remoteBaseVersion" mysql:"remote_base_version"`
	BaseOtaUrl        string `json:"baseOtaUrl" mysql:"base_ota_url"`
	BaseFileSize      int    `json:"baseFileSize" mysql:"base_file_size"`
	RemoteCoreVersion int    `json:"remoteCoreVersion" mysql:"remote_core_version"`
	CoreOtaUrl        string `json:"coreOtaUrl" mysql:"core_ota_url"`
	CoreFileSize      int    `json:"coreFileSize" mysql:"core_file_size"`
}

func QueryLampOtaByCond(filter interface{}, results *[]LampOtaSql) bool {
	res := false
	backFunc := func(rows *sql.Rows) {
		obj := &LampOtaSql{}
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*results = append(*results, *obj)
		}
	}
	res = QueryDao(common.LampOtaTbl, filter, nil, 1, backFunc)
	return res
}

func (me *LampOtaSql) DecodeFromRows(rows *sql.Rows) error {
	err := rows.Scan(&me.ID,
		&me.Upgrade,
		&me.RemoteBaseVersion,
		&me.BaseOtaUrl,
		&me.BaseFileSize,
		&me.RemoteCoreVersion,
		&me.CoreOtaUrl,
		&me.CoreFileSize,
	)
	return err
}
func (me *LampOtaSql) DecodeFromRow(row *sql.Row) error {
	err := row.Scan(&me.ID,
		&me.Upgrade,
		&me.RemoteBaseVersion,
		&me.BaseOtaUrl,
		&me.BaseFileSize,
		&me.RemoteCoreVersion,
		&me.CoreOtaUrl,
		&me.CoreFileSize,
	)
	return err
}

/*
Decode 解析从gin获取的数据 转换成Device
*/
func (me *LampOtaSql) DecodeFromGin(c *gin.Context) {
	if err := c.ShouldBindBodyWith(me, binding.JSON); err != nil {
		exception.Throw(http.StatusAccepted, err.Error())
	}
	if me.ID == 0 {
		exception.Throw(http.StatusAccepted, "id is empty!")
	}
}

func (me *LampOtaSql) QueryByID(id int64) bool {
	return QueryDaoByID(common.LampOtaTbl, id, me)
}

/*
Insert 股票基本信息数据插入
*/
func (me *LampOtaSql) Insert() bool {
	tblName := common.LampOtaTbl
	if !CheckTableExist(tblName) {
		sql := `create table ` + tblName + ` (
            id MEDIUMINT NOT NULL AUTO_INCREMENT,
			upgrade int,
			remote_base_version int,
			base_ota_url varchar(256),
			base_file_size int COMMENT '',
			remote_core_version int COMMENT '',
			core_ota_url varchar(256) COMMENT '',
			core_file_size int COMMENT '',
            PRIMARY KEY(id)
        )`
		CreateTable(sql)
	}
	return InsertDao(tblName, me)
}

/*
Update() 更新股票基本信息
*/
func (me *LampOtaSql) Update() bool {
	return UpdateDaoByID(common.LampOtaTbl, me.ID, me)
}

/*
Delete() 删除指数
*/
func (me *LampOtaSql) Delete() bool {
	return DeleteDaoByID(common.LampOtaTbl, me.ID)
}

/*
设置ID
*/
func (me *LampOtaSql) SetID(id int64) {
	me.ID = id
}

/******************************************************************************
 * function: handleDeviceSyncCmd
 * description:
 * param {*LampMqttMsg} lampMqttMsg
 * return {*}
********************************************************************************/
func handleDeviceSyncCmd(lampMqttMsg *LampMqttMsg) {
	type Ver struct {
		Version string `json:"version"`
	}
	var version *Ver = &Ver{Version: ""}
	err := json.Unmarshal([]byte(lampMqttMsg.Data), &version)
	if err != nil {
		mylog.Log.Errorln(err)
	}
	mylog.Log.Infoln(version)
	var msgRsp = NewLampMqttMsg()
	msgRsp.Cmd = DeviceSyncCmdRsp
	msgRsp.Sn = lampMqttMsg.Sn
	msgRsp.Ts = time.Now().Unix()
	msgRsp.Mac = lampMqttMsg.Mac

	var verRsp *LampVersionRsp = NewLampVersionRsp()
	verRsp.ErrorCode = LampSuccess
	verRsp.SystemTime = time.Now().Unix()
	verRsp.Upgrade = 0
	var otaLst []LampOtaSql
	QueryLampOtaByCond("upgrade=1", &otaLst)
	if len(otaLst) > 0 {
		verRsp.Upgrade = 1
		verRsp.RemoteBaseVersion = otaLst[0].RemoteBaseVersion
		verRsp.BaseOtaUrl = otaLst[0].BaseOtaUrl
		verRsp.BaseFileSize = otaLst[0].BaseFileSize
		verRsp.RemoteCoreVersion = otaLst[0].RemoteCoreVersion
		verRsp.CoreOtaUrl = otaLst[0].CoreOtaUrl
		verRsp.CoreFileSize = otaLst[0].CoreFileSize
	}
	js, err := json.Marshal(verRsp)
	if err != nil {
		mylog.Log.Errorln(err)
	}
	mylog.Log.Infoln(string(js))
	msgRsp.Data = string(js)
	// js, err = json.Marshal(msgRsp)
	// if err != nil {
	// 	mylog.Log.Errorln(err)
	// }
	// 发送一个灯状态的请求MQ
	mq.PublishData(MakeHl77PublishTopicByMac(lampMqttMsg.Mac), msgRsp)
}

/******************************************************************************
 * function: handleKeepAlive
 * description:
 * return {*}
********************************************************************************/
func handleKeepAlive(lampMqttMsg *LampMqttMsg) {
	mylog.Log.Infoln("handleKeepAlive:", lampMqttMsg.Cmd)
	SetDeviceOnline(lampMqttMsg.Mac, 1, 0)
}

type RealDataReq struct {
	DeadLine int64 `json:"deadline"`
	Freq     int   `json:"frequency"`
	KeepPush int   `json:"keep_push"`
}

func AskHl77RealData(mac string, freq int, keepPush int) {
	lampRealReq := &RealDataReq{}
	lampRealReq.DeadLine = 0
	lampRealReq.Freq = freq
	lampRealReq.KeepPush = keepPush
	lampRealReq.SendRealDataReq(mac)
}

/******************************************************************************
 * function: SendRealDataReq
 * description:
 * return {*}
********************************************************************************/
func (me *RealDataReq) SendRealDataReq(mac string) {
	var msg *LampMqttMsg = NewLampMqttMsg()
	msg.Cmd = RealDataSet
	msg.Sn = makeSn()
	msg.Ts = time.Now().Unix()
	msg.Mac = mac
	js, err := json.Marshal(me)
	if err != nil {
		mylog.Log.Errorln(err)
	}
	msg.Data = string(js)
	mq.PublishData(MakeHl77PublishTopicByMac(mac), msg)
}

type RealDataJson struct {
	KeepPush     int   `json:"keep_push"` // 0:stop push, 1:start push
	BodyStatus   []int `json:"body_status"`
	Respiratory  []int `json:"respiratory"`
	HeartRate    []int `json:"heart_rate"`
	BodyMovement []int `json:"body_movement"`
	FlowState    []int `json:"flow_state"`
	PostureState []int `json:"posture_state"`
	ActivityFreq []int `json:"activity_freq"`
	BodyPos      []int `json:"body_pos"`
	BodyAngle    []int `json:"body_angle"`
	HeadPos      []int `json:"head_pos"`
	HeadAngle    []int `json:"head_angle"`
	HandPos      []int `json:"hand_pos"`
	HandAngle    []int `json:"hand_angle"`
}

// swagger:model RealDataSql
type RealDataSql struct {
	ID           int64  `json:"id" mysql:"id"`
	Mac          string `json:"mac" mysql:"mac"`
	BodyStatus   int    `json:"body_status" mysql:"body_status"`
	Respiratory  int    `json:"respiratory" mysql:"respiratory"`
	HeartRate    int    `json:"heart_rate" mysql:"heart_rate"`
	BodyMovement int    `json:"body_movement" mysql:"body_movement"`
	FlowState    int    `json:"flow_state" mysql:"flow_state"`
	PostureState int    `json:"posture_state" mysql:"posture_state"`
	ActivityFreq int    `json:"activity_freq" mysql:"activity_freq"`
	BodyPos      int    `json:"body_pos" mysql:"body_pos"`
	BodyAngle    int    `json:"body_angle" mysql:"body_angle"`
	HeadPos      int    `json:"head_pos" mysql:"head_pos"`
	HeadAngle    int    `json:"head_angle" mysql:"head_angle"`
	HandPos      int    `json:"hand_pos" mysql:"hand_pos"`
	HandAngle    int    `json:"hand_angle" mysql:"hand_angle"`
	CreateTime   string `json:"create_time" mysql:"create_time"`
	Remark       string `json:"remark" mysql:"remark"`
}

func NewRealDataSql() *RealDataSql {
	return &RealDataSql{
		ID:           0,
		Mac:          "",
		BodyStatus:   0,
		Respiratory:  0,
		HeartRate:    0,
		BodyMovement: 0,
		FlowState:    0,
		PostureState: 0,
		ActivityFreq: 0,
		BodyPos:      0,
		BodyAngle:    0,
		HeadPos:      0,
		HeadAngle:    0,
		HandPos:      0,
		HandAngle:    0,
		CreateTime:   common.GetNowDate(),
	}
}

func QueryLampRealDataByCond(filter interface{}, page *common.PageDao, sort interface{}, limited int, results *[]RealDataSql) bool {
	res := false
	backFunc := func(rows *sql.Rows) {
		obj := NewRealDataSql()
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*results = append(*results, *obj)
		}
	}
	if page == nil {
		res = QueryDao(common.LampRealDataTbl, filter, sort, limited, backFunc)
	} else {
		res = QueryPage(common.LampRealDataTbl, page, filter, sort, backFunc)
	}
	return res
}

func (me *RealDataSql) DecodeFromRows(rows *sql.Rows) error {
	err := rows.Scan(&me.ID,
		&me.Mac,
		&me.BodyStatus,
		&me.Respiratory,
		&me.HeartRate,
		&me.BodyMovement,
		&me.FlowState,
		&me.PostureState,
		&me.ActivityFreq,
		&me.BodyPos,
		&me.BodyAngle,
		&me.HeadPos,
		&me.HeadAngle,
		&me.HandPos,
		&me.HandAngle,
		&me.CreateTime,
		&me.Remark,
	)
	return err
}
func (me *RealDataSql) DecodeFromRow(row *sql.Row) error {
	err := row.Scan(&me.ID,
		&me.Mac,
		&me.BodyStatus,
		&me.Respiratory,
		&me.HeartRate,
		&me.BodyMovement,
		&me.FlowState,
		&me.PostureState,
		&me.ActivityFreq,
		&me.BodyPos,
		&me.BodyAngle,
		&me.HeadPos,
		&me.HeadAngle,
		&me.HandPos,
		&me.HandAngle,
		&me.CreateTime,
		&me.Remark,
	)
	return err
}

/*
Decode 解析从gin获取的数据 转换成Device
*/
func (me *RealDataSql) DecodeFromGin(c *gin.Context) {
	if err := c.ShouldBindBodyWith(me, binding.JSON); err != nil {
		exception.Throw(http.StatusAccepted, err.Error())
	}
	if me.Mac == "" {
		exception.Throw(http.StatusAccepted, "mac is empty!")
	}
}

func (me *RealDataSql) QueryByID(id int64) bool {
	return QueryDaoByID(common.LampRealDataTbl, id, me)
}

/*
Insert 股票基本信息数据插入
*/
func (me *RealDataSql) Insert() bool {
	tblName := common.LampRealDataTbl
	if !CheckTableExist(tblName) {
		sql := `create table ` + tblName + ` (
            id MEDIUMINT NOT NULL AUTO_INCREMENT,
			mac char(32) NOT NULL COMMENT 'mac地址',
			body_status int NOT NULL COMMENT '',
			respiratory int NOT NULL COMMENT '',
			heart_rate int NOT NULL COMMENT '',
			body_movement int NOT NULL COMMENT '',
			flow_state int COMMENT '',
			posture_state int COMMENT '',
			activity_freq int COMMENT '',
			body_pos int COMMENT '',
			body_angle int COMMENT '',
			head_pos int COMMENT '',
			head_angle int COMMENT '',
			hand_pos int COMMENT '',
			hand_angle int COMMENT '',
            create_time datetime comment '新增日期',
			remark varchar(64) comment '备注',
            PRIMARY KEY (id, mac, create_time)

        )`
		CreateTable(sql)
	}
	ret := InsertDao(common.LampRealDataTbl, me)
	if me.FlowState > 0 && me.HeartRate > 0 {
		BringLampUserToStudyRoom(me.Mac, me.CreateTime)
	}
	return ret
}

// swagger:model UserRoomStatus
type UserRoomStatus struct {
	UserId   int64  `json:"user_id"`
	RoomId   int64  `json:"room_id"`
	Status   int    `json:"status"`
	DateTime string `json:"date_time"`
}

/******************************************************************************
 * function:
 * description: find user who has the lamp and bring him to study room
 * param {string} mac
 * param {string} createTm
 * return {*}
********************************************************************************/
func BringLampUserToStudyRoom(mac string, createTm string) {
	var filter string
	t, err := common.StrToTime(createTm)
	if err != nil {
		mylog.Log.Errorln(err)
		return
	}
	// 30s before
	beforeTm := t.Add(time.Second * -30).Format(cfg.TmFmtStr)
	filter = fmt.Sprintf("mac='%s' and flow_state > 0 and heart_rate > 0 and create_time>='%s' and create_time<='%s'", mac, beforeTm, createTm)
	var realDataSqls []RealDataSql
	QueryLampRealDataByCond(filter, nil, nil, -1, &realDataSqls)
	// lamp report real data one time every 6s
	// there will be 6 records in db every 30s
	// at least greater than 2 records, we can do next step
	if len(realDataSqls) < 2 {
		return
	}
	// at first get user id who bind the lamp
	var userDevices []UserDeviceDetail
	QueryUserDeviceDetailByMac(mac, &userDevices)
	for _, userDevice := range userDevices {
		// query only one study room that the user has been invited
		var studyRoomUsers []StudyRoomUser
		filter = fmt.Sprintf("user_id=%d and status=1", userDevice.UserId)
		QueryStudyRoomUserByCond(filter, nil, nil, 1, &studyRoomUsers)
		for _, studyRoomUser := range studyRoomUsers {
			// check whether the user has enter the study room
			// if not then bring him into
			filter = fmt.Sprintf("user_id=%d and room_id=%d and status=1", studyRoomUser.UserId, studyRoomUser.RoomId)
			var results []UserStudyRecord
			QueryUserStudyRecordByCond(filter, nil, nil, 1, &results)
			if len(results) == 0 {
				obj := NewUserStudyRecord()
				obj.UserId = studyRoomUser.UserId
				obj.RoomId = studyRoomUser.RoomId
				obj.Status = 1
				obj.Sn = studyRoomUser.Sn
				obj.EnterTime = common.GetNowTime()
				obj.LeaveTime = common.GetNowTime()
				if obj.Insert() {
					mylog.Log.Infoln("BringLampUserToStudyRoom, mac:", mac, " user_id:", obj.UserId, " room_id:", obj.RoomId)
					userStatus := &UserRoomStatus{
						UserId:   obj.UserId,
						RoomId:   obj.RoomId,
						Status:   1,
						DateTime: obj.EnterTime,
					}
					mq.PublishData(common.MakeHl77UserEnterRoomTopicByMac(mac), userStatus)
				}
			}
		}
	}
}

/******************************************************************************
 * function: TakeLampUserFromStudyRoom
 * description:
 * param {string} mac
 * return {*}
********************************************************************************/
func TakeLampUserFromStudyRoom(mac string) {
	// query user id who bind the lamp
	var userDevices []UserDeviceDetail
	QueryUserDeviceDetailByMac(mac, &userDevices)
	for _, userDevice := range userDevices {
		// query study room that the user has been invited
		var studyRoomUsers []StudyRoomUser
		var filter = fmt.Sprintf("user_id=%d and status=1", userDevice.UserId)
		QueryStudyRoomUserByCond(filter, nil, nil, 1, &studyRoomUsers)
		for _, studyRoomUser := range studyRoomUsers {
			// check the user has enter the study room
			// if in then take him out
			var filter = fmt.Sprintf("user_id=%d and room_id=%d and status=1", studyRoomUser.UserId, studyRoomUser.RoomId)
			var sort = "enter_time desc"
			var results []UserStudyRecord
			QueryUserStudyRecordByCond(filter, nil, sort, 1, &results)
			for _, result := range results {
				result.LeaveTime = common.GetNowTime()
				result.Status = 0
				if result.Update() {
					mylog.Log.Infoln("TakeLampUserFromStudyRoom, mac:", mac, " user_id:", result.UserId, " room_id:", result.RoomId)
					userStatus := &UserRoomStatus{
						UserId:   result.UserId,
						RoomId:   result.RoomId,
						Status:   0,
						DateTime: result.LeaveTime,
					}
					mq.PublishData(common.MakeHl77UserEnterRoomTopicByMac(mac), userStatus)
				}
			}
		}
	}
}

/*
Update() 更新股票基本信息
*/
func (me *RealDataSql) Update() bool {
	return UpdateDaoByID(common.LampRealDataTbl, me.ID, me)
}

/*
Delete() 删除指数
*/
func (me *RealDataSql) Delete() bool {
	return DeleteDaoByID(common.LampRealDataTbl, me.ID)
}

/*
设置ID
*/
func (me *RealDataSql) SetID(id int64) {
	me.ID = id
}
func (me *RealDataSql) ConvertHl77RealDataToHeartRate() *HeartRate {
	heartObj := &HeartRate{
		ID:           0,
		Mac:          me.Mac,
		HeartRate:    me.HeartRate,
		BreatheRate:  me.Respiratory,
		ActiveStatus: me.ActivityFreq,
		DateTime:     me.CreateTime,
	}
	if heartObj.HeartRate > 0 && heartObj.BreatheRate > 0 {
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

/******************************************************************************
 * function: handleRealDataSetRsp
 * description: handle real data set response from device
 * return {*}
********************************************************************************/
func handleRealDataSetRsp(lampMqttMsg *LampMqttMsg) {
	var realDataJson *RealDataJson = &RealDataJson{}
	err := json.Unmarshal([]byte(lampMqttMsg.Data), &realDataJson)
	if err != nil {
		mylog.Log.Errorln(err)
	}
	dataLen := len(realDataJson.Respiratory)
	if dataLen <= 0 {
		return
	}
	var realDataSql *RealDataSql = NewRealDataSql()
	realDataSql.Mac = lampMqttMsg.Mac
	realDataSql.CreateTime = common.SecondsToTimeStr(lampMqttMsg.Ts)
	totalRespRate := 0
	totalHeartRate := 0
	totalBodyMovement := 0
	totalFlowState := 0
	totalPostureState := 0
	totalActivityFreq := 0
	totalBodyPos := 0
	totalBodyAngle := 0
	totalHeadPos := 0
	totalHeadAngle := 0
	totalHandPos := 0
	totalHandAngle := 0
	for i := 0; i < dataLen; i++ {
		totalRespRate += realDataJson.Respiratory[i]
		totalHeartRate += realDataJson.HeartRate[i]
		totalBodyMovement += realDataJson.BodyMovement[i]
		totalFlowState += realDataJson.FlowState[i]
		totalPostureState += realDataJson.PostureState[i]
		totalActivityFreq += realDataJson.ActivityFreq[i]
		totalBodyPos += realDataJson.BodyPos[i]
		totalBodyAngle += realDataJson.BodyAngle[i]
		totalHeadPos += realDataJson.HeadPos[i]
		totalHeadAngle += realDataJson.HeadAngle[i]
		totalHandPos += realDataJson.HandPos[i]
		totalHandAngle += realDataJson.HandAngle[i]
	}
	if len(realDataJson.BodyStatus) > 0 {
		realDataSql.BodyStatus = realDataJson.BodyStatus[len(realDataJson.BodyStatus)-1]
	}
	realDataSql.Respiratory = realDataJson.Respiratory[len(realDataJson.Respiratory)-1]
	realDataSql.HeartRate = realDataJson.HeartRate[len(realDataJson.HeartRate)-1]
	realDataSql.BodyMovement = realDataJson.BodyMovement[len(realDataJson.BodyMovement)-1]
	realDataSql.FlowState = realDataJson.FlowState[len(realDataJson.FlowState)-1]
	realDataSql.PostureState = realDataJson.PostureState[len(realDataJson.PostureState)-1]
	realDataSql.ActivityFreq = realDataJson.ActivityFreq[len(realDataJson.ActivityFreq)-1]
	realDataSql.BodyPos = realDataJson.BodyPos[len(realDataJson.BodyPos)-1]
	realDataSql.BodyAngle = realDataJson.BodyAngle[len(realDataJson.BodyAngle)-1]
	realDataSql.HeadPos = realDataJson.HeadPos[len(realDataJson.HeadPos)-1]
	realDataSql.HeadAngle = realDataJson.HeadAngle[len(realDataJson.HeadAngle)-1]
	realDataSql.HandPos = realDataJson.HandPos[len(realDataJson.HandPos)-1]
	realDataSql.HandAngle = realDataJson.HandAngle[len(realDataJson.HandAngle)-1]

	if CheckDiffBetweenTwoLampDeviceRecords(LampType, lampMqttMsg.Mac, realDataSql) {
		//publish RealDataSql to APP
		mq.PublishData(common.MakeHl77RealDataTopic(lampMqttMsg.Mac), realDataSql)
		// send a read mq message to lamp control status
		readLampControlStatus(lampMqttMsg.Mac)
		// save to database
		taskPool.Put(&gopool.Task{
			Params: []interface{}{realDataSql},
			Do: func(params ...interface{}) {
				var obj = params[0].(*RealDataSql)
				obj.Insert()
			},
		})
	}
}

func QueryHl77RealDataToHeartRateByCond(filter interface{}, page *common.PageDao, sort interface{}, limited int, results *[]HeartRate) bool {
	res := false
	backFunc := func(rows *sql.Rows) {
		obj := NewRealDataSql()
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*results = append(*results, *obj.ConvertHl77RealDataToHeartRate())
		}
	}
	if page == nil {
		res = QueryDao(common.LampRealDataTbl, filter, sort, limited, backFunc)
	} else {
		res = QueryPage(common.LampRealDataTbl, page, filter, sort, backFunc)
	}
	return res
}

type EventReportJson struct {
	EventType int   `json:"eventType" mysql:"eventType"`
	EventTs   int64 `json:"eventTs" mysql:"eventTs"`
}

// swagger:model EventReportSql
type EventReportSql struct {
	ID         int64  `json:"id" mysql:"id"`
	Mac        string `json:"mac" mysql:"mac"`
	EventType  int    `json:"eventType" mysql:"eventType"`
	EventTs    int64  `json:"eventTs" mysql:"eventTs"`
	CreateTime string `json:"create_time" mysql:"create_time"`
}

func NewEventReportSql() *EventReportSql {
	return &EventReportSql{
		ID:         0,
		Mac:        "",
		EventType:  0,
		EventTs:    0,
		CreateTime: common.GetNowDate(),
	}
}

func QueryLampEventByCond(filter interface{}, page *common.PageDao, sort interface{}, limited int, results *[]EventReportSql) bool {
	res := false
	backFunc := func(rows *sql.Rows) {
		obj := NewEventReportSql()
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*results = append(*results, *obj)
		}
	}
	if page == nil {
		res = QueryDao(common.LampEventTbl, filter, sort, limited, backFunc)
	} else {
		res = QueryPage(common.LampEventTbl, page, filter, sort, backFunc)
	}
	return res
}

/*
 */
func (me *EventReportSql) DecodeFromGin(c *gin.Context) {
	if err := c.ShouldBindBodyWith(me, binding.JSON); err != nil {
		exception.Throw(http.StatusAccepted, err.Error())
	}
	if me.Mac == "" {
		exception.Throw(http.StatusAccepted, "mac is empty!")
	}
}

func (me *EventReportSql) DecodeFromRows(rows *sql.Rows) error {
	err := rows.Scan(&me.ID,
		&me.Mac,
		&me.EventType,
		&me.EventTs,
		&me.CreateTime,
	)
	return err
}
func (me *EventReportSql) DecodeFromRow(row *sql.Row) error {
	err := row.Scan(&me.ID,
		&me.Mac,
		&me.EventType,
		&me.EventTs,
		&me.CreateTime,
	)
	return err
}

func (me *EventReportSql) QueryByID(id int64) bool {
	return QueryDaoByID(common.LampEventTbl, id, me)
}

/*
 */
func (me *EventReportSql) Insert() bool {
	tblName := common.LampEventTbl
	if !CheckTableExist(tblName) {
		sql := `create table ` + tblName + ` (
            id MEDIUMINT NOT NULL AUTO_INCREMENT,
			mac char(32) NOT NULL COMMENT 'mac地址',
			eventType int NOT NULL COMMENT '事件类型',
			eventTs int NOT NULL COMMENT '事件时间',
			create_time datetime comment '新增日期',
            PRIMARY KEY (id, mac, create_time)

        )`
		CreateTable(sql)
	}
	return InsertDao(common.LampEventTbl, me)
}

/*
Update() 更新股票基本信息
*/
func (me *EventReportSql) Update() bool {
	return UpdateDaoByID(common.LampEventTbl, me.ID, me)
}

/*
Delete() 删除指数
*/
func (me *EventReportSql) Delete() bool {
	return DeleteDaoByID(common.LampEventTbl, me.ID)
}

/*
设置ID
*/
func (me *EventReportSql) SetID(id int64) {
	me.ID = id
}

/******************************************************************************
 * function: handleEventReportRsp
 * description:
 * param {*LampMqttMsg} lampMqttMsg
 * return {*}
********************************************************************************/
func handleEventReportRsp(lampMqttMsg *LampMqttMsg) {
	var eventJson *EventReportJson = &EventReportJson{}
	err := json.Unmarshal([]byte(lampMqttMsg.Data), &eventJson)
	if err != nil {
		mylog.Log.Errorln(err)
	}
	var eventSql *EventReportSql = &EventReportSql{}
	eventSql.Mac = lampMqttMsg.Mac
	eventSql.EventType = eventJson.EventType
	eventSql.EventTs = eventJson.EventTs
	var tm time.Duration = time.Duration(eventJson.EventTs) * time.Second
	eventSql.CreateTime = time.Unix(int64(tm.Seconds()), 0).Format(cfg.TmFmtStr)
	taskPool.Put(&gopool.Task{
		Params: []interface{}{eventSql},
		Do: func(params ...interface{}) {
			var obj = params[0].(*EventReportSql)
			obj.Insert()
		},
	})
	type resp struct {
		ErrorCode int `json:"errorCode"`
	}
	var msg *LampMqttMsg = NewLampMqttMsg()
	msg.Cmd = EventReportRsp
	msg.Sn = lampMqttMsg.Sn
	msg.Ts = time.Now().Unix()
	msg.Mac = lampMqttMsg.Mac
	var rsp *resp = &resp{ErrorCode: LampSuccess}
	js, err := json.Marshal(rsp)
	if err != nil {
		mylog.Log.Errorln(err)
	}
	msg.Data = string(js)
	mq.PublishData(MakeHl77PublishTopicByMac(lampMqttMsg.Mac), msg)
}

/******************************************************************************
 * function:
 * description: define report struct
 * return {*}
********************************************************************************/
type LampReportJson struct {
	ReportStart     int64 `json:"report_start"`
	ReportEnd       int64 `json:"report_end"`
	FlowState       []int `json:"flow_state"`
	Evaluation      int   `json:"evaluation"`
	StudyEfficiency int   `json:"study_efficiency"`
	Concentration   int   `json:"concentration"`
	SeqInterval     int   `json:"seq_interval"`
	Respiratory     []int `json:"respiratory"`
	HeartRate       []int `json:"heart_rate"`
	PostureState    []int `json:"posture_state"`
	ActivityFreq    []int `json:"activity_freq"`
	BodyPos         []int `json:"body_pos"`
}

// swagger:model LampReportSql
type LampReportSql struct {
	ID              int64  `json:"id" mysql:"id"`
	Mac             string `json:"mac" mysql:"mac"`
	ReportStart     string `json:"report_start" mysql:"report_start"`
	ReportEnd       string `json:"report_end" mysql:"report_end"`
	FlowState       int    `json:"flow_state" mysql:"flow_state"`
	FlowStateTime   string `json:"flow_state_time" mysql:"flow_state_time"`
	Evaluation      int    `json:"evaluation" mysql:"evaluation"`
	StudyEfficiency int    `json:"study_efficiency" mysql:"study_efficiency"`
	Concentration   int    `json:"concentration" mysql:"concentration"`
	SeqInterval     int    `json:"seq_interval" mysql:"seq_interval"`
	Respiratory     int    `json:"respiratory" mysql:"respiratory"`
	HeartRate       int    `json:"heart_rate" mysql:"heart_rate"`
	PostureState    int    `json:"posture_state" mysql:"posture_state"`
	ActivityFreq    int    `json:"activity_freq" mysql:"activity_freq"`
	BodyPos         int    `json:"body_pos" mysql:"body_pos"`
	CreateTime      string `json:"create_time" mysql:"create_time"`
}

func NewLampReportSql() *LampReportSql {
	return &LampReportSql{
		ID:              0,
		Mac:             "",
		ReportStart:     "",
		ReportEnd:       "",
		FlowState:       0,
		FlowStateTime:   "",
		Evaluation:      0,
		StudyEfficiency: 0,
		Concentration:   0,
		SeqInterval:     0,
		Respiratory:     0,
		HeartRate:       0,
		PostureState:    0,
		ActivityFreq:    0,
		BodyPos:         0,
		CreateTime:      common.GetNowDate(),
	}
}
func (me *LampReportSql) DecodeFromGin(c *gin.Context) {
	if err := c.ShouldBindBodyWith(me, binding.JSON); err != nil {
		exception.Throw(http.StatusAccepted, err.Error())
	}
	if me.Mac == "" {
		exception.Throw(http.StatusAccepted, "mac is empty!")
	}
}

func (me *LampReportSql) DecodeFromRows(rows *sql.Rows) error {
	err := rows.Scan(&me.ID,
		&me.Mac,
		&me.ReportStart,
		&me.ReportEnd,
		&me.FlowState,
		&me.FlowStateTime,
		&me.Evaluation,
		&me.StudyEfficiency,
		&me.Concentration,
		&me.SeqInterval,
		&me.Respiratory,
		&me.HeartRate,
		&me.PostureState,
		&me.ActivityFreq,
		&me.BodyPos,
		&me.CreateTime,
	)
	return err
}
func (me *LampReportSql) DecodeFromRow(row *sql.Row) error {
	err := row.Scan(&me.ID,
		&me.Mac,
		&me.ReportStart,
		&me.ReportEnd,
		&me.FlowState,
		&me.FlowStateTime,
		&me.Evaluation,
		&me.StudyEfficiency,
		&me.Concentration,
		&me.SeqInterval,
		&me.Respiratory,
		&me.HeartRate,
		&me.PostureState,
		&me.ActivityFreq,
		&me.BodyPos,
		&me.CreateTime,
	)
	return err
}

func (me *LampReportSql) QueryByID(id int64) bool {
	return QueryDaoByID(common.LampReportTbl, id, me)
}

/*
 */
func (me *LampReportSql) Insert() bool {
	tblName := common.LampReportTbl
	if !CheckTableExist(tblName) {
		sql := `create table ` + tblName + ` (
            id MEDIUMINT NOT NULL AUTO_INCREMENT,
			mac char(32) NOT NULL COMMENT 'mac地址',
			report_start datetime NOT NULL COMMENT '报告开始时间',
			report_end datetime NOT NULL COMMENT '报告结束时间',
			flow_state int COMMENT '',
			flow_state_time datetime COMMENT '心流状态时间',
			evaluation int COMMENT '',
			study_efficiency int COMMENT '',
			concentration int COMMENT '',
			seq_interval int COMMENT '',
			respiratory int COMMENT '',
			heart_rate int COMMENT '',
			posture_state int COMMENT '',
			activity_freq int COMMENT '',
			body_pos int COMMENT '',
			create_time datetime comment '新增日期',
            PRIMARY KEY (id, mac, create_time)

        )`
		CreateTable(sql)
	}
	return InsertDao(tblName, me)
}

/*
Update() 更新股票基本信息
*/
func (me *LampReportSql) Update() bool {
	return UpdateDaoByID(common.LampReportTbl, me.ID, me)
}

/*
Delete() 删除指数
*/
func (me *LampReportSql) Delete() bool {
	return DeleteDaoByID(common.LampReportTbl, me.ID)
}

/*
设置ID
*/
func (me *LampReportSql) SetID(id int64) {
	me.ID = id
}

func QueryLampReportByCond(filter interface{}, page *common.PageDao, sort interface{}, limited int, results *[]LampReportSql) bool {
	res := false
	backFunc := func(rows *sql.Rows) {
		obj := NewLampReportSql()
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*results = append(*results, *obj)
		}
	}
	if page == nil {
		res = QueryDao(common.LampReportTbl, filter, sort, limited, backFunc)
	} else {
		res = QueryPage(common.LampReportTbl, page, filter, sort, backFunc)
	}
	return res
}

// swigger:model LampReportStatus
type LampReportStatus struct {
	Mac                  string `json:"mac"`
	Evaluation           int    `json:"evaluation"`
	EvaluationTimeLen    int64  `json:"evaluation_time_len"`
	Efficiency           int    `json:"efficiency"`
	EfficiencyTimeLen    int64  `json:"efficiency_time_len"`
	Concentration        int    `json:"concentration"`
	ConcentrationTimeLen int64  `json:"concentration_time_len"`
}

/******************************************************************************
 * function:
 * description:
 * param {string} mac
 * param {string} beginDay
 * param {string} endDay
 * param {*[]LampReportStatus} results
 * return {*}
********************************************************************************/
func QueryLampReportStatusByMac(mac string, beginDay string, endDay string, results *[]LampReportStatus) {
	var filter string
	if beginDay != "" && endDay != "" {
		filter = fmt.Sprintf(" mac='%s' and date(report_end)>=date('%s') and date(report_end)<=date('%s')", mac, beginDay, endDay)
	} else {
		filter = fmt.Sprintf(" mac='%s'", mac)
	}

	var reportSqls []LampReportSql
	QueryLampReportByCond(filter, nil, nil, 1, &reportSqls)
	for _, reportSql := range reportSqls {
		t1, err := common.StrToTime(reportSql.ReportStart)
		if err != nil {
			mylog.Log.Errorln(err)
			continue
		}
		t2, err := common.StrToTime(reportSql.ReportEnd)
		if err != nil {
			mylog.Log.Errorln(err)
			continue
		}
		diff := t2.Sub(t1)
		diffSec := diff.Seconds()
		var status = LampReportStatus{
			Mac:                  reportSql.Mac,
			Evaluation:           reportSql.Evaluation,
			EvaluationTimeLen:    int64(diffSec),
			Efficiency:           reportSql.StudyEfficiency,
			EfficiencyTimeLen:    int64(diffSec * float64(reportSql.StudyEfficiency) / 100),
			Concentration:        reportSql.Concentration,
			ConcentrationTimeLen: int64(diffSec * float64(reportSql.Concentration) / 100),
		}
		*results = append(*results, status)
	}
}

/******************************************************************************
 * function: handleReportSubmit
 * description:
 * return {*}
********************************************************************************/
func handleReportSubmit(lampMqttMsg *LampMqttMsg) {
	var reportJson *LampReportJson = &LampReportJson{}
	err := json.Unmarshal([]byte(lampMqttMsg.Data), &reportJson)
	if err != nil {
		mylog.Log.Errorln(err)
	}
	var reportSql *LampReportSql = NewLampReportSql()
	reportSql.Mac = lampMqttMsg.Mac
	var tm time.Duration = time.Duration(reportJson.ReportStart) * time.Second
	reportSql.ReportStart = time.Unix(int64(tm.Seconds()), 0).Format(cfg.TmFmtStr)
	tm = time.Duration(reportJson.ReportEnd) * time.Second
	reportSql.ReportEnd = time.Unix(int64(tm.Seconds()), 0).Format(cfg.TmFmtStr)
	reportSql.Evaluation = reportJson.Evaluation
	reportSql.StudyEfficiency = reportJson.StudyEfficiency
	reportSql.Concentration = reportJson.Concentration
	reportSql.SeqInterval = reportJson.SeqInterval
	for i := 0; i < len(reportJson.FlowState); i += 2 {
		reportSql.FlowState = reportJson.FlowState[i]
		reportSql.FlowStateTime = common.SecondsToTimeStr(int64(reportJson.FlowState[i+1]))
		respIdx := i
		if respIdx >= len(reportJson.Respiratory) {
			respIdx = len(reportJson.Respiratory) - 1
		}
		reportSql.Respiratory = reportJson.Respiratory[respIdx]
		hrIdx := i
		if hrIdx >= len(reportJson.HeartRate) {
			hrIdx = len(reportJson.HeartRate) - 1
		}
		reportSql.HeartRate = reportJson.HeartRate[hrIdx]
		psIdx := i
		if psIdx >= len(reportJson.PostureState) {
			psIdx = len(reportJson.PostureState) - 1
		}
		reportSql.PostureState = reportJson.PostureState[psIdx]
		afIdx := i
		if afIdx >= len(reportJson.ActivityFreq) {
			afIdx = len(reportJson.ActivityFreq) - 1
		}
		reportSql.ActivityFreq = reportJson.ActivityFreq[afIdx]
		bpIdx := i
		if bpIdx >= len(reportJson.BodyPos) {
			bpIdx = len(reportJson.BodyPos) - 1
		}
		reportSql.BodyPos = reportJson.BodyPos[bpIdx]
		taskPool.Put(&gopool.Task{
			Params: []interface{}{reportSql},
			Do: func(params ...interface{}) {
				var obj = params[0].(*LampReportSql)
				obj.Insert()
			},
		})
	}
}

type LampControlJson struct {
	Model      int    `json:"ctrl_mode"`
	StartTime  string `json:"start_time"`
	EndTime    string `json:"end_time"`
	Switch     int    `json:"switch"`
	Select     int    `json:"select"`
	BrightNess int    `json:"brightness"`
	ColorTemp  int    `json:"color_temp"`
}

func (me *LampControlJson) SendControl(mac string) {
	var msg *LampMqttMsg = NewLampMqttMsg()
	msg.Cmd = ControlLamp
	msg.Sn = makeSn()
	msg.Ts = time.Now().Unix()
	msg.Mac = mac
	js, err := json.Marshal(me)
	if err != nil {
		mylog.Log.Errorln(err)
	}
	msg.Data = string(js)
	mq.PublishData(MakeHl77PublishTopicByMac(mac), msg)
	obj := NewLampControlSql()
	obj.Mac = mac
	obj.Model = me.Model
	obj.StartTime = me.StartTime
	obj.EndTime = me.EndTime
	obj.Switch = me.Switch
	obj.Select = me.Select
	obj.BrightNess = me.BrightNess
	obj.ColorTemp = me.ColorTemp
	var objs []LampControlSql
	QueryLampControlByCond("mac='"+mac+"'", nil, nil, 1, &objs)
	if len(objs) > 0 {
		obj.ID = objs[0].ID
		obj.Update()
	} else {
		obj.Insert()
	}
}

// swagger:model LampControlSql
type LampControlSql struct {
	ID         int64  `json:"id" mysql:"id"`
	Mac        string `json:"mac" mysql:"mac"`
	Model      int    `json:"ctrl_mode" mysql:"ctrl_mode"`
	StartTime  string `json:"start_time" mysql:"start_time"`
	EndTime    string `json:"end_time" mysql:"end_time"`
	Switch     int    `json:"switch" mysql:"switch"`
	Select     int    `json:"select" mysql:"sel"`
	BrightNess int    `json:"brightness" mysql:"brightness"`
	ColorTemp  int    `json:"color_temp" mysql:"color_temp"`
	CreateTime string `json:"create_time" mysql:"create_time"`
}

func NewLampControlSql() *LampControlSql {
	return &LampControlSql{
		ID:         0,
		Mac:        "",
		Model:      0,
		StartTime:  "",
		EndTime:    "",
		Switch:     0,
		Select:     0,
		BrightNess: 0,
		ColorTemp:  0,
		CreateTime: common.GetNowTime(),
	}
}
func (me *LampControlSql) myTable() string {
	return common.LampControlTbl
}
func QueryLampControlByCond(filter interface{}, page *common.PageDao, sort interface{}, limited int, results *[]LampControlSql) bool {
	res := false
	backFunc := func(rows *sql.Rows) {
		obj := NewLampControlSql()
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*results = append(*results, *obj)
		}
	}
	if page == nil {
		res = QueryDao(common.LampControlTbl, filter, sort, limited, backFunc)
	} else {
		res = QueryPage(common.LampControlTbl, page, filter, sort, backFunc)
	}
	return res
}
func (me *LampControlSql) DecodeFromGin(c *gin.Context) {
	if err := c.ShouldBindBodyWith(me, binding.JSON); err != nil {
		exception.Throw(http.StatusAccepted, err.Error())
	}
	if me.Mac == "" {
		exception.Throw(http.StatusAccepted, "mac is empty!")
	}
}
func (me *LampControlSql) DecodeFromRows(rows *sql.Rows) error {
	err := rows.Scan(&me.ID,
		&me.Mac,
		&me.Model,
		&me.StartTime,
		&me.EndTime,
		&me.Switch,
		&me.Select,
		&me.BrightNess,
		&me.ColorTemp,
		&me.CreateTime,
	)
	return err
}
func (me *LampControlSql) DecodeFromRow(row *sql.Row) error {
	err := row.Scan(&me.ID,
		&me.Mac,
		&me.Model,
		&me.StartTime,
		&me.EndTime,
		&me.Switch,
		&me.Select,
		&me.BrightNess,
		&me.ColorTemp,
		&me.CreateTime,
	)
	return err
}
func (me *LampControlSql) QueryByID(id int64) bool {
	return QueryDaoByID(me.myTable(), id, me)
}
func (me *LampControlSql) Insert() bool {
	tblName := me.myTable()
	if !CheckTableExist(tblName) {
		sql := `create table ` + tblName + ` (
			id MEDIUMINT NOT NULL AUTO_INCREMENT,
			mac char(32) NOT NULL COMMENT 'mac地址',
			ctrl_mode int NOT NULL COMMENT '控制模式',
			start_time varchar(32) NOT NULL COMMENT '开始时间',
			end_time varchar(32) NOT NULL COMMENT '结束时间',
			switch int NOT NULL COMMENT '开关',
			sel int NOT NULL COMMENT '选择',
			brightness int NOT NULL COMMENT '亮度',
			color_temp int NOT NULL COMMENT '色温',
			create_time datetime comment '新增日期',
			PRIMARY KEY (id, mac)
		)`
		CreateTable(sql)
	}
	return InsertDao(tblName, me)
}
func (me *LampControlSql) Update() bool {
	return UpdateDaoByID(common.LampControlTbl, me.ID, me)
}
func (me *LampControlSql) Delete() bool {
	return DeleteDaoByID(common.LampControlTbl, me.ID)
}
func (me *LampControlSql) SetID(id int64) {
	me.ID = id
}

/**
 * @description: 向设备发送MQ读取灯的控制状态
 * @param {string} mac
 * @return {*}
 */
func readLampControlStatus(mac string) {
	var msg *LampMqttMsg = NewLampMqttMsg()
	msg.Cmd = ReadLampStatus
	msg.Sn = makeSn()
	msg.Ts = time.Now().Unix()
	msg.Mac = mac
	type Version struct {
		Version string `json:"version"`
	}
	var version *Version = &Version{Version: ""}
	var otaLst []LampOtaSql
	QueryLampOtaByCond("upgrade=1", &otaLst)
	if len(otaLst) > 0 {
		version.Version = fmt.Sprintf("%d_%d", otaLst[0].RemoteBaseVersion, otaLst[0].RemoteCoreVersion)
	}
	js, err := json.Marshal(version)
	if err != nil {
		mylog.Log.Errorln(err)
		return
	}
	msg.Data = string(js)
	mq.PublishData(MakeHl77PublishTopicByMac(mac), msg)
}

/******************************************************************************
 * function: handleControlLampRsp
 * description: response from lamp device
 * return {*}
********************************************************************************/
type LampControlRsp struct {
	Mac        string `json:"mac" mysql:"mac"`
	Model      int    `json:"ctrl_mode" mysql:"ctrl_mode"`
	Switch     int    `json:"switch" mysql:"switch"`
	BrightNess int    `json:"brightness" mysql:"brightness"`
	ColorTemp  int    `json:"color_temp" mysql:"color_temp"`
}

/**
 * @description: receive lamp control response from device, save to db and publish mqtt message
 * @param {*LampMqttMsg} lampMqttMsg
 * @return {*}
 */
func handleControlLampRsp(lampMqttMsg *LampMqttMsg) {
	var lampControlRsp *LampControlRsp = &LampControlRsp{}
	err := json.Unmarshal([]byte(lampMqttMsg.Data), &lampControlRsp)
	if err != nil {
		mylog.Log.Errorln(err)
	}
	lampControlRsp.Mac = lampMqttMsg.Mac
	filter := fmt.Sprintf("mac='%s'", lampMqttMsg.Mac)
	var objs []LampControlSql
	QueryLampControlByCond(filter, nil, nil, 0, &objs)
	if len(objs) > 0 {
		objs[0].Model = lampControlRsp.Model
		objs[0].Switch = lampControlRsp.Switch
		objs[0].BrightNess = lampControlRsp.BrightNess
		objs[0].ColorTemp = lampControlRsp.ColorTemp
		objs[0].Update()
	} else {
		var obj *LampControlSql = NewLampControlSql()
		obj.Mac = lampMqttMsg.Mac
		obj.Model = lampControlRsp.Model
		obj.Switch = lampControlRsp.Switch
		obj.BrightNess = lampControlRsp.BrightNess
		obj.ColorTemp = lampControlRsp.ColorTemp
		obj.CreateTime = common.GetNowTime()
		obj.Insert()
	}
	mq.PublishData(common.MakeHl77ControlStatusTopic(lampMqttMsg.Mac), lampControlRsp)
}

/******************************************************************************
 * description: define class named studyroom
 * return {*}
********************************************************************************/

// swagger:model StudyRoom
type StudyRoom struct {
	ID         int64  `json:"id" mysql:"id"`
	Name       string `json:"name" mysql:"name"`
	CreateId   int64  `json:"create_id" mysql:"create_id"`
	Capacity   int    `json:"capacity" mysql:"capacity"`
	CurrentNum int    `json:"current_num" mysql:"current_num"`
	Status     int    `json:"status" mysql:"status"`
	CreateTime string `json:"create_time" mysql:"create_time"`
	CloseTime  string `json:"close_time" mysql:"close_time"`
}

func NewStudyRoom() *StudyRoom {
	return &StudyRoom{
		ID:         0,
		Name:       "",
		CreateId:   0,
		Capacity:   0,
		CurrentNum: 0,
		Status:     0,
		CreateTime: common.GetNowTime(),
		CloseTime:  common.GetNowTime(),
	}
}

func QueryStudyRoomByCond(filter interface{}, page *common.PageDao, sort interface{}, limited int, results *[]StudyRoom) bool {
	res := false
	backFunc := func(rows *sql.Rows) {
		obj := NewStudyRoom()
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*results = append(*results, *obj)
		}
	}
	if page == nil {
		res = QueryDao(common.StudyRoomTbl, filter, sort, limited, backFunc)
	} else {
		res = QueryPage(common.StudyRoomTbl, page, filter, sort, backFunc)
	}
	return res
}

func QueryStudyRoomByUser(userId int64, results *[]StudyRoom) bool {
	backFunc := func(rows *sql.Rows) {
		obj := NewStudyRoom()
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
			return
		} else {
			*results = append(*results, *obj)
		}
	}
	sql := "select a.* from " + common.StudyRoomTbl + " a," + common.StudyRoomUserTbl +
		" b where a.id=b.room_id and a.status=1 and b.status=1"
	if userId > 0 {
		sql += " and b.user_id=" + strconv.FormatInt(userId, 10)
	}
	rows, err := mDb.Query(sql)
	if err != nil {
		mylog.Log.Errorln(err)
		return false
	}
	defer rows.Close()
	for rows.Next() {
		backFunc(rows)
	}
	return true
}

func (me *StudyRoom) DecodeFromGin(c *gin.Context) {
	if err := c.ShouldBindBodyWith(me, binding.JSON); err != nil {
		exception.Throw(http.StatusAccepted, err.Error())
	}
	if me.Name == "" || me.CreateId == 0 {
		exception.Throw(http.StatusAccepted, "Name not empty and create id not null!")
	}
}

func (me *StudyRoom) DecodeFromRows(rows *sql.Rows) error {
	err := rows.Scan(&me.ID,
		&me.Name,
		&me.CreateId,
		&me.Capacity,
		&me.CurrentNum,
		&me.Status,
		&me.CreateTime,
		&me.CloseTime,
	)
	return err
}
func (me *StudyRoom) DecodeFromRow(row *sql.Row) error {
	err := row.Scan(&me.ID,
		&me.Name,
		&me.CreateId,
		&me.Capacity,
		&me.CurrentNum,
		&me.Status,
		&me.CreateTime,
		&me.CloseTime,
	)
	return err
}

func (me *StudyRoom) QueryByID(id int64) bool {
	return QueryDaoByID(common.StudyRoomTbl, id, me)
}

/*
 */
func (me *StudyRoom) Insert() bool {
	tblName := common.StudyRoomTbl
	if !CheckTableExist(tblName) {
		sql := `create table ` + tblName + ` (
            id MEDIUMINT NOT NULL AUTO_INCREMENT,
			name varchar(64) NOT NULL COMMENT '房间名称',
			create_id int NOT NULL COMMENT '创建人id',
			capacity int NOT NULL default 6 COMMENT '容量',
			current_num int NOT NULL default 0 COMMENT '当前人数',
			status int NOT NULL default 1 COMMENT '状态 1:使用 0:关闭',
			create_time datetime comment '新增日期',
			close_time datetime comment '关闭日期',
            PRIMARY KEY (id, create_id)

        )`
		CreateTable(sql)
	}
	return InsertDao(tblName, me)
}

/*
 */
func (me *StudyRoom) Update() bool {
	return UpdateDaoByID(common.StudyRoomTbl, me.ID, me)
}

/*
 */
func (me *StudyRoom) Delete() bool {
	return DeleteDaoByID(common.StudyRoomTbl, me.ID)
}

/*
设置ID
*/
func (me *StudyRoom) SetID(id int64) {
	me.ID = id
}

/******************************************************************************
 * description:  define class named studyroom user
********************************************************************************/

// swagger:model StudyRoomUser
type StudyRoomUser struct {
	ID     int64 `json:"id" mysql:"id"`
	RoomId int64 `json:"room_id" mysql:"room_id"`
	UserId int64 `json:"user_id" mysql:"user_id"`
	// 1:邀请 0:移除
	Status     int    `json:"status" mysql:"status"`
	Sn         int    `json:"sn" mysql:"sn"`
	CreateTime string `json:"create_time" mysql:"create_time"`
}

func NewStudyRoomUser() *StudyRoomUser {
	return &StudyRoomUser{
		ID:         0,
		RoomId:     0,
		UserId:     0,
		Status:     0,
		Sn:         0,
		CreateTime: common.GetNowTime(),
	}
}
func QueryStudyRoomUserByCond(filter interface{}, page *common.PageDao, sort interface{}, limited int, results *[]StudyRoomUser) bool {
	res := false
	backFunc := func(rows *sql.Rows) {
		obj := NewStudyRoomUser()
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*results = append(*results, *obj)
		}
	}
	if page == nil {
		res = QueryDao(common.StudyRoomUserTbl, filter, sort, limited, backFunc)
	} else {
		res = QueryPage(common.StudyRoomUserTbl, page, filter, sort, backFunc)
	}
	return res
}
func UserInAnyStudyRoom(userId int64) bool {
	var results []StudyRoomUser
	filter := fmt.Sprintf("user_id=%d and status=1", userId)
	QueryStudyRoomUserByCond(filter, nil, nil, 1, &results)
	return len(results) > 0
}
func CleanUserStudyRoomStatus(userId int64, roomId int64) {
	var sql string
	if userId > 0 {
		sql = "update " + common.StudyRoomUserTbl +
			" set status=0 where status=1 and user_id=" + strconv.FormatInt(userId, 10) + " and room_id=" + strconv.FormatInt(roomId, 10)
	} else {
		sql = "update " + common.StudyRoomUserTbl +
			" set status=0 where status=1 and room_id=" + strconv.FormatInt(roomId, 10)
	}
	_, err := mDb.Exec(sql)
	if err != nil {
		mylog.Log.Errorln(err)
	}
}

func (me *StudyRoomUser) DecodeFromGin(c *gin.Context) {
	if err := c.ShouldBindBodyWith(me, binding.JSON); err != nil {
		exception.Throw(http.StatusAccepted, err.Error())
	}
}

func (me *StudyRoomUser) DecodeFromRows(rows *sql.Rows) error {
	err := rows.Scan(&me.ID,
		&me.RoomId,
		&me.UserId,
		&me.Status,
		&me.Sn,
		&me.CreateTime,
	)
	return err
}
func (me *StudyRoomUser) DecodeFromRow(row *sql.Row) error {
	err := row.Scan(&me.ID,
		&me.RoomId,
		&me.UserId,
		&me.Status,
		&me.Sn,
		&me.CreateTime,
	)
	return err
}

func (me *StudyRoomUser) QueryByID(id int64) bool {
	return QueryDaoByID(common.StudyRoomUserTbl, id, me)
}

/*
 */
func (me *StudyRoomUser) Insert() bool {
	tblName := common.StudyRoomUserTbl
	if !CheckTableExist(tblName) {
		sql := `create table ` + tblName + ` (
            id MEDIUMINT NOT NULL AUTO_INCREMENT,
			room_id int NOT NULL COMMENT '房间id',
			user_id int NOT NULL COMMENT '用户id',
			status int NOT NULL default 1 COMMENT '状态 1:邀请 0:移除',
			sn int NOT NULL default 0 COMMENT '序号',
			create_time datetime comment '创建时间',
            PRIMARY KEY (id)

        )`
		CreateTable(sql)
	}
	return InsertDao(tblName, me)
}

/*
 */
func (me *StudyRoomUser) Update() bool {
	return UpdateDaoByID(common.StudyRoomUserTbl, me.ID, me)
}

/*
 */
func (me *StudyRoomUser) Delete() bool {
	return DeleteDaoByID(common.StudyRoomUserTbl, me.ID)
}

/*
设置ID
*/
func (me *StudyRoomUser) SetID(id int64) {
	me.ID = id
}

/******************************************************************************
 * function: QueryStudyRoomUserDetailByRoomId
 * description: query study room users detail information with room id
 * param {int64} roomId
 * param {*[]StudyRoomUser} results
 * return {*}
********************************************************************************/

// swagger:model StudyRoomUserDetail
type StudyRoomUserDetail struct {
	ID         int64  `json:"id" mysql:"id"`
	RoomId     int64  `json:"room_id" mysql:"room_id"`
	RoomName   string `json:"room_name" mysql:"room_name"`
	CreateId   int64  `json:"create_id" mysql:"create_id"`
	Capacity   int    `json:"capacity" mysql:"capacity"`
	CurrentNum int    `json:"current_num" mysql:"current_num"`
	Sn         int    `json:"sn" mysql:"sn"`
	UserId     int64  `json:"user_id" mysql:"user_id"`
	NickName   string `json:"nick_name" mysql:"nick_name"`
	UserPhone  string `json:"user_phone" mysql:"user_phone"`
	CreateTime string `json:"create_time" mysql:"create_time"`
}

func (me *StudyRoomUserDetail) DecodeFromRows(rows *sql.Rows) error {
	err := rows.Scan(&me.ID,
		&me.RoomId,
		&me.RoomName,
		&me.CreateId,
		&me.Capacity,
		&me.CurrentNum,
		&me.Sn,
		&me.UserId,
		&me.NickName,
		&me.UserPhone,
		&me.CreateTime,
	)
	return err
}

/******************************************************************************
 * function: QueryStudyRoomUserDetailByRoomId
 * description:
 * param {int64} roomId
 * param {int} flag 0:all 1:study user
 * param {*[]StudyRoomUserDetail} results
 * return {*}
********************************************************************************/
func QueryStudyRoomUserDetailByRoomId(roomId int64, createId int64, flag int, results *[]StudyRoomUserDetail) bool {
	res := false
	backFunc := func(rows *sql.Rows) {
		obj := &StudyRoomUserDetail{}
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*results = append(*results, *obj)
		}
	}
	var sql string
	if flag == 0 {
		sql = "select a.id, a.room_id, b.name, b.create_id, b.capacity, b.current_num, a.sn, a.user_id, c.nick_name, c.phone, a.create_time from " +
			common.StudyRoomUserTbl + " a left join " +
			common.StudyRoomTbl +
			" b on a.room_id=b.id left join " +
			common.UserTbl + " c on a.user_id=c.id where a.status=1 "

	} else {
		sql = "select a.id, a.room_id, b.name, b.create_id, b.capacity, b.current_num, a.sn, a.user_id, c.nick_name, c.phone, a.create_time from " +
			common.StudyRoomUserTbl + " a left join " +
			common.StudyRoomTbl +
			" b on a.room_id=b.id left join " +
			common.UserTbl + " c on a.user_id=c.id left join " +
			common.StudyRecordTbl + " d on c.id=d.user_id and b.id=d.room_id where a.status=1 and d.status=1 "
	}
	if roomId > 0 {
		sql += " and a.room_id=" + strconv.FormatInt(roomId, 10)
	}
	if createId > 0 {
		sql += " and b.create_id=" + strconv.FormatInt(createId, 10)
	}

	rows, err := mDb.Query(sql)
	if err != nil {
		mylog.Log.Errorln(err)
		return false
	}
	defer rows.Close()
	for rows.Next() {
		backFunc(rows)
	}
	return res
}

/******************************************************************************
 * function: QueryInviteStudyRoomDetailByUserId
 * description: query study room detail with user id who invited
 * param {int64} userId
 * param {*[]StudyRoomUserDetail} results
 * return {*}
********************************************************************************/
func QueryInviteStudyRoomDetailByUserId(userId int64, results *[]StudyRoomUserDetail) bool {
	res := false
	backFunc := func(rows *sql.Rows) {
		obj := &StudyRoomUserDetail{}
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*results = append(*results, *obj)
		}
	}
	sql := "select a.id, a.room_id, b.name, b.create_id, b.capacity, b.current_num, a.sn, a.user_id, c.nick_name, c.phone, a.create_time from " +
		common.StudyRoomUserTbl + " a left join " +
		common.StudyRoomTbl +
		" b on a.room_id=b.id left join " +
		common.UserTbl + " c on a.user_id=c.id where a.user_id=" + strconv.FormatInt(userId, 10) +
		" and a.status=1 "

	rows, err := mDb.Query(sql)
	if err != nil {
		mylog.Log.Errorln(err)
		return false
	}
	defer rows.Close()
	for rows.Next() {
		backFunc(rows)
	}
	return res
}

// swagger:model LampUserWithStudyRoom
type LampUserWithStudyRoom struct {
	UserId   int64  `json:"user_id" mysql:"user_id"`
	NickName string `json:"nick_name" mysql:"nick_name"`
	Phone    string `json:"phone" mysql:"phone"`
	RoomId   int64  `json:"room_id" mysql:"room_id"`
}

func (me *LampUserWithStudyRoom) DecodeFromRows(rows *sql.Rows) error {
	err := rows.Scan(
		&me.UserId,
		&me.NickName,
		&me.Phone,
		&me.RoomId,
	)
	return err
}

/******************************************************************************
 * function: QueryLampUsersDetailByRoomId
 * description: query users deatil who add lamp device and join study room
 * param {string} roomId
 * param {*[]LampUserWithStudyRoom} results
 * return {*}
********************************************************************************/
func QueryLampUsersDetailByRoomId(roomId string, results *[]LampUserWithStudyRoom) bool {
	var sql string
	if roomId > "0" {
		sql = "select a.*, ifnull(b.room_id, 0) as room_id from " +
			" (select distinct  a.user_id, b.nick_name, b.phone from " +
			common.UserDeviceRelationTbl + " a, " + common.UserTbl + " b, " + common.DeviceTbl + " c " +
			" where a.user_id=b.id and a.device_id=c.id and c.type='lamp_type') a left join " +
			" (select distinct room_id, user_id from " + common.StudyRoomUserTbl + " where status=1) b on a.user_id=b.user_id and b.room_id=" + roomId
	} else {
		sql = "select a.*, ifnull(b.room_id, 0) as room_id from " +
			" (select distinct  a.user_id, b.nick_name, b.phone from " +
			common.UserDeviceRelationTbl + " a, " + common.UserTbl + " b, " + common.DeviceTbl + " c " +
			" where a.user_id=b.id and a.device_id=c.id and c.type='lamp_type') a left join " +
			" (select distinct room_id, user_id from " + common.StudyRoomUserTbl + " where status=1) b on a.user_id=b.user_id"
	}
	rows, err := mDb.Query(sql)
	if err != nil {
		mylog.Log.Errorln(err)
		return false
	}
	defer rows.Close()
	for rows.Next() {
		obj := &LampUserWithStudyRoom{}
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*results = append(*results, *obj)
		}
	}
	return true
}

func QueryLampUsersInFriend(userId string, results *[]LampUserWithStudyRoom) bool {
	sql := "select a.*, ifnull(b.room_id, 0) as room_id from " +
		" (select distinct  a.user_id, b.nick_name, b.phone from " +
		common.UserDeviceRelationTbl + " a, " + common.UserTbl + " b, " + common.DeviceTbl + " c, " + common.FriendsTbl + " d " +
		" where a.user_id=b.id and a.device_id=c.id and c.type='lamp_type' and d.friend_id=b.id and d.user_id=" + userId + ") a left join " +
		" (select distinct room_id, user_id from " + common.StudyRoomUserTbl + " where status=1) b on a.user_id=b.user_id"
	rows, err := mDb.Query(sql)
	if err != nil {
		mylog.Log.Errorln(err)
		return false
	}
	defer rows.Close()
	for rows.Next() {
		obj := &LampUserWithStudyRoom{}
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*results = append(*results, *obj)
		}
	}
	return true
}

/******************************************************************************
 * description:  define class named user study record
********************************************************************************/

// swagger:model UserStudyRecord
type UserStudyRecord struct {
	ID     int64 `json:"id" mysql:"id"`
	UserId int64 `json:"user_id" mysql:"user_id"`
	RoomId int64 `json:"room_id" mysql:"room_id"`
	// 1:进入 0:离开
	Status    int    `json:"status" mysql:"status"`
	Sn        int    `json:"sn" mysql:"sn"`
	EnterTime string `json:"enter_time" mysql:"enter_time"`
	LeaveTime string `json:"leave_time" mysql:"leave_time"`
}

func QueryUserStudyRecordByCond(filter interface{}, page *common.PageDao, sort interface{}, limited int, results *[]UserStudyRecord) bool {
	res := false
	backFunc := func(rows *sql.Rows) {
		obj := NewUserStudyRecord()
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*results = append(*results, *obj)
		}
	}
	if page == nil {
		res = QueryDao(common.StudyRecordTbl, filter, sort, limited, backFunc)
	} else {
		res = QueryPage(common.StudyRecordTbl, page, filter, sort, backFunc)
	}
	return res
}

/******************************************************************************
 * function: CleanStudyRecordStatus
 * description: clean study record status with user id and room id
 * param {int64} userId
 * param {int64} roomId
 * return {*}
********************************************************************************/
func CleanStudyRecordStatus(userId int64, roomId int64) {
	var sql string
	if userId <= 0 {
		sql = "update " + common.StudyRecordTbl + " set status=0, leave_time=now() where room_id=" + strconv.FormatInt(roomId, 10) + " and status=1 order by enter_time desc limit 1"
	} else {
		sql = "update " + common.StudyRecordTbl + " set status=0, leave_time=now() where user_id=" + strconv.FormatInt(userId, 10) +
			" and room_id=" + strconv.FormatInt(roomId, 10) + " and status=1 order by enter_time desc limit 1"
	}
	mylog.Log.Infoln(sql)
	_, err := mDb.Exec(sql)
	if err != nil {
		mylog.Log.Errorln(err)
	}
}

func NewUserStudyRecord() *UserStudyRecord {
	return &UserStudyRecord{
		ID:        0,
		UserId:    0,
		RoomId:    0,
		Status:    0,
		Sn:        0,
		EnterTime: common.GetNowTime(),
		LeaveTime: common.GetNowTime(),
	}
}
func (me *UserStudyRecord) DecodeFromGin(c *gin.Context) {
	if err := c.ShouldBindBodyWith(me, binding.JSON); err != nil {
		exception.Throw(http.StatusAccepted, err.Error())
	}
}

func (me *UserStudyRecord) DecodeFromRows(rows *sql.Rows) error {
	err := rows.Scan(&me.ID,
		&me.UserId,
		&me.RoomId,
		&me.Status,
		&me.Sn,
		&me.EnterTime,
		&me.LeaveTime,
	)
	return err
}
func (me *UserStudyRecord) DecodeFromRow(row *sql.Row) error {
	err := row.Scan(&me.ID,
		&me.UserId,
		&me.RoomId,
		&me.Status,
		&me.Sn,
		&me.EnterTime,
		&me.LeaveTime,
	)
	return err
}

func (me *UserStudyRecord) QueryByID(id int64) bool {
	return QueryDaoByID(common.StudyRecordTbl, id, me)
}

/*
 */
func (me *UserStudyRecord) Insert() bool {
	tblName := common.StudyRecordTbl
	if !CheckTableExist(tblName) {
		sql := `create table ` + tblName + ` (
            id MEDIUMINT NOT NULL AUTO_INCREMENT,
			user_id int NOT NULL COMMENT '用户id',
			room_id int NOT NULL COMMENT '房间id',
			status int NOT NULL default 1 COMMENT '状态 1:进入 0:离开',
			sn int NOT NULL default 0 COMMENT '序号',
			enter_time datetime comment '进入时间',
			leave_time datetime comment '离开时间',
			PRIMARY KEY (id)
        )`
		CreateTable(sql)
	}
	return InsertDao(tblName, me)
}

/*
 */
func (me *UserStudyRecord) Update() bool {
	return UpdateDaoByID(common.StudyRecordTbl, me.ID, me)
}

/*
 */
func (me *UserStudyRecord) Delete() bool {
	return DeleteDaoByID(common.StudyRecordTbl, me.ID)
}

/*
设置ID
*/
func (me *UserStudyRecord) SetID(id int64) {
	me.ID = id
}

/******************************************************************************
* description:  define class named study room ranking
* return {*}
********************************************************************************/

// swagger:model StudyRoomRanking
type StudyRoomRanking struct {
	// 房间id
	RoomId int64 `json:"room_id" mysql:"room_id"`
	// 房间名称
	RoomName string `json:"room_name" mysql:"room_name"`
	// 用户id
	UserId int64 `json:"user_id" mysql:"user_id"`
	// 用户昵称
	NickName string `json:"nick_name" mysql:"nick_name"`
	// 用户手机号
	Phone string `json:"phone" mysql:"phone"`
	// 总天数
	TotayDays int `json:"today_days" mysql:"today_days"`
	// 总时长 秒
	TotalSeconds int64 `json:"total_seconds" mysql:"total_seconds"`
}

func (me *StudyRoomRanking) DecodeFromRows(rows *sql.Rows) error {
	err := rows.Scan(
		&me.RoomId,
		&me.RoomName,
		&me.UserId,
		&me.NickName,
		&me.Phone,
		&me.TotayDays,
		&me.TotalSeconds,
	)
	return err
}

/******************************************************************************
 * function: QueryRankingByStudyRoom
 * description: query ranking by study room
 * param {int64} roomId
 * return {*}
********************************************************************************/
func QueryRankingByStudyRoom(roomId int64, results *[]StudyRoomRanking) bool {
	res := false
	backFunc := func(rows *sql.Rows) {
		obj := &StudyRoomRanking{}
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*results = append(*results, *obj)
		}
	}
	var sql string
	if roomId <= 0 {
		sql = "select 0 as room_id, '' as name, a.user_id, c.nick_name, c.phone, a.total_days, a.total_seconds from " +
			" (select user_id, timestampdiff(DAY,min(date(enter_time)), max(date(leave_time))) + 1 as total_days, sum(timestampdiff(SECOND, enter_time, leave_time)) as total_seconds from " +
			common.StudyRecordTbl + " group by user_id) a join " +
			common.UserTbl + " c on a.user_id=c.id order by a.total_days desc, a.total_seconds desc"
	} else {
		sql = "select a.room_id, b.name, a.user_id, c.nick_name, c.phone, a.total_days, a.total_seconds from " +
			" (select room_id, user_id, timestampdiff(DAY,min(date(enter_time)), max(date(leave_time))) + 1 as total_days, sum(timestampdiff(SECOND, enter_time, leave_time)) as total_seconds from " +
			common.StudyRecordTbl + " where room_id=" + strconv.FormatInt(roomId, 10) + " group by user_id) a join " +
			common.StudyRoomTbl + " b on a.room_id=b.id join " +
			common.UserTbl + " c on a.user_id=c.id order by a.total_days desc, a.total_seconds desc"
	}

	rows, err := mDb.Query(sql)
	if err != nil {
		mylog.Log.Errorln(err)
		return false
	}
	defer rows.Close()
	for rows.Next() {
		backFunc(rows)
	}
	return res
}

// swagger:model UserStudyRoomData
type UserStudyRoomData struct {
	// 房间id
	RoomId int64 `json:"room_id" mysql:"room_id"`
	// 用户id
	UserId int64 `json:"user_id" mysql:"user_id"`
	// 总时长
	TotalSeconds int64 `json:"total_seconds" mysql:"total_seconds"`
	// 总天数
	TotalDays int `json:"total_days" mysql:"total_days"`
	// 平均时长
	AvgSeconds int64 `json:"avg_seconds" mysql:"avg_seconds"`
	// 最大时长
	MaxSeconds int64 `json:"max_seconds" mysql:"max_seconds"`
}

func (me *UserStudyRoomData) DecodeFromRows(rows *sql.Rows) error {
	err := rows.Scan(
		&me.RoomId,
		&me.UserId,
		&me.TotalSeconds,
		&me.TotalDays,
		&me.AvgSeconds,
		&me.MaxSeconds,
	)
	return err
}
func QueryUserStudyRoomData(userId int64, roomId int64, startTm string, endTm string, results *[]UserStudyRoomData) bool {
	backFunc := func(rows *sql.Rows) {
		obj := &UserStudyRoomData{}
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*results = append(*results, *obj)
		}
	}
	var sql string
	if roomId <= 0 {
		sql = "select 0 as room_id, b.user_id, b.total_seconds, b.total_days, b.avg_seconds, c.max_seconds from " +
			"(select *, (case total_days when 0 then 0 else convert(total_seconds/total_days, signed) end) as avg_seconds from " +
			"(select user_id, sum(timestampdiff(second, enter_time, leave_time)) as total_seconds, " +
			"timestampdiff(DAY,min(date(enter_time)), max(date(leave_time)))+1 as total_days from " +
			common.StudyRecordTbl + " where date(enter_time)>=date('" + startTm + "') and date(leave_time)<=date('" + endTm + "') and status=0 group by user_id) a) b left join " +
			"(select user_id, max(timestampdiff(second,enter_time, leave_time)) as max_seconds from " +
			common.StudyRecordTbl + " where date(enter_time)>=date('" + startTm + "') and date(leave_time)<=date('" + endTm + "') and status=0 group by user_id) c " +
			"on b.user_id=c.user_id where b.user_id=" + strconv.FormatInt(userId, 10)
	} else {
		sql = "select b.room_id, b.user_id, b.total_seconds, b.total_days, b.avg_seconds, c.max_seconds from " +
			"(select *, (case total_days when 0 then 0 else convert(total_seconds/total_days, signed) end) as avg_seconds from " +
			"(select room_id, user_id, sum(timestampdiff(second, enter_time, leave_time)) as total_seconds, " +
			"timestampdiff(DAY,min(date(enter_time)), max(date(leave_time)))+1 as total_days from " +
			common.StudyRecordTbl + " where date(enter_time)>=date('" + startTm + "') and date(leave_time)<=date('" + endTm + "') and status=0 group by room_id, user_id) a) b left join " +
			"(select room_id, user_id, max(timestampdiff(second,enter_time, leave_time)) as max_seconds from " +
			common.StudyRecordTbl + " where date(enter_time)>=date('" + startTm + "') and date(leave_time)<=date('" + endTm + "') and status=0 group by room_id, user_id) c " +
			"on b.room_id=c.room_id and b.user_id=c.user_id where b.user_id=" + strconv.FormatInt(userId, 10) +
			" and b.room_id=" + strconv.FormatInt(roomId, 10)
	}

	rows, err := mDb.Query(sql)
	if err != nil {
		mylog.Log.Errorln(err)
		return false
	}
	defer rows.Close()
	for rows.Next() {
		backFunc(rows)
	}
	return true
}

// swagger:model UserStudyDataByDay
type UserStudyDataByDay struct {
	DaySeconds int64  `json:"day_seconds" mysql:"day_seconds"`
	EnterDay   string `json:"enter_day" mysql:"enter_day"`
}

func QueryUserStudyDataByDay(userId int64, roomId int64, startTm string, endTm string, result *[]UserStudyDataByDay) bool {
	var sql string
	if roomId > 0 {
		sql = "select sum(timestampdiff(second, enter_time, leave_time)) as day_seconds, date(enter_time) as enter_day " +
			" from " + common.StudyRecordTbl + " where date(enter_time)>=date('" + startTm + "') and date(leave_time)<=date('" + endTm + "') " +
			" and room_id=" + strconv.FormatInt(roomId, 10) + " and user_id=" + strconv.FormatInt(userId, 10) + " group by date(enter_time)"
	} else {
		sql = "select sum(timestampdiff(second, enter_time, leave_time)) as day_seconds, date(enter_time) as enter_day " +
			" from " + common.StudyRecordTbl + " where date(enter_time)>=date('" + startTm + "') and date(leave_time)<=date('" + endTm + "') " +
			" and user_id=" + strconv.FormatInt(userId, 10) + " group by date(enter_time)"
	}
	rows, err := mDb.Query(sql)
	if err != nil {
		mylog.Log.Errorln(err)
		return false
	}
	defer rows.Close()
	for rows.Next() {
		obj := &UserStudyDataByDay{}
		err := rows.Scan(&obj.DaySeconds, &obj.EnterDay)
		if err != nil {
			mylog.Log.Errorln(err)
			return false
		}
		*result = append(*result, *obj)
	}
	return true
}

/******************************************************************************
 * function:
 * description:
 * param {string} startTm
 * param {string} endTm
 * return {*}
********************************************************************************/
func StatsLampFlowDataByTime(mac string, startDay string, endDay string) int {
	sql := "select (case b.total_count when 0 then 0 else convert(100*a.flow_count / b.total_count, signed) end) as flow_data from " +
		" (SELECT count(*) as flow_count FROM " + common.LampRealDataTbl + " where mac like '" + mac + "'" +
		" and flow_state=2 and date(create_time) >=date('" + startDay + "') and date(create_time)<=date('" + endDay + "')) a," +
		" (SELECT count(*) as total_count FROM " + common.LampRealDataTbl + " where mac like '" + mac + "'" +
		" and flow_state>0 and date(create_time) >=date('" + startDay + "') and date(create_time)<=date('" + endDay + "')) b"
	var flowData int = 0
	rows, err := mDb.Query(sql)
	if err != nil {
		mylog.Log.Errorln(err)
		return flowData
	}
	defer rows.Close()
	if rows.Next() {
		err := rows.Scan(&flowData)
		if err != nil {
			mylog.Log.Errorln(err)
		}
	}
	return flowData
}

// swagger:model UserStudyTime
type UserStudyTime struct {
	// required: true
	UserId int64 `json:"user_id" mysql:"user_id"`
	// required: true
	EnterTime string `json:"enter_time" mysql:"enter_time"`
	//
	LeaveTime string `json:"leave_time" mysql:"leave_time"`
}

func QueryUserStudyTimeByDate(userId int64, roomId int64, startTm string, endTm string, results *[]UserStudyTime) bool {
	backFunc := func(rows *sql.Rows) {
		obj := &UserStudyTime{}
		err := rows.Scan(&obj.UserId, &obj.EnterTime, &obj.LeaveTime)
		if err != nil {
			mylog.Log.Errorln(err)
			return
		} else {
			*results = append(*results, *obj)
		}
	}
	var sql string
	if roomId <= 0 {
		sql = "select user_id, enter_time, leave_time from " +
			common.StudyRecordTbl + " where user_id=" + strconv.FormatInt(userId, 10) +
			" and date(enter_time)>=date('" + startTm + "') and date(leave_time)<=date('" + endTm + "') and status=0"
	} else {
		sql = "select user_id, enter_time, leave_time from " +
			common.StudyRecordTbl + " where user_id=" + strconv.FormatInt(userId, 10) +
			" and room_id=" + strconv.FormatInt(roomId, 10) +
			" and date(enter_time)>=date('" + startTm + "') and date(leave_time)<=date('" + endTm + "') and status=0"
	}
	rows, err := mDb.Query(sql)
	if err != nil {
		mylog.Log.Errorln(err)
		return false
	}
	defer rows.Close()
	for rows.Next() {
		backFunc(rows)
	}
	return true
}
