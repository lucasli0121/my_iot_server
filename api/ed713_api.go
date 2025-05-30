/******************************************************************************
 * Author: liguoqiang
 * Date: 2024-05-10 14:02:25
 * LastEditors: liguoqiang
 * LastEditTime: 2024-08-07 11:01:36
 * Description:
********************************************************************************/
/*********************************************************************
*
**********************************************************************/

package api

import (
	"net/http"

	"hjyserver/exception"
	"hjyserver/mdb"

	"github.com/gin-gonic/gin"
)

// askEd713RealData godoc
//
//	@Summary	askEd713RealData
//	@Schemes
//	@Description	ask Ed713 device to send real data
//	@Tags			ED713 device
//	@Param			token	query	string		false	"token"
//	@Param			in	body	mdb.AskEd713RealDataReq	true	"ack Ed713 info"
//	@Produce		json
//	@Success		200	{none} {none}
//	@Router			/device/askEd713RealData [post]
func askEd713RealData(c *gin.Context) {
	exception.TryEx{
		Try: func() {
			status, result := mdb.AskEd713RealData(c)
			respJSON(c, status, result)
		},
		Catch: func(e exception.Exception) {
			respJSON(c, http.StatusBadRequest, e.Msg)
		},
	}.Run()
}
