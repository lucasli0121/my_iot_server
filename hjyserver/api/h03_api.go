/******************************************************************************
 * Author: liguoqiang
 * Date: 2023-12-26 19:38:45
 * LastEditors: liguoqiang
 * LastEditTime: 2025-01-21 20:26:17
 * Description:
********************************************************************************/
package api

import (
	"hjyserver/mdb"

	"github.com/gin-gonic/gin"
)

func InitH03Actions() (map[string]gin.HandlerFunc, map[string]gin.HandlerFunc) {
	postAction := make(map[string]gin.HandlerFunc)
	getAction := make(map[string]gin.HandlerFunc)
	postAction["/h03/askH03SyncVersion"] = askH03SyncVersion
	postAction["/h03/askH03Reboot"] = askH03Reboot
	postAction["/h03/setH03Param"] = setH03Param
	postAction["/h03/setH03ReportSwitch"] = setH03ReportSwitch

	getAction["/h03/queryH03Version"] = queryH03Version
	getAction["/h03/queryH03LatestAttrs"] = queryH03LatestAttrs
	getAction["/h03/queryH03LatestStudyStatus"] = queryH03LatestStudyStatus
	getAction["/h03/queryH03CurDayFocusStatus"] = queryH03CurDayFocusStatus
	getAction["/h03/queryH03StudyReport"] = queryH03StudyReport
	getAction["/h03/queryH03ReportSwitch"] = queryH03ReportSwitch
	getAction["/h03/queryH03CurrentDayStudyTimeDetail"] = queryH03CurrentDayStudyTimeDetail
	getAction["/h03/queryH03WeekReport"] = queryH03WeekReport
	getAction["/h03/queryH03WarningEventStatDaily"] = queryH03WarningEventStatDaily
	getAction["/h03/queryH03WarningEventStatWeekly"] = queryH03WarningEventStatWeekly
	return postAction, getAction
}

// askH03SyncVersion godoc
//
//	@Summary	askH03SyncVersion
//	@Schemes
//	@Description	ask H03 device to sync version
//	@Tags			H03
//	@Param			token	query	string		false	"token"
//	@Param			mac	query	string		true	"device mac address"
//	@Produce		json
//	@Success		200	{none} {none}
//	@Router			/h03/askH03SyncVersion [post]
func askH03SyncVersion(c *gin.Context) {
	apiCommonFunc(c, mdb.AskH03SyncVersion)
}

// askH03Reboot godoc
//
//	@Summary	askH03Reboot
//	@Schemes
//	@Description	ask H03 device to reboot
//	@Tags			H03
//	@Param			token	query	string		false	"token"
//	@Param			in	body	mdb.H03RebootReq		true	"reboot information"
//	@Produce		json
//	@Success		200	{none} {none}
//	@Router			/h03/askH03Reboot [post]
func askH03Reboot(c *gin.Context) {
	apiCommonFunc(c, mdb.AskH03Reboot)
}

// setH03Param godoc
//
//	@Summary	setH03Param
//	@Schemes
//	@Description	set H03 control params
//	@Tags			H03
//	@Param			token	query	string		false	"token"
//	@Param			in	body	mdb.H03SettingReq		true	"params settting"
//	@Produce		json
//	@Success		200	{none} {none}
//	@Router			/h03/setH03Param [post]
func setH03Param(c *gin.Context) {
	apiCommonFunc(c, mdb.SetH03Param)
}

// setH03ReportSwitch godoc
//
//	@Summary	setH03ReportSwitch
//	@Schemes
//	@Description	set H03 report setting
//	@Tags			H03
//	@Param			token	query	string		false	"token"
//	@Param			in	body	mdb.H03SwitchSettingReq		true	"params settting"
//	@Produce		json
//	@Success		200	{none} {none}
//	@Router			/h03/setH03ReportSwitch [post]
func setH03ReportSwitch(c *gin.Context) {
	apiCommonFunc(c, mdb.SetH03ReportSwitch)
}

// queryH03Version godoc
//
//	@Summary	queryH03Version
//	@Schemes
//	@Description	查询H03版本
//	@Tags			H03
//
//	@Param			mac	query	string	true	"device mac address"
//
//	@Produce		json
//	@Success		200	{object}	mdb.H03VersionResp
//	@Router			/h03/queryH03Version [get]
func queryH03Version(c *gin.Context) {
	apiCommonFunc(c, mdb.QueryH03Version)
}

// queryH03LatestAttrs godoc
//
//	@Summary	queryH03LatestAttrs
//	@Schemes
//	@Description	查询H03属性
//	@Tags			H03
//
//	@Param			mac	query	string	true	"device mac address"
//
//	@Produce		json
//	@Success		200	{object}	mysql.H03AttrData
//	@Router			/h03/queryH03LatestAttrs [get]
func queryH03LatestAttrs(c *gin.Context) {
	apiCommonFunc(c, mdb.QueryH03LatestAttrs)
}

// queryH03LatestStudyStatus godoc
//
//	@Summary	queryH03LatestStudyStatus
//	@Schemes
//	@Description	查询H03学习状态
//	@Tags			H03
//
//	@Param			mac	query	string	true	"device mac address"
//
//	@Produce		json
//	@Success		200	{object}	mysql.H03Event
//	@Router			/h03/queryH03LatestStudyStatus [get]
func queryH03LatestStudyStatus(c *gin.Context) {
	apiCommonFunc(c, mdb.QueryH03LatestStudyStatus)
}

// queryH03CurDayFocusStatus godoc
//
//	@Summary	queryH03CurDayFocusStatus
//	@Schemes
//	@Description	查询H03当天的专注状态
//	@Tags			H03
//
//	@Param			mac	query	string	true	"device mac address"
//
//	@Produce		json
//	@Success		200	{object}	mdb.H03CurDayFocusStatus
//	@Router			/h03/queryH03CurDayFocusStatus [get]
func queryH03CurDayFocusStatus(c *gin.Context) {
	apiCommonFunc(c, mdb.QueryH03CurDayFocusStatus)
}

// queryH03StudyReport godoc
//
//	@Summary	queryH03StudyReport
//	@Schemes
//	@Description	根据日期查询H03学习报告
//	@Tags			H03
//
//	@Param			mac	query	string	true	"device mac address"
//	@Param			start_day	query	string	true	"start_day"
//	@Param			end_day	query	string	true	"end_day"
//
//	@Produce		json
//	@Success		200	{object}	mdb.H03StudyReportResp
//	@Router			/h03/queryH03StudyReport [get]
func queryH03StudyReport(c *gin.Context) {
	apiCommonFunc(c, mdb.QueryH03StudyReport)
}

// queryH03CurrentDayStudyTimeDetail godoc
//
//	@Summary	queryH03CurrentDayStudyTimeDetail
//	@Schemes
//	@Description	查询H03当天学习时长的详情数据
//	@Tags			H03
//
//	@Param			mac	query	string	true	"device mac address"
//
//	@Produce		json
//	@Success		200	{object}	mdb.H03CurrentDayStudyTimeDetail
//	@Router			/h03/queryH03CurrentDayStudyTimeDetail [get]
func queryH03CurrentDayStudyTimeDetail(c *gin.Context) {
	apiCommonFunc(c, mdb.QueryH03CurrentDayStudyTimeDetail2)
}

// queryH03ReportSwitch godoc
//
//	@Summary	queryH03ReportSwitch
//	@Schemes
//	@Description	根据mac查报告开关设置
//	@Tags			H03
//
//	@Param			mac	query	string	true	"device mac address"
//
//	@Produce		json
//	@Success		200	{object}	mysql.H03ReportSwitchSetting
//	@Router			/h03/queryH03ReportSwitch [get]
func queryH03ReportSwitch(c *gin.Context) {
	apiCommonFunc(c, mdb.QueryH03ReportSwitch)
}

// queryH03WeekReport godoc
//
//	@Summary	queryH03WeekReport
//	@Schemes
//	@Description	根据mac以及日期查询周报告
//	@Tags			H03
//
//	@Param			mac	query	string	true	"device mac address"
//	@Param			week_date	query	string	true	"周报日期"
//
//	@Produce		json
//	@Success		200	{object}	mdb.H03WeekReportResp
//	@Router			/h03/queryH03WeekReport [get]
func queryH03WeekReport(c *gin.Context) {
	apiCommonFunc(c, mdb.QueryH03WeekReport)
}

// queryH03WarningEventStatDaily godoc
//
//	@Summary	queryH03WarningEventStatDaily
//	@Schemes
//	@Description	根据mac以及日期查询一周七天每天的告警事件统计次数
//	@Tags			H03
//
//	@Param			mac	query	string	true	"device mac address"
//	@Param			daily_date	query	string	true	"查询的日期"
//
//	@Produce		json
//	@Success		200	{object}	mysql.H03WarningEventNotifyDailyStat
//	@Router			/h03/queryH03WarningEventStatDaily [get]
func queryH03WarningEventStatDaily(c *gin.Context) {
	apiCommonFunc(c, mdb.QueryH03WarningEventDaily)
}

// queryH03WarningEventStatWeekly godoc
//
//	@Summary	queryH03WarningEventStatWeekly
//	@Schemes
//	@Description	根据mac以及日期查询一周告警事件统计次数
//	@Tags			H03
//
//	@Param			mac	query	string	true	"device mac address"
//	@Param			week_date	query	string	true	"周报日期"
//
//	@Produce		json
//	@Success		200	{object}	mysql.H03WarningEventNotifyWeekStat
//	@Router			/h03/queryH03WarningEventStatWeekly [get]
func queryH03WarningEventStatWeekly(c *gin.Context) {
	apiCommonFunc(c, mdb.QueryH03WarningEventWeekly)
}
