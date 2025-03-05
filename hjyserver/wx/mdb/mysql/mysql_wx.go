/******************************************************************************
 * Author: liguoqiang
 * Date: 2024-07-12 18:15:45
 * LastEditors: liguoqiang
 * LastEditTime: 2024-09-09 14:05:27
 * Description:
********************************************************************************/
package mysqlwx

import (
	"database/sql"
	"fmt"
	"hjyserver/exception"
	mylog "hjyserver/log"
	"hjyserver/mdb/common"
	"hjyserver/mdb/mysql"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

/******************************************************************************
 * function:
 * description: 小程序登录的结构定义
 * return {*}
********************************************************************************/
type WxMiniProgram struct {
	ID         int64  `json:"id" mysql:"id" binding:"omitempty"`
	UserId     int64  `json:"user_id" mysql:"user_id" unique:"true" comment:"用户id"`
	OpenId     string `json:"open_id" mysql:"open_id" size:"64" unique:"true" comment:"小程序open_id" `
	SessionKey string `json:"session_key" mysql:"session_key" size:"64" comment:"小程序session" `
	NickName   string `json:"nick_name" mysql:"nick_name" size:"32" comment:"昵称" `
	Gender     int    `json:"gender" mysql:"gender" default:"0" comment:"性别" `
	AvatarUrl  string `json:"avatar_url" mysql:"avatar_url" comment:"微信头像" `
	UnionId    string `json:"union_id" mysql:"union_id" size:"64" comment:"微信公衆號union_id" `
	Version    string `json:"version" mysql:"version" size:"16" comment:"版本号" `
	CreateTime string `json:"create_time" mysql:"create_time" binding:"datetime=2006-01-02 15:04:05" comment:"创建时间" `
}

func (WxMiniProgram) TableName() string {
	return "wx_mini_program_tbl"
}
func NewWxMiniProgram() *WxMiniProgram {
	return &WxMiniProgram{}
}
func (me *WxMiniProgram) Insert() bool {
	if !mysql.CheckTableExist(me.TableName()) {
		mysql.CreateTableWithStruct(me.TableName(), me)
	}
	return mysql.InsertDao(me.TableName(), me)
}
func (me *WxMiniProgram) Update() bool {
	return mysql.UpdateDaoByID(me.TableName(), me.ID, me)
}
func (me *WxMiniProgram) Delete() bool {
	return mysql.DeleteDaoByID(me.TableName(), me.ID)
}
func (me *WxMiniProgram) SetID(id int64) {
	me.ID = id
}
func (me *WxMiniProgram) QueryByID(id int64) bool {
	return mysql.QueryDaoByID(me.TableName(), id, me)
}
func (me *WxMiniProgram) DecodeFromRows(rows *sql.Rows) error {
	err := rows.Scan(&me.ID, &me.UserId, &me.OpenId, &me.SessionKey, &me.NickName, &me.Gender, &me.AvatarUrl, &me.UnionId, &me.Version, &me.CreateTime)
	return err
}
func (me *WxMiniProgram) DecodeFromRow(row *sql.Row) error {
	err := row.Scan(&me.ID, &me.UserId, &me.OpenId, &me.SessionKey, &me.NickName, &me.Gender, &me.AvatarUrl, &me.UnionId, &me.Version, &me.CreateTime)
	return err
}
func (me *WxMiniProgram) DecodeFromGin(c *gin.Context) {
	if err := c.ShouldBindBodyWith(me, binding.JSON); err != nil {
		exception.Throw(common.JsonError, err.Error())
	}
}

func DeleteMiniProgramByUserId(userId int64) bool {
	filter := fmt.Sprintf("user_id = %d", userId)
	return mysql.DeleteDaoByFilter(NewWxMiniProgram().TableName(), filter)
}

/******************************************************************************
 * function: QueryWxMiniProgramByOpenId
 * description: 根据openId查询小程序用户信息
 * param {string} openId
 * param {*[]WxMiniProgram} results
 * return {*}
********************************************************************************/
func QueryWxMiniProgramByOpenId(openId string, results *[]WxMiniProgram) bool {
	filter := fmt.Sprintf("open_id = '%s'", openId)
	mysql.QueryDao(NewWxMiniProgram().TableName(), filter, nil, -1, func(rows *sql.Rows) {
		obj := NewWxMiniProgram()
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*results = append(*results, *obj)
		}
	})
	return len(*results) > 0
}

/******************************************************************************
 * function: QueryWxMiniProgramByUnionId
 * description: 根据unionId查询小程序用户信息
 * param {string} unionId
 * param {*[]WxMiniProgram} results
 * return {*}
********************************************************************************/
func QueryWxMiniProgramByUnionId(unionId string, results *[]WxMiniProgram) bool {
	filter := fmt.Sprintf("union_id = '%s'", unionId)
	mysql.QueryDao(NewWxMiniProgram().TableName(), filter, nil, -1, func(rows *sql.Rows) {
		obj := NewWxMiniProgram()
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*results = append(*results, *obj)
		}
	})
	return len(*results) > 0
}
func QueryWxMiniProgramByUserId(userId int64, results *[]WxMiniProgram) bool {
	filter := fmt.Sprintf("user_id = %d", userId)
	mysql.QueryDao(NewWxMiniProgram().TableName(), filter, nil, -1, func(rows *sql.Rows) {
		obj := NewWxMiniProgram()
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*results = append(*results, *obj)
		}
	})
	return len(*results) > 0
}

/******************************************************************************
 * function:
 * description: 定义用户关注公众号的结构
 * return {*}
********************************************************************************/
type WxOfficalAccount struct {
	ID          int64  `json:"id" mysql:"id" binding:"omitempty"`
	ToUserName  string `json:"to_user_name" mysql:"to_user_name" size:"64" comment:"开发者微信号" `
	FromOpenId  string `json:"from_open_id" mysql:"from_open_id" size:"64" comment:"发送方帐号（一个OpenID）" `
	FromUnionId string `json:"from_union_id" mysql:"from_union_id" size:"64" comment:"微信公衆號union_id" `
	MsgType     string `json:"msg_type" mysql:"msg_type" size:"32" comment:"消息类型 事件为event" `
	Event       string `json:"event" mysql:"event" size:"32" comment:"事件类型 关注事件为subscribe, 取消关注为unscribe" `
	CreateTime  string `json:"create_time" mysql:"create_time" binding:"datetime=2006-01-02 15:04:05" comment:"创建时间" `
}

func (WxOfficalAccount) TableName() string {
	return "wx_offical_account_tbl"
}
func NewWxOfficalAccount() *WxOfficalAccount {
	return &WxOfficalAccount{}
}
func (me *WxOfficalAccount) Insert() bool {
	if !mysql.CheckTableExist(me.TableName()) {
		mysql.CreateTableWithStruct(me.TableName(), me)
	}
	return mysql.InsertDao(me.TableName(), me)
}
func (me *WxOfficalAccount) Update() bool {
	return mysql.UpdateDaoByID(me.TableName(), me.ID, me)
}
func (me *WxOfficalAccount) Delete() bool {
	return mysql.DeleteDaoByID(me.TableName(), me.ID)
}
func (me *WxOfficalAccount) SetID(id int64) {
	me.ID = id
}
func (me *WxOfficalAccount) QueryByID(id int64) bool {
	return mysql.QueryDaoByID(me.TableName(), id, me)
}
func (me *WxOfficalAccount) DecodeFromRows(rows *sql.Rows) error {
	err := rows.Scan(&me.ID, &me.ToUserName, &me.FromOpenId, &me.FromUnionId, &me.MsgType, &me.Event, &me.CreateTime)
	return err
}
func (me *WxOfficalAccount) DecodeFromRow(row *sql.Row) error {
	err := row.Scan(&me.ID, &me.ToUserName, &me.FromOpenId, &me.FromUnionId, &me.MsgType, &me.Event, &me.CreateTime)
	return err
}
func (me *WxOfficalAccount) DecodeFromGin(c *gin.Context) {
	if err := c.ShouldBindBodyWith(me, binding.JSON); err != nil {
		exception.Throw(common.JsonError, err.Error())
	}
}
func DeleteWxOfficalAccountByEvent(openId string, msgType string, event string) bool {
	filter := fmt.Sprintf("from_open_id = '%s' and msg_type = '%s' and event = '%s'", openId, msgType, event)
	return mysql.DeleteDaoByFilter(NewWxOfficalAccount().TableName(), filter)
}

func QueryWxOfficalAccountByOpenIdAndMsgType(openId string, msgType string) []WxOfficalAccount {
	filter := fmt.Sprintf("from_open_id = '%s' and msg_type = '%s'", openId, msgType)
	var gList []WxOfficalAccount
	mysql.QueryDao(NewWxOfficalAccount().TableName(), filter, nil, -1, func(rows *sql.Rows) {
		obj := NewWxOfficalAccount()
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			gList = append(gList, *obj)
		}
	})
	return gList
}

func QueryWxOfficalAccountSubscribeByUnionId(unionId string, results *[]WxOfficalAccount) bool {
	filter := fmt.Sprintf("from_union_id = '%s' and event = 'subscribe'", unionId)
	mysql.QueryDao(NewWxOfficalAccount().TableName(), filter, nil, -1, func(rows *sql.Rows) {
		obj := NewWxOfficalAccount()
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*results = append(*results, *obj)
		}
	})
	return len(*results) > 0
}
