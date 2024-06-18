/******************************************************************************
 * Author: liguoqiang
 * Date: 2023-11-20 11:58:18
 * LastEditors: liguoqiang
 * LastEditTime: 2024-06-01 16:36:37
 * Description:
********************************************************************************/
package mdb

import (
	"fmt"
	"hjyserver/cfg"
	"hjyserver/exception"
	"hjyserver/mdb/common"
	"hjyserver/mdb/mysql"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// swagger:model LampRealDataReq
type LampRealDataReq struct {
	Mac string `json:"mac"`
	Req mysql.RealDataReq
}

func OpenLampRealData(c *gin.Context) (int, interface{}) {
	var req *LampRealDataReq = &LampRealDataReq{}
	if err := c.ShouldBindBodyWith(req, binding.JSON); err != nil {
		exception.Throw(http.StatusAccepted, err.Error())
	}
	if req.Mac == "" {
		return http.StatusBadRequest, "device mac required"
	}
	if req.Req.DeadLine == 0 {
		req.Req.DeadLine = time.Now().Add(time.Minute * 30).Unix()
	}
	if req.Req.Freq == 0 {
		req.Req.Freq = 6
	}
	req.Req.SendRealDataReq(req.Mac)
	return http.StatusOK, nil
}

/******************************************************************************
 * function: QueryLampRealData
 * description: query lamp real data by mac and time
 * param {*gin.Context} c
 * return {*}
********************************************************************************/
func QueryLampRealData(c *gin.Context) (int, interface{}) {
	mac := c.Query("mac")
	if mac == "" {
		return http.StatusBadRequest, "device mac required"
	}
	beginDay := c.Query("begin_day")
	endDay := c.Query("end_day")
	var filter string
	var limit int
	if beginDay == "" && endDay == "" {
		filter = fmt.Sprintf("mac='%s' and respiratory > 0", mac)
		limit = 1
	} else {
		if beginDay == "" {
			t1, _ := time.ParseDuration("-12h") // 12 hour before
			beginDay = time.Now().Add(t1).Format(cfg.DateFmtStr)
		}
		if endDay == "" {
			endDay = common.GetNowDate()
		}
		filter = fmt.Sprintf("mac='%s' and create_time >= '%s' and create_time <= '%s' and respiratory > 0", mac, beginDay, endDay)
		limit = -1
	}
	var gList []mysql.RealDataSql
	mysql.QueryLampRealDataByCond(filter, nil, "create_time desc", limit, &gList)
	return http.StatusOK, gList
}

/******************************************************************************
 * function:QueryLampReportStatus
 * description: query lamp report status by mac and time
 * param {*gin.Context} c
 * return {*}
********************************************************************************/
func QueryLampReportStatus(c *gin.Context) (int, interface{}) {
	mac := c.Query("mac")
	if mac == "" {
		return common.ParamError, "device mac required"
	}

	beginDay := c.Query("begin_day")
	endDay := c.Query("end_day")
	if beginDay == "" {
		beginDay = common.GetNowDate()
	}
	if endDay == "" {
		endDay = common.GetNowDate()
	}
	var gList []mysql.LampReportStatus
	mysql.QueryLampReportStatusByMac(mac, beginDay, endDay, &gList)
	return http.StatusOK, gList
}

/******************************************************************************
 * function:
 * description:
 * return {*}
********************************************************************************/
func QueryLampEvent(c *gin.Context) (int, interface{}) {
	mac := c.Query("mac")
	if mac == "" {
		return http.StatusBadRequest, "device mac required"
	}
	beginDay := c.Query("begin_day")
	endDay := c.Query("end_day")
	var filter string
	var limit int
	if beginDay == "" && endDay == "" {
		filter = fmt.Sprintf("mac='%s'", mac)
		limit = 1
	} else {
		if beginDay == "" {
			t1, _ := time.ParseDuration("-12h") // 12 hour before
			beginDay = time.Now().Add(t1).Format(cfg.DateFmtStr)
		}
		if endDay == "" {
			endDay = common.GetNowDate()
		}
		filter = fmt.Sprintf("mac='%s' and date(create_time) >= date('%s') and date(create_time) <= date('%s')", mac, beginDay, endDay)
		limit = -1
	}
	var gList []mysql.EventReportSql
	mysql.QueryLampEventByCond(filter, nil, "create_time desc", limit, &gList)
	return http.StatusOK, gList
}

// swagger:model ControlLampReq
type ControlLampReq struct {
	Mac string `json:"mac"`
	mysql.LampControlJson
}

/******************************************************************************
 * function:
 * description:
 * return {*}
********************************************************************************/
func ControlLamp(c *gin.Context) (int, interface{}) {
	var req *ControlLampReq = &ControlLampReq{}
	if err := c.ShouldBindBodyWith(req, binding.JSON); err != nil {
		exception.Throw(http.StatusAccepted, err.Error())
	}
	if req.Mac == "" {
		return http.StatusBadRequest, "device mac required"
	}
	req.SendControl(req.Mac)
	return http.StatusOK, nil
}

func QueryLampControl(c *gin.Context) (int, interface{}) {
	mac := c.Query("mac")
	if mac == "" {
		return http.StatusBadRequest, "device mac required"
	}
	var gList []mysql.LampControlSql
	mysql.QueryLampControlByCond("mac='"+mac+"'", nil, "create_time desc", 1, &gList)
	if len(gList) == 0 {
		return http.StatusBadRequest, "no control data"
	}
	return http.StatusOK, gList[0]
}

// swagger:model CreateStudyRoomReq
type CreateStudyRoomReq struct {
	// required: true
	// example: study room
	Name string `json:"name"`
	// required: true
	// example: 1
	CreateId int64 `json:"create_id"`
	// required: true
	// example: 0,  1--auto invite create
	Flag int `json:"flag"`
}

/******************************************************************************
 * function: CreateStudyRoom
 * description:  create study room
 * return {*}
********************************************************************************/
func CreateStudyRoom(c *gin.Context) (int, interface{}) {
	var req *CreateStudyRoomReq = &CreateStudyRoomReq{}
	err := c.ShouldBindJSON(req)
	if err != nil {
		exception.Throw(http.StatusAccepted, err.Error())
	}
	var filter string = fmt.Sprintf("create_id=%d and name='%s' and status=1", req.CreateId, req.Name)
	var objs []mysql.StudyRoom
	if mysql.QueryStudyRoomByCond(filter, nil, "", 1, &objs) && len(objs) > 0 {
		return common.RepeatData, "study room name already exist"
	}
	if mysql.UserInAnyStudyRoom(req.CreateId) {
		return common.HasExist, "user already in study room"
	}
	var room *mysql.StudyRoom = mysql.NewStudyRoom()
	room.Capacity = 6
	room.Status = 1
	room.CreateId = req.CreateId
	room.CurrentNum = 0
	room.Name = req.Name
	room.CreateTime = common.GetNowTime()
	if !room.Insert() {
		return http.StatusBadRequest, "create study room failed"
	}
	if req.Flag == 1 {
		var userRoom *mysql.StudyRoomUser = mysql.NewStudyRoomUser()
		userRoom.RoomId = room.ID
		userRoom.UserId = req.CreateId
		userRoom.Status = 1
		userRoom.Sn = 0
		userRoom.CreateTime = common.GetNowTime()
		userRoom.Insert()
	}
	return http.StatusOK, room
}

/******************************************************************************
 * function: ModifyStudyRoom
 * description:
 * param {*gin.Context} c
 * return {*}
********************************************************************************/
// swagger:model ModifyStudyRoomReq
type ModifyStudyRoomReq struct {
	// required: true
	// example: 1
	RoomId int64 `json:"room_id"`
	// required: true
	// example: "study room name"
	Name string `json:"name"`
}

func ModifyStudyRoom(c *gin.Context) (int, interface{}) {
	var req *ModifyStudyRoomReq = &ModifyStudyRoomReq{}
	err := c.ShouldBindJSON(req)
	if err != nil {
		exception.Throw(http.StatusAccepted, err.Error())
	}
	var filter string = fmt.Sprintf("id=%d and status=1", req.RoomId)
	var objs []mysql.StudyRoom
	if !mysql.QueryStudyRoomByCond(filter, nil, "", 1, &objs) || len(objs) == 0 {
		return http.StatusBadRequest, "study room not exist"
	}
	var obj = &objs[0]
	obj.Name = req.Name
	if !obj.Update() {
		return http.StatusBadRequest, "modify study room failed"
	}
	return http.StatusOK, "modify study room success"
}

/******************************************************************************
 * function: ReleaseStudyRoom
 * description: release study room
 * return {*}
********************************************************************************/
// swagger:model ReleaseStudyRoomReq
type ReleaseStudyRoomReq struct {
	// required: true
	// example: 1
	RoomId int64 `json:"room_id"`
	// required: true
	// example: 1
	CreateId int64 `json:"create_id"`
}

func ReleaseStudyRoom(c *gin.Context) (int, interface{}) {
	var req *ReleaseStudyRoomReq = &ReleaseStudyRoomReq{}
	err := c.ShouldBindJSON(req)
	if err != nil {
		exception.Throw(http.StatusAccepted, err.Error())
	}
	var obj *mysql.StudyRoom = mysql.NewStudyRoom()
	obj.ID = req.RoomId
	obj.CreateId = req.CreateId
	var filter string = fmt.Sprintf("id=%d and create_id=%d and status=1", obj.ID, obj.CreateId)
	var objs []mysql.StudyRoom
	if !mysql.QueryStudyRoomByCond(filter, nil, "", 1, &objs) || len(objs) == 0 {
		return http.StatusBadRequest, "study room not exist"
	}
	obj = &objs[0]
	obj.Status = 0
	if !obj.Update() {
		return http.StatusBadRequest, "release study room failed"
	}
	mysql.CleanUserStudyRoomStatus(0, req.RoomId)
	// mysql.CleanStudyRecordStatus(0, req.RoomId)
	return http.StatusOK, "release study room success"
}

/******************************************************************************
 * function: QueryStudyRoom
 * description:  query study room by create user id or be invited user id
 * return {*}
********************************************************************************/
func QueryStudyRoom(c *gin.Context) (int, interface{}) {
	userId := c.Query("user_id")
	if userId == "" {
		userId = "0"
	}
	var objs []mysql.StudyRoom
	idInt, _ := strconv.ParseInt(userId, 10, 64)
	if !mysql.QueryStudyRoomByUser(idInt, &objs) {
		return common.RecordNotFound, "not found any record"
	}
	return http.StatusOK, objs
}

/******************************************************************************
 * function: QueryInviteStudyRoom
 * description: query invite study room by user id
 * return {*}
********************************************************************************/
func QueryInviteStudyRoom(c *gin.Context) (int, interface{}) {
	userId := c.Query("user_id")
	if userId == "" {
		return http.StatusBadRequest, "user id required"
	}
	var gList []mysql.StudyRoomUserDetail
	idInt, _ := strconv.ParseInt(userId, 10, 64)
	mysql.QueryInviteStudyRoomDetailByUserId(idInt, &gList)
	return http.StatusOK, gList
}

/******************************************************************************
 * function: QueryRankingByStudyRoom
 * description:
 * return {*}
********************************************************************************/
func QueryRankingByStudyRoom(c *gin.Context) (int, interface{}) {
	roomId := c.Query("room_id")
	if roomId == "" {
		return http.StatusBadRequest, "room id required"
	}
	var gList []mysql.StudyRoomRanking
	idInt, _ := strconv.ParseInt(roomId, 10, 64)
	mysql.QueryRankingByStudyRoom(idInt, &gList)
	return http.StatusOK, gList
}

/******************************************************************************
 * function:
 * description:
 * param {*gin.Context} c
 * return {*}
********************************************************************************/
func StatsLampFlowData(c *gin.Context) (int, interface{}) {
	mac := c.Query("mac")
	if mac == "" {
		return http.StatusBadRequest, "mac required"
	}
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")
	if startTime == "" || endTime == "" {
		return http.StatusBadRequest, "start time and end time required"
	}
	flowData := mysql.StatsLampFlowDataByTime(mac, startTime, endTime)
	type FlowData struct {
		Flow int `json:"flow"`
	}
	var obj = &FlowData{}
	obj.Flow = flowData
	return http.StatusOK, obj
}
