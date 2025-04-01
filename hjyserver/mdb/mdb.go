/******************************************************************************
 * Author: liguoqiang
 * Date: 2023-09-06 17:50:12
 * LastEditors: liguoqiang
 * LastEditTime: 2025-03-09 00:02:12
 * Description:
********************************************************************************/
package mdb

import (
	"hjyserver/cfg"
	"hjyserver/mdb/mysql"
)

/******************************************************
* 定义 数据库初始化函数
* 在open函数中实现数据库的打开操作
*******************************************************/

func Open() bool {
	result := mysql.Open()
	if result {
		if cfg.This.Svr.EnableH03 {
			H03MdbInit()
		}
		if cfg.This.Svr.EnableT1 {
			T1MdbInit()
		}
		if cfg.This.Svr.EnableX1s {
			X1sMdbInit()
		}
	}
	return result
}

func Close() {
	if cfg.This.Svr.EnableH03 {
		H03MdbUnini()
	}
	if cfg.This.Svr.EnableT1 {
		T1MdbUnini()
	}
	if cfg.This.Svr.EnableX1s {
		X1sMdbUnini()
	}
	mysql.Close()
}
