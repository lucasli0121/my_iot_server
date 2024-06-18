/*
 * @Author: liguoqiang
 * @Date: 2022-06-15 14:27:42
 * @LastEditors: liguoqiang
 * @LastEditTime: 2023-05-16 19:28:41
 * @Description:
 */
/**********************************************************
* 此文件定义股票相关结构
* 包含： 股票信息，股票综述，股票行情，股票历史等
**********************************************************/
package mysql

import (
	"database/sql"
	"hjyserver/cfg"
	"hjyserver/exception"
	mylog "hjyserver/log"
	"hjyserver/mdb/common"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

const (
	HeatRateType  = "heart_rate"
	FallCheckType = "fall_check"
	LampType      = "lamp_type"
	Ed713Type     = "ed713_type"
	Ed719Type     = "ed719_type"
	X1Type        = "x1_type"
)

func GetDeviceTypeByName(name string) string {
	var typeStr string
	if len(name) < 2 {
		return typeStr
	}
	if name == "ESP_GATTS_DEVI" || strings.ToUpper(name)[0:2] == "P2" {
		typeStr = HeatRateType
	} else if strings.ToUpper(name)[0:2] == "P3" {
		typeStr = FallCheckType
	} else if strings.ToUpper(name)[0:2] == "X1" {
		typeStr = X1Type
	} else if len(name) >= 4 && strings.ToUpper(name)[0:4] == "HL77" {
		typeStr = LampType
	} else if len(name) >= 5 && strings.ToUpper(name)[0:5] == "ED713" {
		typeStr = Ed713Type
	} else if len(name) >= 5 && strings.ToUpper(name)[0:5] == "ED719" {
		typeStr = Ed719Type
	}
	return typeStr
}

/******************************************************
* 为mysql 数据库提供的结构
* device基本信息结构体
*******************************************************/

// swagger:model Device
type Device struct {
	// 设备ID
	// required: true
	ID   int64  `json:"id" mysql:"id" binding:"omitempty"`
	Name string `json:"name" mysql:"name" `
	// 设备类型
	// required: false
	// enum: heart_rate,fall_check,lamp_type
	Type string `json:"type" mysql:"type" `
	// 设备mac地址
	// required: true
	Mac     string `json:"mac" mysql:"mac"`
	RoomNum string `json:"room_num" mysql:"room_num"`
	// 是否在线
	// required: false
	// enum: 0,1
	Online     int    `json:"online" mysql:"online"`
	OnlineTime string `json:"online_time" mysql:"online_time"`
	CreateTime string `json:"create_time" mysql:"create_time"`
	Remark     string `json:"remark" mysql:"remark"`
}

func NewDevice() *Device {
	return &Device{
		ID:         0,
		Name:       "",
		Type:       "",
		Mac:        "",
		RoomNum:    "",
		Online:     0,
		OnlineTime: time.Now().Format(cfg.TmFmtStr),
		CreateTime: time.Now().Format(cfg.TmFmtStr),
		Remark:     "",
	}
}

/*
*  QueryAllDevice...
*  查询所有Device基本信息
 */
func QueryAllDevice(results *[]Device) bool {
	res := QueryDao(common.DeviceTbl, nil, nil, -1, func(rows *sql.Rows) {
		var v *Device = NewDevice()
		err := v.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*results = append(*results, *v)
		}
	})
	return res
}

/*
QueryDeviceByCond...
根据条件查询股票基本信息
*/
func QueryDeviceByCond(filter interface{}, page *common.PageDao, sort interface{}, results *[]Device) bool {
	res := false
	backFunc := func(rows *sql.Rows) {
		obj := NewDevice()
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*results = append(*results, *obj)
		}
	}
	if page == nil {
		res = QueryDao(common.DeviceTbl, filter, sort, -1, backFunc)
	} else {
		res = QueryPage(common.DeviceTbl, page, filter, sort, backFunc)
	}
	return res
}

func (me *Device) DecodeFromRows(rows *sql.Rows) error {
	err := rows.Scan(&me.ID, &me.Name, &me.Type, &me.Mac, &me.RoomNum, &me.Online, &me.OnlineTime, &me.CreateTime, &me.Remark)
	return err
}
func (me *Device) DecodeFromRow(row *sql.Row) error {
	err := row.Scan(&me.ID, &me.Name, &me.Type, &me.Mac, &me.RoomNum, &me.Online, &me.OnlineTime, &me.CreateTime, &me.Remark)
	return err
}

/*
Decode 解析从gin获取的数据 转换成Device
*/
func (me *Device) DecodeFromGin(c *gin.Context) {
	if err := c.ShouldBindBodyWith(me, binding.JSON); err != nil {
		exception.Throw(http.StatusAccepted, err.Error())
	}
	if me.Name == "" || me.Mac == "" {
		exception.Throw(http.StatusAccepted, "name or mac is empty!")
	}
}

/*
QueryByID() 查询股票基本信息
*/
func (me *Device) QueryByID(id int64) bool {
	return QueryDaoByID(common.DeviceTbl, id, me)
}

/*
Insert 股票基本信息数据插入
*/
func (me *Device) Insert() bool {
	tblName := common.DeviceTbl
	if !CheckTableExist(tblName) {
		sql := `create table ` + tblName + ` (
            id MEDIUMINT NOT NULL AUTO_INCREMENT,
            name varchar(32) NOT NULL COMMENT '名称',
            type varchar(32) NOT NULL COMMENT '类型',
			mac char(32) NOT NULL COMMENT 'mac地址',
			room_num char(32) NOT NULL COMMENT '房间号',
			online int NOT NULL COMMENT '是否在线',
			online_time datetime COMMENT '在线时间',
            create_time datetime comment '新增日期',
			remark varchar(64) comment '备注',
            PRIMARY KEY (id, mac, create_time)
        )`
		CreateTable(sql)
	}
	return InsertDao(common.DeviceTbl, me)
}

/*
Update() 更新股票基本信息
*/
func (me *Device) Update() bool {
	return UpdateDaoByID(common.DeviceTbl, me.ID, me)
}

/*
Delete() 删除指数
*/
func (me *Device) Delete() bool {
	return DeleteDaoByID(common.DeviceTbl, me.ID)
}

/*
设置ID
*/
func (me *Device) SetID(id int64) {
	me.ID = id
}
