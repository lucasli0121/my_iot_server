/******************************************************************************
 * Author: liguoqiang
 * Date: 2024-07-14 11:48:18
 * LastEditors: liguoqiang
 * LastEditTime: 2024-07-31 14:27:36
 * Description:
********************************************************************************/
package common

import (
	"bytes"
	"fmt"
	mylog "hjyserver/log"
	"io"
	"net/http"
	"net/url"
)

const Tag = "http_client"

func HttpGet(apiUrl string, params map[string]string) ([]byte, error) {
	// TODO
	data := url.Values{}
	for k, v := range params {
		data.Set(k, v)
	}
	u, err := url.ParseRequestURI(apiUrl)
	if err != nil {
		mylog.Log.Error(Tag, "url parse error: ", err)
		return nil, err
	}
	u.RawQuery = data.Encode()
	resp, err := http.Get(u.String())
	if err != nil {
		mylog.Log.Error(Tag, "http get error: ", err)
		return nil, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	return b, err
}

func HttpPostJson(apiUrl string, params []byte) ([]byte, error) {
	// TODO
	req, err := http.NewRequest("POST", apiUrl, bytes.NewBuffer(params))
	if err != nil {
		mylog.Log.Error(Tag, "http post error: ", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		mylog.Log.Error(Tag, "http post error: ", err)
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		mylog.Log.Error(Tag, "http post error: ", resp.Status)
		return nil, fmt.Errorf("errcode: %d, errmsg: %s, url: %s", resp.StatusCode, resp.Status, apiUrl)
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	return b, err
}
