/******************************************************************************
 * Author: liguoqiang
 * Date: 2024-07-12 17:19:37
 * LastEditors: liguoqiang
 * LastEditTime: 2024-12-16 17:04:14
 * Description:
********************************************************************************/
package mdbwx

import (
	"encoding/json"
	"fmt"
	"hjyserver/cfg"
	mylog "hjyserver/log"
	"hjyserver/mdb/common"
	"hjyserver/mdb/mysql"
	mysqlwx "hjyserver/wx/mdb/mysql"
	wxtools "hjyserver/wx/tools"

	"github.com/gin-gonic/gin"
)

const tag = "mdb_wx"

//swagger:model WxMiniLoginReq
type WxMiniLoginReq struct {
	LoginCode   string `json:"login_code"`
	PhoneCode   string `json:"phone_code"`
	NickName    string `json:"nick_name"`
	Gender      int    `json:"gender"`
	AvatarUrl   string `json:"avatar_url"`
	EncryptData string `json:"encript_data"`
	Iv          string `json:"iv"`
	Version     string `json:"version"`
}

//swagger:model WxMiniSession
type WxMiniSession struct {
	OpenId     string `json:"openId"`
	SessionKey string `json:"session_key" `
	UnionId    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

//swagger:model WxMiniLoginResp
type WxMiniLoginResp struct {
	mysql.User
	Gender int    `json:"gender"`
	Token  string `json:"token"`
}

/******************************************************************************
 * function: WxMiniProgramLogin
 * description: 微信小程序注册登录
 * param {*gin.Context} c
 * return {*}
********************************************************************************/
func WxMiniProgramLogin(c *gin.Context) (int, interface{}) {
	var req *WxMiniLoginReq = &WxMiniLoginReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		return common.JsonError, err.Error()
	}
	if req.LoginCode == "" {
		return common.CodeError, "code required"
	}
	// 如果没有注册，则根据手机号码注册用户
	if req.PhoneCode == "" {
		return common.PhoneError, "phone required"
	}
	if req.EncryptData == "" || req.Iv == "" {
		return common.ParamError, "encrypt or iv field required"
	}
	mylog.Log.Debugln(tag, "WxMiniProgramLogin, req: ", req)
	// 获取小程序的session_key
	session, err := getWxMiniSessionByCode(req.LoginCode)
	if err != nil {
		mylog.Log.Errorln(tag, "getWxMiniSessionByCode", err)
		return common.WxError, err.Error()
	}
	if session.OpenId == "" {
		return common.CodeError, "not get openId"
	}
	// 解密用户信息
	key, err := common.ConvertBase64ToBytes(session.SessionKey)
	if err != nil {
		mylog.Log.Errorln(tag, "ConvertBase64ToBytes", err)
		return common.WxError, err.Error()
	}
	iv, err := common.ConvertBase64ToBytes(req.Iv)
	if err != nil {
		mylog.Log.Errorln(tag, "ConvertBase64ToBytes", err)
		return common.WxError, err.Error()
	}
	decryptData, err := common.DecryptDataWithCBC(key, iv, req.EncryptData)
	if err != nil {
		mylog.Log.Errorln(tag, "DecryptData", err)
		return common.WxError, err.Error()
	}
	mylog.Log.Infoln(tag, "decryptData: ", decryptData)
	type WxMiniUserInfo struct {
		NickName  string `json:"nickName"`
		Gender    int    `json:"gender"`
		AvatarUrl string `json:"avatarUrl"`
		City      string `json:"city"`
		Country   string `json:"country"`
		Language  string `json:"language"`
		Province  string `json:"province"`
		WaterMark struct {
			AppId string `json:"appid"`
		} `json:"watermark"`
	}
	wxUserInfo := &WxMiniUserInfo{}
	err = json.Unmarshal([]byte(decryptData), wxUserInfo)
	if err != nil {
		mylog.Log.Errorln(tag, "Unmarshal encrypt failed,", err)
	}
	req.Gender = wxUserInfo.Gender
	req.NickName = wxUserInfo.NickName
	if req.AvatarUrl == "" {
		req.AvatarUrl = wxUserInfo.AvatarUrl
	}
	// 获取用户手机号码
	phoneInfo, err := getWxMiniPhoneNumber(req.PhoneCode, session.OpenId)
	if err != nil {
		mylog.Log.Errorln(tag, "getWxMiniPhoneNumber", err)
		return common.WxError, err.Error()
	}
	phone := fmt.Sprintf("+%s%s", phoneInfo.CountryCode, phoneInfo.PhoneNumber)
	// 查询用户是否已经注册，如果已经注册则返回用户信息mysql.user
	var wxMiniProgram []mysqlwx.WxMiniProgram
	mysqlwx.QueryWxMiniProgramByOpenId(session.OpenId, &wxMiniProgram)
	if len(wxMiniProgram) > 0 {
		// 有可能是多个用户使用同一个微信小程序，需要遍历查找
		// 查找手机号码相同的用户ID
		for _, wx := range wxMiniProgram {
			userId := wx.UserId
			resp := &WxMiniLoginResp{}
			if resp.QueryByID(userId) {
				if resp.Phone == phone {
					resp.Token, _ = mysql.GetUserToken(&resp.User)
					resp.IsLogin = 1
					resp.LoginTime = common.GetNowTime()
					resp.Update()
					if wx.Gender != req.Gender {
						wx.Gender = req.Gender
						wx.Update()
					}
					return common.Success, resp
				}
			}
		}
	}
	// 先根据手机号码注册用户
	status, result := registerUserWithPhone(phone, req.NickName, req.Gender, req.AvatarUrl)
	if status != common.Success {
		return status, result
	}
	user := result.(*mysql.User)
	wxObj := mysqlwx.NewWxMiniProgram()
	wxObj.UserId = user.ID
	wxObj.OpenId = session.OpenId
	wxObj.SessionKey = session.SessionKey
	wxObj.NickName = req.NickName
	wxObj.Gender = req.Gender
	wxObj.AvatarUrl = req.AvatarUrl
	wxObj.Version = req.Version
	wxObj.UnionId = session.UnionId
	wxObj.CreateTime = common.GetNowTime()
	// 插入微信小程序信息
	if !wxObj.Insert() {
		return common.DBError, "insert wx mini program failed"
	}
	resp := &WxMiniLoginResp{}
	resp.User = *result.(*mysql.User)
	resp.Gender = req.Gender
	resp.Token, _ = mysql.GetUserToken(&resp.User)
	return common.Success, resp
}

/******************************************************************************
 * function: getWxMiniSessionByCode
 * description: 获取小程序的session_key
 * param {string} code
 * return {*}
********************************************************************************/
func getWxMiniSessionByCode(code string) (*WxMiniSession, error) {
	uri := "https://api.weixin.qq.com/sns/jscode2session"
	params := map[string]string{
		"appid":      cfg.This.Wx.MinAppId,
		"secret":     cfg.This.Wx.MinAppSecret,
		"js_code":    code,
		"grant_type": "authorization_code",
	}
	result, err := common.HttpGet(uri, params)
	if err != nil {
		return nil, err
	}
	mylog.Log.Debugln(tag, "getWxMiniSessionByCode, result: ", string(result))
	wxSession := &WxMiniSession{}
	err = json.Unmarshal(result, wxSession)
	if err != nil {
		return nil, err
	}
	if wxSession.ErrCode != 0 {
		return nil, fmt.Errorf("errcode: %d, errmsg: %s", wxSession.ErrCode, wxSession.ErrMsg)
	}
	return wxSession, nil
}

type WxMiniUserPhone struct {
	PhoneNumber     string `json:"phoneNumber"`
	PurePhoneNumber string `json:"purePhoneNumber"`
	CountryCode     string `json:"countryCode"`
}

/******************************************************************************
 * function: getWxMiniPhoneNumber
 * description: 获取微信小程序用户的手机号码
 * param {string} phoneCode
 * param {string} openId
 * return {*}
********************************************************************************/
func getWxMiniPhoneNumber(phoneCode string, openId string) (*WxMiniUserPhone, error) {
	token, err := wxtools.QueryWxMiniAccessToken()
	if err != nil {
		return nil, err
	}
	type PhoneReq struct {
		Code   string `json:"code"`
		OpenId string `json:"openid"`
	}

	req := &PhoneReq{
		Code:   phoneCode,
		OpenId: openId,
	}
	params, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	mylog.Log.Debugln(tag, "getWxMiniPhoneNumber, token: ", token)
	uri := "https://api.weixin.qq.com/wxa/business/getuserphonenumber?access_token=" + token
	result, err := common.HttpPostJson(uri, params)
	if err != nil {
		return nil, err
	}
	mylog.Log.Debugln(tag, "getWxMiniPhoneNumber, result: ", string(result))
	type WxMiniUserPhoneResp struct {
		ErrCode   int             `json:"errcode"`
		ErrMsg    string          `json:"errmsg"`
		PhoneInfo WxMiniUserPhone `json:"phone_info"`
	}
	phone := &WxMiniUserPhoneResp{}
	err = json.Unmarshal(result, phone)
	if err != nil {
		return nil, err
	}
	if phone.ErrCode != 0 {
		return nil, fmt.Errorf("errcode: %d, errmsg: %s", phone.ErrCode, phone.ErrMsg)
	}
	return &phone.PhoneInfo, nil
}

/******************************************************************************
 * function: registerUserWithPhone
 * description: 根据手机号码注册用户，如果用户已经存在，则直接返回用户信息
 *  默认密码为888888
 * param {string} phone
 * param {string} nickName
 * param {string} avatarUrl
 * return {*}
********************************************************************************/
func registerUserWithPhone(phone string, nickName string, gender int, avatarUrl string) (int, interface{}) {
	filter := fmt.Sprintf("phone like '%%%s%%'", phone)
	users := make([]mysql.User, 0)
	mysql.QueryUserByCond(filter, nil, nil, &users)
	if len(users) > 0 {
		users[0].Gender = gender
		users[0].IsLogin = 1
		users[0].LoginTime = common.GetNowTime()
		users[0].Update()
		return common.Success, &users[0]
	}
	userObj := mysql.NewUser()
	userObj.Account = phone
	userObj.Phone = phone
	userObj.NickName = nickName
	userObj.Gender = gender
	userObj.Face = avatarUrl
	userObj.Password = "888888"
	userObj.IsLogin = 1
	return mysql.RegisterWithUserObj(userObj)
}

/******************************************************************************
 * function: WxPublicSubmit
 * description:  接收微信公众号推送的服务器地址
 * param {*gin.Context} c
 * return {*}
********************************************************************************/
//swagger:model WxPublicSubmitReq
type WxPublicSubmitReq struct {
	ToUserName   string `xml:"ToUserName"`
	FromUserName string `xml:"FromUserName"`
	CreateTime   int64  `xml:"CreateTime"`
	MsgType      string `xml:"MsgType"`
	Event        string `xml:"Event"`
}

// 接收微信公众号推送的服务器地址
func WxPublicSubmit(c *gin.Context) (int, interface{}) {
	if c.Request.Method == "GET" {
		echoStr := c.Query("echostr")
		return common.Success, echoStr
	}
	var req *WxPublicSubmitReq = &WxPublicSubmitReq{}
	err := c.ShouldBindXML(req)
	if err != nil {
		mylog.Log.Errorln(tag, err)
		return common.JsonError, err.Error()
	}
	mylog.Log.Debugln(tag, "WxPublicSubmit", req)
	// 取消关注时，则删除关注记录
	if req.MsgType == "event" && req.Event == "unsubscribe" {
		mysqlwx.DeleteWxOfficalAccountByEvent(req.FromUserName, req.MsgType, "subscribe")
		return common.Success, "success"
	}
	if req.MsgType != "event" || req.Event != "subscribe" {
		return common.Success, "success"
	}
	// 查询access_token
	accessToken, err := wxtools.QueryWxOfficalAccessToken()
	if err != nil {
		mylog.Log.Errorln(tag, "QueryWxAccessToken", err)
		return common.WxError, err.Error()
	}
	userInfo, err := wxtools.QueryWxUserBaseInfo(accessToken, req.FromUserName)
	if err != nil {
		mylog.Log.Errorln(tag, "QueryWxUserBaseInfo", err)
		return common.WxError, err.Error()
	}
	mylog.Log.Debugln(tag, "accessToken:", accessToken, "user unionId:", userInfo.UnionId)
	// 插入关注记录
	officalAccount := mysqlwx.NewWxOfficalAccount()
	officalAccount.ToUserName = req.ToUserName
	officalAccount.FromOpenId = req.FromUserName
	officalAccount.MsgType = req.MsgType
	officalAccount.Event = req.Event
	officalAccount.FromUnionId = userInfo.UnionId
	officalAccount.CreateTime = common.SecondsToTimeStr(req.CreateTime)
	officalList := mysqlwx.QueryWxOfficalAccountByOpenIdAndMsgType(req.FromUserName, req.MsgType)
	if len(officalList) > 0 {
		officalAccount.ID = officalList[0].ID
		officalAccount.Update()
	} else {
		if !officalAccount.Insert() {
			return common.DBError, "insert wx offical account failed"
		}
	}
	// 发送一条欢迎消息
	go func() {
		// var content string = fmt.Sprintf("<a data-miniprogram-appid=\"%s\" data-miniprogram-path=\"pages/welcome/welcome\">点击跳小程序</a>", cfg.This.Wx.MinAppId)
		wxtools.SendCustomerWelcomeCardToOfficalAccount(req.FromUserName)
	}()
	return common.Success, "success"
}

/******************************************************************************
 * function: QueryMiniAccessToken
 * description:  获取微信小程序的access_token
 * param {*gin.Context} c
 * return {*}
********************************************************************************/

//swagger:model QueryMiniAccessTokenResp
type QueryMiniAccessTokenResp struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

func QueryMiniAccessToken(c *gin.Context) (int, interface{}) {
	token, err := wxtools.QueryWxMiniAccessToken()
	if err != nil {
		return common.WxError, err.Error()
	}
	resp := &QueryMiniAccessTokenResp{
		AccessToken: token,
		ExpiresIn:   7200,
	}
	return common.Success, resp
}

// swagger:model WxMiniUrlReq
type WxMiniUrlReq struct {
	Path  string `json:"path"`
	Query string `json:"query"`
}

func AskMiniUrl(c *gin.Context) (int, interface{}) {
	req := &WxMiniUrlReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		return common.JsonError, err.Error()
	}
	return wxtools.GeneratorWxMiniUrl(req.Path, req.Query)
}
