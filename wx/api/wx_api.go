/******************************************************************************
 * Author: liguoqiang
 * Date: 2024-07-12 10:33:35
 * LastEditors: liguoqiang
 * LastEditTime: 2024-12-16 19:58:57
 * Description:
********************************************************************************/

package wxapi

import (
	"hjyserver/exception"
	"hjyserver/mdb/common"
	mdbwx "hjyserver/wx/mdb"
	"net/http"

	"github.com/gin-gonic/gin"
)

/******************************************************************************
 * function:
 * description:
 * return {*}
********************************************************************************/
func InitWxActions() (map[string]gin.HandlerFunc, map[string]gin.HandlerFunc) {
	postAction := make(map[string]gin.HandlerFunc)
	getAction := make(map[string]gin.HandlerFunc)
	postAction["/wx/wxMiniProgramLogin"] = wxMiniProgramLogin
	postAction["/wx/wxPublicSubmit"] = wxPublicSubmit
	postAction["/wx/askMiniUrl"] = askMiniUrl

	getAction["/wx/wxPublicSubmit"] = wxPublicSubmit
	getAction["/wx/queryMiniAccessToken"] = queryMiniAccessToken

	return postAction, getAction
}

/******************************************************************************
 * function: wxMiniProgramLogin
 * description: winxin login api
 * param {*gin.Context} c
 * return {*}
********************************************************************************/
// wxMiniProgramLogin godoc
//
//	@Summary	wxMiniProgramLogin
//	@Schemes
//	@Description	winxin mini program login api
//	@Tags			wx
//
//	@Param			in	body	mdbwx.WxMiniLoginReq	true	"wx mini program login request"
//	@Produce		json
//	@Success		200	{object}	mysql.User
//	@Router			/wx/wxMiniProgramLogin [post]
func wxMiniProgramLogin(c *gin.Context) {
	apiCommonFunc(c, mdbwx.WxMiniProgramLogin)
}

/******************************************************************************
 * function: wxPublicSubmit
 * description:
 * param {*gin.Context} c
 * return {*}
********************************************************************************/
// wxPublicSubmit godoc
//
//	@Summary	wxPublicSubmit
//	@Schemes
//	@Description	公众号推送的服务器地址
//	@Tags			wx
//
//	@Param			in	body	string	true	"xml字符串"
//	@Produce		json
//	@Success		200	{none} {none}
//	@Router			/wx/wxPublicSubmit [post]
func wxPublicSubmit(c *gin.Context) {
	status, result := mdbwx.WxPublicSubmit(c)
	c.String(status, result.(string))
}

/******************************************************************************
 * function: queryMiniAccessToken
 * description:
 * param {*gin.Context} c
 * return {*}
********************************************************************************/
// queryMiniAccessToken godoc
//
//	@Summary	queryMiniAccessToken
//	@Schemes
//	@Description	获取小程序的access_token
//	@Tags			wx
//
//	@Produce		json
//	@Success		200	{none} {none}
//	@Router			/wx/queryMiniAccessToken [get]
func queryMiniAccessToken(c *gin.Context) {
	apiCommonFunc(c, mdbwx.QueryMiniAccessToken)
}

/******************************************************************************
 * function: askMiniUrl
 * description:
 * param {*gin.Context} c
 * return {*}
********************************************************************************/
// askMiniUrl godoc
//
//	@Summary	askMiniUrl
//	@Schemes
//	@Description	获取url
//	@Tags			wx
//	@Param			in	body	mdbwx.WxMiniUrlReq	true	"请求参数"
//	@Produce		json
//	@Success		200	{string} {url}
//	@Router			/wx/askMiniUrl [post]
func askMiniUrl(c *gin.Context) {
	apiCommonFunc(c, mdbwx.AskMiniUrl)
}

func respPublic(c *gin.Context, status int, msg string) {
	var result string
	switch status {
	case http.StatusOK:
		result = "Success"
	default:
		result = ""
	}
	c.String(status, result)
}

// 返回response http 回应函数，返回为json，格式为
// 错误信息：{ code: 201, message: "" }，正常信息： {code: 200, data: {} }
func respJSON(c *gin.Context, status int, msg interface{}) {
	if status != http.StatusOK {
		c.JSON(status, gin.H{"code": status, "message": msg})
	} else {
		c.JSON(status, gin.H{"code": status, "data": msg})
	}
}

// 返回带页号的response 回应，格式为{code: 200, pageNo: 1, pageSize 20, data: {} }
func respJSONWithPage(c *gin.Context, status int, page *common.PageDao, msg interface{}) {
	if status != http.StatusOK {
		c.JSON(status, gin.H{"code": status, "message": msg})
	} else {
		c.JSON(status, gin.H{"code": status, "pageNo": page.PageNo, "pageSize": page.PageSize, "totalPage": page.TotalPages, "data": msg})
	}
}

/******************************************************************************
 * function: apiCommonFunc
 * description: define a common function for api
 * param {*gin.Context} c
 * return {*}
********************************************************************************/
func apiCommonFunc(c *gin.Context, mdbFunc func(c *gin.Context) (int, interface{})) {
	exception.TryEx{
		Try: func() {
			status, result := mdbFunc(c)
			respJSON(c, status, result)
		},
		Catch: func(e exception.Exception) {
			respJSON(c, http.StatusBadRequest, e.Msg)
		},
	}.Run()
}
