/******************************************************************************
 * Author: liguoqiang
 * Date: 2023-09-06 17:50:12
 * LastEditors: liguoqiang
 * LastEditTime: 2024-01-31 15:31:04
 * Description:
********************************************************************************/
package mdb

import (
	"fmt"
	"hjyserver/mdb/common"
	"hjyserver/mdb/mysql"
	"hjyserver/mq"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// swagger:model NewDeviceReq
type NewDeviceReq struct {
	UserId int64 `json:"user_id" mysql:"user_id" `
	// required: false
	// flag 0:自己创建 1:共享
	Flag int `json:"flag" mysql:"flag"`
	// required: true
	// 设备名称
	Name string `json:"name" mysql:"name" `
	// 设备类型
	// required: false
	// enum: heart_rate,fall_check,lamp_type
	Type string `json:"type" mysql:"type" `
	// 设备mac地址
	// required: true
	Mac string `json:"mac" mysql:"mac"`
}

/******************************************************************************
 * function: insertDevice
 * description: insert device information into database
 * return {*}
********************************************************************************/
func InsertDevice(c *gin.Context) (int, interface{}) {
	body := mysql.NewUserDevice()
	body.DecodeFromGin(c)
	if body.Name == "" {
		return http.StatusAccepted, "device name required!"
	}
	if body.Mac == "" {
		return http.StatusAccepted, "device mac required!"
	}
	if body.Type == "" {
		body.Type = mysql.GetDeviceTypeByName(body.Name)
	}
	if body.OnlineTime == "" {
		body.OnlineTime = common.GetNowTime()
	}
	if body.CreateTime == "" {
		body.CreateTime = common.GetNowTime()
	}
	var gList []mysql.Device
	filter := fmt.Sprintf("mac = '%s'", body.Mac)
	mysql.QueryDeviceByCond(filter, nil, nil, &gList)
	var ok bool = true
	if len(gList) == 0 {
		ok = body.Insert()
	} else {
		body.ID = gList[0].ID
		ok = body.Update()
	}
	if ok {
		// 添加用户和设备的关联关系
		filter = fmt.Sprintf("device_id=%d and flag=%d", body.ID, common.NormalDeviceFlag)
		var gList []mysql.UserDeviceRelation
		mysql.QueryUserDeviceRelationByCond(filter, nil, nil, &gList)
		if len(gList) > 0 {
			return http.StatusAlreadyReported, "device already exist and not been insert"
		}
		// subscribe topic
		if body.Type == mysql.LampType {
			mq.SubscribeTopic(mysql.MakeHl77DeliverTopicByMac(body.Mac), mysql.NewLampMqttMsgProc())
			mysql.AskHl77RealData(body.Mac, 6, 1)
		} else if body.Type == mysql.Ed713Type {
			mysql.SubscribeEd713MqttTopic(body.Mac)
			mysql.AskEd713RealData(body.Mac, 6, 1)
		} else if body.Type == mysql.X1Type {
			mysql.SubscribeX1MqttTopic(body.Mac)
			mysql.AskEd713RealData(body.Mac, 6, 1)
		}
		userDevice := mysql.NewUserDeviceRelation()
		userDevice.UserId = body.UserId
		userDevice.DeviceId = body.ID
		userDevice.Flag = common.NormalDeviceFlag // 默认为0
		userDevice.Insert()
		body.Flag = common.NormalDeviceFlag
		return http.StatusOK, body
	} else {
		return http.StatusAccepted, "insert error!"
	}
}

// swagger:model UpdateDeviceReq
type UpdateDeviceReq struct {
	ID   int64  `json:"id" mysql:"id" binding:"omitempty"`
	Name string `json:"name" mysql:"name" `
	// 设备类型
	// required: false
	// enum: heart_rate,fall_check,lamp_type
	Type string `json:"type" mysql:"type" `
	// 设备mac地址
	// required: true
	Mac string `json:"mac" mysql:"mac"`
	// required: false
	// 设备备注
	Remark string `json:"remark" mysql:"remark"`
}

/******************************************************************************
 * function: UpdateDevice
 * description:
 * return {*}
********************************************************************************/
func UpdateDevice(c *gin.Context) (int, interface{}) {
	me := mysql.NewDevice()
	me.DecodeFromGin(c)
	if me.Type == "" {
		me.Type = mysql.GetDeviceTypeByName(me.Name)
	}
	if me.Update() {
		return http.StatusOK, me
	}
	return http.StatusAccepted, "update failed!"
}

/******************************************************************************
 * function: QueryDeviceById
 * description:
 * return {*}
********************************************************************************/
func QueryDeviceById(id int64) (int, interface{}) {
	me := mysql.NewDevice()
	if me.QueryByID(id) {
		return http.StatusOK, me
	}
	return http.StatusAccepted, "query failed"
}

/******************************************************************************
 * function:
 * description:
 * param {string} mac
 * return {*}
********************************************************************************/

// swagger:model DeviceBindResp
type DeviceBindResp struct {
	// required: true
	Mac string `json:"mac" mysql:"mac"`
	// required: true
	Bind bool `json:"bind" mysql:"bind"`
}

func QueryBindDeviceByMac(userId int64, mac string) (int, interface{}) {
	filter := fmt.Sprintf("mac='%s'", mac)
	var gList []mysql.Device
	mysql.QueryDeviceByCond(filter, nil, nil, &gList)
	if len(gList) > 0 {
		var device = &gList[0]
		filter = fmt.Sprintf("device_id=%d", device.ID)
		var vList []mysql.UserDeviceRelation
		obj := &DeviceBindResp{}
		obj.Mac = mac
		mysql.QueryUserDeviceRelationByCond(filter, nil, nil, &vList)
		if len(vList) > 0 {
			obj.Bind = true
		} else {
			obj.Bind = false
		}
		return http.StatusOK, obj
	}
	return http.StatusAccepted, "query failed"
}

/***********************************************************
* QueryDeviceByUser
* 根据device information
***********************************************************/
func QueryDeviceByUser(c *gin.Context) (int, interface{}) {
	userId := c.Query("user_id")
	if userId == "" {
		return http.StatusBadRequest, "user id required"
	}
	flag := c.Query("flag")
	mUserId, _ := strconv.ParseInt(userId, 10, 64)
	mFlag, err := strconv.ParseInt(flag, 10, 32)
	if err != nil {
		mFlag = -1
	}
	var gList []mysql.UserDevice
	mysql.QueryUserDeviceByUserId(mUserId, int(mFlag), &gList)
	return http.StatusOK, gList
}

// swagger:model ShareDeviceReq
type ShareDeviceReq struct {
	UserId int64 `json:"user_id" mysql:"user_id" `
	// required: true
	// flag 0:自己创建 1:共享
	Flag int `json:"flag" mysql:"flag"`
	// required: true
	// 设备ID
	ID int64 `json:"id" mysql:"id" `
}

/******************************************************************************
 * function: ShareDeviceToOtherUser
 * description:
 * return {*}
********************************************************************************/
func ShareDeviceToOtherUser(c *gin.Context) (int, interface{}) {
	body := mysql.NewUserDevice()
	body.DecodeFromGin(c)
	if body.ID == 0 || body.UserId == 0 {
		return http.StatusBadRequest, "device id and user id required"
	}
	if !body.QueryByID(body.ID) {
		return http.StatusBadRequest, "device not exist"
	}
	userDevice := mysql.NewUserDeviceRelation()
	userDevice.UserId = body.UserId
	userDevice.DeviceId = body.ID
	userDevice.Flag = common.ShareDeviceFlag // share device is 1
	body.Flag = common.ShareDeviceFlag
	filter := fmt.Sprintf("user_id = %d and device_id=%d", body.UserId, body.ID)
	var gList []mysql.UserDeviceRelation
	mysql.QueryUserDeviceRelationByCond(filter, nil, nil, &gList)
	if len(gList) > 0 {
		return http.StatusOK, body
	}
	if userDevice.Insert() {
		mq.PublishData(common.MakeShareDeviceNotifyTopic(body.UserId), body)
		return http.StatusOK, body
	}
	return http.StatusAccepted, "insert failed!"
}

/******************************************************************************
 * function: RemoveUserDevice
 * description:
 * return {*}
********************************************************************************/
func RemoveUserDevice(c *gin.Context) (int, interface{}) {
	var userDevice = mysql.NewUserDeviceRelation()
	userDevice.DecodeFromGin(c)
	if userDevice.DeviceId <= 0 {
		return http.StatusBadRequest, "device id need!"
	}
	if userDevice.DeleteWithUser() {
		return http.StatusOK, "remove user device ok"
	}
	return http.StatusAccepted, "remove error"
}
