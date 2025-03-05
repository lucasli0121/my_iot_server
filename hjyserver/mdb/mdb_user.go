/******************************************************************************
 * Author: liguoqiang
 * Date: 2024-08-27 18:47:25
 * LastEditors: liguoqiang
 * LastEditTime: 2024-12-20 19:59:50
 * Description:
********************************************************************************/
/******************************************************************************
 * Author: liguoqiang
 * Date: 2023-09-06 17:50:12
 * LastEditors: liguoqiang
 * LastEditTime: 2024-09-11 17:41:50
 * Description:
********************************************************************************/
package mdb

import (
	"fmt"
	"hjyserver/cfg"
	"hjyserver/mdb/common"
	"hjyserver/mdb/mysql"
	"hjyserver/mq"
	mysqlwx "hjyserver/wx/mdb/mysql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// swagger:model VerifyTokenResp
type VerifyTokenResp struct {
	Token string `json:"token"`
	// 0: 过期 1: 有效
	Result int `json:"result"`
}

func VerifyUserToken(c *gin.Context) (int, interface{}) {
	token := c.Query("token")
	if token == "" {
		return common.ParamError, "token required"
	}
	resp := VerifyTokenResp{}
	resp.Token = token
	resp.Result = func() int {
		if mysql.VerifyUserToken(token) {
			return 1
		}
		return 0
	}()
	return common.Success, resp
}

// swagger:model LoginReq
type LoginReq struct {
	// required: true
	// example: 1
	Account string `json:"account" mysql:"account"`
	// required: true
	// example: 1
	Passwd string `json:"passwd" mysql:"passwd"`
	// required: false
	// example: 1
	LoginType int `json:"login_type" mysql:"login_type"` // 0:phone 1:email
}

/******************************************************************************
 * function: UserLogin
 * description: user login
 * param {*gin.Context} c
 * return {*}
********************************************************************************/
func UserLogin(c *gin.Context) (int, interface{}) {
	me := LoginReq{}
	if err := c.ShouldBindJSON(&me); err != nil {
		return common.ParamError, "json format error"
	}
	if me.Account == "" {
		return common.ParamError, "account required"
	}
	if me.Account != "guest" && me.Passwd == "" {
		return common.ParamError, "password required"
	}

	filter := fmt.Sprintf("account = '%s'", me.Account)
	var gList []mysql.User
	mysql.QueryUserByCond(filter, nil, nil, &gList)
	if len(gList) > 0 {
		obj := gList[0]
		if obj.Account == "guest" {
			return http.StatusOK, obj
		}
		passwd, _ := common.EncryptDataWithDefaultkey(me.Passwd)
		if passwd == obj.Password {
			obj.IsLogin = 1
			obj.LoginTime = common.GetNowTime()
			obj.Update()
			return http.StatusOK, obj
		}
		return common.PasswdError, "password error!"
	}
	return common.NoExist, "account is not exist!"
}

/******************************************************************************
 * function: UserRegister
 * description: register a new user
 * return {*}
********************************************************************************/
func UserRegister(c *gin.Context) (int, interface{}) {
	me := mysql.NewUser()
	me.DecodeFromGin(c)
	return mysql.RegisterWithUserObj(me)
}

/******************************************************************************
 * function: LoginOut
 * description: login in or register a new user
 * return {*}
********************************************************************************/
// swagger:model LoginoutReq
type LoginoutReq struct {
	// required: false
	// example: 1
	ID int64 `json:"id"`
}

func LoginOut(c *gin.Context) (int, interface{}) {
	var req *LoginoutReq = &LoginoutReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		return http.StatusBadRequest, "json format error"
	}
	me := mysql.NewUser()
	me.ID = req.ID
	if me.ID != 0 {
		me.QueryByID(me.ID)
		me.IsLogin = 0
		me.Update()
		return http.StatusOK, me
	} else if me.Account != "" {
		var gList []mysql.User
		mysql.QueryUserByCond(fmt.Sprintf("account = '%s'", me.Account), nil, nil, &gList)
		if len(gList) > 0 {
			obj := gList[0]
			obj.IsLogin = 0
			obj.Update()
			return http.StatusOK, obj
		}
	}
	return http.StatusAccepted, "loginout failed!"
}

// swagger:model DeleteUserReq
type DeleteUserReq struct {
	// required: false
	// example: 1
	ID int64 `json:"id"`
}

func DeleteUser(c *gin.Context) (int, interface{}) {
	var req *DeleteUserReq = &DeleteUserReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		return common.JsonError, "json format error"
	}
	me := mysql.NewUser()
	me.ID = req.ID
	if me.Delete() {
		mysql.DeleteDeviceRelationByUserId(me.ID)
		mysql.DeleteUserRelationByUserId(me.ID, 0)
		// 删除用户分享过的记录
		mysql.DeleteUserShareDeviceByUserId(me.ID, 0, 0)
		// 删除用户被分享过的记录
		mysql.DeleteUserShareDeviceByUserId(0, me.ID, 0)
		// 删除此用户过户出去的记录
		mysql.DeleteUserTransferDeviceByUserId(me.ID, 0, 0)
		// 删除此用户过户进来的记录
		mysql.DeleteUserTransferDeviceByUserId(0, me.ID, 0)
		if cfg.This.Svr.EnableWx {
			mysqlwx.DeleteMiniProgramByUserId(me.ID)
		}
		return common.Success, "delete user success!"
	}
	return common.DBError, "delete failed!"
}

/******************************************************************************
 * function: UserOnline
 * description:
 * return {*}
********************************************************************************/
func UserOnline(c *gin.Context) (int, interface{}) {
	me := mysql.NewUser()
	me.DecodeFromGin(c)
	if me.ID != 0 {
		me.QueryByID(me.ID)
		me.IsLogin = 1
		me.LoginTime = common.GetNowTime()
		me.Update()
		return http.StatusOK, me
	} else if me.Account != "" {
		var gList []mysql.User
		mysql.QueryUserByCond(fmt.Sprintf("account = '%s'", me.Account), nil, nil, &gList)
		if len(gList) > 0 {
			obj := gList[0]
			obj.IsLogin = 1
			obj.LoginTime = common.GetNowTime()
			obj.Update()
			return http.StatusOK, obj
		}
	}
	return http.StatusAccepted, "user online failed!"
}

/******************************************************************************
 * function: UserOffline
 * description:
 * return {*}
********************************************************************************/
func UserOffline(c *gin.Context) (int, interface{}) {
	me := mysql.NewUser()
	me.DecodeFromGin(c)
	if me.ID != 0 {
		me.QueryByID(me.ID)
		me.IsLogin = 0
		me.Update()
		return http.StatusOK, me
	} else if me.Account != "" {
		var gList []mysql.User
		mysql.QueryUserByCond(fmt.Sprintf("account = '%s'", me.Account), nil, nil, &gList)
		if len(gList) > 0 {
			obj := gList[0]
			obj.IsLogin = 0
			obj.Update()
			return http.StatusOK, obj
		}
	}
	return http.StatusAccepted, "user offline failed!"
}

// swagger:model NickNameReq
type NickNameReq struct {
	// required: true
	// example: 1
	UserId   int64  `json:"user_id"`
	NickName string `json:"nick_name"`
}

/******************************************************************************
 * function: UpdateNickName
 * description: update user information
 * return {*}
********************************************************************************/
func UpdateNickName(c *gin.Context) (int, interface{}) {
	var req *NickNameReq = &NickNameReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		return common.ParamError, "json format error"
	}
	me := mysql.NewUser()
	me.ID = req.UserId
	if me.QueryByID(me.ID) {
		me.NickName = req.NickName
		if me.Update() {
			return common.Success, me
		}
	}
	return common.DBError, "update failed!"
}

// swagger:model UserGenderReq
type UserGenderReq struct {
	// required: true
	// example: 1
	UserId int64 `json:"user_id"`
	Gender int   `json:"gender"`
}

/******************************************************************************
 * function:
 * description:
 * param {*gin.Context} c
 * return {*}
********************************************************************************/
func UpdateGender(c *gin.Context) (int, interface{}) {
	var req *UserGenderReq = &UserGenderReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		return common.ParamError, "json format error"
	}
	me := mysql.NewUser()
	me.ID = req.UserId
	if me.QueryByID(me.ID) {
		me.Gender = req.Gender
		if me.Update() {
			return common.Success, me
		}
	}
	return common.DBError, "update failed!"
}

// swagger:model UserFacePicReq
type UserFacePicReq struct {
	// required: true
	// example: 1
	UserId int64  `json:"user_id"`
	Face   string `json:"face"`
}

/**
 * @description: update user head picture
 * @param {*gin.Context} c
 * @return {*}
 */
func UpdateHeadPic(c *gin.Context) (int, interface{}) {
	var req *UserFacePicReq = &UserFacePicReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		return common.ParamError, "json format error"
	}
	me := mysql.NewUser()
	me.ID = req.UserId
	if me.QueryByID(me.ID) {
		me.Face = req.Face
		if me.Update() {
			return common.Success, me
		}
	}
	return common.DBError, "update failed!"
}

/******************************************************************************
 * function:
 * description:
 * param {*gin.Context} c
 * return {*}
********************************************************************************/
func UpdateUser(c *gin.Context) (int, interface{}) {
	me := mysql.NewUser()
	me.DecodeFromGin(c)
	if me.ID != 0 {
		obj := mysql.NewUser()
		obj.ID = me.ID
		if obj.QueryByID(me.ID) {
			if obj.Account != me.Account {
				return common.NoPermission, "account can not be modified"
			}
		}
		me.Update()
		return common.Success, me
	}
	return common.DBError, "update failed!"
}

/******************************************************************************
 * function: ModifyPhone
 * description: modify user's phone information
 * return {*}
********************************************************************************/

// swagger:model NewPhone
type NewPhone struct {
	UserId int64  `json:"user_id"`
	Phone  string `json:"phone"`
}

func ModifyPhone(c *gin.Context) (int, interface{}) {
	var newPhone = &NewPhone{
		UserId: 0,
		Phone:  "",
	}
	if err := c.ShouldBindJSON(newPhone); err != nil {
		return http.StatusBadRequest, "json format error"
	}
	if newPhone.UserId == 0 || newPhone.Phone == "" {
		return http.StatusBadRequest, "user id and phone required"
	}
	newPhone.Phone = common.FixPlusInPhoneString(newPhone.Phone)
	var gList []mysql.User
	mysql.QueryUserByCond(fmt.Sprintf("phone = '%s'", newPhone.Phone), nil, nil, &gList)
	if len(gList) > 0 {
		return http.StatusBadRequest, "new phone has registered"
	}

	me := mysql.NewUser()
	me.SetID(newPhone.UserId)
	if me.ID != 0 {
		me.QueryByID(me.ID)
		me.Phone = newPhone.Phone
		me.Update()
		return http.StatusOK, me
	}
	return http.StatusAccepted, "update failed!"
}

/******************************************************************************
 * function: ModifyEmergentPhone
 * description: modify user emergent phone
 * param {*gin.Context} c
 * return {*}
********************************************************************************/
func ModifyEmergentPhone(c *gin.Context) (int, interface{}) {
	var newPhone = &NewPhone{
		UserId: 0,
		Phone:  "",
	}
	if err := c.ShouldBindJSON(newPhone); err != nil {
		return http.StatusBadRequest, "json format error"
	}
	if newPhone.UserId == 0 || newPhone.Phone == "" {
		return http.StatusBadRequest, "user id and phone required"
	}
	newPhone.Phone = common.FixPlusInPhoneString(newPhone.Phone)
	me := mysql.NewUser()
	me.SetID(newPhone.UserId)
	if me.ID != 0 {
		me.QueryByID(me.ID)
		me.EmergentPhone = newPhone.Phone
		me.Update()
		return http.StatusOK, me
	}
	return http.StatusAccepted, "update failed!"
}

// swagger:model NewEmail
type NewEmail struct {
	UserId int64  `json:"user_id"`
	Email  string `json:"email"`
}

func ModifyEmail(c *gin.Context) (int, interface{}) {
	var newEmail = &NewEmail{
		UserId: 0,
		Email:  "",
	}
	if err := c.ShouldBindJSON(newEmail); err != nil {
		return common.ParamError, "json format error"
	}
	if newEmail.UserId == 0 || newEmail.Email == "" {
		return common.ParamError, "user id and email required"
	}
	var gList []mysql.User
	mysql.QueryUserByCond(fmt.Sprintf("email = '%s'", newEmail.Email), nil, nil, &gList)
	if len(gList) > 0 {
		return common.RepeatData, "new email has registered"
	}

	me := mysql.NewUser()
	me.SetID(newEmail.UserId)
	if me.ID != 0 {
		me.QueryByID(me.ID)
		me.Email = newEmail.Email
		me.Update()
		return common.Success, me
	}
	return common.DBError, "update failed!"
}

// swagger:model NewPasswd
type NewPasswd struct {
	UserId int64  `json:"user_id"`
	Passwd string `json:"password"`
}

/******************************************************************************
 * function:
 * description:
 * param {*gin.Context} c
 * return {*}
********************************************************************************/
func ModifyPasswd(c *gin.Context) (int, interface{}) {
	var newPasswd = &NewPasswd{
		UserId: 0,
		Passwd: "",
	}
	if err := c.ShouldBindJSON(newPasswd); err != nil {
		return common.ParamError, "json format error"
	}
	if newPasswd.UserId == 0 || newPasswd.Passwd == "" {
		return common.ParamError, "user id and passwd required"
	}
	me := mysql.NewUser()
	me.ID = newPasswd.UserId
	if !me.QueryByID(newPasswd.UserId) {
		return common.NoExist, "user record not exist"
	}
	me.Password, _ = common.EncryptDataWithDefaultkey(newPasswd.Passwd)
	me.Update()
	return http.StatusOK, me
}

// swagger:model CheckPhoneReq
type CheckPhoneReq struct {
	UserId int64  `json:"user_id"`
	Phone  string `json:"phone"`
}

func CheckUserPhone(c *gin.Context) (int, interface{}) {
	var checkReq = &CheckPhoneReq{
		UserId: 0,
		Phone:  "",
	}
	if err := c.ShouldBindJSON(checkReq); err != nil {
		return common.ParamError, "json format error"
	}
	if checkReq.UserId == 0 || checkReq.Phone == "" {
		return common.ParamError, "user id and phone required"
	}
	me := mysql.NewUser()
	me.ID = checkReq.UserId
	if !me.QueryByID(checkReq.UserId) {
		return common.NoExist, "user record not exist"
	}
	if me.Phone != checkReq.Phone {
		return common.PhoneNotMatch, "phone not match"
	}
	return common.Success, me
}

/***********************************************************
* func QueryAllUsers()
********************************************************/
func QueryAllUsers(c *gin.Context) (int, interface{}) {
	var gList []mysql.User
	mysql.QueryAllUsers(&gList)
	return http.StatusOK, gList
}

/***********************************************************
* func QueryUserById()
********************************************************/
func QueryUserById(c *gin.Context) (int, interface{}) {
	userId := c.Query("id")
	if userId == "" {
		return http.StatusBadRequest, "user id required"
	}
	me := mysql.NewUser()
	mId, err := strconv.ParseInt(userId, 10, 64)
	if err == nil && me.QueryByID(mId) {
		return http.StatusOK, me
	}
	return http.StatusAccepted, "query failed"
}

/**
 * @description: query user by phone
 * @return {*}
 */
func QueryUserByPhone(c *gin.Context) (int, interface{}) {
	phone := c.Query("phone")
	if phone == "" {
		return common.ParamError, "phone number required"
	}
	phone = common.FixPlusInPhoneString(phone)
	var gList []mysql.User
	mysql.QueryUserByCond(fmt.Sprintf("phone like '%%%s%%'", phone), nil, nil, &gList)
	if len(gList) == 0 {
		return common.NoExist, "user is not exist"
	}
	return common.Success, gList
}

/******************************************************************************
 * function:
 * description:
 * param {*gin.Context} c
 * return {*}
********************************************************************************/
func QueryUserByEmail(c *gin.Context) (int, interface{}) {
	email := c.Query("email")
	if email == "" {
		return common.ParamError, "email required"
	}
	var gList []mysql.User
	mysql.QueryUserByCond(fmt.Sprintf("email like '%%%s%%'", email), nil, nil, &gList)
	if len(gList) == 0 {
		return common.NoExist, "user is not exist"
	}
	return common.Success, gList
}

/**
 * @description: QueryUserGroup
 * @return {*}
 */
func QueryUserGroup(c *gin.Context) (int, interface{}) {
	userId := c.Query("user_id")
	if userId == "" {
		return http.StatusBadRequest, "user id required"
	}
	filter := fmt.Sprintf("user_id=%s", userId)
	var gList []mysql.UserGroup
	mysql.QueryGroupByCond(filter, &gList)
	return http.StatusOK, gList
}

/**
 * @description: insert user group information
 * @return {*}
 */
func InsertUserGroup(c *gin.Context) (int, interface{}) {
	body := mysql.NewUserGroup()
	body.DecodeFromGin(c)
	if body.Insert() {
		return http.StatusOK, body
	}
	return http.StatusAccepted, "insert error!"
}

/**
 * @description: delete user group bu userid
 * @return {*}
 */
func DeleteUserGroup(c *gin.Context) (int, interface{}) {
	body := mysql.NewUserGroup()
	body.DecodeFromGin(c)
	if body.Delete() {
		return http.StatusOK, body
	}
	return http.StatusAccepted, "delete error!"
}

/**
 * @function: QueryFriendsByUser
 * @description: query user's friends by user id etc.
 * @return {*}
 */
func QueryFriendsByUser(c *gin.Context) (int, interface{}) {
	userId := c.Query("user_id")
	mUserId, _ := strconv.ParseInt(userId, 10, 64)
	var gList []mysql.UserFriend
	mysql.QueryUserFriendByUserId(mUserId, &gList)
	return http.StatusOK, gList
}

/**
 * @description: insert user friend
 * @return {*}
 */
func InsertUserFriend(c *gin.Context) (int, interface{}) {
	var friend = mysql.NewUserFriend()
	friend.DecodeFromGin(c)
	if friend.UserId == 0 || friend.Phone == "" {
		return http.StatusBadRequest, "user id and phone required"
	}
	var gList []mysql.User
	mysql.QueryUserByCond(fmt.Sprintf("phone = '%s'", friend.Phone), nil, nil, &gList)
	if len(gList) == 0 {
		// var user = mysql.NewUser()
		// user.Phone = friend.Phone
		// user.Account = friend.Phone
		// user.Insert()
		// friend.FriendId = user.ID
		return http.StatusBadRequest, "phone is not registered"
	}
	friend.NickName = gList[0].NickName
	friend.FriendId = gList[0].ID
	friend.CreateTime = common.GetNowTime()
	var fList []mysql.UserRelation
	mysql.QueryUserRelationByCond(fmt.Sprintf("user_id = %d and friend_id = %d", friend.UserId, friend.FriendId), &fList)
	if len(fList) > 0 {
		friend.ID = fList[0].ID
	} else if !friend.Insert() {
		return http.StatusAccepted, "insert error!"
	}
	mysql.QueryUserRelationByCond(fmt.Sprintf("user_id = %d and friend_id = %d", friend.FriendId, friend.UserId), &fList)
	if len(fList) == 0 {
		var f = mysql.UserRelation{}
		f.UserId = friend.FriendId
		f.FriendId = friend.UserId
		f.CreateTime = common.GetNowTime()
		f.Insert()
	}
	// notify friend
	type AddFriendNotify struct {
		UserId   int64  `json:"user_id"`
		FriendId int64  `json:"friend_id"`
		NickName string `json:"nick_name"`
		Phone    string `json:"phone"`
	}
	var user = mysql.NewUser()
	if user.QueryByID(friend.UserId) {
		var notify = &AddFriendNotify{
			UserId:   friend.UserId,
			FriendId: friend.FriendId,
			NickName: user.NickName,
			Phone:    user.Phone,
		}
		mq.PublishData(common.MakeAddFriendNotifyTopic(friend.FriendId), notify)
	}
	return http.StatusOK, friend
}

/******************************************************************************
 * function:
 * description:
 * param {*gin.Context} c
 * return {*}
********************************************************************************/
// swagger:model ModifyFriendReq
type ModifyFriendReq struct {
	// required: true
	UserId   int64  `json:"user_id"`
	FriendId int64  `json:"friend_id"`
	NickName string `json:"nick_name"`
	Face     string `json:"face"`
}

func ModifyUserFriend(c *gin.Context) (int, interface{}) {
	var req *ModifyFriendReq = &ModifyFriendReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		return http.StatusBadRequest, "json format error"
	}
	filter := fmt.Sprintf("user_id=%d and friend_id=%d", req.UserId, req.FriendId)
	var fList []mysql.UserRelation
	if !mysql.QueryUserRelationByCond(filter, &fList) {
		return http.StatusBadRequest, "friend not exist"
	}
	var obj = mysql.NewUser()
	if !obj.QueryByID(req.FriendId) {
		return http.StatusBadRequest, "friend not exist"
	}
	obj.NickName = req.NickName
	obj.Face = req.Face
	if obj.Update() {
		return http.StatusOK, "modify friend success"
	}
	return http.StatusAccepted, "update error!"
}

/******************************************************************************
 * function:
 * description:
 * param {*gin.Context} c
 * return {*}
********************************************************************************/

// swagger:model RemoveFriendReq
type RemoveFriendReq struct {
	// required: true
	UserID int64 `json:"user_id"`
	// required: true
	FriendID int64 `json:"friend_id"`
}

func RemoveUserFriend(c *gin.Context) (int, interface{}) {
	var req *RemoveFriendReq = &RemoveFriendReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		return http.StatusBadRequest, "json format error"
	}
	if mysql.DeleteUserRelationByUserId(req.UserID, req.FriendID) {
		return http.StatusOK, "remove friend success"
	}
	return http.StatusAccepted, "delete error!"
}

/******************************************************************************
 * function: AddUserToStudyRoom
 * description:  add user id into study room
 * param {*gin.Context} c
 * return {*}
********************************************************************************/
// swagger:model UserToStudyRoomReq
type UserToStudyRoomReq struct {
	// required: true
	// example: 1
	CreateId int64 `json:"create_id"`
	// required: true
	// example: 1
	UserId int64 `json:"user_id"`
	// required: true
	// example: 1
	RoomId int64 `json:"room_id"`
	// required: false
	// example: 1
	Sn int `json:"sn"`
}

func AddUserToStudyRoom(c *gin.Context) (int, interface{}) {
	var req *UserToStudyRoomReq = &UserToStudyRoomReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		return http.StatusBadRequest, "json format error"
	}
	if req.CreateId == 0 || req.UserId == 0 || req.RoomId == 0 {
		return http.StatusBadRequest, "create id, user id and room id required"
	}
	var filter string = fmt.Sprintf("id=%d and create_id=%d and status=1", req.RoomId, req.CreateId)
	var objs []mysql.StudyRoom
	if !mysql.QueryStudyRoomByCond(filter, nil, "", 1, &objs) || len(objs) == 0 {
		return common.NoExist, "study room not exist"
	}
	var room = objs[0]
	if room.CurrentNum >= room.Capacity {
		return http.StatusBadRequest, "study room is full"
	}
	if mysql.UserInAnyStudyRoom(req.UserId) {
		return common.HasExist, "user already in other study room"
	}
	var obj *mysql.StudyRoomUser = mysql.NewStudyRoomUser()
	filter = fmt.Sprintf("user_id=%d and room_id=%d and status=1", req.UserId, req.RoomId)
	var gList []mysql.StudyRoomUser
	mysql.QueryStudyRoomUserByCond(filter, nil, nil, 1, &gList)
	if len(gList) > 0 {
		return http.StatusBadRequest, "A user can only join a study room once"
		// obj = &gList[0]
		// obj.Status = 1
		// obj.Sn = req.Sn
		// obj.CreateTime = common.GetNowTime()
		// if obj.Update() {
		// 	return http.StatusOK, obj
		// }
	}
	room.CurrentNum++
	room.Update()
	if req.Sn == 0 {
		req.Sn = room.CurrentNum
	}
	obj.UserId = req.UserId
	obj.RoomId = req.RoomId
	obj.Status = 1
	obj.Sn = req.Sn
	obj.CreateTime = common.GetNowTime()
	if obj.Insert() {
		return http.StatusOK, obj
	}
	return http.StatusBadRequest, "insert error!"
}

/******************************************************************************
 * function: RemoveUserFromStudyRoom
 * description: remove user id from study room
 * param {*gin.Context} c
 * return {*}
********************************************************************************/
func RemoveUserFromStudyRoom(c *gin.Context) (int, interface{}) {
	var req *UserToStudyRoomReq = &UserToStudyRoomReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		return http.StatusBadRequest, "json format error"
	}
	if req.CreateId == 0 || req.UserId == 0 || req.RoomId == 0 {
		return http.StatusBadRequest, "create id, user id and room id required"
	}
	var filter string = fmt.Sprintf("id=%d and create_id=%d and status=1", req.RoomId, req.CreateId)
	var objs []mysql.StudyRoom
	if !mysql.QueryStudyRoomByCond(filter, nil, "", 1, &objs) || len(objs) == 0 {
		return http.StatusBadRequest, "study room not exist"
	}
	if objs[0].CurrentNum > 0 {
		objs[0].CurrentNum--
		objs[0].Update()
	}
	mysql.CleanUserStudyRoomStatus(req.UserId, req.RoomId)
	// mysql.CleanStudyRecordStatus(req.UserId, req.RoomId)
	return http.StatusOK, "remove user from study room success"
}

/******************************************************************************
 * function: QueryStudyRoomUser
 * description:  query user list from study room
 * param {*gin.Context} c
 * return {*}
********************************************************************************/
func QueryStudyRoomUser(c *gin.Context) (int, interface{}) {
	createId := c.Query("create_id")
	roomId := c.Query("room_id")
	flag := c.Query("flag")
	if roomId == "" {
		roomId = "0"
		// return http.StatusBadRequest, "room id required"
	}
	if createId == "" {
		createId = "0"
	}
	roomIdInt, err := strconv.ParseInt(roomId, 10, 64)
	if err != nil {
		roomIdInt = 0
	}
	createIdInt, err := strconv.ParseInt(createId, 10, 64)
	if err != nil {
		createIdInt = 0
	}
	if roomIdInt > 0 {
		var filter string
		if createId == "" || createId == "0" {
			filter = fmt.Sprintf("id=%s and status=1", roomId)
		} else {
			filter = fmt.Sprintf("id=%s and create_id=%s and status=1", roomId, createId)
		}
		var objs []mysql.StudyRoom
		if !mysql.QueryStudyRoomByCond(filter, nil, "", 1, &objs) || len(objs) == 0 {
			return http.StatusBadRequest, "study room not exist"
		}
	}
	var gList []mysql.StudyRoomUserDetail
	flagInt, _ := strconv.ParseInt(flag, 10, 32)
	mysql.QueryStudyRoomUserDetailByRoomId(roomIdInt, createIdInt, int(flagInt), &gList)
	return http.StatusOK, gList
}

/******************************************************************************
 * function: QueryUserStudyData
 * description:
 * param {*gin.Context} c
 * return {*}
********************************************************************************/

// swagger:model UserStudyData
type UserStudyData struct {
	TotalData []mysql.UserStudyRoomData  `json:"total_data"`
	DayData   []mysql.UserStudyDataByDay `json:"day_data"`
}

func QueryUserStudyData(c *gin.Context) (int, interface{}) {
	userId := c.Query("user_id")
	roomId := c.Query("room_id")
	if userId == "" || roomId == "" {
		return http.StatusBadRequest, "user id and room id required"
	}
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")
	if startTime == "" || endTime == "" {
		return http.StatusBadRequest, "start time and end time required"
	}
	var userData = &UserStudyData{}
	userId1, _ := strconv.ParseInt(userId, 10, 64)
	roomId1, _ := strconv.ParseInt(roomId, 10, 64)
	mysql.QueryUserStudyRoomData(userId1, roomId1, startTime, endTime, &userData.TotalData)
	mysql.QueryUserStudyDataByDay(userId1, roomId1, startTime, endTime, &userData.DayData)
	return http.StatusOK, userData
}

/******************************************************************************
 * function: QueryUserStudyTimeByDay
 * description:
 * param {*gin.Context} c
 * return {*}
********************************************************************************/
func QueryUserStudyTimeByDay(c *gin.Context) (int, interface{}) {
	userId := c.Query("user_id")
	roomId := c.Query("room_id")
	if userId == "" || roomId == "" {
		return http.StatusBadRequest, "user id and room id required"
	}
	startDate := c.Query("query_date")
	if startDate == "" {
		return http.StatusBadRequest, "query date required"
	}
	var gList []mysql.UserStudyTime
	userId1, _ := strconv.ParseInt(userId, 10, 64)
	roomId1, _ := strconv.ParseInt(roomId, 10, 64)
	mysql.QueryUserStudyTimeByDate(userId1, roomId1, startDate, startDate, &gList)
	return http.StatusOK, gList
}

/******************************************************************************
 * function: QueryLampUsersByRoom
 * description:
 * param {*gin.Context} c
 * return {*}
********************************************************************************/
func QueryLampUsersByRoom(c *gin.Context) (int, interface{}) {
	roomId := c.Query("room_id")
	if roomId == "" {
		return http.StatusBadRequest, "room id required"
	}
	var gList []mysql.LampUserWithStudyRoom
	mysql.QueryLampUsersDetailByRoomId(roomId, &gList)
	return http.StatusOK, gList
}

/**
 * @description: query user who has bind lamp in friend list
 * @param {*gin.Context} c
 * @return {*}
 */
func QueryLampUserInFriend(c *gin.Context) (int, interface{}) {
	userId := c.Query("user_id")
	if userId == "" {
		return common.ParamError, "user id required"
	}
	var gList []mysql.LampUserWithStudyRoom
	mysql.QueryLampUsersInFriend(userId, &gList)
	return common.Success, gList
}

/******************************************************************************
 * function: EnterStudyRoom
 * description: user enter study room
 * param {*gin.Context} c
 * return {*}
********************************************************************************/

// swagger:model UserEnterStudyReq
type UserEnterStudyReq struct {
	// required: true
	// example: 1
	UserId int64 `json:"user_id"`
	// required: true
	// example: 1
	RoomId int64 `json:"room_id"`
}

func EnterStudyRoom(c *gin.Context) (int, interface{}) {
	var req *UserEnterStudyReq = &UserEnterStudyReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		return http.StatusBadRequest, "json format error"
	}
	if req.UserId == 0 || req.RoomId == 0 {
		return http.StatusBadRequest, "user id and room id required"
	}
	var filter string = fmt.Sprintf("user_id=%d and room_id=%d and status=1", req.UserId, req.RoomId)
	var objs []mysql.StudyRoomUser
	if !mysql.QueryStudyRoomUserByCond(filter, nil, nil, 1, &objs) || len(objs) == 0 {
		return http.StatusBadRequest, "study room not exist or user not been invited"
	}

	// mysql.CleanStudyRecordStatus(req.UserId, req.RoomId)
	obj := mysql.NewUserStudyRecord()
	obj.UserId = req.UserId
	obj.RoomId = req.RoomId
	obj.Status = 1
	obj.Sn = objs[0].Sn
	obj.EnterTime = common.GetNowTime()
	obj.LeaveTime = common.GetNowTime()
	if obj.Insert() {
		return http.StatusOK, obj
	}
	return http.StatusBadRequest, "enter room failed!"
}

/******************************************************************************
 * function: LeaveStudyRoom
 * description: user enter study room
 * param {*gin.Context} c
 * return {*}
********************************************************************************/
func LeaveStudyRoom(c *gin.Context) (int, interface{}) {
	var req *UserEnterStudyReq = &UserEnterStudyReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		return http.StatusBadRequest, "json format error"
	}
	if req.UserId == 0 || req.RoomId == 0 {
		return http.StatusBadRequest, "user id and room id required"
	}

	mysql.CleanStudyRecordStatus(req.UserId, req.RoomId)

	// var filter string = fmt.Sprintf("user_id=%d and room_id=%d and status=1", req.UserId, req.RoomId)
	// var objs []mysql.UserStudyRecord
	// mysql.QueryUserStudyRecordByCond(filter, nil, nil, 1, &objs)
	// if len(objs) > 0 {
	// 	obj := objs[0]
	// 	obj.Status = 0
	// 	obj.LeaveTime = common.GetNowTime()
	// 	if obj.Update() {
	// 		return http.StatusOK, obj
	// 	}
	// }
	return http.StatusOK, "leave room success!"
}

/******************************************************************************
 * function: QueryUserOverview
 * description: 根据用户ID查询用户概况
 * param {*gin.Context} c
 * return {*}
********************************************************************************/
// swagger:model UserOverview
type UserOverview struct {
	// required: true
	UserId int64 `json:"user_id"`
	// 0 未知 1 男 2 女
	Gender int `json:"gender" mysql:"gender"`
	// 出生日期
	BornDate string `json:"born_date"`
	// 年级
	Grade string `json:"grade"`
}

func QueryUserOverview(c *gin.Context) (int, interface{}) {
	userId := c.Query("user_id")
	if userId == "" {
		return common.ParamError, "user id required"
	}
	mId, err := strconv.ParseInt(userId, 10, 64)
	if err != nil {
		return common.ParamError, "user id format error"
	}
	userOverview := &UserOverview{
		UserId: mId,
	}
	user := mysql.NewUser()
	if user.QueryByID(mId) {
		userOverview.BornDate = user.BornDate
		userOverview.Gender = user.Gender
		userOverview.Grade = user.Grade
	} else {
		return common.NoExist, "user not exist"
	}
	return http.StatusOK, userOverview
}

/******************************************************************************
 * function: UpdateUserOverview
 * description: 更新用户概况数据
 * param {*gin.Context} c
 * return {*}
********************************************************************************/
func UpdateUserOverview(c *gin.Context) (int, interface{}) {
	var userOverview = &UserOverview{}
	if err := c.ShouldBindJSON(userOverview); err != nil {
		return common.JsonError, "json format error"
	}
	if userOverview.UserId <= 0 {
		return common.ParamError, "user id required"
	}
	user := mysql.NewUser()
	if user.QueryByID(userOverview.UserId) {
		user.BornDate = userOverview.BornDate
		user.Gender = userOverview.Gender
		user.Grade = userOverview.Grade
		if user.Update() {
			return common.Success, "user overview update success"
		} else {
			return common.DBError, "update failed"
		}
	}
	return common.NoExist, "user not exist"
}
