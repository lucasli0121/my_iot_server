/******************************************************************************
 *                        _oo0oo_
 *                       o8888888o
 *                       88" . "88
 *                       (| -_- |)
 *                       0\  =  /0
 *                     ___/`---'\___
 *                   .' \\|     |// '.
 *                  / \\|||  :  |||// \
 *                 / _||||| -:- |||||- \
 *                |   | \\\  - /// |   |
 *                | \_|  ''\---/''  |_/ |
 *                \  .-\__  '-'  ___/-. /
 *              ___'. .'  /--.--\  `. .'___
 *           ."" '<  `.___\_<|>_/___.' >' "".
 *          | | :  `- \`.;`\ _ /`;.`/ - ` : | |
 *          \  \ `_.   \_ __\ /__ _/   .-` /  /
 *      =====`-.____`.___ \_____/___.-`___.-'=====
 *                        `=---='
 *
 *
 *      ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
 *
 *            佛祖保佑     永不宕机     永无BUG
********************************************************************************/

/******************************************************************************
 *                        _oo0oo_
 *                       o8888888o
 *                       88" . "88
 *                       (| -_- |)
 *                       0\  =  /0
 *                     ___/`---'\___
 *                   .' \\|     |// '.
 *                  / \\|||  :  |||// \
 *                 / _||||| -:- |||||- \
 *                |   | \\\  - /// |   |
 *                | \_|  ''\---/''  |_/ |
 *                \  .-\__  '-'  ___/-. /
 *              ___'. .'  /--.--\  `. .'___
 *           ."" '<  `.___\_<|>_/___.' >' "".
 *          | | :  `- \`.;`\ _ /`;.`/ - ` : | |
 *          \  \ `_.   \_ __\ /__ _/   .-` /  /
 *      =====`-.____`.___ \_____/___.-`___.-'=====
 *                        `=---='
 *
 *
 *      ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
 *
 *            佛祖保佑     永不宕机     永无BUG
********************************************************************************/

/******************************************************************************
 * Author: liguoqiang
 * Date: 2023-11-17 23:31:03
 * LastEditors: liguoqiang
 * LastEditTime: 2024-07-09 14:36:39
 * Description:
********************************************************************************/
package mdb

import (
	"fmt"
	"hjyserver/cfg"
	mylog "hjyserver/log"
	"hjyserver/mdb/common"
	"hjyserver/mdb/mysql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

/******************************************************************************
 * function: QueryHeartRate
 * description: if not day condition then query the latest record from database
 * return {*}
********************************************************************************/
func QueryHeartRate(c *gin.Context) (int, interface{}) {
	mac := c.Query("mac")
	if mac == "" {
		return http.StatusBadRequest, "device mac required"
	}
	beginDay := c.Query("begin_day")
	endDay := c.Query("end_day")
	var devices []mysql.Device
	mysql.QueryDeviceByCond(fmt.Sprintf("mac='%s'", mac), nil, "", &devices)
	if len(devices) == 0 {
		return http.StatusAccepted, "not find any device in the condition"
	}
	device := devices[0]
	switch device.Type {
	case mysql.HeatRateType:
		return queryHeartRateTypeData(mac, beginDay, endDay)
	case mysql.Ed713Type:
		return queryEd713TypeData(mac, beginDay, endDay)
	}
	return http.StatusAccepted, "not support device type"
}

/******************************************************************************
 * function: QueryHeartRateByMinute
 * description: stats heart rate data by minute
 * param {*gin.Context} c
 * return {*}
********************************************************************************/

// swagger:model StatsHeartRate
type StatsHeartRate struct {
	// required: true
	HeartRate       int                  `json:"heart_rate" mysql:"heart_rate"`
	BreathRate      int                  `json:"breath_rate" mysql:"breath_rate"`
	Physical        int                  `json:"physical" mysql:"physical"`
	HeartRateRange  []int                `json:"heart_rate_range" mysql:"heart_rate_range"`
	BreathRateRange []int                `json:"breath_rate_range" mysql:"breath_rate_range"`
	PhysicalRange   []int                `json:"physical_range" mysql:"physical_range"`
	StatsItems      []StatsHeartRateItem `json:"stats_item" mysql:"stats_item"`
}

// swagger:model StatsHeartRateItem
type StatsHeartRateItem struct {
	HighHeartRate  int `json:"high_heart_rate"`
	LowHeartRate   int `json:"low_heart_rate"`
	HighBreathRate int `json:"high_breath_rate"`
	LowBreathRate  int `json:"low_breath_rate"`
	HighPhysical   int `json:"high_physical"`
	LowPhysical    int `json:"low_physical"`
}

func StatsHeartRateByMinute(c *gin.Context) (int, interface{}) {
	mac := c.Query("mac")
	if mac == "" {
		return http.StatusBadRequest, "device mac required"
	}
	beginDay := c.Query("begin_day")
	endDay := c.Query("end_day")
	var devices []mysql.Device
	mysql.QueryDeviceByCond(fmt.Sprintf("mac='%s'", mac), nil, "", &devices)
	if len(devices) == 0 {
		return http.StatusAccepted, "not find any device in the condition"
	}
	device := devices[0]
	var status = http.StatusOK
	var result interface{}
	switch device.Type {
	case mysql.HeatRateType:
		status, result = queryHeartRateTypeData(mac, beginDay, endDay)
	case mysql.Ed713Type:
		status, result = queryEd713TypeData(mac, beginDay, endDay)
	case mysql.LampType:
		status, result = queryHl77TypeData(mac, beginDay, endDay)
	case mysql.X1Type:
		status, result = queryX1TypeData(mac, beginDay, endDay)
	default:
		status = http.StatusNoContent
	}
	if status == http.StatusOK && result != nil {
		var gList []mysql.HeartRate = result.([]mysql.HeartRate)
		return statsHeartRateByMinute(gList)
	}
	return http.StatusAccepted, "not support device type"
}

/******************************************************************************
 * function:
 * description:
 * param {[]mysql.HeartRate} gList
 * return {*}
********************************************************************************/
func statsHeartRateByMinute(gList []mysql.HeartRate) (int, interface{}) {
	var stats StatsHeartRate
	var fistTm time.Time
	obj := &StatsHeartRateItem{}
	var minHeartRate = 0
	var maxHeartRate = 0
	var minBreathRate = 0
	var maxBreathRate = 0
	var minPhysical = 0
	var maxPhysical = 0
	var totalHeartRate = 0
	var totalBreathRate = 0
	var totalPhysical = 0
	for i, v := range gList {
		totalHeartRate += v.HeartRate
		totalBreathRate += v.BreatheRate
		totalPhysical += v.PhysicalRate
		if i == 0 {
			t, err := common.StrToTime(v.DateTime)
			if err != nil {
				mylog.Log.Error("parse time failed, ", err)
				continue
			}
			obj.HighHeartRate = v.HeartRate
			obj.LowHeartRate = v.HeartRate
			obj.HighBreathRate = v.BreatheRate
			obj.LowBreathRate = v.BreatheRate
			obj.HighPhysical = v.PhysicalRate
			obj.LowPhysical = v.PhysicalRate
			minHeartRate = v.HeartRate
			maxHeartRate = v.HeartRate
			minBreathRate = v.BreatheRate
			maxBreathRate = v.BreatheRate
			minPhysical = v.PhysicalRate
			maxPhysical = v.PhysicalRate
			fistTm = t
			continue
		}
		if v.HeartRate > maxHeartRate {
			maxHeartRate = v.HeartRate
		}
		if v.HeartRate < minHeartRate {
			minHeartRate = v.HeartRate
		}
		if v.BreatheRate > maxBreathRate {
			maxBreathRate = v.BreatheRate
		}
		if v.BreatheRate < minBreathRate {
			minBreathRate = v.BreatheRate
		}
		if v.PhysicalRate > maxPhysical {
			maxPhysical = v.PhysicalRate
		}
		if v.PhysicalRate < minPhysical {
			minPhysical = v.PhysicalRate
		}
		t, err := common.StrToTime(v.DateTime)
		if err != nil {
			mylog.Log.Error("parse time failed, ", err)
			continue
		}
		diffM := t.Sub(fistTm).Minutes()
		if diffM <= 29 {
			if v.HeartRate > obj.HighHeartRate {
				obj.HighHeartRate = v.HeartRate
			}
			if v.HeartRate < obj.LowHeartRate {
				obj.LowHeartRate = v.HeartRate
			}
			if v.BreatheRate > obj.HighBreathRate {
				obj.HighBreathRate = v.BreatheRate
			}
			if v.BreatheRate < obj.LowBreathRate {
				obj.LowBreathRate = v.BreatheRate
			}
			if v.PhysicalRate > obj.HighPhysical {
				obj.HighPhysical = v.PhysicalRate
			}
			if v.PhysicalRate < obj.LowPhysical {
				obj.LowPhysical = v.PhysicalRate
			}
		} else {
			stats.StatsItems = append(stats.StatsItems, *obj)
			obj = &StatsHeartRateItem{}
			obj.HighHeartRate = v.HeartRate
			obj.LowHeartRate = v.HeartRate
			obj.HighBreathRate = v.BreatheRate
			obj.LowBreathRate = v.BreatheRate
			obj.HighPhysical = v.PhysicalRate
			obj.LowPhysical = v.PhysicalRate
			fistTm = t
		}
	}
	var l = len(gList)
	var avgHeartRate = 0
	var avgBreathRate = 0
	var avgPhysical = 0

	if l > 0 {
		avgHeartRate = totalHeartRate / l
		avgBreathRate = totalBreathRate / l
		avgPhysical = totalPhysical / l
	}
	stats.BreathRate = avgBreathRate
	stats.HeartRate = avgHeartRate
	stats.Physical = avgPhysical
	stats.HeartRateRange = append(stats.HeartRateRange, avgHeartRate)
	stats.HeartRateRange = append(stats.HeartRateRange, maxHeartRate)
	stats.BreathRateRange = append(stats.BreathRateRange, avgBreathRate)
	stats.BreathRateRange = append(stats.BreathRateRange, maxBreathRate)
	stats.PhysicalRange = append(stats.PhysicalRange, avgPhysical)
	stats.PhysicalRange = append(stats.PhysicalRange, maxPhysical)
	return http.StatusOK, stats
}

func queryHeartRateTypeData(mac string, beginDay string, endDay string) (int, interface{}) {
	var filter string
	var limit int
	if beginDay == "" && endDay == "" {
		filter = fmt.Sprintf("mac='%s' and person_num > 0 and heart_rate > 0", mac)
		limit = 1
	} else {
		if beginDay == "" {
			t1, _ := time.ParseDuration("-12h") // 12 hour before
			beginDay = time.Now().Add(t1).Format(cfg.DateFmtStr)
		}
		if endDay == "" {
			endDay = common.GetNowDate()
		}
		filter = fmt.Sprintf("mac='%s' and date(create_time) >= date('%s') and date(create_time) <= date('%s') and person_num > 0 and heart_rate > 0", mac, beginDay, endDay)
		limit = -1
	}
	var gList []mysql.HeartRate
	mysql.QueryHeartRateByCond(filter, nil, "create_time", limit, &gList)
	return http.StatusOK, gList
}

func queryEd713TypeData(mac string, beginDay string, endDay string) (int, interface{}) {
	var filter string
	var limit int
	if beginDay == "" && endDay == "" {
		filter = fmt.Sprintf("mac='%s' and heart_rate > 0", mac)
		limit = 1
	} else {
		if beginDay == "" {
			t1, _ := time.ParseDuration("-12h") // 12 hour before
			beginDay = time.Now().Add(t1).Format(cfg.DateFmtStr)
		}
		if endDay == "" {
			endDay = common.GetNowDate()
		}
		filter = fmt.Sprintf("mac='%s' and date(create_time) >= date('%s') and date(create_time) <= date('%s') and heart_rate > 0", mac, beginDay, endDay)
		limit = -1
	}
	var gList []mysql.HeartRate
	mysql.QueryEd713RealDataToHeartRateByCond(filter, nil, "create_time", limit, &gList)
	return http.StatusOK, gList
}

/******************************************************************************
 * function:
 * description: 按照日期统计X1实时数据并转换成HeartRate格式
 * param {string} mac
 * param {string} beginDay
 * param {string} endDay
 * return {*}
********************************************************************************/
func queryX1TypeData(mac string, beginDay string, endDay string) (int, interface{}) {
	var filter string
	var limit int
	if beginDay == "" && endDay == "" {
		filter = fmt.Sprintf("mac='%s' and heart_rate > 0", mac)
		limit = 1
	} else {
		if beginDay == "" {
			t1, _ := time.ParseDuration("-12h") // 12 hour before
			beginDay = time.Now().Add(t1).Format(cfg.DateFmtStr)
		}
		if endDay == "" {
			endDay = common.GetNowDate()
		}
		filter = fmt.Sprintf("mac='%s' and date(create_time) >= date('%s') and date(create_time) <= date('%s') and heart_rate > 0", mac, beginDay, endDay)
		limit = -1
	}
	var gList []mysql.HeartRate
	mysql.QueryX1RealDataToHeartRateByCond(filter, nil, "create_time", limit, &gList)
	return http.StatusOK, gList
}

func queryHl77TypeData(mac string, beginDay string, endDay string) (int, interface{}) {
	var filter string
	var limit int
	if beginDay == "" && endDay == "" {
		filter = fmt.Sprintf("mac='%s' and heart_rate > 0", mac)
		limit = 1
	} else {
		if beginDay == "" {
			t1, _ := time.ParseDuration("-12h") // 12 hour before
			beginDay = time.Now().Add(t1).Format(cfg.DateFmtStr)
		}
		if endDay == "" {
			endDay = common.GetNowDate()
		}
		filter = fmt.Sprintf("mac='%s' and date(create_time) >= date('%s') and date(create_time) <= date('%s') and respiratory > 0", mac, beginDay, endDay)
		limit = -1
	}
	var gList []mysql.HeartRate
	mysql.QueryHl77RealDataToHeartRateByCond(filter, nil, "create_time", limit, &gList)
	return http.StatusOK, gList
}

/******************************************************************************
 * function: QuerySleepReport
 * description: analyze heart rate data and give a report include different sleep time and sleep quality
 * query condition include mac, between start day and end day
 * return {*}
********************************************************************************/
func QuerySleepReport(c *gin.Context) (int, interface{}) {
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
	var devices []mysql.Device
	mysql.QueryDeviceByCond(fmt.Sprintf("mac='%s'", mac), nil, "", &devices)
	if len(devices) == 0 {
		return http.StatusAccepted, "not find any device in the condition"
	}
	device := devices[0]
	switch device.Type {
	case mysql.HeatRateType:
		return queryHeartRateTypeSleepReport(mac, beginDay, endDay)
	case mysql.Ed713Type:
		return queryEd713TypeSleepReport(mac, beginDay, endDay)
	case mysql.X1Type:
		return queryX1TypeSleepReport(mac, beginDay, endDay)
	}
	return http.StatusAccepted, "not support device type"

}

func queryHeartRateTypeSleepReport(mac string, beginDay string, endDay string) (int, interface{}) {
	sleepReport := mysql.NewSleepReport()
	var gList []mysql.HeartRate
	// 24 hour before
	filter := fmt.Sprintf("mac='%s' and date(create_time) >= date('%s') and date(create_time) <= date('%s') and person_num > 0 and heart_rate > 0", mac, beginDay, endDay)
	mysql.QueryHeartRateByCond(filter, nil, "create_time", -1, &gList)
	if len(gList) == 0 {
		return http.StatusAccepted, "not find any data in the condition"
	}
	var hasAwake bool = false
	var hasSleep bool = false
	var hasLight bool = false
	var hasDeep bool = false
	var hasLeaveBed bool = false
	var awakeTm time.Time
	var lightTm time.Time
	var deepTm time.Time

	var beginSleepTm time.Time
	var beginLightTm time.Time
	var beginDeepTm time.Time
	var beginAwakeTm time.Time

	sleepReport.StartTime = gList[0].DateTime
	sleepReport.EndTime = gList[len(gList)-1].DateTime

	t1, err := time.ParseInLocation(cfg.TmFmtStr, sleepReport.StartTime, time.Local)
	if err != nil {
		mylog.Log.Error("parse time failed, ", err)
		return http.StatusInternalServerError, "parse time failed"
	}
	t2, err := time.ParseInLocation(cfg.TmFmtStr, sleepReport.EndTime, time.Local)
	if err != nil {
		mylog.Log.Error("parse time failed, ", err)
		return http.StatusInternalServerError, "parse time failed"
	}
	var totalSeconds = int64(t2.Sub(t1).Seconds())

	for _, v := range gList {
		t, err := time.ParseInLocation(cfg.TmFmtStr, v.DateTime, time.Local)
		if err != nil {
			mylog.Log.Error("parse time failed, ", err)
			continue
		}
		// if sleep feature == 1 it is turn over
		if v.SleepFeatures == 1 {
			sleepReport.TurnOver++
		}
		// StagesStatus > 1 mean it is sleep
		// StagesStatus == 1 mean it is awake or maybe just turn over
		// so we also check the SleepFeatures and ActiveStatus
		// when continuous awake time < 2 mintues mean it is leave bed not end sleep
		if v.StagesStatus > 1 {
			if !hasSleep {
				beginSleepTm = t
				hasSleep = true
			}
			if hasAwake {
				sleepReport.StagesSleepTime = append(sleepReport.StagesSleepTime,
					mysql.StagesSleepTime{
						StagesStatus:   1,
						BeginSleepTime: beginAwakeTm.Format(cfg.TmFmtStr),
						EndSleepTime:   t.Format(cfg.TmFmtStr),
					})
				var diff = t.Sub(awakeTm)
				sleepReport.AwakeLong += int64(diff.Seconds())
				hasAwake = false
			}
			if hasLeaveBed {
				hasLeaveBed = false
			}

			// if light sleep at first time remember it
			// else calculate between now and last light sleep time
			if v.StagesStatus == 2 {
				if hasDeep {
					sleepReport.StagesSleepTime = append(sleepReport.StagesSleepTime,
						mysql.StagesSleepTime{
							StagesStatus:   3,
							BeginSleepTime: beginDeepTm.Format(cfg.TmFmtStr),
							EndSleepTime:   t.Format(cfg.TmFmtStr),
						})
					var diff = t.Sub(deepTm)
					sleepReport.SleepDeep += int64(diff.Seconds())
					hasDeep = false
				}
				if !hasLight {
					beginLightTm = t
					hasLight = true
				} else {
					var diff = t.Sub(lightTm)
					sleepReport.SleepLight += int64(diff.Seconds())
				}
				lightTm = t
			}

			if v.StagesStatus == 3 {
				if hasLight {
					sleepReport.StagesSleepTime = append(sleepReport.StagesSleepTime,
						mysql.StagesSleepTime{
							StagesStatus:   2,
							BeginSleepTime: beginLightTm.Format(cfg.TmFmtStr),
							EndSleepTime:   t.Format(cfg.TmFmtStr),
						})
					var diff = t.Sub(lightTm)
					sleepReport.SleepLight += int64(diff.Seconds())
					hasLight = false
				}
				if !hasDeep {
					beginDeepTm = t
					hasDeep = true
				} else {
					var diff = t.Sub(deepTm)
					sleepReport.SleepDeep += int64(diff.Seconds())
				}
				deepTm = t
			}

		} else {
			if hasSleep {
				sleepReport.SleepNum++
				sleepReport.SleepTimeList = append(sleepReport.SleepTimeList,
					mysql.SleepTime{
						BeginSleepTime: beginSleepTm.Format(cfg.TmFmtStr),
						EndSleepTime:   t.Format(cfg.TmFmtStr),
					})
				hasSleep = false
			}
			if hasLight {
				sleepReport.StagesSleepTime = append(sleepReport.StagesSleepTime,
					mysql.StagesSleepTime{
						StagesStatus:   2,
						BeginSleepTime: beginLightTm.Format(cfg.TmFmtStr),
						EndSleepTime:   t.Format(cfg.TmFmtStr),
					})
				var diff = t.Sub(lightTm)
				sleepReport.SleepLight += int64(diff.Seconds())
				hasLight = false
			}
			if hasDeep {
				sleepReport.StagesSleepTime = append(sleepReport.StagesSleepTime,
					mysql.StagesSleepTime{
						StagesStatus:   3,
						BeginSleepTime: beginDeepTm.Format(cfg.TmFmtStr),
						EndSleepTime:   t.Format(cfg.TmFmtStr),
					})
				var diff = t.Sub(deepTm)
				sleepReport.SleepDeep += int64(diff.Seconds())
				hasDeep = false
			}
			// if awake at first time just remember it
			if !hasAwake {
				beginAwakeTm = t
				hasAwake = true
			} else {
				var diff = t.Sub(awakeTm)
				sleepReport.AwakeLong += int64(diff.Seconds())
			}
			awakeTm = t
			// calculate the leave from bed number at awake time
			if v.ActiveStatus > 0 && v.ActiveStatus < 3 && v.SleepFeatures == 0 {
				if !hasLeaveBed {
					sleepReport.LeaveBedTime = append(sleepReport.LeaveBedTime, t.Format(cfg.TmFmtStr))
					sleepReport.LeaveBedNum++
					hasLeaveBed = true
				}
			} else {
				if hasLeaveBed {
					hasLeaveBed = false
				}
			}
		}
	}

	sleepReport.SleepLong = sleepReport.SleepLight + sleepReport.SleepDeep
	if (sleepReport.SleepLong + sleepReport.AwakeLong) > totalSeconds {
		sleepReport.AwakeLong = totalSeconds - sleepReport.SleepLong
	}
	return http.StatusOK, sleepReport
}

/******************************************************************************
 * function:
 * description:
 * param {string} mac
 * param {string} beginDay
 * param {string} endDay
 * return {*}
********************************************************************************/

func queryEd713TypeSleepReport(mac string, beginDay string, endDay string) (int, interface{}) {
	var dayReport []mysql.Ed713DayReportSql
	mysql.QueryEd713DayReportByMacAndTime(mac, beginDay, endDay, &dayReport)
	if len(dayReport) == 0 {
		return http.StatusAccepted, "not find any data in the condition"
	}
	return handleEd713OrX1Report(mysql.Ed713Type, dayReport)
}
func handleEd713OrX1Report(deviceType string, dayReport interface{}) (int, interface{}) {
	sleepReport := mysql.NewSleepReport()
	var isAwake bool = false
	var isLight bool = false
	var isDeep bool = false
	var isSleep bool = false

	var startTm time.Time
	var endTm time.Time

	// var beginSleepTm time.Time
	var beginLightTm time.Time
	var beginDeepTm time.Time
	var beginAwakeTm time.Time

	var lastCreateTm string = ""

	for i, v := range dayReport.([]mysql.Ed713DayReportSql) {
		t1, err := common.StrToTime(v.SleepStartTime)
		if err != nil {
			mylog.Log.Error("parse time failed, ", err)
			continue
		}
		t2, err := common.StrToTime(v.SleepEndTime)
		if err != nil {
			mylog.Log.Error("parse time failed, ", err)
			continue
		}
		if i == 0 {
			startTm = t1
			endTm = t2
		} else {
			if t1.Before(startTm) {
				startTm = t1
			}
			if t2.After(endTm) {
				endTm = t2
			}
		}
		t, err := common.StrToTime(v.PeriodizationTime)
		if err != nil {
			mylog.Log.Error("parse time failed, ", err)
			continue
		}
		if v.CreateTime != lastCreateTm {
			sleepReport.OnBedTime = v.GoBedTime
			sleepTime := mysql.SleepTime{
				BeginSleepTime: v.SleepStartTime,
				EndSleepTime:   v.SleepEndTime,
			}
			sleepReport.SleepTimeList = append(sleepReport.SleepTimeList, sleepTime)
			sleepReport.SleepNum++
			lastCreateTm = v.CreateTime
		}
		if v.SleepEvents == 1 {
			sleepReport.TurnOver++
		}
		if v.SleepEvents == 2 {
			sleepReport.LeaveBedTime = append(sleepReport.LeaveBedTime, v.SleepEventsTime)
			sleepReport.LeaveBedNum++
		}
		if v.SleepPeriodization == 1 {
			if !isAwake {
				beginAwakeTm = t
				isAwake = true
			}
			if isLight {
				sleepReport.SleepLight += int64(t.Sub(beginLightTm).Seconds())
				sleepReport.StagesSleepTime = append(sleepReport.StagesSleepTime,
					mysql.StagesSleepTime{
						StagesStatus:   2,
						BeginSleepTime: beginLightTm.Format(cfg.TmFmtStr),
						EndSleepTime:   t.Format(cfg.TmFmtStr),
					})
				isLight = false
			}
			if isDeep {
				sleepReport.SleepDeep += int64(t.Sub(beginDeepTm).Seconds())
				sleepReport.StagesSleepTime = append(sleepReport.StagesSleepTime,
					mysql.StagesSleepTime{
						StagesStatus:   3,
						BeginSleepTime: beginDeepTm.Format(cfg.TmFmtStr),
						EndSleepTime:   t.Format(cfg.TmFmtStr),
					})
				isDeep = false
			}
			if isSleep {
				// sleepTime := mysql.SleepTime{
				// 	BeginSleepTime: beginSleepTm.Format(cfg.TmFmtStr),
				// 	EndSleepTime:   t.Format(cfg.TmFmtStr),
				// }
				// sleepReport.SleepTimeList = append(sleepReport.SleepTimeList, sleepTime)
				isSleep = false
			}
		}
		if v.SleepPeriodization > 1 {
			if isAwake {
				sleepReport.AwakeLong += int64(t.Sub(beginAwakeTm).Seconds())
				sleepReport.StagesSleepTime = append(sleepReport.StagesSleepTime,
					mysql.StagesSleepTime{
						StagesStatus:   1,
						BeginSleepTime: beginAwakeTm.Format(cfg.TmFmtStr),
						EndSleepTime:   t.Format(cfg.TmFmtStr),
					})
				isAwake = false
			}
			// sleepReport.SleepNum++
			if !isSleep {
				isSleep = true
				// beginSleepTm = t
			}

			if v.SleepPeriodization == 2 {
				if !isLight {
					isLight = true
					beginLightTm = t
				}
				if isDeep {
					sleepReport.SleepDeep += int64(t.Sub(beginDeepTm).Seconds())
					sleepReport.StagesSleepTime = append(sleepReport.StagesSleepTime,
						mysql.StagesSleepTime{
							StagesStatus:   3,
							BeginSleepTime: beginDeepTm.Format(cfg.TmFmtStr),
							EndSleepTime:   t.Format(cfg.TmFmtStr),
						})
					isDeep = false
				}
			}
			if v.SleepPeriodization == 3 {
				if !isDeep {
					isDeep = true
					beginDeepTm = t
				}
				if isLight {
					sleepReport.SleepLight += int64(t.Sub(beginLightTm).Seconds())
					sleepReport.StagesSleepTime = append(sleepReport.StagesSleepTime,
						mysql.StagesSleepTime{
							StagesStatus:   2,
							BeginSleepTime: beginLightTm.Format(cfg.TmFmtStr),
							EndSleepTime:   t.Format(cfg.TmFmtStr),
						})
					isLight = false
				}
			}
		}
	}
	if isAwake {
		sleepReport.AwakeLong += int64(endTm.Sub(beginAwakeTm).Seconds())
		sleepReport.StagesSleepTime = append(sleepReport.StagesSleepTime,
			mysql.StagesSleepTime{
				StagesStatus:   1,
				BeginSleepTime: beginAwakeTm.Format(cfg.TmFmtStr),
				EndSleepTime:   endTm.Format(cfg.TmFmtStr),
			})
		isAwake = false
	}
	if isLight {
		sleepReport.SleepLight += int64(endTm.Sub(beginLightTm).Seconds())
		sleepReport.StagesSleepTime = append(sleepReport.StagesSleepTime,
			mysql.StagesSleepTime{
				StagesStatus:   2,
				BeginSleepTime: beginLightTm.Format(cfg.TmFmtStr),
				EndSleepTime:   endTm.Format(cfg.TmFmtStr),
			})
		isLight = false
	}
	if isDeep {
		sleepReport.SleepDeep += int64(endTm.Sub(beginDeepTm).Seconds())
		sleepReport.StagesSleepTime = append(sleepReport.StagesSleepTime,
			mysql.StagesSleepTime{
				StagesStatus:   3,
				BeginSleepTime: beginDeepTm.Format(cfg.TmFmtStr),
				EndSleepTime:   endTm.Format(cfg.TmFmtStr),
			})
		isDeep = false
	}
	if isSleep {
		isSleep = false
		// sleepTime := mysql.SleepTime{
		// 	BeginSleepTime: beginSleepTm.Format(cfg.TmFmtStr),
		// 	EndSleepTime:   endTm.Format(cfg.TmFmtStr),
		// }
		// sleepReport.SleepTimeList = append(sleepReport.SleepTimeList, sleepTime)
	}
	sleepReport.SleepLong = sleepReport.SleepLight + sleepReport.SleepDeep
	sleepReport.StartTime = startTm.Format(cfg.TmFmtStr)
	sleepReport.EndTime = endTm.Format(cfg.TmFmtStr)
	return http.StatusOK, sleepReport
}

// 暂时用的是这个函数
func queryX1TypeSleepReport(mac string, beginDay string, endDay string) (int, interface{}) {
	var dayReport []mysql.X1DayReportSql
	mysql.QueryX1DayReportByMacAndTime(mac, beginDay, endDay, &dayReport)
	if len(dayReport) == 0 {
		return http.StatusAccepted, "not find any data in the condition"
	}
	return handleX1TypeSleepReport(dayReport)
}
func handleX1TypeSleepReport(dayReport interface{}) (int, interface{}) {
	sleepReport := mysql.NewSleepReport()
	var isAwake bool = false
	var isLight bool = false
	var isDeep bool = false
	var isSleep bool = false

	var startTm time.Time
	var endTm time.Time

	// var beginSleepTm time.Time
	var beginLightTm time.Time
	var beginDeepTm time.Time
	var beginAwakeTm time.Time

	var lastCreateTm string = ""

	for i, v := range dayReport.([]mysql.X1DayReportSql) {
		t1, err := common.StrToTime(v.SleepStartTime)
		if err != nil {
			mylog.Log.Error("parse time failed, ", err)
			continue
		}
		t2, err := common.StrToTime(v.SleepEndTime)
		if err != nil {
			mylog.Log.Error("parse time failed, ", err)
			continue
		}
		if i == 0 {
			startTm = t1
			endTm = t2
		} else {
			if t1.Before(startTm) {
				startTm = t1
			}
			if t2.After(endTm) {
				endTm = t2
			}
		}
		t, err := common.StrToTime(v.PeriodizationTime)
		if err != nil {
			mylog.Log.Error("parse time failed, ", err)
			continue
		}
		if v.CreateTime != lastCreateTm {
			sleepReport.OnBedTime = v.GoBedTime
			sleepTime := mysql.SleepTime{
				BeginSleepTime: v.SleepStartTime,
				EndSleepTime:   v.SleepEndTime,
			}
			sleepReport.SleepTimeList = append(sleepReport.SleepTimeList, sleepTime)
			sleepReport.SleepNum++
			lastCreateTm = v.CreateTime
		}
		//翻身
		if v.SleepEvents == 1 {
			sleepReport.TurnOver++
		}
		//离床
		if v.SleepEvents == 2 {
			sleepReport.LeaveBedTime = append(sleepReport.LeaveBedTime, v.SleepEventsTime)
			sleepReport.LeaveBedNum++
		}
		// 如果是清醒
		if v.SleepPeriodization == 1 {
			if !isAwake {
				beginAwakeTm = t
				isAwake = true
			}
			if isLight {
				sleepReport.SleepLight += int64(t.Sub(beginLightTm).Seconds())
				sleepReport.StagesSleepTime = append(sleepReport.StagesSleepTime,
					mysql.StagesSleepTime{
						StagesStatus:   2,
						BeginSleepTime: beginLightTm.Format(cfg.TmFmtStr),
						EndSleepTime:   t.Format(cfg.TmFmtStr),
					})
				isLight = false
			}
			if isDeep {
				sleepReport.SleepDeep += int64(t.Sub(beginDeepTm).Seconds())
				sleepReport.StagesSleepTime = append(sleepReport.StagesSleepTime,
					mysql.StagesSleepTime{
						StagesStatus:   3,
						BeginSleepTime: beginDeepTm.Format(cfg.TmFmtStr),
						EndSleepTime:   t.Format(cfg.TmFmtStr),
					})
				isDeep = false
			}
			if isSleep {
				isSleep = false
			}
		}
		// 如果是睡眠包括浅睡和深睡，则处理如下
		if v.SleepPeriodization > 1 {
			// 如果之前是清醒状态，则计算清醒时间
			if isAwake {
				sleepReport.AwakeLong += int64(t.Sub(beginAwakeTm).Seconds())
				sleepReport.StagesSleepTime = append(sleepReport.StagesSleepTime,
					mysql.StagesSleepTime{
						StagesStatus:   1,
						BeginSleepTime: beginAwakeTm.Format(cfg.TmFmtStr),
						EndSleepTime:   t.Format(cfg.TmFmtStr),
					})
				isAwake = false
			}
			// sleepReport.SleepNum++
			if !isSleep {
				isSleep = true
				// beginSleepTm = t
			}
			// 如果现在是浅睡则开始浅睡
			if v.SleepPeriodization == 2 {
				// 如果之前不是浅睡，则记录开始浅睡时间
				if !isLight {
					isLight = true
					beginLightTm = t
				}
				// 如果之前是深睡，则计算深睡时间
				if isDeep {
					sleepReport.SleepDeep += int64(t.Sub(beginDeepTm).Seconds())
					sleepReport.StagesSleepTime = append(sleepReport.StagesSleepTime,
						mysql.StagesSleepTime{
							StagesStatus:   3,
							BeginSleepTime: beginDeepTm.Format(cfg.TmFmtStr),
							EndSleepTime:   t.Format(cfg.TmFmtStr),
						})
					isDeep = false
				}
			}
			// 如果现在是深睡则开始深睡
			if v.SleepPeriodization == 3 {
				// 如果之前不是深睡，则记录开始深睡时间
				if !isDeep {
					isDeep = true
					beginDeepTm = t
				}
				// 如果之前是浅睡，则计算浅睡时间
				if isLight {
					sleepReport.SleepLight += int64(t.Sub(beginLightTm).Seconds())
					sleepReport.StagesSleepTime = append(sleepReport.StagesSleepTime,
						mysql.StagesSleepTime{
							StagesStatus:   2,
							BeginSleepTime: beginLightTm.Format(cfg.TmFmtStr),
							EndSleepTime:   t.Format(cfg.TmFmtStr),
						})
					isLight = false
				}
			}
			// 如果无人时段，则处理
			if v.SleepPeriodization == 0 {
				// 离床+1
				sleepReport.LeaveBedNum++
				sleepReport.LeaveBedTime = append(sleepReport.LeaveBedTime, t.Format(cfg.TmFmtStr))
				if isAwake {
					sleepReport.AwakeLong += int64(t.Sub(beginAwakeTm).Seconds())
					sleepReport.StagesSleepTime = append(sleepReport.StagesSleepTime,
						mysql.StagesSleepTime{
							StagesStatus:   1,
							BeginSleepTime: beginAwakeTm.Format(cfg.TmFmtStr),
							EndSleepTime:   t.Format(cfg.TmFmtStr),
						})
					isAwake = false
				}
				if isLight {
					sleepReport.SleepLight += int64(t.Sub(beginLightTm).Seconds())
					sleepReport.StagesSleepTime = append(sleepReport.StagesSleepTime,
						mysql.StagesSleepTime{
							StagesStatus:   2,
							BeginSleepTime: beginLightTm.Format(cfg.TmFmtStr),
							EndSleepTime:   t.Format(cfg.TmFmtStr),
						})
					isLight = false
				}
				if isDeep {
					sleepReport.SleepDeep += int64(t.Sub(beginDeepTm).Seconds())
					sleepReport.StagesSleepTime = append(sleepReport.StagesSleepTime,
						mysql.StagesSleepTime{
							StagesStatus:   3,
							BeginSleepTime: beginDeepTm.Format(cfg.TmFmtStr),
							EndSleepTime:   t.Format(cfg.TmFmtStr),
						})
					isDeep = false
				}
				if isSleep {
					isSleep = false
				}
			}
		}
	}
	// 循环结束后，如果还有状态未处理，则处理
	if isAwake {
		sleepReport.AwakeLong += int64(endTm.Sub(beginAwakeTm).Seconds())
		sleepReport.StagesSleepTime = append(sleepReport.StagesSleepTime,
			mysql.StagesSleepTime{
				StagesStatus:   1,
				BeginSleepTime: beginAwakeTm.Format(cfg.TmFmtStr),
				EndSleepTime:   endTm.Format(cfg.TmFmtStr),
			})
		isAwake = false
	}
	if isLight {
		sleepReport.SleepLight += int64(endTm.Sub(beginLightTm).Seconds())
		sleepReport.StagesSleepTime = append(sleepReport.StagesSleepTime,
			mysql.StagesSleepTime{
				StagesStatus:   2,
				BeginSleepTime: beginLightTm.Format(cfg.TmFmtStr),
				EndSleepTime:   endTm.Format(cfg.TmFmtStr),
			})
		isLight = false
	}
	if isDeep {
		sleepReport.SleepDeep += int64(endTm.Sub(beginDeepTm).Seconds())
		sleepReport.StagesSleepTime = append(sleepReport.StagesSleepTime,
			mysql.StagesSleepTime{
				StagesStatus:   3,
				BeginSleepTime: beginDeepTm.Format(cfg.TmFmtStr),
				EndSleepTime:   endTm.Format(cfg.TmFmtStr),
			})
		isDeep = false
	}
	if isSleep {
		isSleep = false
	}
	sleepReport.SleepLong = sleepReport.SleepLight + sleepReport.SleepDeep
	sleepReport.StartTime = startTm.Format(cfg.TmFmtStr)
	sleepReport.EndTime = endTm.Format(cfg.TmFmtStr)
	return http.StatusOK, sleepReport
}

// 暂时弃用
func queryX1TypeSleepReport2(mac string, beginDay string, endDay string) (int, interface{}) {
	sleepReport := mysql.NewSleepReport()
	var dayReport []mysql.X1DayReportSql
	mysql.QueryX1DayReportByMacAndTime(mac, beginDay, endDay, &dayReport)
	if len(dayReport) == 0 {
		return http.StatusAccepted, "not find any data in the condition"
	}
	var isAwake bool = false
	var isLight bool = false
	var isDeep bool = false
	var isSleep bool = false
	var awakeTm time.Time
	var lightTm time.Time
	var deepTm time.Time

	var startTm time.Time
	var endTm time.Time

	var beginSleepTm time.Time
	var endSleepTm time.Time

	for i, v := range dayReport {
		t1, err := common.StrToTime(v.InBedStartTime)
		if err != nil {
			mylog.Log.Error("parse time failed, ", err)
			continue
		}
		t2, err := common.StrToTime(v.InBedEndTime)
		if err != nil {
			mylog.Log.Error("parse time failed, ", err)
			continue
		}
		if i == 0 {
			startTm = t1
			endTm = t2
		} else {
			if t1.Before(startTm) {
				startTm = t1
			}
			if t2.After(endTm) {
				endTm = t2
			}
		}
		// 分区睡眠时间
		t, err := common.StrToTime(v.PeriodizationTime)
		if err != nil {
			mylog.Log.Error("parse time failed, ", err)
			continue
		}
		// 翻身
		if v.SleepEvents == 1 {
			sleepReport.TurnOver++
		}
		// 离床
		if v.SleepEvents == 2 {
			sleepReport.LeaveBedNum++
		}
		// 清醒
		if v.SleepPeriodization == 1 {
			// 如果之前不清醒，改为清醒
			if !isAwake {
				isAwake = true
				awakeTm = t
			}
			// 如果之前是浅睡，现在是清醒，计算浅睡时间
			if isLight {
				isLight = false
				sleepReport.SleepLight += int64(t.Sub(lightTm).Seconds())
			}
			// 如果之前是深睡，现在是清醒，计算深睡时间
			if isDeep {
				isDeep = false
				sleepReport.SleepDeep += int64(t.Sub(deepTm).Seconds())
			}
			if isSleep {
				isSleep = false
				endSleepTm = t
				sleepTime := mysql.SleepTime{
					BeginSleepTime: beginSleepTm.Format(cfg.TmFmtStr),
					EndSleepTime:   endSleepTm.Format(cfg.TmFmtStr),
				}
				sleepReport.SleepTimeList = append(sleepReport.SleepTimeList, sleepTime)
			}
		}
		// 浅睡or深睡
		if v.SleepPeriodization > 1 {
			if isAwake {
				isAwake = false
				sleepReport.AwakeLong += int64(t.Sub(awakeTm).Seconds())
			}
			sleepReport.SleepNum++
			if !isSleep {
				isSleep = true
				beginSleepTm = t
			}

			//浅睡
			if v.SleepPeriodization == 2 {
				if !isLight {
					isLight = true
					lightTm = t
				}
				if isDeep {
					isDeep = false
					sleepReport.SleepDeep += int64(t.Sub(deepTm).Seconds())
				}
			}
			if v.SleepPeriodization == 3 {
				if !isDeep {
					isDeep = true
					deepTm = t
				}
				if isLight {
					isLight = false
					sleepReport.SleepLight += int64(t.Sub(lightTm).Seconds())
				}
			}
		}
	}
	if isAwake {
		isAwake = false
		sleepReport.AwakeLong += int64(endTm.Sub(awakeTm).Seconds())
	}
	if isLight {
		isLight = false
		sleepReport.SleepLight += int64(endTm.Sub(lightTm).Seconds())
	}
	if isDeep {
		isDeep = false
		sleepReport.SleepDeep += int64(endTm.Sub(deepTm).Seconds())
	}
	if isSleep {
		isSleep = false
		endSleepTm = endTm
		sleepTime := mysql.SleepTime{
			BeginSleepTime: beginSleepTm.Format(cfg.TmFmtStr),
			EndSleepTime:   endSleepTm.Format(cfg.TmFmtStr),
		}
		sleepReport.SleepTimeList = append(sleepReport.SleepTimeList, sleepTime)
	}
	if sleepReport.LeaveBedNum == 0 {
		sleepReport.LeaveBedNum = 1
	}
	sleepReport.SleepLong = sleepReport.SleepLight + sleepReport.SleepDeep
	sleepReport.StartTime = startTm.Format(cfg.TmFmtStr)
	sleepReport.EndTime = endTm.Format(cfg.TmFmtStr)
	return http.StatusOK, sleepReport
}

// swagger:model QueryDateListResp
type QueryDateListResp struct {
	// required: true
	Mac  string   `json:"mac" mysql:"mac"`
	Days []string `json:"days" mysql:"days"`
}

/******************************************************************************
 * function:
 * description:
 * param {*gin.Context} c
 * return {*}
********************************************************************************/
func QueryDateListInReport(c *gin.Context) (int, interface{}) {
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
	var devices []mysql.Device
	mysql.QueryDeviceByCond(fmt.Sprintf("mac='%s'", mac), nil, "", &devices)
	if len(devices) == 0 {
		return http.StatusAccepted, "not find any device in the condition"
	}
	device := devices[0]
	resp := &QueryDateListResp{}
	resp.Mac = mac
	ok := false
	switch device.Type {
	case mysql.HeatRateType:
		ok = mysql.QueryHeartDateListInReport(mac, beginDay, endDay, &resp.Days)
	case mysql.Ed713Type:
		ok = mysql.QueryEd713DateListInReport(mac, beginDay, endDay, &resp.Days)
	case mysql.X1Type:
		ok = mysql.QueryX1DateListInReport(mac, beginDay, endDay, &resp.Days)
	case mysql.H03Type:
		ok = mysql.QueryH03DateListInReport(mac, beginDay, endDay, &resp.Days)
	}
	if ok {
		return http.StatusOK, resp
	}
	return http.StatusAccepted, "not support device type"
}
