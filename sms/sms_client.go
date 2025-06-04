/******************************************************************************
 * Author: liguoqiang
 * Date: 2024-05-24 18:25:06
 * LastEditors: liguoqiang
 * LastEditTime: 2025-06-04 09:46:23
 * Description:
********************************************************************************/
package sms

import (
	"encoding/json"
	"fmt"
	"hjyserver/exception"
	mylog "hjyserver/log"
	"hjyserver/mdb/common"
	"strings"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dysmsapi20170525 "github.com/alibabacloud-go/dysmsapi-20170525/v3/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
)

const (
	HKTempleteCode = "SMS_467535067"
	CNTempleteCode = "SMS_467555052"
)

/**
 * 使用AK&SK初始化账号Client
 * @return Client
 * @throws Exception
 */
func createClient() (_result *dysmsapi20170525.Client, _err error) {
	// 工程代码泄露可能会导致 AccessKey 泄露，并威胁账号下所有资源的安全性。以下代码示例仅供参考。
	// 建议使用更安全的 STS 方式，更多鉴权访问方式请参见：https://help.aliyun.com/document_detail/378661.html。
	config := &openapi.Config{
		// 必填，请确保代码运行环境设置了环境变量 ALIBABA_CLOUD_ACCESS_KEY_ID。
		AccessKeyId: tea.String(""),
		// 必填，请确保代码运行环境设置了环境变量 ALIBABA_CLOUD_ACCESS_KEY_SECRET。
		AccessKeySecret: tea.String(""),
	}
	// Endpoint 请参考 https://api.aliyun.com/product/Dysmsapi
	config.Endpoint = tea.String("dysmsapi.aliyuncs.com")
	_result = &dysmsapi20170525.Client{}
	_result, _err = dysmsapi20170525.NewClient(config)
	return _result, _err
}

func SendSms(phone string, nickName string, msg string) error {
	if len(phone) <= 4 {
		return exception.NewException(common.ParamError, "phone is empty")
	}
	var templateCoee string
	if common.IsCNPhone(phone) {
		templateCoee = CNTempleteCode
	} else if common.IsHKPhone(phone) {
		templateCoee = HKTempleteCode
	} else {
		return exception.NewException(common.ParamError, "phone is invalid")
	}
	// 创建 Client
	client, err := createClient()
	if err != nil {
		return err
	}
	// 发送短信
	sendSmsRequest := &dysmsapi20170525.SendSmsRequest{
		PhoneNumbers:  tea.String(phone),
		SignName:      tea.String(""),
		TemplateCode:  tea.String(templateCoee),
		TemplateParam: tea.String(fmt.Sprintf("{\"name\":\"%s\",\"msg\":\"%s\"}", nickName, msg)),
	}
	runtime := &util.RuntimeOptions{}
	tryErr := func() (_e error) {
		defer func() {
			if r := tea.Recover(recover()); r != nil {
				_e = r
			}
		}()
		resp, err := client.SendSmsWithOptions(sendSmsRequest, runtime)
		if err != nil {
			return err
		}

		mylog.Log.Infoln(util.ToJSONString(resp))

		return nil
	}()

	if tryErr != nil {
		var error = &tea.SDKError{}
		if _t, ok := tryErr.(*tea.SDKError); ok {
			error = _t
		} else {
			error.Message = tea.String(tryErr.Error())
		}
		// 此处仅做打印展示，请谨慎对待异常处理，在工程项目中切勿直接忽略异常。
		// 错误 message
		mylog.Log.Errorln(tea.StringValue(error.Message))
		// 诊断地址
		var data interface{}
		d := json.NewDecoder(strings.NewReader(tea.StringValue(error.Data)))
		d.Decode(&data)
		if m, ok := data.(map[string]interface{}); ok {
			recommend, _ := m["Recommend"]
			mylog.Log.Errorln(recommend)
		}
		_, err = util.AssertAsString(error.Message)
		if err != nil {
			return err
		}
	}
	return err
}
