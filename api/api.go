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
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"hjyserver/cfg"
	"hjyserver/exception"
	"hjyserver/mdb/common"
	"hjyserver/mdb/mysql"
	wxapi "hjyserver/wx/api"

	docs "hjyserver/docs"

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
	// 设置路由版本
	verApi := router.Group("/" + cfg.This.Svr.ApiVersion)
	// 设置swagger版本信息
	docs.SwaggerInfo.BasePath = "/" + cfg.This.Svr.ApiVersion
	// 初始化路由
	initActions()

	for k, v := range getAction {
		verApi.GET(k, tollbooth_gin.LimitHandler(limt), v)
	}
	for k, v := range postAction {
		if cfg.This.Svr.ApiVersion == "v1" {
			verApi.POST(k, tollbooth_gin.LimitHandler(limt), v)
		} else {
			verApi.POST(k, tollbooth_gin.LimitHandler(limt), AuthorizeToken, v)
		}
	}

	// 初始化用户接口
	userPosts, userGets := InitUserActions()
	for k, v := range userGets {
		verApi.GET(k, tollbooth_gin.LimitHandler(limt), v)
	}
	for k, v := range userPosts {
		if cfg.This.Svr.ApiVersion == "v1" {
			verApi.POST(k, tollbooth_gin.LimitHandler(limt), v)
		} else {
			verApi.POST(k, tollbooth_gin.LimitHandler(limt), AuthorizeToken, v)
		}
	}
	// 初始化设备接口
	devicesPost, devicesGets := InitDeviceActions()
	for k, v := range devicesGets {
		verApi.GET(k, tollbooth_gin.LimitHandler(limt), v)
	}
	for k, v := range devicesPost {
		if cfg.This.Svr.ApiVersion == "v1" {
			verApi.POST(k, tollbooth_gin.LimitHandler(limt), v)
		} else {
			verApi.POST(k, tollbooth_gin.LimitHandler(limt), AuthorizeToken, v)
		}
	}

	// 初始化微信接口
	wxPosts, wxGets := wxapi.InitWxActions()
	for k, v := range wxGets {
		verApi.GET(k, tollbooth_gin.LimitHandler(limt), v)
	}
	for k, v := range wxPosts {
		verApi.POST(k, tollbooth_gin.LimitHandler(limt), v)
	}
	// 初始化H03接口
	h03Ports, h03Gets := InitH03Actions()
	for k, v := range h03Gets {
		verApi.GET(k, tollbooth_gin.LimitHandler(limt), v)
	}
	for k, v := range h03Ports {
		if cfg.This.Svr.ApiVersion == "v1" {
			verApi.POST(k, tollbooth_gin.LimitHandler(limt), v)
		} else {
			verApi.POST(k, tollbooth_gin.LimitHandler(limt), AuthorizeToken, v)
		}
	}
	// 初始化T1接口
	t1Ports, t1Gets := InitT1Actions()
	for k, v := range t1Gets {
		verApi.GET(k, tollbooth_gin.LimitHandler(limt), v)
	}
	for k, v := range t1Ports {
		if cfg.This.Svr.ApiVersion == "v1" {
			verApi.POST(k, tollbooth_gin.LimitHandler(limt), v)
		} else {
			verApi.POST(k, tollbooth_gin.LimitHandler(limt), AuthorizeToken, v)
		}
	}
	// 初始化X1s接口
	x1sPorts, x1sGets := InitX1sActions()
	for k, v := range x1sGets {
		verApi.GET(k, tollbooth_gin.LimitHandler(limt), v)
	}
	for k, v := range x1sPorts {
		if cfg.This.Svr.ApiVersion == "v1" {
			verApi.POST(k, tollbooth_gin.LimitHandler(limt), v)
		} else {
			verApi.POST(k, tollbooth_gin.LimitHandler(limt), AuthorizeToken, v)
		}
	}
	// 初始化setting接口
	settingPorts, settingGets := InitSettingActions()
	for k, v := range settingGets {
		verApi.GET(k, tollbooth_gin.LimitHandler(limt), v)
	}
	for k, v := range settingPorts {
		if cfg.This.Svr.ApiVersion == "v1" {
			verApi.POST(k, tollbooth_gin.LimitHandler(limt), v)
		} else {
			verApi.POST(k, tollbooth_gin.LimitHandler(limt), AuthorizeToken, v)
		}
	}

	router.MaxMultipartMemory = 8 << 40
	if cfg.This.Svr.ApiVersion == "v1" {
		router.Static("/public", cfg.This.StaticPath)
	} else {
		verApi.Static("/public", cfg.This.StaticPath)
	}
	verApi.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	// router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 单独启动http server，用于后面的关闭操作
	// 如果设置TLS，则启动https服务
	if cfg.This.Svr.EnableTls {
		// tsConfig := tls.Config{
		// 	InsecureSkipVerify:       false,
		// 	MinVersion:               tls.VersionTLS12,
		// 	PreferServerCipherSuites: true,
		// }
		// cert, err := tls.LoadX509KeyPair(cfg.This.Svr.CertFile, cfg.This.Svr.KeyFile)
		// if err != nil {
		// 	mylog.Log.Errorln(err)
		// 	return
		// }
		// tsConfig.Certificates = []tls.Certificate{cert}
		// caPool := x509.NewCertPool()
		// caPem, err := os.ReadFile(cfg.This.Svr.CaFile)
		// if err != nil {
		// 	mylog.Log.Errorln(err)
		// 	return
		// }
		// caPool.AppendCertsFromPEM(caPem)
		// tsConfig.RootCAs = caPool
		// svcHttp = &http.Server{
		// 	Addr:         cfg.This.Svr.Host,
		// 	Handler:      router,
		// 	ReadTimeout:  120 * time.Second,
		// 	WriteTimeout: 120 * time.Second,
		// 	TLSConfig:    &tsConfig,
		// }
		svcHttp = &http.Server{
			Addr:         cfg.This.Svr.Host,
			Handler:      router,
			ReadTimeout:  120 * time.Second,
			WriteTimeout: 120 * time.Second,
		}
	} else {
		svcHttp = &http.Server{
			Addr:         cfg.This.Svr.Host,
			Handler:      router,
			ReadTimeout:  120 * time.Second,
			WriteTimeout: 120 * time.Second,
			TLSConfig:    nil,
		}
	}
	// 启动一个go例程用于启动服务
	go func() {
		var err error
		if cfg.This.Svr.EnableTls {
			err = svcHttp.ListenAndServeTLS(cfg.This.Svr.CertFile, cfg.This.Svr.KeyFile)
		} else {
			err = svcHttp.ListenAndServe()
		}
		if err != nil {
			mylog.Log.Errorf("start web server failed, %s, %v", cfg.This.Svr.Host, err)
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

	getAction["/notify/queryNotifySettingByType"] = queryNotifySettingByType
	getAction["/notify/queryAllNotifySetting"] = queryAllNotifySetting

	postAction = make(map[string]gin.HandlerFunc)

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
	if page == nil {
		page = common.NewPageDao(1, 10)
	}
	return page
}

/******************************************************************************
 * function: AuthorizeToken
 * description: 接口拦截器，用来验证token是否有效
 * return {*}
********************************************************************************/
func AuthorizeToken(c *gin.Context) {
	uri := c.Request.URL.String()
	matched, err := regexp.Match("/swagger/*", []byte(uri))
	fmt.Println("uri: ", uri, "matched: ", matched, "err: ", err)
	if err == nil && matched {
		c.Next()
		return
	}
	// 如果是V1版本的接口，不需要验证token
	if uri[1:3] == "v1" {
		c.Next()
		return
	}

	if uri == "/"+cfg.This.Svr.ApiVersion+"/user/userLogin" ||
		uri == "/"+cfg.This.Svr.ApiVersion+"/user/loginout" ||
		uri == "/"+cfg.This.Svr.ApiVersion+"/user/userRegister" ||
		uri == "/"+cfg.This.Svr.ApiVersion+"/wx/wxMiniProgramLogin" ||
		uri == "/"+cfg.This.Svr.ApiVersion+"/wx/wxPublicSubmit" {
		c.Next()
		return
	}
	token := c.Query("token")
	mylog.Log.Debug("token: ", token)
	if token == "" {
		respJSON(c, common.TokenError, "token required")
		c.Abort()
		return
	}
	if !mysql.VerifyUserToken(token) {
		respJSON(c, common.TokenError, "token invalid")
		c.Abort()
		return
	}
	c.Next()
}

/*
上传图片接口
*/
// uploadPicFun godoc
//
//	@Summary	uploadPicFun
//	@Schemes
//	@Description	上传图片接口
//	@Tags			uploader
//	@Produce		json
//
//	@Param			token	query	string		false	"token"
//	@Param			file formData	file		true	"file to upload"
//
//	@Success		200		{string}	{string}
//	@Router			/upload/picture [post]
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
