/******************************************************************************
 * Author: liguoqiang
 * Date: 2024-04-18 19:19:08
 * LastEditors: liguoqiang
 * LastEditTime: 2024-05-25 10:45:28
 * Description:
********************************************************************************/
package mdb

import (
	"hjyserver/mdb/common"
	"hjyserver/mdb/mysql"
	"strconv"

	"github.com/gin-gonic/gin"
)

func UpdateNotifySetting(c *gin.Context) (int, interface{}) {
	req := mysql.NotifySetting{}
	if err := c.ShouldBindJSON(&req); err != nil {
		return common.ParamError, "json bind failed"
	}
	if req.Mac == "" {
		return common.ParamError, "param error, need mac filed"
	}
	if req.Type == 0 {
		return common.ParamError, "param error, need type filed"
	}
	if req.LastNotifyTime == "" {
		req.LastNotifyTime = common.GetNowTime()
	}
	result := common.Success
	obj, _ := mysql.QueryNotifySettingByType(req.Mac, req.Type)
	if obj == nil {
		if req.Insert() {
			result = common.Success
			obj = &req
		} else {
			result = common.DBError
		}
	} else {
		obj.Switch = req.Switch
		obj.IntervalTime = req.IntervalTime
		obj.HighValue = req.HighValue
		obj.LowValue = req.LowValue
		if obj.Update() {
			result = common.Success
		} else {
			result = common.DBError
		}
	}
	if result == common.Success {
		return result, obj
	}
	return result, "update notify setting failed"
}

// swagger:model PeopleNotifySettingReq
type PeopleNotifySettingReq struct {
	Mac          string `json:"mac"`
	Switch       int    `json:"switch"`
	IntervalTime int    `json:"interval_time"`
}

/******************************************************************************
 * function:
 * description:
 * param {*gin.Context} c
 * return {*}
********************************************************************************/
func PeopleNotifySetting(c *gin.Context) (int, interface{}) {
	req := PeopleNotifySettingReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		return common.ParamError, "json bind failed"
	}
	obj, _ := mysql.QueryNotifySettingByType(req.Mac, common.PeopleType)
	if obj == nil {
		obj = mysql.NewNotifySetting()
		obj.Type = common.PeopleType
		obj.Switch = req.Switch
		obj.IntervalTime = req.IntervalTime
		if obj.Insert() {
			return common.Success, obj
		} else {
			return common.DBError, "insert people notify setting failed"
		}
	} else {
		obj.Switch = req.Switch
		obj.IntervalTime = req.IntervalTime
		if obj.Update() {
			return common.Success, obj
		} else {
			return common.DBError, "update people notify setting failed"
		}
	}
}

// swagger:model BreathNotifySettingReq
type BreathNotifySettingReq struct {
	Mac          string `json:"mac"`
	Switch       int    `json:"switch"`
	IntervalTime int    `json:"interval_time"`
	HighValue    int    `json:"high_value"`
	LowValue     int    `json:"low_value"`
}

/******************************************************************************
 * function:
 * description:
 * param {*gin.Context} c
 * return {*}
********************************************************************************/
func BreathNotifySetting(c *gin.Context) (int, interface{}) {
	req := BreathNotifySettingReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		return common.ParamError, "json bind failed"
	}
	obj, _ := mysql.QueryNotifySettingByType(req.Mac, common.BreathType)
	if obj == nil {
		obj = mysql.NewNotifySetting()
		obj.Type = common.BreathType
		obj.Switch = req.Switch
		obj.IntervalTime = req.IntervalTime
		obj.HighValue = req.HighValue
		obj.LowValue = req.LowValue
		if obj.Insert() {
			return common.Success, obj
		} else {
			return common.DBError, "insert breath notify setting failed"
		}
	} else {
		obj.Switch = req.Switch
		obj.IntervalTime = req.IntervalTime
		obj.HighValue = req.HighValue
		obj.LowValue = req.LowValue
		if obj.Update() {
			return common.Success, obj
		} else {
			return common.DBError, "update breath notify setting failed"
		}
	}
}

// swagger:model BreathAbnormalNotifySettingReq
type BreathAbnormalNotifySettingReq struct {
	Mac          string `json:"mac"`
	Switch       int    `json:"switch"`
	IntervalTime int    `json:"interval_time"`
}

/******************************************************************************
 * function: BreathAbnormalNotifySetting
 * description:
 * param {*gin.Context} c
 * return {*}
********************************************************************************/
func BreathAbnormalNotifySetting(c *gin.Context) (int, interface{}) {
	req := BreathAbnormalNotifySettingReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		return common.ParamError, "json bind failed"
	}
	obj, _ := mysql.QueryNotifySettingByType(req.Mac, common.BreathAbnormalType)
	if obj == nil {
		obj = mysql.NewNotifySetting()
		obj.Type = common.BreathAbnormalType
		obj.Switch = req.Switch
		obj.IntervalTime = req.IntervalTime
		if obj.Insert() {
			return common.Success, obj
		} else {
			return common.DBError, "insert breath abnormal notify setting failed"
		}
	} else {
		obj.Switch = req.Switch
		obj.IntervalTime = req.IntervalTime
		if obj.Update() {
			return common.Success, obj
		} else {
			return common.DBError, "update breath abnormal notify setting failed"
		}
	}
}

// swagger:model HeartRateNotifySettingReq
type HeartRateNotifySettingReq struct {
	Mac          string `json:"mac"`
	Switch       int    `json:"switch"`
	IntervalTime int    `json:"interval_time"`
	HighValue    int    `json:"high_value"`
	LowValue     int    `json:"low_value"`
}

func HeartRateNotifySetting(c *gin.Context) (int, interface{}) {
	req := HeartRateNotifySettingReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		return common.ParamError, "json bind failed"
	}
	obj, _ := mysql.QueryNotifySettingByType(req.Mac, common.HeartRateType)
	if obj == nil {
		obj = mysql.NewNotifySetting()
		obj.Type = common.HeartRateType
		obj.Switch = req.Switch
		obj.IntervalTime = req.IntervalTime
		obj.HighValue = req.HighValue
		obj.LowValue = req.LowValue
		if obj.Insert() {
			return common.Success, obj
		} else {
			return common.DBError, "insert heart rate notify setting failed"
		}
	} else {
		obj.Switch = req.Switch
		obj.IntervalTime = req.IntervalTime
		obj.HighValue = req.HighValue
		obj.LowValue = req.LowValue
		if obj.Update() {
			return common.Success, obj
		} else {
			return common.DBError, "update heart rate notify setting failed"
		}
	}
}

// swagger:model NurseModelSettingReq
type NurseModelSettingReq struct {
	Mac    string `json:"mac"`
	Switch int    `json:"switch"`
}

/******************************************************************************
 * function: NurseModelSetting
 * description:
 * param {*gin.Context} c
 * return {*}
********************************************************************************/
func NurseModelSetting(c *gin.Context) (int, interface{}) {
	req := NurseModelSettingReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		return common.ParamError, "json bind failed"
	}
	obj, _ := mysql.QueryNotifySettingByType(req.Mac, common.NurseModeType)
	if obj == nil {
		obj = mysql.NewNotifySetting()
		obj.Type = common.NurseModeType
		obj.Switch = req.Switch
		if obj.Insert() {
			return common.Success, obj
		} else {
			return common.DBError, "insert nurse model setting failed"
		}
	} else {
		obj.Switch = req.Switch
		if obj.Update() {
			return common.Success, obj
		} else {
			return common.DBError, "update nurse model setting failed"
		}
	}
}

// swagger:model BeeperSettingReq
type BeeperSettingReq struct {
	Mac    string `json:"mac"`
	Switch int    `json:"switch"`
}

/******************************************************************************
 * function:
 * description:
 * param {*gin.Context} c
 * return {*}
********************************************************************************/
func BeeperSetting(c *gin.Context) (int, interface{}) {
	req := BeeperSettingReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		return common.ParamError, "json bind failed"
	}
	obj, _ := mysql.QueryNotifySettingByType(req.Mac, common.BeeperType)
	if obj == nil {
		obj = mysql.NewNotifySetting()
		obj.Type = common.BeeperType
		obj.Switch = req.Switch
		if obj.Insert() {
			return common.Success, obj
		} else {
			return common.DBError, "insert beeper setting failed"
		}
	} else {
		obj.Switch = req.Switch
		if obj.Update() {
			return common.Success, obj
		} else {
			return common.DBError, "update beeper setting failed"
		}
	}
}

// swagger:model LightSettingReq
type LightSettingReq struct {
	Mac    string `json:"mac"`
	Switch int    `json:"switch"`
}

func LightSetting(c *gin.Context) (int, interface{}) {
	req := LightSettingReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		return common.ParamError, "json bind failed"
	}
	obj, _ := mysql.QueryNotifySettingByType(req.Mac, common.LightType)
	if obj == nil {
		obj = mysql.NewNotifySetting()
		obj.Type = common.LightType
		obj.Switch = req.Switch
		if obj.Insert() {
			return common.Success, obj
		} else {
			return common.DBError, "insert light setting failed"
		}
	} else {
		obj.Switch = req.Switch
		if obj.Update() {
			return common.Success, obj
		} else {
			return common.DBError, "update light setting failed"
		}
	}
}

/******************************************************************************
 * function:
 * description:
 * param {*gin.Context} c
 * return {*}
********************************************************************************/
func QueryNotifySettingByType(c *gin.Context) (int, interface{}) {
	mac := c.Query("mac")
	if mac == "" {
		return common.ParamError, "param error, need mac filed"
	}
	typeStr := c.Query("type")
	if typeStr == "" {
		return common.ParamError, "param error, need type filed"
	}
	typeID, err := strconv.ParseInt(typeStr, 10, 32)
	if err != nil {
		return common.ParamError, "param error, type filed must be int"
	}
	obj, _ := mysql.QueryNotifySettingByType(mac, int(typeID))
	if obj == nil {
		return common.RecordNotFound, "no data"
	}
	return common.Success, obj
}

func QueryAllNotifySetting(c *gin.Context) (int, interface{}) {
	mac := c.Query("mac")
	var gList []mysql.NotifySetting
	if !mysql.QueryAllNotifySetting(mac, &gList) {
		return common.RecordNotFound, "no data"
	}
	return common.Success, gList
}
