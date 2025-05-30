/******************************************************************************
 * Author: liguoqiang
 * Date: 2024-08-27 18:47:25
 * LastEditors: liguoqiang
 * LastEditTime: 2025-01-16 14:22:53
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
	wxtools "hjyserver/wx/tools"
	"math"
	"time"

	"github.com/gin-gonic/gin"
)

const tag = "mdb_device_h03"

var dayReportTimer *time.Timer = nil

func H03MdbInit() {
	mysql.H03ReportNotify = &H03ReportNotifyProc{}
	dayReportTimer = time.NewTimer(10 * time.Minute)
	go func() {
		for {
			select {
			case <-dayReportTimer.C:
				checkDayReportTimer()
				dayReportTimer.Reset(10 * time.Minute)
			}
		}
	}()
}
func H03MdbUnini() {
	if dayReportTimer != nil {
		dayReportTimer.Stop()
	}
}

func checkDayReportTimer() {
	mylog.Log.Debugln(tag, "begin call checkDayReportTimer...")
	var switchList []mysql.H03ReportSwitchSetting
	mysql.QueryH03DayReportOpenSwitchSetting(&switchList)
	if len(switchList) == 0 {
		return
	}
	for _, setting := range switchList {
		if setting.DayReportSwitch == 0 {
			continue
		}
		mac := setting.Mac
		// 检查是否已经推送过了
		if setting.DayReportLatestLatestTime != nil {
			lastTm, err := common.StrToTime(*setting.DayReportLatestLatestTime)
			if err == nil {
				lastDay := lastTm.Format("2006-01-02")
				today := time.Now().Format("2006-01-02")
				if lastDay == today {
					continue
				}
			}
		}
		// 如果当前时间小于 设定的推送时间， 则不推送
		nowTm := time.Now().Format("15:04:05")
		if nowTm < setting.DayReportPushSetTime {
			continue
		}
		// 检查mac是否绑定了用户，如果没有绑定则不推送
		userDevices := make([]mysql.UserDeviceDetail, 0)
		mysql.QueryUserDeviceDetailByMac(mac, &userDevices)
		if len(userDevices) == 0 {
			continue
		}
		mylog.Log.Debugln(tag, "begin query study report list mac:", mac)
		// userId := userDevices[0].UserId
		// 昨天有报告则推送通知
		beginDay := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
		endDay := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
		status, resp := QueryH03StudyReportByMac(mac, beginDay, endDay, true)
		if status == common.Success {
			reportResp := resp.(H03StudyReportResp)
			if len(reportResp.ReportList) > 0 {
				endTime := reportResp.ReportList[0].EndTime
				startTime := reportResp.ReportList[len(reportResp.ReportList)-1].StartTime
				mylog.Log.Debugln(tag, "study report startTime:", startTime, "endTime:", endTime)
				type H03DayReportPush struct {
					Mac       string `json:"mac"`
					StartTime string `json:"start_time"`
					EndTime   string `json:"end_time"`
				}
				pushObj := H03DayReportPush{
					Mac:       mac,
					StartTime: startTime,
					EndTime:   endTime,
				}
				// 推送MQ
				topic := mysql.MakeStudyDayReportTopic(mac)
				mq.PublishData(topic, pushObj)
				//向所有关联用户推送微信报告卡片
				for _, userDevice := range userDevices {
					status, _ := wxtools.SendDayReportMsgToOfficalAccount(userDevice.UserId, userDevice.NickName, mac, reportResp.AvgScore, startTime, endTime)
					if status != common.Success {
						wxtools.SendReportMsgToMiniProgram(userDevice.UserId, userDevice.NickName, "日报告", mac, reportResp.AvgScore, startTime, endTime)
					}
				}
				// 更新最新时间
				nowTm := common.GetNowTime()
				setting.DayReportLatestLatestTime = &nowTm
				setting.Update()
			}
		}
	}
}

/******************************************************************************
 * description: 定义回调函数，用来处理mysql包的H03的报告通知
 * return {*}
********************************************************************************/
type H03ReportNotifyProc struct {
}

func (me *H03ReportNotifyProc) NotifyEveryReportToOfficalAccount(userId int64, nickName string, mac string, startTime string, endTime string) (int, string) {
	return wxtools.SendEveryReportMsgToOfficalAccount(userId, nickName, mac, startTime, endTime)
}
func (me *H03ReportNotifyProc) NotifyToMiniProgram(userId int64, nickName string, title string, mac string, score int, startTime string, endTime string) (int, string) {
	return wxtools.SendReportMsgToMiniProgram(userId, nickName, title, mac, score, startTime, endTime)
}

// 用户H03 告警事件通知
func (me *H03ReportNotifyProc) NotifyWarningEventToOfficalAccount(userId int64, nickName string, mac string, tm string, event int) (int, string) {
	var msg string
	switch event {
	case 1:
		msg = "使用人已落座"
	case 2:
		msg = "专注度长时间过低"
	case 3:
		msg = "专注度长时间过高"
	case 4:
		msg = "学习时长超时，建议休息"
	case 5:
		msg = "检测到使用者反复离开"
	case 6:
		msg = "长时间未正坐，请注意坐姿"
	}
	return func(userId int64, nickName string, mac string, tm string, event int, msg string) (int, string) {
		if event == 1 {
			return wxtools.SendH03DeviceOnlineMsgToOfficalAccount(userId, nickName, mac, msg, tm)
		} else {
			return wxtools.SendH03DeviceStatusWarningMsgToOfficalAccount(userId, nickName, mac, msg, tm)
		}
	}(userId, nickName, mac, tm, event, msg)
}

func AskH03SyncVersion(c *gin.Context) (int, interface{}) {
	mac := c.Query("mac")
	if mac == "" {
		return common.ParamError, "mac required!"
	}
	deviceList := make([]mysql.Device, 0)
	filter := fmt.Sprintf("mac = '%s'", mac)
	mysql.QueryDeviceByCond(filter, nil, nil, &deviceList)
	if len(deviceList) == 0 {
		return common.NoExist, "device is not exist!"
	}
	device := deviceList[0]
	if device.Type != mysql.H03Type {
		return common.TypeError, "device's type is not H03pro!"
	}
	if device.Online == 0 {
		return common.DeviceOffLine, "device is offline!"
	}

	mysql.H03SyncRequest(mac)
	return common.Success, "ok"
}

//swagger:model H03RebootReq
type H03RebootReq struct {
	Mac     string `json:"mac"`
	DelayTm int64  `json:"delay_tm"`
}

func AskH03Reboot(c *gin.Context) (int, interface{}) {
	req := H03RebootReq{}
	err := c.ShouldBindJSON(&req)
	if err != nil {
		return common.ParamError, "param error"
	}
	deviceList := make([]mysql.Device, 0)
	filter := fmt.Sprintf("mac = '%s'", req.Mac)
	mysql.QueryDeviceByCond(filter, nil, nil, &deviceList)
	if len(deviceList) == 0 {
		return common.NoExist, "device is not exist!"
	}
	device := deviceList[0]
	if device.Type != mysql.H03Type {
		return common.TypeError, "device's type is not H03pro!"
	}
	if device.Online == 0 {
		return common.DeviceOffLine, "device is offline!"
	}
	mysql.H03RebootRequest(req.Mac, req.DelayTm)
	return common.Success, "ok"
}

//swagger:model H03SettingReq
type H03SettingReq struct {
	// 设备mac地址
	Mac string `json:"mac"`
	// required: true
	// 设置开关状态
	SetOnoffStatus int `json:"set_onoff_status"`
	// required: true
	// 设置控制模式
	SetControlMode int `json:"set_control_mode"`
	// 设置亮度
	SetBrightnessVal int `json:"set_brightness_val"`
	// required: true
	// 设置色温
	SetColorTemp int `json:"set_color_temp"`
	// required: true
	// 设置延时时间
	SetDelayTime int `json:"set_delay_time"`
	// required: true
	// 设置手势模式"
	SetGestureMode int `json:"set_gesture_mode"`
}

func SetH03Param(c *gin.Context) (int, interface{}) {
	var req map[string]interface{} = make(map[string]interface{})
	err := c.ShouldBindJSON(&req)
	// req := H03SettingReq{}
	// err := c.ShouldBindJSON(&req)
	if err != nil {
		return common.ParamError, "param error"
	}
	if _, ok := req["mac"]; !ok {
		return common.ParamError, "mac required!"
	}
	mac := req["mac"].(string)
	deviceList := make([]mysql.Device, 0)
	filter := fmt.Sprintf("mac = '%s'", mac)
	mysql.QueryDeviceByCond(filter, nil, nil, &deviceList)
	if len(deviceList) == 0 {
		return common.NoExist, "device is not exist!"
	}
	device := deviceList[0]
	if device.Type != mysql.H03Type {
		return common.TypeError, "device's type is not H03pro!"
	}
	if device.Online == 0 {
		return common.DeviceOffLine, "device is offline!"
	}
	setting := &mysql.H03Setting{}
	attrDataList := make([]mysql.H03AttrData, 0)
	mysql.QueryH03AttrDataLatestByMac(mac, &attrDataList)
	if len(attrDataList) > 0 {
		setting.SetBrightnessVal = attrDataList[0].BrightnessVal
		setting.SetColorTemp = attrDataList[0].ColorTemp
		setting.SetControlMode = attrDataList[0].ControlMode
		setting.SetDelayTime = attrDataList[0].DelayTime
		setting.SetGestureMode = attrDataList[0].GestureMode
	}
	if _, ok := req["set_onoff_status"]; ok {
		setting.SetOnoffStatus = int(req["set_onoff_status"].(float64))
	}
	if _, ok := req["set_color_temp"]; ok {
		setting.SetColorTemp = int(req["set_color_temp"].(float64))
	}
	if _, ok := req["set_control_mode"]; ok {
		setting.SetControlMode = int(req["set_control_mode"].(float64))
	}
	if _, ok := req["set_brightness_val"]; ok {
		setting.SetBrightnessVal = int(req["set_brightness_val"].(float64))
	}
	if _, ok := req["set_delay_time"]; ok {
		setting.SetDelayTime = int(req["set_delay_time"].(float64))
	}
	if _, ok := req["set_gesture_mode"]; ok {
		setting.SetGestureMode = int(req["set_gesture_mode"].(float64))
	}
	mysql.H03SettingRequest(mac, setting)
	return common.Success, setting
}

/******************************************************************************
 * function: SetH03ReportSwitch
 * description: 设置开关
 * param {*gin.Context} c
 * return {*}
********************************************************************************/
// swagger:model H03SwitchSettingReq
type H03SwitchSettingReq struct {
	Mac string `json:"mac"`
	// required: true
	// 次报告开关 0:关闭 1:开启
	EveryTimeReportSwitch int `json:"every_time_report_switch"`
	// required: true
	// 日报告开关 0:关闭 1:开启
	DayReportSwitch int `json:"day_report_switch"`
	// required: true
	// 日报告推送设定时间 格式: 00:00:00
	DayReportPushSetTime string `json:"day_report_push_set_time"`
	// required: false
	// 小程序订阅消息模板ID，后台返回给小程序端
	TempIdList []string `json:"temp_id_list"`
	// 落座通知开关 0:关闭 1:开启
	SeatNotifySwitch int `json:"seat_notify_switch"`
	// 专注度低通知开关 0:关闭 1:开启
	ConcentrationLowNotifySwitch int `json:"concentration_low_notify_switch"`
	// 专注度高通知开关 0:关闭 1:开启
	ConcentrationHighNotifySwitch int `json:"concentration_high_notify_switch"`
	// 学习超时通知开关 0:关闭 1:开启
	StudyTimeoutNotifySwitch int `json:"study_timeout_notify_switch"`
	// 反复离开通知开关 0:关闭 1:开启
	LeaveNotifySwitch int `json:"leave_notify_switch"`
	// 坐姿提醒通知开关 0:关闭 1:开启
	PostureNotifySwitch int `json:"posture_notify_switch"`
}

func SetH03ReportSwitch(c *gin.Context) (int, interface{}) {
	req := H03SwitchSettingReq{}
	err := c.ShouldBindJSON(&req)
	if err != nil {
		return common.JsonError, "json format error"
	}
	deviceList := make([]mysql.Device, 0)
	filter := fmt.Sprintf("mac = '%s'", req.Mac)
	mysql.QueryDeviceByCond(filter, nil, nil, &deviceList)
	if len(deviceList) == 0 {
		return common.NoExist, "device is not exist!"
	}
	device := deviceList[0]
	if device.Type != mysql.H03Type {
		return common.TypeError, "device's type is not H03pro!"
	}
	settingObj := mysql.NewH03ReportSwitchSetting()
	settingObj.Mac = req.Mac
	settingObj.EveryTimeReportSwitch = req.EveryTimeReportSwitch
	settingObj.DayReportSwitch = req.DayReportSwitch
	settingObj.DayReportPushSetTime = req.DayReportPushSetTime
	settingObj.SeatNotifySwitch = req.SeatNotifySwitch
	settingObj.ConcentrationLowNotifySwitch = req.ConcentrationLowNotifySwitch
	settingObj.ConcentrationHighNotifySwitch = req.ConcentrationHighNotifySwitch
	settingObj.StudyTimeoutNotifySwitch = req.StudyTimeoutNotifySwitch
	settingObj.LeaveNotifySwitch = req.LeaveNotifySwitch
	settingObj.PostureNotifySwitch = req.PostureNotifySwitch
	switchList := make([]mysql.H03ReportSwitchSetting, 0)
	mysql.QueryH03ReportSwitchSetting(settingObj.Mac, &switchList)
	result := false
	if len(switchList) == 0 {
		result = settingObj.Insert()
	} else {
		settingObj.ID = switchList[0].ID
		settingObj.EveryReportLatestTime = switchList[0].EveryReportLatestTime
		settingObj.DayReportLatestLatestTime = switchList[0].DayReportLatestLatestTime
		result = settingObj.Update()
	}
	if !result {
		return common.DBError, "db error"
	}
	req.TempIdList = make([]string, 0)
	req.TempIdList = append(req.TempIdList, cfg.This.Wx.ReportMiniTemplateId)
	return common.Success, req
}

func QueryH03ReportSwitch(c *gin.Context) (int, interface{}) {
	mac := c.Query("mac")
	if mac == "" {
		return common.ParamError, "mac required!"
	}
	deviceList := make([]mysql.Device, 0)
	filter := fmt.Sprintf("mac = '%s'", mac)
	mysql.QueryDeviceByCond(filter, nil, nil, &deviceList)
	if len(deviceList) == 0 {
		return common.NoExist, "device is not exist!"
	}
	device := deviceList[0]
	if device.Type != mysql.H03Type {
		return common.TypeError, "device's type is not H03pro!"
	}
	switchList := make([]mysql.H03ReportSwitchSetting, 0)
	mysql.QueryH03ReportSwitchSetting(mac, &switchList)
	if len(switchList) == 0 {
		return common.NoData, "no data"
	}
	return common.Success, switchList[0]
}

/******************************************************************************
 * function: QueryH03Version
 * description: 查询H03版本
 * param {*gin.Context} c
 * return {*}
********************************************************************************/
//swagger:model H03VersionResp
type H03VersionResp struct {
	mysql.H03VersionData
	IsUpdate bool `json:"is_update"`
}

func QueryH03Version(c *gin.Context) (int, interface{}) {
	mac := c.Query("mac")
	if mac == "" {
		return common.ParamError, "mac required!"
	}
	dataList := make([]mysql.H03VersionData, 0)
	mysql.QueryH03VersionByMac(mac, &dataList)
	if len(dataList) == 0 {
		return common.NoData, "no data"
	}
	resp := &H03VersionResp{
		H03VersionData: dataList[0],
		IsUpdate:       false,
	}
	otaList := make([]mysql.H03SyncOta, 0)
	mysql.QueryH03Ota(&otaList)
	if len(otaList) > 0 {
		ota := otaList[0]
		if ota.RemoteCoreVersion != dataList[0].CoreVersion || ota.RemoteBaseVersion != dataList[0].SoftwareVersion {
			resp.IsUpdate = true
		}
	}
	return common.Success, resp
}

/******************************************************************************
 * function: QueryH03LatestAttrs
 * description: 查询设备最新的一条属性
 * param {*gin.Context} c
 * return {*}
********************************************************************************/
func QueryH03LatestAttrs(c *gin.Context) (int, interface{}) {
	mac := c.Query("mac")
	if mac == "" {
		return common.ParamError, "mac required!"
	}
	deviceList := make([]mysql.Device, 0)
	filter := fmt.Sprintf("mac = '%s'", mac)
	mysql.QueryDeviceByCond(filter, nil, nil, &deviceList)
	if len(deviceList) == 0 {
		return common.NoExist, "device is not exist!"
	}
	device := deviceList[0]
	if device.Type != mysql.H03Type {
		return common.TypeError, "device's type is not H03pro!"
	}
	if device.Online == 0 {
		return common.DeviceOffLine, "device is offline!"
	}
	dataList := make([]mysql.H03AttrData, 0)
	curDay := common.GetNowDate()
	mysql.QueryH03AttrDataByMacAndDay(mac, curDay, curDay, &dataList)
	if len(dataList) == 0 {
		return common.NoData, "no data"
	}
	return common.Success, dataList[0]
}

/******************************************************************************
 * function: QueryH03LatestStudyStatus
 * description: 查询H03最近的一条学习状态
 * param {*gin.Context} c
 * return {*}
********************************************************************************/
func QueryH03LatestStudyStatus(c *gin.Context) (int, interface{}) {
	mac := c.Query("mac")
	if mac == "" {
		return common.ParamError, "mac required!"
	}
	deviceList := make([]mysql.Device, 0)
	filter := fmt.Sprintf("mac = '%s'", mac)
	mysql.QueryDeviceByCond(filter, nil, nil, &deviceList)
	if len(deviceList) == 0 {
		return common.NoExist, "device is not exist!"
	}
	device := deviceList[0]
	if device.Type != mysql.H03Type {
		return common.TypeError, "device's type is not H03pro!"
	}
	if device.Online == 0 {
		return common.DeviceOffLine, "device is offline!"
	}
	dataList := make([]mysql.H03Event, 0)
	mysql.QueryH03LatestEventByMac(mac, &dataList)
	if len(dataList) == 0 {
		return common.NoData, "no data"
	}
	return common.Success, dataList[0]
}

/******************************************************************************
 * function: QueryH03CurDayFocusStatus
 * description: 查询H03当天的专注状态
 * param {*gin.Context} c
 * return {*}
********************************************************************************/

//swagger:model H03CurDayFocusStatus
type H03CurDayFocusStatus struct {
	// 设备mac地址
	Mac string `json:"mac"`
	// 当天学习的总时间，单位分钟
	TotalTime int `json:"total_time"`
	// 学习状态时长, 低度时长 单位分钟
	LowStudyTime int `json:"low_study_time"`
	// 学习状态时长, 中度时长 单位分钟
	MidStudyTime int `json:"mid_study_time"`
	// 学习状态时长 深度学习时长 单位分钟
	DeepStudyTime int `json:"deep_study_time"`
}

func QueryH03CurDayFocusStatus(c *gin.Context) (int, interface{}) {
	mac := c.Query("mac")
	if mac == "" {
		return common.ParamError, "mac required!"
	}
	beginDay := common.GetNowDate()
	endDay := common.GetNowDate()
	attrList := make([]mysql.H03AttrData, 0)
	mysql.QueryH03AttrDataByMacAndDay(mac, beginDay, endDay, &attrList)
	status := H03CurDayFocusStatus{
		Mac:           mac,
		TotalTime:     0,
		LowStudyTime:  0,
		MidStudyTime:  0,
		DeepStudyTime: 0,
	}
	for _, attr := range attrList {
		status.LowStudyTime += attr.LowStudyTime
		status.MidStudyTime += attr.MidStudyTime
		status.DeepStudyTime += attr.DeepStudyTime
	}
	status.TotalTime = status.LowStudyTime + status.MidStudyTime + status.DeepStudyTime
	return common.Success, status
}

/******************************************************************************
 * function: QueryH03CurrentDayStudyTimeDetail
 * description: Query current day study time length by mac
 * param {*gin.Context} c
 * return {*}
********************************************************************************/
//swagger:model H03CurrentDayStudyTimeDetail
type H03CurrentDayStudyTimeDetail struct {
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	// 学习时长，单位分钟
	StudyTimeLen float64 `json:"study_time_len"`
}

func QueryH03CurrentDayStudyTimeDetail(c *gin.Context) (int, interface{}) {
	mac := c.Query("mac")
	if mac == "" {
		return common.ParamError, "mac required!"
	}
	eventList := make([]mysql.H03Event, 0)
	mysql.QueryH03CurrentDayEventByMac(mac, &eventList)
	if len(eventList) == 0 {
		return common.NoData, "no data"
	}
	studyDetail := make([]H03CurrentDayStudyTimeDetail, 0)
	beginStudy := false
	for _, event := range eventList {
		switch event.FlowState {
		case 4:
			fallthrough
		case 5:
			fallthrough
		case 6:
			if !beginStudy {
				beginStudy = true
				item := H03CurrentDayStudyTimeDetail{
					StartTime:    event.CreateTime,
					EndTime:      "",
					StudyTimeLen: 0,
				}
				studyDetail = append(studyDetail, item)
			}
		default:
			if beginStudy {
				beginStudy = false
				studyDetail[len(studyDetail)-1].EndTime = event.CreateTime
				t1, err := common.StrToTime(studyDetail[len(studyDetail)-1].StartTime)
				if err != nil {
					continue
				}
				t2, err := common.StrToTime(studyDetail[len(studyDetail)-1].EndTime)
				if err != nil {
					continue
				}
				studyDetail[len(studyDetail)-1].StudyTimeLen = t2.Sub(t1).Minutes()
			}
		}
	}
	return common.Success, studyDetail
}

func QueryH03CurrentDayStudyTimeDetail2(c *gin.Context) (int, interface{}) {
	mac := c.Query("mac")
	if mac == "" {
		return common.ParamError, "mac required!"
	}
	today := common.GetNowDate()
	reportList := make([]mysql.H03StudyReport, 0)
	mysql.QueryH03StudyReportByDay(mac, today, today, &reportList, false)
	if len(reportList) == 0 {
		return common.NoData, "no data"
	}
	studyDetail := make([]H03CurrentDayStudyTimeDetail, 0)
	for _, report := range reportList {
		e, err := common.StrToTime(report.EndTime)
		if err != nil {
			continue
		}
		s, err := common.StrToTime(report.StartTime)
		if err != nil {
			continue
		}
		study := H03CurrentDayStudyTimeDetail{
			StartTime:    report.StartTime,
			EndTime:      report.EndTime,
			StudyTimeLen: e.Sub(s).Minutes(),
		}
		studyDetail = append(studyDetail, study)
	}
	return common.Success, studyDetail
}

/******************************************************************************
 * function: QueryH03StudyReport
 * description: 查询学习报告
 * param {*gin.Context} c
 * return {*}
********************************************************************************/

//swagger:model H03StudyReportResp
type H03StudyReportResp struct {
	Mac string `json:"mac"`
	// 总学习时间
	TotalTime int `json:"total_time"`
	// 打败用户数
	BeatUsers int `json:"beat_users"`
	// 平均专注度
	AvgFocus int `json:"avg_focus"`
	// 平均坐姿管理
	AvgPosture int `json:"avg_posture"`
	// 平均学习效率
	AvgLearning int `json:"avg_learning"`
	// 平均学习时长
	AvgStudyEff int `json:"avg_study_eff"`
	// 学习评分（平均值）
	AvgScore int `json:"avg_score"`
	// 学习记录列表
	ReportList []H03ReportItem `json:"report_list"`
}

//swagger:model H03ReportItem
type H03ReportItem struct {
	// 学习开始时间
	StartTime string `json:"start_time"`
	// 学习结束时间
	EndTime string `json:"end_time"`
	// 专注度
	Focus int `json:"focus"`
	// 学习效率
	LearningContinuity int `json:"learning_continuity"`
	// 坐姿管理
	Posture int `json:"posture"`
	// 学习评分
	Score int `json:"score"`
	// 学习时长
	StudyEff int `json:"study_eff"`
	// 状态详情列表
	DetailList []H03ReportDetail `json:"detail_list"`
}

//swagger:model H03ReportDetail
type H03ReportDetail struct {
	// 状态 0:无人 1:低度 2:深度 3:离开 4:中度
	Status int `json:"status"`
	// 状态开始时间
	StartTime string `json:"start_time"`
}

func QueryH03StudyReport(c *gin.Context) (int, interface{}) {
	mac := c.Query("mac")
	if mac == "" {
		return common.ParamError, "mac required!"
	}
	startDay := c.Query("start_day")
	endDay := c.Query("end_day")
	if startDay == "" || endDay == "" {
		return common.ParamError, "start_day or end_day required!"
	}
	return QueryH03StudyReportByMac(mac, startDay, endDay, true)
}

func QueryH03StudyReportByTime(c *gin.Context) (int, interface{}) {
	mac := c.Query("mac")
	if mac == "" {
		return common.ParamError, "mac required!"
	}
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")
	if startTime == "" || endTime == "" {
		return common.ParamError, "start_time or end_time required!"
	}
	return QueryH03StudyReportByMac(mac, startTime, endTime, false)
}

func QueryH03StudyReportByMac(mac string, startDay string, endDay string, queryDay bool) (int, interface{}) {
	reportResp := H03StudyReportResp{
		Mac:         mac,
		AvgFocus:    0,
		AvgPosture:  0,
		AvgLearning: 0,
		AvgStudyEff: 0,
		AvgScore:    0,
		ReportList:  make([]H03ReportItem, 0),
	}
	reportList := make([]mysql.H03StudyReport, 0)
	if queryDay {
		// 查询日报表
		mysql.QueryH03StudyReportByDay(mac, startDay, endDay, &reportList, true)
	} else {
		// 查询时间段的日报表
		mysql.QueryH03StudyReportByTime(mac, startDay, endDay, &reportList)
	}
	if len(reportList) == 0 {
		return common.NoData, "no data"
	}
	totalTime := float64(0)
	for _, report := range reportList {
		t1, err := common.StrToTime(report.StartTime)
		if err != nil {
			continue
		}
		t2, err := common.StrToTime(report.EndTime)
		if err != nil {
			continue
		}
		totalTime += t2.Sub(t1).Minutes()

		item := H03ReportItem{
			StartTime:          report.StartTime,
			EndTime:            report.EndTime,
			Focus:              report.Concentration,
			LearningContinuity: report.LearningContinuity,
			StudyEff:           report.StudyEfficiency,
			Posture:            report.PostureEvaluation,
			Score:              report.Evaluation,
		}
		detailList := make([]H03ReportDetail, 0)
		tFlow := t1
		for i := 0; i < len(report.FlowState); i++ {
			detailItem := H03ReportDetail{
				Status:    report.FlowState[i],
				StartTime: tFlow.Format(cfg.TmFmtStr),
			}
			detailList = append(detailList, detailItem)
			// 默认增加seq_interval
			tFlow = tFlow.Add(time.Duration(report.SeqInterval) * time.Second)
		}
		item.DetailList = detailList
		reportResp.AvgFocus += report.Concentration
		reportResp.AvgPosture += report.PostureEvaluation
		reportResp.AvgLearning += report.LearningContinuity
		reportResp.AvgStudyEff += report.StudyEfficiency
		reportResp.AvgScore += report.Evaluation
		reportResp.ReportList = append(reportResp.ReportList, item)
	}
	reportResp.TotalTime = int(totalTime)
	reportResp.BeatUsers = CalculateBeatUsers(reportResp.TotalTime)
	if len(reportList) > 0 {
		reportResp.AvgFocus /= len(reportList)
		reportResp.AvgPosture /= len(reportList)
		reportResp.AvgStudyEff /= len(reportList)
		reportResp.AvgScore /= len(reportList)
		reportResp.AvgLearning /= len(reportList)
	}
	return common.Success, reportResp
}

func CalculateBeatUsers(score int) int {
	maxVal := 960
	result := 0
	if score > maxVal {
		result = 100
	} else {
		result = int(math.Round(float64(score) / float64(maxVal) * 100))
	}
	return result
}

/******************************************************************************
 * function: QueryH03WeekReport
 * description: 查询H03周报告
 * param {*gin.Context} c
 * return {*}
********************************************************************************/
// swagger:model WeekFlowState
type WeekFlowState struct {
	//时间
	FlowTime       string  `json:"flow_time"`
	TotalStudyTime float32 `json:"total_study_time"`
	//轻度占比
	LightConcentration float32 `json:"light_concentration"`
	//轻度时间，单位分
	LightStudyTime float32 `json:"light_study_time"`
	//中度占比
	MidConcentration float32 `json:"mid_concentration"`
	//中度时间，单位分
	MidStudyTime float32 `json:"mid_study_time"`
	//深度占比
	DeepConcentration float32 `json:"deep_concentration"`
	//深度时间，单位分
	DeepStudyTime float32 `json:"deep_study_time"`
}

// swagger:model WeekConcation
type WeekConcation struct {
	//时间
	ConTime string `json:"con_time"`
	//专注度
	Concentration float32 `json:"concentration"`
}

//swagger:model H03WeekReportResp
type H03WeekReportResp struct {
	mysql.H03WeekReport
	WeekFlow      []WeekFlowState `json:"week_flow_state"`
	WeekConcation []WeekConcation `json:"week_concation"`
}

func QueryH03WeekReport(c *gin.Context) (int, interface{}) {
	mac := c.Query("mac")
	if mac == "" {
		return common.ParamError, "mac required!"
	}
	weekDate := c.Query("week_date")
	if weekDate == "" {
		return common.ParamError, "start_date required!"
	}
	// 查询周报告
	var reportList []mysql.H03WeekReport
	mysql.QueryH03WeekReportByMac(mac, weekDate, &reportList)
	if len(reportList) == 0 {
		return common.NoData, "no data"
	}
	weekReport := reportList[0]
	reportResp := H03WeekReportResp{
		H03WeekReport: reportList[0],
	}
	// 查询日报表
	var dailyReportList []mysql.H03DailyReport
	mysql.QueryH03DailyReportByWeek(weekReport.Mac, weekReport.ReportYear, weekReport.ReportWeek, &dailyReportList)
	for _, dailyReport := range dailyReportList {
		totalConcentrationNums := dailyReport.LowConcentrationNum + dailyReport.MidConcentrationNum + dailyReport.HighConcentrationNum
		lightConcentration := (float32)(dailyReport.LowConcentrationNum) / (float32)(totalConcentrationNums)
		midConcentration := (float32)(dailyReport.MidConcentrationNum) / (float32)(totalConcentrationNums)
		deepConcentration := (float32)(dailyReport.HighConcentrationNum) / (float32)(totalConcentrationNums)
		weekFlow := WeekFlowState{
			FlowTime:           dailyReport.DailyDate,
			TotalStudyTime:     dailyReport.TotalStudyTime,
			LightConcentration: lightConcentration,
			LightStudyTime:     lightConcentration * dailyReport.TotalStudyTime,
			MidConcentration:   midConcentration,
			MidStudyTime:       midConcentration * dailyReport.TotalStudyTime,
			DeepConcentration:  deepConcentration,
			DeepStudyTime:      deepConcentration * dailyReport.TotalStudyTime,
		}
		reportResp.WeekFlow = append(reportResp.WeekFlow, weekFlow)
		weekConcation := WeekConcation{
			ConTime:       dailyReport.DailyDate,
			Concentration: dailyReport.AvgConcentration,
		}
		reportResp.WeekConcation = append(reportResp.WeekConcation, weekConcation)
	}
	return common.Success, reportResp
}

func QueryH03WarningEventWeekly(c *gin.Context) (int, interface{}) {
	mac := c.Query("mac")
	if mac == "" {
		return common.ParamError, "mac required!"
	}
	weekDate := c.Query("week_date")
	if weekDate == "" {
		return common.ParamError, "week_date required!"
	}
	// 查询事件告警周统计数据
	var reportList []mysql.H03WarningEventNotifyWeekStat
	mysql.QueryH03WarningEventNotifyWeekStat(mac, -1, weekDate, &reportList)
	if len(reportList) == 0 {
		return common.NoData, "no data"
	}
	return common.Success, reportList
}

func QueryH03WarningEventDaily(c *gin.Context) (int, interface{}) {
	mac := c.Query("mac")
	if mac == "" {
		return common.ParamError, "mac required!"
	}
	dailyDate := c.Query("daily_date")
	if dailyDate == "" {
		return common.ParamError, "daily_date required!"
	}
	t, err := common.StrToDate(dailyDate)
	if err != nil {
		t, err = common.StrToTime(dailyDate)
		if err != nil {
			return common.ParamError, "daily_date format error!"
		}
	}
	y, w := t.ISOWeek()
	// 查询事件告警周统计数据
	var reportList []mysql.H03WarningEventNotifyDailyStat
	mysql.QueryH03WarningEventNotifyDailyStatByWeek(mac, -1, y, w, &reportList)
	if len(reportList) == 0 {
		return common.NoData, "no data"
	}
	return common.Success, reportList
}
