/******************************************************************************
 * Author: liguoqiang
 * Date: 2023-11-16 20:12:48
 * LastEditors: liguoqiang
 * LastEditTime: 2023-12-15 15:20:08
 * Description: define fall check data struct
********************************************************************************/
package mysql

import (
	"database/sql"
	"hjyserver/cfg"
	"hjyserver/exception"
	mylog "hjyserver/log"
	"hjyserver/mdb/common"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// swagger:model FallParams
type FallParams struct {
	ID            int64  `json:"id" mysql:"id" binding:"omitempty"`
	DeviceId      int64  `json:"device_id" mysql:"device_id" binding:"required"`
	InstallHeight int    `json:"install_height" mysql:"install_height"`
	InstallFlag   int    `json:"install_flag" mysql:"install_flag"`
	Beeper        int    `json:"beeper" mysql:"beeper"`
	LeftDist      int    `json:"left_dist" mysql:"left_dist"`
	RightDist     int    `json:"right_dist" mysql:"right_dist"`
	BackDist      int    `json:"back_dist" mysql:"back_dist"`
	FrontDist     int    `json:"front_dist" mysql:"front_dist"`
	Sensitive     int    `json:"sensitive" mysql:"sensi"`
	StateDelay    int    `json:"state_delay" mysql:"state_delay"`
	DateTime      string `json:"create_time" mysql:"create_time" binding:"datetime=2006-01-02 15:04:05"`
}

func NewFallParams() *FallParams {
	return &FallParams{
		ID:            0,
		DeviceId:      0,
		InstallHeight: 0,
		InstallFlag:   0,
		Beeper:        0,
		LeftDist:      0,
		RightDist:     0,
		BackDist:      0,
		FrontDist:     0,
		Sensitive:     0,
		StateDelay:    0,
		DateTime:      time.Now().Format(cfg.TmFmtStr),
	}
}

/*
QueryFallParamsByCond...
*/
func QueryFallParamsByCond(filter interface{}, page *common.PageDao, sort interface{}, limited int, results *[]FallParams) bool {
	res := false
	backFunc := func(rows *sql.Rows) {
		obj := NewFallParams()
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*results = append(*results, *obj)
		}
	}
	if page == nil {
		res = QueryDao(common.FallParamsTbl, filter, sort, limited, backFunc)
	} else {
		res = QueryPage(common.FallParamsTbl, page, filter, sort, backFunc)
	}
	return res
}

/*
Decode 解析从gin获取的数据 转换成FallParams
*/
func (me *FallParams) DecodeFromGin(c *gin.Context) {
	if err := c.ShouldBindBodyWith(me, binding.JSON); err != nil {
		exception.Throw(http.StatusAccepted, err.Error())
	}
	if me.DeviceId == 0 {
		exception.Throw(http.StatusAccepted, "device id required")
	}
}

func (me *FallParams) DecodeFromRows(rows *sql.Rows) error {
	return rows.Scan(&me.ID,
		&me.DeviceId,
		&me.InstallHeight,
		&me.InstallFlag,
		&me.Beeper,
		&me.LeftDist,
		&me.RightDist,
		&me.BackDist,
		&me.FrontDist,
		&me.Sensitive,
		&me.StateDelay,
		&me.DateTime)
}
func (me *FallParams) DecodeFromRow(row *sql.Row) error {
	return row.Scan(&me.ID,
		&me.DeviceId,
		&me.InstallHeight,
		&me.InstallFlag,
		&me.Beeper,
		&me.LeftDist,
		&me.RightDist,
		&me.BackDist,
		&me.FrontDist,
		&me.Sensitive,
		&me.StateDelay,
		&me.DateTime)
}

/*
QueryByID() 查询股票实时行情
*/
func (me *FallParams) QueryByID(id int64) bool {
	me.SetID(id)
	return QueryDaoByID(common.FallParamsTbl, me.ID, me)
}

/*
Insert FallCheck数据插入
*/
func (me *FallParams) Insert() bool {
	tblName := common.FallParamsTbl
	if !CheckTableExist(tblName) {
		sql := `create table ` + tblName + ` (
            id MEDIUMINT NOT NULL AUTO_INCREMENT,
			device_id MEDIUMINT not null comment '设备id,与设备表关联',
			install_height int not null comment '安装高度',
			install_flag int not null comment '安装标志',
			beeper int not null comment '蜂鸣器',
			left_dist int not null comment '左距离',
			right_dist int not null comment '右距离',
			back_dist int not null comment '后距离',
			front_dist int not null comment '前距离',
			sensi int not null comment '灵敏度',
			state_delay int not null comment '状态延时',
            create_time datetime comment '新增日期',
            PRIMARY KEY (id, device_id, create_time)
        )`
		CreateTable(sql)
	}
	var ret = InsertDao(tblName, me)
	return ret
}

/*
 */
func (me *FallParams) Update() bool {
	return UpdateDaoByID(common.FallParamsTbl, me.ID, me)
}

/*
 */
func (me *FallParams) Delete() bool {
	return DeleteDaoByID(common.FallParamsTbl, me.ID)
}

/*
设置ID
*/
func (me *FallParams) SetID(id int64) {
	me.ID = id
}

/*******************************************************************************
* 定义FallCheck结构
******************************************************************************/

// swagger:model FallCheck
type FallCheck struct {
	ID          int64  `json:"id" mysql:"id" binding:"omitempty"`
	Mac         string `json:"mac" mysql:"mac" binding:"required"`
	Type        int    `json:"type" mysql:"type"`
	PersonState int    `json:"person_state" mysql:"person_state"`
	ActiveState int    `json:"active_state" mysql:"active_state"`
	FallState   int    `json:"fall_state" mysql:"fall_state"`
	DateTime    string `json:"create_time" mysql:"create_time" binding:"datetime=2006-01-02 15:04:05"`
}

/*
NewFallCheck...
构造实例
*/
func NewFallCheck() *FallCheck {
	return &FallCheck{
		ID:          0,
		Mac:         "",
		Type:        0,
		PersonState: 0,
		ActiveState: 0,
		FallState:   0,
		DateTime:    time.Now().Format(cfg.TmFmtStr),
	}
}

/*
QueryFallCheckByCond...
根据条件查询FallCheck数据
*/
func QueryFallCheckByCond(filter interface{}, page *common.PageDao, sort interface{}, limited int, results *[]FallCheck) bool {
	res := false
	backFunc := func(rows *sql.Rows) {
		obj := NewFallCheck()
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*results = append(*results, *obj)
		}
	}
	if page == nil {
		res = QueryDao(common.DeviceRecordTbl(FallCheckType), filter, sort, limited, backFunc)
	} else {
		res = QueryPage(common.DeviceRecordTbl(FallCheckType), page, filter, sort, backFunc)
	}
	return res
}

/*
Decode 解析从gin获取的数据 转换成HeartRate
*/
func (me *FallCheck) DecodeFromGin(c *gin.Context) {
	if err := c.ShouldBindBodyWith(me, binding.JSON); err != nil {
		exception.Throw(http.StatusAccepted, err.Error())
	}
}

func (me *FallCheck) DecodeFromRows(rows *sql.Rows) error {
	return rows.Scan(&me.ID,
		&me.Mac,
		&me.Type,
		&me.PersonState,
		&me.ActiveState,
		&me.FallState,
		&me.DateTime)
}
func (me *FallCheck) DecodeFromRow(row *sql.Row) error {
	return row.Scan(&me.ID,
		&me.Mac,
		&me.Type,
		&me.PersonState,
		&me.ActiveState,
		&me.FallState,
		&me.DateTime)
}

/*
QueryByID() 查询股票实时行情
*/
func (me *FallCheck) QueryByID(id int64) bool {
	me.SetID(id)
	return QueryDaoByID(common.DeviceRecordTbl(FallCheckType), me.ID, me)
}

/*
Insert FallCheck数据插入
*/
func (me *FallCheck) Insert() bool {
	tblName := common.DeviceRecordTbl(FallCheckType)
	if !CheckTableExist(tblName) {
		sql := `create table ` + tblName + ` (
            id MEDIUMINT NOT NULL AUTO_INCREMENT,
            mac varchar(32) not null comment '设备mac,与设备表关联',
			type int comment '类型',
            person_state int not null comment '有无人',
			active_state int not null comment '活动状态',
			fall_state int not null comment '跌倒状态',
            create_time datetime comment '新增日期',
            PRIMARY KEY (id, mac, create_time)
        )`
		CreateTable(sql)
	}
	var ret = InsertDao(tblName, me)
	return ret
}

/*
Update() 更新指数表
*/
func (me *FallCheck) Update() bool {
	return UpdateDaoByID(common.DeviceRecordTbl(FallCheckType), me.ID, me)
}

/*
Delete() 删除指数
*/
func (me *FallCheck) Delete() bool {
	return DeleteDaoByID(common.DeviceRecordTbl(FallCheckType), me.ID)
}

/*
设置ID
*/
func (me *FallCheck) SetID(id int64) {
	me.ID = id
}

/*******************************************************************************
* 定义跌倒告警结构
******************************************************************************/

// swagger:model FallAlarm
type FallAlarm struct {
	ID         int64  `json:"id" mysql:"id" binding:"omitempty"`
	Mac        string `json:"mac" mysql:"mac" binding:"required"`
	AlarmEvent int    `json:"alarm_event" mysql:"alarm_event"`
	DateTime   string `json:"create_time" mysql:"create_time" binding:"datetime=2006-01-02 15:04:05"`
}

/*
NewFallAlarm...
构造实例
*/
func NewFallAlarm() *FallAlarm {
	return &FallAlarm{
		ID:         0,
		Mac:        "",
		AlarmEvent: 0,
		DateTime:   time.Now().Format(cfg.TmFmtStr),
	}
}

/*
QueryFallAlarmByCond...
根据条件查询FallCheck数据
*/
func QueryFallAlarmByCond(filter interface{}, page *common.PageDao, sort interface{}, results *[]FallAlarm) bool {
	res := false
	backFunc := func(rows *sql.Rows) {
		obj := NewFallAlarm()
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*results = append(*results, *obj)
		}
	}
	if page == nil {
		res = QueryDao(common.FallAlarmTbl, filter, sort, -1, backFunc)
	} else {
		res = QueryPage(common.FallAlarmTbl, page, filter, sort, backFunc)
	}
	return res
}

/*
Decode 解析从gin获取的数据 转换成FallAlarm
*/
func (me *FallAlarm) DecodeFromGin(c *gin.Context) {
	if err := c.ShouldBindBodyWith(me, binding.JSON); err != nil {
		exception.Throw(http.StatusAccepted, err.Error())
	}
}

func (me *FallAlarm) DecodeFromRows(rows *sql.Rows) error {
	return rows.Scan(&me.ID,
		&me.Mac,
		&me.AlarmEvent,
		&me.DateTime)
}
func (me *FallAlarm) DecodeFromRow(row *sql.Row) error {
	return row.Scan(&me.ID,
		&me.Mac,
		&me.AlarmEvent,
		&me.DateTime)
}
