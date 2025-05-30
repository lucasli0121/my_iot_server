/******************************************************************************
 * Author: liguoqiang
 * Date: 2024-04-18 18:52:18
 * LastEditors: liguoqiang
 * LastEditTime: 2024-08-07 11:03:56
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
//	@Param			token	query	string		false	"token"
//	@Param			in	body	mysql.NotifySetting	true	"people notify setting"
//	@Produce		json
//	@Success		200	{none} {none}
//	@Router			/notify/notifySetting [post]
func notifySetting(c *gin.Context) {
	apiCommonFunc(c, mdb.UpdateNotifySetting)
}

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
//	@Router			/notify/queryNotifySettingByType [get]
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
//	@Router			/notify/queryAllNotifySetting [get]
func queryAllNotifySetting(c *gin.Context) {
	apiCommonFunc(c, mdb.QueryAllNotifySetting)
}
