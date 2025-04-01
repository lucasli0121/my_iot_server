/******************************************************************************
 * Author: liguoqiang
 * Date: 2023-12-26 19:38:45
 * LastEditors: liguoqiang
 * LastEditTime: 2025-03-08 17:43:11
 * Description:
********************************************************************************/
package api

import (
	"hjyserver/mdb"

	"github.com/gin-gonic/gin"
)

func InitT1Actions() (map[string]gin.HandlerFunc, map[string]gin.HandlerFunc) {
	postAction := make(map[string]gin.HandlerFunc)
	getAction := make(map[string]gin.HandlerFunc)
	postAction["/T1/askT1SyncVersion"] = askT1SyncVersion
	postAction["/T1/askT1Reboot"] = askT1Reboot
	postAction["/T1/setT1Param"] = setT1Param
	postAction["/T1/setT1ReportSwitch"] = setT1ReportSwitch

	getAction["/T1/queryT1Version"] = queryT1Version
	getAction["/T1/queryT1LatestAttrs"] = queryT1LatestAttrs
	getAction["/T1/queryT1LatestStudyStatus"] = queryT1LatestStudyStatus
	getAction["/T1/queryT1CurDayFocusStatus"] = queryT1CurDayFocusStatus
	getAction["/T1/queryT1StudyReport"] = queryT1StudyReport
	getAction["/T1/queryT1ReportSwitch"] = queryT1ReportSwitch
	getAction["/T1/queryT1CurrentDayStudyTimeDetail"] = queryT1CurrentDayStudyTimeDetail
	getAction["/T1/queryT1WeekReport"] = queryT1WeekReport
	getAction["/T1/queryT1WarningEventStatDaily"] = queryT1WarningEventStatDaily
	getAction["/T1/queryT1WarningEventStatWeekly"] = queryT1WarningEventStatWeekly
	return postAction, getAction
}

// askT1SyncVersion godoc
//
//	@Summary	askT1SyncVersion
//	@Schemes
//	@Description	ask T1 device to sync version
//	@Tags			T1
//	@Param			token	query	string		false	"token"
//	@Param			mac	query	string		true	"device mac address"
//	@Produce		json
//	@Success		200	{none} {none}
//	@Router			/T1/askT1SyncVersion [post]
func askT1SyncVersion(c *gin.Context) {
	apiCommonFunc(c, mdb.AskT1SyncVersion)
}

// askT1Reboot godoc
//
//	@Summary	askT1Reboot
//	@Schemes
//	@Description	ask T1 device to reboot
//	@Tags			T1
//	@Param			token	query	string		false	"token"
//	@Param			in	body	mdb.T1RebootReq		true	"reboot information"
//	@Produce		json
//	@Success		200	{none} {none}
//	@Router			/T1/askT1Reboot [post]
func askT1Reboot(c *gin.Context) {
	apiCommonFunc(c, mdb.AskT1Reboot)
}

// setT1Param godoc
//
//	@Summary	setT1Param
//	@Schemes
//	@Description	set T1 control params
//	@Tags			T1
//	@Param			token	query	string		false	"token"
//	@Param			in	body	mysql.T1Setting		true	"params settting"
//	@Produce		json
//	@Success		200	{none} {none}
//	@Router			/T1/setT1Param [post]
func setT1Param(c *gin.Context) {
	apiCommonFunc(c, mdb.SetT1Param)
}

// setT1ReportSwitch godoc
//
//	@Summary	setT1ReportSwitch
//	@Schemes
//	@Description	set T1 report setting
//	@Tags			T1
//	@Param			token	query	string		false	"token"
//	@Param			in	body	mdb.T1SwitchSettingReq		true	"params settting"
//	@Produce		json
//	@Success		200	{none} {none}
//	@Router			/T1/setT1ReportSwitch [post]
func setT1ReportSwitch(c *gin.Context) {
	apiCommonFunc(c, mdb.SetT1ReportSwitch)
}

// queryT1Version godoc
//
//	@Summary	queryT1Version
//	@Schemes
//	@Description	查询T1版本
//	@Tags			T1
//
//	@Param			mac	query	string	true	"device mac address"
//
//	@Produce		json
//	@Success		200	{object}	mdb.T1VersionResp
//	@Router			/T1/queryT1Version [get]
func queryT1Version(c *gin.Context) {
	apiCommonFunc(c, mdb.QueryT1Version)
}

// queryT1LatestAttrs godoc
//
//	@Summary	queryT1LatestAttrs
//	@Schemes
//	@Description	查询T1属性
//	@Tags			T1
//
//	@Param			mac	query	string	true	"device mac address"
//
//	@Produce		json
//	@Success		200	{object}	mysql.T1AttrData
//	@Router			/T1/queryT1LatestAttrs [get]
func queryT1LatestAttrs(c *gin.Context) {
	apiCommonFunc(c, mdb.QueryT1LatestAttrs)
}

// queryT1LatestStudyStatus godoc
//
//	@Summary	queryT1LatestStudyStatus
//	@Schemes
//	@Description	查询T1学习状态
//	@Tags			T1
//
//	@Param			mac	query	string	true	"device mac address"
//
//	@Produce		json
//	@Success		200	{object}	mysql.T1Event
//	@Router			/T1/queryT1LatestStudyStatus [get]
func queryT1LatestStudyStatus(c *gin.Context) {
	apiCommonFunc(c, mdb.QueryT1LatestStudyStatus)
}

// queryT1CurDayFocusStatus godoc
//
//	@Summary	queryT1CurDayFocusStatus
//	@Schemes
//	@Description	查询T1当天的专注状态
//	@Tags			T1
//
//	@Param			mac	query	string	true	"device mac address"
//
//	@Produce		json
//	@Success		200	{object}	mdb.T1CurDayFocusStatus
//	@Router			/T1/queryT1CurDayFocusStatus [get]
func queryT1CurDayFocusStatus(c *gin.Context) {
	apiCommonFunc(c, mdb.QueryT1CurDayFocusStatus)
}

// queryT1StudyReport godoc
//
//	@Summary	queryT1StudyReport
//	@Schemes
//	@Description	根据日期查询T1学习报告
//	@Tags			T1
//
//	@Param			mac	query	string	true	"device mac address"
//	@Param			start_day	query	string	true	"start_day"
//	@Param			end_day	query	string	true	"end_day"
//
//	@Produce		json
//	@Success		200	{object}	mdb.T1StudyReportResp
//	@Router			/T1/queryT1StudyReport [get]
func queryT1StudyReport(c *gin.Context) {
	apiCommonFunc(c, mdb.QueryT1StudyReport)
}

// queryT1CurrentDayStudyTimeDetail godoc
//
//	@Summary	queryT1CurrentDayStudyTimeDetail
//	@Schemes
//	@Description	查询T1当天学习时长的详情数据
//	@Tags			T1
//
//	@Param			mac	query	string	true	"device mac address"
//
//	@Produce		json
//	@Success		200	{object}	mdb.T1CurrentDayStudyTimeDetail
//	@Router			/T1/queryT1CurrentDayStudyTimeDetail [get]
func queryT1CurrentDayStudyTimeDetail(c *gin.Context) {
	apiCommonFunc(c, mdb.QueryT1CurrentDayStudyTimeDetail)
}

// queryT1ReportSwitch godoc
//
//	@Summary	queryT1ReportSwitch
//	@Schemes
//	@Description	根据mac查报告开关设置
//	@Tags			T1
//
//	@Param			mac	query	string	true	"device mac address"
//
//	@Produce		json
//	@Success		200	{object}	mysql.T1ReportSwitchSetting
//	@Router			/T1/queryT1ReportSwitch [get]
func queryT1ReportSwitch(c *gin.Context) {
	apiCommonFunc(c, mdb.QueryT1ReportSwitch)
}

// queryT1WeekReport godoc
//
//	@Summary	queryT1WeekReport
//	@Schemes
//	@Description	根据mac以及日期查询周报告
//	@Tags			T1
//
//	@Param			mac	query	string	true	"device mac address"
//	@Param			week_date	query	string	true	"周报日期"
//
//	@Produce		json
//	@Success		200	{object}	mdb.T1WeekReportResp
//	@Router			/T1/queryT1WeekReport [get]
func queryT1WeekReport(c *gin.Context) {
	apiCommonFunc(c, mdb.QueryT1WeekReport)
}

// queryT1WarningEventStatDaily godoc
//
//	@Summary	queryT1WarningEventStatDaily
//	@Schemes
//	@Description	根据mac以及日期查询一周七天每天的告警事件统计次数
//	@Tags			T1
//
//	@Param			mac	query	string	true	"device mac address"
//	@Param			daily_date	query	string	true	"查询的日期"
//
//	@Produce		json
//	@Success		200	{object}	mysql.T1WarningEventNotifyDailyStat
//	@Router			/T1/queryT1WarningEventStatDaily [get]
func queryT1WarningEventStatDaily(c *gin.Context) {
	apiCommonFunc(c, mdb.QueryT1WarningEventDaily)
}

// queryT1WarningEventStatWeekly godoc
//
//	@Summary	queryT1WarningEventStatWeekly
//	@Schemes
//	@Description	根据mac以及日期查询一周告警事件统计次数
//	@Tags			T1
//
//	@Param			mac	query	string	true	"device mac address"
//	@Param			week_date	query	string	true	"周报日期"
//
//	@Produce		json
//	@Success		200	{object}	mysql.T1WarningEventNotifyWeekStat
//	@Router			/T1/queryT1WarningEventStatWeekly [get]
func queryT1WarningEventStatWeekly(c *gin.Context) {
	apiCommonFunc(c, mdb.QueryT1WarningEventWeekly)
}
