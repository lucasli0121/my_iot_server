/******************************************************************************
 * Author: liguoqiang
 * Date: 2023-09-06 17:50:12
 * LastEditors: liguoqiang
 * LastEditTime: 2024-12-09 20:46:38
 * Description:
********************************************************************************/
package api

import (
	"hjyserver/exception"
	"hjyserver/mdb"
	"net/http"

	"github.com/gin-gonic/gin"
)

func InitUserActions() (map[string]gin.HandlerFunc, map[string]gin.HandlerFunc) {
	postAction := make(map[string]gin.HandlerFunc)
	getAction := make(map[string]gin.HandlerFunc)

	getAction["/user/queryById"] = queryUserById
	getAction["/user/queryUserByPhone"] = queryUserByPhone
	getAction["/user/queryUserByEmail"] = queryUserByEmail
	getAction["/user/queryUserGroup"] = queryUserGroup
	getAction["/user/queryFriendsByUser"] = queryFriendsByUser
	getAction["/user/queryLampUsersByRoom"] = queryLampUsersByRoom
	getAction["/user/queryLampUserInFriend"] = queryLampUserInFriend

	getAction["/user/queryStudyRoomUser"] = queryStudyRoomUser
	getAction["/user/queryUserStudyData"] = queryUserStudyData
	getAction["/user/queryUserStudyTimeByDay"] = queryUserStudyTimeByDay

	getAction["/user/queryUserOverview"] = queryUserOverview

	getAction["/user/verifyUserToken"] = verifyUserToken

	// post user tag action
	postAction["/user/userLogin"] = userLogin
	postAction["/user/userRegister"] = userRegister
	postAction["/user/loginout"] = loginOut
	postAction["/user/deleteUser"] = deleteUser
	postAction["/user/online"] = userOnline
	postAction["/user/offline"] = userOffline
	postAction["/user/updateNickName"] = updateNickName
	postAction["/user/updateGender"] = updateGender
	postAction["/user/updateHeadPic"] = updateHeadPic
	postAction["/user/update"] = updateUser
	postAction["/user/modifyPhone"] = modifyPhone
	postAction["/user/modifyEmergentPhone"] = modifyEmergentPhone
	postAction["/user/modifyEmail"] = modifyEmail
	postAction["/user/modifyPasswd"] = modifyPasswd
	postAction["/user/insertGroup"] = insertUserGroup
	postAction["/user/deleteGroup"] = deleteUserGroup
	postAction["/user/insertUserFriend"] = insertUserFriend
	postAction["/user/removeUserFriend"] = removeUserFriend
	postAction["/user/modifyUserFriend"] = modifyUserFriend
	postAction["/user/removeUserDevice"] = removeUserDevice
	postAction["/user/addUserToStudyRoom"] = addUserToStudyRoom
	postAction["/user/removeUserFromStudyRoom"] = removeUserFromStudyRoom
	postAction["/user/enterStudyRoom"] = enterStudyRoom
	postAction["/user/leaveStudyRoom"] = leaveStudyRoom

	postAction["/user/updateUserOverview"] = updateUserOverview
	return postAction, getAction
}

// verifyUserToken godoc
//
//	@Summary	verifyUserToken
//	@Schemes
//	@Description	验证用户token是否过期
//	@Tags			user
//	@Produce		json
//
//	@Param			token	query	string	 true	"user token"
//
//	@Success		200			{object}	mdb.VerifyTokenResp
//	@Router			/user/verifyUserToken [get]
func verifyUserToken(c *gin.Context) {
	apiCommonFunc(c, mdb.VerifyUserToken)
}

/******************************************************************************
 * function: userLogin
 * description: login in or register a new user
 * return {*}
********************************************************************************/

// userLogin godoc
//
//	@Summary	userLogin
//	@Schemes
//	@Description	login in
//	@Tags			user
//	@Produce		json
//
//	@Param			in		body	mdb.LoginReq true	"user info"
//
//	@Success		200			{object}	mysql.User
//	@Router			/user/userLogin [post]
func userLogin(c *gin.Context) {
	exception.TryEx{
		Try: func() {
			status, result := mdb.UserLogin(c)
			respJSON(c, status, result)
		},
		Catch: func(e exception.Exception) {
			respJSON(c, e.Code, e.Msg)
		},
	}.Run()
}

// userRegister godoc
//
//	@Summary	userRegister
//	@Schemes
//	@Description	user register
//	@Tags			user
//	@Produce		json
//
//	@Param			in		body	mysql.User true	"user info"
//
//	@Success		200			{object}	mysql.User
//	@Router			/user/userRegister [post]
func userRegister(c *gin.Context) {
	exception.TryEx{
		Try: func() {
			status, result := mdb.UserRegister(c)
			respJSON(c, status, result)
		},
		Catch: func(e exception.Exception) {
			respJSON(c, http.StatusBadRequest, e.Msg)
		},
	}.Run()
}

/******************************************************************************
 * function: loginOut
 * description: login out from server
 * return {*}
********************************************************************************/

// loginout godoc
//
//	@Summary	loginout
//	@Schemes
//	@Description	login out from server
//	@Tags			user
//	@Produce		json
//	@Param			token	query	string		false	"token"
//	@Param			in	body	mdb.LoginoutReq	 true	"user id"
//
//	@Success		200			{object}	mysql.User
//	@Router			/user/loginout [post]
func loginOut(c *gin.Context) {
	exception.TryEx{
		Try: func() {
			status, result := mdb.LoginOut(c)
			respJSON(c, status, result)
		},
		Catch: func(e exception.Exception) {
			respJSON(c, e.Code, e.Msg)
		},
	}.Run()
}

/******************************************************************************
 * function: deleteUser
 * description:
 * param {*gin.Context} c
 * return {*}
********************************************************************************/

// deleteUser godoc
//
//	@Summary	deleteUser
//	@Schemes
//	@Description	delete user from server
//	@Tags			user
//	@Produce		json
//	@Param			token	query	string		false	"token"
//	@Param			in	body	mdb.DeleteUserReq	 true	"user id"
//
//	@Success		200			{string}	{"delete user ok""}
//	@Router			/user/deleteUser [post]
func deleteUser(c *gin.Context) {
	exception.TryEx{
		Try: func() {
			status, result := mdb.DeleteUser(c)
			respJSON(c, status, result)
		},
		Catch: func(e exception.Exception) {
			respJSON(c, e.Code, e.Msg)
		},
	}.Run()
}

/*
WEB 接口，For query user infomaion
*/

// queryById godoc
//
//	@Summary	queryById
//	@Schemes
//	@Description	query user by id
//	@Tags			user
//	@Produce		json
//
//	@Param			id	query	int	 true	"user id"
//
//	@Success		200			{object}	mysql.User
//	@Router			/user/queryById [get]
func queryUserById(c *gin.Context) {
	exception.TryEx{
		Try: func() {
			status, result := mdb.QueryUserById(c)
			respJSON(c, status, result)
		},
		Catch: func(e exception.Exception) {
			respJSON(c, e.Code, e.Msg)
		},
	}.Run()
}

/**
 * @description: query user by phone
 * @return {*}
 */

// queryUserByPhone godoc
//
//	@Summary	queryUserByPhone
//	@Schemes
//	@Description	query user by phone
//	@Tags			user
//	@Produce		json
//
//	@Param			phone	query	string	 true	"user phone"
//
//	@Success		200			{object}	mysql.User
//	@Router			/user/queryUserByPhone [get]
func queryUserByPhone(c *gin.Context) {
	exception.TryEx{
		Try: func() {
			status, result := mdb.QueryUserByPhone(c)
			respJSON(c, status, result)
		},
		Catch: func(e exception.Exception) {
			respJSON(c, e.Code, e.Msg)

		},
	}.Run()

}

// queryUserByEmail godoc
//
//	@Summary	queryUserByEmail
//	@Schemes
//	@Description	query user by email
//	@Tags			user
//	@Produce		json
//
//	@Param			email	query	string	 true	"user email"
//
//	@Success		200			{object}	mysql.User
//	@Router			/user/queryUserByEmail [get]
func queryUserByEmail(c *gin.Context) {
	apiCommonFunc(c, mdb.QueryUserByEmail)
}

/*
userOnline...
For update user online status
*/

// online godoc
//
//	@Summary	online
//	@Schemes
//	@Description	set user online
//	@Tags			user
//	@Produce		json
//	@Param			token	query	string		false	"token"
//	@Param			id	body	int	 true	"user id"
//
//	@Success		200			{object}	mysql.User
//	@Router			/user/online [post]
func userOnline(c *gin.Context) {
	exception.TryEx{
		Try: func() {
			status, result := mdb.UserOnline(c)
			respJSON(c, status, result)
		},
		Catch: func(e exception.Exception) {
			respJSON(c, e.Code, e.Msg)
		},
	}.Run()
}

/*
userOffline...
For update user online status
*/

// offline godoc
//
//	@Summary	offline
//	@Schemes
//	@Description	set user offline
//	@Tags			user
//	@Produce		json
//	@Param			token	query	string		false	"token"
//	@Param			id	body	int	 true	"user id"
//
//	@Success		200			{object}	mysql.User
//	@Router			/user/offline [post]
func userOffline(c *gin.Context) {
	exception.TryEx{
		Try: func() {
			status, result := mdb.UserOffline(c)
			respJSON(c, status, result)
		},
		Catch: func(e exception.Exception) {
			respJSON(c, e.Code, e.Msg)
		},
	}.Run()
}

/*
updateUser... 根据ID, code,name update user information
*/

// swagger:model UpdateUserReq
type UpdateUserReq struct {
	ID        int64  `json:"id" mysql:"id" binding:"omitempty"`
	Account   string `json:"account" mysql:"account"`
	NickName  string `json:"nick_name" mysql:"nick_name"`
	LoginType int    `json:"login_type" mysql:"login_type"` // 0:phone 1:email
	Phone     string `json:"phone" mysql:"phone"`
	Email     string `json:"email" mysql:"email"`
	Face      string `json:"face" mysql:"face"`
	Address   string `json:"address" mysql:"address"`
}

// updateNickName godoc
//
//	@Summary	updateNickName
//	@Schemes
//	@Description	update user nickname
//	@Tags			user
//	@Produce		json
//	@Param			token	query	string		false	"token"
//	@Param			in	body	mdb.NickNameReq	 true	"user information"
//
//	@Success		200			{object}	mysql.User
//	@Router			/user/updateNickName [post]
func updateNickName(c *gin.Context) {
	apiCommonFunc(c, mdb.UpdateNickName)
}

// updateGender godoc
//
//	@Summary	updateGender
//	@Schemes
//	@Description	update user gender
//	@Tags			user
//	@Produce		json
//	@Param			token	query	string		false	"token"
//	@Param			in	body	mdb.UserGenderReq	 true	"user gender"
//
//	@Success		200			{object}	mysql.User
//	@Router			/user/updateGender [post]
func updateGender(c *gin.Context) {
	apiCommonFunc(c, mdb.UpdateGender)
}

// updateHeadPic godoc
//
//	@Summary	updateHeadPic
//	@Schemes
//	@Description	update user face picture
//	@Tags			user
//	@Produce		json
//	@Param			token	query	string		false	"token"
//	@Param			in	body	mdb.UserFacePicReq	 true	"user picture information"
//
//	@Success		200			{object}	mysql.User
//	@Router			/user/updateHeadPic [post]
func updateHeadPic(c *gin.Context) {
	apiCommonFunc(c, mdb.UpdateHeadPic)
}

// update godoc
//
//	@Summary	update
//	@Schemes
//	@Description	update user
//	@Tags			user
//	@Produce		json
//	@Param			token	query	string		false	"token"
//	@Param			in	body	mysql.User	 true	"user information"
//
//	@Success		200			{object}	mysql.User
//	@Router			/user/update [post]
func updateUser(c *gin.Context) {
	apiCommonFunc(c, mdb.UpdateUser)
}

/**
 * @description: modify user's phone
 * @return {*}
 */

// modifyPhone godoc
//
//	@Summary	modifyPhone
//	@Schemes
//	@Description	modify user phone
//	@Tags			user
//	@Produce		json
//	@Param			token	query	string		false	"token"
//	@Param			in	body	mdb.NewPhone	 true	"user information"
//
//	@Success		200			{object}	mysql.User
//	@Router			/user/modifyPhone [post]
func modifyPhone(c *gin.Context) {
	apiCommonFunc(c, mdb.ModifyPhone)
}

// modifyEmergentPhone godoc
//
//	@Summary	modifyEmergentPhone
//	@Schemes
//	@Description	modify user emergent phone
//	@Tags			user
//	@Produce		json
//	@Param			token	query	string		false	"token"
//	@Param			in	body	mdb.NewPhone	 true	"user information"
//
//	@Success		200			{object}	mysql.User
//	@Router			/user/modifyEmergentPhone [post]
func modifyEmergentPhone(c *gin.Context) {
	apiCommonFunc(c, mdb.ModifyEmergentPhone)
}

// modifyEmail godoc
//
//	@Summary	modifyEmail
//	@Schemes
//	@Description	modify user email
//	@Tags			user
//	@Produce		json
//	@Param			token	query	string		false	"token"
//	@Param			in	body	mdb.NewEmail	 true	"user email"
//
//	@Success		200			{object}	mysql.User
//	@Router			/user/modifyEmail [post]
func modifyEmail(c *gin.Context) {
	apiCommonFunc(c, mdb.ModifyEmail)
}

// modifyPasswd godoc
//
//	@Summary	modifyPasswd
//	@Schemes
//	@Description	modify user passwd
//	@Tags			user
//	@Produce		json
//	@Param			token	query	string		false	"token"
//	@Param			in	body	mdb.NewPasswd	 true	"user new passwd"
//
//	@Success		200			{object}	mysql.User
//	@Router			/user/modifyPasswd [post]
func modifyPasswd(c *gin.Context) {
	apiCommonFunc(c, mdb.ModifyPasswd)
}

/**
 * @description: query use group
 * @return {*}
 */
func queryUserGroup(c *gin.Context) {
	exception.TryEx{
		Try: func() {
			status, result := mdb.QueryUserGroup(c)
			respJSON(c, status, result)
		},
		Catch: func(e exception.Exception) {
			respJSON(c, e.Code, e.Msg)
		},
	}.Run()

}

/**
 * @description: delete user group
 * @return {*}
 */
func deleteUserGroup(c *gin.Context) {
	exception.TryEx{
		Try: func() {
			status, result := mdb.DeleteUserGroup(c)
			respJSON(c, status, result)
		},
		Catch: func(e exception.Exception) {
			respJSON(c, e.Code, e.Msg)
		},
	}.Run()

}

/**
 * @description: insert user group
 * @return {*}
 */
func insertUserGroup(c *gin.Context) {
	exception.TryEx{
		Try: func() {
			status, result := mdb.InsertUserGroup(c)
			respJSON(c, status, result)
		},
		Catch: func(e exception.Exception) {
			respJSON(c, e.Code, e.Msg)
		},
	}.Run()

}

/**
 * @description: query user friends
 * @return {*}
 */

// queryFriendsByUser godoc
//
//	@Summary	queryFriendsByUser
//	@Schemes
//	@Description	query user's friend
//	@Tags			user
//	@Produce		json
//
//	@Param			user_id	query	int	 true	"user id"
//
//	@Success		200			{object}	mysql.UserFriend
//	@Router			/user/queryFriendsByUser [get]
func queryFriendsByUser(c *gin.Context) {
	exception.TryEx{
		Try: func() {
			status, result := mdb.QueryFriendsByUser(c)
			respJSON(c, status, result)
		},
		Catch: func(e exception.Exception) {
			respJSON(c, e.Code, e.Msg)

		},
	}.Run()

}

/**
 * @description: insert user friend
 * @return {*}
 */

type InsertUserFriendReq struct {
	// required: true
	// example: 1
	UserId int64 `json:"user_id"`
	// required: true
	// example: 1
	Phone string `json:"phone"`
}

// insertUserFriend godoc
//
//	@Summary	insertUserFriend
//	@Schemes
//	@Description	insert user's friend
//	@Tags			user
//	@Produce		json
//	@Param			token	query	string		false	"token"
//	@Param			in	body	InsertUserFriendReq	 true	"friend info"
//
//	@Success		200			{object}	mysql.UserFriend
//	@Router			/user/insertUserFriend [post]
func insertUserFriend(c *gin.Context) {
	exception.TryEx{
		Try: func() {
			status, result := mdb.InsertUserFriend(c)
			respJSON(c, status, result)
		},
		Catch: func(e exception.Exception) {
			respJSON(c, e.Code, e.Msg)

		},
	}.Run()

}

// modifyUserFriend godoc
//
//	@Summary	modifyUserFriend
//	@Schemes
//	@Description	modify user's friend
//	@Tags			user
//	@Produce		json
//	@Param			token	query	string		false	"token"
//	@Param			in	body	mdb.ModifyFriendReq	 true	"modify friend request"
//
//	@Success		200			{string}	{"modify user friend ok""}
//	@Router			/user/modifyUserFriend [post]
func modifyUserFriend(c *gin.Context) {
	exception.TryEx{
		Try: func() {
			status, result := mdb.ModifyUserFriend(c)
			respJSON(c, status, result)
		},
		Catch: func(e exception.Exception) {
			respJSON(c, e.Code, e.Msg)

		},
	}.Run()

}

// removeUserFriend godoc
//
//	@Summary	removeUserFriend
//	@Schemes
//	@Description	remove user's friend
//	@Tags			user
//	@Produce		json
//	@Param			token	query	string		false	"token"
//	@Param			in	body	mdb.RemoveFriendReq	 true	"remove friend request"
//
//	@Success		200			{string}	{"remove user friend ok""}
//	@Router			/user/removeUserFriend [post]
func removeUserFriend(c *gin.Context) {
	exception.TryEx{
		Try: func() {
			status, result := mdb.RemoveUserFriend(c)
			respJSON(c, status, result)
		},
		Catch: func(e exception.Exception) {
			respJSON(c, http.StatusBadRequest, e.Msg)
		},
	}.Run()

}

// swagger:model RemoveUserDeviceReq
type RemoveUserDeviceReq struct {
	// required: true
	// example: 1
	UserId int64 `json:"user_id"`
	// required: true
	// example: 1
	DeviceId int64 `json:"device_id"`
}

// removeUserDevice godoc
//
//	@Summary	removeUserDevice
//	@Schemes
//	@Description	remove user device
//	@Tags			user
//	@Produce		json
//	@Param			token	query	string		false	"token"
//	@Param			in		body	RemoveUserDeviceReq		true	"user device"
//
//	@Success		200			{string}	{"remove user device ok"}
//	@Router			/user/removeUserDevice [post]
func removeUserDevice(c *gin.Context) {
	exception.TryEx{
		Try: func() {
			status, result := mdb.RemoveUserDevice(c)
			respJSON(c, status, result)
		},
		Catch: func(e exception.Exception) {
			respJSON(c, e.Code, e.Msg)

		},
	}.Run()
}

// addUserToStudyRoom godoc
//
//	@Summary	addUserToStudyRoom
//	@Schemes
//	@Description	add user id to study room
//	@Tags			room
//	@Produce		json
//
//	@Param			in	body	mdb.UserToStudyRoomReq		true	"add user to study room"
//
//	@Success		200			{object}	mysql.StudyRoomUser
//	@Router			/user/addUserToStudyRoom [post]
func addUserToStudyRoom(c *gin.Context) {
	exception.TryEx{
		Try: func() {
			status, result := mdb.AddUserToStudyRoom(c)
			respJSON(c, status, result)
		},
		Catch: func(e exception.Exception) {
			respJSON(c, e.Code, e.Msg)

		},
	}.Run()
}

// removeUserFromStudyRoom godoc
//
//	@Summary	removeUserFromStudyRoom
//	@Schemes
//	@Description	remove user id from study room
//	@Tags			room
//	@Produce		json
//
//	@Param			in	body	mdb.UserToStudyRoomReq		true	"remove user from study room"
//
//	@Success		200			{string}	{string}	removeUserFromStudyRoom
//	@Router			/user/removeUserFromStudyRoom [post]
func removeUserFromStudyRoom(c *gin.Context) {
	exception.TryEx{
		Try: func() {
			status, result := mdb.RemoveUserFromStudyRoom(c)
			respJSON(c, status, result)
		},
		Catch: func(e exception.Exception) {
			respJSON(c, e.Code, e.Msg)

		},
	}.Run()
}

// queryStudyRoomUser godoc
//
//		@Summary	queryStudyRoomUser
//		@Schemes
//		@Description	query user list from study room
//		@Tags			room
//		@Produce		json
//
//		@Param			create_id	query	int		true	"create user id"
//		@Param			room_id		query	int		true	"room id"
//	 @Param flag query int true "flag 0:all 1:study user"
//
//		@Success		200			{object}	mysql.StudyRoomUserDetail
//		@Router			/user/queryStudyRoomUser [get]
func queryStudyRoomUser(c *gin.Context) {
	apiCommonFunc(c, mdb.QueryStudyRoomUser)
}

// queryUserStudyData godoc
//
//	@Summary	queryUserStudyData
//	@Schemes
//	@Description	查询用户自习的数据,包括自习时长，自习天数，平均时长，最长时间, 以及根据日期进行分组统计每天的自习时长
//	@Tags			room
//	@Produce		json
//
//	@Param			user_id	query	int		true	"user id"
//	@Param			room_id		query	int		true	"room id"
//
// @Param			start_time	query	string	true	"start time"
// @Param			end_time	query	string	true	"end time"
//
//	@Success		200			{object}	mdb.UserStudyData
//	@Router			/user/queryUserStudyData [get]
func queryUserStudyData(c *gin.Context) {
	exception.TryEx{
		Try: func() {
			status, result := mdb.QueryUserStudyData(c)
			respJSON(c, status, result)
		},
		Catch: func(e exception.Exception) {
			respJSON(c, e.Code, e.Msg)

		},
	}.Run()
}

// queryUserStudyTimeByDay godoc
//
//	@Summary	queryUserStudyTimeByDay
//	@Schemes
//	@Description	查询用户学习时间段，根据指定日期查询
//	@Tags			room
//	@Produce		json
//
//	@Param			user_id	query	int		true	"user id"
//	@Param			room_id		query	int		true	"room id"
//
// @Param			query_date	query	string	true	"query date"
//
//	@Success		200			{object}	mysql.UserStudyTime
//	@Router			/user/queryUserStudyTimeByDay [get]
func queryUserStudyTimeByDay(c *gin.Context) {
	exception.TryEx{
		Try: func() {
			status, result := mdb.QueryUserStudyTimeByDay(c)
			respJSON(c, status, result)
		},
		Catch: func(e exception.Exception) {
			respJSON(c, e.Code, e.Msg)

		},
	}.Run()
}

// queryLampUsersByRoom godoc
//
//	@Summary	queryLampUsersByRoom
//	@Schemes
//	@Description	查询拥有灯的用户，条件自习室ID,id=0 时拥有灯控的所有用户
//	@Tags			room
//	@Produce		json
//
//	@Param			room_id		query	int		true	"room id"
//
// @Success		200			{object}	mysql.LampUserWithStudyRoom
// @Router			/user/queryLampUsersByRoom [get]
func queryLampUsersByRoom(c *gin.Context) {
	exception.TryEx{
		Try: func() {
			status, result := mdb.QueryLampUsersByRoom(c)
			respJSON(c, status, result)
		},
		Catch: func(e exception.Exception) {
			respJSON(c, e.Code, e.Msg)
		},
	}.Run()

}

// queryLampUserInFriend godoc
//
//	@Summary	queryLampUserInFriend
//	@Schemes
//	@Description	查询好友中绑定灯的好友
//	@Tags			room
//	@Produce		json
//
//	@Param			user_id		query	int		true	"user id"
//
// @Success		200			{object}	mysql.LampUserWithStudyRoom
// @Router			/user/queryLampUserInFriend [get]
func queryLampUserInFriend(c *gin.Context) {
	apiCommonFunc(c, mdb.QueryLampUserInFriend)
}

// enterStudyRoom godoc
//
//	@Summary	enterStudyRoom
//	@Schemes
//	@Description	用户进入已邀请的自习室
//	@Tags			room
//	@Produce		json
//
//	@Param			in		body	mdb.UserEnterStudyReq		true	"study info"
//
// @Success		200			{object}	mysql.UserStudyRecord
// @Router			/user/enterStudyRoom [post]
func enterStudyRoom(c *gin.Context) {
	exception.TryEx{
		Try: func() {
			status, result := mdb.EnterStudyRoom(c)
			respJSON(c, status, result)
		},
		Catch: func(e exception.Exception) {
			respJSON(c, e.Code, e.Msg)
		},
	}.Run()
}

// leaveStudyRoom godoc
//
//	@Summary	leaveStudyRoom
//	@Schemes
//	@Description	用户离开已邀请的自习室
//	@Tags			room
//	@Produce		json
//
//	@Param			in		body	mdb.UserEnterStudyReq		true	"study info"
//
// @Success		200			{string}	{"leave room success"}
// @Router			/user/leaveStudyRoom [post]
func leaveStudyRoom(c *gin.Context) {
	exception.TryEx{
		Try: func() {
			status, result := mdb.LeaveStudyRoom(c)
			respJSON(c, status, result)
		},
		Catch: func(e exception.Exception) {
			respJSON(c, e.Code, e.Msg)
		},
	}.Run()
}

// queryUserOverview godoc
//
//	@Summary	queryUserOverview
//	@Schemes
//	@Description	查询用户概况数据
//	@Tags			user
//	@Produce		json
//
//	@Param			user_id		query	int		true	"user id"
//
// @Success		200			{object}	mdb.UserOverview
// @Router			/user/queryUserOverview [get]
func queryUserOverview(c *gin.Context) {
	apiCommonFunc(c, mdb.QueryUserOverview)
}

// updateUserOverview godoc
//
//	@Summary	updateUserOverview
//	@Schemes
//	@Description	更新用户概况数据
//	@Tags			user
//	@Produce		json
//	@Param			token	query	string		false	"token"
//	@Param			in	body	mdb.UserOverview	true	"用户概况数据"
//
// @Success		200			{string}	{"update success"}
// @Router			/user/updateUserOverview [post]
func updateUserOverview(c *gin.Context) {
	apiCommonFunc(c, mdb.UpdateUserOverview)
}
