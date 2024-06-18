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
	LoginType     int    `json:"login_type" mysql:"login_type"` // 0:phone 1:email
	Phone         string `json:"phone" mysql:"phone"`
	Email         string `json:"email" mysql:"email"`
	EmergentPhone string `json:"emergent_phone" mysql:"emergent_phone"`
	Face          string `json:"face" mysql:"face"`
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
		LoginType:     0,
		Phone:         "",
		Email:         "",
		EmergentPhone: "",
		Face:          "",
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
	err := rows.Scan(&me.ID, &me.Account, &me.Password, &me.NickName, &me.LoginType, &me.Phone, &me.Email, &emergentPhone, &me.Face, &me.Address, &me.RoomNum, &me.IsLogin, &me.LoginTime, &me.CreateTime)
	if emergentPhone.Valid {
		me.EmergentPhone = emergentPhone.String
	}
	return err
}
func (me *User) DecodeFromRow(row *sql.Row) error {
	var emergentPhone sql.NullString
	err := row.Scan(&me.ID, &me.Account, &me.Password, &me.NickName, &me.LoginType, &me.Phone, &me.Email, &emergentPhone, &me.Face, &me.Address, &me.RoomNum, &me.IsLogin, &me.LoginTime, &me.CreateTime)
	if emergentPhone.Valid {
		me.EmergentPhone = emergentPhone.String
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
			login_type int NOT NULL COMMENT '登录类型 0:phone 1:email',
            phone varchar(32) comment '手机号',
			email varchar(32) comment '邮箱',
			emergent_phone varchar(32) comment '紧急联系电话',
            face varchar(128) comment '头像',
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
	UserId     int64  `json:"user_id" mysql:"user_id" binding:"required"`
	DeviceId   int64  `json:"device_id" mysql:"device_id" binding:"required"`
	Flag       int    `json:"flag" mysql:"flag"`
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
		sql := `create table ` + tblName + ` (
            id MEDIUMINT NOT NULL AUTO_INCREMENT,
            user_id MEDIUMINT NOT NULL COMMENT '用户id',
			device_id MEDIUMINT NOT NULL COMMENT '设备id',
			flag int NOT NULL COMMENT '标志 0:自己创建 1:共享',
            create_time datetime comment '创建时间',
            PRIMARY KEY (id, user_id, create_time)
        )`
		CreateTable(sql)
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
	sql := "select a.user_id as user_id, a.flag, b.* from " +
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
	err := rows.Scan(&me.UserId, &me.Flag, &me.ID, &me.Name, &me.Type, &me.Mac, &me.RoomNum, &me.Online, &me.OnlineTime, &me.CreateTime, &me.Remark)
	return err
}
func (me *UserDevice) DecodeFromRow(row *sql.Row) error {
	err := row.Scan(&me.UserId, &me.Flag, &me.ID, &me.Name, &me.Type, &me.Mac, &me.RoomNum, &me.Online, &me.OnlineTime, &me.CreateTime, &me.Remark)
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
