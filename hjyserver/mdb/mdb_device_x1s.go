/******************************************************************************
 * Author: liguoqiang
 * Date: 2024-08-27 18:47:25
 * LastEditors: liguoqiang
 * LastEditTime: 2024-12-13 11:27:54
 * Description:
********************************************************************************/
package mdb

import (
	mylog "hjyserver/log"
	"hjyserver/mdb/common"
	"hjyserver/mdb/mysql"
	"strings"

	"github.com/gin-gonic/gin"
)

const x1sTag = "mdb_device_x1s"

func X1sMdbInit() {
	mylog.Log.Debugln(x1sTag, "begin call X1sMdbInit...")
}
func X1sMdbUnini() {
	mylog.Log.Debugln(x1sTag, "begin call X1sMdbUnini...")
}

func AskX1sSyncVersion(c *gin.Context) (int, interface{}) {
	mac := c.Query("mac")
	if mac == "" {
		return common.ParamError, "mac required!"
	}
	// deviceList := make([]mysql.Device, 0)
	// filter := fmt.Sprintf("mac = '%s'", mac)
	// mysql.QueryDeviceByCond(filter, nil, nil, &deviceList)
	// if len(deviceList) == 0 {
	// 	return common.NoExist, "device is not exist!"
	// }
	// device := deviceList[0]
	// if device.Type != mysql.X1sType {
	// 	return common.TypeError, "device's type is not " + mysql.X1sType
	// }
	// if device.Online == 0 {
	// 	return common.DeviceOffLine, "device is offline!"
	// }

	mysql.X1sSyncRequest(mac, 1)
	return common.Success, "ok"
}

/******************************************************************************
 * function: InsertX1sWhiteList
 * description: 插入X1s白名单
 * param {*gin.Context} c
 * return {*}
********************************************************************************/
//swagger:model X1sWhiteListReq
type X1sWhiteListReq struct {
	Mac string `json:"mac"`
}

func InsertX1sWhiteList(c *gin.Context) (int, interface{}) {
	var req X1sWhiteListReq
	err := c.ShouldBindJSON(&req)
	if err != nil {
		return common.ParamError, "param error!"
	}
	if req.Mac == "" {
		return common.ParamError, "mac required!"
	}
	macs := strings.Split(req.Mac, ";")
	for _, mac := range macs {
		if !mysql.X1sCheckMacInOtaWhiteList(mac) {
			whiteList := mysql.NewX1sOtaWhiteList()
			whiteList.Mac = mac
			whiteList.Insert()
		}
	}
	return common.Success, "ok"
}

func QueryX1sWhiteList(c *gin.Context) (int, interface{}) {
	whiteList := make([]mysql.X1sOtaWhiteList, 0)
	mysql.QueryX1sOtaWhiteList(&whiteList)
	return common.Success, whiteList
}
