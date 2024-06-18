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

	"github.com/gin-gonic/gin"
)

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
