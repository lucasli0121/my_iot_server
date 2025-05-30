/******************************************************************************
 * Author: liguoqiang
 * Date: 2024-04-18 19:58:59
 * LastEditors: liguoqiang
 * LastEditTime: 2024-06-02 22:15:21
 * Description:
********************************************************************************/
package mysql

import (
	"database/sql"
	"fmt"
	"hjyserver/exception"
	mylog "hjyserver/log"
	"hjyserver/mdb/common"
	"hjyserver/mq"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// swagger:model NotifySetting
type NotifySetting struct {
	ID  int64  `json:"id" mysql:"id"`
	Mac string `json:"mac" mysql:"mac"`
	// 通知类型
	// required: true
	// PeopleType         = 1
	// BreathType         = 2
	// BreathAbnormalType = 3
	// HeartRateType      = 4
	// NurseModeType      = 5
	// BeeperType         = 6
	// LightType          = 7
	// ImproveType = 8
	Type           int    `json:"type" mysql:"type"`
	Switch         int    `json:"switch" mysql:"switch"`
	IntervalTime   int    `json:"interval_time" mysql:"interval_time"`
	HighValue      int    `json:"high_value" mysql:"high_value"`
	LowValue       int    `json:"low_value" mysql:"low_value"`
	LastStatus     int    `json:"last_status" mysql:"last_status"`
	LastNotifyTime string `json:"last_notify_time" mysql:"last_notify_time"`
}

func NewNotifySetting() *NotifySetting {
	return &NotifySetting{
		ID:             0,
		Mac:            "",
		Type:           0,
		Switch:         0,
		IntervalTime:   0,
		HighValue:      0,
		LowValue:       0,
		LastStatus:     -1,
		LastNotifyTime: common.GetNowTime(),
	}
}

/******************************************************************************
 * function: DecodeFromGin
 * description: decode notify setting from gin context
 * param {*gin.Context} c
 * return {*}
********************************************************************************/
func (me *NotifySetting) DecodeFromGin(c *gin.Context) {
	err := c.ShouldBindBodyWith(me, binding.JSON)
	if err != nil {
		exception.Throw(common.ParamError, err.Error())
	}
}
func (me *NotifySetting) DecodeFromRows(rows *sql.Rows) error {
	err := rows.Scan(&me.ID, &me.Mac, &me.Type, &me.Switch, &me.IntervalTime, &me.HighValue, &me.LowValue, &me.LastStatus, &me.LastNotifyTime)
	return err
}

func (me *NotifySetting) DecodeFromRow(row *sql.Row) error {
	err := row.Scan(&me.ID, &me.Mac, &me.Type, &me.Switch, &me.IntervalTime, &me.HighValue, &me.LowValue, &me.LastStatus, &me.LastNotifyTime)
	return err

}

func (me *NotifySetting) QueryByID(id int64) bool {
	me.SetID(id)
	return QueryDaoByID(common.NotifySettingTbl, me.ID, me)
}

func (me *NotifySetting) Insert() bool {
	tblName := common.NotifySettingTbl
	if !CheckTableExist(tblName) {
		sql := `create table ` + tblName + ` (
			id MEDIUMINT NOT NULL AUTO_INCREMENT,
			mac varchar(32) not null comment 'mac地址',
			type int not null comment '通知类型',
			switch int not null comment '开关',
			interval_time int default 0 comment '间隔时间',
			high_value int default 0 comment '高值',
			low_value int default 0 comment '低值',
			last_status int default 0 comment '最后状态',
			last_notify_time datetime default now() comment '最后通知时间',
			PRIMARY KEY (id, type)
		)`
		CreateTable(sql)
	}
	mqSettingToDevice(*me)
	return InsertDao(tblName, me)
}
func (me *NotifySetting) Update() bool {
	mqSettingToDevice(*me)
	return UpdateDaoByID(common.NotifySettingTbl, me.ID, me)
}
func (me *NotifySetting) Delete() bool {
	return DeleteDaoByID(common.NotifySettingTbl, me.ID)
}
func (me *NotifySetting) SetID(id int64) {
	me.ID = id
}

/******************************************************************************
 * function: mqSettingToDevice
 * description: publish mq to device
 * param {NotifySetting} obj
 * return {*}
********************************************************************************/
func mqSettingToDevice(obj NotifySetting) {
	if obj.Mac == "" {
		return
	}
	filter := fmt.Sprintf("mac = '%s'", obj.Mac)
	var gList = []Device{}
	QueryDeviceByCond(filter, nil, nil, &gList)
	if len(gList) == 0 {
		return
	}
	deviceObj := gList[0]
	switch obj.Type {
	case common.BreathAbnormalType:
		// 呼吸异常通知
		// BreathAbnormalNotifyTask(notifySetting)
		if deviceObj.Type == X1Type {
			BreathAbnormalX1Switch(obj.Mac, obj.Switch)
			BreathAbnormalX1(obj.Mac, int(obj.IntervalTime))
		}
	case common.NurseModeType:
		// 护士模式通知
		// NurseModeNotifyTask(notifySetting)
		if deviceObj.Type == X1Type {
			NurseModeX1Switch(obj.Mac, obj.Switch)
		}
	case common.BeeperType:
		// 蜂鸣器通知
		if deviceObj.Type == X1Type {
			// BeeperNotifyTask(notifySetting)
		}
	case common.LightType:
		// 灯光通知
		// LightNotifyTask(notifySetting)
		if deviceObj.Type == X1Type {
			SleepX1Switch(obj.Mac, obj.Switch)
		}
	case common.ImproveType:
		if deviceObj.Type == X1Type {
			ImproveDisturbedX1Switch(obj.Mac, obj.Switch)
		}
	default:

	}
}

/******************************************************************************
 * function: QueryNotifySettingByType
 * description: query notify setting by type
 * param {int} notifyType
 * return {*}
********************************************************************************/
func QueryNotifySettingByType(mac string, notifyType int) (*NotifySetting, error) {
	var notifySetting NotifySetting
	filter := fmt.Sprintf("mac = '%s' and type = %d", mac, notifyType)
	if QueryFirstByCond(common.NotifySettingTbl, filter, "", &notifySetting) {
		return &notifySetting, nil
	}
	return nil, &exception.Exception{
		Code: common.RecordNotFound,
		Msg:  "notify setting not found",
	}
}

/******************************************************************************
 * function: QueryAllNotifySetting
 * description: query notify setting
 * return {*}
********************************************************************************/
func QueryAllNotifySetting(mac string, results *[]NotifySetting) bool {
	backFunc := func(rows *sql.Rows) {
		obj := NewNotifySetting()
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*results = append(*results, *obj)
		}
	}
	var filter string
	if mac != "" {
		filter = fmt.Sprintf("mac = '%s'", mac)
	}
	return QueryDao(common.NotifySettingTbl, filter, "type", -1, backFunc)
}

/******************************************************************************
 * function: QueryNotifySettingWithOpen
 * description:
 * param {*[]NotifySetting} results
 * return {*}
********************************************************************************/
func QueryNotifySettingWithOpen(results *[]NotifySetting) bool {
	filter := "switch = 1"
	return QueryDao(common.NotifySettingTbl, filter, "mac", -1, func(rows *sql.Rows) {
		obj := NewNotifySetting()
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*results = append(*results, *obj)
		}
	})
}

/******************************************************************************
 * function: StartNotifyTask
 * description: start a task to find out what notify should be sent
 * return {*}
********************************************************************************/
func NotifyTask() {
	notifySettings := make([]NotifySetting, 0)
	QueryNotifySettingWithOpen(&notifySettings)
	var lastMac string = ""
	for _, notifySetting := range notifySettings {
		var x1DataList = []X1RealDataMysql{}
		var ed713DataList = []Ed713RealDataMysql{}
		deviceType := X1Type
		if lastMac != notifySetting.Mac {
			lastMac = notifySetting.Mac
			// 查询设备信息
			var deviceList = []Device{}
			QueryDeviceByCond(fmt.Sprintf("mac = '%s'", notifySetting.Mac), nil, nil, &deviceList)
			if len(deviceList) == 0 {
				continue
			}
			deviceType = deviceList[0].Type
		}
		switch deviceType {
		case X1Type:
			QueryX1RealDataByCond(fmt.Sprintf("mac = '%s'", notifySetting.Mac), nil, "create_time desc", 1, &x1DataList)
			if len(x1DataList) > 0 {
				checkNotifyAndPublish(notifySetting, x1DataList[0].BodyStatus, x1DataList[0].RespiratoryRate, x1DataList[0].HeartRate, x1DataList[0].CreateTime)
			}
		case Ed713Type:
			QueryEd713RealDataByCond(fmt.Sprintf("mac = '%s'", notifySetting.Mac), nil, "create_time desc", 1, &ed713DataList)
			if len(ed713DataList) > 0 {
				checkNotifyAndPublish(notifySetting, ed713DataList[0].BodyStatus, ed713DataList[0].RespiratoryRate, ed713DataList[0].HeartRate, ed713DataList[0].CreateTime)
			}
		}

	}
}

/******************************************************************************
 * function:
 * description:
 * param {NotifySetting} notifyObj
 * param {int} personStatus
 * param {int} breath
 * param {int} heartRate
 * param {string} dt
 * return {*}
********************************************************************************/
func checkNotifyAndPublish(notifyObj NotifySetting, personStatus int, breath int, heartRate int, dt string) {
	realDataDt, err := common.StrToTime(dt)
	if err != nil {
		mylog.Log.Errorln(err)
		return
	}
	lastDt, err := common.StrToTime(notifyObj.LastNotifyTime)
	if err != nil {
		mylog.Log.Errorln(err)
		return
	}
	realDiffDt := time.Since(realDataDt)
	lastDiffDt := time.Since(lastDt)
	if lastDiffDt.Minutes() < float64(notifyObj.IntervalTime) ||
		realDiffDt.Minutes() > float64(notifyObj.IntervalTime) {
		return
	}
	needNotify := false
	type NotifyStatus struct {
		Type   int `json:"type"`
		Status int `json:"status"`
	}

	var notifyStatus = NotifyStatus{
		Type:   notifyObj.Type,
		Status: 0,
	}
	switch notifyObj.Type {
	case common.PeopleType:
		if personStatus == 0 {
			notifyStatus.Status = 0
		} else {
			notifyStatus.Status = 1
		}
		if notifyStatus.Status != notifyObj.LastStatus {
			notifyObj.LastNotifyTime = common.GetNowTime()
			notifyObj.LastStatus = notifyStatus.Status
			notifyObj.Update()
			needNotify = true
		}
	case common.BreathType:
		if breath > notifyObj.HighValue || breath < notifyObj.LowValue {
			if breath > notifyObj.HighValue {
				notifyStatus.Status = 1
			} else {
				notifyStatus.Status = 0
			}
			if notifyStatus.Status != notifyObj.LastStatus {
				notifyObj.LastNotifyTime = common.GetNowTime()
				notifyObj.LastStatus = notifyStatus.Status
				notifyObj.Update()
				needNotify = true
			}
		}
	case common.HeartRateType:
		if heartRate > notifyObj.HighValue || heartRate < notifyObj.LowValue {
			if heartRate > notifyObj.HighValue {
				notifyStatus.Status = 1
			} else {
				notifyStatus.Status = 0
			}
			if notifyStatus.Status != notifyObj.LastStatus {
				notifyObj.LastNotifyTime = common.GetNowTime()
				notifyObj.LastStatus = notifyStatus.Status
				notifyObj.Update()
				needNotify = true
			}
		}
	default:
		// 其他通知
	}
	if needNotify {
		mq.PublishData(common.MakeDeviceNotifyTopic(notifyObj.Mac), notifyStatus)
		//send sms
		SendSleepNotifySms(notifyObj.Mac, notifyStatus.Type, notifyStatus.Status)
	}
}
