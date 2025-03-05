/******************************************************************************
 * Author: liguoqiang
 * Date: 2023-12-26 19:38:45
 * LastEditors: liguoqiang
 * LastEditTime: 2024-08-07 10:59:39
 * Description:
********************************************************************************/
/*********************************************************************
*
**********************************************************************/

package api

import (
	"net/http"

	"hjyserver/exception"
	"hjyserver/mdb"

	"github.com/gin-gonic/gin"
)

// askX1RealData godoc
//
//	@Summary	askX1RealData
//	@Schemes
//	@Description	ask X1 device to send real data
//	@Tags			X1 device
//	@Param			token	query	string		false	"token"
//	@Param			in	body	mdb.AskX1RealDataReq	true	"ack X1 info"
//	@Produce		json
//	@Success		200	{none} {none}
//	@Router			/device/askX1RealData [post]
func askX1RealData(c *gin.Context) {
	exception.TryEx{
		Try: func() {
			status, result := mdb.AskX1RealData(c)
			respJSON(c, status, result)
		},
		Catch: func(e exception.Exception) {
			respJSON(c, http.StatusBadRequest, e.Msg)
		},
	}.Run()
}

// cleanX1Event godoc
//
//	@Summary	cleanX1Event
//	@Schemes
//	@Description	clean event from X1 device
//	@Tags			X1 device
//	@Param			token	query	string		false	"token"
//	@Param			in	body	mdb.CleanX1EventReq	true	"clean X1 event"
//	@Produce		json
//	@Success		200	{none} {none}
//	@Router			/device/cleanX1Event [post]
func cleanX1Event(c *gin.Context) {
	exception.TryEx{
		Try: func() {
			status, result := mdb.CleanX1Event(c)
			respJSON(c, status, result)
		},
		Catch: func(e exception.Exception) {
			respJSON(c, http.StatusBadRequest, e.Msg)
		},
	}.Run()
}

// sleepX1Switch godoc
//
//	@Summary	sleepX1Switch
//	@Schemes
//	@Description	change sleep switch of X1 device
//	@Tags			X1 device
//	@Param			token	query	string		false	"token"
//	@Param			in	body	mdb.SleepX1SwitchReq	true	"X1 switch"
//	@Produce		json
//	@Success		200	{none} {none}
//	@Router			/device/sleepX1Switch [post]
func sleepX1Switch(c *gin.Context) {
	exception.TryEx{
		Try: func() {
			status, result := mdb.SleepX1Switch(c)
			respJSON(c, status, result)
		},
		Catch: func(e exception.Exception) {
			respJSON(c, http.StatusBadRequest, e.Msg)
		},
	}.Run()
}

// improveDisturbed godoc
//
//	@Summary	improveDisturbed
//	@Schemes
//	@Description	import x1 device disturbed data
//	@Tags			X1 device
//	@Param			token	query	string		false	"token"
//	@Param			in	body	mdb.ImproveDisturbedReq	true	"X1 distrube switch"
//	@Produce		json
//	@Success		200	{none} {none}
//	@Router			/device/improveDisturbed [post]
func improveDisturbed(c *gin.Context) {
	apiCommonFunc(c, mdb.ImproveDisturbed)
}

// queryX1RealDataJson godoc
//
//	@Summary	queryX1RealDataJson
//	@Schemes
//	@Description	query X1 real data in json format
//	@Tags			X1 device
//
//	@Param			mac	query	string	true	"device mac address"
//
// @Param create_date query string false "create date"
//
//	@Produce		json
//	@Success		200	{object} mysql.X1RealDataOrigin
//	@Router			/device/queryX1RealDataJson [get]
func queryX1RealDataJson(c *gin.Context) {
	apiCommonFunc(c, mdb.QueryX1RealDataJson)
}

// queryX1SleepReportJson godoc
//
//	@Summary	queryX1SleepReportJson
//	@Schemes
//	@Description	query X1 sleep report in json format
//	@Tags			X1 device
//
//	@Param			mac	query	string	true	"device mac address"
//
// @Param create_date query string false "create date"
//
//	@Produce		json
//	@Success		200	{object} mysql.X1DayReportOrigin
//	@Router			/device/queryX1SleepReportJson [get]
func queryX1SleepReportJson(c *gin.Context) {
	apiCommonFunc(c, mdb.QueryX1SleepReportJson)
}

// recoverX1SleepReport godoc
//
//	@Summary	recoverX1SleepReport
//	@Schemes
//	@Description	recover x1 device sleep report from json table
//	@Tags			X1 device
//	@Param			token	query	string		false	"token"
//	@Param			in	body	mdb.RecoverX1SleepReportReq	true	"X1 sleep report"
//	@Produce		json
//	@Success		200	{none} {none}
//	@Router			/device/recoverX1SleepReport [post]
func recoverX1SleepReport(c *gin.Context) {
	apiCommonFunc(c, mdb.RecoverX1SleepReport)
}
