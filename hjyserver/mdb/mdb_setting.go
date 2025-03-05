package mdb

import (
	"hjyserver/mdb/common"
	"hjyserver/mdb/mysql"

	"github.com/gin-gonic/gin"
)

/******************************************************************************
 * function:
 * description:
 * param {*gin.Context} c
 * return {*}
********************************************************************************/
func InsertBanner(c *gin.Context) (int, interface{}) {
	banner := mysql.NewBannerSetting()
	err := c.ShouldBindJSON(banner)
	if err != nil {
		return common.JsonError, err.Error()
	}
	bannerList := make([]mysql.BannerSetting, 0)
	if banner.Sort > -1 {
		mysql.QueryBannerBySort(banner.Sort, &bannerList)
		if len(bannerList) > 0 {
			if bannerList[0].Sort == banner.Sort {
				bannerList[0].ImgUrl = banner.ImgUrl
				bannerList[0].Update()
				return common.Success, bannerList[0]
			}
		}
	}
	if banner.Sort == -1 {
		mysql.QueryAllBanner(&bannerList)
		if len(bannerList) > 0 {
			banner.Sort = bannerList[len(bannerList)-1].Sort + 1
		} else {
			banner.Sort = 0
		}
	}
	if banner.Insert() {
		return common.Success, banner
	}
	return common.DBError, "Insert banner failed"
}

/******************************************************************************
 * function:
 * description:
 * param {*gin.Context} c
 * return {*}
********************************************************************************/
func QueryBanner(c *gin.Context) (int, interface{}) {
	bannerList := make([]mysql.BannerSetting, 0)
	mysql.QueryAllBanner(&bannerList)
	return common.Success, bannerList
}
