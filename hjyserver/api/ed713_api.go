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

//	@BasePath	/v1

// askEd713RealData godoc
//
//	@Summary	askEd713RealData
//	@Schemes
//	@Description	ask Ed713 device to send real data
//	@Tags			ED713 device
//
//	@Param			in	body	mdb.AskEd713RealDataReq	true	"ack Ed713 info"
//	@Produce		json
//	@Success		200	{none} {none}
//	@Router			/v1/device/askEd713RealData [post]
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
