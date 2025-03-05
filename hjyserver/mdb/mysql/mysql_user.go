/*
 * @Author: liguoqiang
 * @Date: 2022-06-15 14:27:42
 * @LastEditors: liguoqiang
 * @LastEditTime: 2023-09-29 17:26:46
 * @Description:
 */
/**********************************************************
* 此文件定义股票相关结构
* 包含： 股票信息，股票综述，股票行情，股票历史等
**********************************************************/
package mysql

import (
	"database/sql"
	"fmt"
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

/******************************************************
* 为mysql 数据库提供的结构
*******************************************************/

// swagger:model User
type User struct {
	ID            int64  `json:"id" mysql:"id" binding:"omitempty"`
	Account       string `json:"account" mysql:"account"`
	Password      string `json:"password" mysql:"password" `
	NickName      string `json:"nick_name" mysql:"nick_name"`
	Gender        int    `json:"gender" mysql:"gender"`         // 0 未知 1 男 2 女
	LoginType     int    `json:"login_type" mysql:"login_type"` // 0:phone 1:email
	Phone         string `json:"phone" mysql:"phone"`
	Email         string `json:"email" mysql:"email"`
	EmergentPhone string `json:"emergent_phone" mysql:"emergent_phone"`
	Face          string `json:"face" mysql:"face"`
	BornDate      string `json:"born_date" mysql:"born_date" size:"32"`
	Grade         string `json:"grade" mysql:"grade" size:"32"`
	Address       string `json:"address" mysql:"address"`
	RoomNum       string `json:"room_num" mysql:"room_num"`
	IsLogin       int    `json:"is_login" mysql:"is_login"`
	LoginTime     string `json:"login_time" mysql:"login_time"`
	CreateTime    string `json:"create_time" mysql:"create_time"`
}

func NewUser() *User {
	loginTm := common.GetNowTime()
	createTm := common.GetNowTime()
	return &User{
		ID:            0,
		Account:       "",
		Password:      "",
		NickName:      "",
		Gender:        0,
		LoginType:     0,
		Phone:         "",
		Email:         "",
		EmergentPhone: "",
		Face:          "",
		BornDate:      "",
		Grade:         "",
		Address:       "",
		RoomNum:       "",
		IsLogin:       1,
		LoginTime:     loginTm,
		CreateTime:    createTm,
	}
}

/*
*  QueryAllUsers...
*  查询所有User基本信息
 */
func QueryAllUsers(results *[]User) bool {
	res := QueryDao(common.UserTbl, nil, nil, -1, func(rows *sql.Rows) {
		var v *User = NewUser()
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
QueryUserByCond...
根据条件查询user基本信息
*/
func QueryUserByCond(filter interface{}, page *common.PageDao, sort interface{}, results *[]User) bool {
	res := false
	backFunc := func(rows *sql.Rows) {
		obj := NewUser()
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*results = append(*results, *obj)
		}
	}
	if page == nil {
		res = QueryDao(common.UserTbl, filter, sort, -1, backFunc)
	} else {
		res = QueryPage(common.UserTbl, page, filter, sort, backFunc)
	}
	return res
}

func (me *User) DecodeFromRows(rows *sql.Rows) error {
	var emergentPhone sql.NullString
	var bornDate sql.NullString
	var grade sql.NullString
	err := rows.Scan(
		&me.ID,
		&me.Account,
		&me.Password,
		&me.NickName,
		&me.Gender,
		&me.LoginType,
		&me.Phone,
		&me.Email,
		&emergentPhone,
		&me.Face,
		&bornDate,
		&grade,
		&me.Address,
		&me.RoomNum,
		&me.IsLogin,
		&me.LoginTime,
		&me.CreateTime)
	if emergentPhone.Valid {
		me.EmergentPhone = emergentPhone.String
	}
	if bornDate.Valid {
		me.BornDate = bornDate.String
	}
	if grade.Valid {
		me.Grade = grade.String
	}
	return err
}
func (me *User) DecodeFromRow(row *sql.Row) error {
	var emergentPhone sql.NullString
	var bornDate sql.NullString
	var grade sql.NullString
	err := row.Scan(
		&me.ID,
		&me.Account,
		&me.Password,
		&me.NickName,
		&me.Gender,
		&me.LoginType,
		&me.Phone,
		&me.Email,
		&emergentPhone,
		&me.Face,
		&bornDate,
		&grade,
		&me.Address,
		&me.RoomNum,
		&me.IsLogin,
		&me.LoginTime,
		&me.CreateTime)
	if emergentPhone.Valid {
		me.EmergentPhone = emergentPhone.String
	}
	if bornDate.Valid {
		me.BornDate = bornDate.String
	}
	if grade.Valid {
		me.Grade = grade.String
	}
	return err
}

/*
Decode 解析从gin获取的数据 转换成User
*/
func (me *User) DecodeFromGin(c *gin.Context) {
	if err := c.ShouldBindBodyWith(me, binding.JSON); err != nil {
		exception.Throw(http.StatusAccepted, err.Error())
	}
	if me.LoginType == 0 && me.Phone == "" {
		exception.Throw(http.StatusAccepted, "phone is empty!")
	}
	if me.LoginType == 1 && (me.Email == "") {
		exception.Throw(http.StatusAccepted, "email is empty!")
	}
}

/*
QueryByID() 查询user by id
*/
func (me *User) QueryByID(id int64) bool {
	return QueryDaoByID(common.UserTbl, id, me)
}

/*
Insert 股票基本信息数据插入
*/
func (me *User) Insert() bool {
	tblName := common.UserTbl
	if !CheckTableExist(tblName) {
		sql := `create table ` + tblName + ` (
            id MEDIUMINT NOT NULL AUTO_INCREMENT,
            account char(32) NOT NULL COMMENT '账号',
			password char(32) NOT NULL COMMENT '密码',
            nick_name varchar(32) NOT NULL COMMENT '昵称',
			gender int default 0 comment '性别 0:未知 1:男 2:女',
			login_type int NOT NULL COMMENT '登录类型 0:phone 1:email',
            phone varchar(32) comment '手机号',
			email varchar(32) comment '邮箱',
			emergent_phone varchar(32) comment '紧急联系电话',
            face varchar(255) comment '头像',
			born_date varchar(32) comment '出生日期',
			grade varchar(32) comment '年级',
            address varchar(128) comment '地址',
            room_num varchar(32) comment '房间号',
			is_login int default 0 comment '是否登录',
			login_time datetime comment '登录时间',
            create_time datetime comment '创建时间',
            PRIMARY KEY (id, phone, create_time)
        )`
		CreateTable(sql)
	}
	return InsertDao(tblName, me)
}

/*
Update() 更新股票基本信息
*/
func (me *User) Update() bool {
	return UpdateDaoByID(common.UserTbl, me.ID, me)
}

/*
Delete() 删除指数
*/
func (me *User) Delete() bool {
	return DeleteDaoByID(common.UserTbl, me.ID)
}

/*
设置ID
*/
func (me *User) SetID(id int64) {
	me.ID = id
}

/******************************************************
* 为mysql 数据库提供的结构
* define user group
*
*******************************************************/
type UserGroup struct {
	ID        int64  `json:"id" mysql:"id" binding:"omitempty"`
	UserId    int64  `json:"user_id" mysql:"user_id" binding:"required"`
	GroupName string `json:"group_name" mysql:"group_name"`
}

func NewUserGroup() *UserGroup {
	return &UserGroup{
		ID:        0,
		UserId:    0,
		GroupName: "",
	}
}

/**
 * @description: init user group add default group name
 * @return {*}
 */
func InitUserGroup(userId int64) {
	var gList []UserGroup
	if CheckTableExist(common.UserGroupTbl) {
		filter := fmt.Sprintf("user_id = %d", userId)
		QueryGroupByCond(filter, &gList)
	}
	if len(gList) == 0 {
		group := NewUserGroup()
		group.UserId = userId
		if cfg.IsCN() {
			group.GroupName = "客服"
		} else {
			group.GroupName = "客服"
		}
		group.Insert()
		group.UserId = userId
		if cfg.IsCN() {
			group.GroupName = "亲属"
		} else {
			group.GroupName = "親屬"
		}
		group.Insert()
		group.UserId = userId
		if cfg.IsCN() {
			group.GroupName = "好友"
		} else {
			group.GroupName = "好友"
		}
		group.Insert()
	}
}

/*
*  QueryGroupByName...
*  查询Group基本信息
 */
func QueryGroupByCond(filter interface{}, results *[]UserGroup) bool {
	res := QueryDao(common.UserGroupTbl, filter, " group_name", -1, func(rows *sql.Rows) {
		var v *UserGroup = NewUserGroup()
		err := v.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*results = append(*results, *v)
		}
	})
	return res
}

func (me *UserGroup) DecodeFromRows(rows *sql.Rows) error {
	err := rows.Scan(&me.ID, &me.UserId, &me.GroupName)
	return err
}
func (me *UserGroup) DecodeFromRow(row *sql.Row) error {
	err := row.Scan(&me.ID, &me.UserId, &me.GroupName)
	return err
}

/*
Decode 解析从gin获取的数据 转换成UserGroup
*/
func (me *UserGroup) DecodeFromGin(c *gin.Context) {
	if err := c.ShouldBindBodyWith(me, binding.JSON); err != nil {
		exception.Throw(http.StatusAccepted, err.Error())
	}
}

/*
QueryByID() 查询股票基本信息
*/
func (me *UserGroup) QueryByID(id int64) bool {
	return QueryDaoByID(common.UserGroupTbl, id, me)
}

/*
Insert Group基本信息数据插入
*/
func (me *UserGroup) Insert() bool {
	tblName := common.UserGroupTbl
	if !CheckTableExist(tblName) {
		sql := `create table ` + tblName + ` (
            id MEDIUMINT NOT NULL AUTO_INCREMENT,
			user_id MEDIUMINT NOT NULL COMMENT '用户id',
            group_name varchar(32) NOT NULL COMMENT '群組名稱',
            PRIMARY KEY (id)
        )`
		CreateTable(sql)
	}
	return InsertDao(tblName, me)
}

/*
Update() 更新股票基本信息
*/
func (me *UserGroup) Update() bool {
	return UpdateDaoByID(common.UserGroupTbl, me.ID, me)
}

/*
Delete() 删除指数
*/
func (me *UserGroup) Delete() bool {
	return DeleteDaoByID(common.UserGroupTbl, me.ID)
}

/*
设置ID
*/
func (me *UserGroup) SetID(id int64) {
	me.ID = id
}

/******************************************************
* 为mysql 数据库提供的结构
* define user friends related struct
*
*******************************************************/

// swagger:model UserRelation
type UserRelation struct {
	ID         int64  `json:"id" mysql:"id" binding:"omitempty"`
	UserId     int64  `json:"user_id" mysql:"user_id" binding:"required"`
	FriendId   int64  `json:"friend_id" mysql:"friend_id"`
	GroupName  string `json:"group_name" mysql:"group_name"`
	CreateTime string `json:"create_time" mysql:"create_time"`
}

func NewUserRelation() *UserRelation {
	return &UserRelation{
		ID:         0,
		UserId:     0,
		FriendId:   0,
		GroupName:  "",
		CreateTime: time.Now().Format(cfg.TmFmtStr),
	}
}

/*
*  QueryUserRelationByCond...
*  查询UserFriendRelation基本信息
 */
func QueryUserRelationByCond(filter interface{}, results *[]UserRelation) bool {
	res := QueryDao(common.FriendsTbl, filter, " group_name", -1, func(rows *sql.Rows) {
		var v *UserRelation = NewUserRelation()
		err := v.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*results = append(*results, *v)
		}
	})
	return res
}

func (me *UserRelation) DecodeFromRows(rows *sql.Rows) error {
	err := rows.Scan(&me.ID, &me.UserId, &me.FriendId, &me.GroupName, &me.CreateTime)
	return err
}
func (me *UserRelation) DecodeFromRow(row *sql.Row) error {
	err := row.Scan(&me.ID, &me.UserId, &me.FriendId, &me.GroupName, &me.CreateTime)
	return err
}

/*
Decode 解析从gin获取的数据 转换成Friends
*/
func (me *UserRelation) DecodeFromGin(c *gin.Context) {
	if err := c.ShouldBindBodyWith(me, binding.JSON); err != nil {
		exception.Throw(http.StatusAccepted, err.Error())
	}
	if me.UserId == 0 {
		exception.Throw(http.StatusAccepted, "user id is empty!")
	}
}

/*
QueryByID() 查询股票基本信息
*/
func (me *UserRelation) QueryByID(id int64) bool {
	return QueryDaoByID(common.FriendsTbl, id, me)
}

/*
Insert Friends基本信息数据插入
*/
func (me *UserRelation) Insert() bool {
	tblName := common.FriendsTbl
	if !CheckTableExist(tblName) {
		sql := `create table ` + tblName + ` (
            id MEDIUMINT NOT NULL AUTO_INCREMENT,
            user_id MEDIUMINT NOT NULL COMMENT '用户id',
			friend_id MEDIUMINT NOT NULL COMMENT '好友id',
			group_name varchar(32) NOT NULL COMMENT '分组名稱',
            create_time datetime comment '创建时间',
            PRIMARY KEY (id, user_id, create_time)
        )`
		CreateTable(sql)
	}
	return InsertDao(tblName, me)
}

/*
Update() 更新股票基本信息
*/
func (me *UserRelation) Update() bool {
	return UpdateDaoByID(common.FriendsTbl, me.ID, me)
}

/*
Delete() 删除指数
*/
func (me *UserRelation) Delete() bool {
	return DeleteDaoByID(common.FriendsTbl, me.ID)
}

/*
设置ID
*/
func (me *UserRelation) SetID(id int64) {
	me.ID = id
}
func DeleteUserRelationByUserId(userId int64, friendId int64) bool {
	var filter string
	if friendId == 0 {
		filter = fmt.Sprintf("user_id=%d", userId)
	} else {
		filter = fmt.Sprintf("user_id=%d and friend_id=%d", userId, friendId)
	}
	return DeleteDaoByFilter(common.FriendsTbl, filter)
}

/*
********************************************************************************

	UserFriend struct

	define user friend related struct

********************************************************************************
*/

// swagger:model UserFriend
type UserFriend struct {
	UserRelation
	Phone    string `json:"phone" mysql:"phone"`
	Email    string `json:"email" mysql:"email"`
	Face     string `json:"face" mysql:"face"`
	NickName string `json:"nick_name" mysql:"nick_name"`
}

func NewUserFriend() *UserFriend {
	return &UserFriend{
		UserRelation: *NewUserRelation(),
		Phone:        "",
		Email:        "",
		Face:         "",
		NickName:     "",
	}
}

/*
QueryUserFriendByUserId...
根据条件查询股票基本信息
*/
func QueryUserFriendByUserId(userId int64, results *[]UserFriend) bool {
	sql := "select a.*, b.phone, b.email, b.face, b.nick_name from " +
		common.FriendsTbl + " a join " + common.UserTbl + " b on a.friend_id = b.id and a.user_id = " + fmt.Sprintf("%d", userId)
	sql += " order by a.create_time desc"
	rows, err := mDb.Query(sql)
	if err != nil {
		mylog.Log.Errorln(err)
		return false
	}
	defer rows.Close()
	for rows.Next() {
		obj := NewUserFriend()
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*results = append(*results, *obj)
		}
	}
	return true
}

func (me *UserFriend) DecodeFromRows(rows *sql.Rows) error {
	err := rows.Scan(&me.ID, &me.UserId, &me.FriendId, &me.GroupName, &me.CreateTime, &me.Phone, &me.Email, &me.Face, &me.NickName)
	return err
}
func (me *UserFriend) DecodeFromRow(row *sql.Row) error {
	err := row.Scan(&me.ID, &me.UserId, &me.FriendId, &me.GroupName, &me.CreateTime, &me.Phone, &me.Email, &me.Face, &me.NickName)
	return err
}

/*
Decode 解析从gin获取的数据 转换成UserDevice
*/
func (me *UserFriend) DecodeFromGin(c *gin.Context) {
	if err := c.ShouldBindBodyWith(me, binding.JSON); err != nil {
		exception.Throw(http.StatusAccepted, err.Error())
	}
	if me.UserId == 0 {
		exception.Throw(http.StatusAccepted, "user id is empty!")
	}
}

/*
********************************************************************************

	UserDeviceRelation 表

	define user device related struct

********************************************************************************
*/
type UserDeviceRelation struct {
	ID         int64  `json:"id" mysql:"id" binding:"omitempty"`
	UserId     int64  `json:"user_id" mysql:"user_id" key:"true" binding:"required"`
	DeviceId   int64  `json:"device_id" mysql:"device_id" binding:"required"`
	Flag       int    `json:"flag" mysql:"flag"` // 0:自己创建 1:共享
	CreateTime string `json:"create_time" mysql:"create_time"`
}

func NewUserDeviceRelation() *UserDeviceRelation {
	return &UserDeviceRelation{
		ID:         0,
		UserId:     0,
		DeviceId:   0,
		Flag:       0,
		CreateTime: time.Now().Format(cfg.TmFmtStr),
	}
}

/*
QueryUserDeviceRelationByCond...
根据条件查询股票基本信息
*/
func QueryUserDeviceRelationByCond(filter interface{}, page *common.PageDao, sort interface{}, results *[]UserDeviceRelation) bool {
	res := false
	backFunc := func(rows *sql.Rows) {
		obj := NewUserDeviceRelation()
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*results = append(*results, *obj)
		}
	}
	if page == nil {
		res = QueryDao(common.UserDeviceRelationTbl, filter, sort, -1, backFunc)
	} else {
		res = QueryPage(common.UserDeviceRelationTbl, page, filter, sort, backFunc)
	}
	return res
}

func (me *UserDeviceRelation) DecodeFromRows(rows *sql.Rows) error {
	err := rows.Scan(&me.ID, &me.UserId, &me.DeviceId, &me.Flag, &me.CreateTime)
	return err
}
func (me *UserDeviceRelation) DecodeFromRow(row *sql.Row) error {
	err := row.Scan(&me.ID, &me.UserId, &me.DeviceId, &me.Flag, &me.CreateTime)
	return err
}

/*
Decode 解析从gin获取的数据 转换成UserDevice
*/
func (me *UserDeviceRelation) DecodeFromGin(c *gin.Context) {
	if err := c.ShouldBindBodyWith(me, binding.JSON); err != nil {
		exception.Throw(http.StatusAccepted, err.Error())
	}
	if me.UserId == 0 {
		exception.Throw(http.StatusAccepted, "user id is empty!")
	}
}

/*
QueryByID() 查询股票基本信息
*/
func (me *UserDeviceRelation) QueryByID(id int64) bool {
	return QueryDaoByID(common.UserDeviceRelationTbl, id, me)
}

/*
Insert UserDevice
*/
func (me *UserDeviceRelation) Insert() bool {
	tblName := common.UserDeviceRelationTbl
	if !CheckTableExist(tblName) {
		CreateTableWithStruct(tblName, me)
	}
	return InsertDao(tblName, me)
}

/*
Update() 更新
*/
func (me *UserDeviceRelation) Update() bool {
	return UpdateDaoByID(common.UserDeviceRelationTbl, me.ID, me)
}

/*
Delete() 删除
*/
func (me *UserDeviceRelation) Delete() bool {
	return DeleteDaoByID(common.UserDeviceRelationTbl, me.ID)
}

/*
设置ID
*/
func (me *UserDeviceRelation) SetID(id int64) {
	me.ID = id
}

/******************************************************************************
 * function: DeleteWithUser
 * description: delete user device relation table with user id and device id
 * return {*}
********************************************************************************/
func (me *UserDeviceRelation) DeleteWithUser() bool {
	filter := fmt.Sprintf("user_id=%d and device_id=%d", me.UserId, me.DeviceId)
	var vList []UserDeviceRelation
	QueryUserDeviceRelationByCond(filter, nil, nil, &vList)
	if len(vList) == 0 {
		return true
	}
	me.ID = vList[0].ID
	me.Flag = vList[0].Flag
	if me.Flag == common.NormalDeviceFlag {
		filter = fmt.Sprintf("device_id=%d", me.DeviceId)
	} else {
		filter = fmt.Sprintf("user_id=%d and device_id=%d", me.UserId, me.DeviceId)
	}
	return DeleteDaoByFilter(common.UserDeviceRelationTbl, filter)
}

func DeleteDeviceRelationByUserId(userId int64) bool {
	filter := fmt.Sprintf("user_id=%d", userId)
	return DeleteDaoByFilter(common.UserDeviceRelationTbl, filter)
}

/*
********************************************************************************

	UserDevice 表

	define user device related struct
	only a struct don't save database

********************************************************************************
*/

// swagger:model UserDevice
type UserDevice struct {
	// required: true
	// user id
	UserId int64 `json:"user_id" mysql:"user_id" `
	// required: false
	// flag 0:自己创建 1:共享
	Flag int `json:"flag" mysql:"flag"`
	Device
}

func NewUserDevice() *UserDevice {
	return &UserDevice{
		UserId: 0,
		Flag:   0,
		Device: *NewDevice(),
	}
}

/*
QueryUserDeviceByUserId...
*/
func QueryUserDeviceByUserId(userId int64, flag int, results *[]UserDevice) bool {
	sql := "select a.user_id as user_id, a.flag, b.id, b.name, b.type, b.mac, b.online ,b.online_time, b.create_time, b.remark from " +
		common.UserDeviceRelationTbl + " a join " + common.DeviceTbl + " b on a.device_id = b.id and a.user_id = " + fmt.Sprintf("%d", userId)
	if flag != -1 {
		sql += " and a.flag = " + fmt.Sprintf("%d", flag)
	}
	sql += " order by b.create_time desc"

	rows, err := mDb.Query(sql)
	if err != nil {
		mylog.Log.Errorln(err)
		return false
	}
	defer rows.Close()
	for rows.Next() {
		obj := NewUserDevice()
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*results = append(*results, *obj)
		}
	}
	return true
}

func (me *UserDevice) DecodeFromRows(rows *sql.Rows) error {
	err := rows.Scan(&me.UserId, &me.Flag, &me.ID, &me.Name, &me.Type, &me.Mac, &me.Online, &me.OnlineTime, &me.CreateTime, &me.Remark)
	return err
}
func (me *UserDevice) DecodeFromRow(row *sql.Row) error {
	err := row.Scan(&me.UserId, &me.Flag, &me.ID, &me.Name, &me.Type, &me.Mac, &me.Online, &me.OnlineTime, &me.CreateTime, &me.Remark)
	return err
}

/*
Decode 解析从gin获取的数据 转换成UserDevice
*/
func (me *UserDevice) DecodeFromGin(c *gin.Context) {
	if err := c.ShouldBindBodyWith(me, binding.JSON); err != nil {
		exception.Throw(http.StatusAccepted, err.Error())
	}
	if me.UserId == 0 {
		exception.Throw(http.StatusAccepted, "user id is empty!")
	}
}

type UserDeviceDetail struct {
	// required: true
	// user id
	UserId        int64  `json:"user_id" mysql:"user_id" `
	NickName      string `json:"nick_name" mysql:"nick_name" `
	Phone         string `json:"phone" mysql:"phone" `
	EmergentPhone string `json:"emergent_phone" mysql:"emergent_phone" `
	DeviceId      int64  `json:"device_id" mysql:"device_id" `
	DeviceName    string `json:"device_name" mysql:"device_name" `
	Mac           string `json:"mac" mysql:"mac" `
	DeviceType    string `json:"device_type" mysql:"device_type" `
	Flag          int    `json:"flag" mysql:"flag" `
	Remark        string `json:"remark" mysql:"remark" `
}

/******************************************************************************
 * function: QueryUserDeviceByMac
 * description: query user device by mac
 * param {string} mac
 * param {*[]UserDevice} results
 * return {*}
********************************************************************************/
func QueryUserDeviceDetailByMac(mac string, results *[]UserDeviceDetail) bool {
	sqlStr := "select a.id as user_id, a.nick_name, a.phone, a.emergent_phone, b.id as device_id, b.name as device_name, b.mac, b.type as device_type, c.flag, b.remark from " +
		common.UserTbl + " a," + common.DeviceTbl + " b, " + common.UserDeviceRelationTbl + " c where a.id=c.user_id and b.id=c.device_id and b.mac='" + mac + "'"
	rows, err := mDb.Query(sqlStr)
	if err != nil {
		mylog.Log.Errorln(err)
		return false
	}
	defer rows.Close()
	var emergentPhone sql.NullString
	for rows.Next() {
		obj := &UserDeviceDetail{}
		err := rows.Scan(&obj.UserId, &obj.NickName, &obj.Phone, &emergentPhone, &obj.DeviceId, &obj.DeviceName, &obj.Mac, &obj.DeviceType, &obj.Flag, &obj.Remark)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			if emergentPhone.Valid {
				obj.EmergentPhone = emergentPhone.String
			}
			*results = append(*results, *obj)
		}
	}
	return true
}

/******************************************************************************
 * function: RegisterWithUserObj
 * description: 注册一个用户，参数为User对象
 * param {*mysql.User} me
 * return {*}
********************************************************************************/
func RegisterWithUserObj(me *User) (int, interface{}) {
	me.Account = strings.Trim(me.Account, " ")
	me.Phone = strings.Trim(me.Phone, " ")
	me.Email = strings.Trim(me.Email, " ")
	if me.Account == "" {
		return common.ParamError, "account required"
	}
	if me.Phone == "" && me.Email == "" {
		return common.ParamError, "password or email required"
	}
	if me.Phone != "" {
		me.Phone = common.FixPlusInPhoneString(me.Phone)
	}
	filter := fmt.Sprintf("account = '%s'", me.Account)
	var gList []User
	QueryUserByCond(filter, nil, nil, &gList)
	if len(gList) > 0 {
		return common.AccountHasReg, "account has registered"
	}
	if me.Phone != "" {
		filter = fmt.Sprintf("phone = '%s'", me.Phone)
		QueryUserByCond(filter, nil, nil, &gList)
		if len(gList) > 0 {
			return common.PhoneHasReg, "phone has registered"
		}
	}
	if me.Email != "" {
		filter = fmt.Sprintf("email = '%s'", me.Email)
		QueryUserByCond(filter, nil, nil, &gList)
		if len(gList) > 0 {
			return common.EmailHasReg, "email has registered"
		}
	}
	me.Password, _ = common.EncryptDataWithDefaultkey(me.Password)
	me.IsLogin = 1
	me.LoginTime = common.GetNowTime()
	me.CreateTime = common.GetNowTime()
	if me.Insert() {
		return http.StatusOK, me
	}
	return common.RegisterFail, "account registe failed!"
}

/******************************************************************************
 * class: UserShareDevice
 * description: 定义用户共享设备表
 * return {*}
********************************************************************************/

// swagger:model UserShareDevice
type UserShareDevice struct {
	ID         int64  `json:"id" mysql:"id" binding:"omitempty"`
	FromUserId int64  `json:"from_user_id" mysql:"from_user_id" key:"true" binding:"required"`
	ToUserId   int64  `json:"to_user_id" mysql:"to_user_id" key:"true" binding:"required"`
	DeviceId   int64  `json:"device_id" mysql:"device_id" binding:"required"`
	Confirm    int    `json:"confirm" mysql:"confirm"`
	Remark     string `json:"remark" mysql:"remark"`
	CreateTime string `json:"create_time" mysql:"create_time"`
}

func NewUserShareDevice() *UserShareDevice {
	return &UserShareDevice{
		ID:         0,
		FromUserId: 0,
		ToUserId:   0,
		DeviceId:   0,
		Confirm:    0,
		Remark:     "",
		CreateTime: time.Now().Format(cfg.TmFmtStr),
	}
}
func (me *UserShareDevice) MyTableName() string {
	return common.UserShareDeviceTbl
}

func (me *UserShareDevice) DecodeFromRows(rows *sql.Rows) error {
	err := rows.Scan(&me.ID, &me.FromUserId, &me.ToUserId, &me.DeviceId, &me.Confirm, &me.Remark, &me.CreateTime)
	return err
}
func (me *UserShareDevice) DecodeFromRow(row *sql.Row) error {
	err := row.Scan(&me.ID, &me.FromUserId, &me.ToUserId, &me.DeviceId, &me.Confirm, &me.Remark, &me.CreateTime)
	return err
}

/*
Decode 解析从gin获取的数据 转换成UserDevice
*/
func (me *UserShareDevice) DecodeFromGin(c *gin.Context) {
	if err := c.ShouldBindBodyWith(me, binding.JSON); err != nil {
		exception.Throw(common.JsonError, err.Error())
	}
	if me.FromUserId == 0 {
		exception.Throw(common.ParamError, "user id is empty!")
	}
}

/*
QueryByID() 查询股票基本信息
*/
func (me *UserShareDevice) QueryByID(id int64) bool {
	return QueryDaoByID(me.MyTableName(), id, me)
}

/*
Insert UserDevice
*/
func (me *UserShareDevice) Insert() bool {
	if !CheckTableExist(me.MyTableName()) {
		CreateTableWithStruct(me.MyTableName(), me)
	}
	return InsertDao(me.MyTableName(), me)
}

/*
Update() 更新
*/
func (me *UserShareDevice) Update() bool {
	return UpdateDaoByID(me.MyTableName(), me.ID, me)
}

/*
Delete() 删除
*/
func (me *UserShareDevice) Delete() bool {
	return DeleteDaoByID(me.MyTableName(), me.ID)
}

/*
设置ID
*/
func (me *UserShareDevice) SetID(id int64) {
	me.ID = id
}

func DeleteUserShareDeviceByUserId(fromUserId int64, toUserId int64, deviceId int64) bool {
	filter := ""
	if fromUserId != 0 {
		filter += " from_user_id=" + fmt.Sprintf("%d", fromUserId)
	}
	if toUserId != 0 {
		filter += " and to_user_id=" + fmt.Sprintf("%d", toUserId)
	}
	if deviceId != 0 {
		filter += " and device_id=" + fmt.Sprintf("%d", deviceId)
	}
	return DeleteDaoByFilter(common.UserShareDeviceTbl, filter)
}

/******************************************************************************
 * function:
 * description:
 * return {*}
********************************************************************************/
func QueryUserShareDevice(fromUserId int64, toUserId int64, deviceId int64, confirm int, results *[]UserShareDevice) bool {
	sql := "select * from " + common.UserShareDeviceTbl
	filter := ""
	if fromUserId > 0 {
		filter = " from_user_id = " + fmt.Sprintf("%d", fromUserId)
	}
	if toUserId > 0 {
		if filter != "" {
			filter += " and "
		}
		filter += " to_user_id = " + fmt.Sprintf("%d", toUserId)
	}
	if deviceId > 0 {
		if filter != "" {
			filter += " and "
		}
		filter += " device_id = " + fmt.Sprintf("%d", deviceId)
	}
	if confirm > -1 {
		if filter != "" {
			filter += " and "
		}
		filter += " confirm = " + fmt.Sprintf("%d", confirm)
	}
	if filter != "" {
		sql += " where " + filter
	}
	sql += " order by create_time desc"
	rows, err := mDb.Query(sql)
	if err != nil {
		mylog.Log.Errorln(err)
		return false
	}
	defer rows.Close()
	for rows.Next() {
		obj := NewUserShareDevice()
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*results = append(*results, *obj)
		}
	}
	return true
}

// swagger:model UserShareDeviceDetail
type UserShareDeviceDetail struct {
	ID           int64  `json:"id" mysql:"id" binding:"omitempty"`
	FromUserId   int64  `json:"from_user_id" mysql:"from_user_id" `
	FromNickName string `json:"from_nick_name" mysql:"from_nick_name" `
	FromPhone    string `json:"from_phone" mysql:"from_phone" `
	ToUserId     int64  `json:"to_user_id" mysql:"to_user_id" `
	// 分享用户昵称
	ToNickName string `json:"to_nick_name" mysql:"to_nick_name" `
	ToFace     string `json:"to_face" mysql:"to_face" `
	// 分享用户备注
	ToRemark   string `json:"to_remark" mysql:"to_remark" `
	ToPhone    string `json:"to_phone" mysql:"to_phone" `
	DeviceId   int64  `json:"device_id" mysql:"device_id" `
	DeviceName string `json:"device_name" mysql:"device_name" `
	Mac        string `json:"mac" mysql:"mac" `
	DeviceType string `json:"device_type" mysql:"device_type" `
	CreateTime string `json:"create_time" mysql:"create_time" `
}

func (me *UserShareDeviceDetail) DecodeFromRows(rows *sql.Rows) error {
	err := rows.Scan(
		&me.ID,
		&me.FromUserId,
		&me.FromNickName,
		&me.FromPhone,
		&me.ToUserId,
		&me.ToNickName,
		&me.ToFace,
		&me.ToRemark,
		&me.ToPhone,
		&me.DeviceId,
		&me.DeviceName,
		&me.Mac,
		&me.DeviceType,
		&me.CreateTime,
	)
	return err
}
func (me *UserShareDeviceDetail) DecodeFromRow(row *sql.Row) error {
	err := row.Scan(
		&me.ID,
		&me.FromUserId,
		&me.FromNickName,
		&me.FromPhone,
		&me.ToUserId,
		&me.ToNickName,
		&me.ToFace,
		&me.ToRemark,
		&me.ToPhone,
		&me.DeviceId,
		&me.DeviceName,
		&me.Mac,
		&me.DeviceType,
		&me.CreateTime,
	)
	return err
}

func QueryUserShareDeviceDetail(fromUserId int64, toUserId int64, deviceId int64, confirm int, result *[]UserShareDeviceDetail) bool {
	sqlStr := "select a.id, a.from_user_id, b.nick_name as from_nick_name, b.phone as from_phone, a.to_user_id, c.nick_name as to_nick_name, c.face as to_face, a.remark as to_remark, c.phone as to_phone, a.device_id, d.name as device_name, d.mac, d.type as device_type, a.create_time from " +
		common.UserShareDeviceTbl + " a, " +
		common.UserTbl + " b, " +
		common.UserTbl + " c, " +
		common.DeviceTbl + " d where a.from_user_id=b.id and a.to_user_id=c.id and a.device_id=d.id "

	if fromUserId > 0 {
		sqlStr += " and a.from_user_id=" + fmt.Sprintf("%d", fromUserId)
	}
	if toUserId > 0 {
		sqlStr += " and a.to_user_id=" + fmt.Sprintf("%d", toUserId)
	}
	if deviceId > 0 {
		sqlStr += " and a.device_id=" + fmt.Sprintf("%d", deviceId)
	}
	if confirm != -1 {
		sqlStr += " and a.confirm=" + fmt.Sprintf("%d", confirm)
	}
	sqlStr += " order by a.create_time desc"

	rows, err := mDb.Query(sqlStr)
	if err != nil {
		mylog.Log.Errorln(err)
		return false
	}
	defer rows.Close()
	for rows.Next() {
		obj := &UserShareDeviceDetail{}
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*result = append(*result, *obj)
		}
	}
	return true
}

/******************************************************************************
 * function: UserTransferDevice
 * description: 设备过户类，与UserShareDevice结构一样
 * return {*}
********************************************************************************/
// swagger:model UserTransferDevice
type UserTransferDevice UserShareDevice

func NewUserTransferDevice() *UserTransferDevice {
	return &UserTransferDevice{
		ID:         0,
		FromUserId: 0,
		ToUserId:   0,
		DeviceId:   0,
		Confirm:    0,
		Remark:     "",
		CreateTime: time.Now().Format(cfg.TmFmtStr),
	}
}
func (me *UserTransferDevice) MyTableName() string {
	return common.UserTransferDeviceTbl
}

func (me *UserTransferDevice) DecodeFromRows(rows *sql.Rows) error {
	err := rows.Scan(
		&me.ID,
		&me.FromUserId,
		&me.ToUserId,
		&me.DeviceId,
		&me.Confirm,
		&me.Remark,
		&me.CreateTime)
	return err
}
func (me *UserTransferDevice) DecodeFromRow(row *sql.Row) error {
	err := row.Scan(
		&me.ID,
		&me.FromUserId,
		&me.ToUserId,
		&me.DeviceId,
		&me.Confirm,
		&me.Remark,
		&me.CreateTime)
	return err
}

/*
Decode 解析从gin获取的数据 转换成UserDevice
*/
func (me *UserTransferDevice) DecodeFromGin(c *gin.Context) {
	if err := c.ShouldBindBodyWith(me, binding.JSON); err != nil {
		exception.Throw(common.JsonError, err.Error())
	}
	if me.FromUserId == 0 {
		exception.Throw(common.ParamError, "user id is empty!")
	}
}

/*
QueryByID() 查询股票基本信息
*/
func (me *UserTransferDevice) QueryByID(id int64) bool {
	return QueryDaoByID(me.MyTableName(), id, me)
}

/*
Insert UserDevice
*/
func (me *UserTransferDevice) Insert() bool {
	if !CheckTableExist(me.MyTableName()) {
		CreateTableWithStruct(me.MyTableName(), me)
	}
	return InsertDao(me.MyTableName(), me)
}

/*
Update() 更新
*/
func (me *UserTransferDevice) Update() bool {
	return UpdateDaoByID(me.MyTableName(), me.ID, me)
}

/*
Delete() 删除
*/
func (me *UserTransferDevice) Delete() bool {
	return DeleteDaoByID(me.MyTableName(), me.ID)
}

/*
设置ID
*/
func (me *UserTransferDevice) SetID(id int64) {
	me.ID = id
}

/******************************************************************************
 * function: DeleteUserTransferDeviceByUserId
 * description: 根据用户id删除设备过户记录
 * param {int64} fromUserId
 * param {int64} toUserId
 * param {int64} deviceId
 * return {*}
********************************************************************************/
func DeleteUserTransferDeviceByUserId(fromUserId int64, toUserId int64, deviceId int64) bool {
	filter := ""
	if fromUserId > 0 {
		filter += " from_user_id=" + fmt.Sprintf("%d", fromUserId)
	}
	if toUserId > 0 {
		filter += " and to_user_id=" + fmt.Sprintf("%d", toUserId)
	}
	if deviceId > 0 {
		filter += " and device_id=" + fmt.Sprintf("%d", deviceId)
	}
	return DeleteDaoByFilter(common.UserTransferDeviceTbl, filter)
}

/******************************************************************************
 * function:
 * description:
 * return {*}
********************************************************************************/
func QueryUserTransferDevice(fromUserId int64, toUserId int64, deviceId int64, confirm int, results *[]UserTransferDevice) bool {
	sql := "select * from " + common.UserTransferDeviceTbl
	filter := ""
	if fromUserId != 0 {
		filter = " from_user_id = " + fmt.Sprintf("%d", fromUserId)
	}
	if toUserId != 0 {
		if filter != "" {
			filter += " and "
		}
		filter += " to_user_id = " + fmt.Sprintf("%d", toUserId)
	}
	if deviceId != 0 {
		if filter != "" {
			filter += " and "
		}
		filter += " device_id = " + fmt.Sprintf("%d", deviceId)
	}
	if confirm != -1 {
		if filter != "" {
			filter += " and "
		}
		filter += " confirm = " + fmt.Sprintf("%d", confirm)
	}
	if filter != "" {
		sql += " where " + filter
	}
	sql += " order by create_time desc"
	rows, err := mDb.Query(sql)
	if err != nil {
		mylog.Log.Errorln(err)
		return false
	}
	defer rows.Close()
	for rows.Next() {
		obj := NewUserTransferDevice()
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*results = append(*results, *obj)
		}
	}
	return true
}

/******************************************************************************
 * function: UserTransferDeviceDetail
 * description: 定义设备过户详情表，与UserShareDeviceDetail结构一样
 * return {*}
********************************************************************************/
// swagger:model UserTransferDeviceDetail
type UserTransferDeviceDetail UserShareDeviceDetail

func (me *UserTransferDeviceDetail) DecodeFromRows(rows *sql.Rows) error {
	err := rows.Scan(
		&me.ID,
		&me.FromUserId,
		&me.FromNickName,
		&me.FromPhone,
		&me.ToUserId,
		&me.ToNickName,
		&me.ToFace,
		&me.ToRemark,
		&me.ToPhone,
		&me.DeviceId,
		&me.DeviceName,
		&me.Mac,
		&me.DeviceType,
		&me.CreateTime,
	)
	return err
}
func (me *UserTransferDeviceDetail) DecodeFromRow(row *sql.Row) error {
	err := row.Scan(
		&me.ID,
		&me.FromUserId,
		&me.FromNickName,
		&me.FromPhone,
		&me.ToUserId,
		&me.ToNickName,
		&me.ToFace,
		&me.ToRemark,
		&me.ToPhone,
		&me.DeviceId,
		&me.DeviceName,
		&me.Mac,
		&me.DeviceType,
		&me.CreateTime,
	)
	return err
}

func QueryUserTransferDeviceDetail(fromUserId int64, toUserId int64, deviceId int64, confirm int, result *[]UserTransferDeviceDetail) bool {
	sqlStr := "select a.id, a.from_user_id, b.nick_name as from_nick_name, b.phone as from_phone, a.to_user_id, c.nick_name as to_nick_name, c.face as to_face, a.remark as to_remark, c.phone as to_phone, a.device_id, d.name as device_name, d.mac, d.type as device_type, a.create_time from " +
		common.UserTransferDeviceTbl + " a, " +
		common.UserTbl + " b, " +
		common.UserTbl + " c, " +
		common.DeviceTbl + " d where a.from_user_id=b.id and a.to_user_id=c.id and a.device_id=d.id "

	if fromUserId > 0 {
		sqlStr += " and a.from_user_id=" + fmt.Sprintf("%d", fromUserId)
	}
	if toUserId > 0 {
		sqlStr += " and a.to_user_id=" + fmt.Sprintf("%d", toUserId)
	}
	if deviceId > 0 {
		sqlStr += " and a.device_id=" + fmt.Sprintf("%d", deviceId)
	}
	if confirm != -1 {
		sqlStr += " and a.confirm=" + fmt.Sprintf("%d", confirm)
	}
	sqlStr += " order by a.create_time desc"

	rows, err := mDb.Query(sqlStr)
	if err != nil {
		mylog.Log.Errorln(err)
		return false
	}
	defer rows.Close()
	for rows.Next() {
		obj := &UserTransferDeviceDetail{}
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*result = append(*result, *obj)
		}
	}
	return true
}
