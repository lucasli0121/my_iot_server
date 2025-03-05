/******************************************************************************
 * Author: liguoqiang
 * Date: 2024-09-11 18:15:55
 * LastEditors: liguoqiang
 * LastEditTime: 2024-09-11 23:21:07
 * Description:
********************************************************************************/
package mysql

import (
	"database/sql"
	"hjyserver/exception"
	mylog "hjyserver/log"
	"hjyserver/mdb/common"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// swagger:model BannerSetting
type BannerSetting struct {
	ID     int64  `json:"id" mysql:"id" binding:"omitempty"`
	Sort   int    `json:"sort" mysql:"sort"`
	ImgUrl string `json:"img_url" mysql:"img_url"`
}

func (BannerSetting) TableName() string {
	return "banner_tbl"
}
func NewBannerSetting() *BannerSetting {
	return &BannerSetting{
		ID:     0,
		Sort:   0,
		ImgUrl: "",
	}
}
func (me *BannerSetting) Insert() bool {
	if !CheckTableExist(me.TableName()) {
		CreateTableWithStruct(me.TableName(), me)
	}
	return InsertDao(me.TableName(), me)
}
func (me *BannerSetting) Update() bool {
	return UpdateDaoByID(me.TableName(), me.ID, me)
}
func (me *BannerSetting) Delete() bool {
	return DeleteDaoByID(me.TableName(), me.ID)
}
func (me *BannerSetting) SetID(id int64) {
	me.ID = id
}
func (me *BannerSetting) QueryByID(id int64) bool {
	return QueryDaoByID(me.TableName(), id, me)
}
func (me *BannerSetting) DecodeFromRows(rows *sql.Rows) error {
	err := rows.Scan(
		&me.ID,
		&me.Sort,
		&me.ImgUrl)
	return err
}
func (me *BannerSetting) DecodeFromRow(row *sql.Row) error {
	err := row.Scan(
		&me.ID,
		&me.Sort,
		&me.ImgUrl)
	return err
}
func (me *BannerSetting) DecodeFromGin(c *gin.Context) {
	if err := c.ShouldBindBodyWith(me, binding.JSON); err != nil {
		exception.Throw(common.JsonError, err.Error())
	}
}

func QueryAllBanner(results *[]BannerSetting) bool {
	QueryDao(NewBannerSetting().TableName(), nil, nil, -1, func(rows *sql.Rows) {
		obj := NewBannerSetting()
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*results = append(*results, *obj)
		}
	})
	return true
}

func QueryBannerBySort(sort int, results *[]BannerSetting) bool {
	QueryDao(NewBannerSetting().TableName(), "sort=?", sort, -1, func(rows *sql.Rows) {
		obj := NewBannerSetting()
		err := obj.DecodeFromRows(rows)
		if err != nil {
			mylog.Log.Errorln(err)
		} else {
			*results = append(*results, *obj)
		}
	})
	return true
}
