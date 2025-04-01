package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hjyserver/cfg"
	mylog "hjyserver/log"
	"hjyserver/mdb/common"
	"hjyserver/redis"
	mysqlwx "hjyserver/wx/mdb/mysql"
	"strings"
)

const tag = "wx_tools"
const (
	WxOfficalAccessTokenKey = "wx_offical_access_token_key"
	WxMiniAccessTokenKey    = "wx_mini_access_token_key"
)

type RestuleError struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

/******************************************************************************
 * function: QueryWxOfficalAccessToken
 * description:  查询微信公众号的access_token， 返回access_token, error
 * return {*}
********************************************************************************/
func QueryWxOfficalAccessToken() (string, error) {
	accessToken, err := redis.GetValue(WxOfficalAccessTokenKey)
	if err != nil || accessToken == "" {
		newToken, expiresIn, err := QueryWxAccessToken(cfg.This.Wx.PublicAppId, cfg.This.Wx.PublicAppSecret)
		if err != nil {
			return "", err
		}
		accessToken = newToken
		redis.SetValueEx(WxOfficalAccessTokenKey, accessToken, expiresIn-60)
	}
	return accessToken, nil
}

/******************************************************************************
 * function: QueryWxMiniAccessToken
 * description: 获取微信小程序的access_token
 * return {*}
********************************************************************************/
func QueryWxMiniAccessToken() (string, error) {
	accessToken, err := redis.GetValue(WxMiniAccessTokenKey)
	if err != nil || accessToken == "" {
		newToken, expiresIn, err := QueryWxAccessToken(cfg.This.Wx.MinAppId, cfg.This.Wx.MinAppSecret)
		if err != nil {
			return "", err
		}
		accessToken = newToken
		redis.SetValueEx(WxMiniAccessTokenKey, accessToken, expiresIn-60)
	}
	return accessToken, nil
}

/******************************************************************************
 * function: QueryWxAccessToken
 * description: 根据APPiD和secret查询微信的access_token
 * param {string} appId
 * param {string} secret
 * return {*}
********************************************************************************/
func QueryWxAccessToken(appId string, secret string) (string, int, error) {
	uri := cfg.This.Wx.AccessTokenUri
	type TokenReq struct {
		GrantType    string `json:"grant_type"`
		AppId        string `json:"appid"`
		Secret       string `json:"secret"`
		ForceRefresh bool   `json:"force_refresh"`
	}
	req := &TokenReq{
		GrantType:    "client_credential",
		AppId:        appId,
		Secret:       secret,
		ForceRefresh: false,
	}
	params, err := json.Marshal(req)
	if err != nil {
		return "", 0, err
	}
	result, err := common.HttpPostJson(uri, params)
	if err != nil {
		return "", 0, err
	}
	type TokenResult struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
	}
	mylog.Log.Info(tag, "QueryWxAccessToken, result: ", string(result))
	var tokenResult TokenResult
	err = json.Unmarshal(result, &tokenResult)
	if err != nil {
		return "", 0, err
	}
	if tokenResult.ErrCode != 0 {
		return "", 0, fmt.Errorf("errcode: %d, errmsg: %s", tokenResult.ErrCode, tokenResult.ErrMsg)
	}
	return tokenResult.AccessToken, tokenResult.ExpiresIn, nil
}

type WxUserBaseInfo struct {
	Subscribe     int    `json:"subscribe"`
	OpenId        string `json:"openid"`
	SubscribeTime int64  `json:"subscribe_time"`
	UnionId       string `json:"unionid"`
	Remark        string `json:"remark"`
	GroupId       int    `json:"groupid"`
	ErrCode       int    `json:"errcode"`
	ErrMsg        string `json:"errmsg"`
}

/******************************************************************************
 * function: QueryWxUserBaseInfo
 * description: 查询微信公众号用户基本信息
 * param {string} accessToken
 * param {string} openId
 * return {*}
********************************************************************************/
func QueryWxUserBaseInfo(accessToken string, openId string) (*WxUserBaseInfo, error) {
	mylog.Log.Debugln(tag, "QueryWxUserBaseInfo, accessToken: ", accessToken, "openId: ", openId)
	uri := "https://api.weixin.qq.com/cgi-bin/user/info"
	var params map[string]string = make(map[string]string)
	params["access_token"] = accessToken
	params["openid"] = openId
	params["lang"] = "zh_CN"
	result, err := common.HttpGet(uri, params)
	if err != nil {
		mylog.Log.Errorln(tag, "QueryWxUserBaseInfo", err)
		return nil, err
	}
	mylog.Log.Debug(tag, "QueryWxUserBaseInfo, result: ", string(result))
	userInfo := &WxUserBaseInfo{
		ErrCode: 0,
	}
	err = json.Unmarshal(result, &userInfo)
	if err != nil {
		mylog.Log.Errorln(tag, "QueryWxUserBaseInfo", err)
		return nil, err
	}
	if userInfo.ErrCode != 0 {
		return nil, fmt.Errorf("errcode: %d, errmsg: %s", userInfo.ErrCode, userInfo.ErrMsg)
	}
	return userInfo, nil
}

/******************************************************************************
 * function: SendCustomerTextMsgToOfficalAccount
 * description: 发送一个客服文本消息给微信公众号，文本消息根据结构定义
 * param {string} openId
 * param {string} content
 * return {*}
********************************************************************************/
func SendCustomerTextMsgToOfficalAccount(openId string, content string) (int, string) {
	accessToken, err := QueryWxOfficalAccessToken()
	if err != nil {
		return common.WxError, err.Error()
	}
	uri := "https://api.weixin.qq.com/cgi-bin/message/custom/send?access_token=" + accessToken
	type CustomMsgReq struct {
		ToUser  string `json:"touser"`
		MsgType string `json:"msgtype"`
		Text    struct {
			Content string `json:"content"`
		} `json:"text"`
	}
	req := &CustomMsgReq{
		ToUser:  openId,
		MsgType: "text",
	}
	req.Text.Content = content
	params, err := json.Marshal(req)
	if err != nil {
		return common.JsonError, err.Error()
	}
	result, err := common.HttpPostJson(uri, params)
	if err != nil {
		return common.WxError, err.Error()
	}
	mylog.Log.Debugln(tag, "SendCustomMsgToOfficalAccount, result: ", string(result))
	errResult := &RestuleError{}
	err = json.Unmarshal(result, errResult)
	if err != nil {
		return common.JsonError, err.Error()
	}
	if errResult.ErrCode != 0 {
		return common.WxError, fmt.Sprintf("errcode: %d, errmsg: %s", errResult.ErrCode, errResult.ErrMsg)
	}
	return common.Success, "success"
}

/******************************************************************************
 * function: SendCustomerWelcomeCardToOfficalAccount
 * description: 发送欢迎卡片到公众号
 * param {string} 公众号的openId
 * return {*}
********************************************************************************/
func SendCustomerWelcomeCardToOfficalAccount(openId string) {
	SendCustomerMiniCardToOfficalAccount(openId, cfg.This.Wx.WelcomePath, cfg.This.Wx.WelcomeTitle, cfg.This.Wx.WelcomeCardMediaId)
}

/******************************************************************************
 * function: SendCustomerMiniCardToOfficalAccount
 * description: 发送小程序卡片
 * param {string} openId
 * return {*}
********************************************************************************/
func SendCustomerMiniCardToOfficalAccount(openId string, path string, title string, mediaId string) (int, string) {
	accessToken, err := QueryWxOfficalAccessToken()
	if err != nil {
		return common.WxError, err.Error()
	}
	uri := "https://api.weixin.qq.com/cgi-bin/message/custom/send?access_token=" + accessToken
	type CustomMsgReq struct {
		ToUser      string `json:"touser"`
		MsgType     string `json:"msgtype"`
		Miniprogram struct {
			Title        string `json:"title"`
			AppId        string `json:"appid"`
			PagePath     string `json:"pagepath"`
			ThumbMediaId string `json:"thumb_media_id"`
		} `json:"miniprogrampage"`
	}
	req := &CustomMsgReq{
		ToUser:  openId,
		MsgType: "miniprogrampage",
	}
	req.Miniprogram.AppId = cfg.This.Wx.MinAppId
	req.Miniprogram.PagePath = path
	req.Miniprogram.Title = title
	req.Miniprogram.ThumbMediaId = mediaId
	params, err := json.Marshal(req)
	if err != nil {
		return common.JsonError, err.Error()
	}
	result, err := common.HttpPostJson(uri, params)
	if err != nil {
		return common.WxError, err.Error()
	}
	mylog.Log.Debugln(tag, "SendCustomerMiniCardToOfficalAccount, result: ", string(result))
	errResult := &RestuleError{}
	err = json.Unmarshal(result, errResult)
	if err != nil {
		return common.JsonError, err.Error()
	}
	if errResult.ErrCode != 0 {
		return common.WxError, fmt.Sprintf("errcode: %d, errmsg: %s", errResult.ErrCode, errResult.ErrMsg)
	}
	return common.Success, "success"
}

/******************************************************************************
 * function: SendCustomerReportCardToOfficalAccount
 * description:
 * param {int64} userId
 * param {string} mac
 * param {string} reportDate
 * return {*}
********************************************************************************/
func SendCustomerReportMsgToOfficalAccount(userId int64, nickName string, mac string, startTime string, endTime string) (int, string) {
	miniList := make([]mysqlwx.WxMiniProgram, 0)
	mysqlwx.QueryWxMiniProgramByUserId(userId, &miniList)
	if len(miniList) == 0 {
		return common.NoData, "can not find mini program user in datebase"
	}
	// 小程序unionId
	unionId := miniList[0].UnionId
	// 查询公众号的openId
	officalList := make([]mysqlwx.WxOfficalAccount, 0)
	mysqlwx.QueryWxOfficalAccountSubscribeByUnionId(unionId, &officalList)
	if len(officalList) == 0 {
		return common.NoData, "can not find offical account user in datebase"
	}
	openId := officalList[0].FromOpenId
	// 发送模板消息
	data := map[string]interface{}{
		"thing1": map[string]interface{}{"value": nickName},
		"time2":  map[string]interface{}{"value": startTime},
		"time3":  map[string]interface{}{"value": endTime},
	}
	page := fmt.Sprintf("%s/%s", cfg.This.Wx.ReportPath, strings.ToLower(mac))
	return SendTempMessageToOfficalAccount(openId, cfg.This.Wx.ReportOfficalTemplateId, page, data)
}

/******************************************************************************
 * function: 向公众号发送每日报告
 * description:
 * param {int64} userId
 * param {string} nickName
 * param {string} mac
 * param {int} score
 * param {string} startTime
 * param {string} endTime
 * return {*}
********************************************************************************/
func SendDayReportMsgToOfficalAccount(userId int64, nickName string, mac string, score int, startTime string, endTime string) (int, string) {
	miniList := make([]mysqlwx.WxMiniProgram, 0)
	mysqlwx.QueryWxMiniProgramByUserId(userId, &miniList)
	if len(miniList) == 0 {
		return common.NoData, "can not find mini program user in datebase"
	}
	// 小程序unionId
	unionId := miniList[0].UnionId
	// 查询公众号的openId
	officalList := make([]mysqlwx.WxOfficalAccount, 0)
	mysqlwx.QueryWxOfficalAccountSubscribeByUnionId(unionId, &officalList)
	if len(officalList) == 0 {
		return common.NoData, "can not find offical account user in datebase"
	}
	openId := officalList[0].FromOpenId
	// 发送模板消息
	data := map[string]interface{}{
		"thing1":            map[string]interface{}{"value": "日报告"},
		"thing2":            map[string]interface{}{"value": nickName},
		"character_string3": map[string]interface{}{"value": score},
		"time5":             map[string]interface{}{"value": startTime},
	}
	page := fmt.Sprintf("%s/%s", cfg.This.Wx.ReportPath, strings.ToLower(mac))
	return SendTempMessageToOfficalAccount(openId, cfg.This.Wx.ReportOfficalDayTemplateId, page, data)
}

/******************************************************************************
 * function: 发送次报告到公众账号
 * description:
 * param {int64} userId
 * param {string} nickName
 * param {string} mac
 * param {string} startTime
 * param {string} endTime
 * return {*}
********************************************************************************/
func SendEveryReportMsgToOfficalAccount(userId int64, nickName string, mac string, startTime string, endTime string) (int, string) {
	miniList := make([]mysqlwx.WxMiniProgram, 0)
	mysqlwx.QueryWxMiniProgramByUserId(userId, &miniList)
	if len(miniList) == 0 {
		return common.NoData, "can not find mini program user in datebase"
	}
	// 小程序unionId
	unionId := miniList[0].UnionId
	// 查询公众号的openId
	officalList := make([]mysqlwx.WxOfficalAccount, 0)
	mysqlwx.QueryWxOfficalAccountSubscribeByUnionId(unionId, &officalList)
	if len(officalList) == 0 {
		return common.NoData, "can not find offical account user in datebase"
	}
	openId := officalList[0].FromOpenId
	// 发送模板消息
	data := map[string]interface{}{
		"thing10": map[string]interface{}{"value": "次报告"},
		"thing7":  map[string]interface{}{"value": nickName},
		"time18":  map[string]interface{}{"value": startTime},
		"time19":  map[string]interface{}{"value": endTime},
	}
	page := fmt.Sprintf("%s/%s", cfg.This.Wx.ReportPath, strings.ToLower(mac))
	return SendTempMessageToOfficalAccount(openId, cfg.This.Wx.ReportOfficalEveryTemplateId, page, data)
}

/******************************************************************************
 * function: SendH03DeviceOnlineMsgToOfficalAccount
 * description: 发送H03设备上线消息到公众号
 * param {int64} userId
 * param {string} nickName
 * param {string} mac
 * param {string} msg
 * param {string} tm
 * return {*}
********************************************************************************/
func SendH03DeviceOnlineMsgToOfficalAccount(userId int64, nickName string, mac string, msg string, tm string) (int, string) {
	miniList := make([]mysqlwx.WxMiniProgram, 0)
	mysqlwx.QueryWxMiniProgramByUserId(userId, &miniList)
	if len(miniList) == 0 {
		return common.NoData, "can not find mini program user in datebase"
	}
	// 小程序unionId
	unionId := miniList[0].UnionId
	// 查询公众号的openId
	officalList := make([]mysqlwx.WxOfficalAccount, 0)
	mysqlwx.QueryWxOfficalAccountSubscribeByUnionId(unionId, &officalList)
	if len(officalList) == 0 {
		return common.NoData, "can not find offical account user in datebase"
	}
	openId := officalList[0].FromOpenId
	// 发送模板消息
	data := map[string]interface{}{
		"thing16": map[string]interface{}{"value": nickName},
		"thing10": map[string]interface{}{"value": msg},
		"time4":   map[string]interface{}{"value": tm},
	}
	return SendTempMessageToOfficalAccount(openId, cfg.This.Wx.DeviceOnlineOfficalTemplateId, "", data)
}

/******************************************************************************
 * function: SendT1DeviceOnlineMsgToOfficalAccount
 * description: 发送T1设备上线消息到公众号
 * param {int64} userId
 * param {string} nickName
 * param {string} mac
 * param {string} msg
 * param {string} tm
 * return {*}
********************************************************************************/
func SendT1DeviceOnlineMsgToOfficalAccount(userId int64, nickName string, mac string, msg string, tm string) (int, string) {
	miniList := make([]mysqlwx.WxMiniProgram, 0)
	mysqlwx.QueryWxMiniProgramByUserId(userId, &miniList)
	if len(miniList) == 0 {
		return common.NoData, "can not find mini program user in datebase"
	}
	// 小程序unionId
	unionId := miniList[0].UnionId
	// 查询公众号的openId
	officalList := make([]mysqlwx.WxOfficalAccount, 0)
	mysqlwx.QueryWxOfficalAccountSubscribeByUnionId(unionId, &officalList)
	if len(officalList) == 0 {
		return common.NoData, "can not find offical account user in datebase"
	}
	openId := officalList[0].FromOpenId
	// 发送模板消息
	data := map[string]interface{}{
		"thing16": map[string]interface{}{"value": nickName},
		"thing10": map[string]interface{}{"value": msg},
		"time4":   map[string]interface{}{"value": tm},
	}
	return SendTempMessageToOfficalAccount(openId, cfg.This.Wx.DeviceOnlineOfficalTemplateId, "", data)
}

/******************************************************************************
 * function: SendH03DeviceStatusWarningMsgToOfficalAccount
 * description: 发送H03设备状态告警消息到公众号
 * param {int64} userId
 * param {string} nickName
 * param {string} mac
 * param {string} msg
 * param {string} tm
 * return {*}
********************************************************************************/
func SendH03DeviceStatusWarningMsgToOfficalAccount(userId int64, nickName string, mac string, msg string, tm string) (int, string) {
	miniList := make([]mysqlwx.WxMiniProgram, 0)
	mysqlwx.QueryWxMiniProgramByUserId(userId, &miniList)
	if len(miniList) == 0 {
		return common.NoData, "can not find mini program user in datebase"
	}
	// 小程序unionId
	unionId := miniList[0].UnionId
	// 查询公众号的openId
	officalList := make([]mysqlwx.WxOfficalAccount, 0)
	mysqlwx.QueryWxOfficalAccountSubscribeByUnionId(unionId, &officalList)
	if len(officalList) == 0 {
		return common.NoData, "can not find offical account user in datebase"
	}
	openId := officalList[0].FromOpenId
	// 发送模板消息
	data := map[string]interface{}{
		"thing10": map[string]interface{}{"value": nickName},
		"thing2":  map[string]interface{}{"value": msg},
		"time4":   map[string]interface{}{"value": tm},
	}
	return SendTempMessageToOfficalAccount(openId, cfg.This.Wx.DeviceStatusOfficalTemplateId, "", data)
}

/******************************************************************************
 * function: SendT1DeviceStatusWarningMsgToOfficalAccount
 * description: 发送T1设备状态告警消息到公众号
 * param {int64} userId
 * param {string} nickName
 * param {string} mac
 * param {string} msg
 * param {string} tm
 * return {*}
********************************************************************************/
func SendT1DeviceStatusWarningMsgToOfficalAccount(userId int64, nickName string, mac string, msg string, tm string) (int, string) {
	miniList := make([]mysqlwx.WxMiniProgram, 0)
	mysqlwx.QueryWxMiniProgramByUserId(userId, &miniList)
	if len(miniList) == 0 {
		return common.NoData, "can not find mini program user in datebase"
	}
	// 小程序unionId
	unionId := miniList[0].UnionId
	// 查询公众号的openId
	officalList := make([]mysqlwx.WxOfficalAccount, 0)
	mysqlwx.QueryWxOfficalAccountSubscribeByUnionId(unionId, &officalList)
	if len(officalList) == 0 {
		return common.NoData, "can not find offical account user in datebase"
	}
	openId := officalList[0].FromOpenId
	// 发送模板消息
	data := map[string]interface{}{
		"thing10": map[string]interface{}{"value": nickName},
		"thing2":  map[string]interface{}{"value": msg},
		"time4":   map[string]interface{}{"value": tm},
	}
	return SendTempMessageToOfficalAccount(openId, cfg.This.Wx.DeviceStatusOfficalTemplateId, "", data)
}

/******************************************************************************
 * function:
 * description:
 * param {string} openId
 * param {string} templateId
 * param {string} path
 * param {map[string]interface{}} data
 * return {*}
********************************************************************************/
func SendTempMessageToOfficalAccount(openId string, templateId string, path string, data map[string]interface{}) (int, string) {
	accessToken, err := QueryWxOfficalAccessToken()
	if err != nil {
		return common.WxError, err.Error()
	}
	uri := "https://api.weixin.qq.com/cgi-bin/message/template/send?access_token=" + accessToken
	type TempMsgReq struct {
		ToUser      string                 `json:"touser"`
		TemplateId  string                 `json:"template_id"`
		Data        map[string]interface{} `json:"data"`
		Url         string                 `json:"url"`
		MiniProgram map[string]interface{} `json:"miniprogram"`
	}
	req := &TempMsgReq{
		ToUser:     openId,
		TemplateId: templateId,
		Data:       data,
		// MiniProgram: map[string]interface{}{
		// 	"appid":    cfg.This.Wx.MinAppId,
		// 	"pagepath": path,
		// },
	}

	params, err := json.Marshal(req)
	if err != nil {
		return common.JsonError, err.Error()
	}
	result, err := common.HttpPostJson(uri, params)
	if err != nil {
		return common.WxError, err.Error()
	}
	mylog.Log.Debugln(tag, "SendTempMessageToOfficalAccount, result: ", string(result))
	errResult := &RestuleError{}
	err = json.Unmarshal(result, errResult)
	if err != nil {
		return common.JsonError, err.Error()
	}
	if errResult.ErrCode != 0 {
		return common.WxError, fmt.Sprintf("errcode: %d, errmsg: %s", errResult.ErrCode, errResult.ErrMsg)
	}
	return common.Success, "success"
}

func SendReportMsgToMiniProgram(userId int64, nickName string, reportType string, mac string, score int, startTime string, endTime string) (int, string) {
	miniList := make([]mysqlwx.WxMiniProgram, 0)
	mysqlwx.QueryWxMiniProgramByUserId(userId, &miniList)
	if len(miniList) == 0 {
		return common.NoData, "can not find mini program user in datebase"
	}
	openId := miniList[0].OpenId
	// 发送模板消息
	data := map[string]interface{}{
		"thing26":  map[string]interface{}{"value": nickName},
		"thing24":  map[string]interface{}{"value": reportType},
		"number16": map[string]interface{}{"value": score}, // 分数
		"time3":    map[string]interface{}{"value": startTime},
	}
	page := fmt.Sprintf("%s/?mac=%s", cfg.This.Wx.ReportPath, strings.ToLower(mac))
	return SendTempMessageToMiniProgram(openId, cfg.This.Wx.ReportMiniTemplateId, page, data)
}

/******************************************************************************
 * function: SendTempMessageToMiniProgram
 * description: 发送模板消息到小程序
 * param {string} openId
 * param {string} templateId
 * param {string} path
 * param {map[string]interface{}} data
 * return {*}
********************************************************************************/
func SendTempMessageToMiniProgram(openId string, templateId string, path string, data map[string]interface{}) (int, string) {
	accessToken, err := QueryWxMiniAccessToken()
	if err != nil {
		return common.WxError, err.Error()
	}
	uri := "https://api.weixin.qq.com/cgi-bin/message/subscribe/send?access_token=" + accessToken
	type TempMsgReq struct {
		ToUser           string                 `json:"touser"`
		TemplateId       string                 `json:"template_id"`
		Page             string                 `json:"page"`
		MiniprogramState string                 `json:"miniprogram_state"`
		Lang             string                 `json:"lang"`
		Data             map[string]interface{} `json:"data"`
	}
	req := &TempMsgReq{
		ToUser:           openId,
		TemplateId:       templateId,
		Page:             path,
		MiniprogramState: "trial", // 体验版 trial  正式版  formmal 开发版 developer
		Lang:             "zh_CN",
		Data:             data,
	}

	params, err := json.Marshal(req)
	if err != nil {
		return common.JsonError, err.Error()
	}
	mylog.Log.Debugln(tag, "SendTempMessageToMiniProgram, params: ", string(params))
	result, err := common.HttpPostJson(uri, params)
	if err != nil {
		return common.WxError, err.Error()
	}
	mylog.Log.Debugln(tag, "SendTempMessageToMiniProgram, result: ", string(result))
	errResult := &RestuleError{}
	err = json.Unmarshal(result, errResult)
	if err != nil {
		return common.JsonError, err.Error()
	}
	if errResult.ErrCode != 0 {
		return common.WxError, fmt.Sprintf("errcode: %d, errmsg: %s", errResult.ErrCode, errResult.ErrMsg)
	}
	return common.Success, "success"
}

/******************************************************************************
 * function: GeneratorWxMiniUrl
 * description: 生成微信 URL
 * param {string} path
 * param {string} query
 * return {string, error}
********************************************************************************/
func GeneratorWxMiniUrl(path string, query string) (int, string) {
	accessToken, err := QueryWxMiniAccessToken()
	if err != nil {
		return common.WxError, err.Error()
	}
	uri := fmt.Sprintf("https://api.weixin.qq.com/wxa/generate_urllink?access_token=%s", accessToken)
	type UrlLinkReq struct {
		Path           string `json:"path"`
		Query          string `json:"query"`
		ExpireType     int    `json:"expire_type"`
		ExpireInterval int    `json:"expire_interval"`
	}
	req := &UrlLinkReq{
		Path:           path,
		Query:          query,
		ExpireType:     1,
		ExpireInterval: 1,
	}
	params := bytes.NewBuffer([]byte{})
	jsonEncoder := json.NewEncoder(params)
	jsonEncoder.SetEscapeHTML(false)
	err = jsonEncoder.Encode(req)
	if err != nil {
		return common.JsonError, err.Error()
	}
	mylog.Log.Debugln(tag, "GeneratorWxMiniUrl, params: ", string(params.String()))
	result, err := common.HttpPostJson(uri, params.Bytes())
	if err != nil {
		return common.WxError, err.Error()
	}
	mylog.Log.Debugln(tag, "GeneratorWxMiniUrl, result: ", string(result))
	type UrlLinkResult struct {
		UrlLink string `json:"url_link"`
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	var urlLinkResult UrlLinkResult
	err = json.Unmarshal(result, &urlLinkResult)
	if err != nil {
		return common.JsonError, err.Error()
	}
	if urlLinkResult.ErrCode != 0 {
		return common.WxError, fmt.Sprintf("errcode: %d, errmsg: %s", urlLinkResult.ErrCode, urlLinkResult.ErrMsg)
	}
	return common.Success, urlLinkResult.UrlLink
}
