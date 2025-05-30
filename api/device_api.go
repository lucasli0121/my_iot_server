/*********************************************************************
*
**********************************************************************/

package api

import (
	"net/http"
	"strconv"

	"hjyserver/exception"
	"hjyserver/mdb"

	"github.com/gin-gonic/gin"
)

func InitDeviceActions() (map[string]gin.HandlerFunc, map[string]gin.HandlerFunc) {
	postAction := make(map[string]gin.HandlerFunc)
	getAction := make(map[string]gin.HandlerFunc)

	getAction["/device/queryById"] = queryDeviceById
	getAction["/device/queryByUser"] = queryDeviceByUser
	getAction["/device/queryBindByMac"] = queryBindByMac
	getAction["/device/queryHeartRate"] = queryHeartRate
	getAction["/device/statsHeartRateByMinute"] = statsHeartRateByMinute
	getAction["/device/queryX1RealDataJson"] = queryX1RealDataJson
	getAction["/device/queryX1SleepReportJson"] = queryX1SleepReportJson
	getAction["/device/querySleepReport"] = querySleepReport
	getAction["/device/queryDateListInReport"] = queryDateListInReport
	getAction["/device/queryFallCheckStatus"] = queryFallCheckStatus
	getAction["/device/queryAlarmRecord"] = queryAlarmRecord
	getAction["/device/queryFallExistRecord"] = queryFallExistRecord
	getAction["/device/queryFallParams"] = queryFallParams
	getAction["/device/queryLampRealData"] = queryLampRealData
	getAction["/device/queryLampEvent"] = queryLampEvent
	getAction["/device/queryLampControl"] = queryLampControl
	getAction["/device/queryLampReportStatus"] = queryLampReportStatus
	getAction["/device/queryStudyRoom"] = queryStudyRoom
	getAction["/device/queryInviteStudyRoom"] = queryInviteStudyRoom
	getAction["/device/queryRankingByStudyRoom"] = queryRankingByStudyRoom
	getAction["/device/statsLampFlowData"] = statsLampFlowData
	getAction["/device/queryShareUsers"] = queryShareUsers
	getAction["/device/queryUnconfirmedShareDevices"] = queryUnconfirmedShareDevices
	getAction["/device/queryTransferUsers"] = queryTransferUsers
	getAction["/device/queryUnconfirmedTransferDevices"] = queryUnconfirmedTransferDevices
	getAction["/device/queryDeviceOverview"] = queryDeviceOverview

	// post device tag action
	postAction["/device/insert"] = insertDevice
	postAction["/device/update"] = updateDevice
	postAction["/device/share"] = shareDevice
	postAction["/device/shareDeviceToPhoneWithMac"] = shareDeviceToPhoneWithMac
	postAction["/device/confirmSharedDevice"] = confirmSharedDevice
	postAction["/device/modifySharedDeviceRemark"] = modifySharedDeviceRemark
	postAction["/device/removeSharedDevice"] = removeSharedDevice
	postAction["/device/transferDeviceToPhoneWithMac"] = transferDeviceToPhoneWithMac
	postAction["/device/confirmTransferDevice"] = confirmTransferDevice
	postAction["/device/insertFallParams"] = insertFallParams
	postAction["/device/openLampRealData"] = openLampRealData
	postAction["/device/controlLamp"] = controlLamp
	postAction["/device/createStudyRoom"] = createStudyRoom
	postAction["/device/modifyStudyRoom"] = modifyStudyRoom
	postAction["/device/releaseStudyRoom"] = releaseStudyRoom
	postAction["/device/askEd713RealData"] = askEd713RealData
	postAction["/device/askX1RealData"] = askX1RealData
	postAction["/device/cleanX1Event"] = cleanX1Event
	postAction["/device/sleepX1Switch"] = sleepX1Switch
	postAction["/device/improveDisturbed"] = improveDisturbed
	postAction["/device/recoverX1SleepReport"] = recoverX1SleepReport
	postAction["/device/updateDeviceOverview"] = updateDeviceOverview

	return postAction, getAction
}

// queryById godoc
//
//	@Summary	queryById
//	@Schemes
//	@Description	query device infomaion
//	@Tags			device
//
//	@Param			id	query	int	true	"device id"
//
//	@Produce		json
//	@Success		200	{object}	mysql.Device
//	@Router			/device/queryById [get]
func queryDeviceById(c *gin.Context) {
	exception.TryEx{
		Try: func() {
			id := c.Query("id")
			if id == "" {
				respJSON(c, http.StatusBadRequest, "id required")
			}
			mId, err := strconv.ParseInt(id, 10, 64)
			if err != nil {
				respJSON(c, http.StatusBadRequest, "id wrong")
			}
			status, result := mdb.QueryDeviceById(mId)
			respJSON(c, status, result)
		},
		Catch: func(e exception.Exception) {
			respJSON(c, e.Code, e.Msg)
		},
	}.Run()
}

// queryBindByMac godoc
//
//	@Summary	queryBindByMac
//	@Schemes
//	@Description	query device has bind by mac
//	@Tags			device
//
// @Param user_id query int true "user id"
//
//	@Param			mac	query	string	true	"device mac address"
//
//	@Produce		json
//	@Success		200	{object}	mdb.DeviceBindResp
//	@Router			/device/queryBindByMac [get]
func queryBindByMac(c *gin.Context) {
	exception.TryEx{
		Try: func() {
			userId := c.Query("user_id")
			if userId == "" {
				respJSON(c, http.StatusBadRequest, "user id required")
			}
			mac := c.Query("mac")
			if mac == "" {
				respJSON(c, http.StatusBadRequest, "mac required")
			}
			mId, err := strconv.ParseInt(userId, 10, 64)
			if err != nil {
				respJSON(c, http.StatusBadRequest, "id wrong")
			}
			status, result := mdb.QueryBindDeviceByMac(mId, mac)
			respJSON(c, status, result)
		},
		Catch: func(e exception.Exception) {
			respJSON(c, http.StatusBadRequest, e.Msg)
		},
	}.Run()
}

// queryByUser godoc
//
//	@Summary	queryByUser
//	@Schemes
//	@Description	query device infomaion by user id
//	@Tags			device
//	@Produce		json
//
//	@Param			user_id	query	int		true	"user id"
//	@Param			flag	query	int		true	"flag"
//
//	@Success		200		{object}	mysql.Device
//	@Router			/device/queryByUser [get]
func queryDeviceByUser(c *gin.Context) {
	apiCommonFunc(c, mdb.QueryDeviceByUser)
}

// insertDevice godoc
//
//	@Summary	insertDevice
//	@Schemes
//	@Description	insert user device
//	@Tags			device
//	@Produce		json
//
//	@Param			token	query	string		false	"token"
//	@Param			in	body	mdb.NewDeviceReq		true	"user device information"
//
//	@Success		200		{object}	mysql.UserDevice
//	@Router			/device/insert [post]
func insertDevice(c *gin.Context) {
	apiCommonFunc(c, mdb.InsertDevice)
}

/*
updateDevice...
*/

// updateDevice godoc
//
//	@Summary	updateDevice
//	@Schemes
//	@Description	update user device
//	@Tags			device
//	@Produce		json
//	@Param			token	query	string		false	"token"
//	@Param			in	body	mdb.UpdateDeviceReq		true	"user device information"
//
//	@Success		200		{object}	mysql.Device
//	@Router			/device/update [post]
func updateDevice(c *gin.Context) {
	apiCommonFunc(c, mdb.UpdateDevice)
}

/******************************************************************************
 * function: ShareDevice
 * description: share a device to other user
 * return {*}
********************************************************************************/

// shareDevice godoc
//
//	@Summary	shareDevice
//	@Schemes
//	@Description	update user device
//	@Tags			device
//	@Produce		json
//	@Param			token	query	string		false	"token"
//	@Param			in	body	mdb.ShareDeviceReq		true	"user device information"
//
//	@Success		200		{object}	mysql.UserDevice
//	@Router			/device/share [post]
func shareDevice(c *gin.Context) {
	apiCommonFunc(c, mdb.ShareDeviceToOtherUser)
}

// shareDeviceToPhoneWithMac godoc
//
//	@Summary	shareDeviceToPhoneWithMac
//	@Schemes
//	@Description	share device to other user
//	@Tags			device
//	@Produce		json
//	@Param			token	query	string		false	"token"
//	@Param			in	body	mdb.ShareDeviceWithMacReq		true	"user device information"
//
//	@Success		200		{object}	mysql.UserShareDevice
//	@Router			/device/shareDeviceToPhoneWithMac [post]
func shareDeviceToPhoneWithMac(c *gin.Context) {
	apiCommonFunc(c, mdb.ShareDeviceToPhoneWithMac)
}

// confirmSharedDevice godoc
//
//	@Summary	confirmSharedDevice
//	@Schemes
//	@Description	confirm shared device
//	@Tags			device
//	@Produce		json
//	@Param			token	query	string		false	"token"
//	@Param			in	body	mdb.ConfirmSharedReq		true	"confirm shared device request"
//
//	@Success		200		{string}	{"Confirmed successfully"}
//	@Router			/device/confirmSharedDevice [post]
func confirmSharedDevice(c *gin.Context) {
	apiCommonFunc(c, mdb.ConfirmSharedDevice)
}

// modifySharedDeviceRemark godoc
//
//	@Summary	modifySharedDeviceRemark
//	@Schemes
//	@Description	modify shared device remark
//	@Tags			device
//	@Produce		json
//	@Param			token	query	string		false	"token"
//	@Param			in	body	mdb.ModifySharedDeviceRemarkReq		true	"modify share device remark"
//
//	@Success		200		{object}	mysql.UserShareDevice
//	@Router			/device/modifySharedDeviceRemark [post]
func modifySharedDeviceRemark(c *gin.Context) {
	apiCommonFunc(c, mdb.ModifySharedDeviceRemark)
}

// removeSharedDevice godoc
//
//	@Summary	removeSharedDevice
//	@Schemes
//	@Description	delete shared device
//	@Tags			device
//	@Produce		json
//	@Param			token	query	string		false	"token"
//	@Param			in	body	mdb.RemoveSharedDeviceReq		true	"delete share device"
//
//	@Success		200		{string}	{"delete shared device successfully"}
//	@Router			/device/removeSharedDevice [post]
func removeSharedDevice(c *gin.Context) {
	apiCommonFunc(c, mdb.RemoveSharedDevice)
}

// queryShareUsers godoc
//
//	@Summary	queryShareUsers
//	@Schemes
//	@Description	query users to which device shared
//	@Tags			device
//	@Produce		json
//	@Param			user_id	query	string		true	"login user id"
//	@Param			mac	query	string		true	"device mac address"
//
//	@Success		200		{object}	mysql.UserShareDeviceDetail
//	@Router			/device/queryShareUsers [get]
func queryShareUsers(c *gin.Context) {
	apiCommonFunc(c, mdb.QueryShareUsers)
}

// queryUnconfirmedShareDevices godoc
//
//	@Summary	queryUnconfirmedShareDevices
//	@Schemes
//	@Description	query unconfirmed share devices
//	@Tags			device
//	@Produce		json
//	@Param			user_id	query	string		true	"login user id"
//	@Param			mac	query	string		false	"设备mac地址"
//	@Success		200		{object}	mysql.UserShareDeviceDetail
//	@Router			/device/queryUnconfirmedShareDevices [get]
func queryUnconfirmedShareDevices(c *gin.Context) {
	apiCommonFunc(c, mdb.QueryUnconfirmedShareDevices)
}

// queryTransferUsers godoc
//
//	@Summary	queryTransferUsers
//	@Schemes
//	@Description	query users to which device transfered
//	@Tags			device
//	@Produce		json
//	@Param			user_id	query	string		true	"login user id"
//	@Param			mac	query	string		true	"device mac address"
//
//	@Success		200		{object}	mysql.UserTransferDeviceDetail
//	@Router			/device/queryTransferUsers [get]
func queryTransferUsers(c *gin.Context) {
	apiCommonFunc(c, mdb.QueryTransferedUsers)
}

// queryUnconfirmedTransferDevices godoc
//
//	@Summary	queryUnconfirmedTransferDevices
//	@Schemes
//	@Description	query unconfirmed transfered devices
//	@Tags			device
//	@Produce		json
//	@Param			user_id	query	string		true	"login user id"
//	@Param			mac	query	string		false	"device mac address"
//	@Success		200		{object}	mysql.UserTransferDeviceDetail
//	@Router			/device/queryUnconfirmedTransferDevices [get]
func queryUnconfirmedTransferDevices(c *gin.Context) {
	apiCommonFunc(c, mdb.QueryUnconfirmedTransferDevices)
}

// transferDeviceToPhoneWithMac godoc
//
//	@Summary	transferDeviceToPhoneWithMac
//	@Schemes
//	@Description	transfer device to other user
//	@Tags			device
//	@Produce		json
//	@Param			token	query	string		false	"token"
//	@Param			in	body	mdb.TransferDeviceWithMacReq		true	"user device information"
//
//	@Success		200		{object}	mysql.UserDevice
//	@Router			/device/transferDeviceToPhoneWithMac [post]
func transferDeviceToPhoneWithMac(c *gin.Context) {
	apiCommonFunc(c, mdb.TransferDeviceToPhoneWithMac)
}

// confirmTransferDevice godoc
//
//	@Summary	confirmTransferDevice
//	@Schemes
//	@Description	confirmed transfer device to other user
//	@Tags			device
//	@Produce		json
//	@Param			token	query	string		false	"token"
//	@Param			in	body	mdb.ConfirmTransferReq		true	"confirmed transfer device"
//
//	@Success		200		{string}	{"Confirmed successfully"}
//	@Router			/device/confirmTransferDevice [post]
func confirmTransferDevice(c *gin.Context) {
	apiCommonFunc(c, mdb.ConfirmTransferedDevice)
}

/******************************************************************************
 * function: queryHeartRate
 * description: query heart rate status from heartrate table
 * return {*}
********************************************************************************/

// queryHeartRate godoc
//
//	@Summary	queryHeartRate
//	@Schemes
//	@Description	query heart data from heart rate table
//	@Tags			sleep device
//	@Produce		json
//
//	@Param			mac	query	string		true	"device mac address"
//
// @Param begin_day query string false "begin day, format yyyy-MM-dd"
// @Param end_day query string false "end day, format yyyy-MM-dd"
//
//	@Success		200		{object}	mysql.HeartRate
//	@Router			/device/queryHeartRate [get]
func queryHeartRate(c *gin.Context) {
	exception.TryEx{
		Try: func() {
			status, result := mdb.QueryHeartRate(c)
			respJSON(c, status, result)
		},
		Catch: func(e exception.Exception) {
			respJSON(c, e.Code, e.Msg)
		},
	}.Run()

}

// statsHeartRateByMinute godoc
//
//	@Summary	statsHeartRateByMinute
//	@Schemes
//	@Description	stats heart data per half of hour from heart rate table
//	@Tags			sleep device
//	@Produce		json
//
//	@Param			mac	query	string		true	"device mac address"
//
// @Param begin_day query string false "begin day, format yyyy-MM-dd"
// @Param end_day query string false "end day, format yyyy-MM-dd"
//
//	@Success		200		{object}	mdb.StatsHeartRate
//	@Router			/device/statsHeartRateByMinute [get]
func statsHeartRateByMinute(c *gin.Context) {
	exception.TryEx{
		Try: func() {
			status, result := mdb.StatsHeartRateByMinute(c)
			respJSON(c, status, result)
		},
		Catch: func(e exception.Exception) {
			respJSON(c, http.StatusBadRequest, e.Msg)
		},
	}.Run()

}

/******************************************************************************
 * function: querySleepReport
 * description:
 * return {*}
********************************************************************************/

// querySleepReport godoc
//
//	@Summary	querySleepReport
//	@Schemes
//	@Description	query sleep report data
//	@Tags			sleep device
//	@Produce		json
//
//	@Param			mac	query	string		true	"device mac address"
//
// @Param begin_day query string false "begin day, format yyyy-MM-dd"
// @Param end_day query string false "end day, format yyyy-MM-dd"
//
//	@Success		200		{object}	mysql.SleepReport
//	@Router			/device/querySleepReport [get]
func querySleepReport(c *gin.Context) {
	exception.TryEx{
		Try: func() {
			status, result := mdb.QuerySleepReport(c)
			respJSON(c, status, result)
		},
		Catch: func(e exception.Exception) {
			respJSON(c, e.Code, e.Msg)
		},
	}.Run()

}

/******************************************************************************
 * function: queryFallCheckStatus
 * description:
 * return {*}
********************************************************************************/

// queryFallCheckStatus godoc
//
//	@Summary	queryFallCheckStatus
//	@Schemes
//	@Description	query fall check data
//	@Tags			fall check device
//	@Produce		json
//
//	@Param			mac	query	string		true	"device mac address"
//
//	@Success		200		{object}	mysql.FallCheck
//	@Router			/device/queryFallCheckStatus [get]
func queryFallCheckStatus(c *gin.Context) {
	exception.TryEx{
		Try: func() {
			status, result := mdb.QueryFallCheckStatus(c)
			respJSON(c, status, result)
		},
		Catch: func(e exception.Exception) {
			respJSON(c, e.Code, e.Msg)
		},
	}.Run()

}

/******************************************************************************
 * function: queryAlarmRecord
 * description:
 * return {*}
********************************************************************************/

// queryAlarmRecord godoc
//
//	@Summary	queryAlarmRecord
//	@Schemes
//	@Description	query alarm record report from fall check device
//	@Tags			fall check device
//	@Produce		json
//
//	@Param			mac	query	string		true	"device mac address"
//
// @Param begin_day query string false "begin day, format yyyy-MM-dd"
// @Param end_day query string false "end day, format yyyy-MM-dd"
//
//	@Success		200		{object}	mysql.FallAlarm
//	@Router			/device/queryAlarmRecord [get]
func queryAlarmRecord(c *gin.Context) {
	exception.TryEx{
		Try: func() {
			status, result := mdb.QueryAlarmRecord(c)
			respJSON(c, status, result)
		},
		Catch: func(e exception.Exception) {
			respJSON(c, e.Code, e.Msg)
		},
	}.Run()
}

/******************************************************************************
 * function: queryFallExistRecord
 * description:
 * return {*}
********************************************************************************/

// queryFallExistRecord godoc
//
//	@Summary	queryFallExistRecord
//	@Schemes
//	@Description	query fall check data if has person exist
//	@Tags			fall check device
//	@Produce		json
//
//	@Param			mac	query	string		true	"device mac address"
//
// @Param begin_day query string false "begin day, format yyyy-MM-dd"
// @Param end_day query string false "end day, format yyyy-MM-dd"
//
//	@Success		200		{object}	mysql.FallCheck
//	@Router			/device/queryFallExistRecord [get]
func queryFallExistRecord(c *gin.Context) {
	exception.TryEx{
		Try: func() {
			status, result := mdb.QueryFallExistRecord(c)
			respJSON(c, status, result)
		},
		Catch: func(e exception.Exception) {
			respJSON(c, e.Code, e.Msg)
		},
	}.Run()
}

/******************************************************************************
 * function: queryFallParams
 * description:
 * return {*}
********************************************************************************/

// queryFallParams godoc
//
//	@Summary	queryFallParams
//	@Schemes
//	@Description	query fall check device install params
//	@Tags			fall check device
//	@Produce		json
//
//	@Param			device_id	query	int		true	"device id"
//
//	@Success		200		{object}	mysql.FallParams
//	@Router			/device/queryFallParams [get]
func queryFallParams(c *gin.Context) {
	exception.TryEx{
		Try: func() {
			status, result := mdb.QueryFallParams(c)
			respJSON(c, status, result)
		},
		Catch: func(e exception.Exception) {
			respJSON(c, e.Code, e.Msg)
		},
	}.Run()
}

/******************************************************************************
 * function: insertFallParams
 * description:
 * return {*}
********************************************************************************/

// insertFallParams godoc
//
//	@Summary	insertFallParams
//	@Schemes
//	@Description	query fall check device install params
//	@Tags			fall check device
//	@Produce		json
//	@Param			token	query	string		false	"token"
//	@Param			in	body	mysql.FallParams		true	"fall check device params"
//
//	@Success		200		{object}	mysql.FallParams
//	@Router			/device/insertFallParams [post]
func insertFallParams(c *gin.Context) {
	exception.TryEx{
		Try: func() {
			status, result := mdb.InsertFallParams(c)
			respJSON(c, status, result)
		},
		Catch: func(e exception.Exception) {
			respJSON(c, e.Code, e.Msg)
		},
	}.Run()
}

/******************************************************************************
 * function: openLampRealData
 * description:
 * return {*}
********************************************************************************/

// openLampRealData godoc
//
//	@Summary	openLampRealData
//	@Schemes
//	@Description	query fall check device install params
//	@Tags			lamp device
//	@Produce		json
//	@Param			token	query	string		false	"token"
//	@Param			in	body	mdb.LampRealDataReq		true	"lamp real data"
//
//	@Success		200		{none} {none}
//	@Router			/device/openLampRealData [post]
func openLampRealData(c *gin.Context) {
	exception.TryEx{
		Try: func() {
			status, result := mdb.OpenLampRealData(c)
			respJSON(c, status, result)
		},
		Catch: func(e exception.Exception) {
			respJSON(c, e.Code, e.Msg)
		},
	}.Run()
}

// queryLampRealData godoc
//
//	@Summary	queryLampRealData
//	@Schemes
//	@Description	query lamp device real-data by mac and time
//	@Tags			lamp device
//	@Produce		json
//
//	@Param			mac			query	string	true	"mac address"
//	@Param			begin_day	query	string	false	"begin day"
//	@Param			end_day		query	string	false	"end day"
//
//	@Success		200			{object} mysql.RealDataSql
//	@Router			/device/queryLampRealData [get]
func queryLampRealData(c *gin.Context) {
	exception.TryEx{
		Try: func() {
			status, result := mdb.QueryLampRealData(c)
			respJSON(c, status, result)
		},
		Catch: func(e exception.Exception) {
			respJSON(c, http.StatusBadRequest, e.Msg)
		},
	}.Run()
}

// queryLampReportStatus godoc
//
//	@Summary	queryLampReportStatus
//	@Schemes
//	@Description	query lamp device report by mac and date
//	@Tags			lamp device
//	@Produce		json
//
//	@Param			mac			query	string	true	"mac address"
//	@Param			begin_day	query	string	false	"begin day"
//	@Param			end_day		query	string	false	"end day"
//
//	@Success		200			{object} mysql.LampReportStatus
//	@Router			/device/queryLampReportStatus [get]
func queryLampReportStatus(c *gin.Context) {
	apiCommonFunc(c, mdb.QueryLampReportStatus)
}

// queryLampEvent godoc
//
//	@Summary	queryLampEvent
//	@Schemes
//	@Description	query lamp device real-data by mac and time
//	@Tags			lamp device
//	@Produce		json
//
//	@Param			mac			query	string	true	"mac address"
//	@Param			begin_day	query	string	false	"begin day"
//	@Param			end_day		query	string	false	"end day"
//
//	@Success		200			{object}	mysql.EventReportSql
//	@Router			/device/queryLampEvent [get]
func queryLampEvent(c *gin.Context) {
	exception.TryEx{
		Try: func() {
			status, result := mdb.QueryLampEvent(c)
			respJSON(c, status, result)
		},
		Catch: func(e exception.Exception) {
			respJSON(c, http.StatusBadRequest, e.Msg)
		},
	}.Run()
}

// controlLamp godoc
//
//	@Summary	controlLamp
//	@Schemes
//	@Description	control lamp status
//	@Tags			lamp device
//	@Produce		json
//	@Param			token	query	string		false	"token"
//	@Param			in	body	mdb.ControlLampReq	true	"control lamp"
//
// @Success	200	{none}	{none}
// @Router		/device/controlLamp [post]
func controlLamp(c *gin.Context) {
	exception.TryEx{
		Try: func() {
			status, result := mdb.ControlLamp(c)
			respJSON(c, status, result)
		},
		Catch: func(e exception.Exception) {
			respJSON(c, http.StatusBadRequest, e.Msg)
		},
	}.Run()
}

// queryLampControl godoc
//
//	@Summary	queryLampControl
//	@Schemes
//	@Description	query lamp control data by mac
//	@Tags			lamp device
//	@Produce		json
//
//	@Param			mac			query	string	true	"mac address"
//
//	@Success		200			{object}	mysql.LampControlSql
//	@Router			/device/queryLampControl [get]
func queryLampControl(c *gin.Context) {
	exception.TryEx{
		Try: func() {
			status, result := mdb.QueryLampControl(c)
			respJSON(c, status, result)
		},
		Catch: func(e exception.Exception) {
			respJSON(c, http.StatusBadRequest, e.Msg)
		},
	}.Run()
}

// createStudyRoom godoc
//
//	@Summary	createStudyRoom
//	@Schemes
//	@Description	create study room
//	@Tags			room
//	@Produce		json
//	@Param			token	query	string		false	"token"
//	@Param			in	body	mdb.CreateStudyRoomReq	true	"room information"
//
//	@Success		200	{object}	mysql.StudyRoom
//	@Router			/device/createStudyRoom [post]
func createStudyRoom(c *gin.Context) {
	exception.TryEx{
		Try: func() {
			status, result := mdb.CreateStudyRoom(c)
			respJSON(c, status, result)
		},
		Catch: func(e exception.Exception) {
			respJSON(c, http.StatusBadRequest, e.Msg)
		},
	}.Run()
}

// modifyStudyRoom godoc
//
//	@Summary	modifyStudyRoom
//	@Schemes
//	@Description	修改自习室信息
//	@Tags			room
//	@Produce		json
//	@Param			token	query	string		false	"token"
//	@Param			in	body	mdb.ModifyStudyRoomReq	true	"room information"
//
//	@Success		200	 {string}	{"修改成功"}
//	@Router			/device/modifyStudyRoom [post]
func modifyStudyRoom(c *gin.Context) {
	exception.TryEx{
		Try: func() {
			status, result := mdb.ModifyStudyRoom(c)
			respJSON(c, status, result)
		},
		Catch: func(e exception.Exception) {
			respJSON(c, http.StatusBadRequest, e.Msg)
		},
	}.Run()

}

// releaseStudyRoom godoc
//
//	@Summary	releaseStudyRoom
//	@Schemes
//	@Description	release study room
//	@Tags			room
//	@Produce		json
//	@Param			token	query	string		false	"token"
//	@Param			in		body	mdb.ReleaseStudyRoomReq		true	"release room"
//
//	@Success		200			{string}	{"release study room success"}
//	@Router			/device/releaseStudyRoom [post]
func releaseStudyRoom(c *gin.Context) {
	exception.TryEx{
		Try: func() {
			status, result := mdb.ReleaseStudyRoom(c)
			respJSON(c, status, result)
		},
		Catch: func(e exception.Exception) {
			respJSON(c, http.StatusBadRequest, e.Msg)
		},
	}.Run()
}

// queryStudyRoom godoc
//
//	@Summary	queryStudyRoom
//	@Schemes
//	@Description	query study room by create user id
//	@Tags			room
//	@Produce		json
//
//	@Param			user_id	query	int		true	"user id in study room"
//
//	@Success		200			{object}	mysql.StudyRoom
//	@Router			/device/queryStudyRoom [get]
func queryStudyRoom(c *gin.Context) {
	apiCommonFunc(c, mdb.QueryStudyRoom)
}

// queryInviteStudyRoom godoc
//
//	@Summary	queryInviteStudyRoom
//	@Schemes
//	@Description	query invited study room by create user
//	@Tags			room
//	@Produce		json
//
//	@Param			user_id	query	int		true	"user id"
//
//	@Success		200			{object} mysql.StudyRoomUserDetail
//	@Router			/device/queryInviteStudyRoom [get]
func queryInviteStudyRoom(c *gin.Context) {
	exception.TryEx{
		Try: func() {
			status, result := mdb.QueryInviteStudyRoom(c)
			respJSON(c, status, result)
		},
		Catch: func(e exception.Exception) {
			respJSON(c, http.StatusBadRequest, e.Msg)
		},
	}.Run()
}

// queryRankingByStudyRoom godoc
//
//	@Summary	queryRankingByStudyRoom
//	@Schemes
//	@Description	查询自习室排行数据，查询条件自习室id
//	@Tags			room
//	@Produce		json
//
//	@Param			room_id	query	int		true	"room id"
//
//	@Success		200	{object}	mysql.StudyRoomRanking
//	@Router			/device/queryRankingByStudyRoom [get]
func queryRankingByStudyRoom(c *gin.Context) {
	exception.TryEx{
		Try: func() {
			status, result := mdb.QueryRankingByStudyRoom(c)
			respJSON(c, status, result)
		},
		Catch: func(e exception.Exception) {
			respJSON(c, http.StatusBadRequest, e.Msg)
		},
	}.Run()
}

// statsLampFlowData godoc
//
//	@Summary	statsLampFlowData
//	@Schemes
//	@Description	统计台灯的心流数据，根据日期返回心流的比率数据
//	@Tags			lamp device
//	@Produce		json
//
//	@Param			mac	query	string		true	"mac address"
//
// @Param			start_time	query	string	true	"start time"
// @Param			end_time	query	string	true	"end time"
//
//	@Success		200			{int}	{int}
//	@Router			/device/statsLampFlowData [get]
func statsLampFlowData(c *gin.Context) {
	exception.TryEx{
		Try: func() {
			status, result := mdb.StatsLampFlowData(c)
			respJSON(c, status, result)
		},
		Catch: func(e exception.Exception) {
			respJSON(c, http.StatusBadRequest, e.Msg)

		},
	}.Run()
}

// queryDateListInReport godoc
//
//	@Summary	queryDateListInReport
//	@Schemes
//	@Description	查詢报表中含有数据的日期列表
//	@Tags			device
//	@Produce		json
//
//	@Param			mac	query	string		true	"mac address"
//
// @Param begin_day query string false "begin day, format yyyy-MM-dd"
// @Param end_day query string false "end day, format yyyy-MM-dd"
//
//	@Success		200	{object}	mdb.QueryDateListResp
//	@Router			/device/queryDateListInReport [get]
func queryDateListInReport(c *gin.Context) {
	exception.TryEx{
		Try: func() {
			status, result := mdb.QueryDateListInReport(c)
			respJSON(c, status, result)
		},
		Catch: func(e exception.Exception) {
			respJSON(c, http.StatusBadRequest, e.Msg)
		},
	}.Run()

}

// queryDeviceOverview godoc
//
//	@Summary	queryDeviceOverview
//	@Schemes
//	@Description	查询设备的概览数据
//	@Tags			device
//	@Produce		json
//
//	@Param			mac	query	string		true	"mac address"
//
//	@Success		200	{object}	mysql.DeviceOverview
//	@Router			/device/queryDeviceOverview [get]
func queryDeviceOverview(c *gin.Context) {
	apiCommonFunc(c, mdb.QueryDeviceOverview)
}

// updateDeviceOverview godoc
//
//	@Summary	updateDeviceOverview
//	@Schemes
//	@Description	更新设备概况
//	@Tags			device
//	@Produce		json
//	@Param			token	query	string		false	"token"
//	@Param			in		body	mdb.DeviceOverviewReq		true	"概况数据"
//
//	@Success		200	{object}	mysql.DeviceOverview
//	@Router			/device/updateDeviceOverview [post]
func updateDeviceOverview(c *gin.Context) {
	apiCommonFunc(c, mdb.UpdateDeviceOverview)
}
