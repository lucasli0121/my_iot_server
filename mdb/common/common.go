/******************************************************************************
 * Author: liguoqiang
 * Date: 2023-08-29 20:20:29
 * LastEditors: liguoqiang
 * LastEditTime: 2025-01-08 16:40:41
 * Description:
********************************************************************************/

package common

import (
	"crypto/md5"
	"fmt"
	"hjyserver/cfg"
	"strings"
	"time"

	"github.com/google/uuid"
)

// define tables name
const (
	DeviceTbl             = "device_tbl"
	DeviceOverviewTbl     = "device_overview_tbl"
	UserTbl               = "user_tbl"
	UserDeviceRelationTbl = "user_device_relation_tbl"
	UserShareDeviceTbl    = "user_share_device_tbl"
	UserTransferDeviceTbl = "user_transfer_device_tbl"
	UserGroupTbl          = "user_group_tbl"
	FriendsTbl            = "friends_tbl"
	FallAlarmTbl          = "fall_alarm_tbl"
	FallParamsTbl         = "fall_params_tbl"
	SettingTbl            = "setting_tbl"
	LampRealDataTbl       = "lamp_real_data_tbl"
	LampEventTbl          = "lamp_event_tbl"
	LampReportTbl         = "lamp_report_tbl"
	LampControlTbl        = "lamp_control_tbl"
	LampOtaTbl            = "lamp_ota_tbl"
	X1OtaTbl              = "x1_ota_tbl"
	StudyRoomTbl          = "study_room_tbl"
	StudyRoomUserTbl      = "study_room_user_tbl"
	StudyRecordTbl        = "study_record_tbl"
	NotifySettingTbl      = "notify_setting_tbl"
)

// define sleep device notify type
const (
	PeopleType         = 1
	BreathType         = 2
	BreathAbnormalType = 3
	HeartRateType      = 4
	NurseModeType      = 5
	BeeperType         = 6
	LightType          = 7
	ImproveType        = 8
)

// 设备确认标志
const (
	DeviceUnConfirmFlag = 0
	DeviceConfirmFlag   = 1
)

// define device flag, normal device or share device
const (
	NormalDeviceFlag = 0
	ShareDeviceFlag  = 1
)

// define API response result code
const (
	Success        = 200
	RepeatData     = -21
	HasExist       = -22
	NoExist        = -23
	NoData         = -24
	NoPermission   = -25
	PasswdError    = -26
	ParamError     = -27
	RegisterFail   = -28
	RecordNotFound = -29
	DBError        = -30
	AccountHasReg  = -31
	PhoneHasReg    = -32
	EmailHasReg    = -33
	PhoneNotMatch  = -34
	EmailNotMatch  = -35
	JsonError      = -36
	CodeError      = -37
	AppIdError     = -38
	PhoneError     = -39
	WxError        = -40
	TokenError     = -41
	TypeError      = -42
	DeviceOffLine  = -43
	SameUser       = -44
	AlreadyBind    = -45
)

// define all MQ topies prefix
const HEART_RATE_DATA_TOPIC_PREFIX string = "heart/real_data"
const HEART_EVENT_TOPIC_PREFIX string = "heart/event_data"
const FAIL_CHECK_DATA_TOPIC_PREFIX string = "fall_check/real_data"
const FALL_ALARM_DATA_TOPIC_PREFIX string = "fall_alarm/real_data"
const HL77_DATA_TOPIC_PREFIX string = "hl77/real_data"
const HL77_USER_ENTER_ROOM_TOPIC string = "hl77/user_enter_room"
const HL77_CONTROL_STATUS_TOPIC string = "hl77/control_status"
const ADD_FRIEND_NOTIFY string = "add_friend/notify"
const SHARE_DEVICE_NOTIFY string = "hjy-dev/share_device/notify"
const SHARE_DEVICE_REMOVE_NOTIFY string = "hjy-dev/share_device/remove_notify"
const SHARE_DEVICE_CONFIRM string = "hjy-dev/share_device/confirm"
const TRANSFER_DEVICE_CONFIRM string = "hjy-dev/transfer_device/confirm"
const TRANSFER_CONFIRM_FINISHED string = "hjy-dev/transfer_device/confirm_finished"
const DEVICE_HEART_BEAT_TOPIC string = "hjy-dev/device/heart_beat"
const DEVICE_NOTIFY_TOPIC string = "hjy-dev/device/notify"

// const DEVICE_ONLINE_TOPIC string = "device/online"

func DeviceRecordTbl(deviceType string) string {

	return deviceType + "_record_tbl"
}
func DeviceDayReportTbl(deviceType string) string {
	return deviceType + "_day_report_tbl"
}
func DeviceRecordJsonTbl(deviceType string) string {
	return deviceType + "_record_json_tbl"
}
func DeviceDayReportJsonTbl(deviceType string) string {
	return deviceType + "_day_report_json_tbl"
}
func DeviceEventTbl(deviceType string) string {
	return deviceType + "_event_tbl"
}

func DeviceLedTbl(deviceType string) string {
	return deviceType + "_led_tbl"
}

type PageDao struct {
	PageNo     int64
	PageSize   int64
	TotalPages int64
}

// 返回一个缺省的Page信息
func NewPageDao(pageNo, pageSize int64) *PageDao {
	return &PageDao{
		PageNo:     pageNo,
		PageSize:   pageSize,
		TotalPages: 0,
	}
}

/******************************************************************************
 * function: MakeMD5
 * description: encrypt string with md5
 * return {*}
********************************************************************************/
func MakeMD5(str string) string {
	data := []byte(str)
	md5Inst := md5.New()
	md5Inst.Write(data)
	result := md5Inst.Sum([]byte(""))
	md5Str := fmt.Sprintf("%x", result)
	return md5Str
}

/******************************************************************************
 * function: GetNowTime
 * description: return current time format as "2006-01-02 15:04:05"
 * return {*}
********************************************************************************/
func GetNowTime() string {
	return time.Now().Format(cfg.TmFmtStr)
}

/******************************************************************************
 * function: GetNowDate
 * description: return current time format as "2006-01-02"
 * return {*}
********************************************************************************/
func GetNowDate() string {
	return time.Now().Format(cfg.DateFmtStr)
}

/******************************************************************************
 * function: SecondsToTimeStr
 * description: convert seconds to time string format as "2006-01-02 15:04:05"
 * param {int64} seconds
 * return {*}
********************************************************************************/
func SecondsToTimeStr(seconds int64) string {
	var tm time.Duration = time.Duration(seconds) * time.Second
	return time.Unix(int64(tm.Seconds()), 0).Format(cfg.TmFmtStr)
}

/******************************************************************************
 * function: StrToTime
 * description: convert string to time format as location time
 * param {string} tmStr
 * return {*}
********************************************************************************/
func StrToTime(tmStr string) (time.Time, error) {
	return time.ParseInLocation(cfg.TmFmtStr, tmStr, time.Local)
}
func StrToDate(tmStr string) (time.Time, error) {
	return time.ParseInLocation(cfg.DateFmtStr, tmStr, time.Local)
}

/******************************************************************************
 * function: FixPlusInPhoneString
 * description: fix + in string, replace space to + in string
 * return {*}
********************************************************************************/
func FixPlusInPhoneString(v string) string {
	var b = []byte(v)
	if b[0] == ' ' {
		b[0] = '+'
	}
	v = string(b[:])
	v = strings.Replace(v, " ", "", -1)
	v = strings.Replace(v, "\n", "", -1)
	return v
}

func MakeHeartRateTopic(mac string) string {
	return HEART_RATE_DATA_TOPIC_PREFIX + "/" + strings.ToLower(mac)
}
func MakeHeartEventTopic(mac string) string {
	return HEART_EVENT_TOPIC_PREFIX // + "/" + strings.ToLower(mac)
}
func MakeFallCheckTopic(mac string) string {
	return FAIL_CHECK_DATA_TOPIC_PREFIX // + "/" + strings.ToLower(mac)
}
func MakeFallAlarmTopic(mac string) string {
	return FALL_ALARM_DATA_TOPIC_PREFIX // + "/" + strings.ToLower(mac)
}
func MakeHl77RealDataTopic(mac string) string {
	return HL77_DATA_TOPIC_PREFIX + "/" + strings.ToLower(mac)
}
func MakeHl77ControlStatusTopic(mac string) string {
	return HL77_CONTROL_STATUS_TOPIC + "/" + strings.ToLower(mac)
}
func MakeHl77UserEnterRoomTopicByMac(mac string) string {
	return HL77_USER_ENTER_ROOM_TOPIC + "/" + strings.ToLower(mac)
}
func MakeDeviceHeartBeatTopic(mac string) string {
	return DEVICE_HEART_BEAT_TOPIC + "/" + strings.ToLower(mac)
}

func MakeAddFriendNotifyTopic(userId int64) string {
	return ADD_FRIEND_NOTIFY //+ "/" + fmt.Sprintf("%d", userId)
}

func MakeShareDeviceNotifyTopic(userId int64) string {
	return SHARE_DEVICE_NOTIFY + "/" + fmt.Sprintf("%d", userId)
}

func MakeShareDeviceRemoveNotifyTopic(userId int64) string {
	return SHARE_DEVICE_REMOVE_NOTIFY + "/" + fmt.Sprintf("%d", userId)
}

func MakeShareDeviceConfirmTopic(userId int64) string {
	return SHARE_DEVICE_CONFIRM + "/" + fmt.Sprintf("%d", userId)
}

func MakeTransferDeviceConfirmTopic(userId int64) string {
	return TRANSFER_DEVICE_CONFIRM + "/" + fmt.Sprintf("%d", userId)
}
func MakeTransferConfirmedFinishedTopic(userId int64) string {
	return TRANSFER_CONFIRM_FINISHED + "/" + fmt.Sprintf("%d", userId)
}

func MakeDeviceNotifyTopic(mac string) string {
	return DEVICE_NOTIFY_TOPIC + "/" + strings.ToLower(mac)
}

// func MakeDeviceOnlineTopic(mac string) string {
// 	return DEVICE_ONLINE_TOPIC + "/" + strings.ToLower(mac)
// }

func IsCNPhone(phone string) bool {
	if len(phone) <= 3 {
		return false
	}
	if phone[:3] == "+86" {
		return true
	} else if len(phone) == 11 && phone[0] == '1' {
		return true
	}
	return false
}
func IsHKPhone(phone string) bool {
	if len(phone) <= 4 {
		return false
	}
	if phone[:4] == "+852" {
		return true
	} else if len(phone) == 8 &&
		(phone[0] == '5' || phone[0] == '6' || phone[0] == '9' || phone[0] == '8' || phone[0] == '2' || phone[0] == '3') {
		return true
	}
	return false
}

/******************************************************************************
 * function:
 * description:
 * param {string} phone
 * param {int} alarm
 * return {*}
********************************************************************************/
func GetSleepAlarmDesc(phone string, alarm int) string {
	if len(phone) <= 4 {
		return ""
	}

	switch alarm {
	case 3001:
		if IsCNPhone(phone) {
			return cfg.This.AlarmMsg.LeaveBedAlarmMsgCN
		} else if IsHKPhone(phone) {
			return cfg.This.AlarmMsg.LeaveBedAlarmMsgHK
		} else {
			return cfg.This.AlarmMsg.LeaveBedAlarmMsgEN
		}
	case 3012:
		if IsCNPhone(phone) {
			return cfg.This.AlarmMsg.InBedAlarmMsgCN
		} else if IsHKPhone(phone) {
			return cfg.This.AlarmMsg.InBedAlarmMsgHK
		} else {
			return cfg.This.AlarmMsg.InBedAlarmMsgEN
		}
	case 3006:
		if IsCNPhone(phone) {
			return cfg.This.AlarmMsg.BreathingAbnormalAlarmMsgCN
		} else if IsHKPhone(phone) {
			return cfg.This.AlarmMsg.BreathingAbnormalAlarmMsgHK
		} else {
			return cfg.This.AlarmMsg.BreathingAbnormalAlarmMsgEN
		}
	case 3007:
		if IsCNPhone(phone) {
			return cfg.This.AlarmMsg.HeartRateAbnormalAlarmMsgCN
		} else if IsHKPhone(phone) {
			return cfg.This.AlarmMsg.HeartRateAbnormalAlarmMsgHK
		} else {
			return cfg.This.AlarmMsg.HeartRateAbnormalAlarmMsgEN
		}
	case 3008:
		if IsCNPhone(phone) {
			return cfg.This.AlarmMsg.PullRopeAlarmMsgCN
		} else if IsHKPhone(phone) {
			return cfg.This.AlarmMsg.PullRopeAlarmMsgHK
		} else {
			return cfg.This.AlarmMsg.PullRopeAlarmMsgEN
		}
	case 3009:
		if IsCNPhone(phone) {
			return cfg.This.AlarmMsg.DisarmedPullRopeAlarmMsgCN
		} else if IsHKPhone(phone) {
			return cfg.This.AlarmMsg.DisarmedPullRopeAlarmMsgHK
		} else {
			return cfg.This.AlarmMsg.DisarmedPullRopeAlarmMsgEN
		}
	case 3010:
		if IsCNPhone(phone) {
			return cfg.This.AlarmMsg.ApneaAlarmMsgCN
		} else if IsHKPhone(phone) {
			return cfg.This.AlarmMsg.ApneaAlarmMsgHK
		} else {
			return cfg.This.AlarmMsg.ApneaAlarmMsgEN
		}
	default:
		return ""
	}
}

/******************************************************************************
 * function:
 * description:
 * param {string} phone
 * param {int} status
 * return {*}
********************************************************************************/
func GetNotifyStatusDesc(phone string, notifyType int, status int) string {
	var desc string
	var isCN bool
	if len(phone) <= 4 {
		return ""
	}
	if IsCNPhone(phone) {
		isCN = true
	} else {
		isCN = false
	}
	switch notifyType {
	case PeopleType:
		if status == 0 {
			if isCN {
				desc = cfg.This.AlarmMsg.CheckPersonNoActivityMsgCN
			} else {
				desc = cfg.This.AlarmMsg.CheckPersonNoActivityMsgHK
			}
		} else {
			if isCN {
				desc = cfg.This.AlarmMsg.CheckPersonActivityMsgCN
			} else {
				desc = cfg.This.AlarmMsg.CheckPersonActivityMsgHK
			}
		}
	case BreathType:
		if status == 0 {
			if isCN {
				desc = cfg.This.AlarmMsg.CheckPersonBreathLowMsgCN
			} else {
				desc = cfg.This.AlarmMsg.CheckPersonBreathLowMsgHK
			}
		} else {
			if isCN {
				desc = cfg.This.AlarmMsg.CheckPersonBreathHighMsgCN
			} else {
				desc = cfg.This.AlarmMsg.CheckPersonBreathHighMsgHK
			}
		}
	case HeartRateType:
		if status == 0 {
			if isCN {
				desc = cfg.This.AlarmMsg.CheckPersonHeartLowMsgCN
			} else {
				desc = cfg.This.AlarmMsg.CheckPersonHeartLowMsgHK
			}
		} else {
			if isCN {
				desc = cfg.This.AlarmMsg.CheckPersonHeartHighMsgCN
			} else {
				desc = cfg.This.AlarmMsg.CheckPersonHeartHighMsgHK
			}
		}

	default:
		// 其他通知
	}
	return desc
}

func GenerateUUID() string {
	return uuid.New().String()
}

func GenerateTimeRandom() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

/******************************************************************************
 * function: IsOneDayApart
 * description: 判断是否是同一天，如果是同一天返回false，否则返回true
 * param {*} t1
 * param {time.Time} t2
 * return {*}
********************************************************************************/
func IsOneDayApart(t1, t2 time.Time) bool {
	if t1.Year() == t2.Year() && t1.YearDay() == t2.YearDay() {
		return false
	}
	return true
}
