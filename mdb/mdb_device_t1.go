/******************************************************************************
 * Author: liguoqiang
 * Date: 2024-08-27 18:47:25
 * LastEditors: liguoqiang
 * LastEditTime: 2025-03-18 16:22:06
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
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

const t1Tag = "mdb_device_t1"

var t1DayReportTimer *time.Timer = nil

func T1MdbInit() {
	mysql.T1ReportNotify = &T1ReportNotifyProc{}
	t1DayReportTimer = time.NewTimer(10 * time.Minute)
	go func() {
		for {
			select {
			case <-t1DayReportTimer.C:
				checkT1DayReportTimer()
				t1DayReportTimer.Reset(10 * time.Minute)
			}
		}
	}()
}
func T1MdbUnini() {
	if t1DayReportTimer != nil {
		t1DayReportTimer.Stop()
	}
}

func checkT1DayReportTimer() {
	mylog.Log.Debugln(tag, "begin call checkT1DayReportTimer...")
	var switchList []mysql.T1ReportSwitchSetting
	mysql.QueryT1DayReportOpenSwitchSetting(&switchList)
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
		status, resp := QueryT1StudyReportByMac(mac, beginDay, endDay)
		if status == common.Success {
			reportResp := resp.(T1StudyReportResp)
			if len(reportResp.ReportList) > 0 {
				endTime := reportResp.ReportList[0].EndTime
				startTime := reportResp.ReportList[len(reportResp.ReportList)-1].StartTime
				mylog.Log.Debugln(tag, "study report startTime:", startTime, "endTime:", endTime)
				type T1DayReportPush struct {
					Mac       string `json:"mac"`
					StartTime string `json:"start_time"`
					EndTime   string `json:"end_time"`
				}
				pushObj := T1DayReportPush{
					Mac:       mac,
					StartTime: startTime,
					EndTime:   endTime,
				}
				// 推送MQ
				topic := mysql.MakeT1ServerDayReportTopic(mac)
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
 * description: 定义回调函数，用来处理mysql包的T1的报告通知
 * return {*}
********************************************************************************/
type T1ReportNotifyProc struct {
}

func (me *T1ReportNotifyProc) NotifyEveryReportToOfficalAccount(userId int64, nickName string, mac string, startTime string, endTime string) (int, string) {
	return wxtools.SendEveryReportMsgToOfficalAccount(userId, nickName, mac, startTime, endTime)
}
func (me *T1ReportNotifyProc) NotifyToMiniProgram(userId int64, nickName string, title string, mac string, score int, startTime string, endTime string) (int, string) {
	return wxtools.SendReportMsgToMiniProgram(userId, nickName, title, mac, score, startTime, endTime)
}

// 用户T1 告警事件通知
func (me *T1ReportNotifyProc) NotifyWarningEventToOfficalAccount(userId int64, nickName string, mac string, tm string, event int) (int, string) {
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
			return wxtools.SendT1DeviceOnlineMsgToOfficalAccount(userId, nickName, mac, msg, tm)
		} else {
			return wxtools.SendT1DeviceStatusWarningMsgToOfficalAccount(userId, nickName, mac, msg, tm)
		}
	}(userId, nickName, mac, tm, event, msg)
}

func AskT1SyncVersion(c *gin.Context) (int, interface{}) {
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
	if device.Type != mysql.T1Type {
		return common.TypeError, "device's type is not T1_type!"
	}
	if device.Online == 0 {
		return common.DeviceOffLine, "device is offline!"
	}

	mysql.T1SyncRequest(mac)
	return common.Success, "ok"
}

//swagger:model T1RebootReq
type T1RebootReq struct {
	Mac     string `json:"mac"`
	DelayTm int64  `json:"delay_tm"`
}

func AskT1Reboot(c *gin.Context) (int, interface{}) {
	req := T1RebootReq{}
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
	if device.Type != mysql.T1Type {
		return common.TypeError, "device's type is not T1_type!"
	}
	if device.Online == 0 {
		return common.DeviceOffLine, "device is offline!"
	}
	mysql.T1RebootRequest(req.Mac, req.DelayTm)
	return common.Success, "ok"
}

func SetT1Param(c *gin.Context) (int, interface{}) {
	var req map[string]interface{} = make(map[string]interface{})
	err := c.ShouldBindJSON(&req)
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
	if device.Type != mysql.T1Type {
		return common.TypeError, "device's type is not T1_type!"
	}
	if device.Online == 0 {
		return common.DeviceOffLine, "device is offline!"
	}
	setting := &mysql.T1Setting{}
	attrDataList := make([]mysql.T1AttrData, 0)
	mysql.QueryT1AttrDataLatestByMac(mac, &attrDataList)
	if len(attrDataList) > 0 {
		setting.SetNlMode = attrDataList[0].NlMode
		setting.SetNlBrightness = attrDataList[0].NlBrightness
		setting.SetBlMode = attrDataList[0].BlMode
		setting.SetBlBrightness = attrDataList[0].BlBrightness
		setting.SetBlDelay = attrDataList[0].BlDelay
		setting.SetAlarmMode = attrDataList[0].AlarmMode
		hh, _ := strconv.Atoi(attrDataList[0].AlarmTime[:2])
		ss, _ := strconv.Atoi(attrDataList[0].AlarmTime[3:])
		setting.SetAlarmTime = []int{
			hh,
			ss,
		}
		setting.SetAlarmVol = attrDataList[0].AlarmVol
		setting.SetGestureMode = attrDataList[0].GestureMode
	}
	if _, ok := req["set_nl_mode"]; ok {
		setting.SetNlMode = int(req["set_nl_mode"].(float64))
	}
	if _, ok := req["set_nl_brightness"]; ok {
		setting.SetNlBrightness = int(req["set_nl_brightness"].(float64))
	}
	if _, ok := req["set_bl_mode"]; ok {
		setting.SetBlMode = int(req["set_bl_mode"].(float64))
	}
	if _, ok := req["set_bl_brightness"]; ok {
		setting.SetBlBrightness = int(req["set_bl_brightness"].(float64))
	}
	if _, ok := req["set_bl_delay"]; ok {
		setting.SetBlDelay = int(req["set_bl_delay"].(float64))
	}
	if _, ok := req["set_alarm_mode"]; ok {
		setting.SetAlarmMode = int(req["set_alarm_mode"].(float64))
	}
	if _, ok := req["set_alarm_time"]; ok {
		setting.SetAlarmTime = make([]int, 0)
		hh, _ := strconv.Atoi(req["set_alarm_time"].(string)[:2])
		ss, _ := strconv.Atoi(req["set_alarm_time"].(string)[3:])
		setting.SetAlarmTime = []int{
			hh,
			ss,
		}
	}
	if _, ok := req["set_alarm_vol"]; ok {
		setting.SetAlarmVol = int(req["set_alarm_vol"].(float64))
	}
	if _, ok := req["set_gesture_mode"]; ok {
		setting.SetGestureMode = int(req["set_gesture_mode"].(float64))
	}
	mysql.T1SettingRequest(mac, setting)
	return common.Success, setting
}

/******************************************************************************
 * function: SetT1ReportSwitch
 * description: 设置开关
 * param {*gin.Context} c
 * return {*}
********************************************************************************/
// swagger:model T1SwitchSettingReq
type T1SwitchSettingReq struct {
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

func SetT1ReportSwitch(c *gin.Context) (int, interface{}) {
	req := T1SwitchSettingReq{}
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
	if device.Type != mysql.T1Type {
		return common.TypeError, "device's type is not T1_type!"
	}
	settingObj := mysql.NewT1ReportSwitchSetting()
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
	switchList := make([]mysql.T1ReportSwitchSetting, 0)
	mysql.QueryT1ReportSwitchSetting(settingObj.Mac, &switchList)
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

func QueryT1ReportSwitch(c *gin.Context) (int, interface{}) {
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
	if device.Type != mysql.T1Type {
		return common.TypeError, "device's type is not T1_type!"
	}
	switchList := make([]mysql.T1ReportSwitchSetting, 0)
	mysql.QueryT1ReportSwitchSetting(mac, &switchList)
	if len(switchList) == 0 {
		return common.NoData, "no data"
	}
	return common.Success, switchList[0]
}

/******************************************************************************
 * function: QueryT1Version
 * description: 查询T1版本
 * param {*gin.Context} c
 * return {*}
********************************************************************************/
//swagger:model T1VersionResp
type T1VersionResp struct {
	mysql.T1VersionData
	IsUpdate bool `json:"is_update"`
}

func QueryT1Version(c *gin.Context) (int, interface{}) {
	mac := c.Query("mac")
	if mac == "" {
		return common.ParamError, "mac required!"
	}
	dataList := make([]mysql.T1VersionData, 0)
	mysql.QueryT1VersionByMac(mac, &dataList)
	if len(dataList) == 0 {
		return common.NoData, "no data"
	}
	resp := &T1VersionResp{
		T1VersionData: dataList[0],
		IsUpdate:      false,
	}
	otaList := make([]mysql.T1SyncOta, 0)
	mysql.QueryT1Ota(&otaList)
	if len(otaList) > 0 {
		ota := otaList[0]
		if ota.RemoteCoreVersion != dataList[0].CoreVersion || ota.RemoteBaseVersion != dataList[0].SoftwareVersion {
			resp.IsUpdate = true
		}
	}
	return common.Success, resp
}

/******************************************************************************
 * function: QueryT1LatestAttrs
 * description: 查询设备最新的一条属性
 * param {*gin.Context} c
 * return {*}
********************************************************************************/
func QueryT1LatestAttrs(c *gin.Context) (int, interface{}) {
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
	if device.Type != mysql.T1Type {
		return common.TypeError, "device's type is not T1_type!"
	}
	if device.Online == 0 {
		return common.DeviceOffLine, "device is offline!"
	}
	dataList := make([]mysql.T1AttrData, 0)
	curDay := common.GetNowDate()
	mysql.QueryT1AttrDataByMacAndDay(mac, curDay, curDay, &dataList)
	if len(dataList) == 0 {
		return common.NoData, "no data"
	}
	return common.Success, dataList[0]
}

/******************************************************************************
 * function: QueryT1LatestStudyStatus
 * description: 查询T1最近的一条学习状态
 * param {*gin.Context} c
 * return {*}
********************************************************************************/
func QueryT1LatestStudyStatus(c *gin.Context) (int, interface{}) {
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
	if device.Type != mysql.T1Type {
		return common.TypeError, "device's type is not T1_type!"
	}
	if device.Online == 0 {
		return common.DeviceOffLine, "device is offline!"
	}
	dataList := make([]mysql.T1Event, 0)
	mysql.QueryT1LatestEventByMac(mac, &dataList)
	if len(dataList) == 0 {
		return common.NoData, "no data"
	}
	return common.Success, dataList[0]
}

/******************************************************************************
 * function: QueryT1CurDayFocusStatus
 * description: 查询T1当天的专注状态
 * param {*gin.Context} c
 * return {*}
********************************************************************************/

//swagger:model T1CurDayFocusStatus
type T1CurDayFocusStatus struct {
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

func QueryT1CurDayFocusStatus(c *gin.Context) (int, interface{}) {
	mac := c.Query("mac")
	if mac == "" {
		return common.ParamError, "mac required!"
	}
	beginDay := common.GetNowDate()
	endDay := common.GetNowDate()
	attrList := make([]mysql.T1AttrData, 0)
	mysql.QueryT1AttrDataByMacAndDay(mac, beginDay, endDay, &attrList)
	status := T1CurDayFocusStatus{
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
 * function: QueryT1CurrentDayStudyTimeDetail
 * description: Query current day study time length by mac
 * param {*gin.Context} c
 * return {*}
********************************************************************************/
//swagger:model T1CurrentDayStudyTimeDetail
type T1CurrentDayStudyTimeDetail struct {
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	// 学习时长，单位分钟
	StudyTimeLen float64 `json:"study_time_len"`
}

func QueryT1CurrentDayStudyTimeDetail(c *gin.Context) (int, interface{}) {
	mac := c.Query("mac")
	if mac == "" {
		return common.ParamError, "mac required!"
	}
	today := common.GetNowDate()
	reportList := make([]mysql.T1StudyReport, 0)
	mysql.QueryT1StudyReportByDay(mac, today, today, &reportList, false)
	if len(reportList) == 0 {
		return common.NoData, "no data"
	}
	studyDetail := make([]T1CurrentDayStudyTimeDetail, 0)
	for _, report := range reportList {
		e, err := common.StrToTime(report.EndTime)
		if err != nil {
			continue
		}
		s, err := common.StrToTime(report.StartTime)
		if err != nil {
			continue
		}
		study := T1CurrentDayStudyTimeDetail{
			StartTime:    report.StartTime,
			EndTime:      report.EndTime,
			StudyTimeLen: e.Sub(s).Minutes(),
		}
		studyDetail = append(studyDetail, study)
	}
	return common.Success, studyDetail
}

/******************************************************************************
 * function: QueryT1StudyReport
 * description: 查询学习报告
 * param {*gin.Context} c
 * return {*}
********************************************************************************/

//swagger:model T1StudyReportResp
type T1StudyReportResp struct {
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
	ReportList []T1ReportItem `json:"report_list"`
}

//swagger:model T1ReportItem
type T1ReportItem struct {
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
	DetailList []T1ReportDetail `json:"detail_list"`
}

//swagger:model T1ReportDetail
type T1ReportDetail struct {
	// 状态 0:无人 1:低度 2:深度 3:离开 4:中度
	Status int `json:"status"`
	// 状态开始时间
	StartTime string `json:"start_time"`
}

func QueryT1StudyReport(c *gin.Context) (int, interface{}) {
	mac := c.Query("mac")
	if mac == "" {
		return common.ParamError, "mac required!"
	}
	startDay := c.Query("start_day")
	endDay := c.Query("end_day")
	if startDay == "" || endDay == "" {
		return common.ParamError, "start_day or end_day required!"
	}
	return QueryT1StudyReportByMac(mac, startDay, endDay)
}

func QueryT1StudyReportByMac(mac string, startDay string, endDay string) (int, interface{}) {
	reportResp := T1StudyReportResp{
		Mac:         mac,
		AvgFocus:    0,
		AvgPosture:  0,
		AvgLearning: 0,
		AvgStudyEff: 0,
		AvgScore:    0,
		ReportList:  make([]T1ReportItem, 0),
	}
	reportList := make([]mysql.T1StudyReport, 0)
	mysql.QueryT1StudyReportByDay(mac, startDay, endDay, &reportList, true)
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

		item := T1ReportItem{
			StartTime:          report.StartTime,
			EndTime:            report.EndTime,
			Focus:              report.Concentration,
			LearningContinuity: report.LearningContinuity,
			StudyEff:           report.StudyEfficiency,
			Posture:            report.PostureEvaluation,
			Score:              report.Evaluation,
		}
		detailList := make([]T1ReportDetail, 0)
		tFlow := t1
		for i := 0; i < len(report.FlowState); i++ {
			detailItem := T1ReportDetail{
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

/******************************************************************************
 * function: QueryT1WeekReport
 * description: 查询T1周报告
 * param {*gin.Context} c
 * return {*}
********************************************************************************/
// swagger:model T1WeekFlowState
type T1WeekFlowState struct {
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

// swagger:model T1WeekConcation
type T1WeekConcation struct {
	//时间
	ConTime string `json:"con_time"`
	//专注度
	Concentration float32 `json:"concentration"`
}

//swagger:model T1WeekReportResp
type T1WeekReportResp struct {
	mysql.T1WeekReport
	WeekFlow      []T1WeekFlowState `json:"week_flow_state"`
	WeekConcation []T1WeekConcation `json:"week_concation"`
}

func QueryT1WeekReport(c *gin.Context) (int, interface{}) {
	mac := c.Query("mac")
	if mac == "" {
		return common.ParamError, "mac required!"
	}
	weekDate := c.Query("week_date")
	if weekDate == "" {
		return common.ParamError, "start_date required!"
	}
	// 查询周报告
	var reportList []mysql.T1WeekReport
	mysql.QueryT1WeekReportByMac(mac, weekDate, &reportList)
	if len(reportList) == 0 {
		return common.NoData, "no data"
	}
	weekReport := reportList[0]
	reportResp := T1WeekReportResp{
		T1WeekReport: reportList[0],
	}
	// 查询日报表
	var dailyReportList []mysql.T1DailyReport
	mysql.QueryT1DailyReportByWeek(weekReport.Mac, weekReport.ReportYear, weekReport.ReportWeek, &dailyReportList)
	for _, dailyReport := range dailyReportList {
		totalConcentrationNums := dailyReport.LowConcentrationNum + dailyReport.MidConcentrationNum + dailyReport.HighConcentrationNum
		lightConcentration := (float32)(dailyReport.LowConcentrationNum) / (float32)(totalConcentrationNums)
		midConcentration := (float32)(dailyReport.MidConcentrationNum) / (float32)(totalConcentrationNums)
		deepConcentration := (float32)(dailyReport.HighConcentrationNum) / (float32)(totalConcentrationNums)
		weekFlow := T1WeekFlowState{
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
		weekConcation := T1WeekConcation{
			ConTime:       dailyReport.DailyDate,
			Concentration: dailyReport.AvgConcentration,
		}
		reportResp.WeekConcation = append(reportResp.WeekConcation, weekConcation)
	}
	return common.Success, reportResp
}

func QueryT1WarningEventWeekly(c *gin.Context) (int, interface{}) {
	mac := c.Query("mac")
	if mac == "" {
		return common.ParamError, "mac required!"
	}
	weekDate := c.Query("week_date")
	if weekDate == "" {
		return common.ParamError, "week_date required!"
	}
	// 查询事件告警周统计数据
	var reportList []mysql.T1WarningEventNotifyWeekStat
	mysql.QueryT1WarningEventNotifyWeekStat(mac, -1, weekDate, &reportList)
	if len(reportList) == 0 {
		return common.NoData, "no data"
	}
	return common.Success, reportList
}

func QueryT1WarningEventDaily(c *gin.Context) (int, interface{}) {
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
	var reportList []mysql.T1WarningEventNotifyDailyStat
	mysql.QueryT1WarningEventNotifyDailyStatByWeek(mac, -1, y, w, &reportList)
	if len(reportList) == 0 {
		return common.NoData, "no data"
	}
	return common.Success, reportList
}
