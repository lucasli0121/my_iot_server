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
	go func() {
		for {
			time.Sleep(5 * time.Minute)
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
			select {
			case <-quit:
				mylog.Log.Errorln("quit check device online")
				return
			case <-checkOnlineTimeout:
				checkDeviceOnline()
				askAllRealData()
			case <-checkDataTimeout:
				checkNoRealDataLamp()
			}
		}
	}()

	go func() {
		for {
			NotifyTask()
			time.Sleep(30 * time.Second)
		}
	}()
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
	filter := fmt.Sprintf("online=1 and online_time<='%s'", time.Now().Add(-6*time.Minute).Format(cfg.TmFmtStr))
	var gList = []Device{}
	QueryDeviceByCond(filter, nil, "", &gList)
	for _, v := range gList {
		v.Online = 0
		v.Update()
		status := HeartBeatMsg{Mac: v.Mac, Online: 0}
		mq.PublishData(common.MakeDeviceHeartBeatTopic(v.Mac), status)
	}
}

/**
 * @description:设置设备在线状态
 * @param {string} mac
 * @param {int} online
 * @return {*}
 */
func SetDeviceOnline(mac string, online int) {
	var gList = []Device{}
	QueryDeviceByCond(fmt.Sprintf("mac='%s'", mac), nil, nil, &gList)
	if len(gList) > 0 {
		gList[0].Online = online
		gList[0].OnlineTime = time.Now().Format(cfg.TmFmtStr)
		gList[0].Update()
		status := HeartBeatMsg{Mac: mac, Online: online}
		mq.PublishData(common.MakeDeviceHeartBeatTopic(mac), status)
	}
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
		case Ed713Type:
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
		case LampType:
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

/******************************************************************************
 * function: subscribeDeviceTopic
 * description: query device which type is lamp_type record and subscribe mac as topic
 * return {*}
********************************************************************************/
func subscribeDeviceTopic() {
	filter := fmt.Sprintf("type='%s'", LampType)
	var results []Device
	QueryDeviceByCond(filter, nil, "create_time desc", &results)
	if len(results) > 0 {
		for _, v := range results {
			mq.SubscribeTopic(MakeHl77DeliverTopicByMac(v.Mac), NewLampMqttMsgProc())
		}
	}
	results = nil
	filter = fmt.Sprintf("type='%s'", Ed713Type)
	QueryDeviceByCond(filter, nil, "create_time desc", &results)
	if len(results) > 0 {
		for _, v := range results {
			SubscribeEd713MqttTopic(v.Mac)
		}
	}
	results = nil
	filter = fmt.Sprintf("type='%s'", X1Type)
	QueryDeviceByCond(filter, nil, "create_time desc", &results)
	if len(results) > 0 {
		for _, v := range results {
			SubscribeX1MqttTopic(v.Mac)
		}
	}
}

/********************************************************************
* 分页查询功能
* 通过limit, skip 实现简单分页
* pageNo==1时返回总页数
********************************************************************/
func QueryPage(table string, page *common.PageDao, filter interface{}, sort interface{}, cb func(*sql.Rows)) bool {
	totalPages := int64(0)
	sql := "select SQL_CALC_FOUND_ROWS * from " + table
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
	row := mDb.QueryRow("select FOUND_ROWS()")
	totalCount := int64(0)
	if row != nil {
		row.Scan(&totalCount)
	}
	totalPages = int64(float32(totalCount)/float32(page.PageSize) + float32(0.5))
	page.TotalPages = totalPages
	return true
}

/*
 * func Query, support method for any query
 *
 */
func QueryDao(table string, filter interface{}, sort interface{}, limited int, cb func(*sql.Rows)) bool {
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
	sql := "show tables"
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
		if len(fields) > 0 {
			fields += `,`
		}
		tag := f.Tag.Get("mysql")
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
			case reflect.Float64:
				fields += " double"
			case reflect.String:
				binding := f.Tag.Get("binding")
				if len(binding) > 0 && strings.Split(binding, "=")[0] == "datetime" {
					if len(strings.Split(binding, "=")) > 1 && len(strings.Split(binding, "=")[1]) <= 11 {
						fields += " date"
					} else {
						fields += " datetime"
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
					if len(binding) > 0 && strings.Split(binding, "=")[0] == "datetime" {
						if len(strings.Split(binding, "=")) > 1 && len(strings.Split(binding, "=")[1]) <= 11 {
							fields += " date"
						} else {
							fields += " datetime"
						}
					} else {
						if f.Tag.Get("size") != "" {
							fields += " varchar(" + f.Tag.Get("size") + ")"
						} else {
							fields += " varchar(255)"
						}
					}
				}
			}
			if f.Tag.Get("isnull") == "false" || f.Tag.Get("binding") == "required" {
				fields += " not null"
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
				}
			}
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
		if f.Tag.Get("mysql") != "id" {
			setval = fmt.Sprintf(" %s=", f.Tag.Get("mysql"))
		}
		switch v.Kind() {
		case reflect.Int64:
			if f.Name != "ID" {
				setval += fmt.Sprintf("%d", v.Int())
			}
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
