/******************************************************************************
 * Author: liguoqiang
 * Date: 2023-11-20 11:58:18
 * LastEditors: liguoqiang
 * LastEditTime: 2023-12-15 16:54:57
 * Description:
********************************************************************************/
package mdb

import (
	"hjyserver/exception"
	"hjyserver/mdb/mysql"
	"net/http"

	"github.com/gin-gonic/gin"
)

// swagger:model AskEd713RealDataReq
type AskEd713RealDataReq struct {
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

func AskEd713RealData(c *gin.Context) (int, interface{}) {
	var req *AskEd713RealDataReq = &AskEd713RealDataReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		exception.Throw(http.StatusAccepted, err.Error())
	}
	if req.Mac == "" {
		return http.StatusBadRequest, "device mac required"
	}
	if req.Freq == 0 {
		req.Freq = 6
	}
	mysql.AskEd713RealData(req.Mac, req.Freq, req.KeepPush)
	return http.StatusOK, nil
}
