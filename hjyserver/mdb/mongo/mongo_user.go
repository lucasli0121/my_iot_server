/*
 * @Author: liguoqiang
 * @Date: 2022-05-30 23:25:52
 * @LastEditors: liguoqiang
 * @LastEditTime: 2023-02-12 11:16:49
 * @Description:
 */
package mongo

import (
	"hjyserver/cfg"
	"hjyserver/exception"
	mylog "hjyserver/log"
	"hjyserver/mdb/common"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

/*
* User... user table
 */
type User struct {
	ID         primitive.ObjectID `bson:"_id" json:"id" form:"id" binding:"omitempty"`
	Account    string             `json:"account" bson:"account" binding:"required"`
	Password   string             `json:"password" bson:"password" binding:"required"`
	NickName   string             `json:"nick_name" bson:"nick_name" binding:"required"`
	Phone      string             `json:"phone" bson:"phone" binding:"omitempty"`
	Face       string             `json:"face" bson:"face" binding:"omitempty"`
	Address    string             `json:"address" bson:"address" binding:"omitempty"`
	RoomNum    string             `json:"room_num" bson:"room_num"`
	CreateTime string             `json:"create_time" bson:"create_time"`
}

/*
* User...
 */
func NewUser() *User {
	return &User{
		ID:         primitive.NilObjectID,
		Account:    "",
		Password:   "",
		NickName:   "",
		Phone:      "",
		Face:       "",
		Address:    "",
		RoomNum:    "",
		CreateTime: time.Now().Format(cfg.TmFmtStr),
	}
}

/*
* QueryUser...
 */
func QueryUser(results *[]User) bool {
	res := QueryDao(common.UserTbl, bson.M{}, func(cur *mongo.Cursor) {
		var v *User = NewUser()
		err := cur.Decode(v)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*results = append(*results, *v)
		}
	})
	return res
}

/*
 */
func QueryUserByCond(filter bson.M, results *[]User) bool {
	res := QueryDao(common.UserTbl, filter, func(cur *mongo.Cursor) {
		var v *User = NewUser()
		err := cur.Decode(v)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*results = append(*results, *v)
		}
	})
	return res
}

/*
Decode 解析从gin获取的数据
*/
func (me *User) Decode(c *gin.Context) {
	if err := c.ShouldBindWith(me, binding.JSON); err != nil {
		exception.Throw(http.StatusAccepted, "Binding error!")
	}
	if me.Account == "" {
		exception.Throw(http.StatusAccepted, "Account is empty!")
	}
}

/*
QueryByID() 查询指数表
*/
func (me *User) QueryByID(id primitive.ObjectID) bool {
	me.SetID(id)
	return QueryDaoByID(common.UserTbl, me.ID, me)
}

/*
Insert 指数表插入
*/
func (me *User) Insert() bool {
	me.ID = primitive.NewObjectID()
	me.Password = common.MakeMD5(me.Password)
	return InsertDao(common.UserTbl, me)
}

/******************************************************************************
 * function: Update
 * description: update user table
 * return {*}
********************************************************************************/
func (me *User) Update() bool {
	u := bson.M{
		"$set": bson.M{
			"account":     me.Account,
			"nick_name":   me.NickName,
			"phone":       me.Phone,
			"face":        me.Face,
			"address":     me.Address,
			"room_num":    me.RoomNum,
			"create_time": me.CreateTime,
		},
	}
	return UpdateDaoByID(common.UserTbl, me.ID, u)
}

/*
Update() 更新指数表
*/
func (me *User) UpdatePassword() bool {
	u := bson.M{
		"$set": bson.M{
			"password": common.MakeMD5(me.Password),
		},
	}
	return UpdateDaoByID(common.UserTbl, me.ID, u)
}

/*
Delete() 删除指数
*/
func (index *User) Delete() bool {
	return DeleteDaoByID(common.UserTbl, index.ID)
}

/*
设置ID
*/
func (me *User) SetID(id primitive.ObjectID) {
	me.ID = id
}
