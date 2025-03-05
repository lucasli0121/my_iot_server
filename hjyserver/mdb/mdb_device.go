/******************************************************************************
 * Author: liguoqiang
 * Date: 2023-09-06 17:50:12
 * LastEditors: liguoqiang
 * LastEditTime: 2025-01-07 23:28:14
 * Description:
********************************************************************************/
package mdb

import (
	"fmt"
	"hjyserver/cfg"
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
	// 设备类型包括: heart_rate, fall_check, lamp_type, ed713_type, ed719_type, x1_type, H03pro
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
		return common.ParamError, "device name required!"
	}
	if body.Mac == "" {
		return common.ParamError, "device mac required!"
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
			return common.HasExist, "device already exist and not been insert"
		}
		// subscribe topic
		if body.Type == mysql.LampType {
			if cfg.This.Svr.EnableHl77 {
				mq.SubscribeTopic(mysql.MakeHl77DeliverTopicByMac(body.Mac), mysql.NewLampMqttMsgProc())
				mysql.AskHl77RealData(body.Mac, 6, 1)
			}
		} else if body.Type == mysql.Ed713Type {
			if cfg.This.Svr.EnableEd713 {
				mysql.SubscribeEd713MqttTopic(body.Mac)
				mysql.AskEd713RealData(body.Mac, 6, 1)
			}
		} else if body.Type == mysql.X1Type {
			if cfg.This.Svr.EnableX1 {
				mysql.SubscribeX1MqttTopic(body.Mac)
				mysql.AskEd713RealData(body.Mac, 6, 1)
			}
		} else if body.Type == mysql.X1sType {
			if cfg.This.Svr.EnableX1s {
				mysql.SubscribeX1sMqttTopic(body.Mac)
			}
		} else if body.Type == mysql.H03Type {
			if cfg.This.Svr.EnableH03 {
				mysql.SubscribeH03MqttTopic(body.Mac)
			}
		}
		userDevice := mysql.NewUserDeviceRelation()
		userDevice.UserId = body.UserId
		userDevice.DeviceId = body.ID
		userDevice.Flag = common.NormalDeviceFlag // 默认为0
		if userDevice.Insert() {
			body.Flag = common.NormalDeviceFlag
			return common.Success, body
		}
	}
	return common.DBError, "insert error!"
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
	req := make(map[string]interface{}, 0)
	err := c.ShouldBindJSON(&req)
	if err != nil {
		return common.JsonError, "json format error"
	}
	if _, ok := req["id"]; !ok {
		return common.ParamError, "device id required"
	}
	id, _ := req["id"].(float64)
	me := mysql.NewDevice()
	if !me.QueryByID(int64(id)) {
		return common.NoExist, "device not exist"
	}
	if me.Type == "" {
		me.Type = mysql.GetDeviceTypeByName(me.Name)
	}
	for k, v := range req {
		switch k {
		case "name":
			me.Name = v.(string)
		case "remark":
			me.Remark = v.(string)
		}
	}
	if me.Update() {
		return common.Success, me
	}
	return common.DBError, "update failed!"
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
	FromUserId int64 `json:"from_user_id" mysql:"from_user_id" `
	ToUserId   int64 `json:"to_user_id" mysql:"to_user_id" `
	// required: true
	// flag 0:自己创建 1:共享
	Flag int `json:"flag" mysql:"flag"`
	// required: true
	// 设备ID
	DeviceId int64 `json:"device_id" mysql:"device_id" `
	//备注
	Remark string `json:"remark"`
	// 是否等待确认
	WaitConfirm bool `json:"wait_confirm"`
}

/******************************************************************************
 * function: ShareDeviceToOtherUser
 * description:
 * return {*}
********************************************************************************/
func ShareDeviceToOtherUser(c *gin.Context) (int, interface{}) {
	req := ShareDeviceReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		return common.JsonError, "json format error"
	}
	if req.DeviceId == 0 || req.FromUserId == 0 {
		return common.ParamError, "device id and user id required"
	}
	if req.FromUserId == req.ToUserId {
		return common.SameUser, "can't share device to yourself"
	}
	// check device exist
	userDevice := mysql.NewUserDevice()
	userDevice.ID = req.DeviceId
	if !userDevice.QueryByID(userDevice.ID) {
		return common.NoData, "device not exist"
	}
	// 用户分享设备
	userShare := mysql.NewUserShareDevice()
	userShare.FromUserId = req.FromUserId
	userShare.ToUserId = req.ToUserId
	userShare.DeviceId = req.DeviceId
	userShare.Remark = req.Remark
	var userShareList []mysql.UserShareDevice
	mysql.QueryUserShareDevice(userShare.FromUserId, userShare.ToUserId, userShare.DeviceId, -1, &userShareList)
	if len(userShareList) > 0 {
		return common.RepeatData, "device already shared, don't need to share again"
	}
	if req.WaitConfirm {
		userShare.Confirm = common.DeviceUnConfirmFlag
	} else {
		userShare.Confirm = common.DeviceConfirmFlag
	}
	status, result := SaveConfirmShareDevice(userDevice, userShare)
	if req.WaitConfirm && status == common.Success {
		// 向对方发送共享设备请求，表示需要确认共享设备
		mq.PublishData(common.MakeShareDeviceConfirmTopic(userShare.ToUserId), userShare)
	}
	return status, result
}

/******************************************************************************
 * function: ShareDeviceWithMac
 * description: 共享设备给其他用户， Req参数包括分享目的用户的手机号码
 * return {*}
********************************************************************************/

// swagger:model ShareDeviceWithMacReq
type ShareDeviceWithMacReq struct {
	// required: true
	FromUserId int64 `json:"from_user_id" `
	// required: true
	ToUserPhone string `json:"to_user_phone" `
	// required: true
	// flag 0:自己创建 1:共享
	Flag int `json:"flag"`
	// required: true
	// 设备Mac
	Mac string `json:"mac"`
	//备注
	Remark string `json:"remark"`
	// 是否等待确认
	WaitConfirm bool `json:"wait_confirm"`
}

func ShareDeviceToPhoneWithMac(c *gin.Context) (int, interface{}) {
	req := ShareDeviceWithMacReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		return common.JsonError, "json format error"
	}
	if req.Mac == "" || req.FromUserId == 0 || req.ToUserPhone == "" {
		return common.ParamError, "request param error"
	}
	// 查询用户号码是否存在
	var userList []mysql.User
	filter := fmt.Sprintf("phone='%s'", req.ToUserPhone)
	mysql.QueryUserByCond(filter, nil, nil, &userList)
	if len(userList) == 0 {
		return common.NoData, "user phone not exist"
	}
	user := &userList[0]
	// 检查分享的用户和被分享用户是否同一人
	if user.ID == req.FromUserId {
		return common.SameUser, "can't share device to yourself"
	}
	// check device exist
	userDevice := mysql.NewUserDevice()
	filter = fmt.Sprintf("mac='%s'", req.Mac)
	var deviceList []mysql.Device
	mysql.QueryDeviceByCond(filter, nil, nil, &deviceList)
	if len(deviceList) == 0 {
		return common.NoData, "device not exist"
	}
	userDevice.Device = deviceList[0]
	// 用户分享设备
	userShare := mysql.NewUserShareDevice()
	userShare.FromUserId = req.FromUserId
	userShare.ToUserId = user.ID
	userShare.DeviceId = userDevice.ID
	userShare.Remark = req.Remark
	var userShareList []mysql.UserShareDevice
	mysql.QueryUserShareDevice(userShare.FromUserId, userShare.ToUserId, userShare.DeviceId, -1, &userShareList)
	if len(userShareList) > 0 {
		return common.RepeatData, "device already shared, don't need to share again"
	}
	if req.WaitConfirm {
		userShare.Confirm = common.DeviceUnConfirmFlag
	} else {
		userShare.Confirm = common.DeviceConfirmFlag
	}
	status, result := SaveConfirmShareDevice(userDevice, userShare)
	if status != common.Success {
		return status, result
	}
	if req.WaitConfirm && status == common.Success {
		mq.PublishData(common.MakeShareDeviceConfirmTopic(userShare.ToUserId), userShare)
	}
	return status, userShare
}

func SaveConfirmShareDevice(userDevice *mysql.UserDevice, userShare *mysql.UserShareDevice) (int, interface{}) {
	// 检查是否已经共享相同的设备, 没有则添加, 有则更新
	var userShareList []mysql.UserShareDevice
	mysql.QueryUserShareDevice(userShare.FromUserId, userShare.ToUserId, userShare.DeviceId, -1, &userShareList)
	if len(userShareList) == 0 {
		if !userShare.Insert() {
			return common.DBError, "insert error"
		}
	} else {
		userShare.ID = userShareList[0].ID
		if !userShare.Update() {
			return common.DBError, "update error"
		}
	}

	userDevice.UserId = userShare.ToUserId
	userDevice.Flag = common.ShareDeviceFlag
	// 如果用户已经确认了共享设备，则添加设备和用户的关系
	if userShare.Confirm == common.DeviceConfirmFlag {
		// 添加用户和设备的关联关系
		userDeviceRelation := mysql.NewUserDeviceRelation()
		userDeviceRelation.UserId = userShare.ToUserId
		userDeviceRelation.DeviceId = userShare.DeviceId
		userDeviceRelation.Flag = common.ShareDeviceFlag // share device is 1
		// check if the relation exist
		filter := fmt.Sprintf("user_id=%d and device_id=%d", userDeviceRelation.UserId, userDeviceRelation.DeviceId)
		var gList []mysql.UserDeviceRelation
		mysql.QueryUserDeviceRelationByCond(filter, nil, nil, &gList)
		if len(gList) > 0 {
			userDeviceRelation = &gList[0]
			// 如果已经存在用户和设备关系记录，并且用户和设备关系不是分享关系，则不再添加，返回错误
			if userDeviceRelation.Flag == common.NormalDeviceFlag {
				userShare.Delete()
				return common.AlreadyBind, "device has binded the user, not allow shared"
			}
		} else {
			if !userDeviceRelation.Insert() {
				return common.DBError, "insert error"
			}
		}
		mq.PublishData(common.MakeShareDeviceNotifyTopic(userDevice.UserId), userDevice)
	}
	return common.Success, userDevice
}

/******************************************************************************
 * function: QueryShareUsers
 * description:  查询设备分享给哪些用户
 * param {*gin.Context} c
 * return {*}
********************************************************************************/
func QueryShareUsers(c *gin.Context) (int, interface{}) {
	mac := c.Query("mac")
	if mac == "" {
		return common.ParamError, "device mac address required"
	}
	userId := c.Query("user_id")
	if userId == "" {
		return common.ParamError, "user id required"
	}
	filter := fmt.Sprintf("mac='%s'", mac)
	var deviceList []mysql.Device
	mysql.QueryDeviceByCond(filter, nil, nil, &deviceList)
	if len(deviceList) == 0 {
		return common.NoData, "device not exist"
	}
	device := deviceList[0]
	mUserId, _ := strconv.ParseInt(userId, 10, 64)
	var gList []mysql.UserShareDeviceDetail
	mysql.QueryUserShareDeviceDetail(mUserId, 0, device.ID, common.DeviceConfirmFlag, &gList)
	return http.StatusOK, gList
}

/******************************************************************************
 * function: QueryUnconfirmedShareDevices
 * description: 查询未确认的共享设备
 * param {*gin.Context} c
 * return {*}
********************************************************************************/
func QueryUnconfirmedShareDevices(c *gin.Context) (int, interface{}) {
	userId := c.Query("user_id")
	if userId == "" {
		return common.ParamError, "user id required"
	}
	mac := c.Query("mac")
	device := mysql.NewDevice()
	var deviceList []mysql.Device
	if mac != "" {
		filter := fmt.Sprintf("mac='%s'", mac)
		mysql.QueryDeviceByCond(filter, nil, nil, &deviceList)
		if len(deviceList) == 0 {
			return common.NoData, "device not exist"
		}
		device = &deviceList[0]
	}
	mUserId, _ := strconv.ParseInt(userId, 10, 64)
	var gList []mysql.UserShareDeviceDetail
	mysql.QueryUserShareDeviceDetail(0, mUserId, device.ID, common.DeviceUnConfirmFlag, &gList)
	if len(gList) == 0 {
		return common.NoData, "no unconfirmed share device"
	}
	return http.StatusOK, gList
}

/******************************************************************************
 * function: ConfirmSharedDevice
 * description: 确认共享设备
 * param {*gin.Context} c
 * return {*}
********************************************************************************/
// swagger:model ConfirmSharedReq
type ConfirmSharedReq struct {
	UserId   int64 `json:"user_id" mysql:"user_id" `
	DeviceId int64 `json:"device_id" mysql:"device_id" `
	// required: true
	// 是否确认共享设备, true:确认共享设备, false:拒绝共享设备
	Confirm bool `json:"confirm" mysql:"confirm" `
}

func ConfirmSharedDevice(c *gin.Context) (int, interface{}) {
	req := &ConfirmSharedReq{}
	err := c.ShouldBindJSON(&req)
	if err != nil {
		return common.JsonError, "json format error"
	}
	if req.UserId == 0 || req.DeviceId == 0 {
		return common.ParamError, "user id and device id required"
	}
	var userShareDevices []mysql.UserShareDevice
	mysql.QueryUserShareDevice(0, req.UserId, req.DeviceId, common.DeviceUnConfirmFlag, &userShareDevices)
	if len(userShareDevices) == 0 {
		return common.NoData, "no unconfirmed share device"
	}
	userShareDevice := &userShareDevices[0]
	if req.Confirm {
		userShareDevice.Confirm = common.DeviceConfirmFlag
		if userShareDevice.Update() {
			// 添加用户和设备的关联关系
			userDeviceRelation := mysql.NewUserDeviceRelation()
			userDeviceRelation.UserId = userShareDevice.ToUserId
			userDeviceRelation.DeviceId = userShareDevice.DeviceId
			userDeviceRelation.Flag = common.ShareDeviceFlag // share device is 1
			// check if the relation exist
			filter := fmt.Sprintf("user_id=%d and device_id=%d", userDeviceRelation.UserId, userDeviceRelation.DeviceId)
			var gList []mysql.UserDeviceRelation
			mysql.QueryUserDeviceRelationByCond(filter, nil, nil, &gList)
			if len(gList) > 0 {
				userDeviceRelation = &gList[0]
				if userDeviceRelation.Flag == common.NormalDeviceFlag {
					userShareDevice.Delete()
					return common.NoPermission, "device has binded the user, not allow shared"
				}
			} else {
				if !userDeviceRelation.Insert() {
					return common.DBError, "insert error"
				}
			}
			return common.Success, "Confirmed successfully"
		}
	} else {
		// 如果用户拒绝则删除共享表中的记录
		userShareDevice.Delete()
		// 如果已经存在用户和设备关系记录，也一起删除
		var userDeviceRelation = mysql.NewUserDeviceRelation()
		userDeviceRelation.UserId = userShareDevice.ToUserId
		userDeviceRelation.DeviceId = userShareDevice.DeviceId
		userDeviceRelation.Flag = common.ShareDeviceFlag
		userDeviceRelation.DeleteWithUser()
		return common.Success, "Unconfirmed successfully"
	}
	return common.DBError, "update failed!"
}

/******************************************************************************
 * function: RemoveSharedDevice
 * description: 删除共享设备
 * param {*gin.Context} c
 * return {*}
********************************************************************************/
// swagger:model RemoveSharedDeviceReq
type RemoveSharedDeviceReq struct {
	ShareId int64 `json:"share_id" mysql:"share_id" `
	// UserId   int64 `json:"user_id" mysql:"user_id" `
	// DeviceId int64 `json:"device_id" mysql:"device_id" `
}

func RemoveSharedDevice(c *gin.Context) (int, interface{}) {
	req := &RemoveSharedDeviceReq{}
	err := c.ShouldBindJSON(&req)
	if err != nil {
		return common.JsonError, "json format error"
	}
	if req.ShareId <= 0 {
		return common.ParamError, "share device id need!"
	}
	var userShareDevice = mysql.NewUserShareDevice()
	if !userShareDevice.QueryByID(req.ShareId) {
		return common.NoData, "share device not exist"
	}
	// 先删除用户和设备关系表
	var userDeviceRelation = mysql.NewUserDeviceRelation()
	userDeviceRelation.UserId = userShareDevice.ToUserId
	userDeviceRelation.DeviceId = userShareDevice.DeviceId
	userDeviceRelation.Flag = common.ShareDeviceFlag
	userDeviceRelation.DeleteWithUser()
	if userShareDevice.Delete() {
		mq.PublishData(common.MakeShareDeviceRemoveNotifyTopic(userShareDevice.ToUserId), userShareDevice)
		return common.Success, "remove shared device success"
	}
	return common.DBError, "remove shared device failed"
}

/******************************************************************************
 * function: ModifySharedDeviceRemark
 * description: 修改共享设备备注
 * param {*gin.Context} c
 * return {*}
********************************************************************************/
// swagger:model ModifySharedDeviceRemarkReq
type ModifySharedDeviceRemarkReq struct {
	ShareId int64  `json:"share_id" mysql:"share_id" `
	Remark  string `json:"remark" mysql:"remark" `
}

func ModifySharedDeviceRemark(c *gin.Context) (int, interface{}) {
	req := &ModifySharedDeviceRemarkReq{}
	err := c.ShouldBindJSON(&req)
	if err != nil {
		return common.JsonError, "json format error"
	}
	if req.ShareId == 0 || req.Remark == "" {
		return common.ParamError, "share device id and remark required"
	}
	var userShareDevice = mysql.NewUserShareDevice()
	if !userShareDevice.QueryByID(req.ShareId) {
		return common.NoData, "share device not exist"
	}
	userShareDevice.Remark = req.Remark
	if userShareDevice.Update() {
		return common.Success, userShareDevice
	}
	return common.DBError, "update failed!"
}

/******************************************************************************
 * function: RemoveUserDevice
 * description: 移除用户指定的设备
 * return {*}
********************************************************************************/
func RemoveUserDevice(c *gin.Context) (int, interface{}) {
	var userDeviceRelation = mysql.NewUserDeviceRelation()
	userDeviceRelation.DecodeFromGin(c)
	if userDeviceRelation.DeviceId <= 0 {
		return common.ParamError, "device id need!"
	}
	if userDeviceRelation.UserId <= 0 {
		return common.ParamError, "user id need!"
	}

	filter := fmt.Sprintf("device_id=%d and user_id=%d", userDeviceRelation.DeviceId, userDeviceRelation.UserId)
	userDeviceList := make([]mysql.UserDeviceRelation, 0)
	mysql.QueryUserDeviceRelationByCond(filter, nil, nil, &userDeviceList)
	if len(userDeviceList) == 0 {
		res := fmt.Sprintf("cann't find the relation between %s", filter)
		return common.NoData, res
	}
	userDeviceRelation = &userDeviceList[0]
	userDeviceRelation.DeleteWithUser()
	// 如果是原始设备，则删除相关的所有共享设备
	if userDeviceRelation.Flag == common.NormalDeviceFlag {
		// 删除此User已共享给其他用户的记录
		mysql.DeleteUserShareDeviceByUserId(userDeviceRelation.UserId, 0, userDeviceRelation.DeviceId)
		// 删除所有和此设备相关的已过户用户的记录
		mysql.DeleteUserTransferDeviceByUserId(0, 0, userDeviceRelation.DeviceId)
	} else {
		// 删除共享给UserId的记录
		mysql.DeleteUserShareDeviceByUserId(0, userDeviceRelation.UserId, userDeviceRelation.DeviceId)
	}
	// 设置设备概况数据为不可见
	var device = mysql.NewDevice()
	device.ID = userDeviceRelation.DeviceId
	device.QueryByID(device.ID)
	mysql.RemoveDeviceOverviewByMac(device.Mac)
	// 移除订阅的topic
	mysql.UnsubscribeDeviceTopic(device.Mac)
	return common.Success, "remove device ok"
}

/******************************************************************************
 * function: QueryTransferedUsers
 * description:  查询设备过户给哪些用户
 * param {*gin.Context} c
 * return {*}
********************************************************************************/
func QueryTransferedUsers(c *gin.Context) (int, interface{}) {
	mac := c.Query("mac")
	if mac == "" {
		return common.ParamError, "device mac address required"
	}
	userId := c.Query("user_id")
	if userId == "" {
		return common.ParamError, "user id required"
	}
	filter := fmt.Sprintf("mac='%s'", mac)
	var deviceList []mysql.Device
	mysql.QueryDeviceByCond(filter, nil, nil, &deviceList)
	if len(deviceList) == 0 {
		return common.NoData, "device not exist"
	}
	device := deviceList[0]
	mUserId, _ := strconv.ParseInt(userId, 10, 64)
	var gList []mysql.UserTransferDeviceDetail
	mysql.QueryUserTransferDeviceDetail(mUserId, 0, device.ID, common.DeviceConfirmFlag, &gList)
	return http.StatusOK, gList
}

/******************************************************************************
 * function: QueryUnconfirmedTransferDevices
 * description: 查询未确认的过户设备
 * param {*gin.Context} c
 * return {*}
********************************************************************************/
func QueryUnconfirmedTransferDevices(c *gin.Context) (int, interface{}) {
	userId := c.Query("user_id")
	if userId == "" {
		return common.ParamError, "user id required"
	}
	mac := c.Query("mac")
	device := mysql.NewDevice()
	var deviceList []mysql.Device
	if mac != "" {
		filter := fmt.Sprintf("mac='%s'", mac)
		mysql.QueryDeviceByCond(filter, nil, nil, &deviceList)
		if len(deviceList) == 0 {
			return common.NoData, "device not exist"
		}
		device = &deviceList[0]
	}
	mUserId, _ := strconv.ParseInt(userId, 10, 64)
	var gList []mysql.UserTransferDeviceDetail
	mysql.QueryUserTransferDeviceDetail(0, mUserId, device.ID, common.DeviceUnConfirmFlag, &gList)
	if len(gList) == 0 {
		return common.NoData, "no unconfirmed share device"
	}
	return http.StatusOK, gList
}

// swagger:model TransferDeviceWithMacReq
type TransferDeviceWithMacReq struct {
	FromUserId  int64  `json:"from_user_id" `
	ToUserPhone string `json:"to_user_phone" `
	// required: true
	// 设备Mac
	Mac string `json:"mac"`
	//备注
	Remark string `json:"remark"`
	// 是否等待确认
	WaitConfirm bool `json:"wait_confirm"`
}

func TransferDeviceToPhoneWithMac(c *gin.Context) (int, interface{}) {
	req := TransferDeviceWithMacReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		return common.JsonError, "json format error"
	}
	if req.Mac == "" || req.FromUserId == 0 || req.ToUserPhone == "" {
		return common.ParamError, "request param error"
	}
	// 查询用户号码是否存在
	var userList []mysql.User
	filter := fmt.Sprintf("phone='%s'", req.ToUserPhone)
	mysql.QueryUserByCond(filter, nil, nil, &userList)
	if len(userList) == 0 {
		return common.NoExist, "user phone not exist"
	}
	user := &userList[0]
	// 检查过户的用户和被过户用户是否同一人
	if user.ID == req.FromUserId {
		return common.SameUser, "can't share device to yourself"
	}
	// check device exist
	userDevice := mysql.NewUserDevice()
	filter = fmt.Sprintf("mac='%s'", req.Mac)
	var deviceList []mysql.Device
	mysql.QueryDeviceByCond(filter, nil, nil, &deviceList)
	if len(deviceList) == 0 {
		return common.NoExist, "device not exist"
	}
	userDevice.Device = deviceList[0]
	// 用户过户设备
	userTransfer := mysql.NewUserTransferDevice()
	userTransfer.FromUserId = req.FromUserId
	userTransfer.ToUserId = user.ID
	userTransfer.DeviceId = userDevice.ID
	userTransfer.Remark = req.Remark
	var userTransferList []mysql.UserTransferDevice
	mysql.QueryUserTransferDevice(
		userTransfer.FromUserId,
		userTransfer.ToUserId,
		userTransfer.DeviceId,
		-1,
		&userTransferList)
	if len(userTransferList) > 0 {
		return common.RepeatData, "device already been transfered, don't need do again"
	}
	if req.WaitConfirm {
		userTransfer.Confirm = common.DeviceUnConfirmFlag
	} else {
		userTransfer.Confirm = common.DeviceConfirmFlag
	}
	status, result := SaveUserTransferDevice(userDevice, userTransfer)
	if req.WaitConfirm && status == common.Success {
		// 向对方发送共享过户请求，表示需要确认才能过户设备
		mq.PublishData(common.MakeTransferDeviceConfirmTopic(userTransfer.ToUserId), userTransfer)
	}
	return status, result
}

func SaveUserTransferDevice(userDevice *mysql.UserDevice, userTransfer *mysql.UserTransferDevice) (int, interface{}) {
	userDevice.UserId = userTransfer.ToUserId
	// 检查是否已经相同过户的设备, 没有则添加, 有则更新
	var userTransferList []mysql.UserTransferDevice
	mysql.QueryUserTransferDevice(
		userTransfer.FromUserId,
		userTransfer.ToUserId,
		userTransfer.DeviceId,
		-1,
		&userTransferList)
	if len(userTransferList) == 0 {
		if !userTransfer.Insert() {
			return common.DBError, "insert error"
		}
	} else {
		userTransfer.ID = userTransferList[0].ID
		if !userTransfer.Update() {
			return common.DBError, "update error"
		}
	}

	// 如果用户已经确认了过户设备，则添加设备和用户的关系
	if userTransfer.Confirm == common.DeviceConfirmFlag {
		status, result := SaveConfirmedTransferDevice(userTransfer)
		if status != common.Success {
			return status, result
		}
	}
	return common.Success, userDevice
}

func SaveConfirmedTransferDevice(userTransfer *mysql.UserTransferDevice) (int, interface{}) {
	// 添加用户和设备的关联关系
	userDeviceRelation := mysql.NewUserDeviceRelation()
	userDeviceRelation.UserId = userTransfer.ToUserId
	userDeviceRelation.DeviceId = userTransfer.DeviceId
	userDeviceRelation.Flag = common.NormalDeviceFlag // normal device is 0, transfer device flag = 0
	// check if the relation exist
	// 如果设备已经绑定到用户并且是主动绑定的NormalDeviceFlag，则不允许过户
	// 如果是共享设备，还可以继续过户
	filter := fmt.Sprintf("user_id=%d and device_id=%d and flag=%d",
		userDeviceRelation.UserId,
		userDeviceRelation.DeviceId,
		common.NormalDeviceFlag)
	var gList []mysql.UserDeviceRelation
	mysql.QueryUserDeviceRelationByCond(filter, nil, nil, &gList)
	if len(gList) > 0 {
		return common.AlreadyBind, "device has binded to the user, not allow transfer"
	} else {
		// 删除原来用户的设备关系，以及原来设备分享的关系
		orgUserDeviceRelation := mysql.NewUserDeviceRelation()
		orgUserDeviceRelation.UserId = userTransfer.FromUserId
		orgUserDeviceRelation.DeviceId = userTransfer.DeviceId
		orgUserDeviceRelation.Flag = common.NormalDeviceFlag
		orgUserDeviceRelation.DeleteWithUser()
		// 删除原来设备的共享用户
		mysql.DeleteUserShareDeviceByUserId(userTransfer.FromUserId, 0, userTransfer.DeviceId)
		// 保存新的用户和设备关系
		if !userDeviceRelation.Insert() {
			return common.DBError, "insert error"
		}
	}
	return common.Success, userDeviceRelation
}

// swagger:model ConfirmTransferReq
type ConfirmTransferReq struct {
	UserId   int64 `json:"user_id" mysql:"user_id" `
	DeviceId int64 `json:"device_id" mysql:"device_id" `
	// required: true
	// 是否确认共享设备, true:确认共享设备, false:拒绝共享设备
	Confirm bool `json:"confirm" mysql:"confirm" `
}
type ConfirmFinishedMq struct {
	ToUserId int64 `json:"to_user_id" `
	DeviceId int64 `json:"device_id" `
	Result   int   `json:"result" `
}

func ConfirmTransferedDevice(c *gin.Context) (int, interface{}) {
	req := &ConfirmTransferReq{}
	err := c.ShouldBindJSON(&req)
	if err != nil {
		return common.JsonError, "json format error"
	}
	if req.UserId == 0 || req.DeviceId == 0 {
		return common.ParamError, "user id and device id required"
	}
	var userTransferDeviceList []mysql.UserTransferDevice
	mysql.QueryUserTransferDevice(0, req.UserId, req.DeviceId, common.DeviceUnConfirmFlag, &userTransferDeviceList)
	if len(userTransferDeviceList) == 0 {
		return common.NoData, "no unconfirmed transfer device"
	}
	userTransferDevice := &userTransferDeviceList[0]
	confirmFinished := ConfirmFinishedMq{
		ToUserId: userTransferDevice.ToUserId,
		DeviceId: userTransferDevice.DeviceId,
	}
	if req.Confirm {
		userTransferDevice.Confirm = common.DeviceConfirmFlag
		if userTransferDevice.Update() {
			status, result := SaveConfirmedTransferDevice(userTransferDevice)
			if status != common.Success {
				return status, result
			}
			// 向对方from_user发送确认过户设备请求结果
			confirmFinished.Result = 1 // 1:确认
			mq.PublishData(common.MakeTransferConfirmedFinishedTopic(userTransferDevice.FromUserId), confirmFinished)
			return common.Success, "Confirmed successfully"
		}
	} else {
		// 如果用户拒绝则删除过户表中的记录
		userTransferDevice.Delete()
		// 如果已经存在用户和设备关系记录，也一起删除
		var userDeviceRelation = mysql.NewUserDeviceRelation()
		userDeviceRelation.UserId = userTransferDevice.ToUserId
		userDeviceRelation.DeviceId = userTransferDevice.DeviceId
		userDeviceRelation.Flag = common.NormalDeviceFlag
		userDeviceRelation.DeleteWithUser()
		// 向对方from_user发送确认过户设备请求结果
		confirmFinished.Result = 0 // 0:拒绝
		mq.PublishData(common.MakeTransferConfirmedFinishedTopic(userTransferDevice.FromUserId), confirmFinished)
		return common.Success, "Unconfirmed successfully"
	}
	return common.DBError, "update failed!"
}

/******************************************************************************
 * function: QueryDeviceOverview
 * description: 根据设备mac查询设备概况
 * param {*gin.Context} c
 * return {*}
********************************************************************************/

func QueryDeviceOverview(c *gin.Context) (int, interface{}) {
	mac := c.Query("mac")
	if mac == "" {
		return common.ParamError, "mac required"
	}

	overviewList := make([]mysql.DeviceOverview, 0)
	mysql.QueryDeviceOverviewByMac(mac, &overviewList)
	if len(overviewList) == 0 {
		return common.NoData, "no data"
	}
	return common.Success, overviewList[0]
}

/******************************************************************************
 * function: UpdateDeviceOverview
 * description: 更新设备概况数据
 * param {*gin.Context} c
 * return {*}
********************************************************************************/
// swagger:model DeviceOverviewReq
type DeviceOverviewReq struct {
	// required: true
	Mac string `json:"mac"`
	// 0 未知 1 男 2 女
	Gender int `json:"gender" mysql:"gender"`
	// 出生日期
	BornDate string `json:"born_date"`
	// 年级
	Grade string `json:"grade"`
}

func UpdateDeviceOverview(c *gin.Context) (int, interface{}) {
	var req = &DeviceOverviewReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		return common.JsonError, "json format error"
	}
	if req.Mac == "" {
		return common.ParamError, "mac required"
	}
	overview := mysql.NewDeviceOverview()
	overviewList := make([]mysql.DeviceOverview, 0)
	mysql.QueryDeviceOverviewByMac(overview.Mac, &overviewList)
	if len(overviewList) > 0 {
		overview = &overviewList[0]
	}
	overview.Mac = req.Mac
	overview.Gender = req.Gender
	overview.BornDate = &req.BornDate
	overview.Grade = &req.Grade
	if overview.ID == 0 {
		if overview.Insert() {
			return common.Success, overview
		}
	} else {
		if overview.Update() {
			return common.Success, overview
		}
	}
	return common.DBError, "update failed"
}
