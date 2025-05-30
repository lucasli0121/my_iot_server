/*
 * @Author: liguoqiang
 * @Date: 2021-03-07 09:34:20
 * @LastEditors: liguoqiang
 * @LastEditTime: 2024-04-28 17:32:05
 * @Description: 实现 数据库的主函数, 连接mysql 操作
 */

package mysql

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"

	"hjyserver/cfg"
	"hjyserver/gopool"
	mylog "hjyserver/log"
	"hjyserver/mdb/common"
	"hjyserver/mq"
	"hjyserver/redis"
	"hjyserver/sms"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

/*
* MysqlDao... mysql所有数据对象的基类
 */
type Dao interface {
	SetID(int64)
	QueryByID(int64) bool
	Insert() bool
	Update() bool
	Delete() bool
	DecodeFromGin(c *gin.Context)
	DecodeFromRow(row *sql.Row) error
	DecodeFromRows(rows *sql.Rows) error
}

// sleep device alarm event struct
// type 3001 3006 3007 3008 3012
// swagger:model HeartEvent
type HeartEvent struct {
	UserDeviceDetail
	Type            int    `json:"type"`
	HeartRate       int    `json:"heart_rate"`
	RespiratoryRate int    `json:"respiratory_rate"`
	CreateTime      string `json:"create_time"`
}

type HeartBeatMsg struct {
	Mac    string `json:"mac"`
	Online int    `json:"online"`
	Rssi   int    `json:"rssi"`
}

/*
**********************************************************************************

	自定义一个float64类型，用来处理字符串和float转换时null或者nan的情况

**********************************************************************************
*/
type float64JSON float64

func (me *float64JSON) UnmarshalJSON(b []byte) error {
	b = bytes.Trim(b, "\"")
	strval := strings.ToLower(string(b))
	if strval == "nan" || strval == "null" {
		*me = 0.0
	} else {
		val, err := strconv.ParseFloat(strval, 64)
		if err != nil {
			return err
		}
		*me = float64JSON(val)
	}
	return nil
}

var mDb *sql.DB = nil
var taskPool *gopool.Pool = nil
var quit chan bool = make(chan bool)

/******************************************************************************
 * function: Open
 * description: open mysql connection, must first run at main function
 * return {*}
********************************************************************************/
func Open() bool {
	dsn := cfg.This.DB.Username + ":" + cfg.This.DB.Password + "@" + cfg.This.DB.Url + "/" + cfg.This.DB.Dbname
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		mylog.Log.Errorln("open mysql driver error:", err)
		return false
	}
	/* 连接数据库 */
	err = db.Ping()
	if err != nil {
		mylog.Log.Errorln("ping to mysql error:", err)
		return false
	}
	mDb = db
	mDb.SetConnMaxLifetime(time.Second * 30) // 每个连接最大存活时间
	mDb.SetConnMaxIdleTime(time.Second * 30) // 每个连接最大空闲时间
	mDb.SetMaxIdleConns(500)                 // 最大空闲连接数
	mDb.SetMaxOpenConns(1024)                // 连接池最大连接数
	// init task pool
	taskPool, _ = gopool.InitPool(128)
	// subscribe device topic
	subscribeDeviceTopic()
	// open a goroutine to check whether device is online
	checkOnlineTimeout := make(chan struct{}, 1)
	checkDataTimeout := make(chan struct{}, 1)
	cleanupOldRealData := make(chan struct{}, 1)
	go func() {
		for {
			time.Sleep(1 * time.Minute)
			checkOnlineTimeout <- struct{}{}
		}
	}()
	go func() {
		for {
			time.Sleep(30 * time.Second)
			checkDataTimeout <- struct{}{}
		}
	}()
	go func() {
		for {
			time.Sleep(24 * time.Hour)
			cleanupOldRealData <- struct{}{}
		}
	}()

	go func() {
		for {
			select {
			case <-quit:
				mylog.Log.Errorln("quit check device online")
				return
			case <-checkOnlineTimeout:
				checkDeviceOnline()
				askAllRealData()
			case <-checkDataTimeout:
				if cfg.This.Svr.EnableHl77 {
					checkNoRealDataLamp()
				}
			case <-cleanupOldRealData:
				cleanupOldRealDataTbl()
			}
		}
	}()

	if cfg.This.Svr.EnableX1 {
		go func() {
			for {
				NotifyTask()
				time.Sleep(30 * time.Second)
			}
		}()
	}
	return true
}

/******************************************************************************
 * function: Close
 * description: close mysql connection, must run at main function end
 * return {*}
********************************************************************************/
func Close() {
	quit <- true
	taskPool.Close()
	err := mDb.Close()
	if err != nil {
		mylog.Log.Errorln(err)
	}
}

func GetTaskPool() *gopool.Pool {
	return taskPool
}
func GetDB() *sql.DB {
	return mDb
}

/******************************************************************************
 * function: checkDeviceOnline
 * description: check device online, if device is offline, set online=0
 * return {*}
********************************************************************************/
func checkDeviceOnline() {
	// sql := "update " + common.DeviceTbl + " set online=0 where online=1 and online_time<?"
	// result, err := mDb.Exec(sql, time.Now().Add(-6*time.Minute).Format(cfg.TmFmtStr))
	// mylog.Log.Debugln("check device online, sql:", sql)
	// if err != nil {
	// 	mylog.Log.Errorln(err)
	// 	return
	// }
	// _, err = result.RowsAffected()
	// if err != nil {
	// 	mylog.Log.Errorln(err)
	// 	return
	// }
	tm := time.Now().Add(-6 * time.Minute).Format(cfg.TmFmtStr)
	if cfg.This.Svr.EnableH03 || cfg.This.Svr.EnableT1 {
		tm = time.Now().Add(-1 * time.Minute).Format(cfg.TmFmtStr)
	}
	filter := fmt.Sprintf("online=1 and online_time<='%s'", tm)
	var gList = []Device{}
	QueryDeviceByCond(filter, nil, "", &gList)
	for _, v := range gList {
		v.Online = 0
		v.Update()
		status := HeartBeatMsg{Mac: v.Mac, Online: 0, Rssi: v.Rssi}
		mq.PublishData(common.MakeDeviceHeartBeatTopic(v.Mac), status)
	}
}

/**
 * @description:设置设备在线状态
 * @param {string} mac
 * @param {int} online
 * @param {int} rssi
 * @return {*}
 */
func SetDeviceOnline(mac string, online int, rssi int) {
	status := HeartBeatMsg{Mac: mac, Online: online, Rssi: rssi}
	GetTaskPool().Put(&gopool.Task{
		Params: []interface{}{&status},
		Do: func(params ...interface{}) {
			var obj = params[0].(*HeartBeatMsg)
			var gList = []Device{}
			QueryDeviceByCond(fmt.Sprintf("mac='%s'", obj.Mac), nil, nil, &gList)
			if len(gList) > 0 {
				gList[0].Online = obj.Online
				if rssi != 0 {
					gList[0].Rssi = obj.Rssi
				}
				gList[0].OnlineTime = time.Now().Format(cfg.TmFmtStr)
				gList[0].Update()
			}
		},
	})
	mq.PublishData(common.MakeDeviceHeartBeatTopic(mac), status)
}

/******************************************************************************
 * function:
 * description:
 * return {*}
********************************************************************************/
func checkNoRealDataLamp() {
	var curTm = time.Now()
	var beforTm = curTm.Add(-30 * time.Second)
	tmFilter := fmt.Sprintf(" create_time>'%s' and create_time<'%s'", beforTm.Format(cfg.TmFmtStr), curTm.Format(cfg.TmFmtStr))
	filter := fmt.Sprintf("(flow_state = 0 or flow_state = 3) and heart_rate = 0 and %s", tmFilter)
	sql := "select distinct a.mac, ifnull(b.cnt, 0) from device_tbl a left join (select mac, count(*) as cnt from " +
		common.LampRealDataTbl + " where " + filter + " group by mac) b on a.mac=b.mac where a.type like 'lamp_type'"
	mylog.Log.Debugln(sql)
	rows, err := mDb.Query(sql)
	if err != nil {
		mylog.Log.Errorln(err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var mac string
		var cnt int
		err := rows.Scan(&mac, &cnt)
		if err != nil {
			mylog.Log.Errorln(err)
			continue
		}
		if cnt == 0 || cnt > 2 {
			TakeLampUserFromStudyRoom(mac)
		}
	}
}

/******************************************************************************
 * function:
 * description:
 * return {*}
********************************************************************************/
func askAllRealData() {
	var devices = []Device{}
	QueryDeviceByCond("online=1", nil, "online_time desc", &devices)
	for _, v := range devices {
		filter := fmt.Sprintf("mac='%s'", v.Mac)
		var curTm = time.Now()
		switch v.Type {
		case X1Type:
			if cfg.This.Svr.EnableX1 {
				var objs = []X1RealDataMysql{}
				QueryX1RealDataByCond(filter, nil, "create_time desc", 1, &objs)
				if len(objs) > 0 {
					obj := objs[0]
					t, err := common.StrToTime(obj.CreateTime)
					if err != nil {
						mylog.Log.Error(err)
						continue
					}
					diff := curTm.Sub(t)
					if int64(diff.Minutes()) >= 10 {
						AskX1RealData(obj.Mac, 6, 1)
					}
				} else {
					AskX1RealData(v.Mac, 6, 1)
				}
			}
		case Ed713Type:
			if cfg.This.Svr.EnableEd713 {
				var objs = []Ed713RealDataMysql{}
				QueryEd713RealDataByCond(filter, nil, "create_time desc", 1, &objs)
				if len(objs) > 0 {
					obj := objs[0]
					t, err := common.StrToTime(obj.CreateTime)
					if err != nil {
						mylog.Log.Error(err)
						continue
					}
					diff := curTm.Sub(t)
					if int64(diff.Minutes()) >= 10 {
						AskEd713RealData(obj.Mac, 6, 1)
					}
				} else {
					AskEd713RealData(v.Mac, 6, 1)
				}
			}
		case LampType:
			if cfg.This.Svr.EnableHl77 {
				var objs = []RealDataSql{}
				QueryLampRealDataByCond(filter, nil, "create_time desc", 1, &objs)
				if len(objs) > 0 {
					obj := objs[0]
					t, err := common.StrToTime(obj.CreateTime)
					if err != nil {
						mylog.Log.Error(err)
						continue
					}
					diff := curTm.Sub(t)
					if int64(diff.Minutes()) >= 10 {
						AskHl77RealData(obj.Mac, 6, 1)
					}
				} else {
					AskHl77RealData(v.Mac, 6, 1)
				}
			}
		}
	}
}

/******************************************************************************
 * function: cleanupOldRealDataTbl
 * description: clean up expired real data exceed 24 hours
 * return {*}
********************************************************************************/
func cleanupOldRealDataTbl() {
	// cleanup lamp table
	var tmDiff = time.Now().Add(-24 * 30 * time.Hour).Format(cfg.TmFmtStr)
	sql := "delete from " + common.LampRealDataTbl + " where create_time<?"
	_, err := mDb.Exec(sql, tmDiff)
	if err != nil {
		mylog.Log.Errorln(err)
	}
	// cleanup x1 table
	sql = "delete from " + common.DeviceRecordTbl(X1Type) + " where create_time<?"
	_, err = mDb.Exec(sql, tmDiff)
	if err != nil {
		mylog.Log.Errorln(err)
	}
	// cleanup ed713 table
	sql = "delete from " + common.DeviceRecordTbl(Ed713Type) + " where create_time<?"
	_, err = mDb.Exec(sql, tmDiff)
	if err != nil {
		mylog.Log.Errorln(err)
	}
	// cleanup x1 json table
	sql = "delete from " + common.DeviceDayReportJsonTbl(X1Type) + " where create_time<?"
	_, err = mDb.Exec(sql, tmDiff)
	if err != nil {
		mylog.Log.Errorln(err)
	}
	// cleanup h03 attr old data
	sql = "delete from " + H03AttrData{}.TableName() + " where create_time<?"
	_, err = mDb.Exec(sql, tmDiff)
	if err != nil {
		mylog.Log.Errorln(err)
	}
	// clean h03 event old data
	sql = "delete from " + H03Event{}.TableName() + " where create_time<?"
	_, err = mDb.Exec(sql, tmDiff)
	if err != nil {
		mylog.Log.Errorln(err)
	}
}

/******************************************************************************
 * function: subscribeDeviceTopic
 * description: query device which type is lamp_type record and subscribe mac as topic
 * return {*}
********************************************************************************/
func subscribeDeviceTopic() {
	var filter string
	var results []Device
	// subscribe device topic
	// subscribe hl77 topic if enable hl77
	if cfg.This.Svr.EnableHl77 {
		filter = fmt.Sprintf("type='%s'", LampType)
		QueryDeviceByCond(filter, nil, "create_time desc", &results)
		if len(results) > 0 {
			for _, v := range results {
				mq.SubscribeTopic(MakeHl77DeliverTopicByMac(v.Mac), NewLampMqttMsgProc())
			}
		}
	}
	results = nil
	// subscribe ed713 topic if enable ed713
	if cfg.This.Svr.EnableEd713 {
		filter = fmt.Sprintf("type='%s'", Ed713Type)
		QueryDeviceByCond(filter, nil, "create_time desc", &results)
		if len(results) > 0 {
			for _, v := range results {
				SubscribeEd713MqttTopic(v.Mac)
			}
		}
	}
	results = nil
	// subscribe x1 topic if enable x1
	if cfg.This.Svr.EnableX1 {
		filter = fmt.Sprintf("type='%s'", X1Type)
		QueryDeviceByCond(filter, nil, "create_time desc", &results)
		if len(results) > 0 {
			for _, v := range results {
				SubscribeX1MqttTopic(v.Mac)
			}
		}
	}
	results = nil
	// subscribe h03 topic if enable h03
	if cfg.This.Svr.EnableH03 {
		filter = fmt.Sprintf("type='%s'", H03Type)
		QueryDeviceByCond(filter, nil, "create_time desc", &results)
		if len(results) > 0 {
			for _, v := range results {
				SubscribeH03MqttTopic(v.Mac)
			}
		}
	}
	results = nil
	// subscribe t1 topic if enable t1
	if cfg.This.Svr.EnableT1 {
		filter = fmt.Sprintf("type='%s'", T1Type)
		QueryDeviceByCond(filter, nil, "create_time desc", &results)
		if len(results) > 0 {
			for _, v := range results {
				SubscribeT1MqttTopic(v.Mac)
			}
		}
	}
	results = nil
	// subscribe x1s topic if enable x1s
	if cfg.This.Svr.EnableX1s {
		SubscribeX1sWildcardTopic()
		filter = fmt.Sprintf("type='%s'", X1sType)
		QueryDeviceByCond(filter, nil, "create_time desc", &results)
		if len(results) > 0 {
			for _, v := range results {
				SubscribeX1sMqttTopic(v.Mac)
			}
		}
	}
}

func UnsubscribeDeviceTopic(mac string) {
	var results []Device
	var filter string
	// subscribe h03 topic if enable h03
	if cfg.This.Svr.EnableH03 {
		if mac != "" {
			filter = fmt.Sprintf("type='%s' and mac like '%s' ", H03Type, mac)
		} else {
			filter = fmt.Sprintf("type='%s'", H03Type)
		}
		QueryDeviceByCond(filter, nil, "create_time desc", &results)
		if len(results) > 0 {
			for _, v := range results {
				UnsubscribeH03MqttTopic(v.Mac)
			}
		}
	}
	if cfg.This.Svr.EnableT1 {
		if mac != "" {
			filter = fmt.Sprintf("type='%s' and mac like '%s' ", T1Type, mac)
		} else {
			filter = fmt.Sprintf("type='%s'", T1Type)
		}
		QueryDeviceByCond(filter, nil, "create_time desc", &results)
		if len(results) > 0 {
			for _, v := range results {
				UnsubscribeT1MqttTopic(v.Mac)
			}
		}
	}
	results = nil
	// subscribe x1s topic if enable x1s
	if cfg.This.Svr.EnableX1s {
		UnsubscribeX1sWildcardTopic()
		if mac != "" {
			filter = fmt.Sprintf("type='%s' and mac like '%s' ", X1sType, mac)
		} else {
			filter = fmt.Sprintf("type='%s'", X1sType)
		}
		QueryDeviceByCond(filter, nil, "create_time desc", &results)
		if len(results) > 0 {
			for _, v := range results {
				UnsubscribeX1sMqttTopic(v.Mac)
			}
		}
	}
}

/********************************************************************
* 分页查询功能
* 通过limit, skip 实现简单分页
* pageNo==1时返回总页数
********************************************************************/
func QueryPage(table string, page *common.PageDao, filter interface{}, sort interface{}, cb func(*sql.Rows)) bool {
	// 先获取总记录数，计算总页数
	totalPages := int64(0)
	totalCount := int64(0)
	countSql := fmt.Sprintf("select count(*) from %s", table)
	if filter != nil && len(filter.(string)) > 0 {
		countSql += " where " + filter.(string)
	}
	row := mDb.QueryRow(countSql)
	err := row.Scan(&totalCount)
	if err != nil {
		mylog.Log.Errorln(err)
		return false
	}
	if page.PageSize <= 0 {
		page.PageSize = 10
	}
	totalPages = int64(float32(totalCount)/float32(page.PageSize) + float32(0.9))
	page.TotalPages = totalPages
	// 根据页数查询数据
	if page.PageNo <= 0 {
		page.PageNo = 1
	} else if page.PageNo > totalPages {
		page.PageNo = totalPages
	}
	sql := "select * from " + table
	if filter != nil && len(filter.(string)) > 0 {
		sql += " where " + filter.(string)
	}
	if sort != nil && len(sort.(string)) > 0 {
		sql += " order by " + sort.(string)
	}
	sql += fmt.Sprintf(" limit %d offset %d", page.PageSize, page.PageNo-1)
	rows, err := mDb.Query(sql)
	if err != nil {
		mylog.Log.Errorln(err)
		return false
	}
	defer rows.Close()
	for rows.Next() {
		cb(rows)
	}
	return true
}

/*
 * func Query, support method for any query
 *
 */
func QueryDao(table string, filter interface{}, sort interface{}, limited int, cb func(*sql.Rows)) bool {
	if !CheckTableExist(table) {
		return false
	}
	sql := "select * from " + table
	if filter != nil && len(filter.(string)) > 0 {
		sql += " where " + filter.(string)
	}
	if sort != nil && len(sort.(string)) > 0 {
		sql += " order by " + sort.(string)
	}
	if limited > 0 {
		sql += " limit " + strconv.FormatInt(int64(limited), 10)
	}
	mylog.Log.Debugln(sql)
	rows, err := mDb.Query(sql)
	if err != nil {
		mylog.Log.Errorln(err)
		return false
	}
	defer rows.Close()
	for rows.Next() {
		cb(rows)
	}
	return true
}

// Find one by ID
func QueryDaoByID(table string, id int64, obj Dao) bool {
	sql := "select * from " + table + " where id=?"
	row := mDb.QueryRow(sql, id)
	err := obj.DecodeFromRow(row)
	if err != nil {
		mylog.Log.Errorln(err)
		return false
	}
	return true
}

/******************************************************************************
 * function: QueryFirstByCond
 * description: query only one record by condition
 * param {string} table
 * param {string} filter
 * param {string} sort
 * param {Dao} obj
 * return {*}
********************************************************************************/
func QueryFirstByCond(table string, filter string, sort string, obj Dao) bool {
	sql := "select * from " + table
	if len(filter) > 0 {
		sql += " where " + filter
	}
	if len(sort) > 0 {
		sql += " order by " + sort
	}
	sql += " limit 1"
	row := mDb.QueryRow(sql)
	err := obj.DecodeFromRow(row)
	if err != nil {
		mylog.Log.Errorln(err)
		return false
	}
	return true
}

func CheckTableExist(tblName string) bool {
	sql := fmt.Sprintf("show tables like '%%%s%%'", tblName)
	rows, err := mDb.Query(sql)
	if err != nil {
		mylog.Log.Errorln(err)
		return false
	}
	defer rows.Close()
	var table string
	for rows.Next() {
		err := rows.Scan(&table)
		if err != nil {
			mylog.Log.Errorln(err)
		} else if strings.EqualFold(table, tblName) {
			return true
		}
	}
	return false
}

func CreateTable(sql string) error {
	_, err := mDb.Exec(sql)
	return err
}

func CreateTableWithStruct(tblName string, obj interface{}) bool {
	sql := fmt.Sprintf(`create table if not exists %s (`, tblName)
	var fields string
	var keys string = "primary key("
	var unique string
	u := reflect.TypeOf(obj)
	numField := u.Elem().NumField()
	for num := 0; num < numField; num++ {
		f := u.Elem().Field(num)
		tag := f.Tag.Get("mysql")
		if tag == "" {
			continue
		}
		if len(fields) > 0 {
			fields += `,`
		}
		common := f.Tag.Get("common")
		fields += tag
		if tag == "id" {
			fields += " MEDIUMINT not null auto_increment "
			keys += "id"
		} else if f.Type.String() == "time.Time" {
			fields += " datetime"
		} else {
			switch f.Type.Kind() {
			case reflect.Int:
				fields += " int"
			case reflect.Int64:
				fields += " bigint"
			case reflect.Float32:
				fields += " float"
			case reflect.Float64:
				fields += " double"
			case reflect.String:
				binding := f.Tag.Get("binding")
				if len(binding) > 0 && len(strings.Split(binding, "=")) > 1 {
					v1 := strings.Split(binding, "=")[0]
					if v1 == "datetime" {
						fields += " datetime"
					} else if v1 == "date" {
						fields += " date"
					} else if v1 == "time" {
						fields += " time"
					}
				} else {
					if f.Tag.Get("size") != "" {
						fields += " varchar(" + f.Tag.Get("size") + ")"
					} else {
						fields += " varchar(255)"
					}
				}
			case reflect.Pointer:
				if f.Type.Elem().Kind() == reflect.String {
					binding := f.Tag.Get("binding")
					if len(binding) > 0 && len(strings.Split(binding, "=")) > 1 {
						v1 := strings.Split(binding, "=")[0]
						if v1 == "datetime" {
							fields += " datetime"
						} else if v1 == "date" {
							fields += " date"
						} else if v1 == "time" {
							fields += " time"
						}
					} else {
						if f.Tag.Get("size") != "" {
							fields += " varchar(" + f.Tag.Get("size") + ")"
						} else {
							fields += " varchar(255)"
						}
					}
				}
			case reflect.Array:
				fallthrough
			case reflect.Slice:
				if f.Tag.Get("size") != "" {
					fields += " varchar(" + f.Tag.Get("size") + ")"
				} else {
					fields += " varchar(255)"
				}
			}
			if f.Tag.Get("isnull") == "false" || f.Tag.Get("binding") == "required" {
				fields += " not null"
			} else if f.Tag.Get("isnull") == "true" {
				fields += " null"
			}

			if f.Tag.Get("default") != "" {
				fields += " default " + f.Tag.Get("default")
			}
			if f.Tag.Get("unique") == "true" {
				if len(unique) > 0 {
					unique += ","
				}
				unique += tag
			}
			if f.Tag.Get("key") == "true" {
				keys += "," + tag
			}
		}
		if len(common) > 0 {
			fields += " comment '" + common + "'"
		}
	}
	sql += fields + "," + keys + ")"
	if len(unique) > 0 {
		sql += ", constraint " + tblName + "_unique unique(" + unique + ")"
	}
	sql += ") DEFAULT CHARSET=utf8;"
	_, err := mDb.Exec(sql)
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

/*
* insert...
 */
func InsertDao(tblName string, obj Dao) bool {
	sql := fmt.Sprintf("insert into %s ", tblName)
	u := reflect.TypeOf(obj)
	vf := reflect.ValueOf(obj)
	var fields string
	var values string
	numField := u.Elem().NumField()
	for num := 0; num < numField; num++ {
		f := u.Elem().Field(num)
		v := vf.Elem().Field(num)
		if f.Tag.Get("mysql") == "" {
			continue
		}
		if len(fields) > 0 {
			fields += ","
		}
		if len(values) > 0 {
			values += ","
		}
		if f.Tag.Get("mysql") != "id" {
			fields += f.Tag.Get("mysql")
		}
		switch v.Kind() {
		case reflect.Int64:
			if f.Name != "ID" {
				values += fmt.Sprintf("%d", v.Int())
			}
		case reflect.Int:
			values += fmt.Sprintf("%d", v.Int())
		case reflect.Float32:
			fallthrough
		case reflect.Float64:
			if math.IsNaN(v.Float()) {
				values += "NULL"
			} else {
				values += fmt.Sprintf("%v", v.Float())
			}
		case reflect.String:
			values += "'" + v.String() + "'"
		case reflect.Pointer:
			if v.IsNil() {
				values += "NULL"
			} else {
				if f.Type.Elem().Kind() == reflect.String {
					values += "'" + v.Elem().String() + "'"
				} else if f.Type.Elem().Kind() == reflect.Array {
					values += "'" + v.Elem().String() + "'"
				}
			}
		case reflect.Array:
			fallthrough
		case reflect.Slice:
			str := ""
			for i := 0; i < v.Len(); i++ {
				elemVal := v.Index(i)
				s := ""
				if elemVal.Kind() == reflect.String {
					s = elemVal.String()
				} else if elemVal.Kind() == reflect.Int {
					s = fmt.Sprintf("%d", elemVal.Int())
				} else if elemVal.Kind() == reflect.Float64 {
					s = fmt.Sprintf("%v", elemVal.Float())
				}
				if i == 0 {
					str = s
				} else {
					str += "," + s
				}
			}
			values += "'" + str + "'"
		}
	}
	sql += fmt.Sprintf(" (%s) values (%s)", fields, values)
	result, err := mDb.Exec(sql)
	if err != nil {
		mylog.Log.Errorln(err)
		mylog.Log.Errorln(sql)
		return false
	}
	id, err := result.LastInsertId()
	if err != nil {
		mylog.Log.Errorln(err)
		return false
	}
	obj.SetID(id)
	return true
}

/*
* updateDaoById...
 */
func UpdateDaoByID(tblName string, id int64, obj Dao) bool {
	sql := fmt.Sprintf("update %s ", tblName)
	u := reflect.TypeOf(obj)
	vf := reflect.ValueOf(obj)
	var setsql string
	numField := u.Elem().NumField()
	for num := 0; num < numField; num++ {
		f := u.Elem().Field(num)
		v := vf.Elem().Field(num)
		var setval string
		if f.Tag.Get("mysql") == "" {
			continue
		}
		if f.Tag.Get("mysql") != "id" {
			setval = fmt.Sprintf(" %s=", f.Tag.Get("mysql"))
		}
		switch v.Kind() {
		case reflect.Int64:
			if f.Name != "ID" {
				setval += fmt.Sprintf("%d", v.Int())
			}
		case reflect.Float32:
			fallthrough
		case reflect.Float64:
			if math.IsNaN(v.Float()) {
				setval += "NULL"
			} else {
				setval += fmt.Sprintf("%v", v.Float())
			}
		case reflect.Int:
			setval += fmt.Sprintf("%d", v.Int())
		case reflect.String:
			setval += "'" + v.String() + "'"
		case reflect.Pointer:
			if v.IsNil() {
				setval += "NULL"
			} else {
				if f.Type.Elem().Kind() == reflect.String {
					setval += "'" + v.Elem().String() + "'"
				}
			}
		case reflect.Array:
			fallthrough
		case reflect.Slice:
			str := ""
			for i := 0; i < v.Len(); i++ {
				elemVal := v.Index(i)
				s := ""
				if elemVal.Kind() == reflect.String {
					s = elemVal.String()
				} else if elemVal.Kind() == reflect.Int {
					s = fmt.Sprintf("%d", elemVal.Int())
				} else if elemVal.Kind() == reflect.Float64 {
					s = fmt.Sprintf("%v", elemVal.Float())
				}
				if i == 0 {
					str = s
				} else {
					str += "," + s
				}
			}
			setval += "'" + str + "'"
		}
		if len(setsql) > 0 {
			setsql += "," + setval
		} else {
			setsql = setval
		}
	}
	sql += fmt.Sprintf(" set %s where id=%d", setsql, id)
	result, err := mDb.Exec(sql)
	if err != nil {
		mylog.Log.Errorln(err)
		return false
	}
	count, err := result.RowsAffected()
	if err != nil {
		mylog.Log.Errorln(err)
		return false
	}
	mylog.Log.Debugln("Update table:", tblName, ", and affected rows:", count)
	return true
}

/*
* deleteDaoByID...
 */
func DeleteDaoByID(tblName string, id int64) bool {
	sql := fmt.Sprintf("delete from %s where id=%d", tblName, id)
	result, err := mDb.Exec(sql)
	if err != nil {
		mylog.Log.Errorln(err)
		return false
	}
	count, err := result.RowsAffected()
	if err != nil {
		mylog.Log.Errorln(err)
		return false
	}
	mylog.Log.Debugln("Delete table:", tblName, " count:", count)
	return true
}

/******************************************************************************
 * function: DeleteDaoByFilter
 * description:
 * return {*}
********************************************************************************/
func DeleteDaoByFilter(tblName string, filter string) bool {
	sql := fmt.Sprintf("delete from %s where %s", tblName, filter)
	result, err := mDb.Exec(sql)
	if err != nil {
		mylog.Log.Errorln(err)
		return false
	}
	count, err := result.RowsAffected()
	if err != nil {
		mylog.Log.Errorln(err)
		return false
	}
	mylog.Log.Debugln("Delete table:", tblName, " count:", count)
	return true
}

/******************************************************************************
 * function: SendSleepAlarmSms
 * description: send alarm sms to user
 * param {*UserDeviceAlarm} userAlarm
 * return {*}
********************************************************************************/
func SendSleepAlarmSms(userEvent *HeartEvent) {
	if userEvent == nil || userEvent.EmergentPhone == "" {
		return
	}
	alarmDesc := common.GetSleepAlarmDesc(userEvent.EmergentPhone, userEvent.Type)
	if len(alarmDesc) == 0 {
		return
	}
	if userEvent.Type == 3001 || userEvent.Type == 3012 {
		return
	}
	err := sms.SendSms(userEvent.EmergentPhone, userEvent.NickName, alarmDesc)
	if err != nil {
		mylog.Log.Errorln(err)
	}
}

/******************************************************************************
 * function: SendSleepNotifySms
 * description: send notify sms to user
 * param {string} mac
 * param {int} notifyType
 * param {int} status
 * return {*}
********************************************************************************/
func SendSleepNotifySms(mac string, notifyType int, status int) {
	var userDevices []UserDeviceDetail
	QueryUserDeviceDetailByMac(mac, &userDevices)
	if len(userDevices) > 0 {
		phone := userDevices[0].EmergentPhone
		nickName := userDevices[0].NickName
		desc := common.GetNotifyStatusDesc(phone, notifyType, status)
		if len(desc) == 0 {
			return
		}
		err := sms.SendSms(phone, nickName, desc)
		if err != nil {
			mylog.Log.Errorln(err)
		}
	}
}

/******************************************************************************
 * function:CheckDiffBetweenTwoSleepDeviceRecords
 * description: check differance between two real data from sleep devices
 * param {string} deviceType
 * param {string} mac
 * param {interface{}} obj
 * return {*}
********************************************************************************/
func CheckDiffBetweenTwoSleepDeviceRecords(deviceType string, mac string, obj *HeartRate) bool {
	result := true
	oldObj := &HeartRate{}
	err := redis.GetValueFromHash(deviceType, mac, true, oldObj)
	if err == nil {
		result = oldObj.HeartRate != obj.HeartRate ||
			oldObj.BreatheRate != obj.BreatheRate ||
			oldObj.ActiveStatus != obj.ActiveStatus ||
			oldObj.PersonStatus != obj.PersonStatus
	}
	t := time.Now().Add(time.Minute * 2)
	redis.SaveValueToHash(deviceType, mac, &t, obj)
	return result
}

/******************************************************************************
 * function: CheckDiffBetweenTwoLampDeviceRecords
 * description:
 * param {string} deviceType
 * param {string} mac
 * param {*RealDataSql} obj
 * return {*}
********************************************************************************/
func CheckDiffBetweenTwoLampDeviceRecords(deviceType string, mac string, obj *RealDataSql) bool {
	result := true
	oldObj := &RealDataSql{}
	err := redis.GetValueFromHash(deviceType, mac, true, oldObj)
	if err == nil {
		result = oldObj.FlowState != obj.FlowState ||
			oldObj.HeartRate != obj.HeartRate ||
			oldObj.Respiratory != obj.Respiratory ||
			oldObj.ActivityFreq != obj.ActivityFreq ||
			oldObj.PostureState != obj.PostureState ||
			oldObj.BodyMovement != obj.BodyMovement ||
			oldObj.BodyStatus != obj.BodyStatus
	}
	t := time.Now().Add(time.Minute * 2)
	redis.SaveValueToHash(deviceType, mac, &t, obj)
	return result
}

/******************************************************************************
 * function: GetUserToken
 * description: 创建用户的token，创建后保存到redis，并设置过期时间为3个月
 * param {*User} user
 * return {*}
********************************************************************************/
type UserToken struct {
	UserID int64  `json:"user_id"`
	Phone  string `json:"phone"`
}

func GetUserToken(user *User) (string, error) {
	tokenKey := fmt.Sprintf("%s_%d", common.UserTbl, user.ID)
	token, err := redis.GetValue(tokenKey)
	if err == nil && token != "" {
		return token, nil
	}
	userToken := &UserToken{UserID: user.ID, Phone: user.Phone}
	js, err := json.Marshal(userToken)
	if err != nil {
		mylog.Log.Errorln(err)
		return "", err
	}
	token, err = common.EncryptDataWithDefaultkey(string(js))
	if err != nil {
		mylog.Log.Errorln(err)
		return "", err
	}
	redis.SetValueEx(tokenKey, token, 3*30*24*3600)
	return token, nil
}

/******************************************************************************
 * function: VerifyUserToken
 * description: 检查用户的token是否有效
 * param {string} token
 * return {*}
********************************************************************************/
func VerifyUserToken(token string) bool {
	js, err := common.DecryptDataNoCBCWithDefaultkey(token)
	if err != nil {
		mylog.Log.Errorln(err)
		return false
	}
	userToken := &UserToken{}
	err = json.Unmarshal([]byte(js), userToken)
	if err != nil {
		mylog.Log.Errorln(err)
		return false
	}
	tokenKey := fmt.Sprintf("%s_%d", common.UserTbl, userToken.UserID)
	tokenInRedis, err := redis.GetValue(tokenKey)
	if err != nil || tokenInRedis == "" {
		mylog.Log.Errorln("token not found in redis")
		return false
	}
	if tokenInRedis != token {
		mylog.Log.Errorln("token not match")
		return false
	}
	return true
}
