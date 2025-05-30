/******************************************************************************
 * Author: liguoqiang
 * Date: 2023-12-26 19:38:45
 * LastEditors: liguoqiang
 * LastEditTime: 2024-11-30 20:03:00
 * Description:
********************************************************************************/
package api

import (
	"hjyserver/mdb"

	"github.com/gin-gonic/gin"
)

func InitX1sActions() (map[string]gin.HandlerFunc, map[string]gin.HandlerFunc) {
	postAction := make(map[string]gin.HandlerFunc)
	getAction := make(map[string]gin.HandlerFunc)
	postAction["/x1s/askX1sSyncVersion"] = askX1sSyncVersion
	postAction["/x1s/insertX1sWhiteList"] = insertX1sWhiteList

	getAction["/x1s/queryX1sWhiteList"] = queryX1sWhiteList

	return postAction, getAction
}

// askX1sSyncVersion godoc
//
//	@Summary	askX1sSyncVersion
//	@Schemes
//	@Description	ask X1s device to sync version
//	@Tags			X1s
//	@Param			token	query	string		false	"token"
//	@Param			mac	query	string		true	"device mac address"
//	@Produce		json
//	@Success		200	{none} {none}
//	@Router			/x1s/askX1sSyncVersion [post]
func askX1sSyncVersion(c *gin.Context) {
	apiCommonFunc(c, mdb.AskX1sSyncVersion)
}

// insertX1sWhiteList godoc
//
//	@Summary	insertX1sWhiteList
//	@Schemes
//	@Description	insert X1s White list
//	@Tags			X1s
//	@Param			token	query	string		false	"token"
//	@Param			in	body	mdb.X1sWhiteListReq		true	"mac"
//	@Produce		json
//	@Success		200	{string} {"ok"}
//	@Router			/x1s/insertX1sWhiteList [post]
func insertX1sWhiteList(c *gin.Context) {
	apiCommonFunc(c, mdb.InsertX1sWhiteList)
}

// queryX1sWhiteList godoc
//
//	@Summary	queryX1sWhiteList
//	@Schemes
//	@Description	query X1s White list
//	@Tags			X1s
//	@Produce		json
//	@Success		200	{none} {none}
//	@Router			/x1s/queryX1sWhiteList [get]
func queryX1sWhiteList(c *gin.Context) {
	apiCommonFunc(c, mdb.QueryX1sWhiteList)
}
