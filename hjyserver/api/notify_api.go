/******************************************************************************
 * Author: liguoqiang
 * Date: 2024-04-18 18:52:18
 * LastEditors: liguoqiang
 * LastEditTime: 2024-04-19 20:12:30
 * Description:
********************************************************************************/
package api

import (
	"hjyserver/mdb"

	"github.com/gin-gonic/gin"
)

// notifySetting godoc
//
//	@Summary	notifySetting
//	@Schemes
//	@Description	set params for notify
//	@Tags			Notify
//
//	@Param			in	body	mysql.NotifySetting	true	"people notify setting"
//	@Produce		json
//	@Success		200	{none} {none}
//	@Router			/v1/notify/notifySetting [post]
func notifySetting(c *gin.Context) {
	apiCommonFunc(c, mdb.UpdateNotifySetting)
}

// // peopleNotifySetting godoc
// //
// //	@Summary	peopleNotifySetting
// //	@Schemes
// //	@Description	set params for people notify
// //	@Tags			Notify
// //
// //	@Param			in	body	mdb.PeopleNotifySettingReq	true	"people notify setting"
// //	@Produce		json
// //	@Success		200	{none} {none}
// //	@Router			/v1/notify/peopleNotifySetting [post]
// func peopleNotifySetting(c *gin.Context) {
// 	apiCommonFunc(c, mdb.PeopleNotifySetting)
// }

// // breathNotifySetting godoc
// //
// //	@Summary	breathNotifySetting
// //	@Schemes
// //	@Description	set params for breath notify
// //	@Tags			Notify
// //
// //	@Param			in	body	mdb.BreathNotifySettingReq	true	"breath notify setting"
// //	@Produce		json
// //	@Success		200	{none} {none}
// //	@Router			/v1/notify/breathNotifySetting [post]
// func breathNotifySetting(c *gin.Context) {
// 	apiCommonFunc(c, mdb.BreathNotifySetting)
// }

// // breathAbnormalNotifySetting godoc
// //
// //	@Summary	breathAbnormalNotifySetting
// //	@Schemes
// //	@Description	set params for breath abnormal notify
// //	@Tags			Notify
// //
// //	@Param			in	body	mdb.BreathAbnormalNotifySettingReq	true	"breath abnormal notify setting"
// //	@Produce		json
// //	@Success		200	{none} {none}
// //	@Router			/v1/notify/breathAbnormalNotifySetting [post]
// func breathAbnormalNotifySetting(c *gin.Context) {
// 	apiCommonFunc(c, mdb.BreathAbnormalNotifySetting)
// }

// // heartRateNotifySetting godoc
// //
// //	@Summary	heartRateNotifySetting
// //	@Schemes
// //	@Description	set params for heart rate notify
// //	@Tags			Notify
// //
// //	@Param			in	body	mdb.HeartRateNotifySettingReq	true	"heart rate notify setting"
// //	@Produce		json
// //	@Success		200	{none} {none}
// //	@Router			/v1/notify/heartRateNotifySetting [post]
// func heartRateNotifySetting(c *gin.Context) {
// 	apiCommonFunc(c, mdb.HeartRateNotifySetting)
// }

// // nurseModelSetting godoc
// //
// //	@Summary	nurseModelSetting
// //	@Schemes
// //	@Description	set params for nurse model
// //	@Tags			Notify
// //
// //	@Param			in	body	mdb.NurseModelSettingReq	true	"nurse model setting"
// //	@Produce		json
// //	@Success		200	{none} {none}
// //	@Router			/v1/notify/nurseModelSetting [post]
// func nurseModelSetting(c *gin.Context) {
// 	apiCommonFunc(c, mdb.NurseModelSetting)
// }

// // beeperSetting godoc
// //
// //	@Summary	beeperSetting
// //	@Schemes
// //	@Description	set params for beeper
// //	@Tags			Notify
// //
// //	@Param			in	body	mdb.BeeperSettingReq	true	"beeper setting"
// //	@Produce		json
// //	@Success		200	{none} {none}
// //	@Router			/v1/notify/beeperSetting [post]
// func beeperSetting(c *gin.Context) {
// 	apiCommonFunc(c, mdb.BeeperSetting)
// }

// // lightSetting godoc
// //
// //	@Summary	lightSetting
// //	@Schemes
// //	@Description	set params for light
// //	@Tags			Notify
// //
// //	@Param			in	body	mdb.LightSettingReq	true	"light setting"
// //	@Produce		json
// //	@Success		200	{none} {none}
// //	@Router			/v1/notify/lightSetting [post]
// func lightSetting(c *gin.Context) {
// 	apiCommonFunc(c, mdb.LightSetting)
// }

// queryNotifySettingByType godoc
//
//	@Summary	queryNotifySetting
//	@Schemes
//	@Description	query params for notify setting
//	@Tags			Notify
//	@Param		mac query string	true	"mac address"
//	@Param		type query int	true	"notify type"
//	@Produce		json
//	@Success		200	{object} mysql.NotifySetting
//	@Router			/v1/notify/queryNotifySettingByType [get]
func queryNotifySettingByType(c *gin.Context) {
	apiCommonFunc(c, mdb.QueryNotifySettingByType)
}

// queryAllNotifySetting godoc
//
//	@Summary	queryAllNotifySetting
//	@Schemes
//	@Description	query params for notify setting
//	@Tags			Notify
//	@Param		mac query string	true	"mac address"
//	@Produce		json
//	@Success		200	{object} mysql.NotifySetting
//	@Router			/v1/notify/queryAllNotifySetting [get]
func queryAllNotifySetting(c *gin.Context) {
	apiCommonFunc(c, mdb.QueryAllNotifySetting)
}
