package mysql

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"hjyserver/cfg"
	"hjyserver/exception"
	mylog "hjyserver/log"
	"hjyserver/mdb/common"
	"math"
	"net/http"
	"reflect"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

/*
********************************************************************************

	HeatRateDevice 信息表

********************************************************************************
*/

// swagger:model HeartRate
type HeartRate struct {
	ID            int64  `json:"id" mysql:"id" binding:"omitempty"`
	Mac           string `json:"mac" mysql:"mac" binding:"required"`
	PersonNum     int    `json:"person_num" mysql:"person_num"`
	PersonPos     int    `json:"person_pos" mysql:"person_pos"`
	PersonStatus  int    `json:"person_status" mysql:"person_status"`
	SleepFeatures int    `json:"sleep_features" mysql:"sleep_features"`
	HeartRate     int    `json:"heart_rate" mysql:"heart_rate"`
	BreatheRate   int    `json:"breathe_rate" mysql:"breathe_rate"`
	ActiveStatus  int    `json:"active_status" mysql:"active_status"`
	PhysicalRate  int    `json:"physical_rate" mysql:"physical_rate"`
	StagesStatus  int    `json:"stages_status" mysql:"stages_status"`
	DateTime      string `json:"create_time" mysql:"create_time" binding:"datetime=2006-01-02 15:04:05"`
}

/*
NewHeartRate...
构造实例
*/
func NewHeartRate() *HeartRate {
	return &HeartRate{
		ID:            0,
		Mac:           "",
		PersonNum:     0,
		PersonPos:     0,
		PersonStatus:  0,
		SleepFeatures: 0,
		HeartRate:     0,
		BreatheRate:   0,
		ActiveStatus:  0,
		PhysicalRate:  0,
		StagesStatus:  0,
		DateTime:      time.Now().Format(cfg.TmFmtStr),
	}
}
func (me *HeartRate) MarshalJSON() ([]byte, error) {
	var b []byte
	u := reflect.TypeOf(me)
	vf := reflect.ValueOf(me)
	numField := u.Elem().NumField()
	b = append(b, "{"...)
	for num := 0; num < numField; num++ {
		f := u.Elem().Field(num)
		v := vf.Elem().Field(num)
		switch v.Kind() {
		case reflect.Int64:
			var val string
			if f.Name == "ID" && v.Int() <= 0 {
				val = fmt.Sprintf("\"%v\":\"NaN\"", f.Tag.Get("json"))
			} else {
				val = fmt.Sprintf("\"%v\":\"%v\"", f.Tag.Get("json"), v.Int())
			}
			if num < (numField - 1) {
				val += ","
			}
			b = append(b, val...)
		case reflect.Int:
			var val string
			val = fmt.Sprintf("\"%v\":\"%v\"", f.Tag.Get("json"), v.Int())
			if num < (numField - 1) {
				val += ","
			}
			b = append(b, val...)
		case reflect.Float64:
			var val string
			if math.IsNaN(v.Float()) {
				val = fmt.Sprintf("\"%v\":\"NaN\"", f.Tag.Get("json"))
			} else {
				val = fmt.Sprintf("\"%v\":\"%v\"", f.Tag.Get("json"), v.Float())
			}
			if num < (numField - 1) {
				val += ","
			}
			b = append(b, val...)
		case reflect.String:
			val := fmt.Sprintf("\"%v\":\"%v\"", f.Tag.Get("json"), v.String())
			if num < (numField - 1) {
				val += ","
			}
			b = append(b, val...)
		}
	}
	b = append(b, "}"...)
	return b, nil
}

func (me *HeartRate) MarshalBinary() ([]byte, error) {
	return json.Marshal(me)
}
func (me *HeartRate) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, me)
}

/*
QueryHeartRateByCond...
根据条件查询HeartRate数据
*/
func QueryHeartRateByCond(filter interface{}, page *common.PageDao, sort interface{}, limited int, results *[]HeartRate) bool {
	res := false
	backFunc := func(rows *sql.Rows) {
		obj := NewHeartRate()
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*results = append(*results, *obj)
		}
	}
	if page == nil {
		res = QueryDao(common.DeviceRecordTbl(HeatRateType), filter, sort, limited, backFunc)
	} else {
		res = QueryPage(common.DeviceRecordTbl(HeatRateType), page, filter, sort, backFunc)
	}
	return res
}

/*
Decode 解析从gin获取的数据 转换成HeartRate
*/
func (me *HeartRate) DecodeFromGin(c *gin.Context) {
	if err := c.ShouldBindBodyWith(me, binding.JSON); err != nil {
		exception.Throw(http.StatusAccepted, err.Error())
	}
}

func (me *HeartRate) DecodeFromRows(rows *sql.Rows) error {
	return rows.Scan(&me.ID,
		&me.Mac,
		&me.PersonNum,
		&me.PersonPos,
		&me.PersonStatus,
		&me.SleepFeatures,
		&me.HeartRate,
		&me.BreatheRate,
		&me.ActiveStatus,
		&me.PhysicalRate,
		&me.StagesStatus,
		&me.DateTime)
}
func (me *HeartRate) DecodeFromRow(row *sql.Row) error {
	return row.Scan(&me.ID,
		&me.Mac,
		&me.PersonNum,
		&me.PersonPos,
		&me.PersonStatus,
		&me.SleepFeatures,
		&me.HeartRate,
		&me.BreatheRate,
		&me.ActiveStatus,
		&me.PhysicalRate,
		&me.StagesStatus,
		&me.DateTime)
}

/*
QueryByID() 查询股票实时行情
*/
func (me *HeartRate) QueryByID(id int64) bool {
	me.SetID(id)
	return QueryDaoByID(common.DeviceRecordTbl(HeatRateType), me.ID, me)
}

/*
Insert 股票行情数据插入
*/
func (me *HeartRate) Insert() bool {
	tblName := common.DeviceRecordTbl(HeatRateType)
	if !CheckTableExist(tblName) {
		sql := `create table ` + tblName + ` (
            id MEDIUMINT NOT NULL AUTO_INCREMENT,
            mac varchar(32) not null comment '设备mac,与设备表关联',
            person_num int not null comment '人数',
			person_pos int not null comment '人体位置',
			person_status int not null comment '人体状态',
			sleep_features int not null comment '睡眠特征',
			heart_rate int not null comment '心率',
			breathe_rate int not null comment '呼吸率',
			active_status int not null comment '活动状态',
			physical_rate int not null comment '体态评分',
			stages_status int not null comment '睡眠状态',
            create_time datetime comment '新增日期',
            PRIMARY KEY (id, mac, create_time)
        )`
		CreateTable(sql)
	}
	return InsertDao(tblName, me)
}

/*
Update() 更新指数表
*/
func (me *HeartRate) Update() bool {
	return UpdateDaoByID(common.DeviceRecordTbl(HeatRateType), me.ID, me)
}

/*
Delete() 删除指数
*/
func (me *HeartRate) Delete() bool {
	return DeleteDaoByID(common.DeviceRecordTbl(HeatRateType), me.ID)
}

/*
设置ID
*/
func (me *HeartRate) SetID(id int64) {
	me.ID = id
}

func QueryHeartDateListInReport(mac string, startTime string, endTime string, results *[]string) bool {
	var filter string = fmt.Sprintf("mac='%s' and date(create_time)>=date('%s') and date(create_time)<=date('%s')", mac, startTime, endTime)
	sql := fmt.Sprintf("select distinct date(create_time) from %s where %s order by date(create_time)", common.DeviceRecordTbl(HeatRateType), filter)
	rows, err := mDb.Query(sql)
	if err != nil {
		mylog.Log.Errorln(err)
		return false
	}
	defer rows.Close()
	for rows.Next() {
		var date string
		err := rows.Scan(&date)
		if err != nil {
			mylog.Log.Errorln(err)
			return false
		}
		*results = append(*results, date)
	}
	return true
}

/******************************************************************************
 * function:
 * description: define sleep report struct
 * return {*}
********************************************************************************/

// swagger:model SleepTime
type SleepTime struct {
	BeginSleepTime string `json:"begin_sleep_time"`
	EndSleepTime   string `json:"end_sleep_time"`
}

// swagger:model StagesSleepTime
type StagesSleepTime struct {
	StagesStatus   int    `json:"stages_status"` // 0: awake, 1: light sleep, 2: deep sleep
	BeginSleepTime string `json:"begin_sleep_time"`
	EndSleepTime   string `json:"end_sleep_time"`
}

// swagger:model SleepReport
type SleepReport struct {
	SleepLong       int64             `json:"sleep_long"`
	SleepLight      int64             `json:"sleep_light"`
	SleepDeep       int64             `json:"sleep_deep"`
	AwakeLong       int64             `json:"awake_long"`
	StartTime       string            `json:"start_time"`
	EndTime         string            `json:"end_time"`
	SleepTimeList   []SleepTime       `json:"sleep_time"`
	StagesSleepTime []StagesSleepTime `json:"stages_sleep_time"`
	SleepNum        int               `json:"sleep_num"`
	TurnOver        int               `json:"turn_over"`
	LeaveBedNum     int               `json:"leave_bed_num"`
	LeaveBedTime    []string          `json:"leave_bed_time"`
	OnBedTime       string            `json:"on_bed_time"`
}

func NewSleepReport() *SleepReport {
	return &SleepReport{
		SleepLong:       0,
		SleepLight:      0,
		SleepDeep:       0,
		AwakeLong:       0,
		StartTime:       "",
		EndTime:         "",
		SleepTimeList:   []SleepTime{},
		StagesSleepTime: []StagesSleepTime{},
		SleepNum:        0,
		TurnOver:        0,
		LeaveBedNum:     0,
		LeaveBedTime:    []string{},
	}
}
