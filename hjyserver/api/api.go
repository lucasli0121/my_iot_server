/*
 * @Author: liguoqiang
 * @Date: 2022-06-02 17:04:32
 * @LastEditors: liguoqiang
 * @LastEditTime: 2024-04-30 15:10:37
 * @Description:
 */
package api

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"time"

	"hjyserver/cfg"
	"hjyserver/exception"
	"hjyserver/mdb/common"

	_ "hjyserver/docs"

	mylog "hjyserver/log"

	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth_gin"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var postAction map[string]gin.HandlerFunc
var getAction map[string]gin.HandlerFunc
var svcHttp *http.Server = nil

// StartWeb function run a webservice at webPort
func StartWeb() {
	// 设置限流
	limt := tollbooth.NewLimiter(100, nil)
	limt.SetIPLookups([]string{"RemoteAddr", "X-Forwarded-For", "X-Real-IP"}).SetMethods([]string{"GET", "POST"})
	limt.SetMessage("{ \"code\": 201, \"message\": \"reached max request limit\"}")
	router := gin.Default()
	v1 := router.Group("/v1")
	initActions()
	for k, v := range getAction {
		v1.GET(k, tollbooth_gin.LimitHandler(limt), v)
	}
	for k, v := range postAction {
		v1.POST(k, tollbooth_gin.LimitHandler(limt), v)
	}
	router.MaxMultipartMemory = 8 << 40
	router.Static("/public", cfg.This.StaticPath)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	// 单独启动http server，用于后面的关闭操作
	svcHttp = &http.Server{
		Addr:         cfg.This.Svr.Host,
		Handler:      router,
		ReadTimeout:  120 * time.Second,
		WriteTimeout: 120 * time.Second,
	}
	// 启动一个go例程用于启动服务
	go func() {
		err := svcHttp.ListenAndServeTLS(cfg.This.CertFile, cfg.This.KeyFile)
		if err != nil {
			mylog.Log.Errorf("start web server failed, %s", cfg.This.Svr.Host)
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, os.Kill)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	svcHttp.Shutdown(ctx)
	<-ctx.Done()
	mylog.Log.Info("Shutdowning is done!!")
}

func initActions() {
	getAction = make(map[string]gin.HandlerFunc)
	getAction["/device/queryById"] = queryDeviceById
	getAction["/device/queryByUser"] = queryDeviceByUser
	getAction["/device/queryBindByMac"] = queryBindByMac
	getAction["/user/queryById"] = queryUserById
	getAction["/user/queryUserByPhone"] = queryUserByPhone
	getAction["/user/queryUserByEmail"] = queryUserByEmail
	getAction["/user/queryUserGroup"] = queryUserGroup
	getAction["/user/queryFriendsByUser"] = queryFriendsByUser
	getAction["/user/queryLampUsersByRoom"] = queryLampUsersByRoom
	getAction["/user/queryLampUserInFriend"] = queryLampUserInFriend
	getAction["/device/queryHeartRate"] = queryHeartRate
	getAction["/device/statsHeartRateByMinute"] = statsHeartRateByMinute
	getAction["/device/queryX1RealDataJson"] = queryX1RealDataJson
	getAction["/device/queryX1SleepReportJson"] = queryX1SleepReportJson
	getAction["/device/querySleepReport"] = querySleepReport
	getAction["/device/queryDateListInReport"] = queryDateListInReport
	getAction["/device/queryFallCheckStatus"] = queryFallCheckStatus
	getAction["/device/queryAlarmRecord"] = queryAlarmRecord
	getAction["/device/queryFallExistRecord"] = queryFallExistRecord
	getAction["/device/queryFallParams"] = queryFallParams
	getAction["/device/queryLampRealData"] = queryLampRealData
	getAction["/device/queryLampEvent"] = queryLampEvent
	getAction["/device/queryLampControl"] = queryLampControl
	getAction["/device/queryLampReportStatus"] = queryLampReportStatus
	getAction["/device/queryStudyRoom"] = queryStudyRoom
	getAction["/device/queryInviteStudyRoom"] = queryInviteStudyRoom
	getAction["/device/queryRankingByStudyRoom"] = queryRankingByStudyRoom
	getAction["/device/statsLampFlowData"] = statsLampFlowData
	getAction["/user/queryStudyRoomUser"] = queryStudyRoomUser
	getAction["/user/queryUserStudyData"] = queryUserStudyData
	getAction["/user/queryUserStudyTimeByDay"] = queryUserStudyTimeByDay
	getAction["/notify/queryNotifySettingByType"] = queryNotifySettingByType
	getAction["/notify/queryAllNotifySetting"] = queryAllNotifySetting

	postAction = make(map[string]gin.HandlerFunc)
	// post device tag action
	postAction["/device/insert"] = insertDevice
	postAction["/device/update"] = updateDevice
	postAction["/device/share"] = shareDevice
	postAction["/device/insertFallParams"] = insertFallParams
	postAction["/device/openLampRealData"] = openLampRealData
	postAction["/device/controlLamp"] = controlLamp
	postAction["/device/createStudyRoom"] = createStudyRoom
	postAction["/device/modifyStudyRoom"] = modifyStudyRoom
	postAction["/device/releaseStudyRoom"] = releaseStudyRoom
	postAction["/device/askEd713RealData"] = askEd713RealData
	postAction["/device/askX1RealData"] = askX1RealData
	postAction["/device/cleanX1Event"] = cleanX1Event
	postAction["/device/sleepX1Switch"] = sleepX1Switch
	postAction["/device/improveDisturbed"] = improveDisturbed
	// post user tag action
	postAction["/user/userLogin"] = userLogin
	postAction["/user/userRegister"] = userRegister
	postAction["/user/loginout"] = loginOut
	postAction["/user/deleteUser"] = deleteUser
	postAction["/user/online"] = userOnline
	postAction["/user/offline"] = userOffline
	postAction["/user/updateNickName"] = updateNickName
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
	// post notify setting action
	// postAction["/notify/peopleNotifySetting"] = peopleNotifySetting
	// postAction["/notify/breathNotifySetting"] = breathNotifySetting
	// postAction["/notify/breathAbnormalNotifySetting"] = breathAbnormalNotifySetting
	// postAction["/notify/heartRateNotifySetting"] = heartRateNotifySetting
	// postAction["/notify/nurseModelSetting"] = nurseModelSetting
	// postAction["/notify/beeperSetting"] = beeperSetting
	// postAction["/notify/lightSetting"] = lightSetting
	postAction["/notify/notifySetting"] = notifySetting

	postAction["/upload/picture"] = uploadPicFun
	postAction["/upload/video"] = uploadVideoFun
	postAction["/upload/voice"] = uploadVoiceFun
	postAction["/upload/file"] = uploadFileFun
}

func getPageDaoFromGin(c *gin.Context) *common.PageDao {
	pageNo := c.Query("pageNo")
	pageSize := c.Query("pageSize")
	var page *common.PageDao = nil
	if pageNo != "" {
		no, err1 := strconv.ParseInt(pageNo, 10, 64)
		size, err2 := strconv.ParseInt(pageSize, 10, 64)
		if err1 == nil && err2 == nil {
			page = common.NewPageDao(no, size)
		}
	}
	return page
}

/*
上传图片接口
*/
func uploadPicFun(c *gin.Context) {
	uploadFileFunc(c, cfg.StaticPicPath)
}

/*
上传文件接口
*/
func uploadFileFun(c *gin.Context) {
	uploadFileFunc(c, cfg.StaticFilePath)
}

/*
上传视频文件接口
*/
func uploadVideoFun(c *gin.Context) {
	uploadFileFunc(c, cfg.StaticVideoPath)
}

/*
上传音频文件
*/
func uploadVoiceFun(c *gin.Context) {
	uploadFileFunc(c, cfg.StaticVoicePath)
}

func uploadFileFunc(c *gin.Context, staticPath string) {
	exception.TryEx{
		Try: func() {
			file, err := c.FormFile("file")
			if err != nil {
				exception.Throw(http.StatusBadRequest, "upload file error")
			}
			filename := staticPath + filepath.Base(file.Filename)
			if err := c.SaveUploadedFile(file, filename); err != nil {
				exception.Throw(http.StatusBadRequest, "upload file error")
			}
			respJSON(c, http.StatusOK, cfg.This.Svr.OutUrl+filename)
		},
		Catch: func(e exception.Exception) {
			respJSON(c, e.Code, e.Msg)
		},
	}.Run()
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
