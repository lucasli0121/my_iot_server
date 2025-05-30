/******************************************************************************
 * Author: liguoqiang
 * Date: 2024-07-22 09:47:20
 * LastEditors: liguoqiang
 * LastEditTime: 2025-01-15 20:25:39
 * Description:
********************************************************************************/
package mysql

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"hjyserver/exception"
	"hjyserver/gopool"
	mylog "hjyserver/log"
	"hjyserver/mdb/common"
	"hjyserver/mq"
	"hjyserver/redis"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

const (
	H03_STUDY_EVENT_TOPIC_PREFIX = "server-h03/study/event/"
	H03_STUDY_ATTR_TOPIC_PREFIX  = "server-h03/study/attr/"
	H03_STUDY_DAY_REPORT_PREFIX  = "server-h03/study/dayreport/"
	deviceH03TopicPrefix         = "hjy-dev/h03/"
)

func MakeStudyEventTopic(mac string) string {
	return H03_STUDY_EVENT_TOPIC_PREFIX + strings.ToLower(mac)
}
func MakeStudyAttrTopic(mac string) string {
	return H03_STUDY_ATTR_TOPIC_PREFIX + strings.ToLower(mac)
}
func MakeStudyDayReportTopic(mac string) string {
	return H03_STUDY_DAY_REPORT_PREFIX + strings.ToLower(mac)
}

const (
	H03OnlineCmd    = 100
	H03SyncCmd      = 101
	H03SyncCmdRsp   = 201
	H03KeepAlive    = 102
	H03ErrCode      = 103
	H03Attr         = 104
	H03AttrResp     = 204
	H03EventReport  = 105
	H03ReportSubmit = 106
	H03RebootCmd    = 203
	H03SettingCmd   = 205
)

var h03SnVal int = 0

func makeH03Sn() int {
	h03SnVal++
	if h03SnVal >= math.MaxInt32 {
		h03SnVal = 0
	}
	return h03SnVal
}

func MakeH03InfoTopic(mac string) string {
	return deviceH03TopicPrefix + strings.ToLower(mac) + "/info/"
}

func MakeH03AttrTopic(mac string) string {
	return deviceH03TopicPrefix + strings.ToLower(mac) + "/attr/"
}

func MakeH03EventTopic(mac string) string {
	return deviceH03TopicPrefix + strings.ToLower(mac) + "/event/"
}

func MakeH03FuncTopic(mac string) string {
	return deviceH03TopicPrefix + strings.ToLower(mac) + "/func/"
}

func MakeH03ReportTopic(mac string) string {
	return deviceH03TopicPrefix + strings.ToLower(mac) + "/report/"
}

func MakeH03CtlTopic(mac string) string {
	return deviceH03TopicPrefix + strings.ToLower(mac) + "/ctrl/"
}

/******************************************************************************
 * function: SubscribeH03MqttTopic
 * description: 订阅H03设备的MQTT消息
 * param {string} mac
 * return {*}
********************************************************************************/
func SubscribeH03MqttTopic(mac string) {
	msgProc := NewH03MqttMsgProc()
	mq.SubscribeTopic(MakeH03InfoTopic(mac), msgProc)
	mq.SubscribeTopic(MakeH03AttrTopic(mac), msgProc)
	mq.SubscribeTopic(MakeH03EventTopic(mac), msgProc)
	// mq.SubscribeTopic(MakeH03FuncTopic(mac), msgProc)
	mq.SubscribeTopic(MakeH03ReportTopic(mac), msgProc)
}

func UnsubscribeH03MqttTopic(mac string) {
	mq.UnsubscribeTopic(MakeH03InfoTopic(mac))
	mq.UnsubscribeTopic(MakeH03AttrTopic(mac))
	mq.UnsubscribeTopic(MakeH03EventTopic(mac))
	// mq.UnsubscribeTopic(MakeH03FuncTopic(mac))
	mq.UnsubscribeTopic(MakeH03ReportTopic(mac))
}

type H03MqttMsgProc struct {
}

func NewH03MqttMsgProc() *H03MqttMsgProc {
	return &H03MqttMsgProc{}
}

func (me *H03MqttMsgProc) HandleMqttMsg(topic string, payload []byte) {
	mylog.Log.Infoln("HandleH03MqttMsg:", topic, string(payload))
	HandleH03MqttMsg(topic, payload)
}

func HandleH03MqttMsg(topic string, payload []byte) {
	var mqttMsg *H03MqttMsg = NewH03MqttMsg()
	err := json.Unmarshal([]byte(payload), &mqttMsg)
	if err != nil {
		mylog.Log.Errorln(err)
		return
	}

	switch mqttMsg.Cmd {
	case H03OnlineCmd:
		handleH03OnlineCmd(mqttMsg)
	case H03SyncCmd:
		handleH03SyncCmd(mqttMsg)
	case H03KeepAlive:
		handleH03KeepAlive(mqttMsg)
	case H03ErrCode:
		handleH03ErrCode(mqttMsg)
	case H03Attr:
		handleH03Attr(mqttMsg)
	case H03EventReport:
		handleH03Event(mqttMsg)
	case H03ReportSubmit:
		handleH03Report(mqttMsg)
	default:
	}
}

/******************************************************************************
 * function:
 * description: 定义H03 MQTT的消息结构，用于解析MQTT消息
 * return {*}
********************************************************************************/
type H03MqttMsg struct {
	Cmd  int         `json:"cmd"`
	Sn   int         `json:"s"`
	Ts   int64       `json:"time"`
	Mac  string      `json:"id"`
	Data interface{} `json:"data"`
}

func NewH03MqttMsg() *H03MqttMsg {
	return &H03MqttMsg{
		Cmd: 0,
		Sn:  0,
		Ts:  0,
		Mac: "",
	}
}

/*
*****************************************************************************
  - function:
  - description: 处理设备上线的消息
  - param {*H03MqttMsg} mqttMsg
  - return {*}
    测试字符串：{"cmd": 100, "s":  1,"time": 1,"id": "test111","data": { "deviceType": "H03pro", "rssi": -20 }}

*******************************************************************************
*/
type H03OnlineData struct {
	DeviceType string `json:"deviceType"`
	Rssi       int    `json:"rssi"`
}

func handleH03OnlineCmd(mqttMsg *H03MqttMsg) {
	mylog.Log.Infoln("h03", "handleH03OnlineCmd:", mqttMsg.Cmd)
	exception.TryEx{
		Try: func() {
			var data H03OnlineData
			mapData := mqttMsg.Data.(map[string]interface{})
			for key, value := range mapData {
				mylog.Log.Infoln("key:", key, "value:", value)
				if key == "deviceType" {
					data.DeviceType = value.(string)
				} else if key == "rssi" {
					data.Rssi = int(value.(float64))
				}
			}
			SetDeviceOnline(mqttMsg.Mac, 1, data.Rssi)
		},
		Catch: func(e exception.Exception) {
			mylog.Log.Errorln(e)
		},
	}.Run()
}

/*
*****************************************************************************
  - function:
  - description: 处理设备同步的消息
  - param {*H03MqttMsg} mqttMsg
  - return {*}
  - 测试字符串：
    {"cmd": 101, "s":  2, "time": 1720669362,"id": "543204abb252","data": {  \"deviceType\": \"H03pro\",  \"softwareVersion\": \"v1.0.0\",  \"hardwareVersion\": \"24-06-13\",  \"coreVersion\": \"v2.2\"}}

*******************************************************************************
*/

func handleH03SyncCmd(mqttMsg *H03MqttMsg) {
	// 发送同步响应消息

	data := NewH03VersionData()
	mapData := mqttMsg.Data.(map[string]interface{})
	for key, value := range mapData {
		mylog.Log.Infoln("key:", key, "value:", value)
		if key == "deviceType" {
			data.DeviceType = value.(string)
		} else if key == "softwareVersion" {
			data.SoftwareVersion = value.(string)
		} else if key == "hardwareVersion" {
			data.HardwareVersion = value.(string)
		} else if key == "coreVersion" {
			data.CoreVersion = value.(string)
		}
	}
	data.Mac = mqttMsg.Mac
	data.CreateTime = common.GetNowTime()
	GetTaskPool().Put(&gopool.Task{
		Params: []interface{}{data},
		Do: func(params ...interface{}) {
			var obj = params[0].(*H03VersionData)
			objList := make([]H03VersionData, 0)
			QueryH03VersionByMac(obj.Mac, &objList)
			if len(objList) > 0 {
				obj.ID = objList[0].ID
				obj.Update()
			} else {
				obj.Insert()
			}
		},
	})

	mqRespMsg := NewH03MqttMsg()
	syncResp := NewH03SyncOta()
	otaList := make([]H03SyncOta, 0)
	QueryH03Ota(&otaList)
	if len(otaList) > 0 {
		syncResp = &otaList[0]
	}
	syncResp.Upgrade = 0
	mqRespMsg.Cmd = H03SyncCmdRsp
	mqRespMsg.Mac = mqttMsg.Mac
	mqRespMsg.Data = syncResp
	mqRespMsg.Ts = time.Now().Unix()
	mq.PublishData(MakeH03CtlTopic(mqttMsg.Mac), mqRespMsg)
}

// swagger:model H03VersionData
type H03VersionData struct {
	ID              int64  `json:"id" mysql:"id" binding:"omitempty"`
	Mac             string `json:"mac" mysql:"mac" size:"32" comment:"mac地址"`
	DeviceType      string `json:"deviceType" size:"16" mysql:"deviceType"`
	SoftwareVersion string `json:"softwareVersion" size:"16" mysql:"softwareVersion"`
	HardwareVersion string `json:"hardwareVersion" size:"16" mysql:"hardwareVersion"`
	CoreVersion     string `json:"coreVersion" size:"16" mysql:"coreVersion"`
	CreateTime      string `json:"create_time" mysql:"create_time" binding:"datetime=2006-01-02 15:04:05" comment:"创建时间"`
}

func (H03VersionData) TableName() string {
	return H03Type + "_version_tbl"
}
func NewH03VersionData() *H03VersionData {
	return &H03VersionData{
		ID:              0,
		Mac:             "",
		DeviceType:      "",
		SoftwareVersion: "",
		HardwareVersion: "",
		CoreVersion:     "",
		CreateTime:      common.GetNowTime(),
	}
}
func (me *H03VersionData) Insert() bool {
	if !CheckTableExist(me.TableName()) {
		CreateTableWithStruct(me.TableName(), me)
	}
	return InsertDao(me.TableName(), me)
}
func (me *H03VersionData) Update() bool {
	return UpdateDaoByID(me.TableName(), me.ID, me)
}
func (me *H03VersionData) Delete() bool {
	return DeleteDaoByID(me.TableName(), me.ID)
}
func (me *H03VersionData) SetID(id int64) {
	me.ID = id
}
func (me *H03VersionData) QueryByID(id int64) bool {
	return QueryDaoByID(me.TableName(), id, me)
}
func (me *H03VersionData) DecodeFromRows(rows *sql.Rows) error {
	err := rows.Scan(
		&me.ID,
		&me.Mac,
		&me.DeviceType,
		&me.SoftwareVersion,
		&me.HardwareVersion,
		&me.CoreVersion,
		&me.CreateTime)
	return err
}
func (me *H03VersionData) DecodeFromRow(row *sql.Row) error {
	err := row.Scan(
		&me.ID,
		&me.Mac,
		&me.DeviceType,
		&me.SoftwareVersion,
		&me.HardwareVersion,
		&me.CoreVersion,
		&me.CreateTime)
	return err
}
func (me *H03VersionData) DecodeFromGin(c *gin.Context) {
	if err := c.ShouldBindBodyWith(me, binding.JSON); err != nil {
		exception.Throw(common.JsonError, err.Error())
	}
}
func QueryH03VersionByMac(mac string, results *[]H03VersionData) bool {
	filter := fmt.Sprintf("mac='%s'", mac)
	QueryDao(NewH03VersionData().TableName(), filter, nil, -1, func(rows *sql.Rows) {
		obj := NewH03VersionData()
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*results = append(*results, *obj)
		}
	})
	return true
}

// 定义H03设备同步的响应消息结构
// OTA 版本信息
type H03SyncOta struct {
	ID                int64  `json:"id" mysql:"id" binding:"omitempty"`
	Upgrade           int    `json:"upgrade" mysql:"upgrade"`
	RemoteBaseVersion string `json:"remoteBaseVersion" size:"16" mysql:"remoteBaseVersion"`
	BaseOtaUrl        string `json:"baseOtaUrl" mysql:"baseOtaUrl"`
	RemoteCoreVersion string `json:"remoteCoreVersion" size:"16" mysql:"remoteCoreVersion"`
	CoreOtaUrl        string `json:"coreOtaUrl" mysql:"coreOtaUrl"`
}

func (H03SyncOta) TableName() string {
	return H03Type + "_ota_tbl"
}
func NewH03SyncOta() *H03SyncOta {
	return &H03SyncOta{
		ID:                0,
		Upgrade:           0,
		RemoteBaseVersion: "",
		BaseOtaUrl:        "",
		RemoteCoreVersion: "",
		CoreOtaUrl:        "",
	}
}
func (me *H03SyncOta) Insert() bool {
	if !CheckTableExist(me.TableName()) {
		CreateTableWithStruct(me.TableName(), me)
	}
	return InsertDao(me.TableName(), me)
}
func (me *H03SyncOta) Update() bool {
	return UpdateDaoByID(me.TableName(), me.ID, me)
}
func (me *H03SyncOta) Delete() bool {
	return DeleteDaoByID(me.TableName(), me.ID)
}
func (me *H03SyncOta) SetID(id int64) {
	me.ID = id
}
func (me *H03SyncOta) QueryByID(id int64) bool {
	return QueryDaoByID(me.TableName(), id, me)
}
func (me *H03SyncOta) DecodeFromRows(rows *sql.Rows) error {
	err := rows.Scan(
		&me.ID,
		&me.Upgrade,
		&me.RemoteBaseVersion,
		&me.BaseOtaUrl,
		&me.RemoteCoreVersion,
		&me.CoreOtaUrl)
	return err
}
func (me *H03SyncOta) DecodeFromRow(row *sql.Row) error {
	err := row.Scan(
		&me.ID,
		&me.Upgrade,
		&me.RemoteBaseVersion,
		&me.BaseOtaUrl,
		&me.RemoteCoreVersion,
		&me.CoreOtaUrl)
	return err
}
func (me *H03SyncOta) DecodeFromGin(c *gin.Context) {
	if err := c.ShouldBindBodyWith(me, binding.JSON); err != nil {
		exception.Throw(common.JsonError, err.Error())
	}
}

/******************************************************************************
 * function: QueryH03Ota
 * description:  查询OTA信息
 * param {*[]H03SyncOta} results
 * return {*}
********************************************************************************/
func QueryH03Ota(results *[]H03SyncOta) bool {
	QueryDao(NewH03SyncOta().TableName(), nil, nil, -1, func(rows *sql.Rows) {
		obj := NewH03SyncOta()
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*results = append(*results, *obj)
		}
	})
	return true
}

/******************************************************************************
 * function:
 * description: 设备心跳消息处理
 * param {*H03MqttMsg} mqttMsg
 * return {*}
 * 测试字符串：{ "cmd": 102, "s": 2, "time": 1720669362, "id": "test111" }
********************************************************************************/
func handleH03KeepAlive(mqttMsg *H03MqttMsg) {
	SetDeviceOnline(mqttMsg.Mac, 1, 0)
}

/******************************************************************************
 * function:
 * description:
 * param {*H03MqttMsg} mqttMsg
 * return {*}
********************************************************************************/
func handleH03ErrCode(mqttMsg *H03MqttMsg) {
	var data *H03ErrorCode = &H03ErrorCode{}
	mapData := mqttMsg.Data.(map[string]interface{})
	for key, value := range mapData {
		switch key {
		case "softwareVersion":
			data.SoftwareVersion = value.(string)
		case "hardwareVersion":
			data.HardwareVersion = value.(string)
		case "coreVersion":
			data.CoreVersion = value.(string)
		case "rssi":
			data.Rssi = int(value.(float64))
		case "errorCode":
			data.ErrorCode = int(value.(float64))
		}
	}

	// 更新设备错误码，如果不存在设备就不更新
	devcieErrCodeList := make([]H03ErrorCode, 0)
	QueryH03ErrCodeByMac(mqttMsg.Mac, &devcieErrCodeList)
	if len(devcieErrCodeList) > 0 {
		devcieErrCode := devcieErrCodeList[0]
		devcieErrCode.SoftwareVersion = data.SoftwareVersion
		devcieErrCode.HardwareVersion = data.HardwareVersion
		devcieErrCode.CoreVersion = data.CoreVersion
		devcieErrCode.Rssi = data.Rssi
		devcieErrCode.ErrorCode = data.ErrorCode
		devcieErrCode.CreateTime = common.GetNowTime()
		devcieErrCode.Update()
	} else {
		data.Mac = mqttMsg.Mac
		data.CreateTime = common.GetNowTime()
		data.Insert()
	}
}

type H03ErrorCode struct {
	ID              int64  `json:"id" mysql:"id" binding:"omitempty"`
	Mac             string `json:"mac" mysql:"mac" size:"32" comment:"mac地址"`
	SoftwareVersion string `json:"softwareVersion" mysql:"softwareVersion" size:"16" comment:"主板固件版本"`
	HardwareVersion string `json:"hardwareVersion" mysql:"hardwareVersion" size:"16" comment:"硬件版本"`
	CoreVersion     string `json:"coreVersion" mysql:"coreVersion" size:"16" comment:"核心版本"`
	Rssi            int    `json:"rssi" mysql:"rssi" comment:"信号强度"`
	ErrorCode       int    `json:"errorCode" mysql:"errorCode" comment:"错误码"`
	CreateTime      string `json:"create_time" mysql:"create_time" binding:"datetime=2006-01-02 15:04:05" comment:"创建时间"`
}

func (H03ErrorCode) TableName() string {
	return H03Type + "_errcode_tbl"
}
func NewH03ErrorCode() *H03ErrorCode {
	return &H03ErrorCode{
		ID:              0,
		Mac:             "",
		SoftwareVersion: "",
		HardwareVersion: "",
		CoreVersion:     "",
		Rssi:            0,
		ErrorCode:       0,
		CreateTime:      common.GetNowTime(),
	}
}
func (me *H03ErrorCode) Insert() bool {
	if !CheckTableExist(me.TableName()) {
		CreateTableWithStruct(me.TableName(), me)
	}
	return InsertDao(me.TableName(), me)
}
func (me *H03ErrorCode) Update() bool {
	return UpdateDaoByID(me.TableName(), me.ID, me)
}
func (me *H03ErrorCode) Delete() bool {
	return DeleteDaoByID(me.TableName(), me.ID)
}
func (me *H03ErrorCode) SetID(id int64) {
	me.ID = id
}
func (me *H03ErrorCode) QueryByID(id int64) bool {
	return QueryDaoByID(me.TableName(), id, me)
}
func (me *H03ErrorCode) DecodeFromRows(rows *sql.Rows) error {
	err := rows.Scan(
		&me.ID,
		&me.Mac,
		&me.SoftwareVersion,
		&me.HardwareVersion,
		&me.CoreVersion,
		&me.Rssi,
		&me.ErrorCode,
		&me.CreateTime)
	return err
}
func (me *H03ErrorCode) DecodeFromRow(row *sql.Row) error {
	err := row.Scan(
		&me.ID,
		&me.Mac,
		&me.SoftwareVersion,
		&me.HardwareVersion,
		&me.CoreVersion,
		&me.Rssi,
		&me.ErrorCode,
		&me.CreateTime)
	return err
}
func (me *H03ErrorCode) DecodeFromGin(c *gin.Context) {
	if err := c.ShouldBindBodyWith(me, binding.JSON); err != nil {
		exception.Throw(common.JsonError, err.Error())
	}
}
func QueryH03ErrCodeByMac(mac string, results *[]H03ErrorCode) bool {
	filter := fmt.Sprintf("mac='%s'", mac)
	QueryDao(NewH03ErrorCode().TableName(), filter, nil, -1, func(rows *sql.Rows) {
		obj := NewH03ErrorCode()
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*results = append(*results, *obj)
		}
	})
	return true
}

/******************************************************************************
 * function:
 * description: 处理设备属性
 * param {*H03MqttMsg} mqttMsg
 * return {*}
********************************************************************************/
func handleH03Attr(mqttMsg *H03MqttMsg) {
	var mapData map[string]interface{} = mqttMsg.Data.(map[string]interface{})
	attrData := NewH03AttrData()
	// 先到redis中查询，如果没有或者查询出的ID=0,需要再到数据库中查询
	hashKey := "h03:attr"
	hashFiled := strings.ToLower(mqttMsg.Mac)
	err := redis.GetValueFromHash(hashKey, hashFiled, true, attrData)
	if err != nil || attrData.ID == 0 {
		attrDataList := make([]H03AttrData, 0)
		QueryH03AttrDataLatestByMac(mqttMsg.Mac, &attrDataList)
		if len(attrDataList) > 0 {
			attrData = &attrDataList[0]
		}
	}
	curDay := common.GetNowDate()
	// 如果当前日期大于数据库最近的日期，就把最近的日期的学习状态时长清零
	// 做到每天产生一条新记录
	if len(attrData.CreateTime) > 10 && curDay > attrData.CreateTime[:10] {
		attrData.ID = 0
		attrData.LowStudyTime = 0
		attrData.MidStudyTime = 0
		attrData.DeepStudyTime = 0
		attrData.UseLightStudyTime = 0
	}
	for key, value := range mapData {
		switch key {
		case "respiratory":
			attrData.Respiratory = int(value.(float64))
		case "heart_rate":
			attrData.HeartRate = int(value.(float64))
		case "body_movement":
			attrData.BodyMovement = int(value.(float64))
		case "body_angle":
			attrData.BodyAngle = int(value.(float64))
		case "body_distance":
			attrData.BodyDistance = int(value.(float64))
		case "onoff_status":
			attrData.OnoffStatus = int(value.(float64))
		case "control_mode":
			attrData.ControlMode = int(value.(float64))
		case "brightness_val":
			attrData.BrightnessVal = int(value.(float64))
		case "color_temp":
			attrData.ColorTemp = int(value.(float64))
		case "delay_time":
			attrData.DelayTime = int(value.(float64))
		case "gesture_mode":
			attrData.GestureMode = int(value.(float64))
		case "study_time":
			var studyTime = value.([]interface{})
			if len(studyTime) > 0 {
				attrData.LowStudyTime = int(studyTime[0].(float64))
			}
			if len(studyTime) > 1 {
				attrData.MidStudyTime = int(studyTime[1].(float64))
			}
			if len(studyTime) > 2 {
				attrData.DeepStudyTime = int(studyTime[2].(float64))
			}
			if len(studyTime) > 3 {
				attrData.UseLightStudyTime = int(studyTime[3].(float64))
			}
		case "position_interval":
			attrData.PositionInterval = int(value.(float64))
		}
	}
	attrData.Mac = mqttMsg.Mac
	attrData.CreateTime = common.GetNowTime()
	// 先更新到redis中,在没有MQ通知之前不更新到数据库，避免数据库压力，提高MQ通知的效率
	redis.SaveValueToHash(hashKey, hashFiled, nil, attrData)
	// 为了提高
	// if attrData.ID > 0 {
	// 	attrData.Update()
	// } else {
	// 	attrData.Insert()
	// }
	mq.PublishData(MakeStudyAttrTopic(attrData.Mac), attrData)

	// 数据库操作因为会出现性能延迟，所以采用队列处理
	// 队列处理
	GetTaskPool().Put(&gopool.Task{
		Params: []interface{}{attrData},
		Do: func(params ...interface{}) {
			var obj = params[0].(*H03AttrData)
			if obj.ID > 0 {
				obj.Update()
			} else {
				if obj.Insert() {
					// 如果是新插入的数据则需要再保存到redis中
					redis.SaveValueToHash(hashKey, hashFiled, nil, obj)
				}
			}
			// 队列中不需要MQ通知，因为MQ通知已经在上面的代码中处理了
			// mq.PublishData(MakeStudyAttrTopic(obj.Mac), obj)
		},
	})
}

// 定义H03设备的属性数据结构
//
//swagger:model H03AttrData
type H03AttrData struct {
	ID int64 `json:"id" mysql:"id" binding:"omitempty"`
	// 设备mac地址
	Mac string `json:"mac" mysql:"mac" size:"32" comment:"mac地址"`
	// 呼吸频率
	Respiratory int `json:"respiratory" mysql:"respiratory"`
	// 心率
	HeartRate int `json:"heart_rate" mysql:"heart_rate"`
	// 身体活跃度
	BodyMovement int `json:"body_movement" mysql:"body_movement" comment:"身体活跃度"`
	// 身体角度
	BodyAngle int `json:"body_angle" mysql:"body_angle" comment:"身体角度"`
	// 身体距离
	BodyDistance int `json:"body_distance" mysql:"body_distance" comment:"身体距离"`
	// 开关状态
	OnoffStatus int `json:"onoff_status" mysql:"onoff_status" comment:"开关状态"`
	// 控制模式
	ControlMode int `json:"control_mode" mysql:"control_mode" comment:"控制模式"`
	// 亮度值
	BrightnessVal int `json:"brightness_val" mysql:"brightness_val" comment:"亮度值"`
	// 色温
	ColorTemp int `json:"color_temp" mysql:"color_temp" comment:"色温"`
	// 延时时间
	DelayTime int `json:"delay_time" mysql:"delay_time" comment:"延时时间"`
	// 手势模式
	GestureMode int `json:"gesture_mode" mysql:"gesture_mode" comment:"手势模式，关闭，开启"`
	// 学习状态时长 轻度学习时长 单位分钟
	LowStudyTime int `json:"low_study_time" mysql:"low_study_time" comment:"学习状态时长, 轻度时长 单位分钟"`
	// 学习状态时长 中度学习时长 单位分钟
	MidStudyTime int `json:"mid_study_time" mysql:"mid_study_time" comment:"学习状态时长, 中度时长 单位分钟"`
	// 学习状态时长 深度学习时长 单位分钟
	DeepStudyTime int `json:"deep_study_time" mysql:"deep_study_time" comment:"学习状态时长, 深度时长 单位分钟"`
	// 使用灯学习时长 单位分钟
	UseLightStudyTime int `json:"use_light_study_time" mysql:"use_light_study_time" comment:"学习状态时长, 用灯时长 单位分钟"`
	// 坐姿位置间隔
	PositionInterval int `json:"position_interval" mysql:"position_interval" comment:"坐姿位置间隔"`
	// 创建时间
	CreateTime string `json:"create_time" mysql:"create_time" binding:"datetime=2006-01-02 15:04:05" comment:"创建时间"`
}

func (H03AttrData) TableName() string {
	return H03Type + "_attr_tbl"
}
func NewH03AttrData() *H03AttrData {
	return &H03AttrData{
		ID:                0,
		Mac:               "",
		Respiratory:       0,
		HeartRate:         0,
		BodyMovement:      0,
		BodyAngle:         0,
		BodyDistance:      0,
		OnoffStatus:       0,
		ControlMode:       0,
		BrightnessVal:     0,
		ColorTemp:         0,
		DelayTime:         0,
		GestureMode:       0,
		LowStudyTime:      0,
		MidStudyTime:      0,
		DeepStudyTime:     0,
		UseLightStudyTime: 0,
		PositionInterval:  0,
		CreateTime:        common.GetNowTime(),
	}
}
func (me *H03AttrData) Insert() bool {
	if !CheckTableExist(me.TableName()) {
		CreateTableWithStruct(me.TableName(), me)
	}
	return InsertDao(me.TableName(), me)
}
func (me *H03AttrData) Update() bool {
	return UpdateDaoByID(me.TableName(), me.ID, me)
}
func (me *H03AttrData) Delete() bool {
	return DeleteDaoByID(me.TableName(), me.ID)
}
func (me *H03AttrData) SetID(id int64) {
	me.ID = id
}
func (me *H03AttrData) QueryByID(id int64) bool {
	return QueryDaoByID(me.TableName(), id, me)
}
func (me *H03AttrData) DecodeFromRows(rows *sql.Rows) error {
	err := rows.Scan(
		&me.ID,
		&me.Mac,
		&me.Respiratory,
		&me.HeartRate,
		&me.BodyMovement,
		&me.BodyAngle,
		&me.BodyDistance,
		&me.OnoffStatus,
		&me.ControlMode,
		&me.BrightnessVal,
		&me.ColorTemp,
		&me.DelayTime,
		&me.GestureMode,
		&me.LowStudyTime,
		&me.MidStudyTime,
		&me.DeepStudyTime,
		&me.UseLightStudyTime,
		&me.PositionInterval,
		&me.CreateTime)
	return err
}
func (me *H03AttrData) DecodeFromRow(row *sql.Row) error {
	err := row.Scan(
		&me.ID,
		&me.Mac,
		&me.Respiratory,
		&me.HeartRate,
		&me.BodyMovement,
		&me.BodyAngle,
		&me.BodyDistance,
		&me.OnoffStatus,
		&me.ControlMode,
		&me.BrightnessVal,
		&me.ColorTemp,
		&me.DelayTime,
		&me.DelayTime,
		&me.GestureMode,
		&me.LowStudyTime,
		&me.MidStudyTime,
		&me.DeepStudyTime,
		&me.UseLightStudyTime,
		&me.PositionInterval,
		&me.CreateTime)
	return err
}
func (me *H03AttrData) DecodeFromGin(c *gin.Context) {
	if err := c.ShouldBindBodyWith(me, binding.JSON); err != nil {
		exception.Throw(common.JsonError, err.Error())
	}
}

func QueryH03AttrDataByMac(mac string, limited int, results *[]H03AttrData) bool {
	filter := fmt.Sprintf("mac='%s'", mac)
	QueryDao(NewH03AttrData().TableName(), filter, "create_time desc", limited, func(rows *sql.Rows) {
		obj := NewH03AttrData()
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*results = append(*results, *obj)
		}
	})
	return true
}

/******************************************************************************
 * function: QueryH03AttrDataByMacAndDay
 * description: 根据mac地址和日期查询设备属性
 * param {string} mac
 * param {string} startDay
 * param {string} endDay
 * param {*[]H03AttrData} results
 * return {*}
********************************************************************************/
func QueryH03AttrDataByMacAndDay(mac string, startDay string, endDay string, results *[]H03AttrData) bool {
	filter := fmt.Sprintf("mac='%s' and date(create_time)>='%s' and date(create_time)<='%s'", mac, startDay, endDay)
	QueryDao(NewH03AttrData().TableName(), filter, "create_time desc", -1, func(rows *sql.Rows) {
		obj := NewH03AttrData()
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*results = append(*results, *obj)
		}
	})
	return true
}

func QueryH03AttrDataLatestByMac(mac string, results *[]H03AttrData) bool {
	filter := fmt.Sprintf("mac='%s'", mac)
	QueryDao(NewH03AttrData().TableName(), filter, "create_time desc", 1, func(rows *sql.Rows) {
		obj := NewH03AttrData()
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*results = append(*results, *obj)
		}
	})
	return true
}

/******************************************************************************
 * function:
 * description: 处理设备事件, 保存每一条事件数据到数据库
 * param {*H03MqttMsg} mqttMsg
 * return {*}
********************************************************************************/
func handleH03Event(mqttMsg *H03MqttMsg) {
	eventData := NewH03Event()
	dataMap := mqttMsg.Data.(map[string]interface{})
	// 先到redis中查询，如果没有或者查询出的ID=0,需要再到数据库中查询
	hashKey := "h03:event"
	hashFiled := strings.ToLower(mqttMsg.Mac)
	err := redis.GetValueFromHash(hashKey, hashFiled, true, eventData)
	if err != nil || eventData.ID == 0 {
		eventList := make([]H03Event, 0)
		QueryH03LatestEventByMac(mqttMsg.Mac, &eventList)
		if len(eventList) > 0 {
			eventData = &eventList[0]
		}
	}
	curDay := common.GetNowDate()
	// 如果当前日期大于数据库最近的日期，就把最近的日期的学习状态时长清零
	// 做到每天产生一条新记录
	if len(eventData.CreateTime) > 10 && curDay > eventData.CreateTime[:10] {
		eventData.ID = 0
	}
	// warning 需要每次都更新
	eventData.WarningEvent = 0

	for key, value := range dataMap {
		switch key {
		case "body_status":
			eventData.BodyStatus = int(value.(float64))
		case "flow_state":
			eventData.FlowState = int(value.(float64))
		case "posture_state":
			eventData.PostureState = int(value.(float64))
		case "activity_freq":
			eventData.ActivityFreq = int(value.(float64))
		case "warning_event":
			eventData.WarningEvent = int(value.(float64))
		}
	}
	switch eventData.FlowState {
	case 1:
		fallthrough
	case 2:
		fallthrough
	case 3:
		eventData.FocusStatus = 1
	case 4:
		fallthrough
	case 5:
		fallthrough
	case 6:
		eventData.FocusStatus = 2
	case 7:
		fallthrough
	case 8:
		fallthrough
	case 9:
		eventData.FocusStatus = 3
	default:
		eventData.FocusStatus = 0
	}
	eventData.Mac = mqttMsg.Mac
	eventData.CreateTime = common.GetNowTime()
	// 下面代码注释，因为直接操作数据库会引起性能问题，造成MQ通知不及时
	// if eventData.ID > 0 {
	// 	eventData.Update()
	// } else {
	// 	eventData.Insert()
	// }

	// 先更新到redis中,在没有MQ通知之前不更新到数据库，避免数据库压力，提高MQ通知的效率
	redis.SaveValueToHash(hashKey, hashFiled, nil, eventData)
	// 通知event事件
	mq.PublishData(MakeStudyEventTopic(eventData.Mac), eventData)
	// 数据库处理因为会出现性能延迟，所以采用队列处理
	// 队列处理
	GetTaskPool().Put(&gopool.Task{
		Params: []interface{}{eventData},
		Do: func(params ...interface{}) {
			var obj = params[0].(*H03Event)
			if obj.ID > 0 {
				obj.Update()
			} else {
				obj.Insert()
				// 如果是新插入的数据则需要再保存到redis中
				redis.SaveValueToHash(hashKey, hashFiled, nil, obj)
			}
		},
	})
	// 向微信通知告警事件
	if eventData.WarningEvent > 0 {
		switchSettings := make([]H03ReportSwitchSetting, 0)
		QueryH03ReportSwitchSetting(eventData.Mac, &switchSettings)
		if len(switchSettings) == 0 {
			return
		}
		userDevices := make([]UserDeviceDetail, 0)
		QueryUserDeviceDetailByMac(eventData.Mac, &userDevices)
		if len(userDevices) == 0 {
			return
		}
		switchSetting := &switchSettings[0]
		needNotify := false
		switch eventData.WarningEvent {
		case 1:
			needNotify = switchSetting.SeatNotifySwitch == 1
		case 2:
			needNotify = switchSetting.ConcentrationLowNotifySwitch == 1
		case 3:
			needNotify = switchSetting.ConcentrationHighNotifySwitch == 1
		case 4:
			needNotify = switchSetting.StudyTimeoutNotifySwitch == 1
		case 5:
			needNotify = switchSetting.LeaveNotifySwitch == 1
		case 6:
			needNotify = switchSetting.PostureNotifySwitch == 1
		}
		if needNotify {
			// 如果需要通知则先进行统计，再通知公众号
			// 统计放到队列中处理
			GetTaskPool().Put(&gopool.Task{
				Params: []interface{}{eventData},
				Do: func(params ...interface{}) {
					var obj = params[0].(*H03Event)
					// 统计每天的告警事件
					StatH03WarningEventNotifyDaily(obj.Mac, obj.WarningEvent, obj.CreateTime)
					// 统计每周的告警事件
					StatH03WarningEventNotifyWeekly(obj.Mac, obj.WarningEvent, obj.CreateTime)
				},
			})
			// 通知公众号
			for _, userDevice := range userDevices {
				if H03ReportNotify != nil {
					status, _ := H03ReportNotify.NotifyWarningEventToOfficalAccount(
						userDevice.UserId,
						userDevice.NickName,
						eventData.Mac,
						eventData.CreateTime,
						eventData.WarningEvent)
					if status == common.Success {
						mylog.Log.Debugf("notify warning event to offical account success, mac: %s, event: %d", eventData.Mac, eventData.WarningEvent)
					} else {
						mylog.Log.Errorf("notify warning event to offical account failed, mac: %s, event: %d", eventData.Mac, eventData.WarningEvent)
					}
				}
			}
		}
	}

}

// 定义H03设备推送的事件数据结构
//
//swagger:model H03Event
type H03Event struct {
	ID int64 `json:"id" mysql:"id" binding:"omitempty"`
	// required: true
	// mac 号
	Mac string `json:"mac" mysql:"mac" size:"32" comment:"mac地址"`
	// required: true
	// 有人，无人
	BodyStatus int `json:"body_status" mysql:"body_status" comment:"身体状态 有人，无人"`
	// required: true
	// 0-离开; 1\2\3-浅(活动); 4\5\6-中(学习); 7\8\9-深(心流); 10-异常
	FlowState int `json:"flow_state" mysql:"flow_state" comment:"流状态"`
	// required: true
	// 学习状态,1: 轻度专注，2: 中度专注，3: 深度专注
	FocusStatus int `json:"focus_status" mysql:"focus_status" comment:"学习状态,0: 无 1: 轻度专注 2: 中度专注 3: 深度专注"`
	// required: true
	// 0:无人 1:站立 2:端坐 3:趴伏
	PostureState int `json:"posture_state" mysql:"posture_state" comment:"姿态状态"`
	// required: true
	// 0:无人 1:频繁活动 2:轻微活动
	ActivityFreq int `json:"activity_freq" mysql:"activity_freq" comment:"活动频率"`
	// required: true
	// 1:落座;2:专注度分数低; 3:专注度分数高 4:学习时长超时 5:反复离开 6:坐姿太偏
	WarningEvent int `json:"warning_event" mysql:"warning_event" comment:"告警事件"`

	CreateTime string `json:"create_time" mysql:"create_time" binding:"datetime=2006-01-02 15:04:05" comment:"创建时间"`
}

func (H03Event) TableName() string {
	return H03Type + "_event_tbl"
}

func NewH03Event() *H03Event {
	return &H03Event{
		ID:           0,
		Mac:          "",
		BodyStatus:   0,
		FlowState:    0,
		PostureState: 0,
		ActivityFreq: 0,
		WarningEvent: 0,
		CreateTime:   common.GetNowTime(),
	}
}

func (me *H03Event) Insert() bool {
	if !CheckTableExist(me.TableName()) {
		CreateTableWithStruct(me.TableName(), me)
	}
	return InsertDao(me.TableName(), me)
}
func (me *H03Event) Update() bool {
	return UpdateDaoByID(me.TableName(), me.ID, me)
}
func (me *H03Event) Delete() bool {
	return DeleteDaoByID(me.TableName(), me.ID)
}
func (me *H03Event) SetID(id int64) {
	me.ID = id
}
func (me *H03Event) QueryByID(id int64) bool {
	return QueryDaoByID(me.TableName(), id, me)
}
func (me *H03Event) DecodeFromRows(rows *sql.Rows) error {
	err := rows.Scan(
		&me.ID,
		&me.Mac,
		&me.BodyStatus,
		&me.FlowState,
		&me.FocusStatus,
		&me.PostureState,
		&me.ActivityFreq,
		&me.WarningEvent,
		&me.CreateTime)
	return err
}
func (me *H03Event) DecodeFromRow(row *sql.Row) error {
	err := row.Scan(
		&me.ID,
		&me.Mac,
		&me.BodyStatus,
		&me.FlowState,
		&me.FocusStatus,
		&me.PostureState,
		&me.ActivityFreq,
		&me.WarningEvent,
		&me.CreateTime)
	return err
}
func (me *H03Event) DecodeFromGin(c *gin.Context) {
	if err := c.ShouldBindBodyWith(me, binding.JSON); err != nil {
		exception.Throw(common.JsonError, err.Error())
	}
}

/******************************************************************************
 * function: QueryH03CurrentDayEventByMac
 * description: 根据MAC查询H03设备的事件, 查询当天的事件信息
 * param {string} mac
 * param {*[]H03Event} results
 * return {*}
********************************************************************************/
func QueryH03CurrentDayEventByMac(mac string, results *[]H03Event) bool {
	curDay := common.GetNowDate()
	filter := fmt.Sprintf("mac='%s' and date(create_time)=date('%s')", mac, curDay)
	QueryDao(NewH03Event().TableName(), filter, "create_time", -1, func(rows *sql.Rows) {
		obj := NewH03Event()
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*results = append(*results, *obj)
		}
	})
	return true
}

/******************************************************************************
 * function: QueryH03LatestEventByMac
 * description: 根据MAC查询H03设备的事件, 查询最近一条事件记录
 * param {string} mac
 * param {*[]H03Event} results
 * return {*}
********************************************************************************/
func QueryH03LatestEventByMac(mac string, results *[]H03Event) bool {
	filter := fmt.Sprintf("mac='%s'", mac)
	QueryDao(NewH03Event().TableName(), filter, "create_time desc", 1, func(rows *sql.Rows) {
		obj := NewH03Event()
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*results = append(*results, *obj)
		}
	})
	return true
}

/******************************************************************************
 * function: handleH03Report
 * description: h03设备学习报告上报的数据处理
 * param {*H03MqttMsg} mqttMsg
 * return {*}
********************************************************************************/
func handleH03Report(mqttMsg *H03MqttMsg) {
	reportJson := &H03StudyReportOrgJson{}
	reportJson.Mac = mqttMsg.Mac
	jsVal, err := json.Marshal(mqttMsg.Data)
	if err != nil {
		mylog.Log.Errorln(err)
	} else {
		reportJson.Value = string(jsVal)
		reportJson.CreateTime = common.GetNowTime()
		reportOrgList := make([]H03StudyReportOrgJson, 0)
		QueryH03StudyReportOrgJsonByMac(mqttMsg.Mac, reportJson.CreateTime, &reportOrgList)
		if len(reportOrgList) > 0 {
			reportJson.ID = reportOrgList[0].ID
			reportJson.Update()
		} else {
			reportJson.Insert()
		}
	}
	mapData := mqttMsg.Data.(map[string]interface{})
	report := NewH03StudyReport()
	for key, value := range mapData {
		switch key {
		case "report_start":
			report.ReportStart = int64(value.(float64))
		case "report_end":
			report.ReportEnd = int64(value.(float64))
		case "flow_state":
			flowState := value.([]interface{})
			for i := 0; i < len(flowState); i++ {
				report.FlowState = append(report.FlowState, int(flowState[i].(float64)))
			}
		case "evaluation":
			report.Evaluation = int(value.(float64))
		case "learning_continuity":
			report.LearningContinuity = int(value.(float64))
		case "study_efficiency":
			report.StudyEfficiency = int(value.(float64))
		case "concentration":
			report.Concentration = int(value.(float64))
		case "posture_evaluation":
			report.PostureEvaluation = int(value.(float64))
		case "seq_interval":
			report.SeqInterval = int(value.(float64))
		case "respiratory":
			resp := value.([]interface{})
			for i := 0; i < len(resp); i++ {
				report.Respiratory = append(report.Respiratory, int(resp[i].(float64)))
			}
		case "heart_rate":
			heartRate := value.([]interface{})
			for i := 0; i < len(heartRate); i++ {
				report.HeartRate = append(report.HeartRate, int(heartRate[i].(float64)))
			}
		case "posture_state":
			postureState := value.([]interface{})
			for i := 0; i < len(postureState); i++ {
				report.PostureState = append(report.PostureState, int(postureState[i].(float64)))
			}
		case "activity_freq":
			activityFreq := value.([]interface{})
			for i := 0; i < len(activityFreq); i++ {
				report.ActivityFreq = append(report.ActivityFreq, int(activityFreq[i].(float64)))
			}
		case "body_pos":
			bodyPos := value.([]interface{})
			for i := 0; i < len(bodyPos); i++ {
				report.BodyPos = append(report.BodyPos, int(bodyPos[i].(float64)))
			}
		}
	}
	report.Mac = mqttMsg.Mac
	report.StartTime = common.SecondsToTimeStr(report.ReportStart)
	report.EndTime = common.SecondsToTimeStr(report.ReportEnd)
	report.CreateTime = common.GetNowTime()
	// 放到队列中执行入库以及推送通知操作
	GetTaskPool().Put(&gopool.Task{
		Params: []interface{}{report},
		Do: func(params ...interface{}) {
			var obj = params[0].(*H03StudyReport)
			reportList := make([]H03StudyReport, 0)
			QueryH03StudyReportByTime(obj.Mac, obj.StartTime, obj.EndTime, &reportList)
			if len(reportList) > 0 {
				obj.ID = reportList[0].ID
				obj.Update()
			} else {
				obj.Insert()
				if len(obj.EndTime) >= 10 && obj.EndTime[:10] < common.GetNowDate() {
					return
				}
				// 只有新增时才会通知用户
				var switchList []H03ReportSwitchSetting
				QueryH03ReportSwitchSetting(obj.Mac, &switchList)
				if len(switchList) == 0 {
					return
				}
				switchSetting := switchList[0]
				if switchSetting.EveryTimeReportSwitch == 0 {
					return
				}
				userDevices := make([]UserDeviceDetail, 0)
				QueryUserDeviceDetailByMac(obj.Mac, &userDevices)
				if len(userDevices) == 0 {
					return
				}
				// 把报告通知给用户，通过公众号和小程序，优先公众号
				for _, userDevice := range userDevices {
					if H03ReportNotify != nil {
						status, _ := H03ReportNotify.NotifyEveryReportToOfficalAccount(userDevice.UserId, userDevice.NickName, obj.Mac, obj.StartTime, obj.EndTime)
						if status != common.Success {
							status, _ = H03ReportNotify.NotifyToMiniProgram(userDevice.UserId, userDevice.NickName, "次报告", obj.Mac, obj.Evaluation, obj.StartTime, obj.EndTime)
						}
						if status == common.Success {
							nowTm := common.GetNowTime()
							switchSetting.EveryReportLatestTime = &nowTm
							switchSetting.Update()
						}
					}
				}
			}
		},
	})
	// 最后再次放到队列中统计周报告
	GetTaskPool().Put(&gopool.Task{
		Params: []interface{}{report},
		Do: func(params ...interface{}) {
			var obj = params[0].(*H03StudyReport)
			// 先统计当天的日报告
			dailyReport := StatH03DailyReport(obj)
			// 再统计周报告
			StatH03WeekReport(dailyReport)
		},
	})
}

/******************************************************************************
 * description: 定义一个回调函数，用于通知学习报告
********************************************************************************/
var H03ReportNotify H03ReportNotifyCallback = nil

type H03ReportNotifyCallback interface {
	NotifyEveryReportToOfficalAccount(userId int64, nickName string, mac string, startTime string, endTime string) (int, string)
	NotifyToMiniProgram(userId int64, nickName string, reportType string, mac string, evaluation int, startTime string, endTime string) (int, string)
	NotifyWarningEventToOfficalAccount(userId int64, nickName string, mac string, tm string, event int) (int, string)
}

/******************************************************************************
 * function:
 * description: 定义学习报告原始数据结构，用于查询对比
 * return {*}
********************************************************************************/
// swagger:model H03StudyReportOrgJson
type H03StudyReportOrgJson struct {
	ID         int64  `json:"id" mysql:"id" binding:"omitempty"`
	Mac        string `json:"mac" mysql:"mac" size:"32" comment:"mac地址"`
	Value      string `json:"value" size:"4096" mysql:"value"`
	CreateTime string `json:"create_time" mysql:"create_time" binding:"datetime=2006-01-02 15:04:05" comment:"创建时间"`
}

func (H03StudyReportOrgJson) TableName() string {
	return H03Type + "_study_report_json_tbl"
}

func NewH03StudyReportOrgJson() *H03StudyReportOrgJson {
	return &H03StudyReportOrgJson{
		ID:         0,
		Mac:        "",
		Value:      "",
		CreateTime: common.GetNowTime(),
	}
}

func (me *H03StudyReportOrgJson) Insert() bool {
	if !CheckTableExist(me.TableName()) {
		CreateTableWithStruct(me.TableName(), me)
	}
	return InsertDao(me.TableName(), me)
}
func (me *H03StudyReportOrgJson) Update() bool {
	return UpdateDaoByID(me.TableName(), me.ID, me)
}
func (me *H03StudyReportOrgJson) Delete() bool {
	return DeleteDaoByID(me.TableName(), me.ID)
}
func (me *H03StudyReportOrgJson) SetID(id int64) {
	me.ID = id
}
func (me *H03StudyReportOrgJson) QueryByID(id int64) bool {
	return QueryDaoByID(me.TableName(), id, me)
}
func (me *H03StudyReportOrgJson) DecodeFromRows(rows *sql.Rows) error {
	err := rows.Scan(
		&me.ID,
		&me.Mac,
		&me.Value,
		&me.CreateTime)
	return err
}
func (me *H03StudyReportOrgJson) DecodeFromRow(row *sql.Row) error {
	err := row.Scan(
		&me.ID,
		&me.Mac,
		&me.Value,
		&me.CreateTime)
	return err
}
func (me *H03StudyReportOrgJson) DecodeFromGin(c *gin.Context) {
	if err := c.ShouldBindBodyWith(me, binding.JSON); err != nil {
		exception.Throw(common.JsonError, err.Error())
	}
}
func QueryH03StudyReportOrgJsonByMac(mac string, createTime string, results *[]H03StudyReportOrgJson) bool {
	filter := fmt.Sprintf("mac='%s' and create_time='%s'", mac, createTime)
	QueryDao(NewH03StudyReportOrgJson().TableName(), filter, nil, -1, func(rows *sql.Rows) {
		obj := NewH03StudyReportOrgJson()
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*results = append(*results, *obj)
		}
	})
	return true
}

// 定义解析MQ上报学习报告的结构
// swagger:model H03StudyReport
type H03StudyReport struct {
	ID        int64  `json:"id" mysql:"id" binding:"omitempty"`
	Mac       string `json:"mac" mysql:"mac" size:"32" comment:"mac地址"`
	StartTime string `json:"start_time" mysql:"start_time" binding:"datetime=2006-01-02 15:04:05" comment:"开始时间"`
	EndTime   string `json:"end_time" mysql:"end_time" binding:"datetime=2006-01-02 15:04:05" comment:"结束时间"`
	// required: true
	ReportStart        int64  `json:"report_start"`
	ReportEnd          int64  `json:"report_end"`
	FlowState          []int  `json:"flow_state" mysql:"flow_state" size:"512" comment:"流状态 数组格式: 状态,时间"`
	Evaluation         int    `json:"evaluation" mysql:"evaluation" comment:"评价分数"`
	LearningContinuity int    `json:"learning_continuity" mysql:"learning_continuity" comment:"学习效率"`
	StudyEfficiency    int    `json:"study_efficiency" mysql:"study_efficiency" comment:"学习时长"`
	Concentration      int    `json:"concentration" mysql:"concentration" comment:"专注度"`
	PostureEvaluation  int    `json:"posture_evaluation" mysql:"posture_evaluation" comment:"姿态评价"`
	SeqInterval        int    `json:"seq_interval" mysql:"seq_interval" comment:"连续间隔时间"`
	Respiratory        []int  `json:"respiratory" mysql:"respiratory" size:"512" comment:"呼吸频率"`
	HeartRate          []int  `json:"heart_rate" mysql:"heart_rate" size:"512" comment:"心率"`
	PostureState       []int  `json:"posture_state" mysql:"posture_state" size:"512" comment:"姿态状态"`
	ActivityFreq       []int  `json:"activity_freq" mysql:"activity_freq" size:"512" comment:"活动频率"`
	BodyPos            []int  `json:"body_pos" mysql:"body_pos" size:"512" comment:"身体位置"`
	CreateTime         string `json:"create_time" mysql:"create_time" binding:"datetime=2006-01-02 15:04:05" comment:"创建时间"`
}

func (H03StudyReport) TableName() string {
	return H03Type + "_study_report_tbl"
}

func NewH03StudyReport() *H03StudyReport {
	return &H03StudyReport{
		ID:                 0,
		Mac:                "",
		StartTime:          "",
		EndTime:            "",
		ReportStart:        0,
		ReportEnd:          0,
		FlowState:          make([]int, 0),
		Evaluation:         0,
		LearningContinuity: 0,
		StudyEfficiency:    0,
		Concentration:      0,
		PostureEvaluation:  0,
		SeqInterval:        0,
		Respiratory:        make([]int, 0),
		HeartRate:          make([]int, 0),
		PostureState:       make([]int, 0),
		ActivityFreq:       make([]int, 0),
		BodyPos:            make([]int, 0),
		CreateTime:         common.GetNowTime(),
	}
}

func (me *H03StudyReport) Insert() bool {
	if !CheckTableExist(me.TableName()) {
		CreateTableWithStruct(me.TableName(), me)
	}
	return InsertDao(me.TableName(), me)
}
func (me *H03StudyReport) Update() bool {
	return UpdateDaoByID(me.TableName(), me.ID, me)
}
func (me *H03StudyReport) Delete() bool {
	return DeleteDaoByID(me.TableName(), me.ID)
}
func (me *H03StudyReport) SetID(id int64) {
	me.ID = id
}
func (me *H03StudyReport) QueryByID(id int64) bool {
	return QueryDaoByID(me.TableName(), id, me)
}
func (me *H03StudyReport) DecodeFromRows(rows *sql.Rows) error {
	var flowState string
	var respiratory string
	var heartRate string
	var postureState string
	var activityFreq string
	var bodyPos string
	err := rows.Scan(
		&me.ID,
		&me.Mac,
		&me.StartTime,
		&me.EndTime,
		&flowState,
		&me.Evaluation,
		&me.LearningContinuity,
		&me.StudyEfficiency,
		&me.Concentration,
		&me.PostureEvaluation,
		&me.SeqInterval,
		&respiratory,
		&heartRate,
		&postureState,
		&activityFreq,
		&bodyPos,
		&me.CreateTime)
	for _, v := range strings.Split(flowState, ",") {
		if v != "" {
			val, _ := strconv.Atoi(v)
			me.FlowState = append(me.FlowState, val)
		}
	}
	for _, v := range strings.Split(respiratory, ",") {
		if v != "" {
			val, _ := strconv.Atoi(v)
			me.Respiratory = append(me.Respiratory, val)
		}
	}
	for _, v := range strings.Split(heartRate, ",") {
		if v != "" {
			val, _ := strconv.Atoi(v)
			me.HeartRate = append(me.HeartRate, val)
		}
	}
	for _, v := range strings.Split(postureState, ",") {
		if v != "" {
			val, _ := strconv.Atoi(v)
			me.PostureState = append(me.PostureState, val)
		}
	}
	for _, v := range strings.Split(activityFreq, ",") {
		if v != "" {
			val, _ := strconv.Atoi(v)
			me.ActivityFreq = append(me.ActivityFreq, val)
		}
	}
	for _, v := range strings.Split(bodyPos, ",") {
		if v != "" {
			val, _ := strconv.Atoi(v)
			me.BodyPos = append(me.BodyPos, val)
		}
	}

	return err
}
func (me *H03StudyReport) DecodeFromRow(row *sql.Row) error {
	var flowState string
	var respiratory string
	var heartRate string
	var postureState string
	var activityFreq string
	var bodyPos string
	err := row.Scan(
		&me.ID,
		&me.Mac,
		&me.StartTime,
		&me.EndTime,
		&flowState,
		&me.Evaluation,
		&me.LearningContinuity,
		&me.StudyEfficiency,
		&me.Concentration,
		&me.PostureEvaluation,
		&me.SeqInterval,
		&respiratory,
		&heartRate,
		&postureState,
		&activityFreq,
		&bodyPos,
		&me.CreateTime)
	for _, v := range strings.Split(flowState, ",") {
		if v != "" {
			val, _ := strconv.Atoi(v)
			me.FlowState = append(me.FlowState, val)
		}
	}
	for _, v := range strings.Split(respiratory, ",") {
		if v != "" {
			val, _ := strconv.Atoi(v)
			me.Respiratory = append(me.Respiratory, val)
		}
	}
	for _, v := range strings.Split(heartRate, ",") {
		if v != "" {
			val, _ := strconv.Atoi(v)
			me.HeartRate = append(me.HeartRate, val)
		}
	}
	for _, v := range strings.Split(postureState, ",") {
		if v != "" {
			val, _ := strconv.Atoi(v)
			me.PostureState = append(me.PostureState, val)
		}
	}
	for _, v := range strings.Split(activityFreq, ",") {
		if v != "" {
			val, _ := strconv.Atoi(v)
			me.ActivityFreq = append(me.ActivityFreq, val)
		}
	}
	for _, v := range strings.Split(bodyPos, ",") {
		if v != "" {
			val, _ := strconv.Atoi(v)
			me.BodyPos = append(me.BodyPos, val)
		}
	}
	return err
}
func (me *H03StudyReport) DecodeFromGin(c *gin.Context) {
	if err := c.ShouldBindBodyWith(me, binding.JSON); err != nil {
		exception.Throw(common.JsonError, err.Error())
	}
}

/******************************************************************************
 * function: QueryH03StudyReportByTime
 * description: 根据mac，开始日期，结束日期 查询H03设备的学习报告
 * param {string} mac
 * param {string} startDay
 * param {string} endDay
 * param {*[]H03StudyReport} results
 * return {*}
********************************************************************************/
func QueryH03StudyReportByDay(mac string, startDay string, endDay string, results *[]H03StudyReport, desc bool) bool {
	// 查询条件要以报告的结束时间判断，因为报告的开始时间和结束时间是有可能跨天的
	filter := fmt.Sprintf("mac='%s' and date(end_time) >= '%s' and date(end_time) <= '%s'", mac, startDay, endDay)
	sortStr := func() string {
		if desc {
			return "create_time desc"
		} else {
			return "create_time"
		}
	}
	mylog.Log.Info("QueryH03StudyReportByDay sort:", sortStr())
	QueryDao(NewH03StudyReport().TableName(), filter, sortStr(), -1, func(rows *sql.Rows) {
		obj := NewH03StudyReport()
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*results = append(*results, *obj)
		}
	})
	return true
}
func QueryH03StudyReportByTime(mac string, startTime string, endTime string, results *[]H03StudyReport) bool {
	filter := fmt.Sprintf("mac='%s' and start_time >= '%s' and end_time <= '%s'", mac, startTime, endTime)
	QueryDao(NewH03StudyReport().TableName(), filter, "create_time desc", -1, func(rows *sql.Rows) {
		obj := NewH03StudyReport()
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*results = append(*results, *obj)
		}
	})
	return true
}

/******************************************************************************
 * function: QueryH03DateListInReport
 * description: 查询H03设备的学习报告日期列表
 * param {*} mac
 * param {*} startTime
 * param {string} endTime
 * param {*[]string} results
 * return {*}
********************************************************************************/
func QueryH03DateListInReport(mac, startTime, endTime string, results *[]string) bool {
	filter := fmt.Sprintf("mac='%s' and date(start_time)>=date('%s') and date(end_time)<=date('%s')", mac, startTime, endTime)
	sql := "select distinct date(end_time) from " + NewH03StudyReport().TableName() + " where " + filter
	sql += " order by date(end_time)"
	rows, err := GetDB().Query(sql)
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
 * function: H03SyncRequest
 * description: 服务器向设备下发信息同步请求
 * param {string} mac
 * return {*}
********************************************************************************/
func H03SyncRequest(mac string) {
	mqMsg := NewH03MqttMsg()
	mqMsg.Cmd = H03SyncCmdRsp
	mqMsg.Mac = mac
	mqMsg.Ts = time.Now().Unix()
	type SyncRequest struct {
		Upgrade           int    `json:"upgrade"`
		RemoteBaseVersion string `json:"remoteBaseVersion"`
		BaseOtaUrl        string `json:"baseOtaUrl"`
		RemoteCoreVersion string `json:"remoteCoreVersion"`
		CoreOtaUrl        string `json:"coreOtaUrl"`
	}
	syncReq := &SyncRequest{}
	otaList := make([]H03SyncOta, 0)
	QueryH03Ota(&otaList)
	if len(otaList) > 0 {
		syncReq.BaseOtaUrl = otaList[0].BaseOtaUrl
		syncReq.RemoteBaseVersion = otaList[0].RemoteBaseVersion
		syncReq.CoreOtaUrl = otaList[0].CoreOtaUrl
		syncReq.RemoteCoreVersion = otaList[0].RemoteCoreVersion
	}
	syncReq.Upgrade = 1
	mqMsg.Data = syncReq
	mq.PublishData(MakeH03CtlTopic(mac), mqMsg)
}

/******************************************************************************
 * function: H03RebootRequest
 * description: 向设备发送重启请求
 * param {string} mac
 * param {int64} delayTm
 * return {*}
********************************************************************************/
func H03RebootRequest(mac string, delayTm int64) {
	type H03Reboot struct {
		RstDelay int64 `json:"rst_delay"`
		DemoMode int   `json:"demo_mode"`
	}
	mqMsg := NewH03MqttMsg()
	mqMsg.Cmd = H03RebootCmd
	mqMsg.Mac = mac
	mqMsg.Sn = makeH03Sn()
	mqMsg.Ts = time.Now().Unix()
	reboot := &H03Reboot{
		RstDelay: delayTm,
		DemoMode: 0,
	}
	mqMsg.Data = reboot
	mq.PublishData(MakeH03CtlTopic(mac), mqMsg)
}

/******************************************************************************
 * function: H03AttrRequest
 * description: 向设备发送属性请求
 * param {string} mac
 * param {string} attr 可以为具体的属性key 或者 all，all 表示所有属性
 * return {*}
********************************************************************************/
func H03AttrRequest(mac string, attr string) {
	mqMsg := NewH03MqttMsg()
	mqMsg.Cmd = H03AttrResp
	mqMsg.Mac = mac
	mqMsg.Sn = makeH03Sn()
	mqMsg.Ts = time.Now().Unix()
	type H03AttrReq struct {
		Attr string `json:"attr"`
	}
	attrReq := &H03AttrReq{
		Attr: attr,
	}
	mqMsg.Data = attrReq
	mq.PublishData(MakeH03CtlTopic(mac), mqMsg)
}

// 定义H03设备的设置请求结构
// swagger:model H03Setting
type H03Setting struct {
	// required: true
	// 设置开关状态
	SetOnoffStatus int `json:"set_onoff_status"`
	// 设置控制模式
	SetControlMode int `json:"set_control_mode"`
	// 设置亮度
	SetBrightnessVal int `json:"set_brightness_val"`
	//设置色温
	SetColorTemp int `json:"set_color_temp"`
	// 设置延时时间
	SetDelayTime int `json:"set_delay_time"`
	// 设置手势模式
	SetGestureMode int `json:"set_gesture_mode"`
}

/******************************************************************************
 * function: H03SettingRequest
 * description: 发送设置命令
 * param {string} mac
 * param {*H03Setting} setting
 * return {*}
********************************************************************************/
func H03SettingRequest(mac string, setting *H03Setting) {
	mqMsg := NewH03MqttMsg()
	mqMsg.Cmd = H03SettingCmd
	mqMsg.Mac = mac
	mqMsg.Sn = makeH03Sn()
	mqMsg.Ts = time.Now().Unix()
	mqMsg.Data = setting
	mq.PublishData(MakeH03FuncTopic(mac), mqMsg)
}

// swagger:model H03ReportSwitchSetting
type H03ReportSwitchSetting struct {
	ID  int64  `json:"id" mysql:"id" binding:"omitempty"`
	Mac string `json:"mac" mysql:"mac" size:"32" comment:"mac地址"`
	// required: true
	// 次报告开关 0:关闭 1:开启
	EveryTimeReportSwitch int `json:"every_time_report_switch" mysql:"every_time_report_switch" comment:"次报告开关"`
	// required: true
	// 日报告开关 0:关闭 1:开启
	DayReportSwitch int `json:"day_report_switch" mysql:"day_report_switch" comment:"日报告开关"`
	// required: true
	// 日报告推送设定时间 格式: 00:00:00
	DayReportPushSetTime string `json:"day_report_push_set_time" mysql:"day_report_push_set_time" size:"32" comment:"日报告推送设定时间"`
	// 最近次报告推送时间
	EveryReportLatestTime *string `json:"every_report_latest_time" mysql:"every_report_latest_time" isnull:"true" binding:"datetime=2006-01-02 15:04:05" comment:"最近次报告推送时间"`
	// 最近日报告推送时间
	DayReportLatestLatestTime *string `json:"day_report_latest_time" mysql:"day_report_latest_time" isnull:"true" binding:"datetime=2006-01-02 15:04:05" comment:"最近日报告推送时间"`
	// 落座通知开关 0:关闭 1:开启
	SeatNotifySwitch int `json:"seat_notify_switch" mysql:"seat_notify_switch" comment:"落座通知开关"`
	// 专注度低通知开关 0:关闭 1:开启
	ConcentrationLowNotifySwitch int `json:"concentration_low_notify_switch" mysql:"concentration_low_notify_switch" comment:"专注度低通知开关"`
	// 专注度高通知开关 0:关闭 1:开启
	ConcentrationHighNotifySwitch int `json:"concentration_high_notify_switch" mysql:"concentration_high_notify_switch" comment:"专注度高通知开关"`
	// 学习超时通知开关 0:关闭 1:开启
	StudyTimeoutNotifySwitch int `json:"study_timeout_notify_switch" mysql:"study_timeout_notify_switch" comment:"学习超时通知开关"`
	// 反复离开通知开关 0:关闭 1:开启
	LeaveNotifySwitch int `json:"leave_notify_switch" mysql:"leave_notify_switch" comment:"反复离开通知开关"`
	// 坐姿提醒通知开关 0:关闭 1:开启
	PostureNotifySwitch int `json:"posture_notify_switch" mysql:"posture_notify_switch" comment:"坐姿提醒通知开关"`
}

func (H03ReportSwitchSetting) TableName() string {
	return H03Type + "_report_switch_setting_tbl"
}

func NewH03ReportSwitchSetting() *H03ReportSwitchSetting {
	return &H03ReportSwitchSetting{
		ID:                            0,
		Mac:                           "",
		EveryTimeReportSwitch:         0,
		DayReportSwitch:               0,
		DayReportPushSetTime:          "00:00:00",
		EveryReportLatestTime:         nil,
		DayReportLatestLatestTime:     nil,
		SeatNotifySwitch:              0,
		ConcentrationLowNotifySwitch:  0,
		ConcentrationHighNotifySwitch: 0,
		StudyTimeoutNotifySwitch:      0,
		LeaveNotifySwitch:             0,
		PostureNotifySwitch:           0,
	}
}

func (me *H03ReportSwitchSetting) Insert() bool {
	if !CheckTableExist(me.TableName()) {
		CreateTableWithStruct(me.TableName(), me)
	}
	return InsertDao(me.TableName(), me)
}
func (me *H03ReportSwitchSetting) Update() bool {
	return UpdateDaoByID(me.TableName(), me.ID, me)
}
func (me *H03ReportSwitchSetting) Delete() bool {
	return DeleteDaoByID(me.TableName(), me.ID)
}
func (me *H03ReportSwitchSetting) SetID(id int64) {
	me.ID = id
}
func (me *H03ReportSwitchSetting) QueryByID(id int64) bool {
	return QueryDaoByID(me.TableName(), id, me)
}
func (me *H03ReportSwitchSetting) DecodeFromRows(rows *sql.Rows) error {
	err := rows.Scan(
		&me.ID,
		&me.Mac,
		&me.EveryTimeReportSwitch,
		&me.DayReportSwitch,
		&me.DayReportPushSetTime,
		&me.EveryReportLatestTime,
		&me.DayReportLatestLatestTime,
		&me.SeatNotifySwitch,
		&me.ConcentrationLowNotifySwitch,
		&me.ConcentrationHighNotifySwitch,
		&me.StudyTimeoutNotifySwitch,
		&me.LeaveNotifySwitch,
		&me.PostureNotifySwitch)

	return err
}
func (me *H03ReportSwitchSetting) DecodeFromRow(row *sql.Row) error {
	err := row.Scan(
		&me.ID,
		&me.Mac,
		&me.EveryTimeReportSwitch,
		&me.DayReportSwitch,
		&me.DayReportPushSetTime,
		&me.EveryReportLatestTime,
		&me.DayReportLatestLatestTime,
		&me.SeatNotifySwitch,
		&me.ConcentrationLowNotifySwitch,
		&me.ConcentrationHighNotifySwitch,
		&me.StudyTimeoutNotifySwitch,
		&me.LeaveNotifySwitch,
		&me.PostureNotifySwitch)
	return err
}
func (me *H03ReportSwitchSetting) DecodeFromGin(c *gin.Context) {
	if err := c.ShouldBindBodyWith(me, binding.JSON); err != nil {
		exception.Throw(common.JsonError, err.Error())
	}
}

/******************************************************************************
 * function: QueryH03ReportSwitchSetting
 * description:
 * param {string} mac
 * param {*[]H03ReportSwitchSetting} results
 * return {*}
********************************************************************************/
func QueryH03ReportSwitchSetting(mac string, results *[]H03ReportSwitchSetting) bool {
	filter := fmt.Sprintf("mac='%s'", mac)
	QueryDao(NewH03ReportSwitchSetting().TableName(), filter, nil, 1, func(rows *sql.Rows) {
		obj := NewH03ReportSwitchSetting()
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*results = append(*results, *obj)
		}
	})
	return true
}

/******************************************************************************
 * function: QueryH03DayReportOpenSwitchSetting
 * description:
 * param {*[]H03ReportSwitchSetting} results
 * return {*}
********************************************************************************/
func QueryH03DayReportOpenSwitchSetting(results *[]H03ReportSwitchSetting) bool {
	filter := "day_report_switch=1"
	QueryDao(NewH03ReportSwitchSetting().TableName(), filter, nil, -1, func(rows *sql.Rows) {
		obj := NewH03ReportSwitchSetting()
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*results = append(*results, *obj)
		}
	})
	return true
}

/******************************************************************************
 * struct: H03WarningEventNotifyDailyStat
 * description: 记录每日Event警告事件统计，用于统计告警次数
 * return {*}
********************************************************************************/
// swagger:model H03WarningEventNotifyDailyStat
type H03WarningEventNotifyDailyStat struct {
	ID  int64  `json:"id" mysql:"id" binding:"omitempty"`
	Mac string `json:"mac" mysql:"mac" size:"32" comment:"mac地址"`
	// required: true
	// 警告事件
	WarningEvent int `json:"warning_event" mysql:"warning_event" comment:"警告事件"`
	// required: true
	// 警告次数
	WarningNums int    `json:"warning_nums" mysql:"warning_nums" comment:"警告次数"`
	StatYear    int    `json:"stat_year" mysql:"stat_year" comment:"统计年"`
	StatWeek    int    `json:"stat_week" mysql:"stat_week" comment:"统计周"`
	NotifyDate  string `json:"notify_date" mysql:"notify_date" binding:"date=2006-01-02" comment:"通知日期"`
}

func (H03WarningEventNotifyDailyStat) TableName() string {
	return H03Type + "_warning_event_daily_stat_tbl"
}

func NewH03WarningEventNotifyDailyStat() *H03WarningEventNotifyDailyStat {
	return &H03WarningEventNotifyDailyStat{
		ID:           0,
		Mac:          "",
		WarningEvent: 0,
		WarningNums:  0,
		StatYear:     0,
		StatWeek:     0,
		NotifyDate:   common.GetNowDate(),
	}
}

func (me *H03WarningEventNotifyDailyStat) Insert() bool {
	if !CheckTableExist(me.TableName()) {
		CreateTableWithStruct(me.TableName(), me)
	}
	return InsertDao(me.TableName(), me)
}
func (me *H03WarningEventNotifyDailyStat) Update() bool {
	return UpdateDaoByID(me.TableName(), me.ID, me)
}
func (me *H03WarningEventNotifyDailyStat) Delete() bool {
	return DeleteDaoByID(me.TableName(), me.ID)
}
func (me *H03WarningEventNotifyDailyStat) SetID(id int64) {
	me.ID = id
}
func (me *H03WarningEventNotifyDailyStat) QueryByID(id int64) bool {
	return QueryDaoByID(me.TableName(), id, me)
}
func (me *H03WarningEventNotifyDailyStat) DecodeFromRows(rows *sql.Rows) error {
	err := rows.Scan(
		&me.ID,
		&me.Mac,
		&me.WarningEvent,
		&me.WarningNums,
		&me.StatYear,
		&me.StatWeek,
		&me.NotifyDate)

	return err
}
func (me *H03WarningEventNotifyDailyStat) DecodeFromRow(row *sql.Row) error {
	err := row.Scan(
		&me.ID,
		&me.Mac,
		&me.WarningEvent,
		&me.WarningNums,
		&me.StatYear,
		&me.StatWeek,
		&me.NotifyDate)
	return err
}
func (me *H03WarningEventNotifyDailyStat) DecodeFromGin(c *gin.Context) {
	if err := c.ShouldBindBodyWith(me, binding.JSON); err != nil {
		exception.Throw(common.JsonError, err.Error())
	}
}

/******************************************************************************
 * function:
 * description:
 * param {string} mac
 * param {int} warningEvent
 * param {string} notifyDate
 * param {*[]H03WarningEventNotifyDailyStat} results
 * return {*}
********************************************************************************/
func QueryH03WarningEventNotifyDailyStat(
	mac string,
	warningEvent int,
	notifyDate string,
	results *[]H03WarningEventNotifyDailyStat) bool {
	var filter string
	if warningEvent <= 0 {
		filter = fmt.Sprintf("mac='%s' and date(notify_date)=date('%s')", mac, notifyDate)
	} else {
		filter = fmt.Sprintf(
			"mac='%s' and warning_event=%d and date(notify_date)=date('%s')",
			mac,
			warningEvent,
			notifyDate,
		)
	}
	QueryDao(NewH03WarningEventNotifyDailyStat().TableName(), filter, nil, -1, func(rows *sql.Rows) {
		obj := NewH03WarningEventNotifyDailyStat()
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*results = append(*results, *obj)
		}
	})
	return true
}
func QueryH03WarningEventNotifyDailyStatByWeek(
	mac string,
	warningEvent int,
	year int,
	week int,
	results *[]H03WarningEventNotifyDailyStat) bool {
	var filter string
	if warningEvent <= 0 {
		filter = fmt.Sprintf("mac='%s' and stat_year=%d and stat_week=%d", mac, year, week)
	} else {
		filter = fmt.Sprintf(
			"mac='%s' and warning_event=%d and stat_year=%d and stat_week=%d",
			mac,
			warningEvent,
			year,
			week,
		)
	}
	QueryDao(NewH03WarningEventNotifyDailyStat().TableName(), filter, "notify_date", -1, func(rows *sql.Rows) {
		obj := NewH03WarningEventNotifyDailyStat()
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*results = append(*results, *obj)
		}
	})
	return true
}

/******************************************************************************
 * function: StatH03WarningEventNotifyDaily
 * description: 统计每日告警事件通知次数，如果已经存在则更新，否则插入
 * param {string} mac
 * param {int} warningEvent
 * param {string} notifyDate
 * return {*}
********************************************************************************/
func StatH03WarningEventNotifyDaily(mac string, warningEvent int, notifyDate string) {
	t, err := common.StrToTime(notifyDate)
	if err != nil {
		t, err = common.StrToDate(notifyDate)
		if err != nil {
			mylog.Log.Errorln(err)
			return
		}
	}
	y, w := t.ISOWeek()
	var queryResults []H03WarningEventNotifyDailyStat
	QueryH03WarningEventNotifyDailyStat(mac, warningEvent, notifyDate, &queryResults)
	stat := NewH03WarningEventNotifyDailyStat()
	if len(queryResults) > 0 {
		stat = &queryResults[0]
	}
	if stat.ID > 0 {
		stat.WarningNums++
		stat.StatYear = y
		stat.StatWeek = w
		stat.Update()
	} else {
		stat.Mac = mac
		stat.WarningEvent = warningEvent
		stat.WarningNums = 1
		stat.StatYear = y
		stat.StatWeek = w
		stat.NotifyDate = notifyDate
		stat.Insert()
	}
}

/******************************************************************************
 * struct: H03WarningEventNotifyWeekStat
 * description: 记录Event警告事件通知每周统计表
 * return {*}
********************************************************************************/
// swagger:model H03WarningEventNotifyWeekStat
type H03WarningEventNotifyWeekStat struct {
	ID  int64  `json:"id" mysql:"id" binding:"omitempty"`
	Mac string `json:"mac" mysql:"mac" size:"32" comment:"mac地址"`
	// required: true
	// 警告事件
	WarningEvent int `json:"warning_event" mysql:"warning_event" comment:"警告事件"`
	// required: true
	// 警告次数
	WarningNums int `json:"warning_nums" mysql:"warning_nums" comment:"警告次数"`
	// required: true
	// 比上周增加次数
	ThanLastWeek int `json:"than_last_week" mysql:"than_last_week" comment:"比上周增加次数"`
	StatYear     int `json:"stat_year" mysql:"stat_year" comment:"统计年"`
	StatWeek     int `json:"stat_week" mysql:"stat_week" comment:"统计周"`
}

func (H03WarningEventNotifyWeekStat) TableName() string {
	return H03Type + "_warning_event_week_stat_tbl"
}

func NewH03WarningEventNotifyWeekStat() *H03WarningEventNotifyWeekStat {
	return &H03WarningEventNotifyWeekStat{
		ID:           0,
		Mac:          "",
		WarningEvent: 0,
		WarningNums:  0,
		ThanLastWeek: 0,
		StatYear:     0,
		StatWeek:     0,
	}
}

func (me *H03WarningEventNotifyWeekStat) Insert() bool {
	if !CheckTableExist(me.TableName()) {
		CreateTableWithStruct(me.TableName(), me)
	}
	return InsertDao(me.TableName(), me)
}
func (me *H03WarningEventNotifyWeekStat) Update() bool {
	return UpdateDaoByID(me.TableName(), me.ID, me)
}
func (me *H03WarningEventNotifyWeekStat) Delete() bool {
	return DeleteDaoByID(me.TableName(), me.ID)
}
func (me *H03WarningEventNotifyWeekStat) SetID(id int64) {
	me.ID = id
}
func (me *H03WarningEventNotifyWeekStat) QueryByID(id int64) bool {
	return QueryDaoByID(me.TableName(), id, me)
}
func (me *H03WarningEventNotifyWeekStat) DecodeFromRows(rows *sql.Rows) error {
	err := rows.Scan(
		&me.ID,
		&me.Mac,
		&me.WarningEvent,
		&me.WarningNums,
		&me.ThanLastWeek,
		&me.StatYear,
		&me.StatWeek)

	return err
}
func (me *H03WarningEventNotifyWeekStat) DecodeFromRow(row *sql.Row) error {
	err := row.Scan(
		&me.ID,
		&me.Mac,
		&me.WarningEvent,
		&me.WarningNums,
		&me.ThanLastWeek,
		&me.StatYear,
		&me.StatWeek)
	return err
}
func (me *H03WarningEventNotifyWeekStat) DecodeFromGin(c *gin.Context) {
	if err := c.ShouldBindBodyWith(me, binding.JSON); err != nil {
		exception.Throw(common.JsonError, err.Error())
	}
}

/******************************************************************************
 * function:
 * description:
 * param {string} mac
 * param {int} warningEvent
 * param {string} notifyDate
 * param {*[]H03WarningEventNotifyWeekStat} results
 * return {*}
********************************************************************************/
func QueryH03WarningEventNotifyWeekStat(mac string, warningEvent int, notifyDate string, results *[]H03WarningEventNotifyWeekStat) bool {
	t, err := common.StrToDate(notifyDate)
	if err != nil {
		t, err = common.StrToTime(notifyDate)
		if err != nil {
			mylog.Log.Errorln(err)
			return false
		}
	}
	y, w := t.ISOWeek()
	return QueryH03WarningEventNotifyWeekStatByWeek(mac, warningEvent, y, w, results)
}
func QueryH03WarningEventNotifyWeekStatByWeek(mac string, warningEvent int, year, week int, results *[]H03WarningEventNotifyWeekStat) bool {
	var filter string
	if warningEvent <= 0 {
		filter = fmt.Sprintf(
			"mac='%s' and stat_year=%d and stat_week=%d",
			mac,
			year,
			week)
	} else {
		filter = fmt.Sprintf(
			"mac='%s' and warning_event=%d and stat_year=%d and stat_week=%d",
			mac,
			warningEvent,
			year,
			week,
		)
	}
	QueryDao(NewH03WarningEventNotifyWeekStat().TableName(), filter, "warning_event", -1, func(rows *sql.Rows) {
		obj := NewH03WarningEventNotifyWeekStat()
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*results = append(*results, *obj)
		}
	})
	return true
}

/******************************************************************************
 * function: StatH03WarningEventNotifyWeek
 * description: 统计每周告警事件通知次数，如果已经存在则更新，否则插入
 * param {string} mac
 * param {int} warningEvent
 * param {string} notifyDate
 * return {*}
********************************************************************************/
func StatH03WarningEventNotifyWeekly(mac string, warningEvent int, notifyDate string) {
	t, err := common.StrToDate(notifyDate)
	if err != nil {
		t, err = common.StrToTime(notifyDate)
		if err != nil {
			mylog.Log.Errorln(err)
			return
		}
	}
	y, w := t.ISOWeek()
	stat := NewH03WarningEventNotifyWeekStat()
	var queryResults []H03WarningEventNotifyWeekStat
	QueryH03WarningEventNotifyWeekStat(mac, warningEvent, notifyDate, &queryResults)
	if len(queryResults) > 0 {
		stat = &queryResults[0]
	}
	stat.Mac = mac
	stat.WarningEvent = warningEvent
	stat.WarningNums++
	stat.StatYear = y
	stat.StatWeek = w
	lastY, lastW := t.AddDate(0, 0, -7).ISOWeek()
	var lastWeekResults []H03WarningEventNotifyWeekStat
	QueryH03WarningEventNotifyWeekStatByWeek(mac, warningEvent, lastY, lastW, &lastWeekResults)
	if len(lastWeekResults) > 0 {
		stat.ThanLastWeek = stat.WarningNums - lastWeekResults[0].WarningNums
	}
	if stat.ID > 0 {
		stat.Update()
	} else {
		stat.Insert()
	}
}

/******************************************************************************
 * description: 周报统计数据结构
 * return {*}
********************************************************************************/
// swagger:model H03WeekReport
type H03WeekReport struct {
	// required: true
	// ID
	ID int64 `json:"id" mysql:"id" binding:"omitempty"`
	// required: true
	// mac地址
	Mac string `json:"mac" mysql:"mac" size:"32" comment:"mac地址"`
	// required: true
	// 总学习时长,单位分钟
	TotalStudyTime float32 `json:"total_study_time" mysql:"total_study_time" comment:"总学习时长,单位分钟"`
	// required: true
	// 比上周学习时长,单位分钟
	ThanLastStudyTime float32 `json:"than_last_study_time" mysql:"than_last_study_time" comment:"比上周学习时长"`
	// required: true
	// 一周学习天数
	StudyDayNums int `json:"study_day_nums" mysql:"study_day_nums" comment:"学习天数"`
	// required: true
	// 比上周学习天数
	ThanLastStudyDayNums int `json:"than_last_study_day_nums" mysql:"than_last_study_day_nums" comment:"比上周学习天数"`
	// required: true
	// 平均每天学习时长,单位分钟
	AvgDayStudyTime float32 `json:"avg_day_study_time" mysql:"avg_day_study_time" comment:"平均每天学习时长,单位分钟"`
	// required: true
	// 最大学习评分
	MaxStudyEvaluation float32 `json:"max_study_evaluation" mysql:"max_study_evaluation" comment:"最大学习评分"`
	// required: true
	// 平均学习评分
	AvgStudyEvaluation float32 `json:"avg_study_evaluation" mysql:"avg_study_evaluation" comment:"平均学习评分"`
	// required: true
	// 最大学习评分周几
	MaxStudyEvaluationWeekDay int `json:"max_study_evaluation_week_day" mysql:"max_study_evaluation_week_day" comment:"最大学习评分周几"`
	// required: true
	// 周几奖励金奖
	GoldAwardWeekDay int `json:"gold_award_week_day" mysql:"gold_award_week_day" comment:"周几奖励金奖"`
	// required: true
	// 周几最高专注度
	MaxConcentrationWeekDay int `json:"max_concentration_week_day" mysql:"max_concentration_week_day" comment:"周几最高专注度"`
	// required: true
	// 最高专注度分数
	MaxConcentration float32 `json:"max_concentration" mysql:"max_concentration" comment:"最高专注度分数"`
	// required: true
	// 比上周最高专注度
	ThanLastConcentration float32 `json:"than_last_concentration" mysql:"than_last_concentration" comment:"比上周最高专注度"`
	// required: true
	// 总专注度分数
	TotalConcentration float32 `json:"total_concentration" mysql:"total_concentration" comment:"总专注度分数"`
	// required: true
	// 一周平均专注度分数
	AvgConcentration float32 `json:"avg_concentration" mysql:"avg_concentration" comment:"一周平均专注度分数"`
	// required: true
	// 最长学习时长,单位分钟
	MaxStudyTime float32 `json:"max_study_time" mysql:"max_study_time" comment:"最长学习时长"`
	// required: true
	// 比上周最长学习时长
	ThanLastMaxStudyTime float32 `json:"than_last_max_study_time" mysql:"than_last_max_study_time" comment:"比上周最长学习时长"`
	// required: true
	// 最长学习时长周几
	MaxStudyTimeWeekDay int `json:"max_study_time_week_day" mysql:"max_study_time_week_day" comment:"最长学习时长周几"`
	// required: true
	// 金奖数量
	GoldAwardNums int `json:"gold_award_nums" mysql:"gold_award_nums" comment:"金奖数量"`
	// required: true
	// 银奖数量
	SliverAwardNums int `json:"sliver_award_nums" mysql:"sliver_award_nums" comment:"银奖数量"`
	// required: true
	// 铜奖数量
	BronzeAwardNums int `json:"bronze_award_nums" mysql:"bronze_award_nums" comment:"铜奖数量"`
	// required: true
	// 最近一次学习报告结束时间
	LastEndReportTime string `json:"last_end_report_time" mysql:"last_end_report_time" binding:"datetime=2006-01-02 15:04:05" comment:"最近一次学习报告结束时间"`
	// required: true
	// 报告年份
	ReportYear int `json:"report_year" mysql:"report_year" comment:"报告年份"`
	// required: true
	// 报告周数
	ReportWeek int `json:"report_week" mysql:"report_week" comment:"报告周数"`
	// required: true
	// 创建时间
	CreateTime string `json:"create_time" mysql:"create_time" binding:"datetime=2006-01-02 15:04:05" comment:"创建时间"`
}

func (H03WeekReport) TableName() string {
	return H03Type + "_week_report_tbl"
}

func NewH03WeekReport() *H03WeekReport {
	return &H03WeekReport{
		ID:                        0,
		Mac:                       "",
		TotalStudyTime:            0.0,
		ThanLastStudyTime:         0.0,
		StudyDayNums:              0,
		ThanLastStudyDayNums:      0,
		AvgDayStudyTime:           0.0,
		MaxStudyEvaluation:        0.0,
		AvgStudyEvaluation:        0.0,
		MaxStudyEvaluationWeekDay: -1,
		GoldAwardWeekDay:          -1,
		MaxConcentrationWeekDay:   -1,
		MaxConcentration:          0.0,
		ThanLastConcentration:     0.0,
		TotalConcentration:        0.0,
		AvgConcentration:          0.0,
		MaxStudyTime:              0.0,
		ThanLastMaxStudyTime:      0.0,
		MaxStudyTimeWeekDay:       -1,
		GoldAwardNums:             0,
		SliverAwardNums:           0,
		BronzeAwardNums:           0,
		LastEndReportTime:         "",
		ReportYear:                0,
		ReportWeek:                0,
		CreateTime:                common.GetNowTime(),
	}
}

func (me *H03WeekReport) Insert() bool {
	if !CheckTableExist(me.TableName()) {
		CreateTableWithStruct(me.TableName(), me)
	}
	return InsertDao(me.TableName(), me)
}
func (me *H03WeekReport) Update() bool {
	return UpdateDaoByID(me.TableName(), me.ID, me)
}
func (me *H03WeekReport) Delete() bool {
	return DeleteDaoByID(me.TableName(), me.ID)
}
func (me *H03WeekReport) SetID(id int64) {
	me.ID = id
}
func (me *H03WeekReport) QueryByID(id int64) bool {
	return QueryDaoByID(me.TableName(), id, me)
}
func (me *H03WeekReport) DecodeFromRows(rows *sql.Rows) error {
	err := rows.Scan(
		&me.ID,
		&me.Mac,
		&me.TotalStudyTime,
		&me.ThanLastStudyTime,
		&me.StudyDayNums,
		&me.ThanLastStudyDayNums,
		&me.AvgDayStudyTime,
		&me.MaxStudyEvaluation,
		&me.AvgStudyEvaluation,
		&me.MaxStudyEvaluationWeekDay,
		&me.GoldAwardWeekDay,
		&me.MaxConcentrationWeekDay,
		&me.MaxConcentration,
		&me.ThanLastConcentration,
		&me.TotalConcentration,
		&me.AvgConcentration,
		&me.MaxStudyTime,
		&me.ThanLastMaxStudyTime,
		&me.MaxStudyTimeWeekDay,
		&me.GoldAwardNums,
		&me.SliverAwardNums,
		&me.BronzeAwardNums,
		&me.LastEndReportTime,
		&me.ReportYear,
		&me.ReportWeek,
		&me.CreateTime)

	return err
}
func (me *H03WeekReport) DecodeFromRow(row *sql.Row) error {
	err := row.Scan(
		&me.ID,
		&me.Mac,
		&me.TotalStudyTime,
		&me.ThanLastStudyTime,
		&me.StudyDayNums,
		&me.ThanLastStudyDayNums,
		&me.AvgDayStudyTime,
		&me.MaxStudyEvaluation,
		&me.AvgStudyEvaluation,
		&me.MaxStudyEvaluationWeekDay,
		&me.GoldAwardWeekDay,
		&me.MaxConcentrationWeekDay,
		&me.MaxConcentration,
		&me.ThanLastConcentration,
		&me.TotalConcentration,
		&me.AvgConcentration,
		&me.MaxStudyTime,
		&me.ThanLastMaxStudyTime,
		&me.MaxStudyTimeWeekDay,
		&me.GoldAwardNums,
		&me.SliverAwardNums,
		&me.BronzeAwardNums,
		&me.LastEndReportTime,
		&me.ReportYear,
		&me.ReportWeek,
		&me.CreateTime)
	return err
}
func (me *H03WeekReport) DecodeFromGin(c *gin.Context) {
	if err := c.ShouldBindBodyWith(me, binding.JSON); err != nil {
		exception.Throw(common.JsonError, err.Error())
	}
}

/******************************************************************************
 * function: QueryH03WeekReportByMac
 * description: 根据mac以及时间查询周报
 * param {string} mac
 * param {string} currDate
 * param {*[]H03WeekReport} results
 * return {*}
********************************************************************************/
func QueryH03WeekReportByMac(mac string, currDate string, results *[]H03WeekReport) bool {
	t, err := common.StrToDate(currDate)
	if err != nil {
		t, err = common.StrToTime(currDate)
		if err != nil {
			mylog.Log.Errorln(err)
			return false
		}
	}
	y, w := t.ISOWeek()
	// 查询年份和周数相同的记录
	return QueryH03WeekReportByMacAndWeek(mac, y, w, results)
}
func QueryH03WeekReportByMacAndWeek(mac string, y, w int, results *[]H03WeekReport) bool {
	filter := fmt.Sprintf(
		"mac='%s' and report_year=%d and report_week=%d",
		mac,
		y,
		w,
	)
	QueryDao(NewH03WeekReport().TableName(), filter, nil, -1, func(rows *sql.Rows) {
		obj := NewH03WeekReport()
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*results = append(*results, *obj)
		}
	})
	return true
}

/******************************************************************************
 * function: StatH03WeekReportFromStudyReport
 * description: 根据学习报告统计周报,把统计的数据写入周报表
 * return {*}
********************************************************************************/
func StatH03WeekReport(dailyReport *H03DailyReport) {
	if dailyReport == nil {
		mylog.Log.Errorln(mylog.CurrFunName(), " dailyReport is nil")
		return
	}
	y := dailyReport.ReportYear
	w := dailyReport.ReportWeek
	// 查询一周的日报表，准备统计一周数据
	dailyReportList := make([]H03DailyReport, 0)
	QueryH03DailyReportByWeek(dailyReport.Mac, y, w, &dailyReportList)
	if len(dailyReportList) == 0 {
		dailyReportList = append(dailyReportList, *dailyReport)
	}
	weekReport := NewH03WeekReport()
	weekReportList := make([]H03WeekReport, 0)
	QueryH03WeekReportByMac(dailyReport.Mac, dailyReport.DailyDate, &weekReportList)
	if len(weekReportList) > 0 {
		weekReport = &weekReportList[0]
	}
	weekReport.ReportYear = y
	weekReport.ReportWeek = w
	weekReport.LastEndReportTime = dailyReport.LastEndReportTime
	weekReport.StudyDayNums = 0
	weekReport.TotalStudyTime = 0.0
	weekReport.MaxStudyTime = 0.0
	weekReport.MaxConcentration = 0.0
	weekReport.MaxStudyEvaluation = 0.0
	weekReport.MaxStudyEvaluationWeekDay = -1
	weekReport.TotalConcentration = 0.0
	weekReport.GoldAwardNums = 0
	weekReport.SliverAwardNums = 0
	weekReport.BronzeAwardNums = 0
	weekReport.ThanLastConcentration = 0.0
	weekReport.ThanLastMaxStudyTime = 0.0
	weekReport.ThanLastStudyDayNums = 0
	weekReport.ThanLastStudyTime = 0.0
	weekReport.GoldAwardWeekDay = -1
	weekReport.MaxConcentrationWeekDay = -1
	weekReport.MaxStudyTimeWeekDay = -1
	// 统计一周的数据
	for _, report := range dailyReportList {
		t, err := common.StrToDate(report.DailyDate)
		if err != nil {
			t, err = common.StrToTime(report.DailyDate)
			if err != nil {
				mylog.Log.Errorln(err)
				continue
			}
		}
		weekReport.StudyDayNums++
		weekReport.TotalStudyTime += report.ReportTotalTimeLen
		if report.ReportTotalTimeLen > weekReport.MaxStudyTime {
			weekReport.MaxStudyTime = report.ReportTotalTimeLen
			weekReport.MaxStudyTimeWeekDay = int(t.Weekday())
		}
		if report.AvgConcentration > weekReport.MaxConcentration {
			weekReport.MaxConcentration = report.AvgConcentration
			weekReport.MaxConcentrationWeekDay = int(t.Weekday())
		}
		weekReport.TotalConcentration += report.AvgConcentration
		if report.AvgEvaluation > weekReport.MaxStudyEvaluation {
			weekReport.MaxStudyEvaluation = report.AvgEvaluation
			weekReport.MaxStudyEvaluationWeekDay = int(t.Weekday())
		}
		if report.AvgEvaluation >= 80.0 {
			weekReport.GoldAwardNums++
		}
		if report.AvgEvaluation >= 60.0 && report.AvgEvaluation < 80.0 {
			weekReport.SliverAwardNums++
		}
		if report.AvgEvaluation < 60.0 {
			weekReport.BronzeAwardNums++
		}
	}
	if weekReport.MaxStudyEvaluation > 80.0 {
		weekReport.GoldAwardWeekDay = weekReport.MaxStudyEvaluationWeekDay
	}
	weekReport.AvgStudyEvaluation = weekReport.MaxStudyEvaluation / (float32)(weekReport.StudyDayNums)
	weekReport.AvgDayStudyTime = weekReport.TotalStudyTime / (float32)(weekReport.StudyDayNums)
	weekReport.AvgConcentration = (float32)(weekReport.TotalConcentration) / (float32)(weekReport.StudyDayNums)
	weekReport.Mac = dailyReport.Mac
	weekReport.CreateTime = common.GetNowTime()
	//查询上周的周报，进行比较
	t, err := common.StrToDate(dailyReport.DailyDate)
	if err != nil {
		t, err = common.StrToTime(dailyReport.DailyDate)
		if err != nil {
			mylog.Log.Errorln(err)
			return
		}
	}
	// 上周的年份和周数
	y, w = t.AddDate(0, 0, -7).ISOWeek()
	lastWeekReportList := make([]H03WeekReport, 0)
	QueryH03WeekReportByMacAndWeek(dailyReport.Mac, y, w, &lastWeekReportList)
	if len(lastWeekReportList) > 0 {
		lastWeekReport := &lastWeekReportList[0]
		weekReport.ThanLastStudyTime = weekReport.TotalStudyTime - lastWeekReport.TotalStudyTime
		weekReport.ThanLastStudyDayNums = weekReport.StudyDayNums - lastWeekReport.StudyDayNums
		weekReport.ThanLastConcentration = weekReport.MaxConcentration - lastWeekReport.MaxConcentration
		weekReport.ThanLastMaxStudyTime = weekReport.MaxStudyTime - lastWeekReport.MaxStudyTime
	}
	result := false
	if weekReport.ID == 0 {
		result = weekReport.Insert()
	} else {
		result = weekReport.Update()
	}
	if !result {
		mylog.Log.Errorln(mylog.CurrFunName(), " database operation failed")
	}
}

/******************************************************************************
 * function: StatH03DailyReport
 * description: 根据当次学习报告统计日报表
 * param {*H03StudyReport} studyReport
 * return {*} H03DailyReport
********************************************************************************/
func StatH03DailyReport(studyReport *H03StudyReport) *H03DailyReport {
	if studyReport == nil {
		mylog.Log.Errorln(mylog.CurrFunName(), " studyReport is nil")
		return nil
	}
	t1, err := common.StrToTime(studyReport.StartTime)
	if err != nil {
		t1, err = common.StrToDate(studyReport.StartTime)
		if err != nil {
			mylog.Log.Errorln(mylog.CurrFunName(), err)
			return nil
		}
	}
	t2, err := common.StrToTime(studyReport.EndTime)
	if err != nil {
		t2, err = common.StrToDate(studyReport.EndTime)
		if err != nil {
			mylog.Log.Errorln(mylog.CurrFunName(), err)
			return nil
		}
	}
	reportTimeLen := t2.Sub(t1).Minutes()
	y, w := t2.ISOWeek()
	lowConcentrationNum := 0
	midConcentrationNum := 0
	highConcentrationNum := 0
	for flowState := range studyReport.FlowState {
		if flowState <= 54 {
			lowConcentrationNum++
		} else if flowState > 54 && flowState <= 75 {
			midConcentrationNum++
		} else {
			highConcentrationNum++
		}
	}
	dailyReport := NewH03DailyReport()
	dailyReport.Mac = studyReport.Mac
	dailyReportList := make([]H03DailyReport, 0)
	QueryH03DailyReportByDate(studyReport.Mac, &studyReport.EndTime, &dailyReportList)
	if len(dailyReportList) > 0 {
		dailyReport = &dailyReportList[0]
	}
	dailyReport.ReportTotalTimeLen += (float32)(reportTimeLen)
	dailyReport.LowConcentrationNum += lowConcentrationNum
	dailyReport.MidConcentrationNum += midConcentrationNum
	dailyReport.HighConcentrationNum += highConcentrationNum
	dailyReport.TotalConcentration += studyReport.Concentration
	dailyReport.TotalPosture += studyReport.PostureEvaluation
	dailyReport.TotalLearningContinuity += studyReport.LearningContinuity
	dailyReport.TotalStudyTime += (float32)(studyReport.StudyEfficiency) / 60.0
	dailyReport.TotalEvaluation += studyReport.Evaluation
	dailyReport.TotalStudyNums++
	dailyReport.AvgConcentration = (float32)(dailyReport.TotalConcentration) / (float32)(dailyReport.TotalStudyNums)
	dailyReport.AvgPosture = (float32)(dailyReport.TotalPosture) / (float32)(dailyReport.TotalStudyNums)
	dailyReport.AvgLearningContinuity = (float32)(dailyReport.TotalLearningContinuity) / (float32)(dailyReport.TotalStudyNums)
	dailyReport.AvgStudyTime = dailyReport.TotalStudyTime / (float32)(dailyReport.TotalStudyNums)
	dailyReport.AvgEvaluation = (float32)(dailyReport.TotalEvaluation) / (float32)(dailyReport.TotalStudyNums)
	dailyReport.LastEndReportTime = studyReport.EndTime
	dailyReport.ReportYear = y
	dailyReport.ReportWeek = w
	dailyReport.DailyDate = studyReport.EndTime
	result := false
	if dailyReport.ID == 0 {
		result = dailyReport.Insert()
	} else {
		result = dailyReport.Update()
	}
	if !result {
		mylog.Log.Errorln(mylog.CurrFunName(), " database operation failed")
		return nil
	}
	return dailyReport
}

/******************************************************************************
 * class: H03DailyReport
 * description: 统计日报表
 * return {*}
********************************************************************************/
type H03DailyReport struct {
	ID                      int64   `json:"id" mysql:"id" binding:"omitempty"`
	Mac                     string  `json:"mac" mysql:"mac" size:"32" comment:"mac地址"`
	ReportTotalTimeLen      float32 `json:"report_total_time_len" mysql:"report_total_time_len" comment:"报告总时长"`
	LowConcentrationNum     int     `json:"low_concentration_num" mysql:"low_concentration_num" comment:"低专注度次数"`
	MidConcentrationNum     int     `json:"mid_concentration_num" mysql:"mid_concentration_num" comment:"中专注度次数"`
	HighConcentrationNum    int     `json:"high_concentration_num" mysql:"high_concentration_num" comment:"高专注度次数"`
	TotalConcentration      int     `json:"total_concentration" mysql:"total_concentration" comment:"总专注度分数"`
	AvgConcentration        float32 `json:"avg_concentration" mysql:"avg_concentration" comment:"平均专注度分数"`
	TotalPosture            int     `json:"total_posture" mysql:"total_posture" comment:"总坐姿分数"`
	AvgPosture              float32 `json:"avg_posture" mysql:"avg_posture" comment:"平均坐姿分数"`
	TotalLearningContinuity int     `json:"total_learning_continuity" mysql:"total_learning_continuity" comment:"总学习连续性分数"`
	AvgLearningContinuity   float32 `json:"avg_learning_continuity" mysql:"avg_learning_continuity" comment:"平均学习连续性分数"`
	TotalStudyTime          float32 `json:"total_study_time" mysql:"total_study_time" comment:"总学习时长"`
	AvgStudyTime            float32 `json:"avg_study_time" mysql:"avg_study_time" comment:"平均学习时长"`
	TotalEvaluation         int     `json:"total_evaluation" mysql:"total_evaluation" comment:"总评价分数"`
	AvgEvaluation           float32 `json:"avg_evaluation" mysql:"avg_evaluation" comment:"平均评价分数"`
	TotalStudyNums          int     `json:"total_study_nums" mysql:"total_study_nums" comment:"总学习次数"`
	// 最近一次学习报告结束时间
	LastEndReportTime string `json:"last_end_report_time" mysql:"last_end_report_time" binding:"datetime=2006-01-02 15:04:05" comment:"最近一次学习报告结束时间"`
	ReportYear        int    `json:"report_year" mysql:"report_year" comment:"报告年份"`
	ReportWeek        int    `json:"report_week" mysql:"report_week" comment:"报告周数"`
	DailyDate         string `json:"daily_date" mysql:"daily_date" binding:"date=2006-01-02" comment:"每日日期"`
}

func (H03DailyReport) TableName() string {
	return H03Type + "_daily_report_tbl"
}

func NewH03DailyReport() *H03DailyReport {
	return &H03DailyReport{
		ID:                      0,
		Mac:                     "",
		ReportTotalTimeLen:      0.0,
		LowConcentrationNum:     0,
		MidConcentrationNum:     0,
		HighConcentrationNum:    0,
		TotalConcentration:      0,
		AvgConcentration:        0.0,
		TotalPosture:            0,
		AvgPosture:              0.0,
		TotalLearningContinuity: 0,
		AvgLearningContinuity:   0.0,
		TotalStudyTime:          0.0,
		AvgStudyTime:            0.0,
		TotalEvaluation:         0,
		AvgEvaluation:           0.0,
		TotalStudyNums:          0,
		LastEndReportTime:       "",
		ReportYear:              0,
		ReportWeek:              0,
		DailyDate:               "",
	}
}

func (me *H03DailyReport) Insert() bool {
	if !CheckTableExist(me.TableName()) {
		CreateTableWithStruct(me.TableName(), me)
	}
	return InsertDao(me.TableName(), me)
}
func (me *H03DailyReport) Update() bool {
	return UpdateDaoByID(me.TableName(), me.ID, me)
}
func (me *H03DailyReport) Delete() bool {
	return DeleteDaoByID(me.TableName(), me.ID)
}
func (me *H03DailyReport) SetID(id int64) {
	me.ID = id
}
func (me *H03DailyReport) QueryByID(id int64) bool {
	return QueryDaoByID(me.TableName(), id, me)
}
func (me *H03DailyReport) DecodeFromRows(rows *sql.Rows) error {
	err := rows.Scan(
		&me.ID,
		&me.Mac,
		&me.ReportTotalTimeLen,
		&me.LowConcentrationNum,
		&me.MidConcentrationNum,
		&me.HighConcentrationNum,
		&me.TotalConcentration,
		&me.AvgConcentration,
		&me.TotalPosture,
		&me.AvgPosture,
		&me.TotalLearningContinuity,
		&me.AvgLearningContinuity,
		&me.TotalStudyTime,
		&me.AvgStudyTime,
		&me.TotalEvaluation,
		&me.AvgEvaluation,
		&me.TotalStudyNums,
		&me.LastEndReportTime,
		&me.ReportYear,
		&me.ReportWeek,
		&me.DailyDate,
	)
	return err
}
func (me *H03DailyReport) DecodeFromRow(row *sql.Row) error {
	err := row.Scan(
		&me.ID,
		&me.Mac,
		&me.ReportTotalTimeLen,
		&me.LowConcentrationNum,
		&me.MidConcentrationNum,
		&me.HighConcentrationNum,
		&me.TotalConcentration,
		&me.AvgConcentration,
		&me.TotalPosture,
		&me.AvgPosture,
		&me.TotalLearningContinuity,
		&me.AvgLearningContinuity,
		&me.TotalStudyTime,
		&me.AvgStudyTime,
		&me.TotalEvaluation,
		&me.AvgEvaluation,
		&me.TotalStudyNums,
		&me.LastEndReportTime,
		&me.ReportYear,
		&me.ReportWeek,
		&me.DailyDate,
	)
	return err
}
func (me *H03DailyReport) DecodeFromGin(c *gin.Context) {
	if err := c.ShouldBindBodyWith(me, binding.JSON); err != nil {
		exception.Throw(common.JsonError, err.Error())
	}
}

/******************************************************************************
 * function: QueryH03DailyReportByMac
 * description: 根据mac和日期查询日报表
 * param {string} mac
 * param {*string} dailyDate
 * param {*[]H03DailyReport} results
 * return {*}
********************************************************************************/
func QueryH03DailyReportByDate(mac string, dailyDate *string, results *[]H03DailyReport) bool {
	filter := fmt.Sprintf("mac='%s'", mac)
	if dailyDate != nil {
		filter += fmt.Sprintf(" and date(daily_date)=date('%s')", *dailyDate)
	}
	QueryDao(NewH03DailyReport().TableName(), filter, "daily_date", -1, func(rows *sql.Rows) {
		obj := NewH03DailyReport()
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*results = append(*results, *obj)
		}
	})
	return true
}

/******************************************************************************
 * function:
 * description: 根据年和周数查询日报表
 * param {string} mac
 * param {*} year
 * param {int} week
 * param {*[]H03DailyReport} results
 * return {*}
********************************************************************************/
func QueryH03DailyReportByWeek(mac string, year, week int, results *[]H03DailyReport) bool {
	filter := fmt.Sprintf("mac='%s' and report_year=%d and report_week=%d", mac, year, week)
	QueryDao(NewH03DailyReport().TableName(), filter, "daily_date", -1, func(rows *sql.Rows) {
		obj := NewH03DailyReport()
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*results = append(*results, *obj)
		}
	})
	return true
}
