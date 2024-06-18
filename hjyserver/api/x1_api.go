/******************************************************************************
 * Author: liguoqiang
 * Date: 2023-12-26 19:38:45
 * LastEditors: liguoqiang
 * LastEditTime: 2024-06-02 19:08:02
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

//	@BasePath	/v1

// askX1RealData godoc
//
//	@Summary	askX1RealData
//	@Schemes
//	@Description	ask X1 device to send real data
//	@Tags			X1 device
//
//	@Param			in	body	mdb.AskX1RealDataReq	true	"ack X1 info"
//	@Produce		json
//	@Success		200	{none} {none}
//	@Router			/v1/device/askX1RealData [post]
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
//
//	@Param			in	body	mdb.CleanX1EventReq	true	"clean X1 event"
//	@Produce		json
//	@Success		200	{none} {none}
//	@Router			/v1/device/cleanX1Event [post]
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
//
//	@Param			in	body	mdb.SleepX1SwitchReq	true	"X1 switch"
//	@Produce		json
//	@Success		200	{none} {none}
//	@Router			/v1/device/sleepX1Switch [post]
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
//
//	@Param			in	body	mdb.ImproveDisturbedReq	true	"X1 distrube switch"
//	@Produce		json
//	@Success		200	{none} {none}
//	@Router			/v1/device/improveDisturbed [post]
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
//	@Router			/v1/device/queryX1RealDataJson [get]
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
//	@Router			/v1/device/queryX1SleepReportJson [get]
func queryX1SleepReportJson(c *gin.Context) {
	apiCommonFunc(c, mdb.QueryX1SleepReportJson)
}
