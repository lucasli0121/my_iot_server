/******************************************************************************
 * Author: liguoqiang
 * Date: 2024-08-27 18:47:24
 * LastEditors: liguoqiang
 * LastEditTime: 2024-09-11 23:22:28
 * Description:
********************************************************************************/
/*********************************************************************
* File Name: user_api.go
* Author: liguoqiang
* mail:
* Created Time: 2021-03-07 09:31:25
* LastEditors: liguoqiang
**********************************************************************/

package api

import (
	"hjyserver/exception"
	"hjyserver/mdb"

	"github.com/gin-gonic/gin"
)

func InitSettingActions() (map[string]gin.HandlerFunc, map[string]gin.HandlerFunc) {
	postAction := make(map[string]gin.HandlerFunc)
	getAction := make(map[string]gin.HandlerFunc)
	postAction["/setting/insertBanner"] = insertBanner
	getAction["/setting/queryBanner"] = queryBanner

	return postAction, getAction
}

/*
WEB 接口，For query setting infomaion
*/
func querySetting(c *gin.Context) {

	exception.TryEx{
		Try: func() {

		},
		Catch: func(e exception.Exception) {
			respJSON(c, e.Code, e.Msg)
		},
	}.Run()
}

/*
insertSetting...
For insert setting information
*/
func insertSetting(c *gin.Context) {
	exception.TryEx{
		Try: func() {
		},
		Catch: func(e exception.Exception) {
			respJSON(c, e.Code, e.Msg)
		},
	}.Run()
}

/*
updateSetting... 根据ID, code,name update setting information
*/
func updateSetting(c *gin.Context) {
	exception.TryEx{
		Try: func() {
		},
		Catch: func(e exception.Exception) {
			respJSON(c, e.Code, e.Msg)
		},
	}.Run()
}

// insertBanner godoc
//
//	@Summary	insertBanner
//	@Schemes
//	@Description	insert banner picture into database
//	@Tags			setting
//	@Param			token	query	string		false	"token"
//	@Param			in	body	mysql.BannerSetting		true	"banner settting"
//	@Produce		json
//	@Success		200	{none} {none}
//	@Router			/setting/insertBanner [post]
func insertBanner(c *gin.Context) {
	apiCommonFunc(c, mdb.InsertBanner)
}

// queryBanner godoc
//
//	@Summary	queryBanner
//	@Schemes
//	@Description	查询banner图片
//	@Tags			setting
//
//	@Produce		json
//	@Success		200	{object}	mysql.BannerSetting
//	@Router			/setting/queryBanner [get]
func queryBanner(c *gin.Context) {
	apiCommonFunc(c, mdb.QueryBanner)
}
