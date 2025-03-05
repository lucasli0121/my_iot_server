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
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

const (
	X1sOnlineCmd    = 100
	X1sSyncCmd      = 101
	X1sSyncCmdRsp   = 201
	X1sKeepAlive    = 102
	X1sErrCode      = 103
	X1sAttr         = 104
	X1sAttrResp     = 204
	X1sEventReport  = 105
	X1sReportSubmit = 106
	X1sRebootCmd    = 203
	X1sSettingCmd   = 205
)

const (
	deviceX1sTopicPrefix = "hjy-dev/x1/"
)

func MakeX1sInfoTopic(mac string) string {
	return deviceX1sTopicPrefix + strings.ToLower(mac) + "/info/"
}

func MakeX1sAttrTopic(mac string) string {
	return deviceX1sTopicPrefix + strings.ToLower(mac) + "/attr/"
}

func MakeX1sEventTopic(mac string) string {
	return deviceX1sTopicPrefix + strings.ToLower(mac) + "/event/"
}

func MakeX1sFuncTopic(mac string) string {
	return deviceX1sTopicPrefix + strings.ToLower(mac) + "/func/"
}

func MakeX1sReportTopic(mac string) string {
	return deviceX1sTopicPrefix + strings.ToLower(mac) + "/report/"
}
func MakeX1sCtlTopic(mac string) string {
	return deviceX1sTopicPrefix + strings.ToLower(mac) + "/ctrl/"
}

func SubscribeX1sWildcardTopic() {
	msgProc := NewX1sMqttMsgProc()
	topic := MakeX1sInfoTopic("+")
	mylog.Log.Infoln("X1s SubscribeMqttTopic, topic:", topic)
	mq.SubscribeTopic(topic, msgProc)
}

func UnsubscribeX1sWildcardTopic() {
	topic := MakeX1sInfoTopic("+")
	mylog.Log.Infoln("X1s UnsubscribeMqttTopic, topic:", topic)
	mq.UnsubscribeTopic(topic)
}

func SubscribeX1sMqttTopic(mac string) {
	mylog.Log.Infoln("X1s SubscribeMqttTopic, mac:", mac)
	msgProc := NewX1sMqttMsgProc()
	mq.SubscribeTopic(MakeX1sInfoTopic(mac), msgProc)
	mq.SubscribeTopic(MakeX1sAttrTopic(mac), msgProc)
	mq.SubscribeTopic(MakeX1sEventTopic(mac), msgProc)
	mq.SubscribeTopic(MakeX1sReportTopic(mac), msgProc)
}

func UnsubscribeX1sMqttTopic(mac string) {
	mylog.Log.Infoln("X1s UnsubscribeMqttTopic, mac:", mac)
	mq.UnsubscribeTopic(MakeX1sInfoTopic(mac))
	mq.UnsubscribeTopic(MakeX1sAttrTopic(mac))
	mq.UnsubscribeTopic(MakeX1sEventTopic(mac))
	mq.UnsubscribeTopic(MakeX1sReportTopic(mac))
}

type X1sMqttMsgProc struct {
}

func NewX1sMqttMsgProc() *X1sMqttMsgProc {
	return &X1sMqttMsgProc{}
}

func (me *X1sMqttMsgProc) HandleMqttMsg(topic string, payload []byte) {
	mylog.Log.Infoln("HandleX1sMqttMsg:", topic, string(payload))

	var mqttMsg *X1sMqttMsg = NewX1sMqttMsg()
	err := json.Unmarshal([]byte(payload), &mqttMsg)
	if err != nil {
		mylog.Log.Errorln(err)
		return
	}

	switch mqttMsg.Cmd {
	case X1sOnlineCmd:
		handleX1sOnlineCmd(mqttMsg)
	case X1sSyncCmd:
		handleX1sSyncCmd(mqttMsg)
	case X1sKeepAlive:
		handleX1sKeepAlive(mqttMsg)
	case X1sErrCode:
		handleX1sErrCode(mqttMsg)
	case X1sAttr:
		handleX1sAttr(mqttMsg)
	case X1sEventReport:
		handleX1sEvent(mqttMsg)
	case X1sReportSubmit:
		handleX1sReport(mqttMsg)
	default:
	}
}

func SplitX1sMqttTopic(topic string) (string, string) {
	idx := strings.LastIndex(topic, "/")
	if idx != -1 {
		return topic[:idx+1], topic[idx+1:]
	}
	return "", ""
}

/******************************************************************************
 * function:
 * description: 定义X1s MQTT的消息结构，用于解析MQTT消息
 * return {*}
********************************************************************************/
type X1sMqttMsg struct {
	Cmd  int         `json:"cmd"`
	Sn   int         `json:"s"`
	Ts   int64       `json:"time"`
	Mac  string      `json:"id"`
	Data interface{} `json:"data"`
}

func NewX1sMqttMsg() *X1sMqttMsg {
	return &X1sMqttMsg{
		Cmd: 0,
		Sn:  0,
		Ts:  0,
		Mac: "",
	}
}

/*
****************************************************************************
  - function:
  - description: 处理设备上线的消息
  - param {*X1sMqttMsg} mqttMsg
  - return {*}
    测试字符串：{"cmd": 100, "s":  1,"time": 1,"id": "test111","data": { "deviceType": "X1spro", "rssi": -20 }}

*******************************************************************************
*/
type X1sOnlineData struct {
	DeviceType string `json:"deviceType"`
	Rssi       int    `json:"rssi"`
}

func handleX1sOnlineCmd(mqttMsg *X1sMqttMsg) {
	mylog.Log.Infoln("x1s", "handleX1sOnlineCmd:", mqttMsg.Cmd)
	exception.TryEx{
		Try: func() {
			var data X1sOnlineData
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
			X1sSyncRequest(mqttMsg.Mac, 0)
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
  - param {*X1sMqttMsg} mqttMsg
  - return {*}
  - 测试字符串：
    {"cmd": 101, "s":  2, "time": 1720669362,"id": "543204abb252","data": {  \"deviceType\": \"X1spro\",  \"softwareVersion\": \"v1.0.0\",  \"hardwareVersion\": \"24-06-13\",  \"coreVersion\": \"v2.2\"}}

*******************************************************************************
*/

func handleX1sSyncCmd(mqttMsg *X1sMqttMsg) {
	// 发送同步响应消息

	data := NewX1sVersionData()
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
			var obj = params[0].(*X1sVersionData)
			objList := make([]X1sVersionData, 0)
			QueryX1sVersionByMac(obj.Mac, &objList)
			if len(objList) > 0 {
				obj.ID = objList[0].ID
				obj.Update()
			} else {
				obj.Insert()
			}
		},
	})

	mqRespMsg := NewX1sMqttMsg()
	syncResp := NewX1sSyncOta()
	otaList := make([]X1sSyncOta, 0)
	QueryX1sOta(false, &otaList)
	if len(otaList) > 0 {
		syncResp = &otaList[0]
	}
	syncResp.Upgrade = 0
	mqRespMsg.Cmd = X1sSyncCmdRsp
	mqRespMsg.Mac = mqttMsg.Mac
	mqRespMsg.Data = syncResp
	mqRespMsg.Ts = time.Now().Unix()
	mq.PublishData(MakeX1sCtlTopic(mqttMsg.Mac), mqRespMsg)
}

// swagger:model X1sVersionData
type X1sVersionData struct {
	ID              int64  `json:"id" mysql:"id" binding:"omitempty"`
	Mac             string `json:"mac" mysql:"mac" size:"32" comment:"mac地址"`
	DeviceType      string `json:"deviceType" size:"16" mysql:"deviceType"`
	SoftwareVersion string `json:"softwareVersion" size:"16" mysql:"softwareVersion"`
	HardwareVersion string `json:"hardwareVersion" size:"16" mysql:"hardwareVersion"`
	CoreVersion     string `json:"coreVersion" size:"16" mysql:"coreVersion"`
	CreateTime      string `json:"create_time" mysql:"create_time" binding:"datetime=2006-01-02 15:04:05" comment:"创建时间"`
}

func (X1sVersionData) TableName() string {
	return X1sType + "_version_tbl"
}
func NewX1sVersionData() *X1sVersionData {
	return &X1sVersionData{
		ID:              0,
		Mac:             "",
		DeviceType:      "",
		SoftwareVersion: "",
		HardwareVersion: "",
		CoreVersion:     "",
		CreateTime:      common.GetNowTime(),
	}
}
func (me *X1sVersionData) Insert() bool {
	if !CheckTableExist(me.TableName()) {
		CreateTableWithStruct(me.TableName(), me)
	}
	return InsertDao(me.TableName(), me)
}
func (me *X1sVersionData) Update() bool {
	return UpdateDaoByID(me.TableName(), me.ID, me)
}
func (me *X1sVersionData) Delete() bool {
	return DeleteDaoByID(me.TableName(), me.ID)
}
func (me *X1sVersionData) SetID(id int64) {
	me.ID = id
}
func (me *X1sVersionData) QueryByID(id int64) bool {
	return QueryDaoByID(me.TableName(), id, me)
}
func (me *X1sVersionData) DecodeFromRows(rows *sql.Rows) error {
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
func (me *X1sVersionData) DecodeFromRow(row *sql.Row) error {
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
func (me *X1sVersionData) DecodeFromGin(c *gin.Context) {
	if err := c.ShouldBindBodyWith(me, binding.JSON); err != nil {
		exception.Throw(common.JsonError, err.Error())
	}
}
func QueryX1sVersionByMac(mac string, results *[]X1sVersionData) bool {
	filter := fmt.Sprintf("mac='%s'", mac)
	QueryDao(NewX1sVersionData().TableName(), filter, nil, -1, func(rows *sql.Rows) {
		obj := NewX1sVersionData()
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*results = append(*results, *obj)
		}
	})
	return true
}

// 定义X1s设备同步的响应消息结构
// OTA 版本信息
type X1sSyncOta struct {
	ID                int64  `json:"id" mysql:"id" binding:"omitempty"`
	Upgrade           int    `json:"upgrade" mysql:"upgrade"`
	RemoteBaseVersion string `json:"remoteBaseVersion" size:"16" mysql:"remoteBaseVersion"`
	BaseOtaUrl        string `json:"baseOtaUrl" mysql:"baseOtaUrl"`
	RemoteCoreVersion string `json:"remoteCoreVersion" size:"16" mysql:"remoteCoreVersion"`
	CoreOtaUrl        string `json:"coreOtaUrl" mysql:"coreOtaUrl"`
	WhiteList         int    `json:"whiteList" mysql:"whiteList" default:"0" comment:"白名单"`
}

func (X1sSyncOta) TableName() string {
	return X1sType + "_ota_tbl"
}
func NewX1sSyncOta() *X1sSyncOta {
	return &X1sSyncOta{
		ID:                0,
		Upgrade:           0,
		RemoteBaseVersion: "",
		BaseOtaUrl:        "",
		RemoteCoreVersion: "",
		CoreOtaUrl:        "",
		WhiteList:         0,
	}
}
func (me *X1sSyncOta) Insert() bool {
	if !CheckTableExist(me.TableName()) {
		CreateTableWithStruct(me.TableName(), me)
	}
	return InsertDao(me.TableName(), me)
}
func (me *X1sSyncOta) Update() bool {
	return UpdateDaoByID(me.TableName(), me.ID, me)
}
func (me *X1sSyncOta) Delete() bool {
	return DeleteDaoByID(me.TableName(), me.ID)
}
func (me *X1sSyncOta) SetID(id int64) {
	me.ID = id
}
func (me *X1sSyncOta) QueryByID(id int64) bool {
	return QueryDaoByID(me.TableName(), id, me)
}
func (me *X1sSyncOta) DecodeFromRows(rows *sql.Rows) error {
	err := rows.Scan(
		&me.ID,
		&me.Upgrade,
		&me.RemoteBaseVersion,
		&me.BaseOtaUrl,
		&me.RemoteCoreVersion,
		&me.CoreOtaUrl,
		&me.WhiteList)
	return err
}
func (me *X1sSyncOta) DecodeFromRow(row *sql.Row) error {
	err := row.Scan(
		&me.ID,
		&me.Upgrade,
		&me.RemoteBaseVersion,
		&me.BaseOtaUrl,
		&me.RemoteCoreVersion,
		&me.CoreOtaUrl,
		&me.WhiteList)
	return err
}
func (me *X1sSyncOta) DecodeFromGin(c *gin.Context) {
	if err := c.ShouldBindBodyWith(me, binding.JSON); err != nil {
		exception.Throw(common.JsonError, err.Error())
	}
}

/******************************************************************************
 * function: QueryX1sOta
 * description:  查询OTA信息
 * param {*[]X1sSyncOta} results
 * return {*}
********************************************************************************/
func QueryX1sOta(whiteList bool, results *[]X1sSyncOta) bool {
	filter := func() string {
		if whiteList {
			return "whiteList=1"
		} else {
			return "whiteList=0"
		}
	}()
	QueryDao(NewX1sSyncOta().TableName(), filter, nil, -1, func(rows *sql.Rows) {
		obj := NewX1sSyncOta()
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*results = append(*results, *obj)
		}
	})
	return true
}

type X1sOtaWhiteList struct {
	ID  int64  `json:"id" mysql:"id" binding:"omitempty"`
	Mac string `json:"mac" mysql:"mac"`
}

func (X1sOtaWhiteList) TableName() string {
	return X1sType + "_ota_white_list_tbl"
}
func NewX1sOtaWhiteList() *X1sOtaWhiteList {
	return &X1sOtaWhiteList{
		ID:  0,
		Mac: "",
	}
}
func (me *X1sOtaWhiteList) Insert() bool {
	if !CheckTableExist(me.TableName()) {
		CreateTableWithStruct(me.TableName(), me)
	}
	return InsertDao(me.TableName(), me)
}
func (me *X1sOtaWhiteList) Update() bool {
	return UpdateDaoByID(me.TableName(), me.ID, me)
}
func (me *X1sOtaWhiteList) Delete() bool {
	return DeleteDaoByID(me.TableName(), me.ID)
}
func (me *X1sOtaWhiteList) SetID(id int64) {
	me.ID = id
}
func (me *X1sOtaWhiteList) QueryByID(id int64) bool {
	return QueryDaoByID(me.TableName(), id, me)
}
func (me *X1sOtaWhiteList) DecodeFromRows(rows *sql.Rows) error {
	err := rows.Scan(
		&me.ID,
		&me.Mac)
	return err
}
func (me *X1sOtaWhiteList) DecodeFromRow(row *sql.Row) error {
	err := row.Scan(
		&me.ID,
		&me.Mac)
	return err
}
func (me *X1sOtaWhiteList) DecodeFromGin(c *gin.Context) {
	if err := c.ShouldBindBodyWith(me, binding.JSON); err != nil {
		exception.Throw(common.JsonError, err.Error())
	}
}

/******************************************************************************
 * function: X1sCheckMacInOtaWhiteList
 * description: 检查mac地址是否在OTA白名单中
 * param {string} mac
 * return {*}
********************************************************************************/
func X1sCheckMacInOtaWhiteList(mac string) bool {
	var results []X1sOtaWhiteList
	filter := fmt.Sprintf("mac='%s'", mac)
	QueryDao(NewX1sOtaWhiteList().TableName(), filter, nil, -1, func(rows *sql.Rows) {
		obj := NewX1sOtaWhiteList()
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			results = append(results, *obj)
		}
	})
	return len(results) > 0
}
func QueryX1sOtaWhiteList(results *[]X1sOtaWhiteList) bool {
	QueryDao(NewX1sOtaWhiteList().TableName(), "", nil, -1, func(rows *sql.Rows) {
		obj := NewX1sOtaWhiteList()
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
 * function: X1sSyncRequest
 * description: 服务器向设备下发信息同步请求，强制设备升级
 * param {string} mac
 * return {*}
********************************************************************************/
func X1sSyncRequest(mac string, update int) {
	isWhiteList := X1sCheckMacInOtaWhiteList(mac)
	mqMsg := NewX1sMqttMsg()
	mqMsg.Cmd = X1sSyncCmdRsp
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
	otaList := make([]X1sSyncOta, 0)
	QueryX1sOta(isWhiteList, &otaList)
	if len(otaList) > 0 {
		syncReq.BaseOtaUrl = otaList[0].BaseOtaUrl
		syncReq.RemoteBaseVersion = otaList[0].RemoteBaseVersion
		syncReq.CoreOtaUrl = otaList[0].CoreOtaUrl
		syncReq.RemoteCoreVersion = otaList[0].RemoteCoreVersion
	}
	syncReq.Upgrade = update
	mqMsg.Data = syncReq
	mq.PublishData(MakeX1sCtlTopic(mac), mqMsg)
}

/******************************************************************************
 * function:
 * description: 设备心跳消息处理
 * param {*X1sMqttMsg} mqttMsg
 * return {*}
 * 测试字符串：{ "cmd": 102, "s": 2, "time": 1720669362, "id": "test111" }
********************************************************************************/
func handleX1sKeepAlive(mqttMsg *X1sMqttMsg) {
	SetDeviceOnline(mqttMsg.Mac, 1, 0)
}

func handleX1sErrCode(mqttMsg *X1sMqttMsg) {
}

func handleX1sAttr(mqttMsg *X1sMqttMsg) {
}

func handleX1sEvent(mqttMsg *X1sMqttMsg) {
}

func handleX1sReport(mqttMsg *X1sMqttMsg) {
}
