/******************************************************************************
 * Author: liguoqiang
 * Date: 2023-11-16 23:18:36
 * LastEditors: liguoqiang
 * LastEditTime: 2023-11-17 09:59:23
 * Description:
********************************************************************************/
package mdb

import (
	"fmt"
	"hjyserver/cfg"
	"hjyserver/mdb/common"
	"hjyserver/mdb/mysql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

/******************************************************************************
 * function: QueryFallCheckStatus
 * description:
 * return {*}
********************************************************************************/
func QueryFallCheckStatus(c *gin.Context) (int, interface{}) {
	mac := c.Query("mac")
	if mac == "" {
		return http.StatusBadRequest, "device mac required"
	}
	filter := fmt.Sprintf("mac='%s' and create_time>='%s'", mac, time.Now().Add(-6*time.Minute).Format(cfg.TmFmtStr))
	var gList []mysql.FallCheck
	mysql.QueryFallCheckByCond(filter, nil, "create_time desc", 1, &gList)
	return http.StatusOK, gList
}

/******************************************************************************
 * function: QueryFallExistRecord
 * description: query fall check data from database when personstate=1
 * query condition include mac, begin_day, end_day
 * return {*}
********************************************************************************/
func QueryFallExistRecord(c *gin.Context) (int, interface{}) {
	mac := c.Query("mac")
	if mac == "" {
		return http.StatusBadRequest, "device mac required"
	}
	beginDay := c.Query("begin_day")
	if beginDay == "" {
		t1, _ := time.ParseDuration("-24h") // 24 hour before
		beginDay = time.Now().Add(t1).Format(cfg.DateFmtStr)
	}
	endDay := c.Query("end_day")
	if endDay == "" {
		endDay = common.GetNowDate()
	}
	var gList []mysql.FallCheck
	filter := fmt.Sprintf("mac='%s' and date(create_time) >= date('%s') and date(create_time) <= date('%s') and person_state=1", mac, beginDay, endDay)
	mysql.QueryFallCheckByCond(filter, nil, "create_time desc", -1, &gList)
	return http.StatusOK, gList
}

/******************************************************************************
 * function: QueryAlarmRecord
 * description: query alarm record by device mac and date
 * return {*}
********************************************************************************/
func QueryAlarmRecord(c *gin.Context) (int, interface{}) {
	mac := c.Query("mac")
	if mac == "" {
		return http.StatusBadRequest, "device mac required"
	}
	beginDay := c.Query("begin_day")
	if beginDay == "" {
		t1, _ := time.ParseDuration("-24h") // 24 hour before
		beginDay = time.Now().Add(t1).Format(cfg.DateFmtStr)
	}
	endDay := c.Query("end_day")
	if endDay == "" {
		endDay = common.GetNowDate()
	}
	var gList []mysql.FallAlarm
	filter := fmt.Sprintf("mac='%s' and date(create_time) >= date('%s') and date(create_time) <= date('%s')", mac, beginDay, endDay)
	mysql.QueryFallAlarmByCond(filter, nil, "create_time desc", &gList)
	return http.StatusOK, gList
}

/******************************************************************************
 * function: QueryFallParams
 * description:
 * param {*gin.Context} c
 * return {*}
********************************************************************************/
func QueryFallParams(c *gin.Context) (int, interface{}) {
	deviceId := c.Query("device_id")
	if deviceId == "" {
		return http.StatusBadRequest, "device id required"
	}
	var gList []mysql.FallParams
	filter := fmt.Sprintf("device_id=%s", deviceId)
	mysql.QueryFallParamsByCond(filter, nil, "create_time desc", 1, &gList)
	if len(gList) <= 0 {
		return http.StatusFound, "no data found"
	}
	return http.StatusOK, gList[0]
}

/******************************************************************************
 * function: InsertFallParams
 * description: insert fall params data to database
 * return {*}
********************************************************************************/
func InsertFallParams(c *gin.Context) (int, interface{}) {
	var fallParams mysql.FallParams
	fallParams.DecodeFromGin(c)
	fallParams.SetID(0)
	fallParams.DateTime = time.Now().Format(cfg.TmFmtStr)
	filter := fmt.Sprintf("device_id=%d", fallParams.DeviceId)
	var gList []mysql.FallParams
	mysql.QueryFallParamsByCond(filter, nil, "create_time desc", 1, &gList)
	if len(gList) > 0 {
		fallParams.SetID(gList[0].ID)
		fallParams.Update()
	} else {
		fallParams.Insert()
	}
	return http.StatusOK, fallParams
}
