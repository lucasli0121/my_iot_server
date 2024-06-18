/******************************************************************************
 * Author: liguoqiang
 * Date: 2023-11-20 11:58:18
 * LastEditors: liguoqiang
 * LastEditTime: 2024-06-02 19:28:47
 * Description:
********************************************************************************/
package mdb

import (
	"hjyserver/exception"
	"hjyserver/mdb/common"
	"hjyserver/mdb/mysql"
	"net/http"

	"github.com/gin-gonic/gin"
)

// swagger:model AskX1RealDataReq
type AskX1RealDataReq struct {
	// 设备mac地址
	// required: true
	Mac string `json:"mac"`
	// 发送频率
	// required: false
	Freq int `json:"freq"`
	// 是否保持推送
	// required: true
	// enum: 0,1
	KeepPush int `json:"keep_push"`
}

func AskX1RealData(c *gin.Context) (int, interface{}) {
	var req *AskX1RealDataReq = &AskX1RealDataReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		exception.Throw(http.StatusAccepted, err.Error())
	}
	if req.Mac == "" {
		return http.StatusBadRequest, "device mac required"
	}
	if req.Freq == 0 {
		req.Freq = 6
	}
	mysql.AskX1RealData(req.Mac, req.Freq, req.KeepPush)
	return http.StatusOK, nil
}

// swagger:model CleanX1EventReq
type CleanX1EventReq struct {
	// 设备mac地址
	// required: true
	Mac string `json:"mac"`
}

func CleanX1Event(c *gin.Context) (int, interface{}) {
	var req *CleanX1EventReq = &CleanX1EventReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		exception.Throw(http.StatusAccepted, err.Error())
	}
	if req.Mac == "" {
		return http.StatusBadRequest, "device mac required"
	}
	mysql.CleanX1Event(req.Mac)
	return http.StatusOK, nil
}

// swagger:model SleepX1SwitchReq
type SleepX1SwitchReq struct {
	// 设备mac地址
	// required: true
	Mac string `json:"mac"`
	// 开关状态
	// required: true
	// enum: 0,1
	Switch int `json:"switch"`
}

func SleepX1Switch(c *gin.Context) (int, interface{}) {
	var req *SleepX1SwitchReq = &SleepX1SwitchReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		exception.Throw(http.StatusAccepted, err.Error())
	}
	if req.Mac == "" {
		return http.StatusBadRequest, "device mac required"
	}
	mysql.SleepX1Switch(req.Mac, req.Switch)
	return http.StatusOK, nil
}

/******************************************************************************
 * function:
 * description:
 * param {*gin.Context} c
 * return {*}
********************************************************************************/
// swagger:model ImproveDisturbedReq
type ImproveDisturbedReq struct {
	// 设备mac地址
	// required: true
	Mac string `json:"mac"`
	// 开关状态
	// required: true
	// enum: 0,1
	Switch int `json:"switch"`
}

func ImproveDisturbed(c *gin.Context) (int, interface{}) {
	var req *ImproveDisturbedReq = &ImproveDisturbedReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		exception.Throw(http.StatusAccepted, err.Error())
	}
	if req.Mac == "" {
		return common.ParamError, "device mac required"
	}
	mysql.ImproveDisturbedX1Switch(req.Mac, req.Switch)
	return http.StatusOK, nil
}

func QueryX1RealDataJson(c *gin.Context) (int, interface{}) {
	mac := c.Query("mac")
	if mac == "" {
		return http.StatusBadRequest, "device mac required"
	}
	createDate := c.Query("create_date")
	var glist []mysql.X1RealDataOrigin
	mysql.QueryX1RealDataJson(mac, createDate, &glist)
	return common.Success, glist
}

func QueryX1SleepReportJson(c *gin.Context) (int, interface{}) {
	mac := c.Query("mac")
	if mac == "" {
		return http.StatusBadRequest, "device mac required"
	}
	createDate := c.Query("create_date")
	var glist []mysql.X1DayReportOrigin
	mysql.QueryX1DayReportJson(mac, createDate, &glist)
	return common.Success, glist
}
